// Package ring implements a consistent hash ring (MD5) with virtual nodes.
package ring

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
)

// Ring is a consistent hash ring using MD5 and virtual nodes per physical node.
type Ring struct {
	mu         sync.RWMutex
	vnodes     int
	points     map[uint64]string // hash -> nodeID
	sortedKeys []uint64
}

// New creates a ring where each physical node contributes vnodes virtual points.
func New(vnodes int) *Ring {
	return &Ring{
		vnodes: vnodes,
		points: make(map[uint64]string),
	}
}

func hash(key string) uint64 {
	sum := md5.Sum([]byte(key))
	return binary.BigEndian.Uint64(sum[:8])
}

// Add places vnodes virtual points for nodeID onto the ring.
func (r *Ring) Add(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := 0; i < r.vnodes; i++ {
		vkey := fmt.Sprintf("%s#%d", nodeID, i)
		h := hash(vkey)
		r.points[h] = nodeID
		r.sortedKeys = append(r.sortedKeys, h)
	}
	sort.Slice(r.sortedKeys, func(i, j int) bool { return r.sortedKeys[i] < r.sortedKeys[j] })
}

// Remove drops every virtual point owned by nodeID from the ring.
func (r *Ring) Remove(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var kept []uint64
	for _, h := range r.sortedKeys {
		if r.points[h] == nodeID {
			delete(r.points, h)
			continue
		}
		kept = append(kept, h)
	}
	r.sortedKeys = kept
}

// Get returns the node responsible for key.
func (r *Ring) Get(key string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.sortedKeys) == 0 {
		return "", false
	}
	h := hash(key)
	idx := sort.Search(len(r.sortedKeys), func(i int) bool { return r.sortedKeys[i] >= h })
	if idx == len(r.sortedKeys) {
		idx = 0
	}
	return r.points[r.sortedKeys[idx]], true
}

// Nodes returns the distinct physical nodes currently on the ring, in ring order.
func (r *Ring) Nodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var nodes []string
	for _, h := range r.sortedKeys {
		id := r.points[h]
		if !seen[id] {
			seen[id] = true
			nodes = append(nodes, id)
		}
	}
	return nodes
}

// VNodeCount returns how many virtual points nodeID currently owns.
func (r *Ring) VNodeCount(nodeID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, h := range r.sortedKeys {
		if r.points[h] == nodeID {
			count++
		}
	}
	return count
}

// Point is a single virtual-node point on the ring.
type Point struct {
	Hash uint64
	Node string
}

// Points returns every vnode point on the ring sorted by hash, used to
// visualize the ring arcs per physical node.
func (r *Ring) Points() []Point {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pts := make([]Point, 0, len(r.sortedKeys))
	for _, h := range r.sortedKeys {
		pts = append(pts, Point{Hash: h, Node: r.points[h]})
	}
	return pts
}

// Len returns the number of distinct physical nodes on the ring.
func (r *Ring) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	for _, h := range r.sortedKeys {
		seen[r.points[h]] = true
	}
	return len(seen)
}
