package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/model"
	"github.com/ajaxe/email-ingestion/internal/util"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/spf13/cobra"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Run background polling tasks",
	RunE:  func(cmd *cobra.Command, args []string) error { return runCron(cmd.Context()) },
}

func runCron(ctx context.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return err
	}

	dbPool := database.NewDbPool(cfg)
	defer dbPool.Close()

	queries := public.New(dbPool)

	rdsManager := redis.NewManager(cfg)
	defer rdsManager.Close()

	streamName := util.WebhookStreamName(cfg.Environment)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	slog.InfoContext(ctx, "starting cron job runner for webhook retries")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			jobs, err := queries.GetPendingWebhookJobs(ctx, 100)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get pending webhook jobs", "error", err)
				continue
			}

			if len(jobs) > 0 {
				slog.InfoContext(ctx, "found pending webhook jobs", "count", len(jobs))
			}

			for _, job := range jobs {
				// Mark as processing so it isn't picked up again immediately
				err = queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
					ID:             job.ID,
					Status:         public.WebhookStatusPROCESSING,
					RetryCount:     job.RetryCount,
					NextDeliveryAt: job.NextDeliveryAt,
				})
				if err != nil {
					slog.ErrorContext(ctx, "failed to update job status to processing", "job_id", job.ID, "error", err)
					continue
				}

				j, err := util.JSON(&model.WebhookDeliveryPayload{
					ApplicationID:   job.ApplicationID.String(),
					IngestedEmailID: job.IngestedEmailID.String(),
				})

				if err != nil {
					slog.ErrorContext(ctx, "failed to serialize webhook payload", "job_id", job.ID, "error", err)
					continue
				}

				err = rdsManager.Stream.Publish(ctx, streamName, j)
				if err != nil {
					slog.ErrorContext(ctx, "failed to publish job to stream", "job_id", job.ID, "error", err)
					// We might need a mechanism to revert it, but for now we rely on the consumer auto-claim
					// or subsequent polling if it was lost. But since it's PROCESSING, it might be stuck.
					// Ideally we revert to PENDING.
					_ = queries.UpdateWebhookJobStatus(ctx, public.UpdateWebhookJobStatusParams{
						ID:             job.ID,
						Status:         public.WebhookStatusPENDING,
						RetryCount:     job.RetryCount,
						NextDeliveryAt: job.NextDeliveryAt,
					})
				} else {
					slog.InfoContext(ctx, "enqueued pending job to stream", "job_id", job.ID)
				}
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(cronCmd)
}
