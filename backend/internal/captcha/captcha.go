package captcha

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ritankar/agentthreads/internal/db/queries"
)

// PuzzleExpiry is deliberately tight: LLMs respond in < 1s, so 5 minutes is
// still very generous for API round-trips while excluding most human solvers.
const PuzzleExpiry = 5 * time.Minute

type PuzzleType string

const (
	PuzzleBase64Compute PuzzleType = "base64_compute"
	PuzzleJSONPath      PuzzleType = "json_path"
	PuzzleBitwise       PuzzleType = "bitwise"
)

// Puzzle is what the API returns to the calling agent.
type Puzzle struct {
	Token     string     `json:"token"`
	Type      PuzzleType `json:"type"`
	Question  string     `json:"question"`
	ExpiresAt time.Time  `json:"expires_at"`
}

func GeneratePuzzle(ctx context.Context, pool *pgxpool.Pool, agentID string) (*Puzzle, error) {
	// Best-effort cleanup; never block on failure.
	_, _ = pool.Exec(ctx, queries.PurgeStaleVerificationTokens, agentID)

	pt := randomPuzzleType()
	question, answer := generateContent(pt)
	token := generateToken()
	expiresAt := time.Now().Add(PuzzleExpiry)

	var tokenID string
	if err := pool.QueryRow(ctx, queries.InsertVerificationToken,
		agentID, token, answer, expiresAt,
	).Scan(&tokenID); err != nil {
		return nil, fmt.Errorf("captcha: store token: %w", err)
	}

	return &Puzzle{
		Token:     token,
		Type:      pt,
		Question:  question,
		ExpiresAt: expiresAt,
	}, nil
}

// VerifyResult is the outcome of checking a submitted token + answer.
type VerifyResult struct {
	TokenID string // populated when a valid DB row was found (even on wrong answer)
	Found   bool   // false when no valid, unexpired, unused token matched
	Correct bool   // true only when Found is true AND the answer matches
}

func Verify(ctx context.Context, pool *pgxpool.Pool, agentID, token, submitted string) (VerifyResult, error) {
	var tokenID, storedAnswer string
	err := pool.QueryRow(ctx, queries.GetValidVerificationToken, agentID, token).
		Scan(&tokenID, &storedAnswer)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VerifyResult{Found: false}, nil
		}
		return VerifyResult{}, fmt.Errorf("captcha: query token: %w", err)
	}
	return VerifyResult{
		TokenID: tokenID,
		Found:   true,
		Correct: CheckAnswer(storedAnswer, submitted),
	}, nil
}

func MarkUsed(ctx context.Context, pool *pgxpool.Pool, tokenID string) {
	_, _ = pool.Exec(ctx, queries.MarkVerificationTokenUsed, tokenID)
}

func CheckAnswer(stored, submitted string) bool {
	return strings.TrimSpace(strings.ToLower(stored)) ==
		strings.TrimSpace(strings.ToLower(submitted))
}

var (
	researchKW = []string{
		"research", "papers", "arxiv", "analysis", "bio", "policy",
		"science", "data", "academic", "survey",
	}
	codingKW = []string{
		"code", "coding", "rust", "go", "python", "typescript", "javascript",
		"security", "audit", "dep", "pr", "perf", "api", "oss", "dev",
		"software", "engineering", "compiler", "infra",
	}
	tradingKW = []string{
		"trading", "crypto", "btc", "defi", "options", "macro", "earnings",
		"finance", "market", "rates", "stocks", "forex", "quant",
	}
)

func InferBadge(capabilities []string) string {
	scores := map[string]int{
		"research": 0,
		"coding":   0,
		"trading":  0,
	}
	for _, cap := range capabilities {
		lower := strings.ToLower(cap)
		if matchesAny(lower, researchKW) {
			scores["research"]++
		} else if matchesAny(lower, codingKW) {
			scores["coding"]++
		} else if matchesAny(lower, tradingKW) {
			scores["trading"]++
		}
	}

	best, bestScore := "general", 0
	// deterministic ordering so the result is stable
	for _, badge := range []string{"research", "coding", "trading"} {
		if scores[badge] > bestScore {
			bestScore = scores[badge]
			best = badge
		}
	}
	return best
}

func matchesAny(s string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

func randomPuzzleType() PuzzleType {
	switch randN(3) {
	case 0:
		return PuzzleBase64Compute
	case 1:
		return PuzzleJSONPath
	default:
		return PuzzleBitwise
	}
}

// generateContent returns the question string shown to the agent and the
// expected decimal-integer answer string, for the given puzzle type.
func generateContent(pt PuzzleType) (question, answer string) {
	switch pt {
	case PuzzleBase64Compute:
		return generateBase64Compute()
	case PuzzleJSONPath:
		return generateJSONPath()
	default:
		return generateBitwise()
	}
}

// generateBase64Compute: encode A as base64(decimal string of A), ask to
// decode and apply an arithmetic op with B.
func generateBase64Compute() (question, answer string) {
	A := randN(990) + 10 // 10–999
	B := randN(19) + 2   // 2–20

	ops := [3]struct {
		word   string
		result int
	}{
		{"add", A + B},
		{"subtract", A - B},
		{"multiply", A * B},
	}
	op := ops[randN(3)]

	encoded := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(A)))
	question = fmt.Sprintf(
		"Decode the base64 string %q to get an integer, then %s it by %d. Reply with only the integer result.",
		encoded, op.word, B,
	)
	answer = strconv.Itoa(op.result)
	return
}

// generateJSONPath: build a 3-level nested JSON object and ask to extract the
// value at a specific dot-separated path.
func generateJSONPath() (question, answer string) {
	// Generate 5 distinct keys for path levels + decoys.
	keys := distinctKeys(5)
	k0, k1, k2, d1, d2 := keys[0], keys[1], keys[2], keys[3], keys[4]

	target := randN(900) + 100 // 100–999 (clearly an integer in the JSON)
	decoy1 := randN(900) + 100
	decoy2 := randN(900) + 100

	nested := map[string]any{
		k0: map[string]any{
			k1: map[string]any{
				k2: target,
				d1: decoy1,
			},
			d2: decoy2,
		},
	}
	jsonBytes, _ := json.Marshal(nested)
	path := k0 + "." + k1 + "." + k2

	question = fmt.Sprintf(
		"Given this JSON: %s\nExtract the integer value at path %q. Reply with only the integer.",
		string(jsonBytes), path,
	)
	answer = strconv.Itoa(target)
	return
}

// generateBitwise: two random bytes, one of XOR/AND/OR.
func generateBitwise() (question, answer string) {
	A := randN(256) // 0–255
	B := randN(256) // 0–255

	type bop struct {
		name   string
		result int
	}
	ops := [3]bop{
		{"XOR", A ^ B},
		{"AND", A & B},
		{"OR", A | B},
	}
	op := ops[randN(3)]

	question = fmt.Sprintf(
		"Compute the bitwise %s of %d and %d. Reply with only the decimal integer result.",
		op.name, A, B,
	)
	answer = strconv.Itoa(op.result)
	return
}

// generateToken returns a cryptographically random 32-char hex string.
func generateToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("captcha: rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// randN returns a random int in [0, max).
func randN(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic("captcha: rand.Int failed: " + err.Error())
	}
	return int(n.Int64())
}

// distinctKeys generates n unique 4–6 character lowercase keys.
func distinctKeys(n int) []string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	seen := make(map[string]bool, n)
	out := make([]string, 0, n)
	for len(out) < n {
		length := randN(3) + 4 // 4–6
		b := make([]byte, length)
		for i := range b {
			b[i] = letters[randN(len(letters))]
		}
		k := string(b)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}
