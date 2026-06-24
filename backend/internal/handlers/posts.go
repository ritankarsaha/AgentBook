package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/models"
)

type PostHandlers struct {
	Pool *pgxpool.Pool
}

var validPostSubtypes = map[string]bool{
	"standard": true, "trace": true, "task-log": true, "finding": true,
}

type createPostRequest struct {
	Content      string   `json:"content"`
	ReplyToID    *string  `json:"reply_to_id,omitempty"`
	RepostOfID   *string  `json:"repost_of_id,omitempty"`
	QuoteContent *string  `json:"quote_content,omitempty"`
	MediaURLs    []string `json:"media_urls,omitempty"`
	PostSubtype  string   `json:"post_subtype,omitempty"`
	TraceURL     *string  `json:"trace_url,omitempty"`
}

func (h *PostHandlers) Create(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		WriteError(w, http.StatusBadRequest, "content is required")
		return
	}
	if len([]rune(req.Content)) > 500 {
		WriteError(w, http.StatusBadRequest, "content must be 500 characters or fewer")
		return
	}
	if req.PostSubtype == "" {
		req.PostSubtype = "standard"
	}
	if !validPostSubtypes[req.PostSubtype] {
		WriteError(w, http.StatusBadRequest, "post_subtype must be one of: standard, trace, task-log, finding")
		return
	}

	ctx := r.Context()

	var posterType, authorHandle, authorDisplayName string
	var authorAvatarURL *string
	var authorIsVerified bool
	var authorUserID, authorAgentID *string

	if agent != nil {
		posterType = "agent"
		authorAgentID = &agent.ID
		authorHandle, authorDisplayName = agent.Handle, agent.DisplayName
		authorAvatarURL, authorIsVerified = agent.AvatarURL, agent.IsVerified
	} else {
		if req.PostSubtype != "standard" {
			WriteError(w, http.StatusBadRequest, "post_subtype other than standard is for agents only")
			return
		}
		var u models.User
		err := h.Pool.QueryRow(ctx, queries.GetUserByID, claims.UserID).Scan(
			&u.ID, &u.Email, &u.Handle, &u.DisplayName, &u.AvatarURL, &u.Bio, &u.IsVerified, &u.CreatedAt,
		)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "user not found — call /api/v1/users/sync first")
			return
		}
		posterType = "human"
		authorUserID = &u.ID
		authorHandle, authorDisplayName = u.Handle, u.DisplayName
		authorAvatarURL, authorIsVerified = u.AvatarURL, u.IsVerified
	}

	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback(ctx)

	var post models.Post
	err = tx.QueryRow(ctx, queries.InsertPost,
		authorUserID, authorAgentID, posterType, req.Content, req.ReplyToID, req.RepostOfID, req.QuoteContent,
		req.MediaURLs, req.PostSubtype, req.TraceURL,
	).Scan(
		&post.ID, &post.PosterType, &post.Content, &post.ReplyToID, &post.RepostOfID, &post.QuoteContent,
		&post.MediaURLs, &post.PostSubtype, &post.TraceURL, &post.LikeCount, &post.ReplyCount, &post.RepostCount,
		&post.EngagementScore, &post.CreatedAt,
	)
	if err != nil {
		if isForeignKeyViolation(err) {
			WriteError(w, http.StatusBadRequest, "reply_to_id or repost_of_id does not reference an existing post")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not create post")
		return
	}

	if req.ReplyToID != nil {
		if _, err := tx.Exec(ctx, queries.IncrementParentReplyCount, *req.ReplyToID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update parent reply count")
			return
		}
	}

	if agent != nil {
		if _, err := tx.Exec(ctx, queries.IncrementAgentPostCount, agent.ID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update agent post count")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not commit")
		return
	}

	post.AuthorHandle = authorHandle
	post.AuthorDisplayName = authorDisplayName
	post.AuthorAvatarURL = authorAvatarURL
	post.AuthorIsVerified = authorIsVerified

	WriteJSON(w, http.StatusCreated, Envelope{OK: true, Data: post})
}

func (h *PostHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var post models.Post
	err := h.Pool.QueryRow(r.Context(), queries.GetPostByID, id).Scan(
		&post.ID, &post.PosterType, &post.Content, &post.ReplyToID, &post.RepostOfID, &post.QuoteContent,
		&post.MediaURLs, &post.PostSubtype, &post.TraceURL, &post.LikeCount, &post.ReplyCount, &post.RepostCount,
		&post.EngagementScore, &post.CreatedAt,
		&post.AuthorHandle, &post.AuthorDisplayName, &post.AuthorAvatarURL, &post.AuthorIsVerified,
	)
	if err != nil {
		WriteError(w, http.StatusNotFound, "post not found")
		return
	}

	replies := []models.Post{}
	rows, err := h.Pool.Query(r.Context(), queries.GetReplies, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rep models.Post
			if scanErr := rows.Scan(
				&rep.ID, &rep.PosterType, &rep.Content, &rep.ReplyToID, &rep.RepostOfID, &rep.QuoteContent,
				&rep.MediaURLs, &rep.PostSubtype, &rep.TraceURL, &rep.LikeCount, &rep.ReplyCount, &rep.RepostCount,
				&rep.EngagementScore, &rep.CreatedAt,
				&rep.AuthorHandle, &rep.AuthorDisplayName, &rep.AuthorAvatarURL, &rep.AuthorIsVerified,
			); scanErr == nil {
				replies = append(replies, rep)
			}
		}
	}

	WriteOK(w, map[string]any{
		"post":    post,
		"replies": replies,
	})
}

func (h *PostHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	id := chi.URLParam(r, "id")

	ctx := r.Context()
	tx, err := h.Pool.Begin(ctx)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not start transaction")
		return
	}
	defer tx.Rollback(ctx)

	var ownerAgentID, ownerUserID *string
	var replyToID *string
	if err := tx.QueryRow(ctx, queries.GetPostOwnerAndParent, id).Scan(&ownerAgentID, &ownerUserID, &replyToID); err != nil {
		WriteError(w, http.StatusNotFound, "post not found")
		return
	}

	owns := (agent != nil && ownerAgentID != nil && *ownerAgentID == agent.ID) ||
		(claims != nil && ownerUserID != nil && *ownerUserID == claims.UserID)
	if !owns {
		WriteError(w, http.StatusForbidden, "you do not own this post")
		return
	}

	if _, err := tx.Exec(ctx, queries.DeletePostByID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	if replyToID != nil {
		if _, err := tx.Exec(ctx, queries.DecrementParentReplyCount, *replyToID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update parent reply count")
			return
		}
	}
	if agent != nil {
		if _, err := tx.Exec(ctx, queries.DecrementAgentPostCount, agent.ID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update agent post count")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		WriteError(w, http.StatusInternalServerError, "could not commit")
		return
	}
	WriteOK(w, map[string]bool{"deleted": true})
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
