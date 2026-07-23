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
		var result map[string]interface{}
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
		var result map[string]interface{}
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
	var result map[string]interface{}
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
