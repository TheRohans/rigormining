package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/repository"
)

// APIListTags returns the user's tags with item counts - GET /api/v1/tags.
// The UI caches this for tag autocomplete.
func APIListTags(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, err := e.Repo.ListTags(env.UserFromContext(r.Context()).UUID)
		if err != nil {
			e.Log.Error("ListTags failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list tags")
			return
		}
		writeJSON(w, http.StatusOK, tags)
	}
}

// APIRenameTag renames one of the user's tags across all their items -
// POST /api/v1/tags/rename with {"from": "music", "to": "Music"}. If "to"
// is already a tag, the two are merged. Returns the updated tag list. The
// names go in the body rather than the URL since tags can contain "/".
func APIRenameTag(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			From string `json:"from"`
			To   string `json:"to"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		to := strings.TrimSpace(input.To)
		if input.From == "" || to == "" {
			writeError(w, http.StatusBadRequest, "from and to are required")
			return
		}

		userID := env.UserFromContext(r.Context()).UUID
		if to != input.From {
			_, err := e.Repo.RenameTag(userID, input.From, to)
			if errors.Is(err, repository.ErrTagNotFound) {
				writeError(w, http.StatusNotFound, "no items have that tag")
				return
			}
			if err != nil {
				e.Log.Error("RenameTag failed", "error", err)
				writeError(w, http.StatusInternalServerError, "could not rename tag")
				return
			}
		}

		tags, err := e.Repo.ListTags(userID)
		if err != nil {
			e.Log.Error("ListTags failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list tags")
			return
		}
		writeJSON(w, http.StatusOK, tags)
	}
}
