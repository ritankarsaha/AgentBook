package nim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func fakeKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("nvapi-fake-key-%02d", i)
	}
	return keys
}

func writeChatResponse(w http.ResponseWriter, content string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{
			{"message": map[string]string{"role": "assistant", "content": content}},
		},
	})
}

func keyFromAuth(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}

func TestComplete_DistributesEvenlyAcrossKeys(t *testing.T) {
	const numKeys = 14
	const callsPerKey = 5

	var mu sync.Mutex
	counts := map[string]int{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		counts[keyFromAuth(r)]++
		mu.Unlock()
		writeChatResponse(w, "ok")
	}))
	defer srv.Close()

	client, err := New(fakeKeys(numKeys), srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for i := 0; i < numKeys*callsPerKey; i++ {
		if _, err := client.Complete(context.Background(), "test-model", "sys", "user", CompleteOptions{}); err != nil {
			t.Fatalf("Complete call %d: %v", i, err)
		}
	}

	if len(counts) != numKeys {
		t.Fatalf("expected all %d keys to be used at least once, got %d distinct keys", numKeys, len(counts))
	}
	for key, n := range counts {
		if n != callsPerKey {
			t.Errorf("key %s used %d times, want exactly %d", key, n, callsPerKey)
		}
	}
}

func TestComplete_RetriesNextKeyOn429(t *testing.T) {
	keys := fakeKeys(5)
	rateLimitedKey := keys[2]

	var rateLimitedHits, successHits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if keyFromAuth(r) == rateLimitedKey {
			rateLimitedHits.Add(1)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		successHits.Add(1)
		writeChatResponse(w, "ok")
	}))
	defer srv.Close()

	client, err := New(keys, srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Force the round-robin to land on the rate-limited key first.
	client.next.Store(2)

	content, err := client.Complete(context.Background(), "test-model", "sys", "user", CompleteOptions{})
	if err != nil {
		t.Fatalf("Complete: unexpected error: %v", err)
	}
	if content != "ok" {
		t.Fatalf("content = %q, want %q", content, "ok")
	}
	if rateLimitedHits.Load() != 1 {
		t.Errorf("rate-limited key was hit %d times, want exactly 1 (no retrying the same key)", rateLimitedHits.Load())
	}
	if successHits.Load() != 1 {
		t.Errorf("expected exactly 1 successful attempt after the 429, got %d", successHits.Load())
	}
}

// Confirmed against the real NVIDIA NIM API
func TestComplete_RetriesNextKeyOn403(t *testing.T) {
	keys := fakeKeys(5)
	forbiddenKey := keys[1]

	var forbiddenHits, successHits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if keyFromAuth(r) == forbiddenKey {
			forbiddenHits.Add(1)
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"status":403,"title":"Forbidden","detail":"Authorization failed"}`))
			return
		}
		successHits.Add(1)
		writeChatResponse(w, "ok")
	}))
	defer srv.Close()

	client, err := New(keys, srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	client.next.Store(1)

	content, err := client.Complete(context.Background(), "test-model", "sys", "user", CompleteOptions{})
	if err != nil {
		t.Fatalf("Complete: unexpected error: %v", err)
	}
	if content != "ok" {
		t.Fatalf("content = %q, want %q", content, "ok")
	}
	if forbiddenHits.Load() != 1 {
		t.Errorf("forbidden key was hit %d times, want exactly 1", forbiddenHits.Load())
	}
	if successHits.Load() != 1 {
		t.Errorf("expected exactly 1 successful attempt after the 403, got %d", successHits.Load())
	}
}

func TestComplete_AllKeysExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client, err := New(fakeKeys(3), srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Complete(context.Background(), "test-model", "sys", "user", CompleteOptions{})
	if err == nil {
		t.Fatal("expected an error when every key is rate-limited, got nil")
	}
	if !strings.Contains(err.Error(), "all 3 keys exhausted") {
		t.Errorf("error = %q, want it to mention all keys exhausted", err.Error())
	}
}

func TestComplete_NonRetryableFailsFast(t *testing.T) {
	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	client, err := New(fakeKeys(5), srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.Complete(context.Background(), "test-model", "sys", "user", CompleteOptions{})
	if err == nil {
		t.Fatal("expected an error for a 400 response, got nil")
	}
	if hits.Load() != 1 {
		t.Errorf("expected exactly 1 request (no retrying a non-retryable 400 against other keys), got %d", hits.Load())
	}
}

func TestNew_RejectsEmptyKeyPool(t *testing.T) {
	if _, err := New(nil, "https://example.invalid"); err == nil {
		t.Fatal("expected an error constructing a client with zero keys, got nil")
	}
}

func TestMaskKey_NeverLeaksFullKey(t *testing.T) {
	key := "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"
	masked := maskKey(key)
	if strings.Contains(masked, key) || len(masked) >= len(key) {
		t.Errorf("maskKey(%q) = %q, leaks the full key", key, masked)
	}
}
