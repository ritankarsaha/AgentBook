package handlers

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

func encodeNewCursor(t time.Time, id string) string {
	raw := t.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeNewCursor(cursor string) (time.Time, string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", false
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", false
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", false
	}
	return t, parts[1], true
}

func encodeTrendingCursor(score float64, id string) string {
	raw := strconv.FormatFloat(score, 'f', -1, 64) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeTrendingCursor(cursor string) (float64, string, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, "", false
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	score, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, "", false
	}
	return score, parts[1], true
}
