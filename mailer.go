package main

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type MailerConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	From string `mapstructure:"from"`
}

type Mailer struct {
	*MailerConfig
	*gomail.Dialer
}

func (self Mailer) Send(subject, body string, to ...string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", self.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return self.DialAndSend(m)
}

func NewMailer(conf MailerConfig, tlsConfig *tls.Config) *Mailer {
	d := gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password)
	d.TLSConfig = tlsConfig
	return &Mailer{&conf, d}
}
