package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-recipes/models"
	"html/template"
	"image"
	"net/http"
	"strconv"
	"strings"

	"github.com/arpachuilo/go-registerable"
	"github.com/chai2010/webp"
	"github.com/labstack/echo/v4"
	"github.com/nfnt/resize"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type CreateTemplate struct {
	Title string
	Error string
}

func (self App) ServeCreateRecipe() registerable.Registration {
	tmplName := "recipe_create"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_create.html",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)

	return EchoHandlerRegistration{
		Path:        "/create",
		Methods:     []Method{GET},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// get error if any
			err := c.QueryParam("error")

			data := CreateTemplate{
				Title: "Create Recipe",
				Error: err,
			}

			return c.Render(http.StatusOK, tmplName, data)
		},
	}
}

func createRecipe(db *sql.DB, c echo.Context) (id int64, err error) {
	ctx := context.Background()
	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, nil)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err != nil {
		return
	}

	// read image (if any)
	var imgB []byte
	fh, err := c.FormFile("Image")
	if err == nil {
		f, _ := fh.Open()
		if f != nil {
			// decode img
			var img image.Image
			img, _, err = image.Decode(f)
			if err != nil {
				return
			}

			// resize img
			img = resize.Resize(360, 0, img, resize.Lanczos3)
			imgB, err = webp.EncodeRGB(img, webp.DefaulQuality)
			if err != nil {
				return
			}
		}
	}

	// read total time
	var totalTime int64
	parsedTotalTime, err := strconv.ParseInt(c.FormValue("TotalTime"), 10, 64)
	if err == nil {
		totalTime = parsedTotalTime
	}

	// custom path
	var new_path *string
	np := c.FormValue("Path")
	if np != "" {
		new_path = &np
	}

	// update recipe
	recipe := models.Recipe{
		Path:         null.StringFromPtr(new_path),
		Title:        null.StringFrom(c.FormValue("Title")),
		Image:        null.BytesFrom(imgB),
		Author:       null.StringFrom(c.FormValue("Author")),
		Calories:     null.StringFrom(c.FormValue("Calories")),
		ServingSize:  null.StringFrom(c.FormValue("ServingSize")),
		Yields:       null.StringFrom(c.FormValue("Yields")),
		TotalTime:    null.Int64From(totalTime),
		Instructions: null.StringFrom(c.FormValue("Instructions")),
	}

	whitelist := []string{"title", "author", "calories", "serving_size", "yields", "total_time", "instructions", "path"}
	if len(imgB) > 0 {
		whitelist = append(whitelist, "image")
	}

	if err = recipe.Insert(context.Background(), tx, boil.Whitelist(whitelist...)); err != nil {
		return
	}

	// insert new ingredients
	ingredientText := c.FormValue("Ingredients")
	for _, i := range strings.Split(ingredientText, "\n") {
		if strings.TrimSpace(i) == "" {
			continue
		}

		ingredient := models.Ingredient{
			Recipeid:   recipe.ID,
			Ingredient: null.StringFrom(i),
		}

		err = ingredient.Insert(context.Background(), tx, boil.Infer())
		if err != nil {
			return
		}
	}

	// insert new tags
	tagsText := c.FormValue("Tags")
	for _, t := range strings.Split(tagsText, ",") {
		tf := strings.ToLower(strings.TrimSpace(t))
		if tf == "" {
			continue
		}

		tag := models.Tag{
			Recipeid: recipe.ID,
			Tag:      null.StringFrom(tf),
		}

		err = tag.Insert(context.Background(), tx, boil.Infer())
		if err != nil {
			return
		}
	}

	// get new path
	id = recipe.ID.Int64
	return
}

func (self App) CreateRecipe() registerable.Registration {
	return EchoHandlerRegistration{
		Path:        "/create",
		Methods:     []Method{POST},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			id, err := createRecipe(self.DB, c)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := fmt.Sprintf("/edit/%v", id)
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}
