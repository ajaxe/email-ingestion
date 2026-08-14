package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ajaxe/email-ingestion/internal/api/router"
	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/startup"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the HTTP REST API server",
	Long: `Run the Echo HTTP REST API server. Handles tenant management, email ingestion endpoints from the 
SMTP proxy, S3 presigned URL generation, recipient validation, and webhook delivery endpoints.`,
	Example: `  email-ingestion api`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		if err := startup.RunStartupChecks(ctx, cfg); err != nil {
			return fmt.Errorf("failed to run startup checks: %w", err)
		}

		dbPool := database.NewDbPool(cfg)
		defer dbPool.Close()

		rdsManager := redis.NewManager(cfg)
		defer rdsManager.Close()

		e := router.New(cfg, &router.ApiInitOptions{
			Queries:      public.New(dbPool),
			RedisManager: rdsManager,
			DBPool:       dbPool,
		})

		go func() {
			portStr := fmt.Sprintf(":%d", cfg.Server.Port)
			slog.Info("starting server", "port", cfg.Server.Port)
			if err := e.Start(portStr); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("server startup failed", "error", err)
				os.Exit(1)
			}
		}()

		<-ctx.Done()
		slog.Info("received termination signal, shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(shutdownCtx); err != nil {
			slog.Error("server forced to shutdown", "error", err)
		} else {
			slog.Info("server exited properly")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
