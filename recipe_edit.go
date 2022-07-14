package main

import (
	"context"
	"database/sql"
	"errors"
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
	"github.com/labstack/echo/v4"
	"github.com/nfnt/resize"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type EditTemplate struct {
	Title       string
	Error       string
	Recipe      *models.Recipe
	Ingredients models.IngredientSlice
	Tags        models.TagSlice
}

func (self App) ServeEditRecipe() registerable.Registration {
	// read templates dynamically for debug
	tmplName := "recipe_edit"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.gohtml",
		"templates/nav.gohtml",
		"templates/recipe_edit.gohtml",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)
	return EchoHandlerRegistration{
		Path:        "/edit/:id",
		Methods:     []Method{GET},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// get id
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				return err
			}

			// get error if any
			error := c.QueryParam("error")

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
				if errors.Is(err, sql.ErrNoRows) {
					return echo.NewHTTPError(http.StatusNotFound, "Could not find requested recipe.")
				}

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

			// read tags
			tagsQuery := models.Tags(
				models.TagWhere.Recipeid.EQ(null.Int64From(id)),
			)

			tags, err := tagsQuery.All(ctx, tx)
			if err != nil {
				return err
			}

			data := EditTemplate{
				Title:       fmt.Sprintf("Edit %v", recipe.Title.String),
				Error:       error,
				Recipe:      recipe,
				Ingredients: ingredients,
				Tags:        tags,
			}

			return c.Render(http.StatusOK, tmplName, data)
		},
	}
}

func editRecipe(db *sql.DB, id int64, c echo.Context) (path string, err error) {
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

	new_path := fmt.Sprintf("%v", id)
	if np := c.FormValue("Path"); np != "" {
		new_path = np
	}

	// update recipe
	recipe := models.Recipe{
		ID:           null.Int64From(id),
		Path:         null.StringFrom(new_path),
		Image:        null.BytesFrom(imgB),
		Title:        null.StringFrom(c.FormValue("Title")),
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
	ingredientText := c.FormValue("Ingredients")
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

	// delete previous tags
	tagsDeleteQuery := models.Tags(
		models.TagWhere.Recipeid.EQ(null.Int64From(id)),
	)

	if _, err = tagsDeleteQuery.DeleteAll(context.Background(), tx); err != nil {
		return
	}

	// insert new ingredients
	tagsText := c.FormValue("Tags")
	for _, t := range strings.Split(tagsText, ",") {
		tf := strings.ToLower(strings.TrimSpace(t))
		if tf == "" {
			continue
		}

		tag := models.Tag{
			Recipeid: null.Int64From(id),
			Tag:      null.StringFrom(tf),
		}

		err = tag.Insert(context.Background(), tx, boil.Infer())
		if err != nil {
			return
		}
	}

	path = recipe.Path.String
	return
}

func (self App) EditRecipe() registerable.Registration {
	return EchoHandlerRegistration{
		Path:        "/edit/:id",
		Methods:     []Method{POST},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// get id
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			// update recipe
			path, err := editRecipe(self.DB, id, c)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := fmt.Sprintf("/recipe/%v", path)
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}
