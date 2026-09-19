package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
)

// APIWhoAmI lets the SPA (after a cookie login) and the extension/setup
// screen (after pasting a token) confirm who they're authenticated as.
func APIWhoAmI(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, env.UserFromContext(r.Context()))
	}
}

func APIListTokens(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokens, err := e.Repo.GetTokensByUserId(env.UserFromContext(r.Context()).UUID)
		if err != nil {
			e.Log.Error("GetTokensByUserId failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not list tokens")
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

func APICreateToken(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		raw := fmt.Sprintf("%x%x", rand.Int63(), rand.Int63())
		token := models.Token{
			UUID:      uuid.New().String(),
			UserId:    env.UserFromContext(r.Context()).UUID,
			Name:      input.Name,
			Value:     raw,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err := e.Repo.CreateToken(&token); err != nil {
			e.Log.Error("CreateToken failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not create token")
			return
		}
		// This is the only time the raw value is ever returned.
		writeJSON(w, http.StatusCreated, token)
	}
}

func APIDeleteToken(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if err := e.Repo.DeleteToken(id, env.UserFromContext(r.Context()).UUID); err != nil {
			e.Log.Error("DeleteToken failed", "error", err)
			writeError(w, http.StatusInternalServerError, "could not delete token")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
