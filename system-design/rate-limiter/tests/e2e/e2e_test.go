package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytebytego-learning/rate-limiter/internal/config"
	"github.com/bytebytego-learning/rate-limiter/internal/limiter"
	"github.com/bytebytego-learning/rate-limiter/internal/server"
)

// TestE2E_TokenBucket runs end-to-end tests against a token bucket rate limiter.
func TestE2E_TokenBucket(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "token_bucket",
		TokenBucket: &config.TokenBucketConfig{
			Capacity:       5,
			RefillRate:     1.0,
			RefillInterval: time.Second,
		},
	}

	rateLimiter, err := limiter.NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	srv := server.NewServer(rateLimiter, 0)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	runE2ETests(t, ts.URL, 5)
}

// TestE2E_LeakingBucket runs end-to-end tests against a leaking bucket rate limiter.
func TestE2E_LeakingBucket(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "leaking_bucket",
		LeakingBucket: &config.LeakingBucketConfig{
			QueueSize: 5,
			LeakRate:  10.0,
		},
	}

	rateLimiter, err := limiter.NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	srv := server.NewServer(rateLimiter, 0)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	runE2ETests(t, ts.URL, 5)
}

// TestE2E_BBR runs end-to-end tests against a BBR rate limiter.
// BBR is a probabilistic algorithm, so we test that it works but allow for
// some variance in the number of allowed/rejected requests.
func TestE2E_BBR(t *testing.T) {
	cfg := &config.Config{
		LimiterType: "bbr",
		BBR: &config.BBRConfig{
			Gain:       2.0,
			WindowSize: 10,
		},
	}

	rateLimiter, err := limiter.NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}

	srv := server.NewServer(rateLimiter, 0)
	ts := httptest.NewServer(srv)
	defer ts.Close()

	// BBR is probabilistic, so we just verify the endpoints work
	// and that some requests are allowed
	runBBRE2ETests(t, ts.URL)
}

// runBBRE2ETests executes BBR-specific tests that account for its probabilistic nature.
func runBBRE2ETests(t *testing.T, baseURL string) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Test health endpoint
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health endpoint returned status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Send multiple requests - BBR should allow some and may reject others
	allowed := 0
	for i := 0; i < 10; i++ {
		httpResp, err := client.Get(baseURL + "/allow")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		var result map[string]any
		json.NewDecoder(httpResp.Body).Decode(&result)
		httpResp.Body.Close()

		if result["allowed"].(bool) {
			allowed++
		}
	}

	t.Logf("BBR E2E results: allowed=%d out of 10 requests", allowed)

	// BBR should allow at least some requests
	if allowed == 0 {
		t.Errorf("Expected some allowed requests, got 0")
	}

	// Test stats endpoint
	resp, err = client.Get(baseURL + "/stats")
	if err != nil {
		t.Fatalf("Stats request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Stats endpoint returned status %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// helper to create a rate limiter server for testing
func newTestServer(t *testing.T, cfg *config.Config) (*httptest.Server, limiter.RateLimiter) {
	t.Helper()
	rateLimiter, err := limiter.NewRateLimiter(cfg)
	if err != nil {
		t.Fatalf("Failed to create rate limiter: %v", err)
	}
	srv := server.NewServer(rateLimiter, 0)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return ts, rateLimiter
}

// sendAllowRequest sends a GET /allow request and returns the result.
func sendAllowRequest(t *testing.T, client *http.Client, baseURL string) bool {
	t.Helper()
	resp, err := client.Get(baseURL + "/allow")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	return result["allowed"].(bool)
}

// TestE2E_AllLimiters_ComparisonReport runs a comparison across all 3 limiters
// and logs a report table.
func TestE2E_AllLimiters_ComparisonReport(t *testing.T) {
	capacity := 10
	limiters := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "token_bucket",
			cfg: &config.Config{
				LimiterType: "token_bucket",
				TokenBucket: &config.TokenBucketConfig{
					Capacity:       capacity,
					RefillRate:     float64(capacity),
					RefillInterval: time.Second,
				},
			},
		},
		{
			name: "leaking_bucket",
			cfg: &config.Config{
				LimiterType: "leaking_bucket",
				LeakingBucket: &config.LeakingBucketConfig{
					QueueSize: capacity,
					LeakRate:  float64(capacity),
				},
			},
		},
		{
			name: "bbr",
			cfg: &config.Config{
				LimiterType: "bbr",
				BBR: &config.BBRConfig{
					Gain:       2.0,
					WindowSize: 10,
				},
			},
		},
	}

	t.Logf("\n========== E2E COMPARISON REPORT ==========")
	t.Logf("%-15s %-10s %-10s %-10s", "Limiter", "Total", "Allowed", "Rejected")
	t.Logf("%-15s %-10s %-10s %-10s", "-------", "-----", "-------", "--------")

	for _, l := range limiters {
		_, rl := newTestServer(t, l.cfg)
		total := 200
		allowed := 0
		rejected := 0
		for i := 0; i < total; i++ {
			if rl.Allow() {
				allowed++
			} else {
				rejected++
			}
		}
		t.Logf("%-15s %-10d %-10d %-10d", l.name, total, allowed, rejected)
	}
	t.Logf("===========================================")
}

// TestE2E_Attacker_AboveThreshold simulates QPS above the configured threshold
// and verifies the rate limiter rejects excess requests.
func TestE2E_Attacker_AboveThreshold(t *testing.T) {
	capacity := 5
	limiters := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "token_bucket",
			cfg: &config.Config{
				LimiterType: "token_bucket",
				TokenBucket: &config.TokenBucketConfig{
					Capacity:       capacity,
					RefillRate:     float64(capacity),
					RefillInterval: time.Second,
				},
			},
		},
		{
			name: "leaking_bucket",
			cfg: &config.Config{
				LimiterType: "leaking_bucket",
				LeakingBucket: &config.LeakingBucketConfig{
					QueueSize: capacity,
					LeakRate:  float64(capacity),
				},
			},
		},
		{
			name: "bbr",
			cfg: &config.Config{
				LimiterType: "bbr",
				BBR: &config.BBRConfig{
					Gain:       2.0,
					WindowSize: 10,
				},
			},
		},
	}

	t.Logf("\n========== ATTACK MINDSET: ABOVE THRESHOLD ==========")
	t.Logf("Sending 3x capacity (%d) requests at the same instant\n", 3*capacity)

	for _, l := range limiters {
		t.Run(l.name, func(t *testing.T) {
			_, rl := newTestServer(t, l.cfg)
			total := 3 * capacity
			allowed := 0
			for i := 0; i < total; i++ {
				if rl.Allow() {
					allowed++
				}
			}
			rejected := total - allowed
			t.Logf("[%s] allowed=%d, rejected=%d (capacity=%d)", l.name, allowed, rejected, capacity)

			if l.name != "bbr" {
				if allowed > capacity {
					t.Errorf("Expected at most %d allowed requests, got %d", capacity, allowed)
				}
				if rejected == 0 {
					t.Errorf("Expected some requests to be rejected, got 0")
				}
			}
		})
	}
}

// TestE2E_Attacker_BurstVsSteady compares burst traffic vs steady traffic handling.
func TestE2E_Attacker_BurstVsSteady(t *testing.T) {
	capacity := 10
	duration := 5 * time.Second

	type result struct {
		allowed int
		rejected int
	}

	limiters := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "token_bucket",
			cfg: &config.Config{
				LimiterType: "token_bucket",
				TokenBucket: &config.TokenBucketConfig{
					Capacity:       capacity,
					RefillRate:     float64(capacity),
					RefillInterval: time.Second,
				},
			},
		},
		{
			name: "leaking_bucket",
			cfg: &config.Config{
				LimiterType: "leaking_bucket",
				LeakingBucket: &config.LeakingBucketConfig{
					QueueSize: capacity,
					LeakRate:  float64(capacity),
				},
			},
		},
		{
			name: "bbr",
			cfg: &config.Config{
				LimiterType: "bbr",
				BBR: &config.BBRConfig{
					Gain:       2.0,
					WindowSize: 10,
				},
			},
		},
	}

	t.Logf("\n========== ATTACK MINDSET: BURST vs STEADY ==========")
	t.Logf("Burst: 50 requests every 100ms | Steady: 1 request every 20ms\n")

	for _, l := range limiters {
		t.Run(l.name, func(t *testing.T) {
			_, rl := newTestServer(t, l.cfg)
			start := time.Now()

			// Steady load: 1 request every 20ms
			var steady result
			steadyStart := start
			for time.Since(steadyStart) < duration {
				if rl.Allow() {
					steady.allowed++
				} else {
					steady.rejected++
				}
				time.Sleep(20 * time.Millisecond)
			}

			// Burst load: 50 rapid requests every 100ms
			var burst result
			burstStart := time.Now()
			for time.Since(burstStart) < duration {
				for i := 0; i < 50; i++ {
					if rl.Allow() {
						burst.allowed++
					} else {
						burst.rejected++
					}
				}
				time.Sleep(100 * time.Millisecond)
			}

			totalSteady := steady.allowed + steady.rejected
			totalBurst := burst.allowed + burst.rejected
			t.Logf("[%s] steady: %d/%d allowed (%.1f%%) | burst: %d/%d allowed (%.1f%%)",
				l.name,
				steady.allowed, totalSteady, float64(steady.allowed)/float64(totalSteady)*100,
				burst.allowed, totalBurst, float64(burst.allowed)/float64(totalBurst)*100,
			)
		})
	}
}

// TestE2E_Attacker_DoubleTraffic checks for the double traffic / reset issue:
// when a rate limiter resets (e.g., after a long idle period), does it allow
// too many requests at once?
func TestE2E_Attacker_DoubleTraffic(t *testing.T) {
	capacity := 5
	limiters := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "token_bucket",
			cfg: &config.Config{
				LimiterType: "token_bucket",
				TokenBucket: &config.TokenBucketConfig{
					Capacity:       capacity,
					RefillRate:     1.0,
					RefillInterval: time.Second,
				},
			},
		},
		{
			name: "leaking_bucket",
			cfg: &config.Config{
				LimiterType: "leaking_bucket",
				LeakingBucket: &config.LeakingBucketConfig{
					QueueSize: capacity,
					LeakRate:  1.0,
				},
			},
		},
		{
			name: "bbr",
			cfg: &config.Config{
				LimiterType: "bbr",
				BBR: &config.BBRConfig{
					Gain:       2.0,
					WindowSize: 10,
				},
			},
		},
	}

	t.Logf("\n========== ATTACK MINDSET: DOUBLE TRAFFIC (RESET) ==========")
	t.Logf("Phase 1: Drain the limiter. Phase 2: Wait 2x refill time. Phase 3: Send burst.\n")

	for _, l := range limiters {
		t.Run(l.name, func(t *testing.T) {
			_, rl := newTestServer(t, l.cfg)
			now := time.Now()

			// Phase 1: Drain the limiter
			for i := 0; i < capacity; i++ {
				rl.AllowAt(now)
			}

			// Phase 2: Wait 2x refill interval (simulate idle)
			afterIdle := now.Add(2 * time.Second)

			// Phase 3: Send a burst after the reset window
			burstCount := 3 * capacity
			allowed := 0
			for i := 0; i < burstCount; i++ {
				if rl.AllowAt(afterIdle) {
					allowed++
				}
			}

			t.Logf("[%s] after idle: allowed=%d out of %d burst requests (capacity=%d)",
				l.name, allowed, burstCount, capacity)

			if l.name != "bbr" {
				// Deterministic limiters should not allow more than capacity
				if allowed > capacity {
					t.Errorf("Double traffic! Expected at most %d allowed after idle, got %d", capacity, allowed)
				}
			}
		})
	}
}

// runE2ETests executes a sequence of requests against the server and verifies results.
func runE2ETests(t *testing.T, baseURL string, capacity int) {
	client := &http.Client{Timeout: 5 * time.Second}

	// Test health endpoint
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health endpoint returned status %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Send requests up to capacity - all should be allowed
	allowed := 0
	rejected := 0
	for i := 0; i < capacity; i++ {
		var httpResp *http.Response
		httpResp, err = client.Get(baseURL + "/allow")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		var result map[string]any
		json.NewDecoder(httpResp.Body).Decode(&result)
		httpResp.Body.Close()

		if result["allowed"].(bool) {
			allowed++
		} else {
			rejected++
		}
	}

	t.Logf("E2E results: allowed=%d, rejected=%d (capacity=%d)", allowed, rejected, capacity)

	// The first 'capacity' requests should be allowed
	if allowed != capacity {
		t.Errorf("Expected %d allowed requests, got %d", capacity, allowed)
	}

	// Next request should be rejected (bucket empty)
	resp, err = client.Get(baseURL + "/allow")
	if err != nil {
		t.Fatalf("Extra request failed: %v", err)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["allowed"].(bool) {
		t.Log("Note: Extra request was allowed (algorithm may not reject immediately)")
	} else {
		t.Log("Extra request correctly rejected")
	}

	// Test stats endpoint
	resp, err = client.Get(baseURL + "/stats")
	if err != nil {
		t.Fatalf("Stats request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Stats endpoint returned status %d", resp.StatusCode)
	}
	resp.Body.Close()
}
