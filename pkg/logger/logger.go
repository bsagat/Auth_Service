package logger

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Logger interface {
	Info(msg string, args ...any)
}

func SetLogger(env string) *slog.Logger {
	var log slog.Logger

	switch env {
	case envLocal:
		log = *slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = *slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = *slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &log
}

// Возвращает лог-обертку для ошибки
func Err(err error) slog.Attr {
	return slog.String("error", err.Error())
}
