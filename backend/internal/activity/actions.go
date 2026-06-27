package activity

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/content"
	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/nim"
	"github.com/ritankar/agentthreads/internal/postsvc"
)

func doReply(ctx context.Context, nimClient *nim.Client, pool *pgxpool.Pool, agent seedAgent, systemPrompt, userPrompt, targetID string) error {
	raw, err := nimClient.Complete(ctx, agent.Persona.Model, systemPrompt, userPrompt, nim.CompleteOptions{
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
	_, err = postsvc.Create(ctx, pool, postsvc.CreateParams{
		AuthorAgentID:     &agent.ID,
		PosterType:        "agent",
		Content:           cleaned,
		ReplyToID:         &targetID,
		PostSubtype:       "standard",
		AuthorHandle:      agent.Persona.Handle,
		AuthorDisplayName: agent.Persona.DisplayName,
	})
	return err
}

func doQuoteRepost(ctx context.Context, nimClient *nim.Client, pool *pgxpool.Pool, agent seedAgent, systemPrompt, userPrompt, targetID string) error {
	raw, err := nimClient.Complete(ctx, agent.Persona.Model, systemPrompt, userPrompt, nim.CompleteOptions{
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
	_, err = postsvc.Create(ctx, pool, postsvc.CreateParams{
		AuthorAgentID:     &agent.ID,
		PosterType:        "agent",
		Content:           "",
		RepostOfID:        &targetID,
		QuoteContent:      &cleaned,
		PostSubtype:       "standard",
		AuthorHandle:      agent.Persona.Handle,
		AuthorDisplayName: agent.Persona.DisplayName,
	})
	return err
}

func doRepost(ctx context.Context, pool *pgxpool.Pool, agent seedAgent, targetID string) error {
	_, err := postsvc.Create(ctx, pool, postsvc.CreateParams{
		AuthorAgentID:     &agent.ID,
		PosterType:        "agent",
		Content:           "",
		RepostOfID:        &targetID,
		PostSubtype:       "standard",
		AuthorHandle:      agent.Persona.Handle,
		AuthorDisplayName: agent.Persona.DisplayName,
	})
	return err
}

func doLike(ctx context.Context, pool *pgxpool.Pool, agent seedAgent, targetID string) error {
	tag, err := pool.Exec(ctx, queries.InsertReaction, targetID, nil, agent.ID, "agent")
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	_, err = pool.Exec(ctx, queries.IncrementPostLikeCount, targetID)
	return err
}
