package handlers

import (
	"testing"

	"gitlab.com/robrohan/rigormining/internals/models"
)

func TestUniqueCiteKey(t *testing.T) {
	used := map[string]bool{}
	want := []string{"hu2021lora", "hu2021loraa", "hu2021lorab"}
	for _, w := range want {
		if got := uniqueCiteKey("hu2021lora", used); got != w {
			t.Errorf("got %q, want %q", got, w)
		}
	}
	if got := uniqueCiteKey("lecun2015deep", used); got != "lecun2015deep" {
		t.Errorf("unrelated key changed: %q", got)
	}
}

func TestBibtexSuffix(t *testing.T) {
	for i, want := range map[int]string{0: "a", 25: "z", 26: "aa", 27: "ab", 51: "az", 52: "ba"} {
		if got := bibtexSuffix(i); got != want {
			t.Errorf("bibtexSuffix(%d) = %q, want %q", i, got, want)
		}
	}
}

func TestBibtexEscape(t *testing.T) {
	tests := map[string]string{
		"27.9% zero-shot, and 46.0% few-shot": `27.9\% zero-shot, and 46.0\% few-shot`,
		"Taylor & Francis":                    `Taylor \& Francis`,
		"C# and snake_case cost $5":           `C\# and snake\_case cost \$5`,
		"O(n^2) ~ fast":                       `O(n\^{}2) \~{} fast`,
		`a\b`:                                 `a\textbackslash{}b`,
		"{BERT}: Pre-training":                "{BERT}: Pre-training", // balanced braces kept
		"Broken {title":                       "Broken title",         // unbalanced braces dropped
		"Odd } one { out":                     "Odd  one  out",
		"Plain title":                         "Plain title",
	}
	for in, want := range tests {
		if got := bibtexEscape(in); got != want {
			t.Errorf("bibtexEscape(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBibtexEntry(t *testing.T) {
	str := func(s string) *string { return &s }
	year := 2022
	item := &models.LibraryItem{
		Title:     "Mind's Eye: 27.9% better",
		Authors:   "Liu, Ruibo; Wei, Jason",
		Year:      &year,
		Doi:       str("10.48550/arXiv.2210.05359"),
		SourceUrl: str("https://example.com/a%20b?x=1&y=2"),
		Notes:     str("A long abstract that should not be exported."),
	}
	got := bibtexEntry(item, "liu2022minds")
	want := "@misc{liu2022minds,\n" +
		"  title = {Mind's Eye: 27.9\\% better},\n" +
		"  author = {Liu, Ruibo and Wei, Jason},\n" +
		"  year = {2022},\n" +
		"  doi = {10.48550/arXiv.2210.05359},\n" +
		"  url = {https://example.com/a%20b?x=1&y=2},\n" +
		"}\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
