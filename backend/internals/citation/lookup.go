package citation

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Metadata is a looked-up citation, in the same shape as a library item's
// bibliographic fields. Authors use the app's "Last, First; Last, First"
// convention, and ItemType uses the Zotero-style names bibtex.go maps.
type Metadata struct {
	Title     string
	Authors   string
	Doi       string
	Isbn      string
	Year      int
	ItemType  string
	Venue     string
	Volume    string
	Number    string
	Pages     string
	Publisher string

	// Source names the service the record came from (crossref, arxiv,
	// openlibrary), and MatchedBy how the work was identified, e.g.
	// "DOI in PDF" - shown to the user so they can judge the match.
	Source    string
	MatchedBy string
}

// ErrNotFound means the service has no record for the identifier.
var ErrNotFound = errors.New("no record found")

// Client talks to the public bibliographic services. None need an API key.
type Client struct {
	HTTP *http.Client
	// Mailto, when set, is sent to Crossref so requests go to its "polite"
	// pool, which is faster and more reliable than the anonymous one.
	Mailto string

	// Base URLs, overridable in tests.
	CrossrefURL    string
	DataciteURL    string
	ArxivURL       string
	OpenLibraryURL string
}

func NewClient(httpClient *http.Client, mailto string) *Client {
	return &Client{
		HTTP:           httpClient,
		Mailto:         mailto,
		CrossrefURL:    "https://api.crossref.org",
		DataciteURL:    "https://api.datacite.org",
		ArxivURL:       "https://export.arxiv.org/api/query",
		OpenLibraryURL: "https://openlibrary.org",
	}
}

func (c *Client) get(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	ua := "RigorMining/1.0 (https://gitlab.com/robrohan/rigormining)"
	if c.Mailto != "" {
		ua = fmt.Sprintf("RigorMining/1.0 (https://gitlab.com/robrohan/rigormining; mailto:%s)", c.Mailto)
	}
	req.Header.Set("User-Agent", ua)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", u, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4<<20))
}

// -------------------------------------------------------------------------
// Crossref

type crossrefWork struct {
	DOI            string   `json:"DOI"`
	Type           string   `json:"type"`
	Title          []string `json:"title"`
	Subtitle       []string `json:"subtitle"`
	ContainerTitle []string `json:"container-title"`
	Publisher      string   `json:"publisher"`
	Volume         string   `json:"volume"`
	Issue          string   `json:"issue"`
	Page           string   `json:"page"`
	ISBN           []string `json:"ISBN"`
	Author         []struct {
		Given  string `json:"given"`
		Family string `json:"family"`
		Name   string `json:"name"`
	} `json:"author"`
	Issued struct {
		DateParts [][]*int `json:"date-parts"`
	} `json:"issued"`
}

// crossrefTypes maps Crossref work types onto the ItemType names
// bibtex.go understands. Anything else is left blank (BibTeX @misc).
var crossrefTypes = map[string]string{
	"journal-article":     "journalArticle",
	"proceedings-article": "conferencePaper",
	"book-chapter":        "bookSection",
	"book-section":        "bookSection",
	"book-part":           "bookSection",
	"book":                "book",
	"monograph":           "book",
	"edited-book":         "book",
	"reference-book":      "book",
	"dissertation":        "thesis",
	"report":              "report",
	"report-component":    "report",
	"posted-content":      "preprint",
}

func (w *crossrefWork) metadata() *Metadata {
	md := &Metadata{
		Doi:       w.DOI,
		ItemType:  crossrefTypes[w.Type],
		Volume:    w.Volume,
		Number:    w.Issue,
		Pages:     strings.ReplaceAll(w.Page, "–", "-"),
		Publisher: cleanText(w.Publisher),
		Source:    "crossref",
	}
	if len(w.Title) > 0 {
		md.Title = cleanText(w.Title[0])
		if len(w.Subtitle) > 0 && w.Subtitle[0] != "" {
			md.Title += ": " + cleanText(w.Subtitle[0])
		}
	}
	if len(w.ContainerTitle) > 0 {
		md.Venue = cleanText(w.ContainerTitle[0])
	}
	if len(w.ISBN) > 0 {
		md.Isbn = normalizeISBN(w.ISBN[0])
	}
	if len(w.Issued.DateParts) > 0 && len(w.Issued.DateParts[0]) > 0 && w.Issued.DateParts[0][0] != nil {
		md.Year = *w.Issued.DateParts[0][0]
	}

	var authors []string
	for _, a := range w.Author {
		switch {
		case a.Family != "" && a.Given != "":
			authors = append(authors, cleanText(a.Family)+", "+cleanText(a.Given))
		case a.Family != "":
			authors = append(authors, cleanText(a.Family))
		case a.Name != "":
			authors = append(authors, cleanText(a.Name))
		}
	}
	md.Authors = strings.Join(authors, "; ")
	return md
}

// CrossrefDOI fetches the record for a DOI.
func (c *Client) CrossrefDOI(ctx context.Context, doi string) (*Metadata, error) {
	// Escape everything a DOI can contain (#, ?, ;, <...) except its "/".
	u := c.CrossrefURL + "/works/" + strings.ReplaceAll(url.PathEscape(doi), "%2F", "/")
	if c.Mailto != "" {
		u += "?mailto=" + url.QueryEscape(c.Mailto)
	}
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Message crossrefWork `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("crossref: %w", err)
	}
	if len(resp.Message.Title) == 0 {
		return nil, ErrNotFound
	}
	return resp.Message.metadata(), nil
}

// CrossrefSearch returns Crossref's best few bibliographic matches for a
// free-text title. Callers must check the match themselves - Crossref
// always returns something.
func (c *Client) CrossrefSearch(ctx context.Context, title string) ([]*Metadata, error) {
	q := url.Values{}
	q.Set("query.bibliographic", title)
	q.Set("rows", "5")
	if c.Mailto != "" {
		q.Set("mailto", c.Mailto)
	}
	body, err := c.get(ctx, c.CrossrefURL+"/works?"+q.Encode())
	if err != nil {
		return nil, err
	}
	var resp struct {
		Message struct {
			Items []crossrefWork `json:"items"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("crossref: %w", err)
	}
	out := make([]*Metadata, 0, len(resp.Message.Items))
	for i := range resp.Message.Items {
		if len(resp.Message.Items[i].Title) > 0 {
			out = append(out, resp.Message.Items[i].metadata())
		}
	}
	return out, nil
}

// -------------------------------------------------------------------------
// DataCite - registers DOIs Crossref doesn't: arXiv's own (10.48550),
// Zenodo, figshare, many datasets and university repositories.

// dataciteTypes maps DataCite's resourceTypeGeneral onto ItemType names.
var dataciteTypes = map[string]string{
	"JournalArticle":       "journalArticle",
	"ConferencePaper":      "conferencePaper",
	"Book":                 "book",
	"BookChapter":          "bookSection",
	"Dissertation":         "thesis",
	"Report":               "report",
	"Preprint":             "preprint",
	"Text":                 "",
	"ConferenceProceeding": "book",
}

// DataciteDOI fetches the record for a DOI registered with DataCite.
func (c *Client) DataciteDOI(ctx context.Context, doi string) (*Metadata, error) {
	body, err := c.get(ctx, c.DataciteURL+"/dois/"+strings.ReplaceAll(url.PathEscape(doi), "%2F", "/"))
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data struct {
			Attributes struct {
				DOI    string `json:"doi"`
				Titles []struct {
					Title     string `json:"title"`
					TitleType string `json:"titleType"`
				} `json:"titles"`
				Creators []struct {
					Name       string `json:"name"`
					GivenName  string `json:"givenName"`
					FamilyName string `json:"familyName"`
				} `json:"creators"`
				Publisher       json.RawMessage `json:"publisher"`
				PublicationYear int             `json:"publicationYear"`
				Types           struct {
					ResourceTypeGeneral string `json:"resourceTypeGeneral"`
				} `json:"types"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("datacite: %w", err)
	}
	a := resp.Data.Attributes

	md := &Metadata{
		Doi:      a.DOI,
		Year:     a.PublicationYear,
		ItemType: dataciteTypes[a.Types.ResourceTypeGeneral],
		Source:   "datacite",
	}
	for _, t := range a.Titles {
		// The main title has no titleType; subtitles/translations do.
		if t.TitleType == "" {
			md.Title = cleanText(t.Title)
			break
		}
	}
	if md.Title == "" {
		return nil, ErrNotFound
	}
	// publisher is a plain string, or an object in newer API versions.
	var publisher struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(a.Publisher, &md.Publisher) != nil && json.Unmarshal(a.Publisher, &publisher) == nil {
		md.Publisher = publisher.Name
	}
	md.Publisher = cleanText(md.Publisher)

	var authors []string
	for _, cr := range a.Creators {
		switch {
		case cr.FamilyName != "" && cr.GivenName != "":
			authors = append(authors, cleanText(cr.FamilyName)+", "+cleanText(cr.GivenName))
		case cr.Name != "":
			authors = append(authors, cleanText(cr.Name))
		}
	}
	md.Authors = strings.Join(authors, "; ")
	return md, nil
}

// -------------------------------------------------------------------------
// arXiv

type arxivFeed struct {
	Entries []struct {
		ID         string `xml:"id"`
		Title      string `xml:"title"`
		Published  string `xml:"published"`
		DOI        string `xml:"http://arxiv.org/schemas/atom doi"`
		JournalRef string `xml:"http://arxiv.org/schemas/atom journal_ref"`
		Authors    []struct {
			Name string `xml:"name"`
		} `xml:"author"`
	} `xml:"entry"`
}

// ArxivRecord is an arXiv entry plus the DOI of the published version,
// when the authors have recorded one.
type ArxivRecord struct {
	Metadata
	PublishedDOI string
}

// Arxiv fetches an arXiv preprint by id (e.g. 2301.01234 or hep-th/9901001).
func (c *Client) Arxiv(ctx context.Context, id string) (*ArxivRecord, error) {
	body, err := c.get(ctx, c.ArxivURL+"?id_list="+url.QueryEscape(id))
	if err != nil {
		return nil, err
	}
	var feed arxivFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("arxiv: %w", err)
	}
	// An unknown id comes back as an entry whose id isn't an abs/ URL.
	if len(feed.Entries) == 0 || !strings.Contains(feed.Entries[0].ID, "arxiv.org/abs/") {
		return nil, ErrNotFound
	}
	e := feed.Entries[0]

	rec := &ArxivRecord{PublishedDOI: strings.TrimSpace(e.DOI)}
	rec.Title = cleanText(e.Title)
	rec.ItemType = "preprint"
	// arXiv registers a DataCite DOI for every preprint.
	rec.Doi = "10.48550/arXiv." + id
	rec.Venue = "arXiv preprint arXiv:" + id
	rec.Publisher = "arXiv"
	rec.Source = "arxiv"
	if len(e.Published) >= 4 {
		rec.Year, _ = strconv.Atoi(e.Published[:4])
	}
	var authors []string
	for _, a := range e.Authors {
		authors = append(authors, personName(cleanText(a.Name)))
	}
	rec.Authors = strings.Join(authors, "; ")
	return rec, nil
}

// -------------------------------------------------------------------------
// Open Library

var yearPattern = regexp.MustCompile(`\b(1[5-9]\d\d|20\d\d)\b`)

// OpenLibraryISBN fetches a book by ISBN.
func (c *Client) OpenLibraryISBN(ctx context.Context, isbn string) (*Metadata, error) {
	key := "ISBN:" + isbn
	body, err := c.get(ctx, c.OpenLibraryURL+"/api/books?format=json&jscmd=data&bibkeys="+url.QueryEscape(key))
	if err != nil {
		return nil, err
	}
	var resp map[string]struct {
		Title    string `json:"title"`
		Subtitle string `json:"subtitle"`
		Authors  []struct {
			Name string `json:"name"`
		} `json:"authors"`
		Publishers []struct {
			Name string `json:"name"`
		} `json:"publishers"`
		PublishDate string `json:"publish_date"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("openlibrary: %w", err)
	}
	book, ok := resp[key]
	if !ok || book.Title == "" {
		return nil, ErrNotFound
	}

	md := &Metadata{
		Title:    cleanText(book.Title),
		Isbn:     isbn,
		ItemType: "book",
		Source:   "openlibrary",
	}
	if book.Subtitle != "" {
		md.Title += ": " + cleanText(book.Subtitle)
	}
	if len(book.Publishers) > 0 {
		md.Publisher = cleanText(book.Publishers[0].Name)
	}
	if y := yearPattern.FindString(book.PublishDate); y != "" {
		md.Year, _ = strconv.Atoi(y)
	}
	var authors []string
	for _, a := range book.Authors {
		authors = append(authors, personName(cleanText(a.Name)))
	}
	md.Authors = strings.Join(authors, "; ")
	return md, nil
}
