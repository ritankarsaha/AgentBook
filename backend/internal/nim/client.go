
package nim

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type Client struct {
	keys    []string
	next    atomic.Uint64
	baseURL string
	http    *http.Client
}

// A non-empty key pool is enforced here rather
// than discovered on the first call
func New(keys []string, baseURL string) (*Client, error) {
	if len(keys) == 0 {
		return nil, errors.New("nim: at least one API key is required")
	}
	return &Client{
		keys:    keys,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 90 * time.Second},
	}, nil
}

// PoolSize reports how many keys are in rotation — useful for callers to
// log at startup ("seed script running against a pool of N keys").
func (c *Client) PoolSize() int { return len(c.keys) }

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompleteOptions are optional generation parameters. Zero values mean
// "let NIM use its own default" rather than forcing one on every caller.
type CompleteOptions struct {
	MaxTokens   int
	Temperature float64
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func (c *Client) Complete(ctx context.Context, model, systemPrompt, userPrompt string, opts CompleteOptions) (string, error) {
	payload, err := json.Marshal(chatRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	})
	if err != nil {
		return "", fmt.Errorf("nim: encode request: %w", err)
	}

	poolSize := uint64(len(c.keys))
	var lastErr error
	for attempt := uint64(0); attempt < poolSize; attempt++ {
		idx := (c.next.Add(1) - 1) % poolSize
		key := c.keys[idx]

		content, retryable, err := c.attempt(ctx, key, payload)
		if err == nil {
			return content, nil
		}
		lastErr = err
		if !retryable {
			return "", err
		}
	}
	return "", fmt.Errorf("nim: all %d keys exhausted, last error: %w", poolSize, lastErr)
}

// attempt makes one HTTP call with one key. retryable reports whether
// another key might succeed where this one failed.
func (c *Client) attempt(ctx context.Context, key string, payload []byte) (content string, retryable bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", false, fmt.Errorf("nim: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", true, fmt.Errorf("nim: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 16<<10))

	if resp.StatusCode != http.StatusOK {
		return "", isRetryableStatus(resp.StatusCode), fmt.Errorf("nim: %s: status %d: %s", maskKey(key), resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", false, fmt.Errorf("nim: decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", true, errors.New("nim: empty choices in response")
	}
	return parsed.Choices[0].Message.Content, false, nil
}

func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests:
		return true
	default:
		return code >= 500
	}
}

func maskKey(key string) string {
	if len(key) <= 10 {
		return "***"
	}
	return key[:6] + "***"
}
