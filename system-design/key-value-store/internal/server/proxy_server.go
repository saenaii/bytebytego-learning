package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/bytebytego-learning/key-value-store/internal/consistent"
	"github.com/bytebytego-learning/key-value-store/internal/store"
	"github.com/bytebytego-learning/key-value-store/internal/vectorclock"
)

type Quorum struct {
	N int
	W int
	R int
}

type ProxyServer struct {
	mu        sync.RWMutex
	ring      *consistent.Ring
	nodeAddrs map[string]string
	quorum    Quorum
	mux       *http.ServeMux
}

func NewProxyServer(virtualNodes int, quorum Quorum) *ProxyServer {
	ps := &ProxyServer{
		ring:      consistent.New(virtualNodes),
		nodeAddrs: make(map[string]string),
		quorum:    quorum,
		mux:       http.NewServeMux(),
	}
	ps.mux.HandleFunc("/kv/", ps.handleKV)
	ps.mux.HandleFunc("/health", ps.handleHealth)
	ps.mux.HandleFunc("/status", ps.handleStatus)
	ps.mux.HandleFunc("/admin/node", ps.handleAdminNode)
	return ps
}

func (ps *ProxyServer) Serve(ln net.Listener) error {
	log.Printf("Proxy listening on %s", ln.Addr().String())
	return http.Serve(ln, ps.mux)
}

func (ps *ProxyServer) AddNode(nodeID, nodeAddr string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.nodeAddrs[nodeID] = nodeAddr
	ps.ring.Add(nodeID)
	log.Printf("Node added: %s at %s", nodeID, nodeAddr)
}

func (ps *ProxyServer) RemoveNode(nodeID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.nodeAddrs, nodeID)
	ps.ring.Remove(nodeID)
	log.Printf("Node removed: %s", nodeID)
}

func (ps *ProxyServer) handleAdminNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Action   string `json:"action"`
		NodeID   string `json:"node_id"`
		NodeAddr string `json:"node_addr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	switch req.Action {
	case "add":
		if req.NodeID == "" || req.NodeAddr == "" {
			http.Error(w, "node_id and node_addr required", http.StatusBadRequest)
			return
		}
		ps.AddNode(req.NodeID, req.NodeAddr)
		writeJSON(w, map[string]any{"status": "ok", "action": "add", "node": req.NodeID})
	case "remove":
		if req.NodeID == "" {
			http.Error(w, "node_id required", http.StatusBadRequest)
			return
		}
		ps.RemoveNode(req.NodeID)
		writeJSON(w, map[string]any{"status": "ok", "action": "remove", "node": req.NodeID})
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}

func (ps *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ps.mu.RLock()
	nodeCount := len(ps.nodeAddrs)
	ps.mu.RUnlock()
	writeJSON(w, map[string]any{"status": "ok", "nodes": nodeCount})
}

func (ps *ProxyServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	ps.mu.RLock()
	nodes := ps.ring.Nodes()
	addrs := make(map[string]string)
	for _, n := range nodes {
		addrs[n] = ps.nodeAddrs[n]
	}
	ps.mu.RUnlock()
	writeJSON(w, map[string]any{
		"nodes":      nodes,
		"addresses":  addrs,
		"node_count": len(nodes),
	})
}

func (ps *ProxyServer) handleKV(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		ps.handleGet(w, r, key)
	case http.MethodPut:
		ps.handlePut(w, r, key)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Write quorum: send to all N replicas, succeed if >= W acks
func (ps *ProxyServer) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	ps.mu.RLock()
	targets := ps.ring.GetN(key, ps.quorum.N)
	ps.mu.RUnlock()

	if len(targets) < ps.quorum.W {
		http.Error(w, "not enough nodes for write quorum", http.StatusServiceUnavailable)
		return
	}

	// Read current clock from quorum so the write dominates existing versions
	prevClock := ps.readClock(key, targets)

	var (
		acks   int
		lastErr error
		clocks []*vectorclock.Clock
	)
	for _, nodeID := range targets {
		addr, ok := ps.nodeAddr(nodeID)
		if !ok {
			continue
		}
		clock, err := ps.remotePut(addr, key, req.Value, prevClock)
		if err != nil {
			log.Printf("put to %s failed: %v", nodeID, err)
			lastErr = err
			continue
		}
		acks++
		clocks = append(clocks, clock)
	}

	if acks < ps.quorum.W {
		errMsg := fmt.Sprintf("write quorum not met: %d/%d", acks, ps.quorum.W)
		if lastErr != nil {
			errMsg += fmt.Sprintf(": %v", lastErr)
		}
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	merged := vectorclock.New()
	for _, c := range clocks {
		merged.Merge(c)
	}
	writeJSON(w, map[string]any{"status": "ok", "clock": merged})
}

// readClock collects the merged clock from R replicas for read-before-write
func (ps *ProxyServer) readClock(key string, targets []string) *vectorclock.Clock {
	merged := vectorclock.New()
	reads := 0
	for _, nodeID := range targets {
		if reads >= ps.quorum.R {
			break
		}
		addr, ok := ps.nodeAddr(nodeID)
		if !ok {
			continue
		}
		entry, err := ps.remoteGet(addr, key)
		if err != nil {
			continue
		}
		reads++
		if entry != nil {
			for _, vv := range entry.Values {
				merged.Merge(vv.Clock)
			}
		}
	}
	if reads == 0 {
		return nil
	}
	return merged
}

// Read quorum: read from all N replicas, succeed if >= R responses
// Return siblings when concurrent conflicts are detected
func (ps *ProxyServer) handleGet(w http.ResponseWriter, r *http.Request, key string) {
	ps.mu.RLock()
	targets := ps.ring.GetN(key, ps.quorum.N)
	ps.mu.RUnlock()

	if len(targets) < ps.quorum.R {
		http.Error(w, "not enough nodes for read quorum", http.StatusServiceUnavailable)
		return
	}

	var (
		entries []*store.VersionedValue
		reads   int
	)
	for _, nodeID := range targets {
		addr, ok := ps.nodeAddr(nodeID)
		if !ok {
			continue
		}
		entry, err := ps.remoteGet(addr, key)
		if err != nil {
			log.Printf("get from %s failed: %v", nodeID, err)
			continue
		}
		reads++
		if entry != nil {
			for i := 0; i < len(entry.Values); i++ {
				entries = append(entries, &entry.Values[i])
			}
		}
	}

	if reads < ps.quorum.R {
		http.Error(w, "read quorum not met", http.StatusInternalServerError)
		return
	}

	if len(entries) == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	siblings := resolveSiblings(entries)
	writeJSON(w, &store.Entry{Values: siblings})
}

// resolveSiblings reduces a list of versioned values to only the latest/conflicting ones
func resolveSiblings(values []*store.VersionedValue) []store.VersionedValue {
	var latest []store.VersionedValue
	for _, v := range values {
		dominated := false
		for i := len(latest) - 1; i >= 0; i-- {
			rel := v.Clock.Compare(latest[i].Clock)
			switch rel {
			case vectorclock.Before:
				dominated = true
			case vectorclock.After:
				latest = append(latest[:i], latest[i+1:]...)
			}
		}
		if !dominated {
			latest = append(latest, *v)
		}
	}
	if len(latest) > 1 {
		log.Printf("Returning %d siblings", len(latest))
	}
	return latest
}

func (ps *ProxyServer) nodeAddr(nodeID string) (string, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	addr, ok := ps.nodeAddrs[nodeID]
	return addr, ok
}

func (ps *ProxyServer) remotePut(addr, key, value string, clock *vectorclock.Clock) (*vectorclock.Clock, error) {
	body := map[string]any{"value": value}
	if clock != nil {
		body["clock"] = clock
	}
	data, _ := json.Marshal(body)
	url := fmt.Sprintf("http://%s/kv/%s", addr, key)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remote put failed: %d %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		Status string             `json:"status"`
		Clock  *vectorclock.Clock `json:"clock"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Clock, nil
}

func (ps *ProxyServer) remoteGet(addr, key string) (*store.Entry, error) {
	url := fmt.Sprintf("http://%s/kv/%s", addr, key)
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("remote get failed: %d %s", resp.StatusCode, string(respBody))
	}
	var entry store.Entry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
