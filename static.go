package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"runtime/debug"

	"github.com/arpachuilo/go-registerable"
)

//go:embed assets/static
var staticFS embed.FS

type CachedFileServer struct {
	fs   http.Handler
	etag *string
}

func (self CachedFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Cache-Control", "public, max-age=7776000")
	// check etag
	etag := r.Header.Get("If-None-Match")
	if etag == *self.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// set etag
	if self.etag != nil {
		w.Header().Set("ETag", *self.etag)
	}

	self.fs.ServeHTTP(w, r)
}

func NewCachedFileServer(fsys fs.FS) CachedFileServer {
	httpFS := http.FileServer(http.FS(fsys))
	var etag *string

	info, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				v := fmt.Sprintf("1%v", s.Value)
				etag = &v
			}
		}
	}

	return CachedFileServer{httpFS, etag}
}

func (self Router) ServeStatic() registerable.Registration {
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
