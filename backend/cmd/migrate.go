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
	Short: "Database migration commands",
	Long:  "Manage database schema migrations using Pressly Goose",
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
	Short: "Migrate the DB to the most recent version available",
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
	Short: "Dump the migration status for the database",
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
	Short: "Print the current version of the database",
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
	Short: "Set the database version to the next version without running migration queries",
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
