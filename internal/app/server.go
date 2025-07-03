package app

import (
	"authService/internal/dal"
	"authService/internal/domain"
	"authService/internal/handler"
	"authService/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func SetRouter(cfg *domain.Config, log *slog.Logger) (*http.Server, func()) {
	mux := http.NewServeMux()

	UserDal := ConnectAdapters(cfg.Db, log)

	tokenServ := service.NewTokenService(cfg.Secret, UserDal, cfg.RefreshTTL, cfg.AccessTTL, log)
	authServ := service.NewAuthService(UserDal, tokenServ, log)

	authH := handler.NewAuthHandler(authServ, tokenServ, log)

	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("POST /register", authH.Register)
	mux.HandleFunc("POST /isAdmin", authH.IsAdmin)

	serv := http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	cleanup := func() {
		UserDal.Db.Close()
		serv.Close()
	}

	return &serv, cleanup
}

func ConnectAdapters(config domain.DatabaseConf, log *slog.Logger) *dal.UserDal {
	db, err := dal.Connect(config)
	if err != nil {
		log.Error("Failed to connect database", "error", err)
		os.Exit(1)
	}

	log.Info("Adapters connection finished...")
	return dal.NewUserDal(db)
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

	sign := <-stop
	log.Info("Shutdown signal received!", "signal", sign.String())
}
