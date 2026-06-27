package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/models"
)

type UserHandlers struct {
	Pool *pgxpool.Pool
}

var handleSanitizeRe = regexp.MustCompile(`[^a-z0-9_-]`)

func scanUser(row interface{ Scan(...any) error }, u *models.User) error {
	return row.Scan(
		&u.ID, &u.Email, &u.Handle, &u.DisplayName,
		&u.AvatarURL, &u.Bio, &u.WebsiteURL, &u.IsVerified,
		&u.FollowerCount, &u.FollowingCount, &u.PostCount,
		&u.CreatedAt,
	)
}

func (h *UserHandlers) Sync(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserClaimsFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	ctx := r.Context()

	var existing models.User
	err := scanUser(h.Pool.QueryRow(ctx, queries.GetUserByID, claims.UserID), &existing)
	if err == nil {
		WriteOK(w, existing)
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		WriteError(w, http.StatusInternalServerError, "lookup failed")
		return
	}

	displayName := claims.FullName
	if displayName == "" {
		displayName = strings.SplitN(claims.Email, "@", 2)[0]
	}

	handle, err := h.uniqueHandle(ctx, claims.Email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not allocate handle")
		return
	}

	var avatarURL *string
	if claims.AvatarURL != "" {
		avatarURL = &claims.AvatarURL
	}

	var u models.User
	err = scanUser(
		h.Pool.QueryRow(ctx, queries.UpsertUser, claims.UserID, claims.Email, handle, displayName, avatarURL),
		&u,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	WriteJSON(w, http.StatusCreated, Envelope{OK: true, Data: u})
}

func (h *UserHandlers) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserClaimsFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var u models.User
	err := scanUser(h.Pool.QueryRow(r.Context(), queries.GetUserByID, claims.UserID), &u)
	if err != nil {
		WriteError(w, http.StatusNotFound, "user not found — call /api/v1/users/sync first")
		return
	}
	WriteOK(w, u)
}

type updateUserRequest struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	WebsiteURL  *string `json:"website_url"`
}

func (h *UserHandlers) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserClaimsFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	displayName := ""
	if req.DisplayName != nil {
		displayName = strings.TrimSpace(*req.DisplayName)
	}

	var u models.User
	err := scanUser(
		h.Pool.QueryRow(r.Context(), queries.UpdateUserProfile,
			claims.UserID, displayName, req.Bio, req.WebsiteURL),
		&u,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	u.Email = ""
	WriteOK(w, u)
}

func (h *UserHandlers) UserByHandle(w http.ResponseWriter, r *http.Request) {
	handle := chi.URLParam(r, "handle")
	var u models.User
	err := scanUser(h.Pool.QueryRow(r.Context(), queries.GetUserByHandle, handle), &u)
	if err != nil {
		WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	u.Email = ""
	WriteOK(w, u)
}

func (h *UserHandlers) UserPosts(w http.ResponseWriter, r *http.Request) {
	handle := chi.URLParam(r, "handle")
	tab := r.URL.Query().Get("tab")
	if tab != "posts" && tab != "replies" && tab != "traces" {
		tab = "posts"
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	var createdAtCursor *time.Time
	var idCursor *string
	if c := r.URL.Query().Get("cursor"); c != "" {
		if t, id, ok := decodeNewCursor(c); ok {
			createdAtCursor = &t
			idCursor = &id
		}
	}

	rows, err := h.Pool.Query(r.Context(), queries.GetPostsByUserHandle,
		handle, tab, createdAtCursor, idCursor, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query failed")
		return
	}
	defer rows.Close()

	posts, nextCursor, err := scanFeedRows(rows)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "scan failed")
		return
	}
	WriteOKWithCursor(w, posts, nextCursor)
}

func (h *UserHandlers) uniqueHandle(ctx context.Context, email string) (string, error) {
	base := strings.ToLower(strings.SplitN(email, "@", 2)[0])
	base = handleSanitizeRe.ReplaceAllString(base, "")
	if len(base) < 3 {
		base = base + "user"
	}
	if len(base) > 26 {
		base = base[:26]
	}

	candidate := base
	for attempt := 0; attempt < 50; attempt++ {
		if attempt > 0 {
			candidate = fmt.Sprintf("%s%d", base, attempt)
		}

		var reserved bool
		if err := h.Pool.QueryRow(ctx, queries.IsHandleReserved, candidate).Scan(&reserved); err != nil {
			return "", err
		}
		if reserved {
			continue
		}

		var taken bool
		if err := h.Pool.QueryRow(ctx, queries.IsHandleTaken, candidate).Scan(&taken); err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("exhausted handle attempts for base %q", base)
}
