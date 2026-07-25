package consistent

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
)

type Ring struct {
	mu           sync.RWMutex
	virtualNodes int
	nodeMap      map[uint64]string
	sortedKeys   []uint64
}

func New(virtualNodes int) *Ring {
	return &Ring{
		virtualNodes: virtualNodes,
		nodeMap:      make(map[uint64]string),
	}
}

func hash(key string) uint64 {
	h := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(h[:8])
}

func (r *Ring) Add(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < r.virtualNodes; i++ {
		virtualKey := fmt.Sprintf("%s:%d", nodeID, i)
		h := hash(virtualKey)
		r.nodeMap[h] = nodeID
		r.sortedKeys = append(r.sortedKeys, h)
	}
	sort.Slice(r.sortedKeys, func(i, j int) bool {
		return r.sortedKeys[i] < r.sortedKeys[j]
	})
}

func (r *Ring) Remove(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var newKeys []uint64
	for _, h := range r.sortedKeys {
		if r.nodeMap[h] != nodeID {
			newKeys = append(newKeys, h)
		} else {
			delete(r.nodeMap, h)
		}
	}
	r.sortedKeys = newKeys
}

func (r *Ring) Get(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sortedKeys) == 0 {
		return "", false
	}
	h := hash(key)
	idx := sort.Search(len(r.sortedKeys), func(i int) bool {
		return r.sortedKeys[i] >= h
	})
	if idx == len(r.sortedKeys) {
		idx = 0
	}
	return r.nodeMap[r.sortedKeys[idx]], true
}

func (r *Ring) GetN(key string, n int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.nodeMap) == 0 || n <= 0 {
		return nil
	}
	h := hash(key)
	idx := sort.Search(len(r.sortedKeys), func(i int) bool {
		return r.sortedKeys[i] >= h
	})
	var result []string
	seen := make(map[string]bool)
	for i := 0; i < len(r.sortedKeys) && len(result) < n; i++ {
		ringIdx := (idx + i) % len(r.sortedKeys)
		nodeID := r.nodeMap[r.sortedKeys[ringIdx]]
		if !seen[nodeID] {
			seen[nodeID] = true
			result = append(result, nodeID)
		}
	}
	return result
}

func (r *Ring) Nodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var nodes []string
	for _, h := range r.sortedKeys {
		nodeID := r.nodeMap[h]
		if !seen[nodeID] {
			seen[nodeID] = true
			nodes = append(nodes, nodeID)
		}
	}
	return nodes
}
