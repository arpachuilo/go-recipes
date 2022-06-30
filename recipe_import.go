package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"os/exec"
	"syscall"

	"github.com/volatiletech/null/v8"
)

type ImportTemplate struct {
	Title string
	Error string
}

func (self Router) ServeRecipeImport() Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/recipe_import.html",
	))

	return HandlerRegistration{
		Name:    "import",
		Path:    "/import",
		Methods: []string{"GET"},
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
	cmd := exec.Command("python3", "cmds/scrape/scrape.py", db, url)
	if err := cmd.Start(); err != nil {
		return err
	}

	// wait for scraper exit
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("exiterr.Sys: %d", status)
			}
		} else {
			return err
		}

		return err
	}

	return nil
}

func (self Router) ImportRecipe() Registration {
	return HandlerRegistration{
		Name:    "scrape",
		Path:    "/import",
		Methods: []string{"POST"},
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			id, err := func() (int64, error) {
				// parse form
				if err := r.ParseForm(); err != nil {
					return -1, err
				}

				// get url to import
				importURL, ok := r.Form["url"]
				if !ok || len(importURL) == 0 {
					return -1, fmt.Errorf("%v", "not url provided")
				}

				// scrape url
				if err := scrapeRecipe("recipes.db", importURL[0]); err != nil {
					return -1, err
				}

				// read recipe
				ctx := context.Background()
				tx, err := self.DB.BeginTx(ctx, nil)
				defer tx.Commit()
				if err != nil {
					return -1, err
				}

				query := models.Recipes(
					models.RecipeWhere.URL.EQ(null.StringFrom(importURL[0])),
				)

				recipe, err := query.One(ctx, tx)
				if err != nil {
					return -1, err
				}

				return recipe.ID.Int64, nil

			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", r.Header.Get("Referer"), err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			redirectURL := fmt.Sprintf("/edit/%v", id)
			http.Redirect(w, r, redirectURL, http.StatusFound)
		},
	}
}
