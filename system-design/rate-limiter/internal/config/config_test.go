package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
limiter_type: "token_bucket"
server:
  port: 9090
token_bucket:
  capacity: 20
  refill_rate: 2.0
  refill_interval: 500ms
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.LimiterType != "token_bucket" {
		t.Errorf("expected token_bucket, got %s", cfg.LimiterType)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.TokenBucket.Capacity != 20 {
		t.Errorf("expected capacity 20, got %d", cfg.TokenBucket.Capacity)
	}
	if cfg.TokenBucket.RefillRate != 2.0 {
		t.Errorf("expected refill_rate 2.0, got %f", cfg.TokenBucket.RefillRate)
	}
	if cfg.TokenBucket.RefillInterval != 500*time.Millisecond {
		t.Errorf("expected refill_interval 500ms, got %v", cfg.TokenBucket.RefillInterval)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.yaml")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}

	if cfg.LimiterType != "token_bucket" {
		t.Errorf("expected default token_bucket, got %s", cfg.LimiterType)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
limiter_type: "leaking_bucket"
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.LimiterType != "leaking_bucket" {
		t.Errorf("expected leaking_bucket, got %s", cfg.LimiterType)
	}
	// Defaults should apply
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.TokenBucket == nil {
		t.Error("expected token_bucket defaults to be set")
	}
	if cfg.LeakingBucket == nil {
		t.Error("expected leaking_bucket defaults to be set")
	}
	if cfg.BBR == nil {
		t.Error("expected bbr defaults to be set")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
limiter_type: "token_bucket"
  invalid yaml: [}
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.LimiterType != "token_bucket" {
		t.Errorf("expected token_bucket, got %s", cfg.LimiterType)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.TokenBucket.Capacity != 10 {
		t.Errorf("expected capacity 10, got %d", cfg.TokenBucket.Capacity)
	}
	if cfg.LeakingBucket.QueueSize != 10 {
		t.Errorf("expected queue_size 10, got %d", cfg.LeakingBucket.QueueSize)
	}
	if cfg.BBR.Gain != 2.0 {
		t.Errorf("expected gain 2.0, got %f", cfg.BBR.Gain)
	}
}
