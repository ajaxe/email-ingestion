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
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the API server",
	Long:  "Run the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		// The PersistentPreRunE in root.go has already loaded config and setup logger.
		// However, cobra RunE allows us to fetch it again or we can just access it if global.
		// For safety, let's load it here since we don't have a global config variable.
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		e := router.New(cfg)

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

		// Create a deadline to wait for, using context.Background()
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
