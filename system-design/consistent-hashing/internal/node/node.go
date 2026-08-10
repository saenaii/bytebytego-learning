// Package node implements a worker node HTTP service that stores keys
// and reports metrics to the proxy.
package node

import (
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Server is a worker node in the consistent hashing cluster.
type Server struct {
	id         string
	latency    time.Duration
	jitter     time.Duration
	mu         sync.RWMutex
	store      map[string]string
	requests   atomic.Int64
	putCount   atomic.Int64
	getCount   atomic.Int64
	mux        *http.ServeMux
}

// Option configures a worker node server.
type Option func(*Server)

// WithLatency simulates real-world processing latency with the given jitter.
func WithLatency(base, jitter time.Duration) Option {
	return func(s *Server) { s.latency, s.jitter = base, jitter }
}

// New creates a worker node server for id.
func New(id string, opts ...Option) *Server {
	s := &Server{
		id:    id,
		store: make(map[string]string),
		mux:   http.NewServeMux(),
	}
	for _, o := range opts {
		o(s)
	}
	s.mux.HandleFunc("/kv/", s.handleKV)
	s.mux.HandleFunc("/keys", s.handleKeys)
	s.mux.HandleFunc("/stats", s.handleStats)
	s.mux.HandleFunc("/health", s.handleHealth)
	return s
}

// Serve blocks, serving HTTP on ln.
func (s *Server) Serve(ln net.Listener) error {
	log.Printf("node %s listening on %s", s.id, ln.Addr().String())
	return http.Serve(ln, s.mux)
}

// ID returns the node identifier.
func (s *Server) ID() string { return s.id }

func (s *Server) simulateLatency() {
	if s.latency <= 0 {
		return
	}
	d := s.latency
	if s.jitter > 0 {
		d += time.Duration(rand.Int63n(int64(s.jitter)))
	}
	time.Sleep(d)
}

func (s *Server) handleKV(w http.ResponseWriter, r *http.Request) {
	s.simulateLatency()
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPut:
		s.handlePut(w, r, key)
	case http.MethodGet:
		s.handleGet(w, r, key)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var req struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.store[key] = req.Value
	s.mu.Unlock()
	s.putCount.Add(1)
	s.requests.Add(1)
	writeJSON(w, map[string]any{"status": "ok", "node": s.id, "key": key})
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request, key string) {
	s.mu.RLock()
	value, ok := s.store[key]
	s.mu.RUnlock()
	s.getCount.Add(1)
	s.requests.Add(1)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"status": "ok", "node": s.id, "key": key, "value": value})
}

// handleKeys returns the sorted key set stored on this node (used to
// measure key migration across topology changes).
func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	s.mu.RUnlock()
	sort.Strings(keys)
	writeJSON(w, map[string]any{"node": s.id, "count": len(keys), "keys": keys})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	keyCount := len(s.store)
	s.mu.RUnlock()
	writeJSON(w, map[string]any{
		"node":     s.id,
		"requests": s.requests.Load(),
		"puts":     s.putCount.Load(),
		"gets":     s.getCount.Load(),
		"keys":     keyCount,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"status": "ok", "node": s.id})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
