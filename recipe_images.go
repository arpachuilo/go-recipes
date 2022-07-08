package main

import (
	"database/sql"
	"go-recipes/models"
	"net/http"
	"sync"

	"github.com/arpachuilo/go-registerable"
	"github.com/gorilla/mux"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type RecipeImageFileServer struct {
	root  string
	db    *sql.DB
	etags *ETags[int64]

	images map[int64][]byte
	mut    *sync.RWMutex

	h404 http.Handler
}

func NewRecipeImageFileServer(root string, db *sql.DB, etags *ETags[int64]) RecipeImageFileServer {
	return RecipeImageFileServer{
		root:  root,
		db:    db,
		etags: etags,

		images: make(map[int64][]byte, 0),
		mut:    &sync.RWMutex{},

		h404: http.NotFoundHandler(),
	}
}

// TODO cache image hits to prevent extra db reads
func (self RecipeImageFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check etag
	etag := r.Header.Get("If-None-Match")
	if self.etags.HasETag(etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// get id
	vars := mux.Vars(r)
	path := vars["path"]

	// prepare db
	ctx := r.Context()
	tx, err := self.db.BeginTx(ctx, nil)
	defer tx.Commit()
	if err != nil {
		self.h404.ServeHTTP(w, r)
		return
	}

	// read recipe
	recipeQuery := models.Recipes(
		Select("id, image"),
		models.RecipeWhere.Path.EQ(null.StringFrom(path)),
	)

	recipe, err := recipeQuery.One(ctx, tx)
	if err != nil {
		self.h404.ServeHTTP(w, r)
		return
	}

	w.Header().Set("ETag", self.etags.Add(recipe.ID.Int64, recipe.Image.Bytes, false))
	w.Write(recipe.Image.Bytes)
}

func (self Router) ServeRecipeImages() registerable.Registration {
	imageFS := NewRecipeImageFileServer("images", self.DB, self.ImageETags)
	return HandlerRegistration{
		Name:    "images",
		Path:    "/images/recipe/{path}",
		Methods: []string{"GET"},

		RequireAuth: self.Auth.enabled,
		Handler:     imageFS,
	}
}
