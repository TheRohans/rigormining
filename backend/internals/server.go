package server

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ardanlabs/conf"
	ghandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"gitlab.com/robrohan/knotset/internals/auth"
	"gitlab.com/robrohan/knotset/internals/handlers"
	"gitlab.com/robrohan/knotset/internals/models"
	"gitlab.com/robrohan/knotset/internals/repository"

	// for the port 4000 stuff
	_ "net/http/pprof"
)

// will be replaced with git hash
var build = "develop"

func Run() error {
	// =========================================================================
	// Logging
	log := log.New(os.Stdout, "KS : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration
	cfg := models.Config{}

	if err := conf.Parse(os.Args[1:], "KS", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("KS", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// App Starting
	expvar.NewString("build").Set(build)
	log.Printf("Started : Application initializing : version %q", build)
	defer log.Println("Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("Config :\n%v\n", out)

	// =========================================================================
	// Start Database
	log.Println("Initializing database support")

	db, err := repository.OpenDatabase(
		cfg.DB.Driver, cfg.DB.Connection, cfg.Base.Root)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		log.Printf("Database Stopping: %s", cfg.DB.Connection)
		db.Close()
	}()

	///////////////////////////////////////
	// After database is created, do any post settings. Mostly used for
	// sqlite tuning
	posts := strings.Split(cfg.DB.Post, ";")
	for _, ex := range posts {
		cmd := strings.ReplaceAll(ex, "{schema}", cfg.Base.Root)
		_, err = db.Exec(cmd)
		if err != nil {
			log.Printf("%v\n", err.Error())
		}
	}
	///////////////////////////////////////

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.
	log.Println("Initializing debugging support")
	go func() {
		log.Printf("Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("Debug Listener closed: %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// Put the API on top of the connection
	repo := repository.Attach(cfg.Base.Root, db, cfg.DB.Driver)

	// =========================================================================
	// Setup template handling
	templates := handlers.TemplateInit()

	// =========================================================================
	// Start API Service
	log.Println("Initializing API support")

	router := mux.NewRouter() // .StrictSlash(true)

	env := &models.Env{
		Db:  db,
		Log: log,
		// App:    &app,
		Router: router,
	}

	// Routes
	{
		router.PathPrefix("/public").Handler(
			http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))

		secure := router.PathPrefix("/-/").Subrouter()
		secure.Use(auth.JwtVerify(env))

		router.HandleFunc("/", handlers.ServePage(env, templates))
		// router.HandleFunc("/about", handlers.ServePage(env, templates))
		router.HandleFunc("/login", handlers.ServePage(env, templates)).Methods("GET")
		router.HandleFunc("/login", auth.Login(env, repo)).Methods("POST")
	}

	headersOk := ghandlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin", "X-Requested-With"})
	originsOk := ghandlers.AllowedOrigins([]string{"*"})
	methodsOk := ghandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      ghandlers.CORS(originsOk, headersOk, methodsOk)(router),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	log.Printf("API listening on %s", api.Addr)
	api.ListenAndServe()

	return nil
}
