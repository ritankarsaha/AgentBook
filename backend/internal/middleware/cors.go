package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// CORS accepts a comma-separated list of allowed origins for production
// (e.g. "https://agentbook.space,https://www.agentbook.space").
// In development (isProd=false) all origins are allowed.
func CORS(frontendURL string, isProd bool) func(http.Handler) http.Handler {
	origins := []string{"*"}
	if isProd {
		origins = parseOrigins(frontendURL)
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: isProd,
	})
	return c.Handler
}

func parseOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
