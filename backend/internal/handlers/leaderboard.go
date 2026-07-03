package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
)

type LeaderboardHandlers struct {
	Pool *pgxpool.Pool
}

const (
	leaderboardDefaultLimit = 25
	leaderboardMaxLimit     = 25
)

// LeaderboardEntry is an agent with its rank and (sort-mode-dependent)
// ranking metric attached. Only one of PostsThisWeek / EngagementRate is
// ever populated, matching whichever `sort` produced this entry.
type LeaderboardEntry struct {
	models.Agent
	Capabilities   []string `json:"capabilities"`
	Rank           int      `json:"rank"`
	PostsThisWeek  *int64   `json:"posts_this_week,omitempty"`
	EngagementRate *float64 `json:"engagement_rate,omitempty"`
}

// Get handles GET /api/v1/leaderboard?sort=followers|active|engagement|newest&limit=25
func (h *LeaderboardHandlers) Get(w http.ResponseWriter, r *http.Request) {
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "followers"
	}

	limit := leaderboardDefaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= leaderboardMaxLimit {
			limit = n
		}
	}

	ctx := r.Context()
	var entries []LeaderboardEntry
	var err error

	switch sort {
	case "followers":
		entries, err = scanLeaderboardBasic(ctx, h.Pool, queries.LeaderboardMostFollowed, limit)
	case "active":
		entries, err = scanLeaderboardActive(ctx, h.Pool, limit)
	case "engagement":
		entries, err = scanLeaderboardEngagement(ctx, h.Pool, limit)
	case "newest":
		entries, err = scanLeaderboardBasic(ctx, h.Pool, queries.LeaderboardNewest, limit)
	default:
		WriteError(w, http.StatusBadRequest, "sort must be one of: followers, active, engagement, newest")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query failed")
		return
	}

	WriteOK(w, entries)
}

func scanAgentCore(rows interface {
	Scan(dest ...any) error
}, a *models.Agent, caps *[]string, extra ...any) error {
	dest := []any{
		&a.ID, &a.OwnerUserID, &a.Handle, &a.DisplayName, &a.Description,
		&a.Model, &a.Framework, &a.APIKeyHash, &a.IsVerified, &a.VerificationBadge,
		&a.AvatarURL, &a.WebsiteURL, &a.AgentReplayID, &a.LastActiveAt,
		&a.PostCount, &a.FollowerCount, &a.FollowingCount, &a.CreatedAt,
		caps,
	}
	dest = append(dest, extra...)
	return rows.Scan(dest...)
}

func scanLeaderboardBasic(ctx context.Context, pool *pgxpool.Pool, query string, limit int) ([]LeaderboardEntry, error) {
	rows, err := pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LeaderboardEntry{}
	rank := 1
	for rows.Next() {
		var a models.Agent
		var caps []string
		if err := scanAgentCore(rows, &a, &caps); err != nil {
			return nil, err
		}
		a.APIKeyHash = ""
		if caps == nil {
			caps = []string{}
		}
		entries = append(entries, LeaderboardEntry{Agent: a, Capabilities: caps, Rank: rank})
		rank++
	}
	return entries, rows.Err()
}

func scanLeaderboardActive(ctx context.Context, pool *pgxpool.Pool, limit int) ([]LeaderboardEntry, error) {
	rows, err := pool.Query(ctx, queries.LeaderboardMostActive, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LeaderboardEntry{}
	rank := 1
	for rows.Next() {
		var a models.Agent
		var caps []string
		var postsThisWeek int64
		if err := scanAgentCore(rows, &a, &caps, &postsThisWeek); err != nil {
			return nil, err
		}
		a.APIKeyHash = ""
		if caps == nil {
			caps = []string{}
		}
		entries = append(entries, LeaderboardEntry{
			Agent: a, Capabilities: caps, Rank: rank, PostsThisWeek: &postsThisWeek,
		})
		rank++
	}
	return entries, rows.Err()
}

func scanLeaderboardEngagement(ctx context.Context, pool *pgxpool.Pool, limit int) ([]LeaderboardEntry, error) {
	rows, err := pool.Query(ctx, queries.LeaderboardHighestEngagement, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LeaderboardEntry{}
	rank := 1
	for rows.Next() {
		var a models.Agent
		var caps []string
		var engagementRate float64
		if err := scanAgentCore(rows, &a, &caps, &engagementRate); err != nil {
			return nil, err
		}
		a.APIKeyHash = ""
		if caps == nil {
			caps = []string{}
		}
		entries = append(entries, LeaderboardEntry{
			Agent: a, Capabilities: caps, Rank: rank, EngagementRate: &engagementRate,
		})
		rank++
	}
	return entries, rows.Err()
}
