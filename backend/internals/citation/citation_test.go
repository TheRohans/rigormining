package citation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFindDOIs(t *testing.T) {
	tests := map[string][]string{
		"DOI: 10.1145/3290605.3300233.":                    {"10.1145/3290605.3300233"},
		"(https://doi.org/10.1038/nature14539)":            {"10.1038/nature14539"},
		"see 10.1016/S0140-6736(97)11096-0, also":          {"10.1016/S0140-6736(97)11096-0"},
		"chapter 10.1007/978-3-319-46493-0_38 in":          {"10.1007/978-3-319-46493-0_38"},
		"ref ⟨10.1145/1235⟩ end":                           {"10.1145/1235"},
		"two: 10.1000/a1 and 10.1000/b2; dup 10.1000/A1":   {"10.1000/a1", "10.1000/b2"},
		"no doi here, just 10.5 percent":                   nil,
		"doi:10.1109/CVPRW.2019.00320\nNext line of prose": {"10.1109/CVPRW.2019.00320"},
	}
	for in, want := range tests {
		if got := findDOIs(in); !reflect.DeepEqual(got, want) {
			t.Errorf("findDOIs(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestFindArxivIDs(t *testing.T) {
	tests := map[string][]string{
		"arXiv:2106.09685v2 [cs.CL] 16 Oct 2021": {"2106.09685"},
		"arXiv: hep-th/9901001":                  {"hep-th/9901001"},
		// The margin stamp laid out one glyph per line, bottom to top.
		"1\n2\n0\n2\n \nt\nc\nO\n \n6\n1\n \n]\nL\nC\n.\ns\nc\n[\n \n2\nv\n5\n8\n6\n9\n0\n.\n6\n0\n1\n2\n:\nv\ni\nX\nr\na": {"2106.09685"},
		"no identifiers": nil,
	}
	for in, want := range tests {
		if got := findArxivIDs(in); !reflect.DeepEqual(got, want) {
			t.Errorf("findArxivIDs(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNormalizeISBN(t *testing.T) {
	tests := map[string]string{
		"978-0-262-03384-8": "9780262033848",
		"0-262-03384-4":     "0262033844",
		"0-8044-2957-X":     "080442957X",
		"978-0-262-03384-9": "", // bad check digit
		"12345":             "",
	}
	for in, want := range tests {
		if got := normalizeISBN(in); got != want {
			t.Errorf("normalizeISBN(%q) = %q, want %q", in, got, want)
		}
	}
	if got := findISBNs("Copyright 2009. ISBN 978-0-262-03384-8 (hardcover)"); !reflect.DeepEqual(got, []string{"9780262033848"}) {
		t.Errorf("findISBNs = %v", got)
	}
}

func TestTitleInText(t *testing.T) {
	text := "Proceedings of X\nLoRA: Low-Rank Adaptation of Large\nLanguage Models\nEdward Hu, Yelong Shen"
	if !TitleInText("LoRA: Low-Rank Adaptation of Large Language Models", text) {
		t.Error("title spread over two lines should match")
	}
	// Text extracted with the spaces missing, and an "ﬁ" ligature.
	if !TitleInText("Refilming with Depth-Inferred Videos", "ReﬁlmingwithDepth-InferredVideos\nGuofeng Zhang") {
		t.Error("squashed text with a ligature should match")
	}
	if TitleInText("Deep Residual Learning for Image Recognition", text) {
		t.Error("an unrelated title should not match")
	}
}

func TestTitleSimilarity(t *testing.T) {
	if s := TitleSimilarity("Attention Is All You Need", "Attention is all you need."); s != 1 {
		t.Errorf("same title, different case/punctuation: %v", s)
	}
	if s := TitleSimilarity("Attention Is All You Need", "Attention in Graph Neural Networks"); s >= searchThreshold {
		t.Errorf("different titles scored %v", s)
	}
}

func TestPersonName(t *testing.T) {
	tests := map[string]string{
		"Ashish Vaswani":         "Vaswani, Ashish",
		"Laurens van der Maaten": "van der Maaten, Laurens",
		"Yann LeCun":             "LeCun, Yann",
		"J. R. R. Tolkien":       "Tolkien, J. R. R.",
		"Diego de las Casas":     "de las Casas, Diego",
		"Plato":                  "Plato",
	}
	for in, want := range tests {
		if got := personName(in); got != want {
			t.Errorf("personName(%q) = %q, want %q", in, got, want)
		}
	}
}

// -------------------------------------------------------------------------
// Lookups against a fake server

const crossrefLeCun = `{"message":{
	"DOI":"10.1038/nature14539","type":"journal-article",
	"title":["Deep <i>learning</i>"],"container-title":["Nature"],
	"publisher":"Springer Science and Business Media LLC",
	"volume":"521","issue":"7553","page":"436–444",
	"author":[{"given":"Yann","family":"LeCun"},{"given":"Yoshua","family":"Bengio"},{"name":"Some Consortium"}],
	"issued":{"date-parts":[[2015,5,27]]}}}`

const crossrefOther = `{"message":{"DOI":"10.1000/other","type":"journal-article","title":["A Paper That Page One Merely Cites"],"issued":{"date-parts":[[2001]]}}}`

const arxivLoRA = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:arxiv="http://arxiv.org/schemas/atom">
  <entry>
    <id>http://arxiv.org/abs/2106.09685v2</id>
    <published>2021-06-17T17:37:18Z</published>
    <title>LoRA: Low-Rank Adaptation of Large
      Language Models</title>
    <author><name>Edward J. Hu</name></author>
    <author><name>Yelong Shen</name></author>
  </entry>
</feed>`

const arxivUnknown = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom"><entry><id>http://arxiv.org/api/errors#incorrect_id_format</id><title>Error</title></entry></feed>`

const dataciteLoRA = `{"data":{"attributes":{"doi":"10.48550/arxiv.2106.09685",
	"titles":[{"title":"LoRA: Low-Rank Adaptation of Large Language Models"}],
	"creators":[{"name":"Hu, Edward J.","givenName":"Edward J.","familyName":"Hu"},{"name":"Shen, Yelong","givenName":"Yelong","familyName":"Shen"}],
	"publisher":"arXiv","publicationYear":2021,"types":{"resourceTypeGeneral":"Preprint"}}}}`

const openLibraryBook = `{"ISBN:9780262033848":{"title":"Introduction to Algorithms","authors":[{"name":"Thomas H. Cormen"}],"publishers":[{"name":"MIT Press"}],"publish_date":"July 31, 2009"}}`

func fakeServices(t *testing.T) (*Client, *[]string) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.RequestURI())
		switch {
		case r.URL.Path == "/works/10.1038/nature14539":
			w.Write([]byte(crossrefLeCun))
		case r.URL.Path == "/works/10.1000/other":
			w.Write([]byte(crossrefOther))
		case r.URL.Path == "/works" && strings.Contains(r.URL.RawQuery, "query.bibliographic"):
			w.Write([]byte(`{"message":{"items":[` + strings.TrimSuffix(strings.TrimPrefix(crossrefOther, `{"message":`), "}") + `,` +
				strings.TrimSuffix(strings.TrimPrefix(crossrefLeCun, `{"message":`), "}") + `]}}`))
		case r.URL.Path == "/dois/10.48550/arXiv.2106.09685" || r.URL.Path == "/dois/10.48550/arXiv.1111.11111":
			w.Write([]byte(dataciteLoRA))
		case r.URL.Path == "/arxiv" && r.URL.Query().Get("id_list") == "1111.11111":
			// arXiv being slow: longer than arxivWait.
			select {
			case <-time.After(arxivWait + time.Second):
			case <-r.Context().Done():
			}
			w.Write([]byte(arxivUnknown))
		case r.URL.Path == "/arxiv" && r.URL.Query().Get("id_list") == "2106.09685":
			w.Write([]byte(arxivLoRA))
		case r.URL.Path == "/arxiv":
			w.Write([]byte(arxivUnknown))
		case r.URL.Path == "/api/books" && strings.Contains(r.URL.RawQuery, "9780262033848"):
			w.Write([]byte(openLibraryBook))
		case r.URL.Path == "/api/books":
			w.Write([]byte(`{}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.Client(), "")
	c.CrossrefURL = srv.URL
	c.DataciteURL = srv.URL
	c.ArxivURL = srv.URL + "/arxiv"
	c.OpenLibraryURL = srv.URL
	return c, &calls
}

func TestCrossrefDOI(t *testing.T) {
	c, _ := fakeServices(t)
	md, err := c.CrossrefDOI(context.Background(), "10.1038/nature14539")
	if err != nil {
		t.Fatal(err)
	}
	want := &Metadata{
		Title: "Deep learning", Authors: "LeCun, Yann; Bengio, Yoshua; Some Consortium",
		Doi: "10.1038/nature14539", Year: 2015, ItemType: "journalArticle", Venue: "Nature",
		Volume: "521", Number: "7553", Pages: "436-444",
		Publisher: "Springer Science and Business Media LLC", Source: "crossref",
	}
	if !reflect.DeepEqual(md, want) {
		t.Errorf("got  %+v\nwant %+v", md, want)
	}

	if _, err := c.CrossrefDOI(context.Background(), "10.9999/missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing DOI: err = %v, want ErrNotFound", err)
	}
}

func TestArxiv(t *testing.T) {
	c, _ := fakeServices(t)
	rec, err := c.Arxiv(context.Background(), "2106.09685")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Title != "LoRA: Low-Rank Adaptation of Large Language Models" ||
		rec.Authors != "Hu, Edward J.; Shen, Yelong" || rec.Year != 2021 ||
		rec.Doi != "10.48550/arXiv.2106.09685" || rec.ItemType != "preprint" {
		t.Errorf("unexpected record %+v", rec.Metadata)
	}
	if _, err := c.Arxiv(context.Background(), "9999.99999"); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown id: err = %v, want ErrNotFound", err)
	}
}

func TestArxivFallsBackToDataciteWhenSlow(t *testing.T) {
	c, _ := fakeServices(t)
	start := time.Now()
	md, err := c.byArxiv(context.Background(), "1111.11111")
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d > arxivWait+500*time.Millisecond {
		t.Errorf("waited %v for a slow arXiv", d)
	}
	if md.Source != "datacite" || md.Authors != "Hu, Edward J.; Shen, Yelong" ||
		md.ItemType != "preprint" || md.Venue != "arXiv preprint arXiv:1111.11111" || md.Year != 2021 {
		t.Errorf("unexpected record %+v", md)
	}
}

func TestOpenLibraryISBN(t *testing.T) {
	c, _ := fakeServices(t)
	md, err := c.OpenLibraryISBN(context.Background(), "9780262033848")
	if err != nil {
		t.Fatal(err)
	}
	if md.Title != "Introduction to Algorithms" || md.Authors != "Cormen, Thomas H." ||
		md.Publisher != "MIT Press" || md.Year != 2009 || md.ItemType != "book" {
		t.Errorf("unexpected record %+v", md)
	}
}

// Enough words for usableText, containing the LeCun title.
var lecunPage = "Deep learning\nYann LeCun, Yoshua Bengio, Geoffrey Hinton\n" +
	strings.Repeat("representation learning methods allow machines discover patterns ", 20)

func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		hints     Hints
		wantTitle string
		wantBy    string
		wantErr   error
	}{
		{
			name:      "DOI on item wins",
			hints:     Hints{DOI: "10.1038/nature14539"},
			wantTitle: "Deep learning", wantBy: "DOI on item",
		},
		{
			name:      "DOI in text whose title is on the page",
			hints:     Hints{PDF: &PDFIdentifiers{DOIs: []string{"10.1038/nature14539"}, Text: lecunPage}},
			wantTitle: "Deep learning", wantBy: "DOI in PDF",
		},
		{
			name: "a cited paper's DOI is skipped for the paper's own",
			hints: Hints{PDF: &PDFIdentifiers{
				DOIs: []string{"10.1000/other", "10.1038/nature14539"}, Text: lecunPage,
			}},
			wantTitle: "Deep learning", wantBy: "DOI in PDF",
		},
		{
			name: "wrong metadata DOI is rejected when the text says otherwise",
			hints: Hints{PDF: &PDFIdentifiers{
				MetadataDOIs: []string{"10.1000/other", "10.1038/nature14539"}, Text: lecunPage,
			}},
			wantTitle: "Deep learning", wantBy: "DOI in PDF metadata",
		},
		{
			name:      "metadata DOI trusted when the text didn't extract",
			hints:     Hints{PDF: &PDFIdentifiers{MetadataDOIs: []string{"10.1000/other"}, Text: "2024 IEEE"}},
			wantTitle: "A Paper That Page One Merely Cites", wantBy: "DOI in PDF metadata",
		},
		{
			name:      "arXiv id",
			hints:     Hints{PDF: &PDFIdentifiers{ArxivIDs: []string{"2106.09685"}}},
			wantTitle: "LoRA: Low-Rank Adaptation of Large Language Models", wantBy: "arXiv ID in PDF",
		},
		{
			name:      "arXiv DOI on item goes to arXiv, not Crossref",
			hints:     Hints{DOI: "10.48550/arXiv.2106.09685"},
			wantTitle: "LoRA: Low-Rank Adaptation of Large Language Models", wantBy: "DOI on item",
		},
		{
			name:      "ISBN found in the book",
			hints:     Hints{PDF: &PDFIdentifiers{ISBNs: []string{"9780262033848"}}},
			wantTitle: "Introduction to Algorithms", wantBy: "ISBN",
		},
		{
			name:    "title too vague to search",
			hints:   Hints{Title: "Deep Learning", PDF: &PDFIdentifiers{Text: lecunPage}},
			wantErr: ErrNotFound, // under 3 significant words
		},
		{
			name:    "loose title search result rejected",
			hints:   Hints{Title: "Deep <i>learning</i> nature paper review"},
			wantErr: ErrNotFound, // similarity to "Deep learning" too low - better nothing than a guess
		},
		{
			name:    "nothing to go on",
			hints:   Hints{},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := fakeServices(t)
			md, err := c.Resolve(context.Background(), tt.hints)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v (md=%+v)", err, tt.wantErr, md)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if md.Title != tt.wantTitle || md.MatchedBy != tt.wantBy {
				t.Errorf("got %q by %q, want %q by %q", md.Title, md.MatchedBy, tt.wantTitle, tt.wantBy)
			}
		})
	}
}

func TestResolveTitleSearch(t *testing.T) {
	c, _ := fakeServices(t)
	// The fake search always returns the same two papers. A title only
	// loosely like one of them must not match - better no citation than a
	// wrong one - while a close match is accepted.
	md, err := c.Resolve(context.Background(), Hints{Title: "Deep Learning (LeCun Bengio Hinton)"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("loose title matched %+v", md)
	}
	c2, _ := fakeServices(t)
	md, err = c2.Resolve(context.Background(), Hints{Title: "A paper that page one merely cites"})
	if err != nil || md.MatchedBy != "title search" || md.Doi != "10.1000/other" {
		t.Fatalf("got %+v, %v", md, err)
	}
}
