package models

import "github.com/google/uuid"

// UserInfo is the userinfo payload returned by the OAuth provider.
type UserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// User is a signed-in account.
type User struct {
	UUID    string  `db:"uuid"     json:"id"`
	Email   string  `db:"email"    json:"email"`
	Name    *string `db:"username" json:"name,omitempty"`
	Picture *string `db:"picture"  json:"picture,omitempty"`
	AuthId  string  `db:"authid"   json:"-"`
	Salt    *string `db:"salt"     json:"-"`
}

func NewUser(authid string, email string, picture string) *User {
	return &User{
		UUID:    uuid.New().String(),
		AuthId:  authid,
		Email:   email,
		Picture: &picture,
	}
}

// Token is an API bearer token belonging to a user. Used by the browser
// extension, which can't do a browser OAuth redirect flow.
type Token struct {
	UUID      string `db:"uuid"       json:"id"`
	UserId    string `db:"user_uuid"  json:"-"`
	Name      string `db:"name"       json:"name"`
	Value     string `db:"value"      json:"value,omitempty"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

// LibraryItem is a single paper/book in a user's library.
type LibraryItem struct {
	UUID        string  `db:"uuid"         json:"id"`
	UserId      string  `db:"user_uuid"    json:"-"`
	Title       string  `db:"title"        json:"title"`
	Authors     string  `db:"authors"      json:"authors"`
	Doi         *string `db:"doi"          json:"doi,omitempty"`
	Isbn        *string `db:"isbn"         json:"isbn,omitempty"`
	Year        *int    `db:"year"         json:"year,omitempty"`
	SourceUrl   *string `db:"source_url"   json:"source_url,omitempty"`
	FilePath    *string `db:"file_path"    json:"-"`
	FileType    *string `db:"file_type"    json:"file_type,omitempty"`
	FileHash    *string `db:"file_hash"    json:"-"`
	AddedDate   string  `db:"added_date"   json:"added_date"`
	DeliveredAt *string `db:"delivered_at" json:"delivered_at,omitempty"`
	Notes       *string `db:"notes"        json:"notes,omitempty"`
	// The following exist to make BibTeX export possible - see
	// handlers/bibtex.go. ItemType is free text (e.g. "journalArticle",
	// "book") rather than a fixed enum - it's whatever the source (Zotero
	// import, manual entry) called it, and the BibTeX-type mapping lives
	// entirely in bibtex.go so it's the only place that needs to know
	// what values are meaningful.
	ItemType  *string `db:"item_type" json:"item_type,omitempty"`
	Venue     *string `db:"venue"     json:"venue,omitempty"`
	Volume    *string `db:"volume"    json:"volume,omitempty"`
	Number    *string `db:"number"    json:"number,omitempty"`
	Pages     *string `db:"pages"     json:"pages,omitempty"`
	Publisher *string `db:"publisher" json:"publisher,omitempty"`
}

// Tag is a short label attachable to library items.
type Tag struct {
	UUID string `db:"uuid" json:"id"`
	Name string `db:"name" json:"name"`
}
