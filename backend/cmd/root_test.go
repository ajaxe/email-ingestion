package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ajaxe/email-ingestion/pkg/config"
)

func TestRootCmd_ConfigFlag(t *testing.T) {
	config.ResetConfig()

	tmpDir := t.TempDir()
	customConfigPath := filepath.Join(tmpDir, "app_config.yaml")
	yamlContent := `
environment: flag-test-env
log_level: info
server:
  port: 7070
`
	if err := os.WriteFile(customConfigPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	rootCmd.SetArgs([]string{"--config", customConfigPath, "--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute failed: %v", err)
	}

	if cfgFile != customConfigPath {
		t.Errorf("expected cfgFile to be '%s', got '%s'", customConfigPath, cfgFile)
	}
	if path := getConfigPath(); path != customConfigPath {
		t.Errorf("expected getConfigPath() to return '%s', got '%s'", customConfigPath, path)
	}
}
