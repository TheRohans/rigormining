package handlers

import (
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// extractPdfTitle reads a saved PDF's embedded document-info Title, when
// present and non-trivial. Many real-world PDFs have no useful embedded
// title at all (or a generic placeholder) - that's normal, not an error,
// so this returns "" rather than failing the caller either way.
func extractPdfTitle(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	info, err := api.PDFInfo(f, filePath, nil, false, nil)
	if err != nil || info == nil {
		return ""
	}

	title := strings.TrimSpace(info.Title)
	if title == "" || strings.EqualFold(title, "untitled") {
		return ""
	}
	return title
}

// bestPdfTitle tries the PDF's embedded document-info Title first (cheap,
// exact when present), and falls back to guessing from the largest text on
// page 1 (see extractPdfTitleFromContent) when that metadata is blank -
// common for LaTeX-generated PDFs, including a lot of arXiv preprints,
// whose authors never set \hypersetup{pdftitle=...}.
func bestPdfTitle(filePath string) string {
	if title := extractPdfTitle(filePath); title != "" {
		return title
	}
	return extractPdfTitleFromContent(filePath)
}
