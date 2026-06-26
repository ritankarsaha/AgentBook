
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/joho/godotenv"

	"github.com/ritankar/agentthreads/internal/config"
	"github.com/ritankar/agentthreads/internal/db"
	"github.com/ritankar/agentthreads/internal/nim"
	"github.com/ritankar/agentthreads/internal/personas"
	"github.com/ritankar/agentthreads/internal/platform"
)

func main() {
	_ = godotenv.Load()

	var (
		dryRun        bool
		postsPerAgent int
		agentLimit    int
		concurrency   int
		daysBack      int
	)
	flag.BoolVar(&dryRun, "dry-run", false, "print the seeding plan without calling NIM or writing to the DB")
	flag.IntVar(&postsPerAgent, "posts-per-agent", 75, "posts to generate per seed agent")
	flag.IntVar(&agentLimit, "agent-limit", 0, "if > 0, only seed the first N agents (for smoke testing)")
	flag.IntVar(&concurrency, "concurrency", 8, "max concurrent NIM generation calls")
	flag.IntVar(&daysBack, "days-back", 30, "spread backdated posts across this many days")
	flag.Parse()

	agents := personas.SeedPersonas
	if agentLimit > 0 && agentLimit < len(agents) {
		agents = agents[:agentLimit]
	}

	totalCalls := len(agents) * postsPerAgent
	log.Printf("seed: %d agents x %d posts/agent = %d planned NIM calls (dry_run=%v, concurrency=%d, days_back=%d)",
		len(agents), postsPerAgent, totalCalls, dryRun, concurrency, daysBack)

	if dryRun {
		printDryRunPlan(agents, postsPerAgent)
		return
	}

	cfg := config.Load()
	if len(cfg.NIMAPIKeys) == 0 {
		log.Fatal("seed: no NVIDIA NIM API keys configured (NVIDIA_NIM_API_KEYS / NVIDIA_NIM_API_KEY)")
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("seed: db: %v", err)
	}
	defer pool.Close()

	nimClient, err := nim.New(cfg.NIMAPIKeys, cfg.NIMBaseURL)
	if err != nil {
		log.Fatalf("seed: nim client: %v", err)
	}
	log.Printf("seed: NIM client ready, key pool size %d", nimClient.PoolSize())

	ownerID, err := ensurePlatformOwner(ctx, pool, cfg)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("seed: platform owner ready (users.id=%s, handle=@%s)", ownerID, platform.OwnerHandle)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	agentIDs := make(map[string]string, len(agents))
	var totalInserted, totalFailed int

	for _, p := range agents {
		id, alreadySeeded, err := ensureAgent(ctx, pool, ownerID, p)
		if err != nil {
			log.Fatalf("seed: %v", err)
		}
		agentIDs[p.Handle] = id

		if alreadySeeded {
			log.Printf("seed: @%s already has posts — skipping generation (resumable rerun)", p.Handle)
			continue
		}

		jobs := buildJobsForAgent(p, postsPerAgent, daysBack, now, rng)
		log.Printf("seed: generating %d posts for @%s (%s)...", len(jobs), p.Handle, p.Model)
		results := runJobs(ctx, nimClient, jobs, concurrency)

		inserted, failed := insertResults(ctx, pool, id, results)
		totalInserted += inserted
		totalFailed += failed
		log.Printf("seed: @%s done — %d inserted, %d failed", p.Handle, inserted, failed)
	}

	followsCreated, err := seedFollows(ctx, pool, agents, agentIDs, rng)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("seed: %d follow relationships created", followsCreated)

	reactionsCreated, err := seedReactions(ctx, pool, agents, agentIDs, rng)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("seed: %d reactions created", reactionsCreated)

	log.Printf("seed: done — %d agents, %d posts inserted, %d generation failures, %d follows, %d reactions",
		len(agents), totalInserted, totalFailed, followsCreated, reactionsCreated)
}

func printDryRunPlan(agents []personas.Persona, postsPerAgent int) {
	fmt.Println("=== AgentThreads seed — dry run plan (no NIM calls, no DB writes) ===")
	for _, p := range agents {
		fmt.Printf("- @%-16s %-16s model=%-32s badge=%-8s subtype=%-9s capabilities=%v\n",
			p.Handle, p.DisplayName, p.Model, p.Badge, p.PostSubtype, p.Capabilities)
	}
	fmt.Printf("\n%d agents x %d posts/agent = %d total NIM calls\n", len(agents), postsPerAgent, len(agents)*postsPerAgent)
	fmt.Println("Owner: a new dedicated platform account (agents@agentthreads.dev / @agentthreads-hq) would be created or reused.")
	fmt.Println("Follows: each agent follows 3-8 random other agents.")
	fmt.Println("Reactions: each post gets 2-15 likes from random other agents.")
}
