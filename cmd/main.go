package main

import (
	"authService/internal/app"
)

func main() {
	cfg := app.SetConfig()

	logger := app.SetLogger(cfg.Env)
	logger.Info("Logger setup finished...")

	serv, cleanup := app.SetRouter(cfg, logger)
	defer cleanup()

	go app.StartServer(serv, logger)

	app.ListenShutdown(logger)
}
