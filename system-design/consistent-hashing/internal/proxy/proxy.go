// Package proxy implements the consistent hashing gateway: it owns the
// hash ring, routes /kv requests to the responsible worker, serves the
// embedded dashboard and exposes the admin + metrics APIs.
package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytebytego-learning/consistent-hashing/internal/dashboard"
	"github.com/bytebytego-learning/consistent-hashing/internal/ring"
)

// Config configures a proxy server.
type Config struct {
	VNodes        int
	ClientTimeout time.Duration
	Logger        *log.Logger
}

// Proxy is the consistent hashing gateway.
type Proxy struct {
	vnodes int
	ring   *ring.Ring
	log    *log.Logger
	client *http.Client
	mux    *http.ServeMux

	mu       sync.RWMutex
	registry map[string]string // nodeID -> base URL

	topoChanges atomic.Int64

	snapMu      sync.RWMutex
	snapshot    map[string]string // key -> owner at last topology change
	hasSnapshot bool

	statMu   sync.Mutex
	prevStat map[string]int64
	prevTime time.Time
}

// New creates a proxy with the given configuration.
func New(cfg Config) *Proxy {
	if cfg.VNodes <= 0 {
		cfg.VNodes = 50
	}
	timeout := cfg.ClientTimeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}
	p := &Proxy{
		vnodes:   cfg.VNodes,
		ring:     ring.New(cfg.VNodes),
		log:      logger,
		client:   &http.Client{Timeout: timeout},
		mux:      http.NewServeMux(),
		registry: make(map[string]string),
		prevStat: make(map[string]int64),
	}
	p.mux.HandleFunc("/", p.handleRoot)
	p.mux.HandleFunc("/kv/", p.handleKV)
	p.mux.HandleFunc("/register", p.handleRegister)
	p.mux.HandleFunc("/admin/nodes/", p.handleAdminNode)
	p.mux.HandleFunc("/stats", p.handleStats)
	p.mux.HandleFunc("/health", p.handleHealth)
	return p
}

// ServeHTTP implements http.Handler.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

// Serve blocks, serving HTTP on ln.
func (p *Proxy) Serve(ln net.Listener) error {
	p.log.Printf("proxy listening on %s (vnodes=%d)", ln.Addr().String(), p.vnodes)
	return http.Serve(ln, p.mux)
}

// Ring exposes the underlying hash ring for tests and tooling.
func (p *Proxy) Ring() *ring.Ring { return p.ring }

// handleRoot serves the embedded dashboard.
func (p *Proxy) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	html, err := dashboard.FS.ReadFile(dashboard.IndexHTML)
	if err != nil {
		http.Error(w, "dashboard not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}

// handleKV routes PUT/GET /kv/{key} to the responsible worker node.
func (p *Proxy) handleKV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	node, ok := p.ring.Get(key)
	if !ok {
		http.Error(w, "no active nodes", http.StatusServiceUnavailable)
		return
	}
	p.mu.RLock()
	addr := p.registry[node]
	p.mu.RUnlock()
	if addr == "" {
		http.Error(w, fmt.Sprintf("node %s not registered", node), http.StatusServiceUnavailable)
		return
	}
	target := addr + "/kv/" + urlPathEscape(key)
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		http.Error(w, "bad upstream request", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		p.log.Printf("kv %s %s -> %s failed: %v", r.Method, key, node, err)
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func urlPathEscape(key string) string {
	replacer := strings.NewReplacer(
		"/", "%2F",
		" ", "%20",
		"?", "%3F",
		"#", "%23",
		"&", "%26",
		"%", "%25",
	)
	return replacer.Replace(key)
}

// handleRegister stores the address of a worker so the admin API can later
// enable it. Registration alone does not put the node on the ring.
func (p *Proxy) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID   string `json:"id"`
		Addr string `json:"addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.ID == "" || req.Addr == "" {
		http.Error(w, "id and addr required", http.StatusBadRequest)
		return
	}
	p.mu.Lock()
	p.registry[req.ID] = req.Addr
	p.mu.Unlock()
	active := false
	for _, n := range p.ring.Nodes() {
		if n == req.ID {
			active = true
			break
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "id": req.ID, "active": active})
}

// handleAdminNode enables (POST) or disables (DELETE) a registered worker.
func (p *Proxy) handleAdminNode(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/admin/nodes/")
	if id == "" {
		http.Error(w, "missing node id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		p.enableNode(w, id)
	case http.MethodDelete:
		p.disableNode(w, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (p *Proxy) enableNode(w http.ResponseWriter, id string) {
	for _, n := range p.ring.Nodes() {
		if n == id {
			writeJSON(w, http.StatusConflict, map[string]any{"status": "error", "error": "node already active"})
			return
		}
	}
	p.mu.RLock()
	addr := p.registry[id]
	p.mu.RUnlock()
	if addr == "" {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "error", "error": "node not registered"})
		return
	}
	// Snapshot pre-change ownership so /stats can measure the migration caused
	// by this topology change.
	p.takeSnapshot()
	p.ring.Add(id)
	p.topoChanges.Add(1)
	p.log.Printf("node %s enabled (active=%d)", id, p.ring.Len())
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "node": id, "active": true})
}

func (p *Proxy) disableNode(w http.ResponseWriter, id string) {
	found := false
	for _, n := range p.ring.Nodes() {
		if n == id {
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "error", "error": "node not active"})
		return
	}
	p.takeSnapshot()
	p.ring.Remove(id)
	p.topoChanges.Add(1)
	p.log.Printf("node %s disabled (active=%d)", id, p.ring.Len())
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "node": id, "active": false})
}

// takeSnapshot records the keys and their owners as they are *before* a
// topology change; later /stats calls count how many of those keys now map
// to a different node (the "live migration counter").
func (p *Proxy) takeSnapshot() {
	snap := make(map[string]string)
	for _, id := range p.ring.Nodes() {
		p.mu.RLock()
		addr := p.registry[id]
		p.mu.RUnlock()
		if addr == "" {
			continue
		}
		keys, err := p.fetchKeys(addr)
		if err != nil {
			p.log.Printf("snapshot: fetch keys from %s failed: %v", id, err)
			continue
		}
		for _, k := range keys {
			if owner, ok := p.ring.Get(k); ok {
				snap[k] = owner
			}
		}
	}
	p.snapMu.Lock()
	p.snapshot = snap
	p.hasSnapshot = true
	p.snapMu.Unlock()
}

// fetchKeys returns the sorted key set stored on a worker node.
func (p *Proxy) fetchKeys(addr string) ([]string, error) {
	var resp struct {
		Keys []string `json:"keys"`
	}
	if err := p.getJSON(addr+"/keys", &resp); err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// computeMigration counts snapshot keys whose current ring owner differs.
func (p *Proxy) computeMigration() Migration {
	p.snapMu.RLock()
	defer p.snapMu.RUnlock()
	if !p.hasSnapshot {
		return Migration{}
	}
	moved := 0
	for key, owner := range p.snapshot {
		if cur, ok := p.ring.Get(key); ok && cur != owner {
			moved++
		}
	}
	return Migration{
		TrackedKeys:  len(p.snapshot),
		Moved:        int64(moved),
		MigrationPct: pct(float64(moved), float64(len(p.snapshot))),
	}
}

func pct(part, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return part / total * 100
}

// getJSON performs a GET and decodes a JSON response.
func (p *Proxy) getJSON(url string, out any) error {
	resp, err := p.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// NodeInfo is the per-node view returned by /stats.
type NodeInfo struct {
	ID       string  `json:"id"`
	Addr     string  `json:"addr"`
	Active   bool    `json:"active"`
	Requests int64   `json:"requests"`
	Puts     int64   `json:"puts"`
	Gets     int64   `json:"gets"`
	QPS      float64 `json:"qps"`
	Keys     int     `json:"keys"`  // keys physically stored on the node
	Owned    int     `json:"owned"` // keys the ring assigns to the node (ideal load)
	LoadPct  float64 `json:"load_pct"`
}

// RingArc is a vnode point for ring visualization.
type RingArc struct {
	Hash uint64 `json:"hash"`
	Node string `json:"node"`
}

// Migration reports keys moved since the last topology change.
type Migration struct {
	TrackedKeys  int     `json:"tracked_keys"`
	Moved        int64   `json:"moved"`
	MigrationPct float64 `json:"migration_pct"`
}

// Stats is the aggregate /stats response.
type Stats struct {
	Nodes           []NodeInfo `json:"nodes"`
	Standby         []string   `json:"standby"`
	ActiveCount     int        `json:"active_count"`
	TotalKeys       int        `json:"total_keys"`
	TotalRequests   int64      `json:"total_requests"`
	LoadStdDevPct   float64    `json:"load_stddev_pct"`
	MeanLoadPct     float64    `json:"mean_load_pct"`
	Migration       Migration  `json:"migration"`
	TopologyChanges int64      `json:"topology_changes"`
	VNodes          int        `json:"vnodes"`
	RingArcs        []RingArc  `json:"ring_arcs"`
}

type nodeStat struct {
	Requests int64 `json:"requests"`
	Puts     int64 `json:"puts"`
	Gets     int64 `json:"gets"`
	Keys     int   `json:"keys"`
}

func (p *Proxy) fetchNodeStat(addr string) (nodeStat, error) {
	var ns nodeStat
	err := p.getJSON(addr+"/stats", &ns)
	return ns, err
}

// handleStats aggregates per-node metrics, computes load balance, migration
// and the ring arcs for the dashboard.
func (p *Proxy) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p.mu.RLock()
	registry := make(map[string]string, len(p.registry))
	for id, addr := range p.registry {
		registry[id] = addr
	}
	p.mu.RUnlock()

	active := make(map[string]bool)
	for _, id := range p.ring.Nodes() {
		active[id] = true
	}

	st := Stats{
		Nodes:           []NodeInfo{},
		Standby:         []string{},
		VNodes:          p.vnodes,
		TopologyChanges: p.topoChanges.Load(),
		Migration:       p.computeMigration(),
	}

	now := time.Now()
	p.statMu.Lock()
	union := make(map[string]bool)
	for id, addr := range registry {
		info := NodeInfo{ID: id, Addr: addr, Active: active[id]}
		if info.Active {
			ns, err := p.fetchNodeStat(addr)
			if err != nil {
				p.log.Printf("stats: fetch %s failed: %v", id, err)
				st.Nodes = append(st.Nodes, info)
				continue
			}
			info.Requests = ns.Requests
			info.Puts = ns.Puts
			info.Gets = ns.Gets
			info.Keys = ns.Keys
			if prev, ok := p.prevStat[id]; ok && !p.prevTime.IsZero() {
				if elapsed := now.Sub(p.prevTime).Seconds(); elapsed > 0 {
					info.QPS = float64(ns.Requests-prev) / elapsed
					if info.QPS < 0 {
						info.QPS = 0
					}
				}
			}
			p.prevStat[id] = ns.Requests
			if keys, err := p.fetchKeys(addr); err == nil {
				for _, k := range keys {
					union[k] = true
				}
			}
		} else {
			st.Standby = append(st.Standby, id)
		}
		st.Nodes = append(st.Nodes, info)
	}
	p.prevTime = now
	p.statMu.Unlock()

	sort.Slice(st.Nodes, func(i, j int) bool { return st.Nodes[i].ID < st.Nodes[j].ID })
	sort.Strings(st.Standby)

	// Ideal load: how many keys in the cluster does the ring assign to each node.
	owned := make(map[string]int)
	for k := range union {
		if owner, ok := p.ring.Get(k); ok {
			owned[owner]++
		}
	}
	totalOwned := 0
	for _, c := range owned {
		totalOwned += c
	}
	ownedVals := make([]int, 0, len(owned))
	for i := range st.Nodes {
		n := &st.Nodes[i]
		if n.Active {
			n.Owned = owned[n.ID]
			n.LoadPct = pct(float64(n.Owned), float64(totalOwned))
			st.TotalRequests += n.Requests
			ownedVals = append(ownedVals, n.Owned)
		}
	}
	st.TotalKeys = totalOwned
	st.ActiveCount = len(active)
	if len(ownedVals) > 0 {
		st.LoadStdDevPct = loadStdDevPct(ownedVals)
		st.MeanLoadPct = 100 / float64(len(ownedVals))
	}

	for _, pt := range p.ring.Points() {
		st.RingArcs = append(st.RingArcs, RingArc{Hash: pt.Hash, Node: pt.Node})
	}
	writeJSON(w, http.StatusOK, st)
}

// loadStdDevPct returns std-dev / mean in percent for the given key counts.
func loadStdDevPct(counts []int) float64 {
	if len(counts) == 0 {
		return 0
	}
	var sum float64
	for _, c := range counts {
		sum += float64(c)
	}
	mean := sum / float64(len(counts))
	if mean <= 0 {
		return 0
	}
	var sq float64
	for _, c := range counts {
		d := float64(c) - mean
		sq += d * d
	}
	stddev := math.Sqrt(sq / float64(len(counts)))
	return stddev / mean * 100
}

func (p *Proxy) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
