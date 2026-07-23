package limiter

import (
	"sync"
	"testing"
	"time"
)

func TestLeakingBucket_BasicAllow(t *testing.T) {
	lb := NewLeakingBucket(5, 1)
	now := time.Now()

	// First 5 requests should be allowed
	for i := 0; i < 5; i++ {
		if !lb.AllowAt(now) {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
	}
}

func TestLeakingBucket_QueueFullRejection(t *testing.T) {
	lb := NewLeakingBucket(3, 1)
	now := time.Now()

	// Fill the queue
	for i := 0; i < 3; i++ {
		if !lb.AllowAt(now) {
			t.Errorf("request %d: expected allowed, got denied", i)
		}
	}

	// 4th request should be denied (queue full)
	if lb.AllowAt(now) {
		t.Error("expected denied (queue full), got allowed")
	}
}

func TestLeakingBucket_LeakOverTime(t *testing.T) {
	lb := NewLeakingBucket(3, 1)
	now := time.Now()

	// Fill the queue
	for i := 0; i < 3; i++ {
		lb.AllowAt(now)
	}

	// After 1 second, 1 item should have leaked
	afterLeak := now.Add(time.Second)
	if !lb.AllowAt(afterLeak) {
		t.Error("expected allowed after leak, got denied")
	}

	// Queue should be full again (3 items)
	if lb.AllowAt(afterLeak) {
		t.Error("expected denied (queue full again), got allowed")
	}
}

func TestLeakingBucket_ConcurrentAccess(t *testing.T) {
	lb := NewLeakingBucket(50, 50)
	now := time.Now()

	var wg sync.WaitGroup
	allowed := make([]bool, 200)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			allowed[idx] = lb.AllowAt(now)
		}(i)
	}

	wg.Wait()

	allowedCount := 0
	for _, a := range allowed {
		if a {
			allowedCount++
		}
	}

	if allowedCount != 50 {
		t.Errorf("expected 50 allowed, got %d", allowedCount)
	}
}

func TestLeakingBucket_ZeroQueue(t *testing.T) {
	lb := NewLeakingBucket(0, 1)
	if lb.Allow() {
		t.Error("expected denied with zero queue size")
	}
}

func TestLeakingBucket_NegativeLeakRate(t *testing.T) {
	lb := NewLeakingBucket(5, -1)
	if !lb.Allow() {
		t.Error("expected allowed with negative leak rate (queue not full)")
	}
}

func TestLeakingBucket_Name(t *testing.T) {
	lb := NewLeakingBucket(5, 1)
	if lb.Name() != "leaking_bucket" {
		t.Errorf("expected 'leaking_bucket', got '%s'", lb.Name())
	}
}

func TestLeakingBucket_Stats(t *testing.T) {
	lb := NewLeakingBucket(5, 1)
	stats := lb.Stats()
	if stats["queue_size"] != 5 {
		t.Errorf("expected queue_size 5, got %v", stats["queue_size"])
	}
	if stats["limiter_type"] != "leaking_bucket" {
		t.Errorf("expected limiter_type 'leaking_bucket', got %v", stats["limiter_type"])
	}
}
