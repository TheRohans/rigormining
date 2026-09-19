package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ardanlabs/conf"
	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"

	"gitlab.com/robrohan/rigormining/internals/auth"
	"gitlab.com/robrohan/rigormining/internals/env"
	"gitlab.com/robrohan/rigormining/internals/handlers"
	"gitlab.com/robrohan/rigormining/internals/models"
	"gitlab.com/robrohan/rigormining/internals/repository"
)

// will be replaced with git hash at build time
var build = "develop"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	// =========================================================================
	// Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// =========================================================================
	// Configuration
	cfg := models.Config{}
	if err := conf.Parse(os.Args[1:], "RM", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("RM", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	oauthCfg := auth.NewOAuthConfig(&cfg)

	expvar.NewString("build").Set(build)
	logger.Info("started : application initializing", "version", build)
	defer logger.Info("completed")

	// =========================================================================
	// Database
	logger.Info("initializing database support")
	db, err := repository.OpenDatabase(cfg.DB.Driver, cfg.DB.Connection, cfg.Base.Root)
	if err != nil {
		return errors.Wrap(err, "opening database")
	}
	defer db.Close()

	for _, stmt := range strings.Split(cfg.DB.Post, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			logger.Warn("post-connect pragma failed", "stmt", stmt, "error", err)
		}
	}

	if err := os.MkdirAll(cfg.Library.Dir, 0o755); err != nil {
		return errors.Wrap(err, "creating library directory")
	}

	repo := repository.Attach(cfg.Base.Root, db, cfg.DB.Driver)

	// =========================================================================
	// Debug service (pprof/expvar)
	go func() {
		logger.Info("debug listening", "address", cfg.Web.DebugHost)
		logger.Info("debug listener closed", "err", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// =========================================================================
	// Routes
	router := mux.NewRouter()

	randState := fmt.Sprintf("%x", rand.Int())
	e := &env.Env{
		Db:        db,
		Log:       logger,
		Cfg:       &cfg,
		Router:    router,
		RandState: randState,
		Repo:      repo,
	}

	router.HandleFunc("/login", auth.HandleLogin(e, oauthCfg)).Methods("GET")
	router.HandleFunc("/callback", auth.HandleCallback(e, oauthCfg, repo)).Methods("GET")
	router.HandleFunc("/dev-login", auth.HandleDevLogin(e, repo)).Methods("GET")

	secure := router.PathPrefix("/-/").Subrouter()
	secure.Use(auth.LoginVerify(e, repo))
	secure.HandleFunc("/logout", auth.HandleLogout(e)).Methods("GET")

	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(auth.APILoginVerify(e, repo))

	api.HandleFunc("/whoami", handlers.APIWhoAmI(e)).Methods("GET")

	api.HandleFunc("/items", handlers.APIGetItems(e)).Methods("GET")
	api.HandleFunc("/items", handlers.APICreateItem(e)).Methods("POST")
	api.HandleFunc("/capture", handlers.APICapture(e)).Methods("POST")
	api.HandleFunc("/items/{id}", handlers.APIGetItem(e)).Methods("GET")
	api.HandleFunc("/items/{id}", handlers.APIUpdateItem(e)).Methods("PATCH")
	api.HandleFunc("/items/{id}", handlers.APIDeleteItem(e)).Methods("DELETE")
	api.HandleFunc("/items/{id}/file", handlers.APIGetItemFile(e)).Methods("GET")
	api.HandleFunc("/items/{id}/file", handlers.APIUploadItemFile(e)).Methods("POST")
	api.HandleFunc("/items/{id}/export.md", handlers.APIExportMarkdown(e)).Methods("GET")
	api.HandleFunc("/items/{id}/export.bib", handlers.APIExportBibtex(e)).Methods("GET")
	api.HandleFunc("/items/{id}/tags", handlers.APIAddItemTag(e)).Methods("POST")
	api.HandleFunc("/items/{id}/tags/{tagId}", handlers.APIRemoveItemTag(e)).Methods("DELETE")
	api.HandleFunc("/items/{id}/delivered", handlers.APIMarkDelivered(e)).Methods("POST")
	api.HandleFunc("/items/{id}/delivered", handlers.APIClearDelivered(e)).Methods("DELETE")

	api.HandleFunc("/tokens", handlers.APIListTokens(e)).Methods("GET")
	api.HandleFunc("/tokens", handlers.APICreateToken(e)).Methods("POST")
	api.HandleFunc("/tokens/{id}", handlers.APIDeleteToken(e)).Methods("DELETE")

	api.HandleFunc("/extension/{browser}", handlers.APIDownloadExtension(e)).Methods("GET")

	// The built React app is served last, as a catch-all, so it never
	// shadows the routes above.
	router.PathPrefix("/").Handler(handlers.SPAFileServer(cfg.Web.StaticDir))

	corsOrigins := ghandlers.AllowedOrigins([]string{cfg.Web.AllowedOrigin})
	corsHeaders := ghandlers.AllowedHeaders([]string{"Accept", "Content-Type", "Authorization"})
	corsMethods := ghandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PATCH", "DELETE", "OPTIONS"})
	corsCreds := ghandlers.AllowCredentials()

	srv := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      ghandlers.CORS(corsOrigins, corsHeaders, corsMethods, corsCreds)(router),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// =========================================================================
	// Graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.Info("shutting down", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
