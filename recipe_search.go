package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"strconv"

	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SearchTemplate struct {
	Title   string
	Recipes models.RecipeSlice
	Search  string
	Offset  int
	Offsets []int
}

func (self Router) ServeSearch() Registration {
	// read templates dynamically for debug
	fullTemplate := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/recipe_search.html",
		"templates/recipe_search_results.html",
	))

	fragTemplate := template.Must(template.New("fragment").Funcs(templateFns).ParseFiles(
		"templates/recipe_search_results.html",
	))

	return HandlerRegistration{
		Name:    "search",
		Path:    "/",
		Methods: []string{"GET"},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			rq := r.URL.Query()
			// get search if any
			search := ""
			keys, ok := rq["search"]
			if ok && len(keys) > 0 {
				search = keys[0]
			}

			// get offset
			offset := 0
			keys, ok = rq["offset"]
			if ok && len(keys) > 0 {
				if value, err := strconv.Atoi(keys[0]); err == nil {
					offset = value
				}
			}

			// get limit
			limit := 15
			keys, ok = rq["limit"]
			if ok && len(keys) > 0 {
				if value, err := strconv.Atoi(keys[0]); err == nil {
					limit = value
				}
			}

			// check if this is an htmlx request
			htmx := false
			if value, err := strconv.ParseBool(r.Header.Get("HX-Request")); err == nil {
				htmx = value
			}

			// read recipes
			ctx := context.Background()
			tx, err := self.DB.BeginTx(ctx, nil)
			defer tx.Commit()
			if err != nil {
				return err
			}

			// get recipes
			filter := fmt.Sprintf("%%%v%%", search)
			query := models.Recipes(
				Distinct("recipes.id, title, url, instructions, author, total_time, yields, serving_size, calories, image"),
				LeftOuterJoin("ingredients i on i.recipeid = recipes.id"),
				Where("title like ? or i.ingredient like ?", filter, filter),
				OrderBy("title", "recipes.id"),
				Limit(limit),
				Offset(offset),
			)

			recipes, err := query.All(ctx, tx)
			if err != nil {
				return err
			}

			// get total for this query
			query = models.Recipes(
				Distinct("recipes.id"),
				LeftOuterJoin("ingredients i on i.recipeid = recipes.id"),
				Where("title like ? or i.ingredient like ?", filter, filter),
			)

			count, err := query.Count(ctx, tx)
			if err != nil {
				return err
			}

			// generate pages
			offsets := make([]int, 0)
			pages := (int(count) + limit - 1) / limit

			if pages > 1 {
				for i := 0; i < pages; i++ {
					offsets = append(offsets, i*limit)
				}
			}

			// run template
			data := &SearchTemplate{
				Title:   fmt.Sprintf("search recipes for %v", search),
				Search:  search,
				Recipes: recipes,
				Offset:  offset,
				Offsets: offsets,
			}

			if htmx {
				return fragTemplate.Execute(w, data)
			}

			return fullTemplate.Execute(w, data)
		},
	}
}
