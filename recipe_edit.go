package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-recipes/models"
	"html/template"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/arpachuilo/go-registerable"
	"github.com/chai2010/webp"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type EditTemplate struct {
	Title       string
	Error       string
	Recipe      *models.Recipe
	Ingredients models.IngredientSlice
}

func (self Router) ServeEditRecipe() registerable.Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_edit.html",
	))

	return HandlerRegistration{
		Name:        "edit",
		Path:        "/edit/{id:[0-9]+}",
		Methods:     []string{"GET"},
		RequireAuth: self.Auth.enabled,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			// get id
			vars := mux.Vars(r)
			sid := vars["id"]
			id, err := strconv.ParseInt(sid, 10, 64)
			if err != nil {
				return err
			}

			// get error if any
			rq := r.URL.Query()
			error := ""
			keys, ok := rq["error"]
			if ok && len(keys) > 0 {
				error = keys[0]
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

			data := EditTemplate{
				Title:       fmt.Sprintf("Edit %v", recipe.Title.String),
				Error:       error,
				Recipe:      recipe,
				Ingredients: ingredients,
			}

			return tmpl.Execute(w, data)
		},
	}
}

func editRecipe(db *sql.DB, id int64, r *http.Request) (err error) {
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
		ID:           null.Int64From(id),
		Image:        null.BytesFrom(imgB),
		Title:        null.StringFrom(r.Form["Title"][0]),
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

	if _, err = recipe.Update(context.Background(), tx, boil.Whitelist(whitelist...)); err != nil {
		return
	}

	// delete previous ingredients
	ingredientsDeleteQuery := models.Ingredients(
		models.IngredientWhere.Recipeid.EQ(null.Int64From(id)),
	)

	if _, err = ingredientsDeleteQuery.DeleteAll(context.Background(), tx); err != nil {
		return
	}

	// insert new ingredients
	ingredientText := r.Form["Ingredients"][0]
	for _, i := range strings.Split(ingredientText, "\n") {
		if strings.TrimSpace(i) == "" {
			continue
		}

		ingredient := models.Ingredient{
			Recipeid:   null.Int64From(id),
			Ingredient: null.StringFrom(i),
		}

		err = ingredient.Insert(context.Background(), tx, boil.Infer())
		if err != nil {
			return
		}
	}

	return nil
}

func (self Router) EditRecipe() registerable.Registration {
	return HandlerRegistration{
		Name:        "edit recipe",
		Path:        "/edit/{id:[0-9]+}",
		Methods:     []string{"POST"},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			// get id
			vars := mux.Vars(r)
			sid := vars["id"]
			id, err := strconv.ParseInt(sid, 10, 64)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", r.Header.Get("Referer"), err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			err = func() error {
				// parse form
				if err := r.ParseMultipartForm(32 << 20); err != nil {
					return err
				}

				// update recipe
				return editRecipe(self.DB, id, r)
			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", r.Header.Get("Referer"), err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			redirectURL := fmt.Sprintf("/recipe/%v", id)
			http.Redirect(w, r, redirectURL, http.StatusFound)
		},
	}
}
