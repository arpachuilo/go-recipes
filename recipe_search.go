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
	"github.com/labstack/echo/v4"
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

func (self App) ServeSearch() registerable.Registration {
	// read templates dynamically for debug
	tmplFullName := "recipe_search_results"
	fullTemplate := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_search.html",
		"templates/recipe_search_results.html",
	))

	self.TemplateRenderer.Add(tmplFullName, fullTemplate)

	tmplFragName := "recipe_search_results_partial"
	fragTemplate := template.Must(template.New("fragment").Funcs(templateFns).ParseFiles(
		"templates/recipe_search_results.html",
	))

	self.TemplateRenderer.Add(tmplFragName, fragTemplate)

	return EchoHandlerRegistration{
		Path:        "/",
		Methods:     []Method{GET},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			// get search if any
			search := c.QueryParam("search")

			// get tags if any
			var tags []string
			if keys, ok := c.QueryParams()["tags"]; ok && len(keys) > 0 {
				tags = keys
			}

			// get offset
			offset := 0
			if value := c.QueryParam("offset"); value != "" {
				if value, err := strconv.Atoi(value); err == nil {
					offset = value
				}
			}

			// get limit
			limit := 15
			if value := c.QueryParam("limit"); value != "" {
				if value, err := strconv.Atoi(value); err == nil {
					limit = value
				}
			}

			// check if this is an htmlx request
			htmx := false
			if value, err := strconv.ParseBool(c.Request().Header.Get("HX-Request")); err == nil {
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
				return c.Render(http.StatusOK, tmplFragName, data)
			}

			return c.Render(http.StatusOK, tmplFullName, data)
		},
	}
}
