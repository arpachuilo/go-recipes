package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"

	_ "github.com/mattn/go-sqlite3"
)

type AutoCertConfig struct {
	Email string   `mapstructure:"email"`
	Hosts []string `mapstructure:"hosts"`
}

type ServerConfig struct {
	HTTPS         bool           `mapstructure:"https"`
	Autocert      AutoCertConfig `mapstructure:"autocert"`
	Address       string         `mapstructure:"address"`
	ReadTimeout   time.Duration  `mapstructure:"read_timeout"`
	WriteTimeout  time.Duration  `mapstructure:"write_timeout"`
	IdleTimeout   time.Duration  `mapstructure:"idle_timeout"`
	RateLimit     *int           `mapstructure:"rate_limit"`
	AssetsDir     string         `mapstructure:"assets_dir"`
	EnableLogging bool           `mapstructure:"enable_logging"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Mailer   MailerConfig   `mapstructure:"mailer"`
}

// TODO: Look into improvements to prevent multiple db reads on image serving
// TODO: Look into template context to set/unset navigation
func main() {
	// load config
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("viper.ReadInConfig: %w", err))
	}

	conf := &Config{}
	err = viper.Unmarshal(conf)

	// create + start server
	NewApp(conf).Start()
}
