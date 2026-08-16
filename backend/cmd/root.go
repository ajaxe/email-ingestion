package cmd

import (
	"context"
	"log/slog"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/spf13/cobra"
)

var cfgFile string

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return "."
}

var rootCmd = &cobra.Command{
	Use:   "email-ingestion",
	Short: "Single binary multi-service application for email ingestion and processing",
	Long: `Email Ingestion Gateway is a production-grade microservices suite designed to handle
inbound SMTP traffic, parse MIME emails, store attachments securely in S3, and deliver
webhooks to registered SaaS applications with strict multi-tenant isolation.`,
	Example: `  email-ingestion api --config /path/to/config.yaml
  email-ingestion smtp --config /path/to/config.yaml
  email-ingestion worker --streams email,webhook --config /path/to/config.yaml
  email-ingestion cron --config /path/to/config.yaml
  email-ingestion migrate up --config /path/to/config.yaml`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.LoadConfig(getConfigPath())
		if err != nil {
			return err
		}

		config.SetupLogger(cfg)
		slog.Info("Running root command...")

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to config file or directory (default \".\")")
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
