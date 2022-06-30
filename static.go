package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/static
var staticFS embed.FS

type CachedFileServer struct {
	fs http.Handler
}

func (self CachedFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	self.fs.ServeHTTP(w, r)
}

func NewCachedFileServer(fsys fs.FS) CachedFileServer {
	httpFS := http.FileServer(http.FS(fsys))
	return CachedFileServer{httpFS}
}

func (self Router) ServeStatic() Registration {
	fsys, err := fs.Sub(staticFS, "assets/static")
	if err != nil {
		panic(err)
	}

	cachedFS := NewCachedFileServer(fsys)
	return HandlerRegistration{
		Name:    "fs",
		Path:    "/static/",
		Methods: []string{"GET"},

		Handler: http.StripPrefix("/static", cachedFS),
	}
}
