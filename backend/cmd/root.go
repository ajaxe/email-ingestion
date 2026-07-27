package cmd

import (
	"fmt"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "single binary for multi-service application",
	Long:  "single binary for multi-service application",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running root command...")

		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		config.SetupLogger(cfg)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
