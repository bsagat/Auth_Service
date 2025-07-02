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
	helpF = flag.Bool("help", false, "Shows help message")
	portF = flag.String("port", "80", "Default auth service port number")
	hostF = flag.String("host", "localhost", "Default auth service host settings")
	envF  = flag.String("env", "local", "Application environment: local | dev | prod")
	ttlF  = flag.String("tokenttl", "10m", "JWT Token time to live data")
)

func SetConfig() *domain.Config {
	ParseConfig()

	cfg := domain.Config{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
		Env:  os.Getenv("ENV"),
		Addr: fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT")),
	}

	var err error
	if cfg.TokenTTL, err = time.ParseDuration(os.Getenv("TOKENTTL")); err != nil {
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
		"PORT":     portF,
		"HOST":     hostF,
		"TOKENTTL": ttlF,
		"ENV":      envF,
	}

	for key, value := range args {
		if err := os.Setenv(key, *value); err != nil {
			slog.Error("Failed to set env value", "error", err)
			os.Exit(1)
		}
	}

}

func ShowHelp() {
	text :=
		`Auth Service
Flags:		
	--help 	   [ Shows help message ]
	--port     [ Default auth service port number ]
	--host     [ Default auth service host settings ]
	--tokenttl [ JWT Token time to live data ]
	--env      [ Application environment: local | dev | prod ]`
	fmt.Println(text)
	os.Exit(0)
}
