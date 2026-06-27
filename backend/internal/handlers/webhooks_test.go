package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestComputeHMAC(t *testing.T) {
	body := []byte(`{"agentreplay_id":"test","content":"hello"}`)
	secret := "s3cr3t"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))

	got := computeHMAC(body, secret)
	if got != want {
		t.Fatalf("computeHMAC = %q; want %q", got, want)
	}
}

func TestValidateHMAC(t *testing.T) {
	body := []byte(`{"agentreplay_id":"test","content":"hello"}`)
	secret := "s3cr3t"
	sig := "sha256=" + computeHMAC(body, secret)

	if !validateHMAC(body, secret, sig) {
		t.Fatal("validateHMAC returned false for a valid signature")
	}

	// Wrong secret.
	if validateHMAC(body, "wrong", sig) {
		t.Fatal("validateHMAC returned true for a wrong secret")
	}

	// Tampered body.
	if validateHMAC([]byte(`{"agentreplay_id":"evil"}`), secret, sig) {
		t.Fatal("validateHMAC returned true for tampered body")
	}

	// Missing prefix — should fail.
	if validateHMAC(body, secret, computeHMAC(body, secret)) {
		t.Fatal("validateHMAC returned true without sha256= prefix")
	}

	// Garbage signature.
	if validateHMAC(body, secret, "sha256=notahex!") {
		t.Fatal("validateHMAC returned true for garbage signature")
	}
}
