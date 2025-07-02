package app

import (
	"authService/internal/domain"
	"authService/internal/handler"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func SetRouter(cfg *domain.Config, log *slog.Logger) (*http.Server, func()) {
	mux := http.NewServeMux()

	authH := handler.NewAuthHandler()

	mux.HandleFunc("GET /login", authH.Login)

	serv := http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	cleanup := func() {
		// Do cleanup
	}

	return &serv, cleanup
}

func StartServer(serv *http.Server, log *slog.Logger) {
	log.Info("Server started on " + serv.Addr)
	if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("Failed to start server", "error", err)
	}
}

func ListenShutdown(log *slog.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Info("Shutdown signal received!")
}
