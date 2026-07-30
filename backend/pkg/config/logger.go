package config

import (
	"log/slog"
	"os"
	"strings"
)

func SetupLogger(cfg *AppConfig) {
	logLevel := slog.LevelInfo
	if strings.ToLower(cfg.LogLevel) == "debug" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)
}
