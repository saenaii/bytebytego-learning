package limiter

import (
	"sync"
	"testing"
	"time"
)

func TestBBR_InitialState(t *testing.T) {
	b := NewBBR(100, 10*time.Millisecond, 10, 2.0)
	if b.Name() != "bbr" {
		t.Errorf("expected 'bbr', got '%s'", b.Name())
	}
	stats := b.Stats()
	if stats["max_bandwidth"] != 100.0 {
		t.Errorf("expected max_bandwidth 100, got %v", stats["max_bandwidth"])
	}
	if stats["gain"] != 2.0 {
		t.Errorf("expected gain 2.0, got %v", stats["gain"])
	}
}

func TestBBR_Allow(t *testing.T) {
	// Override randomFloat to always return 0 (always allow if probability > 0)
	originalRandom := randomFloat
	randomFloat = func() float64 { return 0.0 }
	defer func() { randomFloat = originalRandom }()

	b := NewBBR(100, 10*time.Millisecond, 10, 2.0)

	// First request should be allowed (initial probability is 1.0)
	if !b.Allow() {
		t.Error("expected initial request to be allowed")
	}
}

func TestBBR_RateAdaptation(t *testing.T) {
	// Override randomFloat to always return 0 (always allow)
	originalRandom := randomFloat
	randomFloat = func() float64 { return 0.0 }
	defer func() { randomFloat = originalRandom }()

	b := NewBBR(100, 10*time.Millisecond, 10, 2.0)
	now := time.Now()

	// Send requests with increasing intervals
	for i := 0; i < 5; i++ {
		b.AllowAt(now.Add(time.Duration(i) * 100 * time.Millisecond))
	}

	stats := b.Stats()
	if stats["allowed"].(int64) == 0 {
		t.Error("expected some requests to be allowed")
	}
}

func TestBBR_ConcurrentAccess(t *testing.T) {
	originalRandom := randomFloat
	randomFloat = func() float64 { return 0.5 }
	defer func() { randomFloat = originalRandom }()

	b := NewBBR(100, 10*time.Millisecond, 10, 2.0)
	now := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			b.AllowAt(now.Add(time.Duration(idx) * time.Millisecond))
		}(i)
	}

	wg.Wait()
	// No race conditions should occur
	stats := b.Stats()
	t.Logf("BBR concurrent stats: allowed=%d, rejected=%d",
		stats["allowed"], stats["rejected"])
}

func TestBBR_ZeroGain(t *testing.T) {
	b := NewBBR(100, 10*time.Millisecond, 10, 0)
	if b.gain != 2.0 {
		t.Errorf("expected default gain 2.0, got %f", b.gain)
	}
}

func TestBBR_NegativeWindow(t *testing.T) {
	b := NewBBR(100, 10*time.Millisecond, -1, 2.0)
	if b.windowSize != 10 {
		t.Errorf("expected default windowSize 10, got %d", b.windowSize)
	}
}

func TestBBR_Stats(t *testing.T) {
	b := NewBBR(100, 10*time.Millisecond, 10, 2.0)
	stats := b.Stats()
	if stats["limiter_type"] != "bbr" {
		t.Errorf("expected limiter_type 'bbr', got %v", stats["limiter_type"])
	}
	if stats["window_size"] != 10 {
		t.Errorf("expected window_size 10, got %v", stats["window_size"])
	}
}
