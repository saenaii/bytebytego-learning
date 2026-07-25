package vectorclock_test

import (
	"testing"

	"github.com/bytebytego-learning/key-value-store/internal/vectorclock"
)

func TestClock(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new clock has empty counters",
			run: func(t *testing.T) {
				c := vectorclock.New()
				if c == nil {
					t.Fatal("expected non-nil clock")
				}
				if len(c.Counters) != 0 {
					t.Errorf("expected empty counters, got %d", len(c.Counters))
				}
			},
		},
		{
			name: "increment increases counter for node",
			run: func(t *testing.T) {
				c := vectorclock.New()
				c.Increment("node-1")
				if c.Counters["node-1"] != 1 {
					t.Errorf("expected 1, got %d", c.Counters["node-1"])
				}
				c.Increment("node-1")
				if c.Counters["node-1"] != 2 {
					t.Errorf("expected 2, got %d", c.Counters["node-1"])
				}
			},
		},
		{
			name: "merge takes max of each counter",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				a.Increment("node-1")
				a.Increment("node-2")

				b := vectorclock.New()
				b.Increment("node-1")
				b.Increment("node-3")

				a.Merge(b)
				tests := []struct {
					node  string
					want  uint64
				}{
					{"node-1", 2},
					{"node-2", 1},
					{"node-3", 1},
				}
				for _, tt := range tests {
					if a.Counters[tt.node] != tt.want {
						t.Errorf("node %s = %d, want %d", tt.node, a.Counters[tt.node], tt.want)
					}
				}
			},
		},
		{
			name: "compare equal clocks",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				b := vectorclock.New()
				b.Increment("node-1")
				if r := a.Compare(b); r != vectorclock.Equal {
					t.Errorf("expected Equal, got %v", r)
				}
			},
		},
		{
			name: "compare before",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				b := vectorclock.New()
				b.Increment("node-1")
				b.Increment("node-1")
				if r := a.Compare(b); r != vectorclock.Before {
					t.Errorf("expected Before, got %v", r)
				}
			},
		},
		{
			name: "compare after",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				a.Increment("node-1")
				b := vectorclock.New()
				b.Increment("node-1")
				if r := a.Compare(b); r != vectorclock.After {
					t.Errorf("expected After, got %v", r)
				}
			},
		},
		{
			name: "compare concurrent",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				b := vectorclock.New()
				b.Increment("node-2")
				if r := a.Compare(b); r != vectorclock.Concurrent {
					t.Errorf("expected Concurrent, got %v", r)
				}
			},
		},
		{
			name: "copy is independent",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				a.Increment("node-2")
				b := a.Copy()
				if a.Compare(b) != vectorclock.Equal {
					t.Error("expected copy to be equal")
				}
				b.Increment("node-1")
				if a.Compare(b) != vectorclock.Before {
					t.Error("expected original before copy after modifying copy")
				}
			},
		},
		{
			name: "marshal and unmarshal round-trip",
			run: func(t *testing.T) {
				a := vectorclock.New()
				a.Increment("node-1")
				a.Increment("node-2")
				data, err := a.Marshal()
				if err != nil {
					t.Fatalf("marshal error: %v", err)
				}
				b := vectorclock.New()
				if err := b.Unmarshal(data); err != nil {
					t.Fatalf("unmarshal error: %v", err)
				}
				if a.Compare(b) != vectorclock.Equal {
					t.Error("expected equal after round-trip")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
