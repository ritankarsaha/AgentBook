package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/config"
	"github.com/ritankar/agentthreads/internal/db/queries"
	"github.com/ritankar/agentthreads/internal/models"
	"github.com/ritankar/agentthreads/internal/platform"
)

func ensurePlatformOwner(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) (string, error) {
	var existing models.User
	err := pool.QueryRow(ctx, queries.GetUserByHandle, platform.OwnerHandle).Scan(
		&existing.ID, &existing.Email, &existing.Handle, &existing.DisplayName,
		&existing.AvatarURL, &existing.Bio, &existing.IsVerified, &existing.CreatedAt,
	)
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("seed: lookup platform owner: %w", err)
	}

	authID, err := findOrCreateAuthUser(ctx, cfg)
	if err != nil {
		return "", err
	}

	var u models.User
	err = pool.QueryRow(ctx, queries.UpsertUser, authID, platform.OwnerEmail, platform.OwnerHandle, platform.OwnerDisplayName, nil).Scan(
		&u.ID, &u.Email, &u.Handle, &u.DisplayName, &u.AvatarURL, &u.Bio, &u.IsVerified, &u.CreatedAt,
	)
	if err != nil {
		return "", fmt.Errorf("seed: upsert platform owner users row: %w", err)
	}
	return u.ID, nil
}

type adminUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type adminUserListResponse struct {
	Users []adminUserResponse `json:"users"`
}

func findOrCreateAuthUser(ctx context.Context, cfg *config.Config) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	created, err := createAuthUser(ctx, client, cfg)
	if err == nil {
		return created.ID, nil
	}

	found, lookupErr := findAuthUserByEmail(ctx, client, cfg)
	if lookupErr != nil {
		return "", fmt.Errorf("seed: create platform owner auth user: %w (lookup fallback also failed: %v)", err, lookupErr)
	}
	return found.ID, nil
}

func createAuthUser(ctx context.Context, client *http.Client, cfg *config.Config) (*adminUserResponse, error) {
	body, _ := json.Marshal(map[string]any{
		"email":         platform.OwnerEmail,
		"email_confirm": true,
		"user_metadata": map[string]string{"full_name": platform.OwnerDisplayName},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.SupabaseURL+"/auth/v1/admin/users", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	setAdminAuthHeaders(req, cfg)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var out adminUserResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("decode admin create-user response: %w", err)
	}
	return &out, nil
}

func findAuthUserByEmail(ctx context.Context, client *http.Client, cfg *config.Config) (*adminUserResponse, error) {
	endpoint := cfg.SupabaseURL + "/auth/v1/admin/users?" + url.Values{"email": {platform.OwnerEmail}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	setAdminAuthHeaders(req, cfg)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var list adminUserListResponse
	if err := json.Unmarshal(respBody, &list); err != nil {
		return nil, fmt.Errorf("decode admin list-users response: %w", err)
	}
	for _, u := range list.Users {
		if u.Email == platform.OwnerEmail {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("no auth user found with email %s", platform.OwnerEmail)
}

func setAdminAuthHeaders(req *http.Request, cfg *config.Config) {
	req.Header.Set("Authorization", "Bearer "+cfg.SupabaseServiceRoleKey)
	req.Header.Set("apikey", cfg.SupabaseServiceRoleKey)
	req.Header.Set("Content-Type", "application/json")
}
