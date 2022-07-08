package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"os/exec"

	"github.com/arpachuilo/go-registerable"
	"github.com/volatiletech/null/v8"
)

type ImportTemplate struct {
	Title string
	Error string
}

func (self Router) ServeRecipeImport() registerable.Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_import.html",
	))

	return HandlerRegistration{
		Name:        "import",
		Path:        "/import",
		Methods:     []string{"GET"},
		RequireAuth: self.Auth.enabled,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			rq := r.URL.Query()
			error := ""
			keys, ok := rq["error"]
			if ok && len(keys) > 0 {
				error = keys[0]
			}

			data := ImportTemplate{
				Title: "Import Recipe from URL",
				Error: error,
			}

			return tmpl.Execute(w, data)
		},
	}
}

func scrapeRecipe(db, url string) error {
	// scrape URL
	out, err := exec.Command("python3", "cmds/scrape/scrape.py", db, url).CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return err
	}

	return nil
}

func (self Router) ImportRecipe() registerable.Registration {
	return HandlerRegistration{
		Name:        "scrape",
		Path:        "/import",
		Methods:     []string{"POST"},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			path, err := func() (string, error) {
				// parse form
				if err := r.ParseForm(); err != nil {
					return "", err
				}

				// get url to import
				importURL, ok := r.Form["url"]
				if !ok || len(importURL) == 0 {
					return "", fmt.Errorf("%v", "not url provided")
				}

				// scrape url
				if err := scrapeRecipe("recipes.db", importURL[0]); err != nil {
					return "", err
				}

				// read recipe
				ctx := context.Background()
				tx, err := self.DB.BeginTx(ctx, nil)
				defer tx.Commit()
				if err != nil {
					return "", err
				}

				query := models.Recipes(
					models.RecipeWhere.URL.EQ(null.StringFrom(importURL[0])),
				)

				recipe, err := query.One(ctx, tx)
				if err != nil {
					return "", err
				}

				return recipe.Path.String, nil

			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", r.Header.Get("Referer"), err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			redirectURL := fmt.Sprintf("/edit/%v", path)
			http.Redirect(w, r, redirectURL, http.StatusFound)
		},
	}
}
