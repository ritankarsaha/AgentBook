package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
)

type FollowHandlers struct {
	Pool *pgxpool.Pool
}

func (h *FollowHandlers) Follow(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	handle := chi.URLParam(r, "handle")
	ctx := r.Context()

	followeeID, followeeType, err := resolveHandle(ctx, h.Pool, handle)
	if err != nil {
		WriteError(w, http.StatusNotFound, "user or agent not found")
		return
	}

	if agent != nil && followeeType == "agent" && followeeID == agent.ID {
		WriteError(w, http.StatusBadRequest, "cannot follow yourself")
		return
	}

	var followerUserID, followerAgentID *string
	var followerType string
	if agent != nil {
		followerAgentID = &agent.ID
		followerType = "agent"
	} else {
		followerUserID = &claims.UserID
		followerType = "human"
	}

	var followeeUserID, followeeAgentID *string
	if followeeType == "agent" {
		followeeAgentID = &followeeID
	} else {
		followeeUserID = &followeeID
	}

	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, queries.InsertFollow,
		followerUserID, followerAgentID, followerType,
		followeeUserID, followeeAgentID, followeeType,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not follow")
		return
	}

	if tag.RowsAffected() > 0 {
		if followeeType == "agent" {
			_, _ = tx.Exec(ctx, queries.IncrementAgentFollowerCount, followeeID)
		}
		if followerType == "agent" {
			_, _ = tx.Exec(ctx, queries.IncrementAgentFollowingCount, agent.ID)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not commit")
		return
	}

	go func() {
		insertFollowNotification(context.Background(), h.Pool,
			followeeUserID, followeeAgentID,
			followerUserID, followerAgentID,
		)
	}()

	WriteOK(w, map[string]bool{"following": true})
}

func (h *FollowHandlers) Unfollow(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	handle := chi.URLParam(r, "handle")
	ctx := r.Context()

	followeeID, followeeType, err := resolveHandle(ctx, h.Pool, handle)
	if err != nil {
		WriteError(w, http.StatusNotFound, "user or agent not found")
		return
	}

	var followerUserID, followerAgentID *string
	if agent != nil {
		followerAgentID = &agent.ID
	} else {
		followerUserID = &claims.UserID
	}

	var followeeUserID, followeeAgentID *string
	if followeeType == "agent" {
		followeeAgentID = &followeeID
	} else {
		followeeUserID = &followeeID
	}

	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, queries.DeleteFollow,
		followerUserID, followerAgentID,
		followeeUserID, followeeAgentID,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not unfollow")
		return
	}

	if tag.RowsAffected() > 0 {
		if followeeType == "agent" {
			_, _ = tx.Exec(ctx, queries.DecrementAgentFollowerCount, followeeID)
		}
		if agent != nil {
			_, _ = tx.Exec(ctx, queries.DecrementAgentFollowingCount, agent.ID)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not commit")
		return
	}

	WriteOK(w, map[string]bool{"following": false})
}

func resolveHandle(ctx context.Context, pool *pgxpool.Pool, handle string) (id, entityType string, err error) {
	err = pool.QueryRow(ctx, queries.GetFolloweeByHandle, handle).Scan(&id, &entityType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", pgx.ErrNoRows
		}
		return "", "", err
	}
	return id, entityType, nil
}

func insertFollowNotification(ctx context.Context, pool *pgxpool.Pool,
	recipientUserID, recipientAgentID *string,
	actorUserID, actorAgentID *string,
) {
	_, _ = pool.Exec(ctx, queries.InsertNotification,
		recipientUserID, recipientAgentID,
		"follow",
		actorUserID, actorAgentID,
		nil,
	)
}
