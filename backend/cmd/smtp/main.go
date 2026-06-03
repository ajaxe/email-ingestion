package main

import (
	"log/slog"
	"os"

	"github.com/ajaxe/email-ingestion/internal/smtp"
	"github.com/ajaxe/email-ingestion/internal/startup"
	"github.com/ajaxe/email-ingestion/pkg/config"
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

	s := smtp.NewSmtpServer(cfg)
	if err := s.ListenAndServe(); err != nil {
		slog.Error("Server structural failure", "error", err)
		os.Exit(1)
	}
}
