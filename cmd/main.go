package main

import (
	"auth/config"
	"auth/internal/app"
	"auth/pkg/logger"
	"os"
)

func main() {

	cfg := config.New()

	log := logger.SetLogger(cfg.App.Env)
	log.Info("Logger setup finished...")

	app, err := app.New(cfg, log)
	if err != nil {
		log.Error("Failed to setup application", logger.Err(err))
		os.Exit(1)
	}
	defer app.CleanUp(log)

	app.Start(log)
}
