package cmd

import (
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/smtp"
	"github.com/ajaxe/email-ingestion/internal/startup"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/spf13/cobra"
)

var smtpCmd = &cobra.Command{
	Use:   "smtp",
	Short: "Run the SMTP server",
	Long:  "Run the SMTP server",
	RunE:  func(cmd *cobra.Command, args []string) error { return runSMTP() },
}

func runSMTP() error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return err
	}

	slog.Info("starting application initialization")

	dbPool := startup.NewDbPool(cfg)
	defer dbPool.Close()

	queries := public.New(dbPool)
	cache := startup.NewCache()

	storageService := storage.NewStorageService(&cfg.Storage)

	s := smtp.NewSmtpServer(cfg, queries, cache, storageService)
	if err := s.ListenAndServe(); err != nil {
		slog.Error("Server structural failure", "error", err)
		// os.Exit(1)
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(smtpCmd)
}
