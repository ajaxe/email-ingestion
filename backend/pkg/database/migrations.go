package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*/*.sql
var embedMigrations embed.FS

// GetMigrationsFS returns an fs.FS sub-filesystem scoped to a specific version folder (default "v0").
func GetMigrationsFS(version string) (fs.FS, error) {
	if version == "" {
		version = "v0"
	}
	dir := fmt.Sprintf("migrations/%s", version)
	subFS, err := fs.Sub(embedMigrations, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-filesystem for %s: %w", dir, err)
	}
	return subFS, nil
}

// RunMigrations runs goose migration commands for default schema version "v0".
func RunMigrations(ctx context.Context, dsn string, command string, args ...string) error {
	return RunMigrationsForVersion(ctx, dsn, "v0", command, args...)
}

// RunMigrationsForVersion runs goose migration commands for a specified schema version folder.
func RunMigrationsForVersion(ctx context.Context, dsn string, version string, command string, args ...string) error {
	subFS, err := GetMigrationsFS(version)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection for migrations: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database for migrations: %w", err)
	}

	goose.SetBaseFS(subFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.RunContext(ctx, command, db, ".", args...); err != nil {
		return fmt.Errorf("goose %s failed: %w", command, err)
	}

	return nil
}
