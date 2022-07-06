package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go-recipes/models"
	"html/template"
	"net/http"
	"time"

	"github.com/arpachuilo/go-registerable"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type AuthConfig struct {
	MagicLinkHost            string        `mapstructure:"magic_link_host"`
	Enabled                  bool          `mapstructure:"enabled"`
	Secret                   string        `mapstructure:"secret"`
	VerificationExpiresAfter time.Duration `mapstructure:"verification_expires_after"`

	TokenName         string        `mapstructure:"token_name"`
	TokenExpiresAfter time.Duration `mapstructure:"token_expires_after"`
}

type Auth struct {
	magicLinkHost            string
	enabled                  bool
	secret                   []byte
	verificationExpiresAfter time.Duration
	tokenName                string
	tokenExpiresAfter        time.Duration
}

func NewAuth(conf AuthConfig) *Auth {
	return &Auth{conf.MagicLinkHost, conf.Enabled, []byte(conf.Secret), conf.VerificationExpiresAfter, conf.TokenName, conf.TokenExpiresAfter}
}

func (self *Auth) VerifyToken(r *http.Request) bool {
	cookie, err := r.Cookie(self.tokenName)
	if err != nil {
		return false
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return self.secret, nil
	})
	if err != nil {
		return false
	}

	return token.Valid
}

func (self *Auth) UseFunc(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if self.enabled && !self.VerifyToken(r) {
			w.Header().Add("Content-Type", "text/html; charset=UTF-8")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		next(w, r)
	})
}

func (self *Auth) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if self.enabled && !self.VerifyToken(r) {
			w.Header().Add("Content-Type", "text/html; charset=UTF-8")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type LoginTemplate struct {
	Title string
	Error string
}

func (self Router) ServeLogin() registerable.Registration {
	// read templates dynamically for debug
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/empty_nav.html",
		"templates/login.html",
	))

	return HandlerRegistration{
		Name:    "login",
		Path:    "/login",
		Methods: []string{"GET"},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			// get error if any
			rq := r.URL.Query()
			error := ""
			keys, ok := rq["error"]
			if ok && len(keys) > 0 {
				error = keys[0]
			}

			data := LoginTemplate{
				Title: "Login",
				Error: error,
			}

			return tmpl.Execute(w, data)
		},
	}
}

type MagicLinkTemplate struct {
	URL    template.URL
	Expiry string
}

func (self Router) SendLink() registerable.Registration {
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/magic_link.html",
	))

	return HandlerRegistration{
		Name:    "send link",
		Path:    "/send-link",
		Methods: []string{"POST"},
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				r.ParseForm()
				// get email
				email := r.Form.Get("email")
				if email == "" {
					return errors.New("email missing")
				}

				// check email exist
				// prepare db
				ctx := context.Background()
				tx, err := self.DB.BeginTx(ctx, nil)
				defer tx.Commit()
				if err != nil {
					return err
				}

				// read user by email
				userQuery := models.Users(
					models.UserWhere.Email.EQ(null.StringFrom(email)),
				)

				user, err := userQuery.One(ctx, tx)
				if err != nil {
					return err
				}

				// create verification code
				verificationCode, err := uuid.NewRandom()
				if err != nil {
					return err
				}

				// get expiry for verification code
				expiration := time.Now().Add(self.Auth.verificationExpiresAfter)

				// update user
				user.VerificationCode = null.StringFrom(verificationCode.String())
				user.VerificationCodeExpiry = null.TimeFrom(expiration)

				_, err = user.Update(ctx, tx, boil.Infer())
				if err != nil {
					return err
				}

				data := MagicLinkTemplate{
					URL:    template.URL(fmt.Sprintf("%v/verify-link?verification_code=%v", self.magicLinkHost, verificationCode)),
					Expiry: fmt.Sprintf("%s", self.Auth.verificationExpiresAfter),
				}

				var body bytes.Buffer
				if err := tmpl.Execute(&body, data); err != nil {
					return err
				}

				return self.Mailer.Send("Magic Link for Recipes DB", body.String(), email)
			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", r.Header.Get("Referer"), err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			redirectURL := "/link-sent"
			http.Redirect(w, r, redirectURL, http.StatusFound)
		},
	}
}

func (self Router) ServeLinkSent() registerable.Registration {
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.html",
		"templates/empty_nav.html",
		"templates/link_sent.html",
	))

	return HandlerRegistration{
		Name:    "link sent",
		Path:    "/link-sent",
		Methods: []string{"GET"},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request) error {
			data := LoginTemplate{
				Title: "Magic Link Sent",
			}

			return tmpl.Execute(w, data)
		},
	}
}

func (self Router) VerifyLink() registerable.Registration {
	return HandlerRegistration{
		Name:    "verify link",
		Path:    "/verify-link",
		Methods: []string{"GET"},
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				// parse form
				verificationCode := r.URL.Query().Get("verification_code")
				if verificationCode == "" {
					return errors.New("empty verification code")
				}

				// check verification code exists in database
				// prepare db
				ctx := context.Background()
				tx, err := self.DB.BeginTx(ctx, nil)
				defer tx.Commit()
				if err != nil {
					return err
				}

				// read user by email
				userQuery := models.Users(
					models.UserWhere.VerificationCode.EQ(null.StringFrom(verificationCode)),
				)

				user, err := userQuery.One(ctx, tx)
				if err != nil {
					return err
				}

				// validate code
				if !user.VerificationCode.Valid {
					return errors.New("invalid code")
				}

				if !user.VerificationCodeExpiry.Valid {
					return errors.New("invalid expiration")
				}

				if user.VerificationCodeExpiry.Time.Before(time.Now()) {
					return errors.New("expired code")
				}

				// clear used code
				user.VerificationCode = null.String{}
				user.VerificationCodeExpiry = null.Time{}
				_, err = user.Update(ctx, tx, boil.Infer())
				if err != nil {
					return err
				}

				// set JWT in cookies
				expiresAt := time.Now().Add(time.Hour * 24 * 60)
				tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
					Subject:   user.Email.String,
					ExpiresAt: expiresAt.Unix(),
				}).SignedString(self.Auth.secret)
				if err != nil {
					return err
				}

				http.SetCookie(w, &http.Cookie{
					Name:    "token",
					Value:   tokenString,
					Expires: expiresAt,
				})

				return nil
			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", "/login", err)
				http.Redirect(w, r, redirectURL, http.StatusFound)
				return
			}

			redirectURL := "/"
			http.Redirect(w, r, redirectURL, http.StatusFound)
		},
	}
}
