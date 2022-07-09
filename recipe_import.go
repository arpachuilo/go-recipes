package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"os/exec"

	"github.com/arpachuilo/go-registerable"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
)

type ImportTemplate struct {
	Title string
	Error string
}

func (self App) ServeRecipeImport() registerable.Registration {
	// read templates dynamically for debug
	tmplName := "recipe_import"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_import.html",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)

	return EchoHandlerRegistration{
		Path:        "/import",
		Methods:     []Method{GET},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			error := c.QueryParam("error")

			data := ImportTemplate{
				Title: "Import Recipe from URL",
				Error: error,
			}

			return c.Render(http.StatusOK, tmplName, data)
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

func (self App) ImportRecipe() registerable.Registration {
	return EchoHandlerRegistration{
		Path:        "/import",
		Methods:     []Method{POST},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			path, err := func() (string, error) {
				// get url to import
				importURL := c.FormValue("url")
				if importURL == "" {
					return "", fmt.Errorf("%v", "not url provided")
				}

				// scrape url
				if err := scrapeRecipe("recipes.db", importURL); err != nil {
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
					models.RecipeWhere.URL.EQ(null.StringFrom(importURL)),
				)

				recipe, err := query.One(ctx, tx)
				if err != nil {
					return "", err
				}

				return recipe.Path.String, nil

			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := fmt.Sprintf("/edit/%v", path)
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}
