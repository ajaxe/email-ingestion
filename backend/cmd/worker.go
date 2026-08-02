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
	"github.com/ajaxe/email-ingestion/internal/worker/handlers"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var streamNames []string

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run the stream processor workers",
	Long:  "Run the stream processor workers that continuously poll Redis streams for new jobs. The --streams flag determines which streams this binary will consume.",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWorker(cmd.Context()) },
}

func runWorker(ctx context.Context) error {
	cfg, err := config.LoadConfig(".")
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

		switch s {
		case "email":
			emailProcessor := service.NewEmailProcessor(queries, storageService)
			emailHandler := handlers.NewEmailIngestionHandler(emailProcessor)
			streamName := util.EmailStreamName(cfg.Environment)
			consumer := worker.NewStreamConsumer(rdsManager, streamName, "email_worker_group", emailHandler)
			consumers = append(consumers, consumer)

		case "webhook":
			// TODO: Add webhook processor and handler when implemented
			slog.Warn("webhook stream logic is not fully implemented yet")
			// webhookProcessor := service.NewWebhookProcessor(queries)
			// webhookHandler := handlers.NewWebhookDeliveryHandler(webhookProcessor)
			// streamName := util.WebhookStreamName(cfg.Environment)
			// consumer := worker.NewStreamConsumer(rdsManager, streamName, "webhook_worker_group", webhookHandler)
			// consumers = append(consumers, consumer)

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
