package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
)

// bibtexEntryType maps the free-text ItemType (whatever the source called
// it - Zotero's typeName, or hand-entered) onto a BibTeX entry type. Kept
// as the only place that needs to know what ItemType values mean, so the
// importer/rest of the app can treat ItemType as an opaque label.
func bibtexEntryType(itemType *string) string {
	if itemType == nil {
		return "misc"
	}
	switch strings.ToLower(*itemType) {
	case "journalarticle", "article", "magazinearticle", "newspaperarticle":
		return "article"
	case "conferencepaper":
		return "inproceedings"
	case "book":
		return "book"
	case "booksection":
		return "incollection"
	case "thesis":
		return "phdthesis"
	case "report":
		return "techreport"
	default:
		return "misc"
	}
}

// diacriticFold does a best-effort ASCII fold of the common Latin
// accented letters so a generated citekey stays plain alphanumeric
// without needing a transliteration library - a citekey just needs to be
// valid BibTeX syntax, not a perfect rendering of the name.
var diacriticFold = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ä", "a", "ã", "a", "å", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "ö", "o", "õ", "o", "ø", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ñ", "n", "ç", "c", "ß", "ss",
	"Á", "A", "À", "A", "Â", "A", "Ä", "A", "Ã", "A", "Å", "A",
	"É", "E", "È", "E", "Ê", "E", "Ë", "E",
	"Í", "I", "Ì", "I", "Î", "I", "Ï", "I",
	"Ó", "O", "Ò", "O", "Ô", "O", "Ö", "O", "Õ", "O", "Ø", "O",
	"Ú", "U", "Ù", "U", "Û", "U", "Ü", "U",
	"Ñ", "N", "Ç", "C",
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// slugWord lowercases, ASCII-folds, and strips anything that isn't a
// letter/digit - used to build a citekey component from a name or title
// word.
func slugWord(s string) string {
	return nonAlnum.ReplaceAllString(strings.ToLower(diacriticFold.Replace(s)), "")
}

// bibtexCiteKey generates a citekey in the common "lastname+year+word"
// shape (e.g. "hormann2020portable"). There's no universal citekey
// without a plugin like Better BibTeX, so this generates one rather than
// depending on the source having assigned one - uniqueness only matters
// if this app ever exports a whole library at once, which it doesn't yet.
func bibtexCiteKey(item *models.LibraryItem) string {
	lastName := "unknown"
	if firstAuthor := strings.SplitN(item.Authors, ";", 2)[0]; firstAuthor != "" {
		if last := strings.SplitN(firstAuthor, ",", 2)[0]; slugWord(last) != "" {
			lastName = slugWord(last)
		}
	}

	year := ""
	if item.Year != nil {
		year = strconv.Itoa(*item.Year)
	}

	titleWord := ""
	for _, word := range strings.Fields(item.Title) {
		if w := slugWord(word); len(w) > 3 {
			titleWord = w
			break
		}
	}

	return lastName + year + titleWord
}

// bibtexAuthors turns this app's "Last, First; Last, First" convention
// into BibTeX's "Last, First and Last, First".
func bibtexAuthors(authors string) string {
	parts := strings.Split(authors, ";")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return strings.Join(parts, " and ")
}

func bibtexField(b *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, "  %s = {%s},\n", name, value)
}

// APIExportBibtex serves a single item as a .bib file - GET /api/v1/items/{id}/export.bib.
func APIExportBibtex(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}

		entryType := bibtexEntryType(item.ItemType)
		citeKey := bibtexCiteKey(item)

		var b strings.Builder
		fmt.Fprintf(&b, "@%s{%s,\n", entryType, citeKey)
		bibtexField(&b, "title", item.Title)
		if item.Authors != "" {
			bibtexField(&b, "author", bibtexAuthors(item.Authors))
		}
		if item.Year != nil {
			bibtexField(&b, "year", strconv.Itoa(*item.Year))
		}
		if item.Venue != nil {
			if entryType == "article" {
				bibtexField(&b, "journal", *item.Venue)
			} else {
				bibtexField(&b, "booktitle", *item.Venue)
			}
		}
		if item.Volume != nil {
			bibtexField(&b, "volume", *item.Volume)
		}
		if item.Number != nil {
			bibtexField(&b, "number", *item.Number)
		}
		if item.Pages != nil {
			bibtexField(&b, "pages", *item.Pages)
		}
		if item.Publisher != nil {
			if entryType == "phdthesis" {
				bibtexField(&b, "school", *item.Publisher)
			} else {
				bibtexField(&b, "publisher", *item.Publisher)
			}
		}
		if item.Doi != nil {
			bibtexField(&b, "doi", *item.Doi)
		}
		if item.Isbn != nil {
			bibtexField(&b, "isbn", *item.Isbn)
		}
		if item.SourceUrl != nil {
			bibtexField(&b, "url", *item.SourceUrl)
		}
		if item.Notes != nil {
			bibtexField(&b, "note", *item.Notes)
		}
		b.WriteString("}\n")

		w.Header().Set("Content-Type", "application/x-bibtex; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.bib"`, sanitizeFilename(item.Title)))
		w.Write([]byte(b.String()))
	}
}
