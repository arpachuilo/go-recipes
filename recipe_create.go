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
	"github.com/nfnt/resize"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type CreateTemplate struct {
	Title string
	Error string
}

func (self Router) ServeCreateRecipe() registerable.Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_create.html",
	))

	return HandlerRegistration{
		Name:        "create",
		Path:        "/create",
		Methods:     []string{"GET"},
		RequireAuth: self.Auth.enabled,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			// get error if any
			rq := r.URL.Query()
			error := ""
			keys, ok := rq["error"]
			if ok && len(keys) > 0 {
				error = keys[0]
			}

			data := CreateTemplate{
				Title: "Create Recipe",
				Error: error,
			}

			return tmpl.Execute(w, data)
		},
	}
}

func createRecipe(db *sql.DB, r *http.Request) (id int64, err error) {
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
	f, _, _ := r.FormFile("Image")
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

	// read total time
	var totalTime int64
	parsedTotalTime, err := strconv.ParseInt(r.Form["TotalTime"][0], 10, 64)
	if err == nil {
		totalTime = parsedTotalTime
	}

	// update recipe
	recipe := models.Recipe{
		Title:        null.StringFrom(r.Form["Title"][0]),
		Image:        null.BytesFrom(imgB),
		Author:       null.StringFrom(r.Form["Author"][0]),
		Calories:     null.StringFrom(r.Form["Calories"][0]),
		ServingSize:  null.StringFrom(r.Form["ServingSize"][0]),
		Yields:       null.StringFrom(r.Form["Yields"][0]),
		TotalTime:    null.Int64From(totalTime),
		Instructions: null.StringFrom(r.Form["Instructions"][0]),
	}

	whitelist := []string{"title", "author", "calories", "serving_size", "yields", "total_time", "instructions"}
	if len(imgB) > 0 {
		whitelist = append(whitelist, "image")
	}

	if err = recipe.Insert(context.Background(), tx, boil.Whitelist(whitelist...)); err != nil {
		return
	}

	// insert new ingredients
	ingredientText := r.Form["Ingredients"][0]
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
	tagsText := r.Form["Tags"][0]
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

	id = recipe.ID.Int64
	return
}

func (self Router) CreateRecipe() registerable.Registration {
	return HandlerRegistration{
		Name:        "create recipe",
		Path:        "/create",
		Methods:     []string{"POST"},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			id, err := func() (int64, error) {
				// parse form
				if err := r.ParseMultipartForm(32 << 20); err != nil {
					return 0, err
				}

				return createRecipe(self.DB, r)
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
