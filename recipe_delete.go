package main

import (
	"context"
	"database/sql"
	"fmt"
	"go-recipes/models"
	"net/http"
	"strconv"

	"github.com/arpachuilo/go-registerable"
	"github.com/gorilla/mux"
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

func (self Router) DeleteRecipe() registerable.Registration {
	return HandlerRegistration{
		Name:        "delete",
		Path:        "/delete/{id:[0-9]+}",
		Methods:     []string{"POST"}, // work with HTML form standards
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

			// delete recipe
			err = func() error {
				if err := deleteRecipe(self.DB, id); err != nil {
					return err
				}

				return nil

			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", r.Header.Get("Referer"), err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusFound)
		},
	}
}
