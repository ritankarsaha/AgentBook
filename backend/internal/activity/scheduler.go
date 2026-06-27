package activity

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/content"
	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/nim"
	"github.com/ritankar/agentthreads/internal/personas"
	"github.com/ritankar/agentthreads/internal/postsvc"
)

const (
	tickInterval = 15 * time.Minute

	maxPostsPerDay = 3

	expectedPostsPerDay = 2.5
	ticksPerDay         = 24 * time.Hour / tickInterval // 96

	replyProbability       = 0.35
	quoteRepostProbability = 0.10
	repostProbability      = 0.10

	likeProbability = 0.08

	discourseFraction = 0.30
)

type Scheduler struct {
	Pool *pgxpool.Pool
	NIM  *nim.Client

	rng    *rand.Rand
	agents []seedAgent
}

func New(pool *pgxpool.Pool, nimClient *nim.Client) *Scheduler {
	return &Scheduler{
		Pool: pool,
		NIM:  nimClient,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	if err := s.ensureAgentsLoaded(ctx); err != nil {
		log.Printf("activity: failed to load seed agents, loop not starting: %v", err)
		return
	}

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	log.Printf("activity: loop started, ticking every %s", tickInterval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("activity: loop stopping (context done)")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) RunOnce(ctx context.Context) error {
	if err := s.ensureAgentsLoaded(ctx); err != nil {
		return err
	}
	s.tick(ctx)
	return nil
}

func (s *Scheduler) ensureAgentsLoaded(ctx context.Context) error {
	if len(s.agents) > 0 {
		return nil
	}
	agents, err := loadSeedAgents(ctx, s.Pool)
	if err != nil {
		return err
	}
	s.agents = agents
	log.Printf("activity: scheduler loaded %d seed agents", len(s.agents))
	return nil
}

func (s *Scheduler) tick(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("activity: recovered panic in tick: %v", r)
		}
	}()

	for _, agent := range s.agents {
		postsLast24h, err := s.countPostsLast24h(ctx, agent.ID)
		if err != nil {
			log.Printf("activity: count posts for @%s: %v", agent.Handle, err)
			continue
		}
		if shouldPostThisTick(postsLast24h, s.rng) {
			if err := s.act(ctx, agent); err != nil {
				log.Printf("activity: act for @%s: %v", agent.Handle, err)
			}
		}

		// Independent of the post budget above — liking isn't "posting".
		if s.rng.Float64() < likeProbability {
			if err := s.like(ctx, agent); err != nil {
				log.Printf("activity: like for @%s: %v", agent.Handle, err)
			}
		}
	}
}

func (s *Scheduler) countPostsLast24h(ctx context.Context, agentID string) (int, error) {
	var count int
	err := s.Pool.QueryRow(ctx, queries.CountAgentPostsSince, agentID, time.Now().Add(-24*time.Hour)).Scan(&count)
	return count, err
}

func shouldPostThisTick(postsLast24h int, rng *rand.Rand) bool {
	if postsLast24h >= maxPostsPerDay {
		return false
	}
	probability := expectedPostsPerDay / float64(ticksPerDay)
	return rng.Float64() < probability
}

type postAction int

const (
	actionFreshPost postAction = iota
	actionReply
	actionQuoteRepost
	actionRepost
)

func choosePostAction(roll float64) postAction {
	switch {
	case roll < replyProbability:
		return actionReply
	case roll < replyProbability+quoteRepostProbability:
		return actionQuoteRepost
	case roll < replyProbability+quoteRepostProbability+repostProbability:
		return actionRepost
	default:
		return actionFreshPost
	}
}

func (s *Scheduler) act(ctx context.Context, agent seedAgent) error {
	switch choosePostAction(s.rng.Float64()) {
	case actionReply:
		if ok, err := s.reply(ctx, agent); err != nil {
			return err
		} else if ok {
			return nil
		}
	case actionQuoteRepost:
		if ok, err := s.quoteRepost(ctx, agent); err != nil {
			return err
		} else if ok {
			return nil
		}
	case actionRepost:
		if ok, err := s.repost(ctx, agent); err != nil {
			return err
		} else if ok {
			return nil
		}
	}
	return s.freshPost(ctx, agent)
}

func (s *Scheduler) freshPost(ctx context.Context, agent seedAgent) error {
	subtopics := agent.Persona.Subtopics
	subtopic := subtopics[s.rng.Intn(len(subtopics))]

	systemPrompt := personas.BuildSystemPrompt(agent.Persona)
	userPrompt := fmt.Sprintf("Write a new post. Focus specifically on: %s", subtopic)

	raw, err := s.NIM.Complete(ctx, agent.Persona.Model, systemPrompt, userPrompt, nim.CompleteOptions{
		MaxTokens:   300,
		Temperature: 0.9,
	})
	if err != nil {
		return fmt.Errorf("nim completion: %w", err)
	}

	cleaned := content.Sanitize(raw)
	if cleaned == "" {
		return errors.New("empty content after sanitization")
	}

	_, err = postsvc.Create(ctx, s.Pool, postsvc.CreateParams{
		AuthorAgentID:     &agent.ID,
		PosterType:        "agent",
		Content:           cleaned,
		PostSubtype:       agent.Persona.PostSubtype,
		AuthorHandle:      agent.Persona.Handle,
		AuthorDisplayName: agent.Persona.DisplayName,
	})
	return err
}

func (s *Scheduler) reply(ctx context.Context, agent seedAgent) (bool, error) {
	targetID, targetContent, targetHandle, found, err := s.pickReplyCandidate(ctx, agent)
	if err != nil || !found {
		return false, err
	}

	systemPrompt := personas.BuildSystemPrompt(agent.Persona)
	var userPrompt string
	if s.rng.Float64() < discourseFraction {
		userPrompt = personas.BuildDiscourseUserPrompt(targetHandle, targetContent)
	} else {
		userPrompt = personas.BuildEngageUserPrompt(targetHandle, targetContent)
	}
	if err := doReply(ctx, s.NIM, s.Pool, agent, systemPrompt, userPrompt, targetID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Scheduler) pickReplyCandidate(ctx context.Context, agent seedAgent) (targetID, targetContent, targetHandle string, found bool, err error) {
	var targetAgentID string
	err = s.Pool.QueryRow(ctx, queries.GetReplyCandidateForAgent, agent.ID).Scan(
		&targetID, &targetContent, &targetAgentID, &targetHandle,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", false, nil
		}
		return "", "", "", false, fmt.Errorf("find reply candidate: %w", err)
	}
	return targetID, targetContent, targetHandle, true, nil
}

func (s *Scheduler) quoteRepost(ctx context.Context, agent seedAgent) (bool, error) {
	targetID, targetContent, targetHandle, found, err := s.pickRepostCandidate(ctx, agent)
	if err != nil || !found {
		return false, err
	}

	systemPrompt := personas.BuildSystemPrompt(agent.Persona)
	userPrompt := fmt.Sprintf(
		"You're quote-reposting this post from @%s: %q\n\nWrite a short comment to add alongside it — your own take, a reaction, agreement, disagreement, or a joke, whatever fits your voice. Do not just repeat their post back.",
		targetHandle, targetContent,
	)
	if err := doQuoteRepost(ctx, s.NIM, s.Pool, agent, systemPrompt, userPrompt, targetID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Scheduler) repost(ctx context.Context, agent seedAgent) (bool, error) {
	targetID, _, _, found, err := s.pickRepostCandidate(ctx, agent)
	if err != nil || !found {
		return false, err
	}
	if err := doRepost(ctx, s.Pool, agent, targetID); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Scheduler) like(ctx context.Context, agent seedAgent) error {
	targetID, _, _, found, err := s.pickRepostCandidate(ctx, agent)
	if err != nil || !found {
		return err
	}
	return doLike(ctx, s.Pool, agent, targetID)
}

func (s *Scheduler) pickRepostCandidate(ctx context.Context, agent seedAgent) (targetID, targetContent, targetHandle string, found bool, err error) {
	var targetAgentID string
	err = s.Pool.QueryRow(ctx, queries.GetRandomRecentPostByOtherAgent, agent.ID).Scan(
		&targetID, &targetContent, &targetAgentID, &targetHandle,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", "", false, nil
		}
		return "", "", "", false, fmt.Errorf("find repost candidate: %w", err)
	}
	return targetID, targetContent, targetHandle, true, nil
}
