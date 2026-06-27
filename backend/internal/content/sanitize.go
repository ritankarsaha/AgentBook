package content

import (
	"regexp"
	"strings"
)

var (
	markdownBoldRe  = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	headingMarkerRe = regexp.MustCompile(`(?m)^#{1,6}\s*`)

	hashtagRe = regexp.MustCompile(`(?:^|\s)#\w+`)

	cveRe        = regexp.MustCompile(`(?i)CVE-\d{4}-\d{4,7}`)
	blankLinesRe = regexp.MustCompile(`\n{3,}`)
	multiSpaceRe = regexp.MustCompile(`[ \t]{2,}`)
)

var quoteClosers = map[rune]rune{
	'"': '"', '\'': '\'', '“': '”', '‘': '’',
}

func stripWrappingQuotes(s string) string {
	runes := []rune(s)
	if len(runes) < 2 {
		return s
	}
	closer, isQuote := quoteClosers[runes[0]]
	if !isQuote {
		return s
	}
	if runes[len(runes)-1] == closer {
		return strings.TrimSpace(string(runes[1 : len(runes)-1]))
	}
	return strings.TrimSpace(string(runes[1:]))
}

func isGarbledMathRune(r rune) bool {
	return r >= 0x1D400 && r <= 0x1D7FF
}

func stripGarbledUnicode(s string) string {
	return strings.Map(func(r rune) rune {
		if isGarbledMathRune(r) {
			return -1
		}
		return r
	}, s)
}

const minViableSentenceLen = 20

func isEmojiRune(r rune) bool {
	switch {
	case r >= 0x1F300 && r <= 0x1FAFF:
		return true
	case r >= 0x2600 && r <= 0x27BF:
		return true
	case r >= 0x1F1E6 && r <= 0x1F1FF:
		return true
	case r == 0xFE0F || r == 0x200D:
		return true
	default:
		return false
	}
}

func trimToCompleteSentence(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	last := runes[len(runes)-1]
	if strings.ContainsRune(`.!?"'”’)`, last) || isEmojiRune(last) {
		return s
	}
	idx := strings.LastIndexAny(s, ".!?")
	if idx < minViableSentenceLen {
		return s
	}
	return strings.TrimSpace(s[:idx+1])
}

func Sanitize(raw string) string {
	s := strings.TrimSpace(raw)
	s = stripWrappingQuotes(s)
	s = markdownBoldRe.ReplaceAllString(s, "$1")
	s = headingMarkerRe.ReplaceAllString(s, "")
	s = hashtagRe.ReplaceAllString(s, "")
	s = cveRe.ReplaceAllString(s, "a CVE")
	s = stripGarbledUnicode(s)
	s = blankLinesRe.ReplaceAllString(s, "\n\n")
	s = multiSpaceRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)

	runes := []rune(s)
	if len(runes) > 500 {
		runes = runes[:500]
		// back off to the last word boundary so we don't cut mid-word
		if idx := strings.LastIndexAny(string(runes), " \n"); idx > 400 {
			runes = []rune(string(runes)[:idx])
		}
		s = strings.TrimSpace(string(runes))
	}
	return trimToCompleteSentence(s)
}
