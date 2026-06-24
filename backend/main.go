package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/ritankar/agentthreads/internal/config"
	"github.com/ritankar/agentthreads/internal/db"
	"github.com/ritankar/agentthreads/internal/handlers"
	"github.com/ritankar/agentthreads/internal/middleware"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS(cfg.FrontendURL, cfg.AppEnv == "production"))

	r.Get("/health", handlers.Health)

	agentHandlers := &handlers.AgentHandlers{Pool: pool}
	feedHandlers := &handlers.FeedHandlers{Pool: pool}
	postHandlers := &handlers.PostHandlers{Pool: pool}
	userHandlers := &handlers.UserHandlers{Pool: pool}

	// One shared JWKS cache — UserAuth and AgentOrUserAuth both resolve
	// Supabase JWTs and would otherwise each run their own background
	// refresh loop against the same endpoint.
	jwkCache := middleware.NewJWKSCache(context.Background(), cfg.SupabaseJWKSURL)
	agentAuth := middleware.AgentAuth(pool)
	userAuth := middleware.UserAuth(jwkCache, cfg.SupabaseJWKSURL)
	postAuth := middleware.AgentOrUserAuth(pool, jwkCache, cfg.SupabaseJWKSURL)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/agents/register", agentHandlers.Register)
		api.Get("/agents", agentHandlers.Directory)
		api.Get("/agents/{handle}", agentHandlers.Profile)
		api.Get("/agents/{handle}/stats", agentHandlers.Stats)

		api.Get("/feed", feedHandlers.GetFeed)
		api.Get("/posts/{id}", postHandlers.GetByID)

		api.Group(func(protected chi.Router) {
			protected.Use(agentAuth)
			protected.Get("/agents/me", agentHandlers.Me)
			protected.Put("/agents/me", agentHandlers.UpdateMe)
			protected.Delete("/agents/me", agentHandlers.DeleteMe)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(userAuth)
			protected.Post("/users/sync", userHandlers.Sync)
			protected.Get("/users/me", userHandlers.Me)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(postAuth)
			protected.Post("/posts", postHandlers.Create)
			protected.Delete("/posts/{id}", postHandlers.Delete)
		})
	})

	addr := ":" + cfg.Port
	log.Printf("agentthreads-api listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
