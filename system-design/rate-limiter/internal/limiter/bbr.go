package limiter

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// sample represents a single bandwidth and RTT observation.
type sample struct {
	bandwidth float64
	rtt       time.Duration
}

// BBR implements a simplified BBR (Bottleneck Bandwidth and Round-trip propagation time)
// rate limiting algorithm. It estimates bandwidth and RTT from recent request samples
// to dynamically adjust an allow probability.
type BBR struct {
	mu               sync.Mutex
	maxBandwidth     float64       // Estimated max bandwidth (requests per second)
	minRTT           time.Duration // Estimated minimum RTT
	gain             float64       // Multiplicative gain factor (default: 2.0)
	windowSize       int           // Number of samples in the sliding window
	samples          []sample      // Ring buffer of recent samples
	nextSampleIndex  int           // Next write position in ring buffer
	allowed          int64         // Total allowed requests
	rejected         int64         // Total rejected requests
	allowProbability float64       // Current probability of allowing a request
	lastRequestTime  time.Time     // Time of the last request
}

// NewBBR creates a new BBR rate limiter.
func NewBBR(initialBandwidth float64, initialRTT time.Duration, windowSize int, gain float64) *BBR {
	if initialBandwidth <= 0 {
		initialBandwidth = 100
	}
	if initialRTT <= 0 {
		initialRTT = 10 * time.Millisecond
	}
	if windowSize <= 0 {
		windowSize = 10
	}
	if gain <= 0 {
		gain = 2.0
	}

	return &BBR{
		maxBandwidth:     initialBandwidth,
		minRTT:           initialRTT,
		gain:             gain,
		windowSize:       windowSize,
		samples:          make([]sample, windowSize),
		nextSampleIndex:  0,
		allowProbability: 1.0,
		lastRequestTime:  time.Now(),
	}
}

// Allow checks if a request is allowed at the current time.
func (b *BBR) Allow() bool {
	return b.AllowAt(time.Now())
}

// AllowAt checks if a request would be allowed at the specified time.
func (b *BBR) AllowAt(t time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.updateEstimate(t)

	// Use the allow probability to decide
	if b.allowProbability >= 1.0 || b.allowProbability > randomFloat() {
		b.allowed++
		b.recordSample(t, true)
		return true
	}
	b.rejected++
	b.recordSample(t, false)
	return false
}

// Name returns the name of the rate limiter algorithm.
func (b *BBR) Name() string {
	return "bbr"
}

// Stats returns current statistics about the rate limiter state.
func (b *BBR) Stats() map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()

	return map[string]any{
		"limiter_type":      b.Name(),
		"max_bandwidth":     b.maxBandwidth,
		"min_rtt":           b.minRTT.String(),
		"gain":              b.gain,
		"window_size":       b.windowSize,
		"allow_probability": b.allowProbability,
		"allowed":           b.allowed,
		"rejected":          b.rejected,
	}
}

// updateEstimate adjusts the bandwidth and RTT estimates and recalculates allow probability.
func (b *BBR) updateEstimate(t time.Time) {
	elapsed := t.Sub(b.lastRequestTime)
	if elapsed < 0 {
		elapsed = 0
	}

	// Calculate observed bandwidth from inter-arrival time
	var observedBandwidth float64
	if elapsed > 0 {
		observedBandwidth = 1.0 / elapsed.Seconds()
	} else {
		observedBandwidth = b.maxBandwidth * 2 // Burst: high bandwidth
	}
	observedRTT := elapsed

	// Update max bandwidth (exponential moving average)
	if observedBandwidth > b.maxBandwidth {
		b.maxBandwidth = observedBandwidth
	} else {
		// Decay towards observed bandwidth
		b.maxBandwidth = b.maxBandwidth*0.9 + observedBandwidth*0.1
	}

	// Update min RTT
	if observedRTT < b.minRTT && observedRTT > 0 {
		b.minRTT = observedRTT
	} else if observedRTT > b.minRTT {
		// Allow minRTT to slowly increase if consistently higher
		// Convert to float64 seconds for computation, then back to Duration
		avgRTT := time.Duration(float64(b.minRTT)*0.95 + float64(observedRTT)*0.05)
		b.minRTT = avgRTT
	}

	// Calculate BDP (Bandwidth-Delay Product)
	bdp := b.maxBandwidth * b.minRTT.Seconds()

	// Adjust allow probability based on BDP and gain
	if bdp > 0 {
		// Higher BDP means more capacity available
		targetRate := b.gain * b.maxBandwidth
		// Normalize: if we're below target, allow more; if above, allow less
		currentRate := b.maxBandwidth
		if targetRate > 0 {
			b.allowProbability = math.Min(1.0, currentRate/targetRate)
		}
	} else {
		b.allowProbability = 1.0
	}

	b.lastRequestTime = t
}

// recordSample adds a new observation to the sliding window.
func (b *BBR) recordSample(t time.Time, wasAllowed bool) {
	elapsed := t.Sub(b.lastRequestTime)
	if elapsed < 0 {
		elapsed = 0
	}

	var bw float64
	if elapsed > 0 {
		bw = 1.0 / elapsed.Seconds()
	} else {
		bw = b.maxBandwidth
	}

	b.samples[b.nextSampleIndex] = sample{
		bandwidth: bw,
		rtt:       elapsed,
	}
	b.nextSampleIndex = (b.nextSampleIndex + 1) % b.windowSize
}

// randomFloat returns a random float64 in [0, 1) for probability decisions.
// Using a var so tests can override it with a deterministic version.
var randomFloat = func() float64 {
	return rand.Float64()
}
