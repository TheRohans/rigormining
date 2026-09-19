// import-zotero reads a Zotero data directory (zotero.sqlite + storage/)
// directly and imports every non-trashed item into a running rigormining
// server over its normal HTTP API - the same POST /api/v1/items and
// POST /api/v1/items/{id}/tags the browser extension already uses,
// authenticated with an API token from the Settings page. There is
// deliberately no direct database/repository access here: going through
// the API means this gets the existing file-hash dedup and validation for
// free, and the script works against any running instance, not just one
// with local DB access.
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// itemMetadata mirrors backend/internals/handlers/items.go's
// itemMetadataInput - keep the two in sync if that struct changes.
type itemMetadata struct {
	Title     string `json:"title"`
	Authors   string `json:"authors,omitempty"`
	Doi       string `json:"doi,omitempty"`
	Isbn      string `json:"isbn,omitempty"`
	Year      int    `json:"year,omitempty"`
	SourceUrl string `json:"source_url,omitempty"`
	Notes     string `json:"notes,omitempty"`
	ItemType  string `json:"item_type,omitempty"`
	Venue     string `json:"venue,omitempty"`
	Volume    string `json:"volume,omitempty"`
	Number    string `json:"number,omitempty"`
	Pages     string `json:"pages,omitempty"`
	Publisher string `json:"publisher,omitempty"`
}

type zoteroItem struct {
	itemID   int64
	key      string
	typeName string
}

var yearPattern = regexp.MustCompile(`(19|20)\d{2}`)

func main() {
	zoteroDir := flag.String("zotero-dir", "", "path to the Zotero data directory (contains zotero.sqlite and storage/)")
	server := flag.String("server", "http://localhost:3000", "base URL of the running rigormining server")
	token := flag.String("token", "", "API token, created on the app's Settings page")
	limit := flag.Int("limit", 0, "import at most N items (0 = no limit) - useful for a dry run before importing everything")
	flag.Parse()

	if *zoteroDir == "" || *token == "" {
		fmt.Println("usage: import-zotero --zotero-dir=/path/to/Zotero --token=... [--server=http://localhost:3000]")
		os.Exit(1)
	}

	// Zotero (and Dropbox, if it's synced) may have zotero.sqlite open -
	// never touch the real file, work off a throwaway copy.
	tmpDB, err := copyToTemp(filepath.Join(*zoteroDir, "zotero.sqlite"))
	if err != nil {
		log.Fatalf("could not copy zotero.sqlite: %v", err)
	}
	defer os.Remove(tmpDB)

	db, err := sql.Open("sqlite3", "file:"+tmpDB+"?mode=ro")
	if err != nil {
		log.Fatalf("could not open zotero database: %v", err)
	}
	defer db.Close()

	items, err := loadItems(db)
	if err != nil {
		log.Fatalf("could not read items: %v", err)
	}
	log.Printf("found %d non-trashed items", len(items))
	if *limit > 0 && *limit < len(items) {
		items = items[:*limit]
		log.Printf("--limit=%d set, only importing the first %d", *limit, *limit)
	}

	var created, existing, noFile, failed int
	for i, it := range items {
		meta, attachmentPath, err := buildMetadata(db, *zoteroDir, it)
		if err != nil {
			log.Printf("[%d/%d] SKIP %s (%s): %v", i+1, len(items), it.key, it.typeName, err)
			failed++
			continue
		}
		if attachmentPath == "" {
			noFile++
		}

		id, isNew, err := postItem(*server, *token, meta, attachmentPath)
		if err != nil {
			log.Printf("[%d/%d] FAILED %q: %v", i+1, len(items), meta.Title, err)
			failed++
			continue
		}
		if isNew {
			created++
		} else {
			existing++
		}
		log.Printf("[%d/%d] %s %q", i+1, len(items), statusLabel(isNew), meta.Title)

		tags, err := loadTags(db, it.itemID)
		if err != nil {
			log.Printf("  could not load tags: %v", err)
			continue
		}
		for _, tag := range tags {
			if err := postTag(*server, *token, id, tag); err != nil {
				log.Printf("  could not attach tag %q: %v", tag, err)
			}
		}
	}

	fmt.Printf(
		"\nDone: %d created, %d already existed, %d with no local file, %d failed (of %d total)\n",
		created, existing, noFile, failed, len(items),
	)
}

func statusLabel(isNew bool) string {
	if isNew {
		return "created"
	}
	return "duplicate, skipped file"
}

func copyToTemp(path string) (string, error) {
	src, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer src.Close()

	tmp, err := os.CreateTemp("", "zotero-import-*.sqlite")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, src); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

func loadItems(db *sql.DB) ([]zoteroItem, error) {
	rows, err := db.Query(`
		SELECT i.itemID, i.key, it.typeName
		FROM items i
		JOIN itemTypes it ON it.itemTypeID = i.itemTypeID
		WHERE i.itemID NOT IN (SELECT itemID FROM deletedItems)
		  AND it.typeName NOT IN ('attachment', 'note', 'annotation')
		ORDER BY i.itemID
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []zoteroItem
	for rows.Next() {
		var it zoteroItem
		if err := rows.Scan(&it.itemID, &it.key, &it.typeName); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// loadFields returns every itemData field for this item, keyed by field
// name - fieldsCombined (rather than plain "fields") includes any
// custom/type-specific fields too.
func loadFields(db *sql.DB, itemID int64) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT f.fieldName, idv.value
		FROM itemData id
		JOIN fieldsCombined f ON f.fieldID = id.fieldID
		JOIN itemDataValues idv ON idv.valueID = id.valueID
		WHERE id.itemID = ?
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		out[name] = value
	}
	return out, rows.Err()
}

// loadAuthors formats authors as "Last, First; Last, First" - matches
// this app's existing convention (see ItemDetail.tsx's placeholder text).
func loadAuthors(db *sql.DB, itemID int64) (string, error) {
	rows, err := db.Query(`
		SELECT c.firstName, c.lastName
		FROM itemCreators ic
		JOIN creators c ON c.creatorID = ic.creatorID
		JOIN creatorTypes ct ON ct.creatorTypeID = ic.creatorTypeID
		WHERE ic.itemID = ? AND ct.creatorType = 'author'
		ORDER BY ic.orderIndex
	`, itemID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var first, last sql.NullString
		if err := rows.Scan(&first, &last); err != nil {
			return "", err
		}
		name := last.String
		if first.String != "" {
			name += ", " + first.String
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, "; "), rows.Err()
}

// loadTags merges Zotero tags and collection names into one deduped list
// - both become plain tags in this app, there's no separate "collections"
// concept in the UI.
func loadTags(db *sql.DB, itemID int64) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		out = append(out, name)
	}

	tagRows, err := db.Query(`SELECT t.name FROM itemTags it JOIN tags t ON t.tagID = it.tagID WHERE it.itemID = ?`, itemID)
	if err != nil {
		return nil, err
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var name string
		if err := tagRows.Scan(&name); err != nil {
			return nil, err
		}
		add(name)
	}
	if err := tagRows.Err(); err != nil {
		return nil, err
	}

	collRows, err := db.Query(`
		SELECT c.collectionName FROM collectionItems ci
		JOIN collections c ON c.collectionID = ci.collectionID
		WHERE ci.itemID = ?
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer collRows.Close()
	for collRows.Next() {
		var name string
		if err := collRows.Scan(&name); err != nil {
			return nil, err
		}
		add(name)
	}
	return out, collRows.Err()
}

// loadAttachmentPath finds the best attachment (PDF preferred over EPUB
// over anything else) and resolves it to a real file on disk. Zotero's
// itemAttachments.path for an imported file looks like
// "storage:actual filename.pdf" - the file itself lives at
// storage/<the attachment's own item key>/<that filename>, confirmed
// against a real file while planning this. Anything else (a linked file
// elsewhere on disk, a URL-only attachment, a missing file) comes back as
// "" - the item still imports, just without a file, same as the app's
// existing "bookmark now, attach later" case.
func loadAttachmentPath(db *sql.DB, zoteroDir string, parentItemID int64) (string, error) {
	row := db.QueryRow(`
		SELECT ia.path, i2.key
		FROM itemAttachments ia
		JOIN items i2 ON i2.itemID = ia.itemID
		WHERE ia.parentItemID = ?
		ORDER BY CASE ia.contentType
			WHEN 'application/pdf' THEN 0
			WHEN 'application/epub+zip' THEN 1
			ELSE 2
		END
		LIMIT 1
	`, parentItemID)

	var path, key sql.NullString
	if err := row.Scan(&path, &key); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if !path.Valid || !strings.HasPrefix(path.String, "storage:") {
		return "", nil
	}

	full := filepath.Join(zoteroDir, "storage", key.String, strings.TrimPrefix(path.String, "storage:"))
	if _, err := os.Stat(full); err != nil {
		return "", nil
	}
	return full, nil
}

// firstNonEmpty picks the first non-blank field value - different Zotero
// item types use different field names for what's conceptually the same
// "venue"/"issue"/"publisher" slot (e.g. journalArticle's
// "publicationTitle" vs. bookSection's "bookTitle").
func firstNonEmpty(fields map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(fields[k]); v != "" {
			return v
		}
	}
	return ""
}

func buildMetadata(db *sql.DB, zoteroDir string, it zoteroItem) (itemMetadata, string, error) {
	fields, err := loadFields(db, it.itemID)
	if err != nil {
		return itemMetadata{}, "", fmt.Errorf("load fields: %w", err)
	}
	authors, err := loadAuthors(db, it.itemID)
	if err != nil {
		return itemMetadata{}, "", fmt.Errorf("load authors: %w", err)
	}
	attachmentPath, err := loadAttachmentPath(db, zoteroDir, it.itemID)
	if err != nil {
		return itemMetadata{}, "", fmt.Errorf("load attachment: %w", err)
	}

	title := fields["title"]
	if title == "" {
		title = fmt.Sprintf("(untitled %s - %s)", it.typeName, it.key)
	}

	meta := itemMetadata{
		Title:     title,
		Authors:   authors,
		Doi:       fields["DOI"],
		Isbn:      fields["ISBN"],
		SourceUrl: fields["url"],
		Notes:     fields["abstractNote"],
		ItemType:  it.typeName,
		Venue:     firstNonEmpty(fields, "publicationTitle", "proceedingsTitle", "bookTitle", "conferenceName", "series"),
		Volume:    fields["volume"],
		Number:    firstNonEmpty(fields, "issue", "number"),
		Pages:     fields["pages"],
		Publisher: firstNonEmpty(fields, "publisher", "university"),
	}
	if y := yearPattern.FindString(fields["date"]); y != "" {
		if year, err := strconv.Atoi(y); err == nil {
			meta.Year = year
		}
	}

	return meta, attachmentPath, nil
}

// postItem does what the browser extension already does:
// POST /api/v1/items with a "metadata" JSON part and an optional "file"
// part. The API returns 201 for a genuinely new item and 200 when the
// file's hash already matched an existing item (see attachUploadedFile in
// backend/internals/handlers/items.go) - that status code is how the
// caller tells "created" from "duplicate" apart, no separate check needed.
func postItem(server, token string, meta itemMetadata, attachmentPath string) (id string, isNew bool, err error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return "", false, err
	}
	if err := w.WriteField("metadata", string(metaJSON)); err != nil {
		return "", false, err
	}

	if attachmentPath != "" {
		f, err := os.Open(attachmentPath)
		if err != nil {
			return "", false, fmt.Errorf("open attachment: %w", err)
		}
		defer f.Close()

		part, err := w.CreateFormFile("file", filepath.Base(attachmentPath))
		if err != nil {
			return "", false, err
		}
		if _, err := io.Copy(part, f); err != nil {
			return "", false, err
		}
	}
	if err := w.Close(); err != nil {
		return "", false, err
	}

	req, err := http.NewRequest(http.MethodPost, server+"/api/v1/items", body)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", false, fmt.Errorf("could not parse response: %w", err)
	}
	return out.ID, resp.StatusCode == http.StatusCreated, nil
}

func postTag(server, token, itemID, tagName string) error {
	body, err := json.Marshal(map[string]string{"name": tagName})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, server+"/api/v1/items/"+itemID+"/tags", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
