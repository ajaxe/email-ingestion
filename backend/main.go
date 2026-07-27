package main

import (
	"os"

	"github.com/ajaxe/email-ingestion/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
