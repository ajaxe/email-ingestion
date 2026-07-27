package cmd

import "github.com/spf13/cobra"

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Run the API server",
	Long:  "Run the API server",
	RunE:  func(cmd *cobra.Command, args []string) error { return nil },
}

func init() {
	rootCmd.AddCommand(apiCmd)
}
