package startup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
)

// RunStartupChecks performs startup checks, including running automatic database migrations up to date.
func RunStartupChecks(ctx context.Context, cfg *config.AppConfig) error {
	slog.Info("running database migration startup check...")
	if err := database.RunMigrations(ctx, cfg.Database.DSN, "up"); err != nil {
		return fmt.Errorf("failed automatic startup database migrations: %w", err)
	}
	slog.Info("database migration startup check completed successfully")
	return nil
}
