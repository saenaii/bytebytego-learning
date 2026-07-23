package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bytebytego-learning/rate-limiter/internal/limiter"
)

func TestHealthEndpoint(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 1, time.Second)
	s := NewServer(tb, 8080)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp["status"])
	}
	if resp["limiter_type"] != "token_bucket" {
		t.Errorf("expected limiter_type 'token_bucket', got '%s'", resp["limiter_type"])
	}
}

func TestAllowEndpoint(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 1, time.Second)
	s := NewServer(tb, 8080)

	req := httptest.NewRequest(http.MethodGet, "/allow", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if resp["allowed"] != true {
		t.Errorf("expected allowed true, got %v", resp["allowed"])
	}
}

func TestAllowEndpoint_MethodNotAllowed(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 1, time.Second)
	s := NewServer(tb, 8080)

	req := httptest.NewRequest(http.MethodPost, "/allow", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestStatsEndpoint(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 1, time.Second)
	s := NewServer(tb, 8080)

	// Make a request first to update stats
	tb.Allow()

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	if resp["limiter_type"] != "token_bucket" {
		t.Errorf("expected limiter_type 'token_bucket', got '%v'", resp["limiter_type"])
	}
	allowed := resp["allowed"].(float64)
	if allowed != 1 {
		t.Errorf("expected allowed count 1, got %v", resp["allowed"])
	}
}

func TestRejectEndpoint(t *testing.T) {
	// Create a token bucket with capacity 1, refill 1 per second
	tb := limiter.NewTokenBucket(1, 1, time.Second)
	s := NewServer(tb, 8080)

	// First request should be allowed
	req1 := httptest.NewRequest(http.MethodGet, "/allow", nil)
	w1 := httptest.NewRecorder()
	s.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected first request to be allowed (200), got %d", w1.Code)
	}

	// Second request should be rejected (bucket empty, no refill yet)
	req2 := httptest.NewRequest(http.MethodGet, "/allow", nil)
	w2 := httptest.NewRecorder()
	s.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected second request to be rejected (429), got %d", w2.Code)
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"8080", 8080},
		{"", 9090},
		{"invalid", 9090},
		{"0", 9090},
		{"70000", 9090},
		{"-1", 9090},
	}

	for _, tt := range tests {
		result := ParsePort(tt.input, 9090)
		if result != tt.expected {
			t.Errorf("ParsePort(%q, 9090) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
