package models

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/jwx/jwk"
)

// Env context for db, logger, etc
type Env struct {
	Db     *sqlx.DB
	Log    *log.Logger
	Cfg    *Config
	Router *mux.Router
	KeySet *jwk.Set
}
