package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type userContextKey string

const userClaimsKey userContextKey = "user_claims"

// UserClaims is the subset of a Supabase-issued JWT we care about.
type UserClaims struct {
	UserID    string
	Email     string
	FullName  string
	AvatarURL string
}

var (
	ErrAuthUnavailable    = errors.New("jwks temporarily unavailable")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrInvalidAudience    = errors.New("invalid token audience")
	ErrMissingTokenClaims = errors.New("token missing required claims")
)

func NewJWKSCache(ctx context.Context, jwksURL string) *jwk.Cache {
	cache := jwk.NewCache(ctx)
	_ = cache.Register(jwksURL, jwk.WithMinRefreshInterval(10*time.Minute))
	return cache
}

func ResolveUserClaims(ctx context.Context, cache *jwk.Cache, jwksURL, raw string) (*UserClaims, error) {
	keySet, err := cache.Get(ctx, jwksURL)
	if err != nil {
		return nil, ErrAuthUnavailable
	}

	token, err := jwt.Parse([]byte(raw), jwt.WithKeySet(keySet), jwt.WithValidate(true))
	if err != nil {
		return nil, ErrInvalidToken
	}

	if !slices.Contains(token.Audience(), "authenticated") {
		return nil, ErrInvalidAudience
	}

	claims := &UserClaims{UserID: token.Subject()}
	if email, ok := token.Get("email"); ok {
		claims.Email, _ = email.(string)
	}
	if um, ok := token.Get("user_metadata"); ok {
		if m, ok := um.(map[string]interface{}); ok {
			if v, ok := m["full_name"].(string); ok {
				claims.FullName = v
			}
			if v, ok := m["name"].(string); ok && claims.FullName == "" {
				claims.FullName = v
			}
			if v, ok := m["avatar_url"].(string); ok {
				claims.AvatarURL = v
			} else if v, ok := m["picture"].(string); ok {
				claims.AvatarURL = v
			}
		}
	}

	if claims.UserID == "" || claims.Email == "" {
		return nil, ErrMissingTokenClaims
	}
	return claims, nil
}

func UserAuth(cache *jwk.Cache, jwksURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := bearerToken(r)
			if raw == "" {
				http.Error(w, `{"ok":false,"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			claims, err := ResolveUserClaims(r.Context(), cache, jwksURL, raw)
			if err != nil {
				status := http.StatusUnauthorized
				if errors.Is(err, ErrAuthUnavailable) {
					status = http.StatusInternalServerError
				}
				http.Error(w, fmt.Sprintf(`{"ok":false,"error":%q}`, err.Error()), status)
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserClaimsFromContext(ctx context.Context) *UserClaims {
	c, _ := ctx.Value(userClaimsKey).(*UserClaims)
	return c
}
