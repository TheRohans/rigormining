package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
	"gitlab.com/robrohan/rigormining/internals/repository"
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
// depending on the source having assigned one. Keys only need to be
// unique within one exported file - see uniqueCiteKey.
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

// bibtexSpecials are the characters LaTeX treats as commands when the
// field is typeset - an unescaped % comments out the rest of the line,
// & is a "misplaced alignment tab" error, and so on.
var bibtexSpecials = strings.NewReplacer(
	`\`, `\textbackslash{}`,
	`%`, `\%`,
	`&`, `\&`,
	`#`, `\#`,
	`_`, `\_`,
	`$`, `\$`,
	`^`, `\^{}`,
	`~`, `\~{}`,
)

// bibtexEscape makes plain text safe for a BibTeX field. Braces are kept
// when balanced (a protected "{BERT}" is legitimate BibTeX), but dropped
// when not, since an unbalanced brace breaks parsing of the whole file -
// and BibTeX counts braces even when they're backslash-escaped.
func bibtexEscape(s string) string {
	depth, balanced := 0, true
	for _, r := range s {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
		}
		if depth < 0 {
			balanced = false
			break
		}
	}
	if !balanced || depth != 0 {
		s = strings.NewReplacer("{", "", "}", "").Replace(s)
	}
	return bibtexSpecials.Replace(s)
}

// bibtexField writes a text field, escaped for LaTeX.
func bibtexField(b *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, "  %s = {%s},\n", name, bibtexEscape(value))
}

// bibtexVerbatimField writes a field biblatex reads verbatim (doi, url),
// where escaping would put literal backslashes into the link.
func bibtexVerbatimField(b *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, "  %s = {%s},\n", name, value)
}

// uniqueCiteKey returns key, or key with a letter suffix (hu2021lora,
// hu2021loraa, hu2021lorab...) if it's already used, the usual way
// reference managers disambiguate same-author-same-year keys.
func uniqueCiteKey(key string, used map[string]bool) string {
	unique := key
	for i := 0; used[unique]; i++ {
		unique = key + bibtexSuffix(i)
	}
	used[unique] = true
	return unique
}

// bibtexSuffix is a, b, ... z, aa, ab, ...
func bibtexSuffix(i int) string {
	s := ""
	for i++; i > 0; i = (i - 1) / 26 {
		s = string(rune('a'+(i-1)%26)) + s
	}
	return s
}

// bibtexEntry renders one item as a BibTeX entry.
func bibtexEntry(item *models.LibraryItem, citeKey string) string {
	entryType := bibtexEntryType(item.ItemType)

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
		switch entryType {
		case "article":
			bibtexField(&b, "journal", *item.Venue)
		case "misc":
			// e.g. "arXiv preprint arXiv:2106.09685"
			bibtexField(&b, "howpublished", *item.Venue)
		default:
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
		bibtexVerbatimField(&b, "doi", *item.Doi)
	}
	if item.Isbn != nil {
		bibtexField(&b, "isbn", *item.Isbn)
	}
	if item.SourceUrl != nil {
		bibtexVerbatimField(&b, "url", *item.SourceUrl)
	}
	// Notes are deliberately left out: for Zotero imports they hold the
	// abstract, and most citation styles print "note" in the reference
	// list.
	b.WriteString("}\n")
	return b.String()
}

// APIExportBibtex serves a single item as a .bib file - GET /api/v1/items/{id}/export.bib.
func APIExportBibtex(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item := loadOwnedItem(e, w, r)
		if item == nil {
			return
		}

		w.Header().Set("Content-Type", "application/x-bibtex; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.bib"`, sanitizeFilename(item.Title)))
		w.Write([]byte(bibtexEntry(item, bibtexCiteKey(item))))
	}
}

// APIExportBibtexTag serves every item with a tag as one .bib file -
// GET /api/v1/items/export.bib?tag=<name>. Entries are sorted by citekey,
// and keys are made unique within the file.
func APIExportBibtexTag(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tag := r.URL.Query().Get("tag")
		if tag == "" {
			writeError(w, http.StatusBadRequest, "tag is required")
			return
		}

		items, err := e.Repo.ListItems(env.UserFromContext(r.Context()).UUID, repository.ItemFilter{Tag: tag})
		if err != nil {
			e.Log.Error("ListItems (bibtex export) failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list items")
			return
		}
		if len(items) == 0 {
			writeError(w, http.StatusNotFound, "no items with that tag")
			return
		}

		type entry struct {
			key  string
			item *models.LibraryItem
		}
		entries := make([]entry, len(items))
		for i := range items {
			entries[i] = entry{bibtexCiteKey(&items[i]), &items[i]}
		}
		sort.SliceStable(entries, func(i, j int) bool { return entries[i].key < entries[j].key })

		used := map[string]bool{}
		var b strings.Builder
		for i, en := range entries {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(bibtexEntry(en.item, uniqueCiteKey(en.key, used)))
		}

		w.Header().Set("Content-Type", "application/x-bibtex; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.bib"`, sanitizeFilename(tag)))
		w.Write([]byte(b.String()))
	}
}
