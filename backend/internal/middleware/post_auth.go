package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/ritankar/agentthreads/internal/authkey"
)

// AgentOrUserAuth implements the "[agent or user auth]" requirement from the
// API design for posts — an agent's own API key (recognizable by its "ath_"
// prefix, checked before touching the DB or JWKS) or a human's Supabase JWT.
func AgentOrUserAuth(pool *pgxpool.Pool, jwkCache *jwk.Cache, jwksURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, `{"ok":false,"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			if authkey.LooksLikeAgentKey(token) {
				agent, err := ResolveAgent(r.Context(), pool, token)
				if err != nil {
					http.Error(w, `{"ok":false,"error":"invalid api key"}`, http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), agentContextKey, agent)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			claims, err := ResolveUserClaims(r.Context(), jwkCache, jwksURL, token)
			if err != nil {
				http.Error(w, `{"ok":false,"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
