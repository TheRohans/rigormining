package citation

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Hints is everything known about an item before lookup. All optional.
type Hints struct {
	// DOI and ISBN already recorded on the item (from the user, a Zotero
	// import, or the page's citation meta tags) - trusted as-is.
	DOI  string
	ISBN string
	// Title is the item's current title, or the best guess from the PDF.
	// Used only for a fuzzy Crossref search when nothing else matches.
	Title string
	// PDF is what ExtractPDF found in the attached file, if any.
	PDF *PDFIdentifiers
}

const (
	// maxDOICandidates caps how many DOIs found in a PDF get looked up.
	maxDOICandidates = 3
	// searchThreshold is how similar a title-search result's title must
	// be to ours to be accepted. High on purpose: a wrong citation is
	// worse than none.
	searchThreshold = 0.85
)

// arxivDOIPrefix is arXiv's DataCite DOI prefix, which Crossref doesn't
// serve - those DOIs are looked up through the arXiv API instead.
const arxivDOIPrefix = "10.48550/arxiv."

// Resolve finds the citation for an item, trying the most reliable
// identifier first: a DOI already on the item, then identifiers found in
// the PDF (DOI, arXiv id, ISBN), then a fuzzy title search. It returns
// ErrNotFound when nothing matched confidently.
//
// DOIs found in the PDF's text are only accepted if the record's title
// appears in that text, since page 1 can cite other papers' DOIs too.
func (c *Client) Resolve(ctx context.Context, h Hints) (*Metadata, error) {
	// verified reports whether title checks out against the PDF's text.
	// With no PDF, or a PDF whose text didn't extract, there's nothing to
	// check against, so the identifier is taken on trust.
	verified := func(title string) bool {
		if h.PDF == nil || !usableText(h.PDF.Text) {
			return true
		}
		return TitleInText(title, h.PDF.Text)
	}

	if h.DOI != "" {
		if md, err := c.byDOI(ctx, h.DOI); err == nil {
			md.MatchedBy = "DOI on item"
			return md, nil
		} else if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	if h.PDF != nil {
		for _, doi := range h.PDF.MetadataDOIs {
			md, err := c.byDOI(ctx, doi)
			if errors.Is(err, ErrNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			// Usually right, but one real Springer PDF carried an
			// unrelated article's DOI in its XMP alongside its own.
			if verified(md.Title) {
				md.MatchedBy = "DOI in PDF metadata"
				return md, nil
			}
		}

		for i, doi := range h.PDF.DOIs {
			if i == maxDOICandidates {
				break
			}
			md, err := c.byDOI(ctx, doi)
			if errors.Is(err, ErrNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if verified(md.Title) {
				md.MatchedBy = "DOI in PDF"
				return md, nil
			}
		}

		for i, id := range h.PDF.ArxivIDs {
			if i == maxDOICandidates {
				break
			}
			md, err := c.byArxiv(ctx, id)
			if errors.Is(err, ErrNotFound) {
				continue
			}
			if err != nil {
				return nil, err
			}
			// Same check as DOIs: page 1 can cite other arXiv papers.
			if verified(md.Title) {
				md.MatchedBy = "arXiv ID in PDF"
				return md, nil
			}
		}
	}

	isbns := []string{}
	if h.ISBN != "" {
		isbns = append(isbns, normalizeISBN(h.ISBN))
	}
	if h.PDF != nil {
		isbns = append(isbns, h.PDF.ISBNs...)
	}
	for i, isbn := range isbns {
		if isbn == "" {
			continue
		}
		md, err := c.OpenLibraryISBN(ctx, isbn)
		if errors.Is(err, ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		// An ISBN on the item is trusted; one found in the PDF could be a
		// series or proceedings ISBN, so check the title matches.
		fromItem := i == 0 && h.ISBN != ""
		if fromItem || verified(md.Title) {
			md.MatchedBy = "ISBN"
			return md, nil
		}
	}

	if title := strings.TrimSpace(h.Title); len(significantWords(title)) >= 3 {
		results, err := c.CrossrefSearch(ctx, title)
		if err != nil {
			return nil, err
		}
		for _, md := range results {
			if TitleSimilarity(title, md.Title) >= searchThreshold && verified(md.Title) {
				md.MatchedBy = "title search"
				return md, nil
			}
		}
	}

	return nil, ErrNotFound
}

func (c *Client) byDOI(ctx context.Context, doi string) (*Metadata, error) {
	if strings.HasPrefix(strings.ToLower(doi), arxivDOIPrefix) {
		return c.byArxiv(ctx, doi[len(arxivDOIPrefix):])
	}
	md, err := c.CrossrefDOI(ctx, doi)
	if errors.Is(err, ErrNotFound) {
		// Not every DOI is Crossref's - try the other big registry.
		md, err = c.DataciteDOI(ctx, doi)
	}
	// DOIs are case-insensitive, but registries often return them
	// lowercased; keep the casing the publisher printed.
	if err == nil && strings.EqualFold(md.Doi, doi) {
		md.Doi = doi
	}
	return md, err
}

// arxivWait is how long to wait on the arXiv API before settling for
// DataCite's record. arXiv's API usually answers in milliseconds but has
// been seen taking 37s for an id it hadn't served recently.
const arxivWait = 2 * time.Second

// byArxiv looks up a preprint. It asks arXiv and DataCite at once: arXiv
// is preferred because it links to the published version - when there is
// one, that journal's richer Crossref record is returned instead - but
// DataCite's copy of the preprint record is used if arXiv is slow.
func (c *Client) byArxiv(ctx context.Context, id string) (*Metadata, error) {
	type result struct {
		md  *Metadata
		err error
	}
	datacite := make(chan result, 1)
	go func() {
		md, err := c.DataciteDOI(ctx, "10.48550/arXiv."+id)
		datacite <- result{md, err}
	}()

	arxivCtx, cancel := context.WithTimeout(ctx, arxivWait)
	defer cancel()
	if rec, err := c.Arxiv(arxivCtx, id); err == nil {
		if rec.PublishedDOI != "" {
			if md, err := c.CrossrefDOI(ctx, rec.PublishedDOI); err == nil {
				return md, nil
			}
		}
		return &rec.Metadata, nil
	}

	r := <-datacite
	if r.err != nil {
		return nil, r.err
	}
	// Match the arXiv record's shape: DataCite types arXiv papers as
	// Preprint but has no venue.
	r.md.ItemType = "preprint"
	r.md.Venue = "arXiv preprint arXiv:" + id
	return r.md, nil
}
