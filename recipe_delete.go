package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-recipes/models"
	"net/http"
	"strconv"

	"github.com/arpachuilo/go-registrable"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func deleteRecipe(db *sql.DB, id int64) (err error) {
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

	// delete recipe
	recipeDeleteQuery := models.Recipes(
		models.RecipeWhere.ID.EQ(null.Int64From(id)),
	)

	if _, err = recipeDeleteQuery.DeleteAll(context.Background(), tx); err != nil {
		return
	}

	// delete ingredients
	ingredientsDeleteQuery := models.Ingredients(
		models.IngredientWhere.Recipeid.EQ(null.Int64From(id)),
	)

	if _, err = ingredientsDeleteQuery.DeleteAll(context.Background(), tx); err != nil {
		return
	}

	// delete tags
	tagsDeleteQuery := models.Tags(
		models.TagWhere.Recipeid.EQ(null.Int64From(id)),
	)

	if _, err = tagsDeleteQuery.DeleteAll(context.Background(), tx); err != nil {
		return
	}

	// delete orphaned tags
	orphanedTagsDeleteQuery := models.Tags(
		Where("not exists (select id from recipes where id = recipeid)"),
	)

	if _, err = orphanedTagsDeleteQuery.DeleteAll(context.Background(), tx); err != nil {
		return
	}

	return nil
}

func (self App) DeleteRecipe() registrable.Registration {
	return EchoHandlerRegistration{
		Path:        "/delete/:id",
		Methods:     []Method{POST}, // work with HTML form standards
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// get id
			id, err := strconv.ParseInt(c.Param("id"), 10, 64)
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			// delete recipe
			err = func() error {
				if err := deleteRecipe(self.DB, id); err != nil {
					return err
				}

				return nil

			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := "/"
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}
