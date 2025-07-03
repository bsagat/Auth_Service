package app

import (
	"authService/internal/domain"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bsagat/envzilla"
)

var (
	helpF          = flag.Bool("help", false, "Shows help message")
	portF          = flag.String("port", "80", "Default auth service port number")
	hostF          = flag.String("host", "localhost", "Default auth service host settings")
	envF           = flag.String("env", "local", "Application environment: local | dev | prod")
	adminNameF     = flag.String("name", "", "Administrator name field")
	adminPasswordF = flag.String("password", "", "Administrator password field")
	adminEmailF    = flag.String("email", "", "Administrator email field")
)

func SetConfig() *domain.Config {
	ParseConfig()

	cfg := domain.Config{
		Host:   os.Getenv("HOST"),
		Port:   os.Getenv("PORT"),
		Env:    os.Getenv("ENV"),
		Addr:   fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),
		Secret: os.Getenv("SECRET"),
		Db: domain.DatabaseConf{
			Name:     os.Getenv("DB_NAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Port:     os.Getenv("DB_PORT"),
			UserName: os.Getenv("DB_USER"),
		},
	}

	var err error
	if cfg.RefreshTTL, err = time.ParseDuration(os.Getenv("REFRESHTTL")); err != nil {
		slog.Error("Failed to parse duration", "error", err)
		os.Exit(1)
	}
	if cfg.AccessTTL, err = time.ParseDuration(os.Getenv("ACCESSTTL")); err != nil {
		slog.Error("Failed to parse duration", "error", err)
		os.Exit(1)
	}

	return &cfg
}

// ParseConfig loads ENV/CLI with priority ENV >> CLI
func ParseConfig() {
	flag.Parse()
	if *helpF {
		ShowHelp()
	}

	if err := envzilla.Loader(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Error("Failed to load configuration", "error", err)
			os.Exit(1)
		}

		slog.Warn("Configuration file is not found...")
		slog.Info("Parsing CLI args...")
		ParseFlags()
	}
}

func ParseFlags() {
	args := map[string]*string{
		"PORT":           portF,
		"HOST":           hostF,
		"ENV":            envF,
		"ADMIN_NAME":     adminNameF,
		"ADMIN_PASSWORD": adminPasswordF,
		"ADMIN_EMAIL":    adminEmailF,
	}
	if err := CheckAdminCredentials(args); err != nil {
		slog.Error("Invalid admin credentials", "error", err)
		os.Exit(1)
	}

	for key, value := range args {
		if err := os.Setenv(key, *value); err != nil {
			slog.Error("Failed to set env value", "error", err)
			os.Exit(1)
		}
	}

}

func CheckAdminCredentials(args map[string]*string) error {
	if len(*args["ADMIN_NAME"]) == 0 {
		return errors.New("admin name is required in (ENV/CLI)")
	}
	if len(*args["ADMIN_PASSWORD"]) == 0 {
		return errors.New("admin password is required in (ENV/CLI)")
	}
	if len(*args["ADMIN_EMAIL"]) == 0 {
		return errors.New("admin email is required in (ENV/CLI)")
	}
	return nil
}

func ShowHelp() {
	text :=
		`Auth Service
Flags:		
	--help 	   [ Shows help message ]
	--port     [ Default auth service port number ]
	--host     [ Default auth service host settings ]
	--env      [ Application environment: local | dev | prod ]
	--name 	   []
	--password []
	--email    []`
	fmt.Println(text)
	os.Exit(0)
}
