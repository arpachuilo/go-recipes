package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/volatiletech/sqlboiler/v4/boil"

	_ "github.com/mattn/go-sqlite3"

	"go-recipes/models"
)

func redirect(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.RequestURI
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

type Router struct {
	*sql.DB
	*mux.Router
	*Auth

	ImageETags *ETags[int64]
}

type HandlerRegistration struct {
	// 404 handler
	Is404 bool

	// Auth required
	RequireAuth bool

	// Name of the methdod, should be unique
	Name string

	// Path the endpoint is registered at
	Path string

	// Methods allowed for using this service
	Methods []string

	// Your http handler
	Handler http.Handler

	// Your http handler func
	HandlerFunc http.HandlerFunc

	// Handler func that handles errors
	ErrorHandlerFunc ErrorHandlerFunc
}

func NewRouter() *Router {
	// open db
	db, err := sql.Open("sqlite3", "./recipes.db")
	if err != nil {
		panic(err)
	}

	// open auth
	auth := NewAuth("supersecret")

	// image etags
	imageEtags := NewETags[int64]()

	// hook into etags invalidation
	invalidate := func(ctx context.Context, exec boil.ContextExecutor, r *models.Recipe) error {
		imageEtags.InvalidateByID(r.ID.Int64)
		return nil
	}

	models.AddRecipeHook(boil.AfterInsertHook, invalidate)
	models.AddRecipeHook(boil.AfterUpdateHook, invalidate)
	models.AddRecipeHook(boil.AfterDeleteHook, invalidate)

	r := &Router{
		DB:         db,
		Router:     mux.NewRouter(),
		Auth:       auth,
		ImageETags: imageEtags,
	}

	RegisterMethods[HandlerRegistration](r)

	return r
}

func (self Router) Register(r HandlerRegistration) {
	// check we only have one of
	count := 0

	if r.Handler != nil {
		count++
	}
	if r.HandlerFunc != nil {
		count++
	}
	if r.ErrorHandlerFunc != nil {
		count++
	}

	if count > 1 {
		panic("more than handler type set")
	}

	if r.Is404 && r.Handler != nil {
		self.Router.NotFoundHandler = r.Handler
	} else if r.Handler != nil {
		h := r.Handler
		if r.RequireAuth {
			h = self.Auth.Use(h)
		}

		self.Router.
			PathPrefix(r.Path).
			Handler(h).
			Name(r.Name).
			Methods(r.Methods...)
	}

	if r.ErrorHandlerFunc != nil {
		h := Handle500Middleware(r.ErrorHandlerFunc)
		if r.RequireAuth {
			h = self.Auth.UseFunc(h)
		}

		self.Router.
			HandleFunc(r.Path, h).
			Name(r.Name).
			Methods(r.Methods...)
	}

	if r.HandlerFunc != nil {
		h := r.HandlerFunc
		if r.RequireAuth {
			h = self.Auth.UseFunc(h)
		}

		self.Router.
			HandleFunc(r.Path, r.HandlerFunc).
			Name(r.Name).
			Methods(r.Methods...)
	}
}

// TODO: Improve static asset cache w/ etags
// TODO: Look into impromvements to prevent multiple db reads on image serving
// TODO: Look into using sass

// TODO: Add simple auth mechanism (API keys)
// TODO: Use viper config
// - rate limiter
// - timeouts
// - address
// - db file path
// - https
// - auth keys
func main() {
	mux := NewRouter()

	h := NewLimiter(30, 50, time.Minute*3).Use(mux)
	h = handlers.RecoveryHandler()(h)

	s := &http.Server{
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 120,
		Handler:      h,
	}

	log.Fatal(s.ListenAndServe())
}
