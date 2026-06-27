package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
)

type NotificationHandlers struct {
	Pool *pgxpool.Pool
}

type Notification struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	ActorHandle      *string   `json:"actor_handle,omitempty"`
	ActorDisplayName *string   `json:"actor_display_name,omitempty"`
	ActorAvatarURL   *string   `json:"actor_avatar_url,omitempty"`
	PostID           *string   `json:"post_id,omitempty"`
	Read             bool      `json:"read"`
	CreatedAt        time.Time `json:"created_at"`
}

const defaultNotifLimit = 30
const maxNotifLimit = 100

func (h *NotificationHandlers) List(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var recipientUserID, recipientAgentID *string
	if agent != nil {
		recipientAgentID = &agent.ID
	} else {
		recipientUserID = &claims.UserID
	}

	limit := defaultNotifLimit
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 && v <= maxNotifLimit {
			limit = v
		}
	}

	var cursorTime *time.Time
	if cs := r.URL.Query().Get("cursor"); cs != "" {
		if t, err := time.Parse(time.RFC3339Nano, cs); err == nil {
			cursorTime = &t
		}
	}

	ctx := r.Context()

	rows, err := h.Pool.Query(ctx, queries.GetNotifications,
		recipientUserID, recipientAgentID, cursorTime, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not fetch notifications")
		return
	}
	defer rows.Close()

	notifs := []Notification{}
	var lastCreatedAt *time.Time
	for rows.Next() {
		var n Notification
		var actorUserID, actorAgentID *string
		if err := rows.Scan(
			&n.ID, &n.Type, &actorUserID, &actorAgentID, &n.PostID, &n.Read, &n.CreatedAt,
			&n.ActorHandle, &n.ActorDisplayName, &n.ActorAvatarURL,
		); err != nil {
			continue
		}
		notifs = append(notifs, n)
		lastCreatedAt = &n.CreatedAt
	}
	rows.Close()

	markRead := r.URL.Query().Get("mark_read")
	if markRead != "false" && len(notifs) > 0 {
		go func() {
			_, _ = h.Pool.Exec(ctx, queries.MarkNotificationsRead,
				recipientUserID, recipientAgentID)
		}()
	}

	var nextCursor string
	if len(notifs) == limit && lastCreatedAt != nil {
		nextCursor = lastCreatedAt.Format(time.RFC3339Nano)
	}

	WriteOKWithCursor(w, notifs, nextCursor)
}
