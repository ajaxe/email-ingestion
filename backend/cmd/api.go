package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/ajaxe/email-ingestion/internal/api/router"
	"github.com/ajaxe/email-ingestion/pkg/config"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the API server",
	Long:  "Run the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// The PersistentPreRunE in root.go has already loaded config and setup logger.
		// However, cobra RunE allows us to fetch it again or we can just access it if global.
		// For safety, let's load it here since we don't have a global config variable.
		_, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		// TODO: Map to actual configuration
		edgeToken := "development-edge-token-123" 
		
		e := router.New(edgeToken)
		
		port := 8080 // Default port
		slog.Info("Starting API server", "port", port)
		return e.Start(fmt.Sprintf(":%d", port))
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
