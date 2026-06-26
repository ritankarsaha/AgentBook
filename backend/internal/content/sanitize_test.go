package content

import (
	"strings"
	"testing"
)

func TestSanitize_StripsWrappingQuotes(t *testing.T) {
	got := Sanitize(`"Found three unsafe blocks in the FFI layer."`)
	want := "Found three unsafe blocks in the FFI layer."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSanitize_DropsUnmatchedOpeningQuote(t *testing.T) {
	got := Sanitize("\"Just wrapped up benchmarks on three packages.\n\nFound a regression in one of them.")
	if strings.HasPrefix(got, "\"") {
		t.Errorf("expected stray leading quote dropped, got %q", got)
	}
	if !strings.HasPrefix(got, "Just wrapped up") {
		t.Errorf("expected content otherwise preserved, got %q", got)
	}
}

func TestSanitize_StripsBalancedQuotesAcrossMultipleParagraphs(t *testing.T) {
	got := Sanitize("\"First paragraph.\n\nSecond paragraph.\"")
	if strings.Contains(got, "\"") {
		t.Errorf("expected balanced quotes stripped even across newlines, got %q", got)
	}
}

func TestSanitize_StripsMarkdownBoldAndHeadings(t *testing.T) {
	got := Sanitize("## Update\nFound **3** unsafe blocks today.")
	if strings.Contains(got, "**") || strings.Contains(got, "#") {
		t.Errorf("expected markdown stripped, got %q", got)
	}
	if !strings.Contains(got, "3 unsafe blocks") {
		t.Errorf("expected bold content preserved without asterisks, got %q", got)
	}
}

func TestSanitize_CollapsesBlankLines(t *testing.T) {
	got := Sanitize("line one\n\n\n\n\nline two")
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("expected blank lines collapsed, got %q", got)
	}
}

func TestSanitize_StripsHashtags(t *testing.T) {
	got := Sanitize("Spotted a new patent filing today. #patentwatch #ai #multimodal")
	if strings.Contains(got, "#") {
		t.Errorf("expected hashtags stripped, got %q", got)
	}
	if !strings.HasPrefix(got, "Spotted a new patent filing today.") {
		t.Errorf("expected leading content preserved, got %q", got)
	}
}

func TestSanitize_PreservesEmoji(t *testing.T) {
	// Agent Voice & Content Philosophy (2026-06-25): emoji are voice, not
	// noise — this locks in the deliberate reversal of 1.8's emoji-stripping.
	got := Sanitize("🔍 DeFi Watcher: spotted a rug pull risk. Stay vigilant, folks. 🚨")
	if !strings.Contains(got, "🔍") || !strings.Contains(got, "🚨") {
		t.Errorf("expected emoji preserved, got %q", got)
	}
}

func TestSanitize_PreservesCompoundEmojiWithJoiners(t *testing.T) {
	got := Sanitize("Investigating further. 🕵️‍♂️ Will report back.")
	if !strings.Contains(got, "🕵") {
		t.Errorf("expected compound emoji preserved, got %q", got)
	}
}

func TestSanitize_RedactsRealLookingCVEs(t *testing.T) {
	// Real bug found live during 1.10 verification: a model cited
	// CVE-2022-3749 unprompted despite the system prompt's explicit ban on
	// naming real CVEs. Defensive, same as the hashtag/emoji lessons.
	got := Sanitize("Apparently this was assigned CVE-2022-3749 and nobody caught it for months.")
	if strings.Contains(got, "CVE-2022-3749") {
		t.Errorf("expected real-looking CVE id redacted, got %q", got)
	}
}

func TestSanitize_StripsGarbledMathUnicode(t *testing.T) {
	// U+1D5E7 etc. ("Mathematical Sans-Serif Bold" glyphs) — a model's
	// fake-bold artifact seen once in the real 1.8 backfill, not real text.
	garbled := "🔵 \U0001D5E7\U0001D5C8\U0001D5C2\U0001D5C2\U0001D5C8 News update."
	got := Sanitize(garbled)
	for _, r := range got {
		if isGarbledMathRune(r) {
			t.Fatalf("expected garbled math-alphanumeric runes stripped, got %q in %q", string(r), got)
		}
	}
	if !strings.Contains(got, "News update.") {
		t.Errorf("expected surrounding real text preserved, got %q", got)
	}
}

func TestSanitize_BacksOffIncompleteTrailingSentence(t *testing.T) {
	got := Sanitize("First full sentence here. Second one too! Paper title: 'Evaluating Synthetic Data for Model Training via")
	if got != "First full sentence here. Second one too!" {
		t.Errorf("expected trailing incomplete sentence dropped, got %q", got)
	}
}

func TestSanitize_KeepsContentEndingWithoutPunctuationIfNoGoodBoundary(t *testing.T) {
	got := Sanitize("A short fragment with no terminal punctuation at all")
	if got == "" {
		t.Error("expected content preserved when no safe sentence boundary exists")
	}
}

func TestSanitize_EnforcesMaxLength(t *testing.T) {
	long := strings.Repeat("word ", 200) // 1000 chars
	got := Sanitize(long)
	if len([]rune(got)) > 500 {
		t.Fatalf("expected <= 500 runes, got %d", len([]rune(got)))
	}
}
