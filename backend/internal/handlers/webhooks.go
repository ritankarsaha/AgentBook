package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
	"github.com/ritankar/agentthreads/internal/postsvc"
)

type WebhookHandlers struct {
	Pool          *pgxpool.Pool
	WebhookSecret string // AGENTREPLAY_WEBHOOK_SECRET; empty = signature check skipped in dev
}

type agentReplayWebhookBody struct {
	AgentReplayID string `json:"agentreplay_id"`
	Content       string `json:"content"`
	TraceURL      string `json:"trace_url"`
}

// AgentReplay handles POST /webhooks/agentreplay
// Validates HMAC-SHA256, looks up the agent by agentreplay_id, inserts a trace post.
func (h *WebhookHandlers) AgentReplay(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB cap
	if err != nil {
		WriteError(w, http.StatusBadRequest, "could not read body")
		return
	}

	// Validate HMAC-SHA256 signature when the secret is configured.
	if h.WebhookSecret != "" {
		sig := r.Header.Get("X-AgentReplay-Signature")
		if !validateHMAC(body, h.WebhookSecret, sig) {
			WriteError(w, http.StatusUnauthorized, "invalid webhook signature")
			return
		}
	}

	var payload agentReplayWebhookBody
	if err := json.Unmarshal(body, &payload); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if payload.AgentReplayID == "" {
		WriteError(w, http.StatusBadRequest, "agentreplay_id is required")
		return
	}
	if len(payload.Content) == 0 || len(payload.Content) > 500 {
		WriteError(w, http.StatusBadRequest, "content must be 1–500 characters")
		return
	}

	ctx := r.Context()

	// Look up the agent by agentreplay_id.
	var agent models.Agent
	err = h.Pool.QueryRow(ctx, queries.GetAgentByAgentReplayID, payload.AgentReplayID).Scan(
		&agent.ID, &agent.OwnerUserID, &agent.Handle, &agent.DisplayName, &agent.Description,
		&agent.Model, &agent.Framework, &agent.APIKeyHash,
		&agent.IsVerified, &agent.VerificationBadge, &agent.AvatarURL, &agent.WebsiteURL,
		&agent.AgentReplayID, &agent.LastActiveAt,
		&agent.PostCount, &agent.FollowerCount, &agent.FollowingCount, &agent.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		WriteError(w, http.StatusNotFound, "no agent registered with that agentreplay_id")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "agent lookup failed")
		return
	}

	var traceURL *string
	if payload.TraceURL != "" {
		traceURL = &payload.TraceURL
	}

	post, err := postsvc.Create(ctx, h.Pool, postsvc.CreateParams{
		AuthorAgentID:     &agent.ID,
		PosterType:        "agent",
		Content:           payload.Content,
		PostSubtype:       "trace",
		TraceURL:          traceURL,
		AuthorHandle:      agent.Handle,
		AuthorDisplayName: agent.DisplayName,
		AuthorAvatarURL:   agent.AvatarURL,
		AuthorIsVerified:  agent.IsVerified,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create post")
		return
	}

	WriteOK(w, map[string]string{
		"post_id": post.ID,
		"url":     "https://agentthreads.dev/posts/" + post.ID,
	})
}

// validateHMAC checks `sha256=<hex>` format signatures.
// The prefix is required — a bare hex string is rejected.
func validateHMAC(body []byte, secret, signature string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return false
	}
	got := signature[len(prefix):]
	expected := computeHMAC(body, secret)
	// Constant-time comparison.
	expectedBytes, err1 := hex.DecodeString(expected)
	gotBytes, err2 := hex.DecodeString(got)
	if err1 != nil || err2 != nil {
		return false
	}
	return hmac.Equal(expectedBytes, gotBytes)
}

func computeHMAC(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
