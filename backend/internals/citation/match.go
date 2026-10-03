package citation

import (
	"html"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	markupTag  = regexp.MustCompile(`<[^>]+>`)
	whitespace = regexp.MustCompile(`\s+`)
)

// cleanText strips the JATS/HTML markup and entities Crossref titles
// carry (<i>, <sub>, &amp;) and collapses whitespace.
func cleanText(s string) string {
	s = markupTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(whitespace.ReplaceAllString(s, " "))
}

// fold lowercases, strips accents and expands ligatures (PDF text often
// has "ﬁ" for "fi"), keeping letters, digits and spaces only.
func fold(s string) string {
	s = norm.NFKD.String(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Mn, r):
			// combining accent - drop
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// significantWords is a title's words minus short filler, for comparing
// titles that differ in punctuation, case or small words.
func significantWords(s string) []string {
	var out []string
	for _, w := range strings.Fields(fold(s)) {
		if len(w) >= 3 {
			out = append(out, w)
		}
	}
	return out
}

// minUsableWords is how many real words extracted text needs before it's
// worth checking a title against. Some PDFs' body text doesn't extract
// at all (only a margin stamp and footers come out), and failing the
// check against those would reject correct matches.
const minUsableWords = 100

// usableText reports whether text has enough real words to verify a
// title against.
func usableText(text string) bool {
	return len(significantWords(text)) >= minUsableWords
}

// TitleInText reports whether title plausibly appears in text - used to
// check that a DOI found in a PDF is the paper's own and not one it
// cites. Two tests, because PDF text extraction varies: most of the
// title's words present, or (for PDFs whose text comes out with the
// spaces missing) the whole title present once all spaces are removed.
func TitleInText(title, text string) bool {
	words := significantWords(title)
	if len(words) == 0 {
		return false
	}

	foldedText := fold(text)
	squashedText := strings.ReplaceAll(foldedText, " ", "")
	if strings.Contains(squashedText, strings.Join(words, "")) {
		return true
	}

	present := map[string]bool{}
	for _, w := range strings.Fields(foldedText) {
		present[w] = true
	}
	found := 0
	for _, w := range words {
		if present[w] {
			found++
		}
	}
	return float64(found)/float64(len(words)) >= 0.8
}

// TitleSimilarity is the Dice coefficient of two titles' significant
// words: 1 for the same words, 0 for nothing in common.
func TitleSimilarity(a, b string) float64 {
	wa, wb := significantWords(a), significantWords(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	counts := map[string]int{}
	for _, w := range wa {
		counts[w]++
	}
	shared := 0
	for _, w := range wb {
		if counts[w] > 0 {
			counts[w]--
			shared++
		}
	}
	return 2 * float64(shared) / float64(len(wa)+len(wb))
}

// personName formats a "First Middle Last" name as "Last, First Middle",
// this app's author convention. Particles (van, de, von...) stay with
// the surname.
func personName(full string) string {
	parts := strings.Fields(full)
	if len(parts) < 2 {
		return strings.TrimSpace(full)
	}
	split := len(parts) - 1
	for split > 1 && isParticle(parts[split-1]) {
		split--
	}
	return strings.Join(parts[split:], " ") + ", " + strings.Join(parts[:split], " ")
}

func isParticle(w string) bool {
	switch strings.ToLower(w) {
	case "van", "von", "de", "der", "den", "del", "della", "di", "da", "du", "la", "las", "le", "los", "dos", "das", "ter", "ten":
		return true
	}
	return false
}
