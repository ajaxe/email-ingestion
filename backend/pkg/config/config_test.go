package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_CustomFilePath(t *testing.T) {
	ResetConfig()

	tmpDir := t.TempDir()
	customConfigPath := filepath.Join(tmpDir, "custom.yaml")
	yamlContent := `
environment: test-env
log_level: debug
server:
  port: 9090
`
	err := os.WriteFile(customConfigPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := LoadConfig(customConfigPath)
	if err != nil {
		t.Fatalf("LoadConfig failed for custom file path: %v", err)
	}

	if cfg.Environment != "test-env" {
		t.Errorf("expected environment 'test-env', got '%s'", cfg.Environment)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level 'debug', got '%s'", cfg.LogLevel)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected server.port 9090, got %d", cfg.Server.Port)
	}
}

func TestLoadConfig_DirectoryPath(t *testing.T) {
	ResetConfig()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `
environment: dir-env
log_level: info
server:
  port: 8081
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed for directory path: %v", err)
	}

	if cfg.Environment != "dir-env" {
		t.Errorf("expected environment 'dir-env', got '%s'", cfg.Environment)
	}
	if cfg.Server.Port != 8081 {
		t.Errorf("expected server.port 8081, got %d", cfg.Server.Port)
	}
}
