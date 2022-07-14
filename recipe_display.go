package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"

	"github.com/arpachuilo/go-registerable"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
)

type RecipeTemplate struct {
	Title       string
	Recipe      *models.Recipe
	Ingredients models.IngredientSlice
	Tags        models.TagSlice
}

func (self App) ServeRecipeDisplay() registerable.Registration {
	// read templates dynamically for debug
	tmplName := "recipe_display"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.gohtml",
		"templates/nav.gohtml",
		"templates/recipe_display.gohtml",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)

	return EchoHandlerRegistration{
		Path:        "/recipe/:path",
		Methods:     []Method{GET},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// get path/id
			path := c.Param("path")

			// prepare db
			ctx := context.Background()
			tx, err := self.DB.BeginTx(ctx, nil)
			defer tx.Commit()
			if err != nil {
				return err
			}

			// read recipe
			recipeQuery := models.Recipes(
				models.RecipeWhere.Path.EQ(null.StringFrom(path)),
			)

			recipe, err := recipeQuery.One(ctx, tx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return echo.NewHTTPError(http.StatusNotFound, "Could not find requested recipe.")
				}

				return err
			}

			// read ingredients
			igredientsQuery := models.Ingredients(
				models.IngredientWhere.Recipeid.EQ(recipe.ID),
			)

			ingredients, err := igredientsQuery.All(ctx, tx)
			if err != nil {
				return err
			}

			// read tags
			tagsQuery := models.Tags(
				models.TagWhere.Recipeid.EQ(recipe.ID),
			)

			tags, err := tagsQuery.All(ctx, tx)
			if err != nil {
				return err
			}

			// run template
			data := &RecipeTemplate{
				Title:       fmt.Sprintf("%v", recipe.Title.String),
				Recipe:      recipe,
				Ingredients: ingredients,
				Tags:        tags,
			}

			return c.Render(http.StatusOK, tmplName, data)
		},
	}
}
