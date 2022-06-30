package main

import (
	"net/http"
	"strings"
)

type Auth struct {
	key string
}

func NewAuth(key string) *Auth {
	return &Auth{key}
}

func (self *Auth) Authorize(r *http.Request) bool {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	return token == self.key
}

func (self *Auth) UseFunc(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !self.Authorize(r) {
			http.Redirect(w, r, "/token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}

func (self *Auth) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !self.Authorize(r) {
			http.Redirect(w, r, "/token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
