package models

import "time"

// Config is the config object for the application
type Config struct {
	Base struct {
		Root string `conf:"default:rigormining"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
		// StaticDir is where the built frontend (frontend/dist) is served from.
		StaticDir string `conf:"default:./static"`
		// AllowedOrigin is the frontend origin allowed to make credentialed
		// cross-origin requests during local dev (e.g. webpack-dev-server).
		AllowedOrigin string `conf:"default:http://localhost:8080"`
	}
	// Auth holds the OAuth2 provider settings (defaults to Google).
	Auth struct {
		RedirectURL    string   `conf:"default:http://localhost:3000/callback"`
		ClientID       string   `conf:"default:changeme"`
		ClientSecret   string   `conf:"default:changeme"`
		Scopes         []string `conf:"default:email openid https://www.googleapis.com/auth/userinfo.email"`
		AuthURL        string   `conf:"default:https://accounts.google.com/o/oauth2/auth"`
		TokenURL       string   `conf:"default:https://oauth2.googleapis.com/token"`
		AuthStyle      int      `conf:"default:1"`
		AccessTokenURL string   `conf:"default:https://www.googleapis.com/oauth2/v2/userinfo?access_token="`
		// DevLogin enables GET /dev-login, which logs the browser in as a
		// fixed local user with no OAuth round-trip. Only ever set this true
		// for local development - never in a deployed environment.
		DevLogin bool `conf:"default:false"`
	}
	DB struct {
		Driver     string `conf:"default:sqlite3"`
		Connection string `conf:"default:./datastore/rigormining.db"`
		// journal_mode is DELETE, not WAL - WAL needs mmap'd shared memory
		// and byte-range locking that Cloud Storage FUSE doesn't support,
		// so it fails against the mounted datastore/ volume in production
		// (see the comment in backend/.env.template).
		Post string `conf:"default:PRAGMA synchronous = normal;PRAGMA journal_mode = DELETE;PRAGMA temp_store = memory"`
	}
	// LibraryDir is where uploaded PDFs/EPUBs are stored, one subfolder per
	// user. On GCP this lives on the same Cloud Run GCS-volume mount as the
	// SQLite file.
	Library struct {
		Dir string `conf:"default:./datastore/library"`
	}
	// Extension is where the built browser-extension zips live (see
	// extension/Makefile) - served to logged-in users from the "Get
	// Extension" page while it's still pre-store/beta distribution.
	Extension struct {
		Dir string `conf:"default:./static/extension"`
	}
	// Skills is where the built downloadable agent-skill zips live (see
	// the root Makefile's skills target) - served from the same "Get
	// Extension" page as a zip someone can drop into ~/.agents/skills.
	Skills struct {
		Dir string `conf:"default:./static/skills"`
	}
}
