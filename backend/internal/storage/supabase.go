package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client calls Supabase Storage REST API with the service role key.
type Client struct {
	supabaseURL    string
	serviceRoleKey string
	bucket         string
	httpClient     *http.Client
}

func NewClient(supabaseURL, serviceRoleKey, bucket string) *Client {
	return &Client{
		supabaseURL:    strings.TrimRight(supabaseURL, "/"),
		serviceRoleKey: serviceRoleKey,
		bucket:         bucket,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

type UploadURL struct {
	SignedURL string `json:"signed_url"`
	Token     string `json:"token"`
	Path      string `json:"path"`
}

// GenerateUploadURL creates a presigned upload URL for the given object path.
// The caller then PUTs the file body directly to SignedURL.
func (c *Client) GenerateUploadURL(ctx context.Context, objectPath string) (*UploadURL, error) {
	endpoint := fmt.Sprintf("%s/storage/v1/object/sign/upload/%s/%s",
		c.supabaseURL, c.bucket, strings.TrimLeft(objectPath, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.serviceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("storage api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("storage api returned %d", resp.StatusCode)
	}

	var result UploadURL
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
