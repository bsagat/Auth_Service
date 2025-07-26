package config

import (
	"auth/pkg/postgres"
	"log/slog"
	"os"
	"time"

	"github.com/bsagat/envzilla/v2"
)

// App config settings
type (
	Config struct {
		App        AppConf
		HttpServer HttpServer
		GrpcServer GrpcServer
		Db         postgres.DatabaseConf
	}

	AppConf struct {
		Env        string                    `env:"ENV" default:"local"` // Application environment: local | dev | prod
		Secret     string                    `env:"SECRET"`              // Token generation secret
		AccessTTL  time.Duration             `env:"ACCESSTTL"`           // Access token TTL
		RefreshTTL time.Duration             `env:"REFRESHTTL"`          // Refresh token TTL
		Admin      postgres.AdminCredentials // Admin credentials
	}

	HttpServer struct {
		Port              string        `env:"HTTP_PORT" default:"80"`                  // HTTP server port
		Host              string        `env:"HOST" default:"localhost"`                // HTTP server host
		ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" default:"15s"`         // Max duration for reading request
		IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" default:"60s"`         // Max keep-alive idle duration
		ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" default:"10s"`  // Max duration for reading headers
		MaxHeaderBytes    int           `env:"HTTP_MAX_HEADER_BYTES" default:"1048576"` // Max size of request headers in bytes (default 1MB)
	}

	GrpcServer struct {
		Port             string        `env:"GRPC_PORT" default:"50051"`                 // gRPC server port
		Host             string        `env:"HOST" default:"localhost"`                  // gRPC server host
		MaxRecvMsgSize   int           `env:"GRPC_MAX_RECV_MSG_SIZE" default:"10485760"` // Max gRPC receive message size (default 10MB)
		MaxSendMsgSize   int           `env:"GRPC_MAX_SEND_MSG_SIZE" default:"10485760"` // Max gRPC send message size (default 10MB)
		KeepaliveIdle    time.Duration `env:"GRPC_KEEPALIVE_IDLE" default:"5m"`          // Time before sending keepalive ping
		KeepaliveAge     time.Duration `env:"GRPC_KEEPALIVE_AGE" default:"30m"`          // Max connection age
		KeepaliveGrace   time.Duration `env:"GRPC_KEEPALIVE_GRACE" default:"5m"`         // Grace period after max connection age
		KeepalivePing    time.Duration `env:"GRPC_KEEPALIVE_PING" default:"1m"`          // Interval between pings
		KeepaliveTimeout time.Duration `env:"GRPC_KEEPALIVE_TIMEOUT" default:"20s"`      // Time to wait for ping ack
	}
)

func New() Config {
	if err := envzilla.Loader(".env"); err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	var cfg Config
	if err := envzilla.Parse(&cfg); err != nil {
		slog.Error("Failed to parse env configuration", "error", err)
		os.Exit(1)
	}

	return cfg
}
