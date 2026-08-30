package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ajaxe/email-ingestion/internal/api/router"
	"github.com/ajaxe/email-ingestion/internal/infra/redis"
	"github.com/ajaxe/email-ingestion/internal/startup"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/labstack/echo/v5"
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
		cfg, err := config.LoadConfig(getConfigPath())
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

		sc := echo.StartConfig{
			Address:         fmt.Sprintf(":%d", cfg.Server.Port),
			GracefulTimeout: 10 * time.Second,
		}

		slog.Info("starting server", "port", cfg.Server.Port)
		if err := sc.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server startup failed", "error", err)
			return err
		}

		slog.Info("server exited properly")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
