package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/bytebytego-learning/rate-limiter/internal/limiter"
)

// Server wraps an HTTP server with a rate limiter.
type Server struct {
	limiter limiter.RateLimiter
	port    int
	mux     *http.ServeMux
}

// NewServer creates a new HTTP server with the given rate limiter and port.
func NewServer(l limiter.RateLimiter, port int) *Server {
	s := &Server{
		limiter: l,
		port:    port,
		mux:     http.NewServeMux(),
	}

	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/allow", s.handleAllow)
	s.mux.HandleFunc("/stats", s.handleStats)

	return s
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Rate limiter server starting on %s (limiter: %s)", addr, s.limiter.Name())
	return http.ListenAndServe(addr, s.mux)
}

// ServeHTTP implements the http.Handler interface for testing with httptest.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleHealth returns the health status of the service.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "ok",
		"limiter_type": s.limiter.Name(),
	})
}

// handleAllow checks if a request is allowed by the rate limiter.
func (s *Server) handleAllow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	allowed := s.limiter.Allow()

	w.Header().Set("Content-Type", "application/json")
	if allowed {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusTooManyRequests)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"allowed":      allowed,
		"limiter_type": s.limiter.Name(),
	})
}

// handleStats returns current statistics from the rate limiter.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.limiter.Stats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ParsePort parses a port string to an int, returning the default if invalid.
func ParsePort(portStr string, defaultPort int) int {
	if portStr == "" {
		return defaultPort
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return defaultPort
	}
	return port
}
