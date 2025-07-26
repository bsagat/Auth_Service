package grpcserver

import (
	"auth/config"
	authv1 "auth/internal/adapters/transport/grpc/gen"
	"auth/internal/adapters/transport/grpc/routers"
	"auth/internal/service"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type API struct {
	server *grpc.Server
	cfg    config.GrpcServer

	log *slog.Logger
}

func New(cfg config.GrpcServer, authServ *service.AuthService, adminServ *service.AdminService, tokenServ *service.TokenService, log *slog.Logger) *API {
	grpcServer := grpc.NewServer(GetOptions(cfg, log)...)

	adminHandler := routers.NewAdminHandler(authServ, adminServ, log)
	authHandler := routers.NewAuthHandler(authServ, tokenServ, log)

	authv1.RegisterAdminServiceServer(grpcServer, adminHandler)
	authv1.RegisterAuthServiceServer(grpcServer, authHandler)

	return &API{
		server: grpcServer,
		cfg:    cfg,
		log:    log,
	}
}

func (a *API) StartServer() error {
	address := fmt.Sprintf(":%s", a.cfg.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		a.log.Error("Failed to listen on port", "port", a.cfg.Port, "error", err)
		return err
	}

	a.log.Info("gRPC server is running", "port", a.cfg.Port)
	return a.server.Serve(listener)
}

func (a *API) Close() {
	a.log.Info("Shutting down gRPC server...")
	a.server.GracefulStop()
}
