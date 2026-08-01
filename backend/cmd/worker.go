package cmd

import (
	"context"
	"log/slog"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/startup"
	"github.com/ajaxe/email-ingestion/internal/storage"
	"github.com/ajaxe/email-ingestion/internal/worker"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Run the email processor worker",
	Long:  "Run the email processor worker that continuously  polls the database for new email that are received. Multiple workers can be started to process emails in parallel.",
	RunE:  func(cmd *cobra.Command, args []string) error { return runWorker(cmd.Context()) },
}

func runWorker(ctx context.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return err
	}

	slog.Info("starting worker initialization")

	dbPool := startup.NewDbPool(cfg)
	defer dbPool.Close()

	queries := public.New(dbPool)

	rdsManager := redis.NewManager(cfg)
	defer rdsManager.Close()

	storageService := storage.NewStorageService(&cfg.Storage)

	w := worker.New(cfg, queries, rdsManager, storageService)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		w.Start(ctx)
	}()

	<-ctx.Done()

	slog.InfoContext(ctx, "graceful shutdown of worker requested")

	return nil
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
