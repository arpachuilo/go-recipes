package main

import (
	"bytes"
	"html/template"
	"net/http"
)

type FourOhFourHandler struct {
	Page *bytes.Buffer
}

func New404Handle(data any, tmpl *template.Template) *FourOhFourHandler {
	var page bytes.Buffer
	err := tmpl.Execute(&page, data)
	if err != nil {
		panic(err)
	}

	return &FourOhFourHandler{&page}
}

func (self FourOhFourHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	self.Page.WriteTo(w)
}

func (self Router) Serve404() Registration {
	data := struct {
		Title string
	}{
		Title: "Page Not Found",
	}

	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/404.html",
	))

	return HandlerRegistration{
		Is404:   true,
		Handler: New404Handle(data, tmpl),
	}
}
