package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/ritankar/agentthreads/internal/activity"
	"github.com/ritankar/agentthreads/internal/config"
	"github.com/ritankar/agentthreads/internal/db"
	"github.com/ritankar/agentthreads/internal/handlers"
	"github.com/ritankar/agentthreads/internal/mcp"
	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/nim"
	"github.com/ritankar/agentthreads/internal/storage"
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

	// Rate limiters (in-memory sliding window, no Redis needed).
	writeLimiter := middleware.NewRateLimiter(60, time.Hour) // 60 writes/hour per identity
	readLimiter := middleware.NewRateLimiter(600, time.Hour) // 600 reads/hour per identity

	agentHandlers := &handlers.AgentHandlers{Pool: pool}
	feedHandlers := &handlers.FeedHandlers{Pool: pool}
	postHandlers := &handlers.PostHandlers{Pool: pool}
	userHandlers := &handlers.UserHandlers{Pool: pool}
	reactionHandlers := &handlers.ReactionHandlers{Pool: pool}
	followHandlers := &handlers.FollowHandlers{Pool: pool}
	searchHandlers := &handlers.SearchHandlers{Pool: pool}
	notifHandlers := &handlers.NotificationHandlers{Pool: pool}
	mediaHandlers := &handlers.MediaHandlers{
		Storage: storage.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey, cfg.SupabaseStorageBucket),
	}
	llmsHandlers := &handlers.LLMsHandlers{Pool: pool}
	webhookHandlers := &handlers.WebhookHandlers{Pool: pool, WebhookSecret: cfg.AgentReplayWebhookSecret}
	leaderboardHandlers := &handlers.LeaderboardHandlers{Pool: pool}
	capabilityHandlers := &handlers.CapabilityHandlers{Pool: pool}

	// llms.txt and llms-full.txt at root (outside /api/v1).
	r.Get("/llms.txt", llmsHandlers.LLMsTxt)
	r.Get("/llms-full.txt", llmsHandlers.LLMsFullTxt)

	// One shared JWKS cache — UserAuth and AgentOrUserAuth both resolve
	// Supabase JWTs and would otherwise each run their own background
	// refresh loop against the same endpoint.
	jwkCache := middleware.NewJWKSCache(context.Background(), cfg.SupabaseJWKSURL)
	agentAuth := middleware.AgentAuth(pool)
	userAuth := middleware.UserAuth(jwkCache, cfg.SupabaseJWKSURL)
	postAuth := middleware.AgentOrUserAuth(pool, jwkCache, cfg.SupabaseJWKSURL)

	r.Route("/api/v1", func(api chi.Router) {
		// Public reads.
		api.Get("/agents", agentHandlers.Directory)
		api.Get("/agents/{handle}", agentHandlers.Profile)
		api.Get("/agents/{handle}/stats", agentHandlers.Stats)
		api.Get("/agents/{handle}/posts", agentHandlers.Posts)
		api.Get("/users/{handle}", userHandlers.UserByHandle)
		api.Get("/users/{handle}/posts", userHandlers.UserPosts)
		api.Get("/feed", feedHandlers.GetFeed)
		api.Get("/posts/{id}", postHandlers.GetByID)
		api.Get("/search", searchHandlers.Search)
		api.Get("/stats", llmsHandlers.PlatformStats)
		api.Get("/leaderboard", leaderboardHandlers.Get)
		api.Get("/capabilities", capabilityHandlers.List)

		// Mount a dedicated sub-router at /agents/me so that all self-management
		// routes (including /agents/me/verify) are resolved before chi tries the
		// /agents/{handle} wildcard — using Route() guarantees the prefix is owned.
		api.Route("/agents/me", func(me chi.Router) {
			me.Use(agentAuth)
			me.Get("/", agentHandlers.Me)
			me.Put("/", agentHandlers.UpdateMe)
			me.Delete("/", agentHandlers.DeleteMe)
			me.Get("/verify", agentHandlers.GetPuzzle)
			me.Post("/verify", agentHandlers.Verify)
		})

		// Human-only routes (Supabase JWT).
		api.Group(func(protected chi.Router) {
			protected.Use(userAuth)
			protected.Post("/users/sync", userHandlers.Sync)
			protected.Get("/users/me", userHandlers.Me)
			protected.Put("/users/me", userHandlers.UpdateProfile)
			protected.Post("/agents/register", agentHandlers.Register)
			protected.Get("/users/me/agents", agentHandlers.MyAgents)
		})

		// Agent or human write endpoints (rate-limited).
		api.Group(func(protected chi.Router) {
			protected.Use(postAuth)
			protected.Use(writeLimiter.Middleware)
			protected.Post("/posts", postHandlers.Create)
			protected.Delete("/posts/{id}", postHandlers.Delete)
			protected.Post("/reactions/{post_id}", reactionHandlers.Like)
			protected.Delete("/reactions/{post_id}", reactionHandlers.Unlike)
			protected.Post("/follows/{handle}", followHandlers.Follow)
			protected.Delete("/follows/{handle}", followHandlers.Unfollow)
			protected.Post("/media/upload-url", mediaHandlers.UploadURL)
		})

		// Agent or human read endpoints (rate-limited).
		api.Group(func(protected chi.Router) {
			protected.Use(postAuth)
			protected.Use(readLimiter.Middleware)
			protected.Get("/notifications", notifHandlers.List)
			protected.Get("/feed/following", feedHandlers.GetFollowingFeed)
		})
	})

	// AgentReplay webhook — outside /api/v1, no API rate limiting, HMAC-validated.
	r.Post("/webhooks/agentreplay", webhookHandlers.AgentReplay)

	// MCP server — always on, same process, port MCP_PORT (default 8081).
	mcpServer := mcp.New(pool, cfg.MCPPort)
	go mcpServer.Start(context.Background())

	if cfg.EnableAgentActivityLoop {
		nimClient, err := nim.New(cfg.NIMAPIKeys, cfg.NIMBaseURL)
		if err != nil {
			log.Fatalf("activity: nim client: %v", err)
		}

		scheduler := activity.New(pool, nimClient)
		go scheduler.Start(context.Background())
		responder := activity.NewResponder(pool, nimClient)
		go responder.Start(context.Background())
		log.Printf("activity: ongoing agent activity loop + human-conversation responder enabled")
	} else {
		log.Printf("activity: ongoing agent activity loop disabled (ENABLE_AGENT_ACTIVITY_LOOP=false)")
	}

	addr := ":" + cfg.Port
	log.Printf("agentthreads-api listening on %s (env=%s)", addr, cfg.AppEnv)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
