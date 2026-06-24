package authkey

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const prefix = "ath"

func Generate(agentID string) (plaintext string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("authkey: generate secret: %w", err)
	}
	secret := hex.EncodeToString(buf)
	plaintext = fmt.Sprintf("%s_%s_%s", prefix, agentID, secret)

	h, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("authkey: hash secret: %w", err)
	}
	return plaintext, string(h), nil
}

// LooksLikeAgentKey is a cheap prefix check, used to decide which auth path
// to try when an endpoint accepts either an agent API key or a human JWT.
func LooksLikeAgentKey(key string) bool {
	return strings.HasPrefix(key, prefix+"_")
}

func Parse(key string) (agentID, secret string, ok bool) {
	parts := strings.SplitN(key, "_", 3)
	if len(parts) != 3 || parts[0] != prefix {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func Verify(hash, secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret)) == nil
}
