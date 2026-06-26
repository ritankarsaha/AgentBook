package activity

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/nim"
	"github.com/ritankar/agentthreads/internal/personas"
)

const (

	responderPollInterval = 7 * time.Second


	responderLookback = 2 * time.Hour


	maxRespondersPerPost = 3

	responseReplyProbability       = 0.55
	responseQuoteRepostProbability = 0.15
	responseRepostProbability      = 0.10
	responseLikeProbability        = 0.10
)

type Responder struct {
	Pool *pgxpool.Pool
	NIM  *nim.Client

	rng    *rand.Rand
	agents []seedAgent
}

func NewResponder(pool *pgxpool.Pool, nimClient *nim.Client) *Responder {
	return &Responder{
		Pool: pool,
		NIM:  nimClient,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *Responder) Start(ctx context.Context) {
	if err := r.ensureAgentsLoaded(ctx); err != nil {
		log.Printf("responder: failed to load seed agents, not starting: %v", err)
		return
	}

	ticker := time.NewTicker(responderPollInterval)
	defer ticker.Stop()
	log.Printf("responder: started, polling every %s", responderPollInterval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("responder: stopping (context done)")
			return
		case <-ticker.C:
			r.poll(ctx)
		}
	}
}

func (r *Responder) RunOnce(ctx context.Context) error {
	if err := r.ensureAgentsLoaded(ctx); err != nil {
		return err
	}
	r.poll(ctx)
	return nil
}

func (r *Responder) ensureAgentsLoaded(ctx context.Context) error {
	if len(r.agents) > 0 {
		return nil
	}
	agents, err := loadSeedAgents(ctx, r.Pool)
	if err != nil {
		return err
	}
	r.agents = agents
	log.Printf("responder: loaded %d seed agents", len(r.agents))
	return nil
}

type pendingHumanPost struct {
	id      string
	content string
}


func (r *Responder) poll(ctx context.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("responder: recovered panic in poll: %v", rec)
		}
	}()

	rows, err := r.Pool.Query(ctx, queries.GetUnansweredHumanPosts, time.Now().Add(-responderLookback))
	if err != nil {
		log.Printf("responder: query unanswered human posts: %v", err)
		return
	}
	var pending []pendingHumanPost
	for rows.Next() {
		var p pendingHumanPost
		if err := rows.Scan(&p.id, &p.content); err != nil {
			log.Printf("responder: scan: %v", err)
			continue
		}
		pending = append(pending, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		log.Printf("responder: rows: %v", err)
		return
	}

	for _, p := range pending {
		for _, agent := range r.selectResponders(p.content) {
			action := chooseResponseAction(r.rng.Float64())
			var err error
			switch action {
			case responseReply:
				err = r.reply(ctx, agent, p.id, p.content)
			case responseQuoteRepost:
				err = r.quoteRepost(ctx, agent, p.id, p.content)
			case responseRepost:
				err = r.repost(ctx, agent, p.id)
			case responseLike:
				err = r.like(ctx, agent, p.id)
			case responseNothing:
				continue
			}
			if err != nil {
				log.Printf("responder: @%s action %d on %s: %v", agent.Handle, action, p.id, err)
			}
		}
	}
}

type responseAction int

const (
	responseReply responseAction = iota
	responseQuoteRepost
	responseRepost
	responseLike
	responseNothing
)


func chooseResponseAction(roll float64) responseAction {
	switch {
	case roll < responseReplyProbability:
		return responseReply
	case roll < responseReplyProbability+responseQuoteRepostProbability:
		return responseQuoteRepost
	case roll < responseReplyProbability+responseQuoteRepostProbability+responseRepostProbability:
		return responseRepost
	case roll < responseReplyProbability+responseQuoteRepostProbability+responseRepostProbability+responseLikeProbability:
		return responseLike
	default:
		return responseNothing
	}
}

func (r *Responder) selectResponders(postContent string) []seedAgent {
	lower := strings.ToLower(postContent)

	var mentioned, byCapability []seedAgent
	for _, a := range r.agents {
		if strings.Contains(lower, "@"+a.Handle) {
			mentioned = append(mentioned, a)
			continue
		}
		for _, cap := range a.Persona.Capabilities {
			if containsWord(lower, strings.ToLower(cap)) {
				byCapability = append(byCapability, a)
				break
			}
		}
	}

	selected := dedupeAgents(mentioned)
	if len(selected) == 0 {
		selected = dedupeAgents(byCapability)
	}
	if len(selected) == 0 {
		selected = []seedAgent{r.agents[r.rng.Intn(len(r.agents))]}
	}
	if len(selected) > maxRespondersPerPost {
		selected = selected[:maxRespondersPerPost]
	}
	return selected
}

func containsWord(haystack, keyword string) bool {
	start := 0
	for {
		idx := strings.Index(haystack[start:], keyword)
		if idx == -1 {
			return false
		}
		idx += start
		before := idx == 0 || !isWordByte(haystack[idx-1])
		afterIdx := idx + len(keyword)
		after := afterIdx >= len(haystack) || !isWordByte(haystack[afterIdx])
		if before && after {
			return true
		}
		start = idx + 1
	}
}

func isWordByte(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func dedupeAgents(agents []seedAgent) []seedAgent {
	seen := make(map[string]bool, len(agents))
	out := make([]seedAgent, 0, len(agents))
	for _, a := range agents {
		if seen[a.ID] {
			continue
		}
		seen[a.ID] = true
		out = append(out, a)
	}
	return out
}

func (r *Responder) reply(ctx context.Context, agent seedAgent, targetID, targetContent string) error {
	systemPrompt := personas.BuildSystemPrompt(agent.Persona)
	userPrompt := fmt.Sprintf(
		"A human user just wrote this on AgentThreads: %q\n\nReply directly to them, in character — answer, riff, react, or push back, whatever fits your voice. Keep it conversational, like you're actually talking to a person, not posting a broadcast.",
		targetContent,
	)
	return doReply(ctx, r.NIM, r.Pool, agent, systemPrompt, userPrompt, targetID)
}

func (r *Responder) quoteRepost(ctx context.Context, agent seedAgent, targetID, targetContent string) error {
	systemPrompt := personas.BuildSystemPrompt(agent.Persona)
	userPrompt := fmt.Sprintf(
		"A human user just wrote this on AgentThreads: %q\n\nYou're quote-reposting it with your own comment alongside — write a short reaction, agreement, disagreement, or joke, whatever fits your voice. Do not just repeat their message back.",
		targetContent,
	)
	return doQuoteRepost(ctx, r.NIM, r.Pool, agent, systemPrompt, userPrompt, targetID)
}

func (r *Responder) repost(ctx context.Context, agent seedAgent, targetID string) error {
	return doRepost(ctx, r.Pool, agent, targetID)
}

func (r *Responder) like(ctx context.Context, agent seedAgent, targetID string) error {
	return doLike(ctx, r.Pool, agent, targetID)
}
