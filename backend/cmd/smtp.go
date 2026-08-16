package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/ingest/client"
	"github.com/ajaxe/email-ingestion/internal/smtp"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/spf13/cobra"
)

var smtpCmd = &cobra.Command{
	Use:   "smtp",
	Short: "Run the inbound SMTP edge server",
	Long: `Run the non-blocking SMTP edge daemon. Receives inbound SMTP connections, performs 
perimeter recipient validation against the API/cache, and streams raw email payloads 
to the internal ingestion API.`,
	Example: `  email-ingestion smtp`,
	Args:    cobra.NoArgs,
	RunE:    func(cmd *cobra.Command, args []string) error { return runSMTP(cmd.Context()) },
}

func runSMTP(ctx context.Context) error {
	cfg, err := config.LoadConfig(getConfigPath())
	if err != nil {
		return err
	}

	slog.Info("starting SMTP initialization")

	ingestClient := client.NewIngestClient(cfg.Smtp.ApiURL, cfg.Smtp.MTAAuthToken)

	s := smtp.NewSmtpServer(&cfg.Smtp, ingestClient)

	go func() {
		if err := s.ListenAndServe(); err != nil {
			slog.Error("shutting down smtp server", "error", err)
		}
	}()

	<-ctx.Done()

	slog.InfoContext(ctx, "graceful shutdown of smtp requested")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = s.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "graceful shutdown of smtp failed", "error", err)
	} else {
		slog.InfoContext(ctx, "graceful shutdown of smtp completed")
	}

	return err
}

func init() {
	rootCmd.AddCommand(smtpCmd)
}
