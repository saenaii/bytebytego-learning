package limiter

import (
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_BasicAllow(t *testing.T) {
	tb := NewTokenBucket(5, 1, time.Second)
	now := time.Now()

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !tb.AllowAt(now) {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
	}

	// 6th request should be denied (empty bucket)
	if tb.AllowAt(now) {
		t.Error("expected denied, got allowed")
	}
}

func TestTokenBucket_BasicDeny(t *testing.T) {
	tb := NewTokenBucket(0, 1, time.Second)
	if tb.Allow() {
		t.Error("expected denied with zero capacity, got allowed")
	}
}

func TestTokenBucket_RefillOverTime(t *testing.T) {
	tb := NewTokenBucket(5, 1, time.Second)
	now := time.Now()

	// Drain the bucket
	for i := 0; i < 5; i++ {
		tb.AllowAt(now)
	}

	// After 1 second, 1 token should be refilled
	afterRefill := now.Add(time.Second)
	if !tb.AllowAt(afterRefill) {
		t.Error("expected allowed after refill, got denied")
	}

	// After another second, another token should be available
	afterSecondRefill := now.Add(2 * time.Second)
	if !tb.AllowAt(afterSecondRefill) {
		t.Error("expected allowed after second refill, got denied")
	}
}

func TestTokenBucket_BurstCapacity(t *testing.T) {
	tb := NewTokenBucket(10, 1, time.Second)
	now := time.Now()

	// Allow 10 requests in a burst
	for i := 0; i < 10; i++ {
		if !tb.AllowAt(now) {
			t.Errorf("burst request %d: expected allowed, got denied", i)
		}
	}

	// 11th should be denied
	if tb.AllowAt(now) {
		t.Error("expected denied after burst, got allowed")
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(100, 100, time.Second)
	now := time.Now()

	var wg sync.WaitGroup
	allowed := make([]bool, 200)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			allowed[idx] = tb.AllowAt(now)
		}(i)
	}

	wg.Wait()

	allowedCount := 0
	for _, a := range allowed {
		if a {
			allowedCount++
		}
	}

	if allowedCount != 100 {
		t.Errorf("expected 100 allowed, got %d", allowedCount)
	}
}

func TestTokenBucket_ZeroCapacity(t *testing.T) {
	tb := NewTokenBucket(0, 1, time.Second)
	if tb.Allow() {
		t.Error("expected denied with zero capacity")
	}
}

func TestTokenBucket_NegativeRefill(t *testing.T) {
	tb := NewTokenBucket(5, -1, time.Second)
	// Drain the bucket first (bucket starts full with capacity 5)
	for i := 0; i < 5; i++ {
		tb.Allow()
	}
	// Now with negative refill rate (clamped to 0), no tokens should be added
	if tb.Allow() {
		t.Error("expected denied with negative refill rate after draining bucket")
	}
}

func TestTokenBucket_Name(t *testing.T) {
	tb := NewTokenBucket(5, 1, time.Second)
	if tb.Name() != "token_bucket" {
		t.Errorf("expected 'token_bucket', got '%s'", tb.Name())
	}
}

func TestTokenBucket_Stats(t *testing.T) {
	tb := NewTokenBucket(5, 1, time.Second)
	stats := tb.Stats()
	if stats["capacity"] != float64(5) {
		t.Errorf("expected capacity 5, got %v", stats["capacity"])
	}
	if stats["limiter_type"] != "token_bucket" {
		t.Errorf("expected limiter_type 'token_bucket', got %v", stats["limiter_type"])
	}
}
