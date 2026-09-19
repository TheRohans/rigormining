package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
)

// metaTag mirrors one <meta name="..." content="..."> pair as collected by
// the browser extension's content script.
type metaTag struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

var doiPattern = regexp.MustCompile(`10\.\d{4,9}/[^\s"'<>]+`)

// extractCitation does the actual citation/DOI parsing server-side, so the
// extension can stay a thin client and this logic can improve without
// shipping a new extension version.
func extractCitation(pageURL, pageTitle string, tags []metaTag) itemMetadataInput {
	meta := itemMetadataInput{Title: pageTitle}
	if pageURL != "" {
		meta.SourceUrl = pageURL
	}

	get := func(names ...string) string {
		for _, tag := range tags {
			for _, n := range names {
				if strings.EqualFold(tag.Name, n) && tag.Content != "" {
					return tag.Content
				}
			}
		}
		return ""
	}

	if t := get("citation_title", "dc.title"); t != "" {
		meta.Title = t
	}

	var authors []string
	for _, tag := range tags {
		if (strings.EqualFold(tag.Name, "citation_author") || strings.EqualFold(tag.Name, "dc.creator")) && tag.Content != "" {
			authors = append(authors, tag.Content)
		}
	}
	meta.Authors = strings.Join(authors, "; ")

	if doi := get("citation_doi"); doi != "" {
		meta.Doi = strings.TrimPrefix(doi, "doi:")
	} else if m := doiPattern.FindString(pageURL); m != "" {
		meta.Doi = m
	}

	if isbn := get("citation_isbn"); isbn != "" {
		meta.Isbn = isbn
	}

	if dateStr := get("citation_publication_date", "citation_date", "dc.date"); len(dateStr) >= 4 {
		if y, err := strconv.Atoi(dateStr[:4]); err == nil {
			meta.Year = y
		}
	}

	// These are the same Highwire Press / Google Scholar meta tag
	// conventions already relied on above - widely used by arXiv, IEEE
	// Xplore, ACM DL, SpringerLink, Nature, etc. A page with none of these
	// just leaves the corresponding field blank, same as any other manual
	// entry - nothing downstream requires them.
	if journal := get("citation_journal_title"); journal != "" {
		meta.ItemType = "journalArticle"
		meta.Venue = journal
	} else if conf := get("citation_conference_title"); conf != "" {
		meta.ItemType = "conferencePaper"
		meta.Venue = conf
	} else if inst := get("citation_dissertation_institution"); inst != "" {
		meta.ItemType = "thesis"
		meta.Publisher = inst
	} else if inst := get("citation_technical_report_institution"); inst != "" {
		meta.ItemType = "report"
		meta.Publisher = inst
	}

	meta.Volume = get("citation_volume")
	meta.Number = get("citation_issue")
	if meta.Publisher == "" {
		meta.Publisher = get("citation_publisher")
	}

	first, last := get("citation_firstpage"), get("citation_lastpage")
	switch {
	case first != "" && last != "":
		meta.Pages = first + "-" + last
	case first != "":
		meta.Pages = first
	}

	return meta
}

// APICapture is what the browser extension calls: it sends the page URL,
// title, and its collected <meta> tags (and optionally the fetched PDF
// blob as the "file" part, when the page allowed a same-origin fetch). A
// missing file is expected for paywalled/cross-origin pages - the item is
// still created so it can be attached to later from the library UI.
func APICapture(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "could not parse form")
			return
		}

		var tags []metaTag
		if raw := r.FormValue("meta"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &tags); err != nil {
				writeError(w, http.StatusBadRequest, "invalid meta JSON")
				return
			}
		}

		pageTitle := r.FormValue("title")
		meta := extractCitation(r.FormValue("url"), pageTitle, tags)
		if meta.Title == "" {
			writeError(w, http.StatusBadRequest, "could not determine a title for this page")
			return
		}

		item := models.LibraryItem{
			UUID:      uuid.New().String(),
			UserId:    env.UserFromContext(r.Context()).UUID,
			Title:     meta.Title,
			Authors:   meta.Authors,
			AddedDate: time.Now().UTC().Format(time.RFC3339),
		}
		if meta.Doi != "" {
			item.Doi = &meta.Doi
		}
		if meta.Isbn != "" {
			item.Isbn = &meta.Isbn
		}
		if meta.Year != 0 {
			item.Year = &meta.Year
		}
		if meta.SourceUrl != "" {
			item.SourceUrl = &meta.SourceUrl
		}
		if meta.ItemType != "" {
			item.ItemType = &meta.ItemType
		}
		if meta.Venue != "" {
			item.Venue = &meta.Venue
		}
		if meta.Volume != "" {
			item.Volume = &meta.Volume
		}
		if meta.Number != "" {
			item.Number = &meta.Number
		}
		if meta.Pages != "" {
			item.Pages = &meta.Pages
		}
		if meta.Publisher != "" {
			item.Publisher = &meta.Publisher
		}

		// meta.Title unchanged from pageTitle means extractCitation found no
		// citation_title meta tag, so this is still just the raw fallback
		// the extension sent - a document <title> (often cluttered with a
		// journal/site name suffix), or - for the "no content script could
		// run" case (Firefox's own PDF viewer, a redirected S3 URL, etc.) -
		// a bare filename like "main". A real embedded PDF title beats
		// either of those; an actual citation_title from the page beats a
		// PDF's, so this only kicks in when citation extraction found
		// nothing to work with.
		preferExtractedTitle := meta.Title == pageTitle

		duplicate, err := attachUploadedFile(e, r, &item, preferExtractedTitle)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if duplicate != nil {
			writeJSON(w, http.StatusOK, itemResponse(e, duplicate))
			return
		}

		if err := e.Repo.CreateItem(&item); err != nil {
			e.Log.Error("CreateItem (capture) failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not create item")
			return
		}

		writeJSON(w, http.StatusCreated, itemResponse(e, &item))
	}
}
