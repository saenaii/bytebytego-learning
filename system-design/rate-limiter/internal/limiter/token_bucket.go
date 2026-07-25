package limiter

import (
	"sync"
	"time"
)

// TokenBucket implements the token bucket rate limiting algorithm.
// A bucket holds up to capacity tokens. Each Allow() call consumes 1 token if available.
// Tokens are added back at a fixed refillRate per refillInterval.
type TokenBucket struct {
	mu             sync.Mutex
	capacity       float64       // Maximum number of tokens
	tokens         float64       // Current token count
	refillRate     float64       // Number of tokens added per refillInterval
	refillInterval time.Duration // Time between refills
	lastRefillTime time.Time     // Last time tokens were refilled
	allowed        int64         // Total allowed requests
	rejected       int64         // Total rejected requests
}

// NewTokenBucket creates a new TokenBucket rate limiter.
func NewTokenBucket(capacity int, refillRate float64, refillInterval time.Duration) *TokenBucket {
	if capacity < 0 {
		capacity = 0
	}
	if refillRate < 0 {
		refillRate = 0
	}
	if refillInterval <= 0 {
		refillInterval = time.Second
	}
	return &TokenBucket{
		capacity:       float64(capacity),
		tokens:         float64(capacity),
		refillRate:     refillRate,
		refillInterval: refillInterval,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request is allowed at the current time.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowAt(time.Now())
}

// AllowAt checks if a request would be allowed at the specified time.
func (tb *TokenBucket) AllowAt(t time.Time) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill(t)

	if tb.tokens >= 1 {
		tb.tokens--
		tb.allowed++
		return true
	}
	tb.rejected++
	return false
}

// Name returns the name of the rate limiter algorithm.
func (tb *TokenBucket) Name() string {
	return "token_bucket"
}

// Stats returns current statistics about the rate limiter state.
func (tb *TokenBucket) Stats() map[string]any {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	return map[string]any{
		"limiter_type":    tb.Name(),
		"tokens":          tb.tokens,
		"capacity":        tb.capacity,
		"refill_rate":     tb.refillRate,
		"refill_interval": tb.refillInterval.String(),
		"allowed":         tb.allowed,
		"rejected":        tb.rejected,
	}
}

// refill adds tokens based on elapsed time since last refill.
func (tb *TokenBucket) refill(t time.Time) {
	elapsed := t.Sub(tb.lastRefillTime)
	if elapsed < 0 {
		elapsed = 0
	}

	tokensToAdd := (float64(elapsed) / float64(tb.refillInterval)) * tb.refillRate
	if tokensToAdd > 0 {
		tb.tokens = tb.tokens + tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefillTime = t
	}
}
