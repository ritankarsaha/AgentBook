package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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

func (h *UserHandlers) Sync(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserClaimsFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	ctx := r.Context()

	var existing models.User
	err := h.Pool.QueryRow(ctx, queries.GetUserByID, claims.UserID).Scan(
		&existing.ID, &existing.Email, &existing.Handle, &existing.DisplayName,
		&existing.AvatarURL, &existing.Bio, &existing.IsVerified, &existing.CreatedAt,
	)
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
	err = h.Pool.QueryRow(ctx, queries.UpsertUser, claims.UserID, claims.Email, handle, displayName, avatarURL).Scan(
		&u.ID, &u.Email, &u.Handle, &u.DisplayName, &u.AvatarURL, &u.Bio, &u.IsVerified, &u.CreatedAt,
	)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	WriteJSON(w, http.StatusCreated, Envelope{OK: true, Data: u})
}

// Me implements GET /api/v1/users/me — used by the frontend layout shell
func (h *UserHandlers) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.UserClaimsFromContext(r.Context())
	if claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var u models.User
	err := h.Pool.QueryRow(r.Context(), queries.GetUserByID, claims.UserID).Scan(
		&u.ID, &u.Email, &u.Handle, &u.DisplayName, &u.AvatarURL, &u.Bio, &u.IsVerified, &u.CreatedAt,
	)
	if err != nil {
		WriteError(w, http.StatusNotFound, "user not found — call /api/v1/users/sync first")
		return
	}
	WriteOK(w, u)
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
