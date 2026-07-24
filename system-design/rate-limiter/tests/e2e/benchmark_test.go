package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/bytebytego-learning/rate-limiter/internal/config"
	"github.com/bytebytego-learning/rate-limiter/internal/limiter"
)

// BenchmarkTokenBucket_Allow benchmarks the TokenBucket.Allow() method.
func BenchmarkTokenBucket_Allow(b *testing.B) {
	rl, _ := limiter.NewRateLimiter(&config.Config{
		LimiterType: "token_bucket",
		TokenBucket: &config.TokenBucketConfig{
			Capacity:       1000,
			RefillRate:     1000,
			RefillInterval: time.Second,
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

// BenchmarkLeakingBucket_Allow benchmarks the LeakingBucket.Allow() method.
func BenchmarkLeakingBucket_Allow(b *testing.B) {
	rl, _ := limiter.NewRateLimiter(&config.Config{
		LimiterType: "leaking_bucket",
		LeakingBucket: &config.LeakingBucketConfig{
			QueueSize: 1000,
			LeakRate:  1000,
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

// BenchmarkBBR_Allow benchmarks the BBR.Allow() method.
func BenchmarkBBR_Allow(b *testing.B) {
	rl, _ := limiter.NewRateLimiter(&config.Config{
		LimiterType: "bbr",
		BBR: &config.BBRConfig{
			Gain:       2.0,
			WindowSize: 10,
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

// BenchmarkTokenBucket_Refill benchmarks refill behavior under load.
func BenchmarkTokenBucket_Refill(b *testing.B) {
	rl, _ := limiter.NewRateLimiter(&config.Config{
		LimiterType: "token_bucket",
		TokenBucket: &config.TokenBucketConfig{
			Capacity:       100,
			RefillRate:     10,
			RefillInterval: 10 * time.Millisecond,
		},
	})
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.AllowAt(now.Add(time.Duration(i) * time.Millisecond))
	}
}

// BenchmarkLeakingBucket_Leak benchmarks leak behavior under load.
func BenchmarkLeakingBucket_Leak(b *testing.B) {
	rl, _ := limiter.NewRateLimiter(&config.Config{
		LimiterType: "leaking_bucket",
		LeakingBucket: &config.LeakingBucketConfig{
			QueueSize: 100,
			LeakRate:  100,
		},
	})
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.AllowAt(now.Add(time.Duration(i) * time.Millisecond))
	}
}

// BenchmarkBBR_Adaptation benchmarks BBR adaptation under load.
func BenchmarkBBR_Adaptation(b *testing.B) {
	rl, _ := limiter.NewRateLimiter(&config.Config{
		LimiterType: "bbr",
		BBR: &config.BBRConfig{
			Gain:       2.0,
			WindowSize: 10,
		},
	})
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.AllowAt(now.Add(time.Duration(i) * time.Millisecond))
	}
}

// BenchmarkComparison runs all 3 limiters side-by-side and prints a comparison table.
func BenchmarkComparison(b *testing.B) {
	benchmarks := []struct {
		name string
		rl   limiter.RateLimiter
	}{
		{
			name: "token_bucket",
			rl: func() limiter.RateLimiter {
				rl, _ := limiter.NewRateLimiter(&config.Config{
					LimiterType: "token_bucket",
					TokenBucket: &config.TokenBucketConfig{
						Capacity:       1000,
						RefillRate:     1000,
						RefillInterval: time.Second,
					},
				})
				return rl
			}(),
		},
		{
			name: "leaking_bucket",
			rl: func() limiter.RateLimiter {
				rl, _ := limiter.NewRateLimiter(&config.Config{
					LimiterType: "leaking_bucket",
					LeakingBucket: &config.LeakingBucketConfig{
						QueueSize: 1000,
						LeakRate:  1000,
					},
				})
				return rl
			}(),
		},
		{
			name: "bbr",
			rl: func() limiter.RateLimiter {
				rl, _ := limiter.NewRateLimiter(&config.Config{
					LimiterType: "bbr",
					BBR: &config.BBRConfig{
						Gain:       2.0,
						WindowSize: 10,
					},
				})
				return rl
			}(),
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			rl := bm.rl
			for i := 0; i < b.N; i++ {
				rl.Allow()
			}
		})
	}
}

// BenchmarkResult is a helper to run a sub-benchmark and capture results.
type BenchmarkResult struct {
	Name   string
	N      int64
	NsOp   int64
	Alloc  int64
	Allocs int64
}

// RunBenchmark runs a single benchmark function and returns structured results.
func RunBenchmark(name string, fn func(b *testing.B)) BenchmarkResult {
	result := testing.Benchmark(fn)
	return BenchmarkResult{
		Name:   name,
		N:      int64(result.N),
		NsOp:   result.NsPerOp(),
		Alloc:  result.AllocedBytesPerOp(),
		Allocs: int64(result.AllocsPerOp()),
	}
}

// PrintBenchmarkTable prints a formatted comparison table from benchmark results.
func PrintBenchmarkTable(results []BenchmarkResult) {
	fmt.Printf("\n%-20s %-15s %-15s %-15s %-15s\n",
		"Limiter", "Iterations", "ns/op", "B/op", "allocs/op")
	fmt.Printf("%-20s %-15s %-15s %-15s %-15s\n",
		"-------", "----------", "-----", "-----", "----------")
	for _, r := range results {
		fmt.Printf("%-20s %-15d %-15d %-15d %-15d\n",
			r.Name, r.N, r.NsOp, r.Alloc, r.Allocs)
	}
	fmt.Println()
}
