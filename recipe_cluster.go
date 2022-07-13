package main

import (
	"html/template"
	"net/http"

	"github.com/arpachuilo/go-registerable"
	"github.com/labstack/echo/v4"
)

func (self App) ServeCluster() registerable.Registration {
	// read templates dynamically for debug
	tmplFullName := "recipe_clustering"
	fullTemplate := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/recipe_search.html",
		"templates/recipe_cluster.html",
	))

	self.TemplateRenderer.Add(tmplFullName, fullTemplate)
	return EchoHandlerRegistration{
		Path:        "/cluster",
		Methods:     []Method{GET},
		RequireAuth: self.Auth.enabled,
		HandlerFunc: func(c echo.Context) error {
			return c.Render(http.StatusOK, tmplFullName, nil)
		},
	}
}
