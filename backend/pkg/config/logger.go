package config

import (
	"log"
	"log/slog"
	"os"
	"strings"
)

func SetupLogger(cfg *AppConfig) {
	logLevel := slog.LevelInfo
	if cfg != nil && strings.ToLower(cfg.LogLevel) == "debug" {
		logLevel = slog.LevelDebug
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
	log.SetOutput(slog.NewLogLogger(jsonHandler, logLevel).Writer())
}
