package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"

	"gitlab.com/robrohan/rigormining/internals/env"
)

// extensionFiles maps the {browser} path var to the zip built by
// extension/Makefile - an allowlist, not a passthrough, so this can never
// be used to read an arbitrary file off Extension.Dir.
var extensionFiles = map[string]string{
	"chrome":  "rigormining-chrome.zip",
	"firefox": "rigormining-firefox.zip",
}

// APIDownloadExtension serves the pre-built extension zip for the "Get
// Extension" page. Gated behind APILoginVerify (same as the rest of
// /api/v1) since this is beta/pre-store distribution, not a public
// download - see extension/Makefile for how the zips are built and the
// root Dockerfile for how they end up on Extension.Dir at runtime.
func APIDownloadExtension(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		browser := mux.Vars(r)["browser"]
		filename, ok := extensionFiles[browser]
		if !ok {
			writeError(w, http.StatusNotFound, "unknown browser")
			return
		}

		path := filepath.Join(e.Cfg.Extension.Dir, filename)
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.Header().Set("Content-Type", "application/zip")
		http.ServeFile(w, r, path)
	}
}
