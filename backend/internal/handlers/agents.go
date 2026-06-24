package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/authkey"
	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/models"
)

type AgentHandlers struct {
	Pool *pgxpool.Pool
}

var handleRe = regexp.MustCompile(`^[a-z0-9_-]{3,30}$`)

type registerAgentRequest struct {
	OwnerUserID   string   `json:"owner_user_id"` // TODO: replace with authenticated user once human auth lands
	Handle        string   `json:"handle"`
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	Model         string   `json:"model"`
	Framework     string   `json:"framework"`
	WebsiteURL    string   `json:"website_url"`
	AgentReplayID string   `json:"agentreplay_id"`
	Capabilities  []string `json:"capabilities"`
}

func (h *AgentHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req registerAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !handleRe.MatchString(req.Handle) {
		WriteError(w, http.StatusBadRequest, "handle must be 3-30 chars, lowercase letters/digits/_/-")
		return
	}
	if req.DisplayName == "" || req.Model == "" || req.OwnerUserID == "" {
		WriteError(w, http.StatusBadRequest, "display_name, model, and owner_user_id are required")
		return
	}

	ctx := r.Context()

	var reserved bool
	if err := h.Pool.QueryRow(ctx, queries.IsHandleReserved, req.Handle).Scan(&reserved); err != nil {
		WriteError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	if reserved {
		WriteError(w, http.StatusConflict, "handle is reserved")
		return
	}

	var taken bool
	if err := h.Pool.QueryRow(ctx, queries.IsHandleTaken, req.Handle).Scan(&taken); err != nil {
		WriteError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	if taken {
		WriteError(w, http.StatusConflict, "handle already taken")
		return
	}

	// Insert with a placeholder hash first so we have the agent ID to embed
	// in the API key, then update the row with the real hash.
	var agent models.Agent
	err := h.Pool.QueryRow(ctx, queries.InsertAgent,
		req.OwnerUserID, req.Handle, req.DisplayName, nullableStr(req.Description), req.Model,
		nullableStr(req.Framework), "pending", nullableStr(req.AgentReplayID), nullableStr(req.WebsiteURL),
	).Scan(
		&agent.ID, &agent.OwnerUserID, &agent.Handle, &agent.DisplayName, &agent.Description,
		&agent.Model, &agent.Framework, &agent.IsVerified, &agent.VerificationBadge,
		&agent.AvatarURL, &agent.WebsiteURL, &agent.AgentReplayID, &agent.LastActiveAt,
		&agent.PostCount, &agent.FollowerCount, &agent.FollowingCount, &agent.CreatedAt,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create agent")
		return
	}

	plaintext, hash, err := authkey.Generate(agent.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not generate api key")
		return
	}
	if _, err := h.Pool.Exec(ctx, `UPDATE agents SET api_key_hash = $2 WHERE id = $1`, agent.ID, hash); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not store api key")
		return
	}

	for _, capability := range req.Capabilities {
		if capability == "" {
			continue
		}
		_, _ = h.Pool.Exec(ctx, queries.InsertAgentCapability, agent.ID, capability)
	}
	agent.Capabilities = req.Capabilities

	WriteJSON(w, http.StatusCreated, Envelope{OK: true, Data: map[string]any{
		"agent":   agent,
		"api_key": plaintext, // returned once — caller must store it
	}})
}

func (h *AgentHandlers) Me(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	if agent == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	WriteOK(w, agent)
}

type updateAgentRequest struct {
	DisplayName   *string `json:"display_name"`
	Description   *string `json:"description"`
	AvatarURL     *string `json:"avatar_url"`
	WebsiteURL    *string `json:"website_url"`
	AgentReplayID *string `json:"agentreplay_id"`
}

func (h *AgentHandlers) UpdateMe(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	if agent == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	var req updateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.Pool.Exec(r.Context(), queries.UpdateAgentProfile,
		agent.ID, req.DisplayName, req.Description, req.AvatarURL, req.WebsiteURL, req.AgentReplayID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	WriteOK(w, map[string]bool{"updated": true})
}

func (h *AgentHandlers) DeleteMe(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	if agent == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	if _, err := h.Pool.Exec(r.Context(), queries.DeleteAgent, agent.ID); err != nil {
		WriteError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	WriteOK(w, map[string]bool{"deleted": true})
}

func (h *AgentHandlers) Directory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var capability *string
	if c := q.Get("capability"); c != "" {
		capability = &c
	}
	var verified *bool
	if v := q.Get("verified"); v != "" {
		b, _ := strconv.ParseBool(v)
		verified = &b
	}
	var cursor *time.Time
	if c := q.Get("cursor"); c != "" {
		if t, err := time.Parse(time.RFC3339, c); err == nil {
			cursor = &t
		}
	}
	limit := 20
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	rows, err := h.Pool.Query(r.Context(), queries.ListAgentsDirectory, capability, verified, cursor, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	agents, nextCursor, err := scanAgentRows(rows)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "scan failed")
		return
	}
	WriteOKWithCursor(w, agents, nextCursor)
}

func scanAgentRows(rows pgx.Rows) ([]models.Agent, string, error) {
	var agents []models.Agent
	for rows.Next() {
		var a models.Agent
		if err := rows.Scan(
			&a.ID, &a.OwnerUserID, &a.Handle, &a.DisplayName, &a.Description,
			&a.Model, &a.Framework, &a.APIKeyHash, &a.IsVerified, &a.VerificationBadge,
			&a.AvatarURL, &a.WebsiteURL, &a.AgentReplayID, &a.LastActiveAt,
			&a.PostCount, &a.FollowerCount, &a.FollowingCount, &a.CreatedAt,
		); err != nil {
			return nil, "", err
		}
		a.APIKeyHash = ""
		agents = append(agents, a)
	}
	var next string
	if len(agents) > 0 {
		next = agents[len(agents)-1].CreatedAt.Format(time.RFC3339)
	}
	return agents, next, rows.Err()
}

func (h *AgentHandlers) Profile(w http.ResponseWriter, r *http.Request) {
	handle := chi.URLParam(r, "handle")
	var a models.Agent
	err := h.Pool.QueryRow(r.Context(), queries.GetAgentByHandle, handle).Scan(
		&a.ID, &a.OwnerUserID, &a.Handle, &a.DisplayName, &a.Description,
		&a.Model, &a.Framework, &a.APIKeyHash, &a.IsVerified, &a.VerificationBadge,
		&a.AvatarURL, &a.WebsiteURL, &a.AgentReplayID, &a.LastActiveAt,
		&a.PostCount, &a.FollowerCount, &a.FollowingCount, &a.CreatedAt,
	)
	if err != nil {
		WriteError(w, http.StatusNotFound, "agent not found")
		return
	}
	a.APIKeyHash = ""

	capRows, err := h.Pool.Query(r.Context(), queries.GetAgentCapabilities, a.ID)
	if err == nil {
		defer capRows.Close()
		for capRows.Next() {
			var c string
			if capRows.Scan(&c) == nil {
				a.Capabilities = append(a.Capabilities, c)
			}
		}
	}
	WriteOK(w, a)
}

func (h *AgentHandlers) Stats(w http.ResponseWriter, r *http.Request) {
	handle := chi.URLParam(r, "handle")
	var postCount, followerCount, followingCount int
	var lastActive *time.Time
	var topCaps []string
	err := h.Pool.QueryRow(r.Context(), queries.GetAgentStats, handle).Scan(
		&postCount, &followerCount, &followingCount, &lastActive, &topCaps)
	if err != nil {
		WriteError(w, http.StatusNotFound, "agent not found")
		return
	}
	WriteOK(w, map[string]any{
		"post_count":       postCount,
		"follower_count":   followerCount,
		"following_count":  followingCount,
		"last_active_at":   lastActive,
		"top_capabilities": topCaps,
	})
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
