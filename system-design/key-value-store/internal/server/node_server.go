package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bytebytego-learning/key-value-store/internal/merkle"
	"github.com/bytebytego-learning/key-value-store/internal/store"
	"github.com/bytebytego-learning/key-value-store/internal/vectorclock"
)

type NodeServer struct {
	id           string
	store        *store.Store
	tree         *merkle.Tree
	mux          *http.ServeMux
	replicas     map[string]string
	antiEntropy  bool
}

func NewNodeServer(id string) *NodeServer {
	ns := &NodeServer{
		id:       id,
		store:    store.New(),
		tree:     merkle.New(),
		mux:      http.NewServeMux(),
		replicas: make(map[string]string),
	}
	ns.mux.HandleFunc("/kv/", ns.handleKV)
	ns.mux.HandleFunc("/health", ns.handleHealth)
	ns.mux.HandleFunc("/sync/merkle", ns.handleSyncMerkle)
	ns.mux.HandleFunc("/sync/keys", ns.handleSyncKeys)
	ns.mux.HandleFunc("/sync/repair", ns.handleSyncRepair)
	return ns
}

func (ns *NodeServer) Serve(ln net.Listener) error {
	log.Printf("Node %s listening on %s", ns.id, ln.Addr().String())
	return http.Serve(ln, ns.mux)
}

func (ns *NodeServer) SetReplicas(replicas map[string]string) {
	ns.replicas = replicas
}

func (ns *NodeServer) StartAntiEntropy(interval time.Duration) {
	if ns.antiEntropy {
		return
	}
	ns.antiEntropy = true
	go ns.antiEntropyLoop(interval)
}

func (ns *NodeServer) antiEntropyLoop(interval time.Duration) {
	for {
		time.Sleep(interval)
		ns.runAntiEntropy()
	}
}

func (ns *NodeServer) runAntiEntropy() {
	for replicaID, addr := range ns.replicas {
		if replicaID == ns.id {
			continue
		}
		if err := ns.syncWithReplica(addr); err != nil {
			log.Printf("Anti-entropy with %s (%s) failed: %v", replicaID, addr, err)
		}
	}
}

func (ns *NodeServer) syncWithReplica(addr string) error {
	// 1. Compare Merkle roots
	remoteRoot, err := ns.fetchMerkleRoot(addr)
	if err != nil {
		return err
	}
	localRoot := ns.tree.Root()
	if localRoot == remoteRoot {
		return nil
	}

	// 2. Roots differ, fetch all keys from remote and compare
	remoteData, err := ns.fetchAllKeys(addr)
	if err != nil {
		return err
	}
	localData := ns.store.All()

	// 3. Build repair maps
	var repairToRemote, repairFromRemote map[string]*store.Entry

	for key, remoteEntry := range remoteData {
		localEntry, hasLocal := localData[key]
		if !hasLocal {
			if repairToRemote == nil {
				repairToRemote = make(map[string]*store.Entry)
			}
			repairToRemote[key] = remoteEntry
			continue
		}
		for _, rvv := range remoteEntry.Values {
			stale := true
			for _, lvv := range localEntry.Values {
				rel := lvv.Clock.Compare(rvv.Clock)
				if rel == vectorclock.After || rel == vectorclock.Equal {
					stale = false
					break
				}
			}
			if stale {
				if repairToRemote == nil {
					repairToRemote = make(map[string]*store.Entry)
				}
				repairToRemote[key] = remoteEntry
			}
		}
	}

	for key, remoteEntry := range remoteData {
		localEntry, hasLocal := localData[key]
		if !hasLocal {
			if repairFromRemote == nil {
				repairFromRemote = make(map[string]*store.Entry)
			}
			repairFromRemote[key] = remoteEntry
			continue
		}
		needsRepair := false
		for _, rvv := range remoteEntry.Values {
			localNewer := false
			for _, lvv := range localEntry.Values {
				rel := lvv.Clock.Compare(rvv.Clock)
				if rel == vectorclock.After || rel == vectorclock.Equal {
					localNewer = true
					break
				}
			}
			if !localNewer {
				needsRepair = true
				break
			}
		}
		if needsRepair {
			if repairFromRemote == nil {
				repairFromRemote = make(map[string]*store.Entry)
			}
			repairFromRemote[key] = remoteEntry
		}
	}

	// 5. Send repairs
	if repairToRemote != nil {
		data, _ := json.Marshal(repairToRemote)
		http.Post("http://"+addr+"/sync/repair", "application/json", bytes.NewReader(data))
	}
	if repairFromRemote != nil {
		data, _ := json.Marshal(repairFromRemote)
		http.Post("http://"+addr+"/sync/repair", "application/json", bytes.NewReader(data))
	}

	return nil
}

func (ns *NodeServer) fetchMerkleRoot(addr string) (string, error) {
	resp, err := http.Get("http://" + addr + "/sync/merkle")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Root string `json:"root"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Root, nil
}

func (ns *NodeServer) fetchAllKeys(addr string) (map[string]*store.Entry, error) {
	resp, err := http.Get("http://" + addr + "/sync/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]*store.Entry
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (ns *NodeServer) Store() *store.Store {
	return ns.store
}

func (ns *NodeServer) Tree() *merkle.Tree {
	return ns.tree
}

func (ns *NodeServer) handleKV(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		entry, ok := ns.store.Get(key)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, entry)

	case http.MethodPut:
		var req struct {
			Value string             `json:"value"`
			Clock *vectorclock.Clock `json:"clock,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		clock := req.Clock
		if clock == nil {
			clock = vectorclock.New()
		}
		clock.Increment(ns.id)
		ns.store.Put(key, req.Value, clock)
		ns.tree.Update(key, req.Value)
		writeJSON(w, map[string]any{"status": "ok", "clock": clock})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ns *NodeServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"status": "ok", "node": ns.id})
}

func (ns *NodeServer) handleSyncMerkle(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"root": ns.tree.Root()})
}

func (ns *NodeServer) handleSyncKeys(w http.ResponseWriter, r *http.Request) {
	all := ns.store.All()
	writeJSON(w, all)
}

func (ns *NodeServer) handleSyncRepair(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	var entries map[string]*store.Entry
	if err := json.Unmarshal(body, &entries); err != nil {
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}
	for key, entry := range entries {
		for _, vv := range entry.Values {
			ns.store.Put(key, vv.Value, vv.Clock)
			ns.tree.Update(key, vv.Value)
		}
	}
	writeJSON(w, map[string]any{"status": "ok", "repaired": len(entries)})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}
