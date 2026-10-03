package citation

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// PDFIdentifiers is what could be found inside a PDF that identifies the
// work, plus the text it was found in (used later to sanity-check that a
// looked-up record really is this paper and not something it cites).
type PDFIdentifiers struct {
	// MetadataDOIs come from the PDF's XMP/document info. Normally the
	// document's own DOI, and the best source when the body text doesn't
	// extract.
	MetadataDOIs []string
	// DOIs found in the page text, in reading order. Usually the paper's
	// own, but page 1 can mention others.
	DOIs     []string
	ArxivIDs []string
	ISBNs    []string
	// Text of the first few pages, for verification. May be empty for
	// scanned/image-only PDFs.
	Text string
}

const (
	// textPages is how far into the document to look for a DOI or arXiv
	// id - publishers put the DOI on page 1, occasionally page 2.
	textPages = 2
	// isbnPages: a book's ISBN lives on the copyright page, a few pages in.
	isbnPages = 8
	// maxRawScan bounds how much of the file is scanned for uncompressed
	// XMP metadata.
	maxRawScan = 64 << 20
)

var (
	// Crossref's recommended DOI character set, plus "_" and "+" (Springer
	// chapter DOIs use "_"). ASCII only, so a following "⟩" or curly quote
	// in the text isn't swallowed.
	doiPattern = regexp.MustCompile(`(?i)\b10\.\d{4,9}/[-._;()/:a-z0-9+]+`)

	// New-style (0704.0001 onwards) and old-style (hep-th/9901001) arXiv
	// ids, as stamped in the margin of arXiv PDFs ("arXiv:2301.01234v2").
	arxivPattern = regexp.MustCompile(`(?i)arxiv\s*:\s*(\d{4}\.\d{4,5}|[a-z\-]+(?:\.[a-z]{2})?/\d{7})(?:v\d+)?`)

	lineSpaces = regexp.MustCompile(`[ \t]+`)

	isbnPattern = regexp.MustCompile(`(?i)ISBN(?:-1[03])?[\s:]*((?:97[89][\s\-]?)?(?:\d[\s\-]?){9}[\dX])`)

	// XMP fields publishers (Elsevier, Springer, Wiley, ACM...) use for
	// the work's own DOI. Matched against the raw file bytes, which works
	// because XMP packets are normally stored uncompressed.
	xmpDOIPattern = regexp.MustCompile(`(?is)<(?:prism:doi|pdfx:doi|crossmark:DOI)>\s*(?:doi:)?\s*(10\.[^<\s]+)\s*</|<dc:identifier>\s*(?:<rdf:\w+>\s*<rdf:li>)?\s*(?:doi:|https?://(?:dx\.)?doi\.org/)(10\.[^<\s]+)`)
)

// ExtractPDF reads identifiers out of the PDF at path. It never fails: a
// PDF it can't parse just yields empty identifiers. Page parsing can be
// slow (some pages take ~0.5s), so it stops early once ctx is done and
// only reads past page 2 - looking for a book's ISBN - when no DOI or
// arXiv id turned up.
func ExtractPDF(ctx context.Context, path string) PDFIdentifiers {
	var ids PDFIdentifiers

	ids.MetadataDOIs = metadataDOIs(path)

	pages := openPages(path)
	defer pages.close()

	var early strings.Builder
	for i := 1; i <= textPages && ctx.Err() == nil; i++ {
		early.WriteString(pages.text(i))
		early.WriteString("\n")
	}
	ids.Text = early.String()

	for _, d := range findDOIs(ids.Text) {
		ids.DOIs = appendUnique(ids.DOIs, d)
	}
	// Letter-spaced text lays out as "10.1109/C VPRW .2019.00320". Lower
	// priority: removing spaces can also glue a DOI to the next word, but
	// a wrong candidate just fails lookup or the title check.
	for _, d := range findDOIs(lineSpaces.ReplaceAllString(ids.Text, "")) {
		ids.DOIs = appendUnique(ids.DOIs, d)
	}
	ids.ArxivIDs = findArxivIDs(ids.Text)
	ids.ISBNs = findISBNs(ids.Text)

	if len(ids.DOIs) == 0 && len(ids.ArxivIDs) == 0 {
		for i := textPages + 1; i <= isbnPages && ctx.Err() == nil; i++ {
			for _, isbn := range findISBNs(pages.text(i)) {
				ids.ISBNs = appendUnique(ids.ISBNs, isbn)
			}
		}
	}
	return ids
}

// metadataDOIs looks for the work's DOI in XMP and the document-info
// dictionary (Elsevier, for one, writes "doi:10..." into Subject).
func metadataDOIs(path string) []string {
	var out []string

	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	raw, _ := io.ReadAll(io.LimitReader(f, maxRawScan))
	for _, m := range xmpDOIPattern.FindAllSubmatch(raw, -1) {
		for _, g := range m[1:] {
			if len(g) > 0 {
				out = appendUnique(out, cleanDOI(string(g)))
			}
		}
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return out
	}
	info, err := safePDFInfo(f, path)
	if err == nil && info != nil {
		fields := append([]string{info.Subject}, info.Keywords...)
		for _, d := range findDOIs(strings.Join(fields, " ")) {
			out = appendUnique(out, d)
		}
	}
	return out
}

func safePDFInfo(rs io.ReadSeeker, path string) (info *pdfcpuInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("pdfcpu panic: %v", r)
		}
	}()
	i, err := api.PDFInfo(rs, path, nil, false, nil)
	if err != nil || i == nil {
		return nil, err
	}
	return &pdfcpuInfo{Subject: i.Subject, Keywords: i.Keywords}, nil
}

type pdfcpuInfo struct {
	Subject  string
	Keywords []string
}

// pdfPages reads page text on demand. A file that won't open just reads
// as having no text.
type pdfPages struct {
	f *os.File
	r *pdf.Reader
}

func openPages(path string) (p pdfPages) {
	defer func() {
		// ledongthuc/pdf panics on some malformed files.
		if rec := recover(); rec != nil {
			p = pdfPages{}
		}
	}()
	f, r, err := pdf.Open(path)
	if err != nil {
		return pdfPages{}
	}
	return pdfPages{f: f, r: r}
}

func (p pdfPages) close() {
	if p.f != nil {
		p.f.Close()
	}
}

// text returns page n's text (1-based), or "" past the end or on a parse
// failure. It lays the text out from the positioned fragments rather than
// using GetPlainText, which runs lines together with no separator - so a
// DOI at the end of a line would swallow the first word of the next one.
func (p pdfPages) text(n int) (out string) {
	if p.r == nil || n > p.r.NumPage() {
		return ""
	}
	defer func() {
		if rec := recover(); rec != nil {
			out = ""
		}
	}()
	page := p.r.Page(n)
	if page.V.IsNull() {
		return ""
	}
	return layoutText(page.Content().Text)
}

// layoutText orders fragments top-to-bottom, left-to-right and joins
// them, inserting a space at word-sized gaps and a newline between lines.
// Some PDFs (pdfTeX/arXiv) place every glyph separately, so adjacent
// fragments aren't assumed to be separate words.
func layoutText(texts []pdf.Text) string {
	frags := make([]pdf.Text, 0, len(texts))
	for _, t := range texts {
		if t.S != "" {
			frags = append(frags, t)
		}
	}
	sort.SliceStable(frags, func(i, j int) bool {
		if abs(frags[i].Y-frags[j].Y) > 2 {
			return frags[i].Y > frags[j].Y
		}
		return frags[i].X < frags[j].X
	})

	var b bytes.Buffer
	for i, fr := range frags {
		if i > 0 {
			prev := frags[i-1]
			if abs(fr.Y-prev.Y) > 2 {
				b.WriteByte('\n')
			} else if fr.X-(prev.X+prev.W) > prev.FontSize*0.15 {
				b.WriteByte(' ')
			}
		}
		b.WriteString(fr.S)
	}
	return b.String()
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func findDOIs(text string) []string {
	var out []string
	for _, m := range doiPattern.FindAllString(text, -1) {
		if d := cleanDOI(m); d != "" {
			out = appendUnique(out, d)
		}
	}
	return out
}

// cleanDOI strips the punctuation a DOI picks up from the sentence or
// markup around it.
func cleanDOI(d string) string {
	d = strings.TrimSpace(d)
	d = strings.TrimRight(d, ".,;:]}>")
	// A trailing ")" belongs to the DOI only if it opened a "(" inside it,
	// as in 10.1016/S0140-6736(97)11096-0.
	for strings.HasSuffix(d, ")") && strings.Count(d, "(") < strings.Count(d, ")") {
		d = strings.TrimSuffix(d, ")")
		d = strings.TrimRight(d, ".,;:")
	}
	if !strings.Contains(d, "/") || strings.HasSuffix(d, "/") {
		return ""
	}
	return d
}

// findArxivIDs also searches the text with whitespace removed, forwards
// and reversed: arXiv's id stamp is rotated text in the left margin, which
// positional layout turns into one glyph per line, read bottom to top.
func findArxivIDs(text string) []string {
	squashed := whitespace.ReplaceAllString(text, "")
	runes := []rune(squashed)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	var out []string
	for _, t := range []string{text, squashed, string(runes)} {
		for _, m := range arxivPattern.FindAllStringSubmatch(t, -1) {
			out = appendUnique(out, m[1])
		}
	}
	return out
}

func findISBNs(text string) []string {
	var out []string
	for _, m := range isbnPattern.FindAllStringSubmatch(text, -1) {
		if isbn := normalizeISBN(m[1]); isbn != "" {
			out = appendUnique(out, isbn)
		}
	}
	return out
}

// normalizeISBN strips separators and returns the ISBN only if its check
// digit is valid.
func normalizeISBN(s string) string {
	s = strings.ToUpper(strings.NewReplacer(" ", "", "-", "").Replace(s))
	switch len(s) {
	case 10:
		sum := 0
		for i, c := range s {
			var v int
			switch {
			case c >= '0' && c <= '9':
				v = int(c - '0')
			case c == 'X' && i == 9:
				v = 10
			default:
				return ""
			}
			sum += v * (10 - i)
		}
		if sum%11 == 0 {
			return s
		}
	case 13:
		sum := 0
		for i, c := range s {
			if c < '0' || c > '9' {
				return ""
			}
			v := int(c - '0')
			if i%2 == 1 {
				v *= 3
			}
			sum += v
		}
		if sum%10 == 0 {
			return s
		}
	}
	return ""
}

func appendUnique(list []string, v string) []string {
	if v == "" {
		return list
	}
	for _, x := range list {
		if strings.EqualFold(x, v) {
			return list
		}
	}
	return append(list, v)
}
