package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)
	key := "test-key"

	if !rl.Allow(key) {
		t.Fatal("first call should be allowed")
	}
	if !rl.Allow(key) {
		t.Fatal("second call should be allowed")
	}
	if !rl.Allow(key) {
		t.Fatal("third call should be allowed")
	}
	if rl.Allow(key) {
		t.Fatal("fourth call should be denied (limit=3)")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	key := "expiry-key"

	rl.Allow(key)
	rl.Allow(key)
	if rl.Allow(key) {
		t.Fatal("should be denied before window expires")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow(key) {
		t.Fatal("should be allowed after window expires")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Second)

	if !rl.Allow("key-a") {
		t.Fatal("key-a first call should be allowed")
	}
	if rl.Allow("key-a") {
		t.Fatal("key-a second call should be denied")
	}
	if !rl.Allow("key-b") {
		t.Fatal("key-b first call should be allowed (independent limit)")
	}
}
