package config

import (
	"auth/pkg/postgres"
	"log/slog"
	"os"
	"time"

	"github.com/bsagat/envzilla/v2"
)

// App config settings
type (
	Config struct {
		App    AppConf
		Server ServerConf
		Db     postgres.DatabaseConf
	}

	AppConf struct {
		Env        string        `env:"ENV" default:"local"` // Application environment: local | dev | prod
		Secret     string        `env:"SECRET"`              // Token generate secret key
		RefreshTTL time.Duration `env:"ACCESSTTL"`
		AccessTTL  time.Duration `env:"REFRESHTTL"`
		Admin      postgres.AdminCredentials
	}

	ServerConf struct {
		Host string `env:"HOST" default:"localhost"` // Application Host
		Port string `env:"PORT" default:"80"`        // Application Port number
	}
)

func New() Config {
	if err := envzilla.Loader(".env"); err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	var cfg Config
	if err := envzilla.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env configuration", "error", err)
		os.Exit(1)
	}

	return cfg
}
