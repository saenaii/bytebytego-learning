package limiter

import (
	"sync"
	"time"
)

// LeakingBucket implements the leaking bucket rate limiting algorithm.
// A FIFO queue of fixed size holds incoming requests. Requests are "leaked"
// (processed) at a constant rate. If the queue is full, the request is rejected.
type LeakingBucket struct {
	mu           sync.Mutex
	queueSize    int           // Maximum queue capacity
	queue        []time.Time   // FIFO queue of request timestamps
	leakRate     float64       // Number of items leaked per second
	leakInterval time.Duration // Time between leaks
	lastLeakTime time.Time     // Last time an item was leaked
	allowed      int64         // Total allowed requests
	rejected     int64         // Total rejected requests
}

// NewLeakingBucket creates a new LeakingBucket rate limiter.
func NewLeakingBucket(queueSize int, leakRate float64) *LeakingBucket {
	if queueSize < 0 {
		queueSize = 0
	}
	if leakRate < 0 {
		leakRate = 0
	}

	leakInterval := time.Second
	if leakRate > 0 {
		leakInterval = time.Duration(float64(time.Second) / leakRate)
		if leakInterval < time.Millisecond {
			leakInterval = time.Millisecond
		}
	}

	return &LeakingBucket{
		queueSize:    queueSize,
		queue:        make([]time.Time, 0, queueSize),
		leakRate:     leakRate,
		leakInterval: leakInterval,
		lastLeakTime: time.Now(),
	}
}

// Allow checks if a request is allowed at the current time.
func (lb *LeakingBucket) Allow() bool {
	return lb.AllowAt(time.Now())
}

// AllowAt checks if a request would be allowed at the specified time.
func (lb *LeakingBucket) AllowAt(t time.Time) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak(t)

	if len(lb.queue) < lb.queueSize {
		lb.queue = append(lb.queue, t)
		lb.allowed++
		return true
	}
	lb.rejected++
	return false
}

// Name returns the name of the rate limiter algorithm.
func (lb *LeakingBucket) Name() string {
	return "leaking_bucket"
}

// Stats returns current statistics about the rate limiter state.
func (lb *LeakingBucket) Stats() map[string]any {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	return map[string]any{
		"limiter_type":  lb.Name(),
		"queue_size":    lb.queueSize,
		"queue_length":  len(lb.queue),
		"leak_rate":     lb.leakRate,
		"leak_interval": lb.leakInterval.String(),
		"allowed":       lb.allowed,
		"rejected":      lb.rejected,
	}
}

// leak removes items from the queue based on elapsed time.
func (lb *LeakingBucket) leak(t time.Time) {
	if lb.leakRate <= 0 || lb.queueSize <= 0 {
		return
	}

	elapsed := t.Sub(lb.lastLeakTime)
	if elapsed < 0 {
		elapsed = 0
	}

	itemsToLeak := int(float64(elapsed) / float64(lb.leakInterval))
	if itemsToLeak > 0 {
		if itemsToLeak >= len(lb.queue) {
			lb.queue = lb.queue[:0]
		} else {
			lb.queue = lb.queue[itemsToLeak:]
		}
		lb.lastLeakTime = t
	}
}
