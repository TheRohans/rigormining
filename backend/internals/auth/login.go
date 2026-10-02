package auth

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
	"gitlab.com/robrohan/rigormining/internals/repository"
	"gitlab.com/robrohan/rigormining/internals/secret"
)

const (
	// CookieName holds the raw session id; the session table stores its hash.
	CookieName = "RM_AT"
	// stateCookieName holds the OAuth state for one in-flight login.
	stateCookieName = "RM_OS"

	sessionTTL = 30 * 24 * time.Hour
	stateTTL   = 10 * time.Minute
)

// NewOAuthConfig builds the provider config (Google by default) from the
// app's env-driven Config.
func NewOAuthConfig(cfg *models.Config) *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  cfg.Auth.RedirectURL,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
		Scopes:       cfg.Auth.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:   cfg.Auth.AuthURL,
			TokenURL:  cfg.Auth.TokenURL,
			AuthStyle: oauth2.AuthStyle(cfg.Auth.AuthStyle),
		},
	}
}

func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(ttl),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// startSession creates a server-side session for user and sets its cookie.
func startSession(w http.ResponseWriter, repo *repository.DataRepository, user *models.User) error {
	raw, err := repo.CreateSession(user.UUID, sessionTTL)
	if err != nil {
		return err
	}
	addCookie(w, CookieName, raw, sessionTTL)
	return nil
}

// HandleLogin redirects the browser to the OAuth provider, with a fresh
// state value bound to this browser by a short-lived cookie, so a callback
// can't be forged or replayed into someone else's browser (login CSRF).
func HandleLogin(e *env.Env, oauthCfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := secret.New()
		addCookie(w, stateCookieName, state, stateTTL)
		http.Redirect(w, r, oauthCfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
	}
}

// HandleLogout deletes this browser's session server-side and clears its
// cookie. Other devices stay logged in.
func HandleLogout(e *env.Env, repo *repository.DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(CookieName); err == nil && cookie.Value != "" {
			if err := repo.DeleteSession(cookie.Value); err != nil {
				e.Log.Error("could not delete session", "error", err)
			}
		}
		addCookie(w, CookieName, "", -1*time.Hour)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// HandleCallback checks the OAuth state, exchanges the code, upserts the
// user, and starts a session.
func HandleCallback(e *env.Env, oauthCfg *oauth2.Config, repo *repository.DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie(stateCookieName)
		// The state is single use, whatever happens next.
		addCookie(w, stateCookieName, "", -1*time.Hour)
		if err != nil || stateCookie.Value == "" ||
			subtle.ConstantTimeCompare([]byte(r.FormValue("state")), []byte(stateCookie.Value)) != 1 {
			e.Log.Error("oauth state mismatch")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		token, err := oauthCfg.Exchange(context.Background(), r.FormValue("code"))
		if err != nil {
			e.Log.Error("oauth exchange failed", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		resp, err := http.Get(e.Cfg.Auth.AccessTokenURL + token.AccessToken)
		if err != nil {
			e.Log.Error("could not fetch userinfo", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		content, err := io.ReadAll(resp.Body)
		if err != nil {
			e.Log.Error("could not read userinfo response", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		userInfo := models.UserInfo{}
		if err := json.Unmarshal(content, &userInfo); err != nil {
			e.Log.Error("could not parse userinfo", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		newUser := models.NewUser(userInfo.Id, userInfo.Email, userInfo.Picture)
		if err := repo.UpsertUser(newUser); err != nil {
			e.Log.Error("could not upsert user", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		user, err := repo.GetUser(userInfo.Email)
		if err != nil {
			e.Log.Error("could not load user after upsert", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if err := startSession(w, repo, user); err != nil {
			e.Log.Error("could not create session", "error", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// HandleDevLogin logs the browser in as a fixed local user with no OAuth
// round-trip. Gated behind Cfg.Auth.DevLogin (RM_AUTH_DEV_LOGIN=true) so it
// can never fire in a deployed environment by accident.
func HandleDevLogin(e *env.Env, repo *repository.DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !e.Cfg.Auth.DevLogin {
			http.NotFound(w, r)
			return
		}

		devUser := models.NewUser("dev-local", "dev@localhost", "")
		if err := repo.UpsertUser(devUser); err != nil {
			e.Log.Error("dev login upsert failed", "error", err)
			http.Error(w, "dev login failed", http.StatusInternalServerError)
			return
		}

		user, err := repo.GetUser("dev@localhost")
		if err != nil {
			e.Log.Error("dev login lookup failed", "error", err)
			http.Error(w, "dev login failed", http.StatusInternalServerError)
			return
		}

		if err := startSession(w, repo, user); err != nil {
			e.Log.Error("dev login session failed", "error", err)
			http.Error(w, "dev login failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

// verifyCookie resolves the session cookie to its user. The lookup is by
// the hash of the cookie value, so there's no secret comparison to time.
func verifyCookie(r *http.Request, repo *repository.DataRepository) (*models.User, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return nil, fmt.Errorf("missing auth cookie")
	}
	return repo.GetUserBySession(cookie.Value)
}

// LoginVerify protects HTML/browser routes: cookie only, redirects to /login
// on failure.
func LoginVerify(e *env.Env, repo *repository.DataRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := verifyCookie(r, repo)
			if err != nil {
				e.Log.Error("login verify failed", "error", err)
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r.WithContext(env.WithUser(r.Context(), user)))
		})
	}
}

func apiUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}

// APILoginVerify protects the JSON API: accepts either a Bearer token (used
// by the browser extension) or the session cookie (used by the SPA).
// Returns 401 JSON on failure, never a redirect.
func APILoginVerify(e *env.Env, repo *repository.DataRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
				value := strings.TrimPrefix(authHeader, "Bearer ")
				user, err := repo.GetUserByToken(value)
				if err != nil {
					apiUnauthorized(w)
					return
				}
				next.ServeHTTP(w, r.WithContext(env.WithUser(r.Context(), user)))
				return
			}

			user, err := verifyCookie(r, repo)
			if err != nil {
				apiUnauthorized(w)
				return
			}
			next.ServeHTTP(w, r.WithContext(env.WithUser(r.Context(), user)))
		})
	}
}
