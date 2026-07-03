package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/models"
	"github.com/ritankar/agentthreads/internal/postsvc"
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
	if req.ReplyToID != nil && req.RepostOfID != nil {
		WriteError(w, http.StatusBadRequest, "a post cannot be both a reply and a repost")
		return
	}
	// A repost (plain or quote) carries its own text in quote_content, not
	// content — content is only required for standalone posts and replies.
	if req.Content == "" && req.RepostOfID == nil {
		WriteError(w, http.StatusBadRequest, "content is required")
		return
	}
	if len([]rune(req.Content)) > 500 {
		WriteError(w, http.StatusBadRequest, "content must be 500 characters or fewer")
		return
	}
	if req.QuoteContent != nil {
		if req.RepostOfID == nil {
			WriteError(w, http.StatusBadRequest, "quote_content requires repost_of_id")
			return
		}
		if len([]rune(*req.QuoteContent)) > 500 {
			WriteError(w, http.StatusBadRequest, "quote_content must be 500 characters or fewer")
			return
		}
	}
	if req.PostSubtype == "" {
		req.PostSubtype = "standard"
	}
	if !validPostSubtypes[req.PostSubtype] {
		WriteError(w, http.StatusBadRequest, "post_subtype must be one of: standard, trace, task-log, finding")
		return
	}

	ctx := r.Context()

	params := postsvc.CreateParams{
		Content:      req.Content,
		ReplyToID:    req.ReplyToID,
		RepostOfID:   req.RepostOfID,
		QuoteContent: req.QuoteContent,
		MediaURLs:    req.MediaURLs,
		PostSubtype:  req.PostSubtype,
		TraceURL:     req.TraceURL,
	}

	if agent != nil {
		params.PosterType = "agent"
		params.AuthorAgentID = &agent.ID
		params.AuthorHandle, params.AuthorDisplayName = agent.Handle, agent.DisplayName
		params.AuthorAvatarURL, params.AuthorIsVerified = agent.AvatarURL, agent.IsVerified
	} else {
		if req.PostSubtype != "standard" {
			WriteError(w, http.StatusBadRequest, "post_subtype other than standard is for agents only")
			return
		}
		var u models.User
		if err := scanUser(h.Pool.QueryRow(ctx, queries.GetUserByID, claims.UserID), &u); err != nil {
			WriteError(w, http.StatusUnauthorized, "user not found — call /api/v1/users/sync first")
			return
		}
		params.PosterType = "human"
		params.AuthorUserID = &u.ID
		params.AuthorHandle, params.AuthorDisplayName = u.Handle, u.DisplayName
		params.AuthorAvatarURL, params.AuthorIsVerified = u.AvatarURL, u.IsVerified
	}

	post, err := postsvc.Create(ctx, h.Pool, params)
	if err != nil {
		if errors.Is(err, postsvc.ErrInvalidParent) {
			WriteError(w, http.StatusBadRequest, "reply_to_id or repost_of_id does not reference an existing post")
			return
		}
		WriteError(w, http.StatusInternalServerError, "could not create post")
		return
	}

	WriteJSON(w, http.StatusCreated, Envelope{OK: true, Data: post})
}

func (h *PostHandlers) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var post models.Post
	var postRef refPostFields
	postDest := append([]any{
		&post.ID, &post.PosterType, &post.Content, &post.ReplyToID, &post.RepostOfID, &post.QuoteContent,
		&post.MediaURLs, &post.PostSubtype, &post.TraceURL, &post.LikeCount, &post.ReplyCount, &post.RepostCount,
		&post.EngagementScore, &post.CreatedAt,
		&post.AuthorHandle, &post.AuthorDisplayName, &post.AuthorAvatarURL, &post.AuthorIsVerified,
	}, postRef.dest()...)
	if err := h.Pool.QueryRow(r.Context(), queries.GetPostByID, id).Scan(postDest...); err != nil {
		WriteError(w, http.StatusNotFound, "post not found")
		return
	}
	post.ReferencedPost = postRef.toPost()

	replies := []models.Post{}
	rows, err := h.Pool.Query(r.Context(), queries.GetReplies, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rep models.Post
			var repRef refPostFields
			repDest := append([]any{
				&rep.ID, &rep.PosterType, &rep.Content, &rep.ReplyToID, &rep.RepostOfID, &rep.QuoteContent,
				&rep.MediaURLs, &rep.PostSubtype, &rep.TraceURL, &rep.LikeCount, &rep.ReplyCount, &rep.RepostCount,
				&rep.EngagementScore, &rep.CreatedAt,
				&rep.AuthorHandle, &rep.AuthorDisplayName, &rep.AuthorAvatarURL, &rep.AuthorIsVerified,
			}, repRef.dest()...)
			if scanErr := rows.Scan(repDest...); scanErr == nil {
				rep.ReferencedPost = repRef.toPost()
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
	var replyToID, repostOfID *string
	if err := tx.QueryRow(ctx, queries.GetPostOwnerAndParent, id).Scan(&ownerAgentID, &ownerUserID, &replyToID, &repostOfID); err != nil {
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
	if repostOfID != nil {
		if _, err := tx.Exec(ctx, queries.DecrementParentRepostCount, *repostOfID); err != nil {
			WriteError(w, http.StatusInternalServerError, "could not update parent repost count")
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
