package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
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

func NewRouter(conf *Config) *Router {
	// open db
	db, err := sql.Open("sqlite3", conf.Database.Path)
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

type ConfigRateLimiter struct {
	Limit   int           `mapstructure:"limit"`
	Burst   int           `mapstructure:"burst"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type ConfigServer struct {
	Address      string             `mapstructure:"address"`
	ReadTimeout  time.Duration      `mapstructure:"read_timeout"`
	WriteTimeout time.Duration      `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration      `mapstructure:"idle_timeout"`
	RateLimit    *ConfigRateLimiter `mapstructure:"rate_limiter"`
}

type ConfigDatabase struct {
	Path string `mapstructure:"path"`
}

type Config struct {
	Server   ConfigServer   `mapstructure:"server"`
	Database ConfigDatabase `mapstructure:"database"`
}

// TODO: Improve static asset cache w/ etags
// TODO: Support for list view along side the grid view for search results
// TODO: Look into impromvements to prevent multiple db reads on image serving
// TODO: Look into using sass
// TODO: Add simple auth mechanism (API keys)
// TODO: Add https setting

func main() {
	// load config
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("viper.ReadInConfig: %w", err))
	}

	conf := &Config{}
	err = viper.Unmarshal(conf)

	// setup router
	h := http.Handler(NewRouter(conf))

	// setup server
	if conf.Server.RateLimit != nil {
		conf := conf.Server.RateLimit
		h = NewLimiter(conf.Limit, conf.Burst, conf.Timeout).Use(h)
	}

	h = handlers.RecoveryHandler()(h)
	s := &http.Server{
		Addr:         conf.Server.Address,
		ReadTimeout:  conf.Server.ReadTimeout,
		WriteTimeout: conf.Server.WriteTimeout,
		IdleTimeout:  conf.Server.IdleTimeout,
		Handler:      h,
	}

	// serve
	log.Fatal(s.ListenAndServe())
}
