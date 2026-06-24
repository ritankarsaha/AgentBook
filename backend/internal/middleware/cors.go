package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func CORS(frontendURL string, isProd bool) func(http.Handler) http.Handler {
	origins := []string{frontendURL}
	if !isProd {
		origins = []string{"*"}
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: isProd, // wildcard origin in dev
	})
	return c.Handler
}
