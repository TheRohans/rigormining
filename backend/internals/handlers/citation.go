package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"gitlab.com/robrohan/rigormining/internals/citation"
	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
)

// backgroundLookupTimeout bounds a lookup that outlives its upload
// request (PDF parsing plus a few API calls).
const backgroundLookupTimeout = 30 * time.Second

func citationHints(item *models.LibraryItem) citation.Hints {
	h := citation.Hints{Title: item.Title}
	if item.Doi != nil {
		h.DOI = *item.Doi
	}
	if item.Isbn != nil {
		h.ISBN = *item.Isbn
	}
	return h
}

func pdfPath(item *models.LibraryItem) string {
	if item.FilePath != nil && item.FileType != nil && *item.FileType == "pdf" {
		return *item.FilePath
	}
	return ""
}

// resolveCitation extracts identifiers from the item's PDF (if any) and
// looks the item up.
func resolveCitation(ctx context.Context, c *citation.Client, item *models.LibraryItem) (*citation.Metadata, error) {
	h := citationHints(item)
	if path := pdfPath(item); path != "" {
		ids := citation.ExtractPDF(ctx, path)
		h.PDF = &ids
	}
	return c.Resolve(ctx, h)
}

// applyCitation fills the item's empty bibliographic fields from md,
// never overwriting what's already there - that came from the user, a
// Zotero import or the page's citation tags. The title is the exception
// when replaceTitle is set, for callers whose title is only a weak guess
// (a filename, a page <title>, the biggest text on page 1).
func applyCitation(item *models.LibraryItem, md *citation.Metadata, replaceTitle bool) {
	// Same words means the current title is already right, and probably
	// better capitalised - Crossref often has titles in sentence case.
	if replaceTitle && md.Title != "" && citation.TitleSimilarity(item.Title, md.Title) < 1 {
		item.Title = md.Title
	}
	if item.Authors == "" {
		item.Authors = md.Authors
	}
	if item.Year == nil && md.Year != 0 {
		y := md.Year
		item.Year = &y
	}
	fill := func(field **string, v string) {
		if (*field == nil || **field == "") && v != "" {
			s := v
			*field = &s
		}
	}
	fill(&item.Doi, md.Doi)
	fill(&item.Isbn, md.Isbn)
	fill(&item.ItemType, md.ItemType)
	fill(&item.Venue, md.Venue)
	fill(&item.Volume, md.Volume)
	fill(&item.Number, md.Number)
	fill(&item.Pages, md.Pages)
	fill(&item.Publisher, md.Publisher)
}

// citationAutofill is a lookup started for a newly uploaded or captured
// item. The handler waits a short while for it (applyNow); if it isn't
// done by then, it finishes after the response (finishInBackground), so a
// slow PDF or API never fails or stalls an upload.
type citationAutofill struct {
	result       chan *citation.Metadata
	replaceTitle bool
	titleAtStart string
}

// startAutofill begins a lookup for item, or returns nil when lookup is
// disabled or there's nothing to look up by (no PDF, DOI or ISBN). A nil
// *citationAutofill is safe to use.
func startAutofill(e *env.Env, item *models.LibraryItem, replaceTitle bool) *citationAutofill {
	if e.Citation == nil {
		return nil
	}
	if pdfPath(item) == "" && item.Doi == nil && item.Isbn == nil {
		return nil
	}

	a := &citationAutofill{
		result:       make(chan *citation.Metadata, 1),
		replaceTitle: replaceTitle,
		titleAtStart: item.Title,
	}
	snapshot := *item
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), backgroundLookupTimeout)
		defer cancel()
		md, err := resolveCitation(ctx, e.Citation, &snapshot)
		if err != nil && !errors.Is(err, citation.ErrNotFound) {
			e.Log.Warn("citation lookup failed", "item", snapshot.UUID, "error", err)
		}
		a.result <- md
	}()
	return a
}

// applyNow waits up to the configured Lookup.Wait for the result and
// applies it to item. It returns true if the lookup is still running.
func (a *citationAutofill) applyNow(e *env.Env, item *models.LibraryItem) (pending bool) {
	if a == nil {
		return false
	}
	select {
	case md := <-a.result:
		if md != nil {
			applyCitation(item, md, a.replaceTitle)
		}
		return false
	case <-time.After(e.Cfg.Lookup.Wait):
		return true
	}
}

// finishInBackground waits for a lookup that outlived its request and
// saves the result onto the stored item. Call only after the item has
// been saved.
func (a *citationAutofill) finishInBackground(e *env.Env, itemID, userID string) {
	if a == nil {
		return
	}
	go func() {
		md := <-a.result
		if md == nil {
			return
		}
		item, err := e.Repo.GetItemById(itemID)
		if err != nil || item.UserId != userID {
			return
		}
		// Only replace the title if nobody has edited it in the meantime.
		applyCitation(item, md, a.replaceTitle && item.Title == a.titleAtStart)
		if err := e.Repo.UpdateItem(item); err != nil {
			e.Log.Error("saving background citation lookup failed", "item", itemID, "error", err)
		}
	}()
}

// citationLookupResponse is what GET /items/{id}/citation returns: the
// looked-up fields in the same shape the item update endpoint takes, so
// the UI can drop them into its form for review.
type citationLookupResponse struct {
	Source    string            `json:"source"`
	MatchedBy string            `json:"matched_by"`
	Metadata  itemMetadataInput `json:"metadata"`
}

// APILookupCitation looks up an existing item's citation and returns it
// without saving anything - the user reviews it in the edit form first.
func APILookupCitation(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		if e.Citation == nil {
			writeError(w, http.StatusNotFound, "citation lookup is disabled")
			return
		}

		// Leave enough of the server's write timeout to send the reply.
		budget := e.Cfg.Web.WriteTimeout - time.Second
		if budget < 2*time.Second {
			budget = 2 * time.Second
		}
		ctx, cancel := context.WithTimeout(r.Context(), budget)
		defer cancel()

		md, err := resolveCitation(ctx, e.Citation, item)
		switch {
		case errors.Is(err, citation.ErrNotFound):
			writeError(w, http.StatusNotFound, "no confident match found")
			return
		case errors.Is(err, context.DeadlineExceeded):
			writeError(w, http.StatusGatewayTimeout, "lookup timed out - try again")
			return
		case err != nil:
			e.Log.Warn("citation lookup failed", "item", item.UUID, "error", err)
			writeError(w, http.StatusBadGateway, "lookup service unavailable - try again later")
			return
		}

		writeJSON(w, http.StatusOK, citationLookupResponse{
			Source:    md.Source,
			MatchedBy: md.MatchedBy,
			Metadata: itemMetadataInput{
				Title:     md.Title,
				Authors:   md.Authors,
				Doi:       md.Doi,
				Isbn:      md.Isbn,
				Year:      md.Year,
				ItemType:  md.ItemType,
				Venue:     md.Venue,
				Volume:    md.Volume,
				Number:    md.Number,
				Pages:     md.Pages,
				Publisher: md.Publisher,
			},
		})
	}
}
