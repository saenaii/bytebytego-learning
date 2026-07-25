package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration loaded from config.yaml.
type Config struct {
	LimiterType   string               `yaml:"limiter_type"`
	Server        ServerConfig         `yaml:"server"`
	TokenBucket   *TokenBucketConfig   `yaml:"token_bucket,omitempty"`
	LeakingBucket *LeakingBucketConfig `yaml:"leaking_bucket,omitempty"`
	BBR           *BBRConfig           `yaml:"bbr,omitempty"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int `yaml:"port"`
}

// TokenBucketConfig holds token bucket algorithm settings.
type TokenBucketConfig struct {
	Capacity       int           `yaml:"capacity"`
	RefillRate     float64       `yaml:"refill_rate"`
	RefillInterval time.Duration `yaml:"refill_interval"`
}

// LeakingBucketConfig holds leaking bucket algorithm settings.
type LeakingBucketConfig struct {
	QueueSize int     `yaml:"queue_size"`
	LeakRate  float64 `yaml:"leak_rate"`
}

// BBRConfig holds BBR algorithm settings.
type BBRConfig struct {
	Gain       float64 `yaml:"gain"`
	WindowSize int     `yaml:"window_size"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		LimiterType: "token_bucket",
		Server: ServerConfig{
			Port: 8080,
		},
		TokenBucket: &TokenBucketConfig{
			Capacity:       10,
			RefillRate:     1.0,
			RefillInterval: time.Second,
		},
		LeakingBucket: &LeakingBucketConfig{
			QueueSize: 10,
			LeakRate:  1.0,
		},
		BBR: &BBRConfig{
			Gain:       2.0,
			WindowSize: 10,
		},
	}
}

// LoadConfig reads a YAML configuration file from the given path.
// If the file does not exist, it returns a default configuration.
// If the file exists but is invalid, it returns an error.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found, use defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse YAML into a raw map first to detect partial configs
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// If file is empty (no keys), return defaults
	if len(raw) == 0 {
		return cfg, nil
	}

	// Unmarshal into the config struct
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Apply defaults for missing fields
	cfg.applyDefaults()

	return cfg, nil
}

// applyDefaults ensures nil sub-configs are set to defaults.
func (c *Config) applyDefaults() {
	if c.TokenBucket == nil {
		c.TokenBucket = &TokenBucketConfig{
			Capacity:       10,
			RefillRate:     1.0,
			RefillInterval: time.Second,
		}
	}
	if c.LeakingBucket == nil {
		c.LeakingBucket = &LeakingBucketConfig{
			QueueSize: 10,
			LeakRate:  1.0,
		}
	}
	if c.BBR == nil {
		c.BBR = &BBRConfig{
			Gain:       2.0,
			WindowSize: 10,
		}
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
}
