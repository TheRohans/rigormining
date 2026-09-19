package auth

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/models"
	"gitlab.com/robrohan/rigormining/internals/repository"
)

const CookieName = "RM_AT"

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

// HandleLogin redirects the browser to the OAuth provider.
func HandleLogin(e *env.Env, oauthCfg *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := oauthCfg.AuthCodeURL(e.RandState)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// HandleLogout clears the session cookie.
func HandleLogout(e *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addCookie(w, CookieName, "", -1*time.Hour)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

// HandleCallback exchanges the OAuth code, upserts the user, and sets the
// session cookie.
func HandleCallback(e *env.Env, oauthCfg *oauth2.Config, repo *repository.DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("state") != e.RandState {
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

		salt := fmt.Sprintf("%x", rand.Int())
		newUser := models.NewUser(userInfo.Id, userInfo.Email, userInfo.Picture)
		if err := repo.UpsertUser(newUser, salt); err != nil {
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

		hash := md5.Sum([]byte(user.Email + user.AuthId + salt))
		addCookie(w, CookieName, fmt.Sprintf("%s:%x", user.UUID, hash), 30*24*time.Hour)

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

		salt := fmt.Sprintf("%x", rand.Int())
		devUser := models.NewUser("dev-local", "dev@localhost", "")
		if err := repo.UpsertUser(devUser, salt); err != nil {
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

		hash := md5.Sum([]byte(user.Email + user.AuthId + salt))
		addCookie(w, CookieName, fmt.Sprintf("%s:%x", user.UUID, hash), 30*24*time.Hour)

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func verifyCookie(r *http.Request, repo *repository.DataRepository) (*models.User, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return nil, fmt.Errorf("missing auth cookie")
	}

	parts := strings.SplitN(cookie.Value, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed auth cookie")
	}

	user, err := repo.GetUserById(parts[0])
	if err != nil {
		return nil, err
	}

	salt := ""
	if user.Salt != nil {
		salt = *user.Salt
	}
	hash := md5.Sum([]byte(user.Email + user.AuthId + salt))
	if fmt.Sprintf("%x", hash) != parts[1] {
		return nil, fmt.Errorf("auth cookie hash mismatch")
	}

	return user, nil
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
