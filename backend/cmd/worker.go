package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/internal/worker/handler"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var streamNames []string

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run stream processor workers",
	Long: `Run background stream processor workers that continuously poll Redis streams for new jobs. 
The --streams flag determines which streams this worker instance will consume (e.g., "email" for MIME parsing 
and raw email ingestion, or "webhook" for HTTP webhook dispatch).`,
	Example: `  email-ingestion worker --streams email
  email-ingestion worker --streams webhook
  email-ingestion worker --streams email,webhook`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error { return runWorker(cmd.Context()) },
}

func runWorker(ctx context.Context) error {
	cfg, err := config.LoadConfig(getConfigPath())
	if err != nil {
		return err
	}

	slog.Info("starting worker initialization", "streams", streamNames)

	dbPool := database.NewDbPool(cfg)
	defer dbPool.Close()

	queries := public.New(dbPool)

	rdsManager := redis.NewManager(cfg)
	defer rdsManager.Close()

	storageService := storage.NewStorageService(&cfg.Storage)

	var consumers []*worker.StreamConsumer

	for _, s := range streamNames {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		emailStreamName := util.EmailStreamName(cfg.Environment)
		webhookStreamName := util.WebhookStreamName(cfg.Environment)

		switch s {
		case "email":
			emailRepo := service.NewPgxEmailRepository(dbPool)
			emailProcessor := service.NewEmailProcessor(emailRepo, storageService, rdsManager.Stream, webhookStreamName)
			emailHandler := handler.NewEmailIngestionHandler(emailProcessor)
			consumer := worker.NewStreamConsumer(rdsManager, emailStreamName, "email_worker_group", emailHandler)
			consumers = append(consumers, consumer)

		case "webhook":
			webhookProcessor := service.NewWebhookDeliveryProcessor(queries, storageService, &cfg.Webhook)
			webhookHandler := handler.NewWebhookDeliveryHandler(webhookProcessor)
			consumer := worker.NewStreamConsumer(rdsManager, webhookStreamName, "webhook_worker_group", webhookHandler)
			consumers = append(consumers, consumer)

		default:
			return fmt.Errorf("unknown stream type requested: %s", s)
		}
	}

	if len(consumers) == 0 {
		return fmt.Errorf("no streams configured to consume")
	}

	g, gCtx := errgroup.WithContext(ctx)

	for _, c := range consumers {
		// capture the loop variable for the goroutine
		consumer := c
		g.Go(func() error {
			return consumer.Start(gCtx)
		})
	}

	// Also catch termination signals from the host
	go func() {
		<-ctx.Done()
		slog.InfoContext(ctx, "graceful shutdown of worker requested")
	}()

	return g.Wait()
}

func init() {
	workerCmd.Flags().StringSliceVar(&streamNames, "streams", []string{"email"}, "Comma-separated list of streams to consume (e.g., email,webhook)")
	rootCmd.AddCommand(workerCmd)
}
