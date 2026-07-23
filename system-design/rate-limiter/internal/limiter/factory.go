package limiter

import (
	"fmt"
	"time"

	"github.com/bytebytego-learning/rate-limiter/internal/config"
)

// NewRateLimiter creates a RateLimiter based on the provided configuration.
// It returns the appropriate limiter implementation or an error if the type is unknown.
func NewRateLimiter(cfg *config.Config) (RateLimiter, error) {
	switch cfg.LimiterType {
	case "token_bucket":
		tbCfg := cfg.TokenBucket
		if tbCfg == nil {
			tbCfg = &config.TokenBucketConfig{
				Capacity:       10,
				RefillRate:     1.0,
				RefillInterval: time.Second,
			}
		}
		return NewTokenBucket(tbCfg.Capacity, tbCfg.RefillRate, tbCfg.RefillInterval), nil

	case "leaking_bucket":
		lbCfg := cfg.LeakingBucket
		if lbCfg == nil {
			lbCfg = &config.LeakingBucketConfig{
				QueueSize: 10,
				LeakRate:  1.0,
			}
		}
		return NewLeakingBucket(lbCfg.QueueSize, lbCfg.LeakRate), nil

	case "bbr":
		bbrCfg := cfg.BBR
		if bbrCfg == nil {
			bbrCfg = &config.BBRConfig{
				Gain:       2.0,
				WindowSize: 10,
			}
		}
		// Use sensible defaults for initial bandwidth and RTT
		return NewBBR(100, 10*time.Millisecond, bbrCfg.WindowSize, bbrCfg.Gain), nil

	default:
		return nil, fmt.Errorf("unknown limiter type: %s", cfg.LimiterType)
	}
}
