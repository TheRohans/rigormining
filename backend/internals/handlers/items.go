package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
	"gitlab.com/robrohan/rigormining/internals/repository"
)

func itemResponse(e *env.Env, item *models.LibraryItem) map[string]any {
	tags, _ := e.Repo.GetTagsForItem(item.UUID)
	return map[string]any{
		"id":         item.UUID,
		"title":      item.Title,
		"authors":    item.Authors,
		"doi":        item.Doi,
		"isbn":       item.Isbn,
		"year":       item.Year,
		"source_url": item.SourceUrl,
		"file_type":  item.FileType,
		"added_date": item.AddedDate,
		"sync_state": item.SyncState,
		"notes":      item.Notes,
		"item_type":  item.ItemType,
		"venue":      item.Venue,
		"volume":     item.Volume,
		"number":     item.Number,
		"pages":      item.Pages,
		"publisher":  item.Publisher,
		"tags":       tags,
	}
}

// loadOwnedItem fetches an item and 404/403s if it doesn't belong to the
// requesting user.
func loadOwnedItem(e *env.Env, w http.ResponseWriter, r *http.Request) *models.LibraryItem {
	item, err := e.Repo.GetItemById(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, "item not found")
		return nil
	}
	if item.UserId != env.UserFromContext(r.Context()).UUID {
		writeError(w, http.StatusForbidden, "forbidden")
		return nil
	}
	return item
}

func APIGetItems(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		filter := repository.ItemFilter{
			Query:     q.Get("q"),
			SyncState: parseSyncStateParam(q.Get("sync_state")),
		}

		items, err := e.Repo.ListItems(env.UserFromContext(r.Context()).UUID, filter)
		if err != nil {
			e.Log.Error("ListItems failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list items")
			return
		}

		out := make([]map[string]any, 0, len(items))
		for i := range items {
			out = append(out, itemResponse(e, &items[i]))
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func APIGetItem(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

type itemMetadataInput struct {
	Title     string   `json:"title"`
	Authors   string   `json:"authors"`
	Doi       string   `json:"doi"`
	Isbn      string   `json:"isbn"`
	Year      int      `json:"year"`
	SourceUrl string   `json:"source_url"`
	Notes     string   `json:"notes"`
	Tags      []string `json:"tags"`
	ItemType  string   `json:"item_type"`
	Venue     string   `json:"venue"`
	Volume    string   `json:"volume"`
	Number    string   `json:"number"`
	Pages     string   `json:"pages"`
	Publisher string   `json:"publisher"`
}

// APICreateItem accepts a multipart form: a "metadata" part (JSON, see
// itemMetadataInput) and an optional "file" part (the PDF/EPUB). A missing
// file part is a first-class supported case (bookmark now, attach later).
func APICreateItem(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "could not parse form")
			return
		}

		var meta itemMetadataInput
		if raw := r.FormValue("metadata"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &meta); err != nil {
				writeError(w, http.StatusBadRequest, "invalid metadata JSON")
				return
			}
		}
		if meta.Title == "" {
			writeError(w, http.StatusBadRequest, "title is required")
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
		if meta.Notes != "" {
			item.Notes = &meta.Notes
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

		// Manual uploads from the library page always title from the
		// filename (see frontend Library/index.tsx) - a real embedded PDF
		// title, when present, is essentially always better than that.
		duplicate, err := attachUploadedFile(e, r, &item, true)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if duplicate != nil {
			writeJSON(w, http.StatusOK, itemResponse(e, duplicate))
			return
		}

		if err := e.Repo.CreateItem(&item); err != nil {
			e.Log.Error("CreateItem failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not create item")
			return
		}

		for _, tag := range meta.Tags {
			if tag == "" {
				continue
			}
			if err := e.Repo.AttachTag(item.UUID, tag); err != nil {
				e.Log.Error("AttachTag failed", "error", err, "tag", tag)
			}
		}

		writeJSON(w, http.StatusCreated, itemResponse(e, &item))
	}
}

// attachUploadedFile reads the "file" part of a multipart form (if present),
// hashes it for dedupe, and saves it under the user's library directory,
// filling in item.FilePath/FileType/FileHash. A missing file part is not an
// error - it's the "bookmark now, attach a file later" path. If a file with
// the same hash already exists for this user, it's returned as duplicate
// instead of being saved again.
//
// preferExtractedTitle: when true and the saved file is a PDF with a real
// embedded document-info Title, item.Title is overwritten with it. Callers
// pass true only when they know their own title guess is weak (e.g.
// derived from a filename or a bare URL) - a citation-scraped or
// user-edited title should never be silently replaced.
func attachUploadedFile(e *env.Env, r *http.Request, item *models.LibraryItem, preferExtractedTitle bool) (duplicate *models.LibraryItem, err error) {
	file, header, ferr := r.FormFile("file")
	if ferr != nil {
		return nil, nil
	}
	defer file.Close()

	hash := sha256.New()
	tmp, err := os.CreateTemp("", "rm-upload-*")
	if err != nil {
		return nil, fmt.Errorf("could not stage upload: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(io.MultiWriter(tmp, hash), file); err != nil {
		return nil, fmt.Errorf("could not read upload: %w", err)
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))
	userId := env.UserFromContext(r.Context()).UUID

	if existing, err := e.Repo.GetItemByHash(fileHash, userId); err == nil {
		return existing, nil
	} else if err != sql.ErrNoRows {
		e.Log.Error("GetItemByHash failed", "error", err)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	fileType := strings.TrimPrefix(ext, ".")
	userDir := filepath.Join(e.Cfg.Library.Dir, userId)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		return nil, fmt.Errorf("could not create library dir: %w", err)
	}
	destPath := filepath.Join(userDir, item.UUID+ext)

	tmp.Close()
	if err := os.Rename(tmp.Name(), destPath); err != nil {
		// Rename can fail across devices/mounts (e.g. gcsfuse) - fall back
		// to a copy.
		src, err := os.Open(tmp.Name())
		if err != nil {
			return nil, fmt.Errorf("could not save upload: %w", err)
		}
		defer src.Close()
		dst, err := os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("could not save upload: %w", err)
		}
		defer dst.Close()
		io.Copy(dst, src)
	}

	item.FilePath = &destPath
	item.FileType = &fileType
	item.FileHash = &fileHash

	if preferExtractedTitle && fileType == "pdf" {
		if title := bestPdfTitle(destPath); title != "" {
			item.Title = title
		}
	}

	return nil, nil
}

func APIUpdateItem(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}

		var input itemMetadataInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if input.Title != "" {
			item.Title = input.Title
		}
		item.Authors = input.Authors
		if input.Doi != "" {
			item.Doi = &input.Doi
		}
		if input.Isbn != "" {
			item.Isbn = &input.Isbn
		}
		if input.Year != 0 {
			item.Year = &input.Year
		}
		if input.Notes != "" {
			item.Notes = &input.Notes
		}
		if input.ItemType != "" {
			item.ItemType = &input.ItemType
		}
		if input.Venue != "" {
			item.Venue = &input.Venue
		}
		if input.Volume != "" {
			item.Volume = &input.Volume
		}
		if input.Number != "" {
			item.Number = &input.Number
		}
		if input.Pages != "" {
			item.Pages = &input.Pages
		}
		if input.Publisher != "" {
			item.Publisher = &input.Publisher
		}

		if err := e.Repo.UpdateItem(item); err != nil {
			e.Log.Error("UpdateItem failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not update item")
			return
		}
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

func APIDeleteItem(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		if item.FilePath != nil {
			if err := os.Remove(*item.FilePath); err != nil && !os.IsNotExist(err) {
				e.Log.Error("could not remove item file", "error", err, "path", *item.FilePath)
			}
		}
		if err := e.Repo.DeleteItem(item.UUID, env.UserFromContext(r.Context()).UUID); err != nil {
			e.Log.Error("DeleteItem failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not delete item")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// stringOrEmpty reads a *string's value, treating nil as "" - lets the
// transition table below compare against models.SyncState* without a nil
// check at every branch.
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptr(s string) *string { return &s }

// nextStateOnTick is what "I want this on the Kobo" (checking the box, or
// the skill/desktop app being told to sync an item) means for the current
// state. Already synced or already queued to sync is a no-op; a pending
// removal just gets cancelled, since the file's still actually there.
func nextStateOnTick(current *string) *string {
	switch stringOrEmpty(current) {
	case models.SyncStateRequestRemove:
		return ptr(models.SyncStateSynced)
	case models.SyncStateSynced, models.SyncStateRequestSync:
		return current
	default:
		return ptr(models.SyncStateRequestSync)
	}
}

// nextStateOnUntick is what "I don't want this on the Kobo anymore" means.
// A request that never actually got copied is simply cancelled; an item
// that's really on the device needs a request_remove so a poller goes and
// deletes the file, rather than the app just forgetting about it.
func nextStateOnUntick(current *string) *string {
	switch stringOrEmpty(current) {
	case models.SyncStateSynced:
		return ptr(models.SyncStateRequestRemove)
	case models.SyncStateRequestSync:
		return nil
	default:
		return current
	}
}

// APIRequestSync handles "tick the box" - queues the item to be copied to
// the Kobo (or cancels a pending removal), for whichever poller (browser
// Sync screen, the skill, a future desktop app) picks up request_sync
// next.
func APIRequestSync(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		next := nextStateOnTick(item.SyncState)
		if err := e.Repo.SetItemSyncState(item.UUID, env.UserFromContext(r.Context()).UUID, next); err != nil {
			e.Log.Error("SetItemSyncState failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not request sync")
			return
		}
		item.SyncState = next
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

// APICancelSync handles "untick the box" - either cancels a sync that
// hasn't happened yet, or queues removal of a file that's actually on the
// device.
func APICancelSync(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		next := nextStateOnUntick(item.SyncState)
		if err := e.Repo.SetItemSyncState(item.UUID, env.UserFromContext(r.Context()).UUID, next); err != nil {
			e.Log.Error("SetItemSyncState failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not cancel sync")
			return
		}
		item.SyncState = next
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

// APIAckSync is how a poller (browser, skill, desktop app) reports back
// after actually touching the device: "synced" once a request_sync item's
// file is copied, "removed" once a request_remove item's file is deleted.
// It's a plain report, not itself a transition decision - the caller
// already did the real work, this just records it.
func APIAckSync(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		var input struct {
			Result string `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "could not parse body")
			return
		}

		var next *string
		switch input.Result {
		case "synced":
			next = ptr(models.SyncStateSynced)
		case "removed":
			next = nil
		default:
			writeError(w, http.StatusBadRequest, `result must be "synced" or "removed"`)
			return
		}

		if err := e.Repo.SetItemSyncState(item.UUID, env.UserFromContext(r.Context()).UUID, next); err != nil {
			e.Log.Error("SetItemSyncState failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not ack sync")
			return
		}
		item.SyncState = next
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

// APIUploadItemFile attaches (or replaces) the file on an existing item -
// the path for "the extension made a metadata-only entry because it
// couldn't fetch the PDF automatically, now attach the one I downloaded by
// hand." Multipart form with a single "file" part.
func APIUploadItemFile(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			writeError(w, http.StatusBadRequest, "could not parse form")
			return
		}

		oldPath := item.FilePath

		// true: this is exactly the "the extension couldn't fetch a PDF
		// automatically, so I downloaded and I'm attaching it by hand" path
		// - the existing title is very often just a weak fallback (a page
		// <title>, a bare filename) precisely because nothing better could
		// be scraped when the item was created. extractPdfTitle only ever
		// overrides when the file has a real, non-placeholder embedded
		// title, so this is low-risk even when the current title happens
		// to already be a good one.
		duplicate, err := attachUploadedFile(e, r, item, true)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if duplicate != nil && duplicate.UUID != item.UUID {
			writeError(w, http.StatusConflict, "this exact file is already attached to another item in your library")
			return
		}
		if item.FilePath == nil {
			writeError(w, http.StatusBadRequest, "no file was provided")
			return
		}

		if err := e.Repo.UpdateItemFile(item); err != nil {
			e.Log.Error("UpdateItemFile failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not save file reference")
			return
		}

		if oldPath != nil && *oldPath != *item.FilePath {
			if err := os.Remove(*oldPath); err != nil && !os.IsNotExist(err) {
				e.Log.Error("could not remove replaced item file", "error", err, "path", *oldPath)
			}
		}

		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

func APIGetItemFile(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		if item.FilePath == nil {
			writeError(w, http.StatusNotFound, "no file attached to this item")
			return
		}
		filename := item.Title + "." + safeExt(item.FileType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		http.ServeFile(w, r, *item.FilePath)
	}
}

func safeExt(fileType *string) string {
	if fileType == nil || *fileType == "" {
		return "bin"
	}
	return *fileType
}

func APIExportMarkdown(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		tags, _ := e.Repo.GetTagsForItem(item.UUID)
		tagNames := make([]string, len(tags))
		for i, t := range tags {
			tagNames[i] = t.Name
		}

		var b strings.Builder
		b.WriteString("---\n")
		fmt.Fprintf(&b, "title: %q\n", item.Title)
		if item.Authors != "" {
			fmt.Fprintf(&b, "authors: [%s]\n", item.Authors)
		}
		if item.Year != nil {
			fmt.Fprintf(&b, "year: %d\n", *item.Year)
		}
		if item.Doi != nil {
			fmt.Fprintf(&b, "doi: %s\n", *item.Doi)
		}
		if len(tagNames) > 0 {
			fmt.Fprintf(&b, "tags: [%s]\n", strings.Join(tagNames, ", "))
		}
		if item.SourceUrl != nil {
			fmt.Fprintf(&b, "source: %s\n", *item.SourceUrl)
		}
		fmt.Fprintf(&b, "added: %s\n", item.AddedDate)
		b.WriteString("---\n\n")

		citation := item.Authors
		if item.Year != nil {
			citation += fmt.Sprintf(" (%d)", *item.Year)
		}
		citation += fmt.Sprintf(". *%s*.", item.Title)
		if item.Doi != nil {
			citation += fmt.Sprintf(" https://doi.org/%s", *item.Doi)
		}
		fmt.Fprintf(&b, "> %s\n", strings.TrimSpace(citation))

		if item.Notes != nil && *item.Notes != "" {
			b.WriteString("\n## Notes\n\n")
			b.WriteString(*item.Notes)
			b.WriteString("\n")
		}

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, sanitizeFilename(item.Title)))
		w.Write([]byte(b.String()))
	}
}

func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-")
	return replacer.Replace(s)
}

func APIAddItemTag(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		if err := e.Repo.AttachTag(item.UUID, input.Name); err != nil {
			e.Log.Error("AttachTag failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not attach tag")
			return
		}
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

func APIRemoveItemTag(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}
		tagId := mux.Vars(r)["tagId"]
		if err := e.Repo.DetachTag(item.UUID, tagId); err != nil {
			e.Log.Error("DetachTag failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not remove tag")
			return
		}
		writeJSON(w, http.StatusOK, itemResponse(e, item))
	}
}

// parseSyncStateParam validates ?sync_state= against the known states,
// returning "" (no filter) for anything else - including empty, and
// including typos, rather than silently matching everything.
func parseSyncStateParam(v string) string {
	switch v {
	case models.SyncStateRequestSync, models.SyncStateSynced, models.SyncStateRequestRemove:
		return v
	default:
		return ""
	}
}
