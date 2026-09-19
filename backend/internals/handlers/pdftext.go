package handlers

import (
	"regexp"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

var whitespaceRun = regexp.MustCompile(`\s+`)

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// extractPdfTitleFromContent guesses a title from the largest text on a
// PDF's first page - for PDFs with no usable Title in their document-info
// metadata at all (common for LaTeX-generated PDFs, including a lot of
// arXiv preprints, whose authors never set \hypersetup{pdftitle=...}).
// This is the same basic heuristic many PDF tools use: a paper's title is
// almost always the single biggest piece of text near the top of page 1.
// It's a heuristic, not a guarantee - unusual layouts (multi-column
// templates, a big logo/figure on page 1, etc.) can fool it - but it's a
// meaningfully better guess than a bare filename, and a bad guess is easy
// to fix by hand afterward same as any other title.
func extractPdfTitleFromContent(filePath string) string {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	if r.NumPage() < 1 {
		return ""
	}
	page := r.Page(1)
	if page.V.IsNull() {
		return ""
	}

	content := page.Content()

	var maxSize float64
	for _, t := range content.Text {
		if strings.TrimSpace(t.S) == "" {
			continue
		}
		if t.FontSize > maxSize {
			maxSize = t.FontSize
		}
	}
	if maxSize <= 0 {
		return ""
	}

	type frag struct {
		x, y, w, fontSize float64
		s                 string
	}
	var frags []frag
	for _, t := range content.Text {
		s := t.S // keep as-is, including leading/trailing space characters: some PDFs render an explicit " " as its own fragment
		if strings.TrimSpace(s) == "" && s != " " {
			continue
		}
		if t.FontSize >= maxSize*0.9 {
			frags = append(frags, frag{t.X, t.Y, t.W, t.FontSize, s})
		}
	}
	if len(frags) == 0 {
		return ""
	}

	// Reading order: top of page first (Y decreases going down the page in
	// this library's coordinate system), left to right within a line.
	sort.Slice(frags, func(i, j int) bool {
		if absFloat(frags[i].y-frags[j].y) > 2 {
			return frags[i].y > frags[j].y
		}
		return frags[i].x < frags[j].x
	})

	// Some PDFs (confirmed: arXiv/pdfTeX output) position every individual
	// GLYPH as its own separately-placed text fragment rather than whole
	// words - so "did the previous fragment end and this one begin" can't
	// be assumed to mean "these are different words." Only insert a space
	// when the horizontal gap between fragments is wide relative to the
	// font size (a real inter-word gap), not for the near-zero gaps
	// between adjacent letters of the same word.
	var b strings.Builder
	var prev *frag
	for i := range frags {
		fr := &frags[i]
		if prev != nil {
			sameLine := absFloat(fr.y-prev.y) <= 2
			gap := fr.x - (prev.x + prev.w)
			if !sameLine || gap > prev.fontSize*0.15 {
				b.WriteString(" ")
			}
		}
		b.WriteString(strings.TrimSpace(fr.s))
		prev = fr
	}

	title := strings.TrimSpace(whitespaceRun.ReplaceAllString(b.String(), " "))
	if len(title) < 4 || len(title) > 300 {
		return ""
	}
	return title
}
