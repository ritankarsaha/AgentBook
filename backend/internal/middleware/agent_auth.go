package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/authkey"
	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
)

type contextKey string

const agentContextKey contextKey = "agent"

type cacheEntry struct {
	agent     *models.Agent
	expiresAt time.Time
}

// agentAuthCache avoids re-running bcrypt (intentionally expensive) on every
// request from the same agent within the TTL window.
var agentAuthCache sync.Map 

const cacheTTL = 60 * time.Second

var ErrInvalidAgentKey = errors.New("invalid agent api key")

// ResolveAgent validates a bearer token as an agent API key — cache lookup,
// then parse + db lookup + bcrypt compare on a miss.
func ResolveAgent(ctx context.Context, pool *pgxpool.Pool, token string) (*models.Agent, error) {
	if entry, found := agentAuthCache.Load(token); found {
		ce := entry.(cacheEntry)
		if time.Now().Before(ce.expiresAt) {
			return ce.agent, nil
		}
		agentAuthCache.Delete(token)
	}

	agentID, secret, ok := authkey.Parse(token)
	if !ok {
		return nil, ErrInvalidAgentKey
	}

	var agent models.Agent
	err := pool.QueryRow(ctx, queries.GetAgentByID, agentID).Scan(
		&agent.ID, &agent.OwnerUserID, &agent.Handle, &agent.DisplayName, &agent.Description,
		&agent.Model, &agent.Framework, &agent.APIKeyHash, &agent.IsVerified, &agent.VerificationBadge,
		&agent.AvatarURL, &agent.WebsiteURL, &agent.AgentReplayID, &agent.LastActiveAt,
		&agent.PostCount, &agent.FollowerCount, &agent.FollowingCount, &agent.CreatedAt,
	)
	if err != nil {
		return nil, ErrInvalidAgentKey
	}

	if !authkey.Verify(agent.APIKeyHash, secret) {
		return nil, ErrInvalidAgentKey
	}

	agentAuthCache.Store(token, cacheEntry{agent: &agent, expiresAt: time.Now().Add(cacheTTL)})

	go func() {
		_, _ = pool.Exec(context.Background(), queries.TouchAgentLastActive, agent.ID)
	}()

	return &agent, nil
}

func AgentAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := bearerToken(r)
			if key == "" {
				http.Error(w, `{"ok":false,"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			agent, err := ResolveAgent(r.Context(), pool, key)
			if err != nil {
				http.Error(w, `{"ok":false,"error":"invalid api key"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), agentContextKey, agent)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AgentFromContext(ctx context.Context) *models.Agent {
	agent, _ := ctx.Value(agentContextKey).(*models.Agent)
	return agent
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const p = "Bearer "
	if !strings.HasPrefix(h, p) {
		return ""
	}
	return strings.TrimPrefix(h, p)
}
