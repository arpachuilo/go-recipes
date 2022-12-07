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

	"github.com/arpachuilo/go-registrable"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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

type AuthContext struct {
	echo.Context
	username string
}

func NewAuth(conf AuthConfig) *Auth {
	return &Auth{conf.MagicLinkHost, conf.Enabled, []byte(conf.Secret), conf.VerificationExpiresAfter, conf.TokenName, conf.TokenExpiresAfter}
}

func (self *Auth) ReadToken(c echo.Context) *jwt.Token {
	cookie, err := c.Cookie(self.tokenName)
	if err != nil {
		return nil
	}

	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return self.secret, nil
	})
	if err != nil {
		return nil
	}

	return token
}

func (self *Auth) Use(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ca := &AuthContext{Context: c}
		if self.enabled {
			token := self.ReadToken(c)
			if token == nil || !token.Valid {
				c.Response().Header().Add("Content-Type", "text/html; charset=UTF-8")
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if subject, ok := claims["sub"].(string); ok {
					ca.username = subject
				}
			}
		}

		return next(ca)
	}
}

func (self *Auth) Skip(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ca := &AuthContext{Context: c}
		return next(ca)
	}
}

type LoginTemplate struct {
	Title string
	Error string
}

func (self App) ServeLogin() registrable.Registration {
	// read templates dynamically for debug
	tmplName := "login"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.gohtml",
		"templates/empty_nav.gohtml",
		"templates/login.gohtml",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)

	return EchoHandlerRegistration{
		Path:    "/login",
		Methods: []Method{GET},
		HandlerFunc: func(c echo.Context) error {
			// get error if any
			error := c.QueryParam("error")
			data := LoginTemplate{
				Title: "Login",
				Error: error,
			}

			return c.Render(http.StatusOK, tmplName, data)
		},
	}
}

type MagicLinkTemplate struct {
	URL    template.URL
	Expiry string
}

func (self App) SendLink() registrable.Registration {
	tmplName := "magic_link"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/magic_link.gohtml",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)

	return EchoHandlerRegistration{
		Path:    "/send-link",
		Methods: []Method{POST},
		HandlerFunc: func(c echo.Context) error {
			err := func() error {
				// get email
				email := c.FormValue("email")
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
				redirectURL := fmt.Sprintf("%v?error=%v", c.Request().Referer(), err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := "/link-sent"
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}

func (self App) ServeLinkSent() registrable.Registration {
	tmplName := "link_sent"
	tmpl := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
		"templates/base.gohtml",
		"templates/empty_nav.gohtml",
		"templates/link_sent.gohtml",
	))

	self.TemplateRenderer.Add(tmplName, tmpl)

	return EchoHandlerRegistration{
		Path:    "/link-sent",
		Methods: []Method{GET},
		HandlerFunc: func(c echo.Context) error {
			data := LoginTemplate{
				Title: "Magic Link Sent",
			}

			return c.Render(http.StatusOK, tmplName, data)
		},
	}
}

func (self App) VerifyLink() registrable.Registration {
	return EchoHandlerRegistration{
		Path:    "/verify-link",
		Methods: []Method{GET},
		HandlerFunc: func(c echo.Context) error {
			err := func() error {
				// parse form
				verificationCode := c.QueryParam("verification_code")
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

				c.SetCookie(&http.Cookie{
					Name:    "token",
					Value:   tokenString,
					Expires: expiresAt,
				})

				return nil
			}()
			if err != nil {
				redirectURL := fmt.Sprintf("%v?error=%v", "/login", err)
				return c.Redirect(http.StatusFound, redirectURL)
			}

			redirectURL := "/"
			return c.Redirect(http.StatusFound, redirectURL)
		},
	}
}
