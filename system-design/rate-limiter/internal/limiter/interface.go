package limiter

import "time"

// RateLimiter defines the contract that all rate limiter implementations must satisfy.
type RateLimiter interface {
	// Allow checks if a request is allowed at the current time.
	// Returns true if the request should be processed, false if it should be rejected.
	Allow() bool

	// AllowAt checks if a request would be allowed at the specified time.
	// Useful for testing and simulation.
	AllowAt(t time.Time) bool

	// Name returns the name of the rate limiter algorithm.
	Name() string

	// Stats returns current statistics about the rate limiter state.
	Stats() map[string]any
}
