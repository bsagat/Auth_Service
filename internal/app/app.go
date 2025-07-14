package app

import (
	"auth/config"
	"auth/internal/adapters/repo"
	httpserver "auth/internal/adapters/transport/http"
	"auth/internal/service"
	"auth/pkg/logger"
	"auth/pkg/postgres"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const serviceName = "auth"

type App struct {
	httpServer *httpserver.API
	postgresDB *postgres.PostgreDB
	// grpcServer *grpcserver.API
}

func New(cfg config.Config, log *slog.Logger) (*App, error) {
	log.Info(fmt.Sprintf("Starting %s service", serviceName))

	log.Info("Starting database connection")
	postgresDB, err := postgres.Connect(cfg.Db, cfg.App.Admin)
	if err != nil {
		return nil, err
	}
	log.Info("connection established")

	userDal := repo.NewUserDal(postgresDB.DB)

	tokenServ := service.NewTokenService(cfg.App.Secret, userDal, cfg.App.RefreshTTL, cfg.App.AccessTTL, log)
	authServ := service.NewAuthService(userDal, tokenServ, log)
	adminServ := service.NewAdminService(userDal, tokenServ, log)

	httpServ := httpserver.New(cfg.Server, authServ, adminServ, tokenServ, log)

	return &App{
		httpServer: httpServ,
		postgresDB: postgresDB,
	}, nil
}

func (a *App) Start(log *slog.Logger) {
	go func() {
		if err := a.httpServer.StartServer(); err != nil {
			log.Error("Failed to start http server", logger.Err(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	s := <-stop
	log.Info(fmt.Sprintf("Catched %s signal!!!", s.String()))
	log.Info("Shutting down...")
}

func (a *App) CleanUp(log *slog.Logger) {
	if err := a.httpServer.Close(); err != nil {
		log.Error("Failed to close http server conn", logger.Err(err))
	}

	if err := a.postgresDB.DB.Close(); err != nil {
		log.Error("Failed to close database conn", logger.Err(err))
	}
}
