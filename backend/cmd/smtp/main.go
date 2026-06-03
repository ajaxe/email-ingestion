package main

import (
	"log/slog"
	"os"

	"github.com/ajaxe/email-ingestion/internal/smtp"
	"github.com/ajaxe/email-ingestion/internal/startup"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	config.SetupLogger(cfg)

	slog.Info("starting application initialization")

	dbPool := startup.NewDbPool(cfg)
	defer dbPool.Close()

	queries := public.New(dbPool)
	cache := startup.NewCache()

	s := smtp.NewSmtpServer(cfg, queries, cache)
	if err := s.ListenAndServe(); err != nil {
		slog.Error("Server structural failure", "error", err)
		os.Exit(1)
	}
}
