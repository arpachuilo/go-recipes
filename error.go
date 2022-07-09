package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

type FiveHundredHandler struct {
	tmplName string
}

func New500Handle(tr *TemplateRenderer) *FiveHundredHandler {
	// read templates dynamically for debug
	tmplName := "error"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/error.html",
	))

	tr.Add(tmplName, tmpl)

	return &FiveHundredHandler{tmplName}
}

type ErrorTemplate struct {
	Title  string
	Code   int
	Status string
	Err    string
}

func (self FiveHundredHandler) errorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	data := ErrorTemplate{
		Title:  fmt.Sprintf("Error %v", code),
		Code:   code,
		Status: http.StatusText(code),
		Err:    err.Error(),
	}

	c.Logger().Error(err)
	if err := c.Render(code, self.tmplName, data); err != nil {
		c.Logger().Error(err)
	}
}
