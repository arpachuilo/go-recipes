package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"strconv"

	"github.com/arpachuilo/go-registrable"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type RecipeTemplate struct {
	Title       string
	Recipe      *models.Recipe
	Ingredients models.IngredientSlice
	Tags        models.TagSlice
	Comments    models.CommentSlice
}

func (self App) ServeRecipeDisplay() registrable.Registration {
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

			// read comments
			commentsQuery := models.Comments(
				models.CommentWhere.Recipeid.EQ(recipe.ID),
			)

			comments, err := commentsQuery.All(ctx, tx)
			if err != nil {
				return err
			}

			// run template
			data := &RecipeTemplate{
				Title:       fmt.Sprintf("%v", recipe.Title.String),
				Recipe:      recipe,
				Ingredients: ingredients,
				Tags:        tags,
				Comments:    comments,
			}

			return c.Render(http.StatusOK, tmplName, data)
		},
	}
}

func createComment(db *sql.DB, c echo.Context) (string, error) {
	message := c.FormValue("comment")
	recipe_id := c.FormValue("recipe_id")
	recipe_path := c.FormValue("recipe_path")
	rid, err := strconv.ParseInt(recipe_id, 10, 64)
	if err != nil {
		return "", err
	}

	cc := c.(*AuthContext)
	comment := models.Comment{
		Recipeid: null.Int64From(rid),
		Comment:  null.StringFrom(message),
		Who:      null.StringFrom(cc.username),
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	defer tx.Commit()
	if err != nil {
		return "", err
	}

	err = comment.Insert(ctx, tx, boil.Infer())
	if err != nil {
		return "", err
	}

	return recipe_path, nil
}

func (self App) AddCommentRecipe() registrable.Registration {
	return EchoHandlerRegistration{
		Path:        "/comment",
		Methods:     []Method{POST},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			path, err := createComment(self.DB, c)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := fmt.Sprintf("/recipe/%v", path)
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}

func deleteComment(db *sql.DB, c echo.Context) (string, error) {
	comment_id := c.FormValue("comment_id")
	recipe_path := c.FormValue("recipe_path")
	id, err := strconv.ParseInt(comment_id, 10, 64)
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	defer tx.Commit()
	if err != nil {
		return "", err
	}
	// read comments
	commentsQuery := models.Comments(
		models.CommentWhere.ID.EQ(null.Int64From(id)),
	)

	comment, err := commentsQuery.One(ctx, tx)
	if err != nil {
		return "", err
	}

	cc := c.(*AuthContext)

	if cc.username != comment.Who.String {
		return "", errors.New("invalid user")
	}

	_, err = comment.Delete(ctx, tx)
	if err != nil {
		return "", err
	}

	return recipe_path, nil
}

func (self App) RemoveCommentRecipe() registrable.Registration {
	return EchoHandlerRegistration{
		Path:        "/comment-delete",
		Methods:     []Method{POST},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			path, err := deleteComment(self.DB, c)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := fmt.Sprintf("/recipe/%v", path)
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}
