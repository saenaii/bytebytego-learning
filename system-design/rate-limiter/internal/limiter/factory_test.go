package limiter

import (
	"testing"
	"time"

	"github.com/bytebytego-learning/rate-limiter/internal/config"
)

func TestFactory_TokenBucket(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "token_bucket",
		TokenBucket: &config.TokenBucketConfig{
			Capacity:       5,
			RefillRate:     1.0,
			RefillInterval: time.Second,
		},
	}

	limiter, err := NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter.Name() != "token_bucket" {
		t.Errorf("expected 'token_bucket', got '%s'", limiter.Name())
	}
}

func TestFactory_LeakingBucket(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "leaking_bucket",
		LeakingBucket: &config.LeakingBucketConfig{
			QueueSize: 5,
			LeakRate:  1.0,
		},
	}

	limiter, err := NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter.Name() != "leaking_bucket" {
		t.Errorf("expected 'leaking_bucket', got '%s'", limiter.Name())
	}
}

func TestFactory_BBR(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "bbr",
		BBR: &config.BBRConfig{
			Gain:       2.0,
			WindowSize: 10,
		},
	}

	limiter, err := NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter.Name() != "bbr" {
		t.Errorf("expected 'bbr', got '%s'", limiter.Name())
	}
}

func TestFactory_UnknownType(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "unknown",
	}

	_, err := NewRateLimiter(cfg)
	if err == nil {
		t.Error("expected error for unknown limiter type")
	}
}

func TestFactory_FallbackDefaults(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "token_bucket",
		// TokenBucket config is nil, should use defaults
	}

	limiter, err := NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if limiter.Name() != "token_bucket" {
		t.Errorf("expected 'token_bucket', got '%s'", limiter.Name())
	}
}
