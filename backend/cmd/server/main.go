package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	config.SetupLogger(cfg)

	slog.Info("starting application initialization")

	dbURL := cfg.Database.DSN
	if dbURL == "" {
		slog.Error("database connection string is empty in configuration")
		os.Exit(1)
	}
	ctxDB := context.Background()
	dbPool, err := pgxpool.New(ctxDB, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctxDB); err != nil {
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to database successfully")
}
