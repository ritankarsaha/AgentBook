package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter is an in-memory sliding-window rate limiter.
// It is safe for concurrent use.  The zero value is not useful — use NewRateLimiter.
type RateLimiter struct {
	mu      sync.Mutex
	windows map[string][]time.Time
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		windows: make(map[string][]time.Time),
		limit:   limit,
		window:  window,
	}
}

// Allow returns true if the key is within its rate limit, recording the attempt.
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	times := rl.windows[key]
	// Evict timestamps outside the sliding window.
	j := 0
	for _, t := range times {
		if t.After(cutoff) {
			times[j] = t
			j++
		}
	}
	times = times[:j]

	if len(times) >= rl.limit {
		rl.windows[key] = times
		return false
	}

	rl.windows[key] = append(times, now)
	return true
}

// Middleware wraps a handler, keying on the authenticated identity (agent ID,
// user ID, or remote IP as a last resort).
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rateLimitKey(r)
		if !rl.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"ok":false,"error":"rate limit exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func rateLimitKey(r *http.Request) string {
	if agent := AgentFromContext(r.Context()); agent != nil {
		return "agent:" + agent.ID
	}
	if claims := UserClaimsFromContext(r.Context()); claims != nil {
		return "user:" + claims.UserID
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return "ip:" + ip
	}
	return "ip:" + r.RemoteAddr
}
