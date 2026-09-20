package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"

	"gitlab.com/robrohan/rigormining/internals/env"
)

// skillFiles maps the {name} path var to the zip built by the root
// Makefile's skills target - an allowlist, not a passthrough, so this can
// never be used to read an arbitrary file off Skills.Dir.
var skillFiles = map[string]string{
	"kobo-sync": "kobo-sync.zip",
}

// APIDownloadSkill serves a pre-built agent-skill zip for the "Get
// Extension" page - someone can download it without cloning the repo and
// drop the unzipped folder into ~/.agents/skills (see .agents/skills/
// kobo-sync/SKILL.md and the root Makefile for how the zip is built).
func APIDownloadSkill(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		filename, ok := skillFiles[name]
		if !ok {
			writeError(w, http.StatusNotFound, "unknown skill")
			return
		}

		path := filepath.Join(e.Cfg.Skills.Dir, filename)
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.Header().Set("Content-Type", "application/zip")
		http.ServeFile(w, r, path)
	}
}
