package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ritankar/agentthreads/internal/middleware"
	"github.com/ritankar/agentthreads/internal/storage"
)

type MediaHandlers struct {
	Storage *storage.Client
}

var allowedMIMETypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type uploadURLRequest struct {
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
}

type uploadURLResponse struct {
	SignedURL string `json:"signed_url"`
	Path      string `json:"path"`
	ExpiresAt string `json:"expires_at"`
}

// UploadURL handles POST /api/v1/media/upload-url
func (h *MediaHandlers) UploadURL(w http.ResponseWriter, r *http.Request) {
	agent := middleware.AgentFromContext(r.Context())
	claims := middleware.UserClaimsFromContext(r.Context())
	if agent == nil && claims == nil {
		WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req uploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ext, ok := allowedMIMETypes[strings.ToLower(req.ContentType)]
	if !ok {
		WriteError(w, http.StatusBadRequest, "content_type must be image/png, image/jpeg, image/gif, or image/webp")
		return
	}

	// Build a namespaced path: uploads/{identity}/{uuid}{ext}
	var prefix string
	if agent != nil {
		prefix = "agent-" + agent.ID
	} else {
		prefix = "user-" + claims.UserID
	}
	objectPath := path.Join("uploads", prefix, uuid.New().String()+ext)

	result, err := h.Storage.GenerateUploadURL(r.Context(), objectPath)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, fmt.Sprintf("could not generate upload URL: %v", err))
		return
	}

	WriteOK(w, uploadURLResponse{
		SignedURL: result.SignedURL,
		Path:      objectPath,
		ExpiresAt: time.Now().Add(5 * time.Minute).UTC().Format(time.RFC3339),
	})
}
