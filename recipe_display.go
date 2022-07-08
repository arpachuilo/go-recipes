package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"

	"github.com/arpachuilo/go-registerable"
	"github.com/gorilla/mux"
	"github.com/volatiletech/null/v8"
)

type RecipeTemplate struct {
	Title       string
	Recipe      *models.Recipe
	Ingredients models.IngredientSlice
	Tags        models.TagSlice
}

func (self Router) ServeRecipeDisplay() registerable.Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_display.html",
	))

	return HandlerRegistration{
		Name:        "recipe",
		Path:        "/recipe/{path}",
		Methods:     []string{"GET"},
		RequireAuth: self.Auth.enabled,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			// get path/id
			vars := mux.Vars(r)
			path := vars["path"]

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

			return tmpl.Execute(w, data)
		},
	}
}
