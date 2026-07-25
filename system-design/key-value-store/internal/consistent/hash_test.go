package consistent_test

import (
	"testing"

	"github.com/bytebytego-learning/key-value-store/internal/consistent"
)

func TestRing(t *testing.T) {
	tests := []struct {
		name     string
		run      func(t *testing.T, r *consistent.Ring)
		setup    func(r *consistent.Ring)
	}{
		{
			name: "add and get returns a node",
			setup: func(r *consistent.Ring) {
				r.Add("node-1")
				r.Add("node-2")
			},
			run: func(t *testing.T, r *consistent.Ring) {
				node, ok := r.Get("key1")
				if !ok {
					t.Fatal("expected to find a node")
				}
				if node != "node-1" && node != "node-2" {
					t.Fatalf("expected node-1 or node-2, got %s", node)
				}
			},
		},
		{
			name: "remove redistributes keys",
			setup: func(r *consistent.Ring) {
				r.Add("node-1")
				r.Add("node-2")
				r.Add("node-3")
			},
			run: func(t *testing.T, r *consistent.Ring) {
				keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
				m1 := make(map[string]string)
				for _, k := range keys {
					n, _ := r.Get(k)
					m1[k] = n
				}
				r.Remove("node-2")
				moved := 0
				for _, k := range keys {
					n, ok := r.Get(k)
					if !ok {
						t.Fatalf("key %s has no node after removal", k)
					}
					if m1[k] != n && m1[k] == "node-2" {
						moved++
					} else if m1[k] != n && m1[k] != "node-2" {
						t.Errorf("key %s moved from %s to %s unexpectedly", k, m1[k], n)
					}
				}
				if moved == 0 {
					t.Error("expected some keys to move after node-2 removal")
				}
			},
		},
		{
			name: "GetN returns multiple distinct nodes",
			setup: func(r *consistent.Ring) {
				r.Add("node-1")
				r.Add("node-2")
				r.Add("node-3")
			},
			run: func(t *testing.T, r *consistent.Ring) {
				nodes := r.GetN("mykey", 2)
				if len(nodes) != 2 {
					t.Fatalf("expected 2 nodes, got %d", len(nodes))
				}
				if nodes[0] == nodes[1] {
					t.Error("expected different nodes")
				}
			},
		},
		{
			name: "empty ring returns false",
			run: func(t *testing.T, r *consistent.Ring) {
				_, ok := r.Get("key")
				if ok {
					t.Error("expected false for empty ring")
				}
				nodes := r.GetN("key", 2)
				if nodes != nil {
					t.Error("expected nil for empty ring")
				}
			},
		},
		{
			name: "Nodes returns all added nodes",
			setup: func(r *consistent.Ring) {
				r.Add("node-1")
				r.Add("node-2")
			},
			run: func(t *testing.T, r *consistent.Ring) {
				nodes := r.Nodes()
				if len(nodes) != 2 {
					t.Fatalf("expected 2 nodes, got %d", len(nodes))
				}
			},
		},
		{
			name: "deterministic routing for same key",
			setup: func(r *consistent.Ring) {
				r.Add("node-1")
				r.Add("node-2")
			},
			run: func(t *testing.T, r *consistent.Ring) {
				n1, _ := r.Get("test-key")
				n2, _ := r.Get("test-key")
				if n1 != n2 {
					t.Errorf("expected same node, got %s and %s", n1, n2)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := consistent.New(50)
			if tt.setup != nil {
				tt.setup(r)
			}
			tt.run(t, r)
		})
	}
}
