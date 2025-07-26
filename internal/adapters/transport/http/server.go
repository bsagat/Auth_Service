package httpserver

import (
	"auth/config"
	"auth/internal/adapters/transport/http/routers"
	"auth/internal/service"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	httpSwagger "github.com/swaggo/http-swagger"
)

type API struct {
	server *http.Server
	cfg    config.ServerConf

	adminHandler *routers.AdminHandler
	authHandler  *routers.AuthHandler
	log          *slog.Logger
}

func New(cfg config.ServerConf, authServ *service.AuthService, adminServ *service.AdminService, tokenServ *service.TokenService, log *slog.Logger) *API {
	mux := http.NewServeMux()
	SetSwagger(mux)

	authH := routers.NewAuthHandler(authServ, tokenServ, log)
	adminH := routers.NewAdminHandler(authServ, adminServ, log)

	mux.HandleFunc("POST /login", authH.Login)
	mux.HandleFunc("POST /register", authH.Register)
	mux.HandleFunc("POST /refresh", authH.RefreshToken)
	mux.HandleFunc("GET /role", authH.CheckRole)

	// Admin rights
	mux.HandleFunc("PUT /user", adminH.UpdateUser)
	mux.HandleFunc("GET /user/{id}", adminH.GetUser)
	mux.HandleFunc("DELETE /user/{id}", adminH.DeleteUser)

	serv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: mux,
	}

	return &API{
		server:       serv,
		cfg:          cfg,
		adminHandler: adminH,
		authHandler:  authH,
		log:          log,
	}
}

func SetSwagger(mux *http.ServeMux) {
	swaggerBytes, err := os.ReadFile("docs/swagger.json")
	if err != nil {
		slog.Error("Failed to read swagger docs", "error", err)
		os.Exit(1)
	}

	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
	))
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/openapi+json")
		if _, err := w.Write(swaggerBytes); err != nil {
			slog.Error("Failed to send swagger file", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (a *API) StartServer() error {
	const op = "httpserver.StartServer"

	a.log.Info("Server started on " + a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *API) Close() error {
	return a.server.Close()
}
