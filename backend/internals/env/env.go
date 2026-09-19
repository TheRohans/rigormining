package env

import (
	"context"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"gitlab.com/robrohan/rigormining/internals/models"
	"gitlab.com/robrohan/rigormining/internals/repository"
)

// Env carries per-process shared state (db handle, logger, config) into
// handlers. It is created once and shared by every request goroutine, so it
// must never hold per-request data - see WithUser/UserFromContext below for
// the authenticated user, which is per-request.
type Env struct {
	Db        *sqlx.DB
	Log       *slog.Logger
	Cfg       *models.Config
	Router    *mux.Router
	RandState string
	Repo      *repository.DataRepository
}

type contextKey int

const userContextKey contextKey = iota

// WithUser returns a copy of ctx carrying the authenticated user for this
// request. Used by the auth middleware, once per request - never store the
// result anywhere longer-lived than the request itself.
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext returns the authenticated user the auth middleware put on
// ctx, or nil if none is present (e.g. an unauthenticated route).
func UserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(userContextKey).(*models.User)
	return user
}
