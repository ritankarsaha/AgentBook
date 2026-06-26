package activity

import (
	"math/rand"
	"testing"

	"github.com/ritankar/agentthreads/internal/personas"
)

func testAgents() []seedAgent {
	return []seedAgent{
		{ID: "1", Handle: "rust-auditor", Persona: personas.Persona{Handle: "rust-auditor", Capabilities: []string{"rust", "security"}}},
		{ID: "2", Handle: "dep-scanner", Persona: personas.Persona{Handle: "dep-scanner", Capabilities: []string{"security", "dependencies"}}},
		{ID: "3", Handle: "btc-signal", Persona: personas.Persona{Handle: "btc-signal", Capabilities: []string{"bitcoin", "on-chain"}}},
	}
}

func newTestResponder() *Responder {
	return &Responder{
		rng:    rand.New(rand.NewSource(1)),
		agents: testAgents(),
	}
}

func TestSelectResponders_MentionTakesPriority(t *testing.T) {
	r := newTestResponder()
	got := r.selectResponders("hey @rust-auditor what do you think about this bitcoin thing")
	if len(got) != 1 || got[0].Handle != "rust-auditor" {
		t.Fatalf("expected only the mentioned agent, got %+v", got)
	}
}

func TestSelectResponders_CapabilityMatchWhenNoMention(t *testing.T) {
	r := newTestResponder()
	got := r.selectResponders("anyone know if this crate has a security issue?")
	found := false
	for _, a := range got {
		if a.Handle == "rust-auditor" || a.Handle == "dep-scanner" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a security-capable agent matched, got %+v", got)
	}
}

func TestSelectResponders_FallsBackToRandomWhenNoMatch(t *testing.T) {
	r := newTestResponder()
	got := r.selectResponders("what's everyone's favorite color")
	if len(got) != 1 {
		t.Fatalf("expected exactly one fallback agent, got %d", len(got))
	}
}

func TestSelectResponders_NeverExceedsCap(t *testing.T) {
	r := newTestResponder()
	// Matches all three agents' capabilities at once.
	got := r.selectResponders("this rust crate has a security flaw involving dependencies and bitcoin on-chain data")
	if len(got) > maxRespondersPerPost {
		t.Fatalf("expected at most %d responders, got %d", maxRespondersPerPost, len(got))
	}
}

func TestSelectResponders_CapabilityMatchRespectsWordBoundaries(t *testing.T) {
	// Real bug found live: macro-clock's capability "rates" matched as a
	// substring of "crates", pulling in an unrelated trading/Fed agent for
	// a Rust-security question.
	agents := []seedAgent{
		{ID: "1", Handle: "macro-clock", Persona: personas.Persona{Handle: "macro-clock", Capabilities: []string{"trading", "macro", "fed", "rates"}}},
		{ID: "2", Handle: "rust-auditor", Persona: personas.Persona{Handle: "rust-auditor", Capabilities: []string{"rust", "security"}}},
	}
	r := &Responder{rng: rand.New(rand.NewSource(1)), agents: agents}
	got := r.selectResponders("anyone seen any nasty rust security issues in popular crates lately?")
	for _, a := range got {
		if a.Handle == "macro-clock" {
			t.Fatalf("expected macro-clock NOT matched (false positive via 'crates' containing 'rates'), got %+v", got)
		}
	}
}

func TestContainsWord_RejectsSubstringMatch(t *testing.T) {
	if containsWord("popular crates lately", "rates") {
		t.Error("expected 'rates' to NOT match inside 'crates'")
	}
	if !containsWord("the fed cut rates today", "rates") {
		t.Error("expected 'rates' to match as a standalone word")
	}
	if !containsWord("rust security review", "rust") {
		t.Error("expected exact leading word match")
	}
}

func TestChooseResponseAction_BoundariesMapToExpectedActions(t *testing.T) {
	cases := []struct {
		roll float64
		want responseAction
	}{
		{0, responseReply},
		{responseReplyProbability - 0.001, responseReply},
		{responseReplyProbability, responseQuoteRepost},
		{responseReplyProbability + responseQuoteRepostProbability, responseRepost},
		{responseReplyProbability + responseQuoteRepostProbability + responseRepostProbability, responseLike},
		{responseReplyProbability + responseQuoteRepostProbability + responseRepostProbability + responseLikeProbability, responseNothing},
		{0.999, responseNothing},
	}
	for _, c := range cases {
		if got := chooseResponseAction(c.roll); got != c.want {
			t.Errorf("chooseResponseAction(%v) = %v, want %v", c.roll, got, c.want)
		}
	}
}

func TestChooseResponseAction_ProbabilityMassNeverExceedsOne(t *testing.T) {
	total := responseReplyProbability + responseQuoteRepostProbability + responseRepostProbability + responseLikeProbability
	if total > 1.0 {
		t.Fatalf("response action probabilities sum to %v, must leave room for 'do nothing'", total)
	}
}

func TestDedupeAgents_RemovesDuplicateIDs(t *testing.T) {
	agents := testAgents()
	dup := []seedAgent{agents[0], agents[1], agents[0]}
	got := dedupeAgents(dup)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique agents, got %d", len(got))
	}
}
