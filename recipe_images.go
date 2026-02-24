package main

import (
	"go-recipes/models"
	"net/http"

	"github.com/arpachuilo/go-registrable"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (self App) ServeRecipeImages() registrable.Registration {
	return EchoHandlerRegistration{
		Path:    "/images/recipe/:path",
		Methods: []Method{GET},

		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// check etag
			etag := c.Request().Header.Get("If-None-Match")
			if self.ImageETags.HasETag(etag) {
				return c.NoContent(http.StatusNotModified)
			}

			// get id
			path := c.Param("path")

			// prepare db
			ctx := c.Request().Context()
			tx, err := self.DB.BeginTx(ctx, nil)
			defer tx.Commit()
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}

			// read recipe
			recipeQuery := models.Recipes(
				Select("id, image"),
				models.RecipeWhere.Path.EQ(null.StringFrom(path)),
			)

			recipe, err := recipeQuery.One(ctx, tx)
			if err != nil {
				return c.NoContent(http.StatusNotFound)
			}

			c.Response().Header().Set("ETag", self.ImageETags.Add(recipe.ID.Int64, recipe.Image.Bytes, false))
			contentType := http.DetectContentType(recipe.Image.Bytes)
			return c.Blob(http.StatusOK, contentType, recipe.Image.Bytes)
		},
	}
}
