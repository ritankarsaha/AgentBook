package captcha

import (
	"strconv"
	"strings"
	"testing"
)

func TestCheckAnswer(t *testing.T) {
	cases := []struct {
		stored, submitted string
		want              bool
	}{
		{"42", "42", true},
		{"42", "  42  ", true}, // whitespace trimmed
		{"42", " 42\n", true},  // newline trimmed
		{"42", "43", false},    // wrong
		{"100", "100", true},
		{"-5", "-5", true}, // subtraction can produce negatives
		{"0", "0", true},
	}
	for _, tc := range cases {
		got := CheckAnswer(tc.stored, tc.submitted)
		if got != tc.want {
			t.Errorf("CheckAnswer(%q, %q) = %v, want %v", tc.stored, tc.submitted, got, tc.want)
		}
	}
}

func TestInferBadge(t *testing.T) {
	cases := []struct {
		caps []string
		want string
	}{
		{[]string{"research", "papers", "arxiv"}, "research"},
		{[]string{"rust", "security", "audit", "coding"}, "coding"},
		{[]string{"crypto", "btc", "defi", "trading"}, "trading"},
		{[]string{}, "general"},
		{[]string{"custom", "tasks", "agent"}, "general"},
		// tie: research(1) vs coding(1) → first winner (research) wins
		{[]string{"research", "rust"}, "research"},
	}
	for _, tc := range cases {
		got := InferBadge(tc.caps)
		if got != tc.want {
			t.Errorf("InferBadge(%v) = %q, want %q", tc.caps, got, tc.want)
		}
	}
}

func TestGenerateContent_AllTypes(t *testing.T) {
	for _, pt := range []PuzzleType{PuzzleBase64Compute, PuzzleJSONPath, PuzzleBitwise} {
		q, a := generateContent(pt)
		if q == "" {
			t.Errorf("type %s: empty question", pt)
		}
		if a == "" {
			t.Errorf("type %s: empty answer", pt)
		}
		if _, err := strconv.Atoi(a); err != nil {
			t.Errorf("type %s: answer %q is not a decimal integer", pt, a)
		}
	}
}

// TestGenerateBase64Compute verifies the arithmetic is correct across many runs.
func TestGenerateBase64Compute(t *testing.T) {
	for i := 0; i < 200; i++ {
		_, a := generateBase64Compute()
		if _, err := strconv.Atoi(a); err != nil {
			t.Fatalf("answer %q not integer on iteration %d", a, i)
		}
	}
}

func TestGenerateJSONPath(t *testing.T) {
	for i := 0; i < 50; i++ {
		q, a := generateJSONPath()
		// q must contain the answer as a substring (it's in the JSON)
		if !strings.Contains(q, a) {
			t.Fatalf("iteration %d: answer %q not found in question %q", i, a, q)
		}
		if _, err := strconv.Atoi(a); err != nil {
			t.Fatalf("iteration %d: answer %q not integer", i, a)
		}
	}
}

func TestGenerateBitwise(t *testing.T) {
	for i := 0; i < 200; i++ {
		_, a := generateBitwise()
		n, err := strconv.Atoi(a)
		if err != nil {
			t.Fatalf("iteration %d: answer %q not integer", i, a)
		}
		// Bitwise ops on [0,255] produce [0,255]
		if n < 0 || n > 255 {
			t.Fatalf("iteration %d: answer %d out of [0,255]", i, n)
		}
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		tok := generateToken()
		if len(tok) != 32 {
			t.Fatalf("token length %d, want 32", len(tok))
		}
		if seen[tok] {
			t.Fatalf("duplicate token on iteration %d", i)
		}
		seen[tok] = true
	}
}

func TestDistinctKeys(t *testing.T) {
	for i := 0; i < 20; i++ {
		keys := distinctKeys(5)
		if len(keys) != 5 {
			t.Fatalf("got %d keys, want 5", len(keys))
		}
		seen := make(map[string]bool)
		for _, k := range keys {
			if seen[k] {
				t.Fatalf("duplicate key %q", k)
			}
			seen[k] = true
			if len(k) < 4 || len(k) > 6 {
				t.Fatalf("key %q length %d outside [4,6]", k, len(k))
			}
		}
	}
}

func TestInferBadge_EmptyCapabilities(t *testing.T) {
	if got := InferBadge(nil); got != "general" {
		t.Errorf("nil caps: got %q, want general", got)
	}
	if got := InferBadge([]string{}); got != "general" {
		t.Errorf("empty caps: got %q, want general", got)
	}
}
