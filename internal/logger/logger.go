package logger

import (
	"log/slog"
	"os"

	colorlogger "github.com/zhavkk/L0-test-service/pkg/logger"
)

var Log *slog.Logger

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Init(env string) {
	var handler slog.Handler

	switch env {
	case envLocal:
		handler = colorlogger.NewColorHandler(slog.LevelDebug)
	case envDev:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	Log = slog.New(handler)
}
