package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/arpachuilo/go-registerable"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"

	"go-recipes/models"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	Templates map[string]*template.Template
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{make(map[string]*template.Template)}
}

func (self *TemplateRenderer) Add(name string, tmpl *template.Template) {
	self.Templates[name] = tmpl
}

// Render renders a template document
func (self *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return self.Templates[name].Execute(w, data)
}

type App struct {
	*echo.Echo
	*sql.DB
	*Auth
	*Mailer
	*TemplateRenderer
	*Config

	ImageETags *ETags[int64]
}

type Method string

const (
	GET     Method = "GET"
	HEAD    Method = "HEAD"
	POST    Method = "POST"
	PUT     Method = "PUT"
	PATCH   Method = "PATCH"
	DELETE  Method = "DELETE"
	CONNECT Method = "CONNECT"
	OPTIONS Method = "OPTIONS"
	TRACE   Method = "TRACE"
)

type EchoHandlerRegistration struct {
	// Path the endpoint is registered at
	Path string

	// Methods allowed for route
	Methods []Method

	// Do we need auth?
	RequireAuth bool

	// Your http handler func
	HandlerFunc echo.HandlerFunc
}

func NewApp(conf *Config) *App {
	// setup echo
	e := echo.New()

	e.Server.ReadTimeout = conf.Server.ReadTimeout
	e.Server.WriteTimeout = conf.Server.WriteTimeout
	e.Server.IdleTimeout = conf.Server.IdleTimeout

	if conf.Server.HTTPS {
		e.Pre(middleware.HTTPSRedirect())
		e.AutoTLSManager.Cache = autocert.DirCache("secret-dir")
		e.AutoTLSManager.Prompt = autocert.AcceptTOS
		e.AutoTLSManager.Email = conf.Server.Autocert.Email
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(conf.Server.Autocert.Hosts...)
	}

	e.Use(middleware.Recover())
	if conf.Server.EnableLogging {
		e.Use(middleware.Logger())
	}

	if conf.Server.RateLimit != nil {
		e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(*conf.Server.RateLimit))))
	}

	tr := NewTemplateRenderer()
	e.Renderer = tr
	e.HTTPErrorHandler = New500Handle(tr).errorHandler

	// setup static
	e.Static("/", conf.Server.AssetsDir)

	// open db
	db, err := sql.Open("sqlite3", conf.Database.Path)
	if err != nil {
		panic(err)
	}

	// open auth
	auth := NewAuth(conf.Auth)

	// open mailer
	mailer := NewMailer(conf.Mailer, nil)

	// image etags
	imageEtags := NewETags[int64]()

	// new clustering
	rc := NewRecipeCluster(conf.Database.Path, conf.Server.AssetsDir+"/cluster.html", 5)

	// hook into etags invalidation
	hook := func(ctx context.Context, exec boil.ContextExecutor, r *models.Recipe) error {
		// invalidate tag
		imageEtags.InvalidateByID(r.ID.Int64)

		// cluster
		rc.Run()
		return nil
	}

	models.AddRecipeHook(boil.AfterInsertHook, hook)
	models.AddRecipeHook(boil.AfterUpdateHook, hook)
	models.AddRecipeHook(boil.AfterDeleteHook, hook)

	h := &App{
		Echo:             e,
		TemplateRenderer: tr,

		Config:     conf,
		DB:         db,
		Mailer:     mailer,
		Auth:       auth,
		ImageETags: imageEtags,
	}

	registerable.RegisterMethods[EchoHandlerRegistration](h)

	return h
}

func (self App) Start() {
	if self.Config.Server.HTTPS {
		if err := self.Echo.StartAutoTLS(self.Config.Server.Address); err != http.ErrServerClosed {
			panic(err)
		}
	} else {
		if err := self.Echo.Start(self.Config.Server.Address); err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func (self App) Register(r EchoHandlerRegistration) {
	for _, method := range r.Methods {
		h := r.HandlerFunc

		// wrap auth if needed
		if r.RequireAuth {
			h = self.Auth.Use(h)
		}

		switch method {
		case GET:
			self.GET(r.Path, h)
		case POST:
			self.POST(r.Path, h)
		case PUT:
			self.PUT(r.Path, h)
		case PATCH:
			self.PATCH(r.Path, h)
		case DELETE:
			self.DELETE(r.Path, h)
		case OPTIONS:
			self.OPTIONS(r.Path, h)
		case CONNECT:
			self.CONNECT(r.Path, h)
		case TRACE:
			self.TRACE(r.Path, h)
		default:
			panic(fmt.Errorf("invalid method: %v", method))
		}
	}
}
