package main

import (
	"bytes"
	"html/template"
	"net/http"
)

type ErrorHandlerFunc = func(http.ResponseWriter, *http.Request) error

func Handle500Middleware(fn ErrorHandlerFunc) http.HandlerFunc {
	data := struct {
		Title string
	}{
		Title: "Something Went Wrong",
	}

	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/nav.html",
		"templates/500.html",
	))

	var page bytes.Buffer
	err := tmpl.Execute(&page, data)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			page.WriteTo(w)
		}
	}
}
