package main

import (
	"context"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/arpachuilo/go-registerable"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SearchTemplate struct {
	Title        string
	Recipes      models.RecipeSlice
	PossibleTags models.TagSlice
	SelectedTags []string
	Search       string
	Offset       int
	Offsets      []int
}

func (self Router) ServeSearch() registerable.Registration {
	// read templates dynamically for debug
	fullTemplate := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_search.html",
		"templates/recipe_search_results.html",
	))

	fragTemplate := template.Must(template.New("fragment").Funcs(templateFns).ParseFiles(
		"templates/recipe_search_results.html",
	))

	return HandlerRegistration{
		Name:        "search",
		Path:        "/",
		Methods:     []string{"GET"},
		RequireAuth: self.Auth.enabled,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			rq := r.URL.Query()
			// get search if any
			search := ""
			if keys, ok := rq["search"]; ok && len(keys) > 0 {
				search = keys[0]
			}

			// get tags if any
			var tags []string
			if keys, ok := rq["tags"]; ok && len(keys) > 0 {
				tags = keys
			}

			// get offset
			offset := 0
			if keys, ok := rq["offset"]; ok && len(keys) > 0 {
				if value, err := strconv.Atoi(keys[0]); err == nil {
					offset = value
				}
			}

			// get limit
			limit := 15
			if keys, ok := rq["limit"]; ok && len(keys) > 0 {
				if value, err := strconv.Atoi(keys[0]); err == nil {
					limit = value
				}
			}

			// check if this is an htmlx request
			htmx := false
			if value, err := strconv.ParseBool(r.Header.Get("HX-Request")); err == nil {
				htmx = value
			}

			// prepare db
			ctx := context.Background()
			tx, err := self.DB.BeginTx(ctx, nil)
			defer tx.Commit()
			if err != nil {
				return err
			}

			// get recipes
			filter := fmt.Sprintf("%%%v%%", search)
			args := make([]string, 0)
			for _, t := range tags {
				args = append(args, fmt.Sprintf("'%v'", t))
			}

			whereClause := "(title like ? or i.ingredient like ? or author like ?)"
			if len(args) > 0 {
				whereClause += `and (t.tag in (` + strings.Join(args, ",") + `))`
			}

			where := Where(whereClause, filter, filter, filter)
			query := models.Recipes(
				Distinct("recipes.id, path, title, url, instructions, author, total_time, yields, serving_size, calories, image"),
				LeftOuterJoin("ingredients i on i.recipeid = recipes.id"),
				LeftOuterJoin("tags t on t.recipeid = recipes.id"),
				where,
				OrderBy("title", "recipes.id"),
				Limit(limit),
				Offset(offset),
				GroupBy("recipes.id"),
				Having("count(distinct t.id) >= ?", len(tags)),
			)

			recipes, err := query.All(ctx, tx)
			if err != nil {
				fmt.Println(err)
				return err
			}

			// get total for this query
			countQuery := fmt.Sprintf(`
                          select count(*)
                          from (
                            select recipes.id
                            from recipes
                            left outer join ingredients i on i.recipeid = recipes.id
                            left outer join tags t on t.recipeid = recipes.id
                            where %v
                            group by recipes.id
                            having count(distinct t.id) >= ?
                          )`, whereClause)

			count := 0
			row := tx.QueryRow(countQuery, filter, filter, filter, len(tags))
			err = row.Scan(&count)
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

			// get tags
			queryTags := models.Tags(
				Distinct("tag"),
			)

			possibleTags, err := queryTags.All(ctx, tx)
			if err != nil {
				return err
			}

			// run template
			data := &SearchTemplate{
				Title:        fmt.Sprintf("search recipes for %v", search),
				Search:       search,
				Recipes:      recipes,
				PossibleTags: possibleTags,
				SelectedTags: tags,
				Offset:       offset,
				Offsets:      offsets,
			}

			if htmx {
				return fragTemplate.Execute(w, data)
			}

			return fullTemplate.Execute(w, data)
		},
	}
}
