package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/volatiletech/null/v8"
)

type RecipeTemplate struct {
	Title       string
	Recipe      *models.Recipe
	Ingredients models.IngredientSlice
}

func (self Router) ServeRecipeDisplay() Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/recipe_display.html",
	))

	return HandlerRegistration{
		Name:    "recipe",
		Path:    "/recipe/{id:[0-9]+}",
		Methods: []string{"GET"},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			// get id
			vars := mux.Vars(r)
			sid := vars["id"]
			id, err := strconv.ParseInt(sid, 10, 64)
			if err != nil {
				return err
			}

			// prepare db
			ctx := context.Background()
			tx, err := self.DB.BeginTx(ctx, nil)
			defer tx.Commit()
			if err != nil {
				return err
			}

			// read recipe
			recipeQuery := models.Recipes(
				models.RecipeWhere.ID.EQ(null.Int64From(id)),
			)

			recipe, err := recipeQuery.One(ctx, tx)
			if err != nil {
				return err
			}

			// read ingredients
			igredientsQuery := models.Ingredients(
				models.IngredientWhere.Recipeid.EQ(null.Int64From(id)),
			)

			ingredients, err := igredientsQuery.All(ctx, tx)
			if err != nil {
				return err
			}
			// run template
			data := &RecipeTemplate{
				Title:       fmt.Sprintf("%v", recipe.Title.String),
				Recipe:      recipe,
				Ingredients: ingredients,
			}

			return tmpl.Execute(w, data)
		},
	}
}
