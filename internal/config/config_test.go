package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"watchtower/internal/config"
)

func TestConfig_LoadAndValidate(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config.yml")

	yamlContent := `
workers: 8
targets:
  - name: test-site
    type: http
    address: https://example.com
    interval: 20s
    timeout: 4s
    retries: 2
`
	if err := os.WriteFile(cfgFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Workers != 8 {
		t.Errorf("expected 8 workers, got %d", cfg.Workers)
	}
	if len(cfg.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(cfg.Targets))
	}
	if cfg.Targets[0].Interval != 20*time.Second {
		t.Errorf("expected 20s interval, got %v", cfg.Targets[0].Interval)
	}
}

func TestConfig_EnvExpansion(t *testing.T) {
	t.Setenv("TEST_WATCHTOWER_ADDR", "https://expanded.com")

	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "config_env.yml")

	yamlContent := `
targets:
  - name: env-site
    type: http
    address: ${TEST_WATCHTOWER_ADDR}
`
	if err := os.WriteFile(cfgFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Targets[0].Address != "https://expanded.com" {
		t.Errorf("expected expanded address, got %s", cfg.Targets[0].Address)
	}
}

func TestConfig_ValidationErrors(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "invalid.yml")

	yamlContent := `
targets:
  - name: bad-type
    type: ftp
    address: ftp://example.com
`
	if err := os.WriteFile(cfgFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := config.Load(cfgFile)
	if err == nil {
		t.Fatal("expected validation error for invalid target type, got nil")
	}
}
