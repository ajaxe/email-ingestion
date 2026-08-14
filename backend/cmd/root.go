package cmd

import (
	"context"
	"fmt"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "email-ingestion",
	Short: "Single binary multi-service application for email ingestion and processing",
	Long: `Email Ingestion Gateway is a production-grade microservices suite designed to handle
inbound SMTP traffic, parse MIME emails, store attachments securely in S3, and deliver
webhooks to registered SaaS applications with strict multi-tenant isolation.`,
	Example: `  email-ingestion api
  email-ingestion smtp
  email-ingestion worker --streams email,webhook
  email-ingestion cron
  email-ingestion migrate up`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running root command...")

		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		config.SetupLogger(cfg)

		return nil
	},
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
