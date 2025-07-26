package grpcserver

import (
	"auth/config"
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func GetOptions(cfg config.GrpcServer, log *slog.Logger) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor(log)), // Добавляем interceptor
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     cfg.KeepaliveIdle,
			MaxConnectionAge:      cfg.KeepaliveAge,
			MaxConnectionAgeGrace: cfg.KeepaliveGrace,
			Time:                  cfg.KeepalivePing,
			Timeout:               cfg.KeepaliveTimeout,
		}),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
	}
}

func unaryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		log.Info("gRPC request",
			"method", info.FullMethod,
			"request", fmt.Sprintf("%+v", req),
		)

		resp, err := handler(ctx, req)

		if err != nil {
			log.Error("gRPC error",
				"method", info.FullMethod,
				"error", err.Error(),
			)
		} else {
			log.Info("gRPC response",
				"method", info.FullMethod,
				"response", fmt.Sprintf("%+v", resp),
			)
		}

		return resp, err
	}
}
