package ring

import (
	"fmt"
	"math"
	"testing"
)

func TestGetEmptyRing(t *testing.T) {
	r := New(50)
	if _, ok := r.Get("key"); ok {
		t.Error("expected no node on empty ring")
	}
}

func TestGetSingleNode(t *testing.T) {
	r := New(50)
	r.Add("node-1")
	for _, key := range []string{"user-42", "a", "b", "zzz"} {
		node, ok := r.Get(key)
		if !ok || node != "node-1" {
			t.Errorf("key %s: got (%s,%v), want (node-1,true)", key, node, ok)
		}
	}
}

func TestRemovedNodeNeverServes(t *testing.T) {
	tests := []struct {
		name   string
		remove string
	}{
		{name: "remove node-2 of three", remove: "node-2"},
		{name: "remove last remaining node", remove: "node-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(50)
			for _, n := range []string{"node-1", "node-2", "node-3"} {
				r.Add(n)
			}
			if tt.remove != "" {
				r.Remove(tt.remove)
			}
			for i := 0; i < 500; i++ {
				key := fmt.Sprintf("user-%d", i)
				node, ok := r.Get(key)
				if node == tt.remove {
					t.Fatalf("key %s routed to removed node %s", key, node)
				}
				if tt.remove == "node-1" && !ok {
					continue
				}
				if !ok {
					t.Fatalf("expected a node for key %s", key)
				}
			}
		})
	}
}

func TestVirtualNodeCount(t *testing.T) {
	r := New(50)
	r.Add("node-1")
	r.Add("node-2")
	r.Add("node-3")
	if got := r.VNodeCount("node-1"); got != 50 {
		t.Errorf("vnodes for node-1: got %d, want 50", got)
	}
	r.Remove("node-1")
	if got := r.VNodeCount("node-1"); got != 0 {
		t.Errorf("vnodes after remove: got %d, want 0", got)
	}
}

func TestDistributionBalanced(t *testing.T) {
	r := New(50)
	for i := 1; i <= 3; i++ {
		r.Add(fmt.Sprintf("node-%d", i))
	}
	counts := map[string]int{}
	for i := 0; i < 3000; i++ {
		node, ok := r.Get(fmt.Sprintf("user-%d", i))
		if !ok {
			t.Fatal("expected node")
		}
		counts[node]++
	}
	if len(counts) != 3 {
		t.Fatalf("expected 3 nodes to receive keys, got %d", len(counts))
	}
	var values []float64
	for _, c := range counts {
		values = append(values, float64(c))
	}
	stddev := stddevOf(values)
	mean := meanOf(values)
	if stddev/mean > 0.15 {
		t.Errorf("distribution too skewed: counts=%v stddev/mean=%.3f", counts, stddev/mean)
	}
}

func TestMinimalMigrationOnScaleOut(t *testing.T) {
	r := New(50)
	for i := 1; i <= 3; i++ {
		r.Add(fmt.Sprintf("node-%d", i))
	}
	ownersBefore := map[string]string{}
	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("user-%d", i)
		node, _ := r.Get(key)
		ownersBefore[key] = node
	}
	r.Add("node-4")
	moved := 0
	for key, before := range ownersBefore {
		after, _ := r.Get(key)
		if after != before {
			moved++
		}
		_ = key
	}
	migrationPct := float64(moved) / float64(len(ownersBefore)) * 100
	// Adding 1 node to a 3-node ring moves ~25% of keys; naive rehash moves ~75%.
	if migrationPct < 18 || migrationPct > 32 {
		t.Errorf("migration %.1f%% out of expected 18-32%% range", migrationPct)
	}
}

func TestMinimalMigrationOnScaleIn(t *testing.T) {
	r := New(50)
	for i := 1; i <= 4; i++ {
		r.Add(fmt.Sprintf("node-%d", i))
	}
	ownersBefore := map[string]string{}
	for i := 0; i < 2000; i++ {
		key := fmt.Sprintf("user-%d", i)
		node, _ := r.Get(key)
		ownersBefore[key] = node
	}
	r.Remove("node-3")
	moved, removedOwned := 0, 0
	for key, before := range ownersBefore {
		if before == "node-3" {
			removedOwned++
		}
		after, _ := r.Get(key)
		if after != before {
			moved++
		}
	}
	if moved != removedOwned {
		t.Errorf("moved=%d, but only keys owned by removed node (%d) should move", moved, removedOwned)
	}
	migrationPct := float64(moved) / float64(len(ownersBefore)) * 100
	if migrationPct < 18 || migrationPct > 32 {
		t.Errorf("migration %.1f%% out of expected 18-32%% range", migrationPct)
	}
}

func meanOf(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stddevOf(values []float64) float64 {
	mean := meanOf(values)
	var sq float64
	for _, v := range values {
		d := v - mean
		sq += d * d
	}
	return math.Sqrt(sq / float64(len(values)))
}
