package handlers

import (
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
)

type SearchHandlers struct {
	Pool *pgxpool.Pool
}

const defaultSearchLimit = 20
const maxSearchLimit = 50

func (h *SearchHandlers) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		WriteError(w, http.StatusBadRequest, "q is required")
		return
	}
	if len(q) < 2 {
		WriteError(w, http.StatusBadRequest, "q must be at least 2 characters")
		return
	}

	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "all"
	}
	switch searchType {
	case "posts", "agents", "all":
	default:
		WriteError(w, http.StatusBadRequest, "type must be posts, agents, or all")
		return
	}

	capability := r.URL.Query().Get("capability")
	var capabilityPtr *string
	if capability != "" {
		capabilityPtr = &capability
	}

	limit := defaultSearchLimit
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 && v <= maxSearchLimit {
			limit = v
		}
	}

	ctx := r.Context()

	type searchResult struct {
		Posts  []models.Post  `json:"posts,omitempty"`
		Agents []models.Agent `json:"agents,omitempty"`
	}
	result := searchResult{}

	if searchType == "posts" || searchType == "all" {
		rows, err := h.Pool.Query(ctx, queries.SearchPosts, q, limit)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "search failed")
			return
		}
		defer rows.Close()

		posts := []models.Post{}
		for rows.Next() {
			var p models.Post
			if err := rows.Scan(
				&p.ID, &p.PosterType, &p.Content, &p.ReplyToID, &p.RepostOfID, &p.QuoteContent,
				&p.MediaURLs, &p.PostSubtype, &p.TraceURL, &p.LikeCount, &p.ReplyCount,
				&p.RepostCount, &p.EngagementScore, &p.CreatedAt,
				&p.AuthorHandle, &p.AuthorDisplayName, &p.AuthorAvatarURL, &p.AuthorIsVerified,
			); err != nil {
				continue
			}
			if p.MediaURLs == nil {
				p.MediaURLs = []string{}
			}
			posts = append(posts, p)
		}
		rows.Close()
		result.Posts = posts
	}

	if searchType == "agents" || searchType == "all" {
		rows, err := h.Pool.Query(ctx, queries.SearchAgents, q, capabilityPtr, limit)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "search failed")
			return
		}
		defer rows.Close()

		agentList := []models.Agent{}
		for rows.Next() {
			var a models.Agent
			var caps []string
			if err := rows.Scan(
				&a.ID, &a.OwnerUserID, &a.Handle, &a.DisplayName, &a.Description,
				&a.Model, &a.Framework, &a.APIKeyHash,
				&a.IsVerified, &a.VerificationBadge, &a.AvatarURL, &a.WebsiteURL,
				&a.AgentReplayID, &a.LastActiveAt, &a.PostCount, &a.FollowerCount,
				&a.FollowingCount, &a.CreatedAt, &caps,
			); err != nil {
				continue
			}
			a.APIKeyHash = "" // never expose
			if caps == nil {
				caps = []string{}
			}
			a.Capabilities = caps
			agentList = append(agentList, a)
		}
		rows.Close()
		result.Agents = agentList
	}

	WriteOK(w, result)
}
