package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
)

type ReactionHandlers struct {
	Pool *pgxpool.Pool
}

func (h *ReactionHandlers) Like(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	postID := chi.URLParam(r, "post_id")
	ctx := r.Context()

	var userID, agentID *string
	reactorType := "agent"
	if agent != nil {
		agentID = &agent.ID
	} else {
		reactorType = "human"
		userID = &claims.UserID
	}

	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, queries.InsertReaction, postID, userID, agentID, reactorType)
	if err != nil {
		if isForeignKeyViolation(err) {
			WriteError(w, http.StatusNotFound, "post not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not like post")
		return
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, queries.IncrementPostLikeCount, postID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update like count")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not commit")
		return
	}
	WriteOK(w, map[string]bool{"liked": true})
}

func (h *ReactionHandlers) Unlike(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	postID := chi.URLParam(r, "post_id")
	ctx := r.Context()

	var userID, agentID *string
	if agent != nil {
		agentID = &agent.ID
	} else {
		userID = &claims.UserID
	}

	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, queries.DeleteReaction, postID, userID, agentID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not unlike post")
		return
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, queries.DecrementPostLikeCount, postID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update like count")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not commit")
		return
	}
	WriteOK(w, map[string]bool{"liked": false})
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
