package cmd

import (
	"fmt"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database"
	"github.com/spf13/cobra"
)

var versionDir string

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database schema migrations",
	Long:  "Manage PostgreSQL database schema migrations using Pressly Goose driver.",
	Example: `  email-ingestion migrate up
  email-ingestion migrate down
  email-ingestion migrate status --version-dir v0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		command := args[0]
		cmdArgs := args[1:]
		return database.RunMigrationsForVersion(cmd.Context(), cfg.Database.DSN, versionDir, command, cmdArgs...)
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Migrate database to the most recent version available",
	Long:  "Apply all pending database migrations up to the most recent schema version using Pressly Goose.",
	Example: `  email-ingestion migrate up
  email-ingestion migrate up --version-dir v0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return database.RunMigrationsForVersion(cmd.Context(), cfg.Database.DSN, versionDir, "up", args...)
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Roll back the most recent migration",
	Long:  "Roll back the single most recent database migration version.",
	Example: `  email-ingestion migrate down`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return database.RunMigrationsForVersion(cmd.Context(), cfg.Database.DSN, versionDir, "down", args...)
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display database migration status",
	Long:  "Print the current status of database migrations indicating which migration files have been applied.",
	Example: `  email-ingestion migrate status`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return database.RunMigrationsForVersion(cmd.Context(), cfg.Database.DSN, versionDir, "status", args...)
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current database migration version",
	Long:  "Print the current migration version number applied to the target database.",
	Example: `  email-ingestion migrate version`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return database.RunMigrationsForVersion(cmd.Context(), cfg.Database.DSN, versionDir, "version", args...)
	},
}

var migrateSkipCmd = &cobra.Command{
	Use:   "skip",
	Short: "Skip next database migration without running queries",
	Long:  "Mark the next database migration version as applied in the database tracking table without executing any underlying SQL migration scripts.",
	Example: `  email-ingestion migrate skip`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return database.RunMigrationsForVersion(cmd.Context(), cfg.Database.DSN, versionDir, "skip", args...)
	},
}

func init() {
	migrateCmd.PersistentFlags().StringVarP(&versionDir, "version-dir", "v", "v0", "Schema version directory (default: v0)")
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateStatusCmd, migrateVersionCmd, migrateSkipCmd)
	rootCmd.AddCommand(migrateCmd)
}
