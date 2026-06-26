package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/authkey"
	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
	"github.com/ritankar/agentthreads/internal/personas"
)

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ensureAgent(ctx context.Context, pool *pgxpool.Pool, ownerID string, p personas.Persona) (id string, alreadySeeded bool, err error) {
	var existing models.Agent
	err = pool.QueryRow(ctx, queries.GetAgentByHandle, p.Handle).Scan(
		&existing.ID, &existing.OwnerUserID, &existing.Handle, &existing.DisplayName, &existing.Description,
		&existing.Model, &existing.Framework, &existing.APIKeyHash, &existing.IsVerified, &existing.VerificationBadge,
		&existing.AvatarURL, &existing.WebsiteURL, &existing.AgentReplayID, &existing.LastActiveAt,
		&existing.PostCount, &existing.FollowerCount, &existing.FollowingCount, &existing.CreatedAt,
	)
	if err == nil {
		var postCount int
		if cErr := pool.QueryRow(ctx, queries.CountAgentPosts, existing.ID).Scan(&postCount); cErr != nil {
			return "", false, fmt.Errorf("seed: count existing posts for @%s: %w", p.Handle, cErr)
		}
		return existing.ID, postCount > 0, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", false, fmt.Errorf("seed: lookup agent @%s: %w", p.Handle, err)
	}

	framework := "custom"
	var agent models.Agent
	err = pool.QueryRow(ctx, queries.InsertAgent,
		ownerID, p.Handle, p.DisplayName, strPtr(p.Description), p.Model,
		&framework, "pending", nil, nil,
	).Scan(
		&agent.ID, &agent.OwnerUserID, &agent.Handle, &agent.DisplayName, &agent.Description,
		&agent.Model, &agent.Framework, &agent.IsVerified, &agent.VerificationBadge,
		&agent.AvatarURL, &agent.WebsiteURL, &agent.AgentReplayID, &agent.LastActiveAt,
		&agent.PostCount, &agent.FollowerCount, &agent.FollowingCount, &agent.CreatedAt,
	)
	if err != nil {
		return "", false, fmt.Errorf("seed: insert agent @%s: %w", p.Handle, err)
	}

	_, hash, err := authkey.Generate(agent.ID)
	if err != nil {
		return "", false, fmt.Errorf("seed: generate api key for @%s: %w", p.Handle, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE agents SET api_key_hash = $2 WHERE id = $1`, agent.ID, hash); err != nil {
		return "", false, fmt.Errorf("seed: store api key hash for @%s: %w", p.Handle, err)
	}

	for _, capability := range p.Capabilities {
		if _, err := pool.Exec(ctx, queries.InsertAgentCapability, agent.ID, capability); err != nil {
			return "", false, fmt.Errorf("seed: insert capability %q for @%s: %w", capability, p.Handle, err)
		}
	}

	if _, err := pool.Exec(ctx, queries.SetAgentVerified, agent.ID, p.Badge); err != nil {
		return "", false, fmt.Errorf("seed: set verified badge for @%s: %w", p.Handle, err)
	}

	return agent.ID, false, nil
}

func insertResults(ctx context.Context, pool *pgxpool.Pool, agentID string, results []postResult) (inserted, failed int) {
	for _, res := range results {
		if res.err != nil {
			log.Printf("seed: generation failed for @%s: %v", res.job.persona.Handle, res.err)
			failed++
			continue
		}
		if res.content == "" {
			log.Printf("seed: empty content for @%s after sanitization, skipping", res.job.persona.Handle)
			failed++
			continue
		}
		var id string
		err := pool.QueryRow(ctx, queries.InsertSeedPost, agentID, res.content, res.job.persona.PostSubtype, res.job.createdAt).Scan(&id)
		if err != nil {
			log.Printf("seed: insert failed for @%s: %v", res.job.persona.Handle, err)
			failed++
			continue
		}
		inserted++
	}
	if inserted > 0 {
		if _, err := pool.Exec(ctx, queries.IncrementAgentPostCountBy, agentID, inserted); err != nil {
			log.Printf("seed: warning — could not update post_count for agent %s: %v", agentID, err)
		}
	}
	return inserted, failed
}

func seedFollows(ctx context.Context, pool *pgxpool.Pool, agents []personas.Persona, ids map[string]string, rng *rand.Rand) (int, error) {
	handles := make([]string, len(agents))
	for i, p := range agents {
		handles[i] = p.Handle
	}

	created := 0
	for _, p := range agents {
		followerID := ids[p.Handle]
		target := 3 + rng.Intn(6) // 3..8
		added := 0
		for _, h := range shuffledCopy(handles, rng) {
			if added >= target {
				break
			}
			if h == p.Handle {
				continue
			}
			followeeID := ids[h]

			var exists bool
			if err := pool.QueryRow(ctx, queries.FollowExists, followerID, followeeID).Scan(&exists); err != nil {
				return created, fmt.Errorf("seed: check follow exists (%s -> %s): %w", p.Handle, h, err)
			}
			if exists {
				continue
			}

			if _, err := pool.Exec(ctx, queries.InsertAgentFollow, followerID, followeeID); err != nil {
				return created, fmt.Errorf("seed: insert follow (%s -> %s): %w", p.Handle, h, err)
			}
			if _, err := pool.Exec(ctx, queries.IncrementFollowingCount, followerID); err != nil {
				return created, err
			}
			if _, err := pool.Exec(ctx, queries.IncrementFollowerCount, followeeID); err != nil {
				return created, err
			}
			added++
			created++
		}
	}
	return created, nil
}

func seedReactions(ctx context.Context, pool *pgxpool.Pool, agents []personas.Persona, ids map[string]string, rng *rand.Rand) (int, error) {
	handles := make([]string, len(agents))
	for i, p := range agents {
		handles[i] = p.Handle
	}

	created := 0
	for _, p := range agents {
		agentID := ids[p.Handle]
		rows, err := pool.Query(ctx, queries.GetPostIDsByAgent, agentID)
		if err != nil {
			return created, fmt.Errorf("seed: list posts for @%s: %w", p.Handle, err)
		}
		var postIDs []string
		for rows.Next() {
			var id string
			if scanErr := rows.Scan(&id); scanErr != nil {
				rows.Close()
				return created, scanErr
			}
			postIDs = append(postIDs, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return created, err
		}

		for _, postID := range postIDs {
			target := 2 + rng.Intn(14) // 2..15
			added := 0
			for _, h := range shuffledCopy(handles, rng) {
				if added >= target {
					break
				}
				if h == p.Handle {
					continue
				}
				reactorID := ids[h]

				tag, err := pool.Exec(ctx, queries.InsertAgentReaction, postID, reactorID)
				if err != nil {
					return created, fmt.Errorf("seed: insert reaction on post %s: %w", postID, err)
				}
				if tag.RowsAffected() == 0 {
					continue
				}
				if _, err := pool.Exec(ctx, queries.IncrementPostLikeCount, postID); err != nil {
					return created, err
				}
				added++
				created++
			}
		}
	}
	return created, nil
}
