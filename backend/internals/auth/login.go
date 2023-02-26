package auth

import (
	"encoding/base64"

	"github.com/gorilla/mux"

	"net/http"
	"time"

	"gitlab.com/robrohan/knotset/internals/models"
	"gitlab.com/robrohan/knotset/internals/repository"
)

// addCookie will apply a new cookie to the response of a http request
// with the key/value specified.
func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) *http.Cookie {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expire,
		HttpOnly: true,
		Secure:   false, // only if https!
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	// http.SetCookie(w, &cookie)
	return &cookie
}

func SessionFromCookie(cookie *http.Cookie) (*models.Session, error) {
	session := models.Session{}
	// json.Unmarshal(data, &session)
	return &session, nil
}

// JwtVerify make sure jwt exists in the header
func JwtVerify(env *models.Env) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read cookie
			cookie, err := r.Cookie("HL")
			if err != nil {
				// w.WriteHeader(http.StatusForbidden)
				// json.NewEncoder(w).Encode("Missing auth token")
				env.Log.Printf("Missing auth token")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			if cookie.Value == "" {
				// w.WriteHeader(http.StatusForbidden)
				// json.NewEncoder(w).Encode("Missing auth token")
				env.Log.Printf("Missing auth token")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			if err != nil {
				// w.WriteHeader(http.StatusForbidden)
				// json.NewEncoder(w).Encode("Could not sort out the token")
				env.Log.Printf("Could not sort out the token")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Login handle login form posts
func Login(env *models.Env, repo *repository.ResearcherRepository) http.HandlerFunc {
	// Login handles login scenario.
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		alterego := r.Form.Get("alterego")
		password := r.Form.Get("password")

		_ = map[string]*string{
			"USERNAME": &alterego,
			"PASSWORD": &password,
		}

		token := base64.URLEncoding.EncodeToString([]byte(alterego))
		athleteId := token

		repo.UpsertResearcher(athleteId, "", alterego)

		// Logged in cookie
		cookie := addCookie(w, "KS", token, 30*24*time.Hour)
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/-/home", http.StatusMovedPermanently)
	}
}
