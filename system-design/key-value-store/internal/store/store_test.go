package store_test

import (
	"testing"

	"github.com/bytebytego-learning/key-value-store/internal/store"
	"github.com/bytebytego-learning/key-value-store/internal/vectorclock"
)

func TestStore(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "get nonexistent key returns false",
			run: func(t *testing.T) {
				s := store.New()
				_, ok := s.Get("missing")
				if ok {
					t.Error("expected false for missing key")
				}
			},
		},
		{
			name: "put and get returns stored value",
			run: func(t *testing.T) {
				s := store.New()
				clock := vectorclock.New()
				clock.Increment("node-1")
				s.Put("key1", "value1", clock)
				entry, ok := s.Get("key1")
				if !ok {
					t.Fatal("expected key1 to exist")
				}
				if len(entry.Values) != 1 {
					t.Fatalf("expected 1 value, got %d", len(entry.Values))
				}
				if entry.Values[0].Value != "value1" {
					t.Errorf("expected value1, got %s", entry.Values[0].Value)
				}
				if entry.Values[0].Clock.Compare(clock) != vectorclock.Equal {
					t.Error("expected clocks to match")
				}
			},
		},
		{
			name: "put overwrites when clock dominates",
			run: func(t *testing.T) {
				s := store.New()
				c1 := vectorclock.New()
				c1.Increment("node-1")
				s.Put("k", "v1", c1)

				c2 := vectorclock.New()
				c2.Increment("node-1")
				c2.Increment("node-1")
				s.Put("k", "v2", c2)

				entry, _ := s.Get("k")
				if len(entry.Values) != 1 {
					t.Errorf("expected 1 value, got %d", len(entry.Values))
				}
				if entry.Values[0].Value != "v2" {
					t.Errorf("expected v2, got %s", entry.Values[0].Value)
				}
			},
		},
		{
			name: "put creates siblings on concurrent clocks",
			run: func(t *testing.T) {
				s := store.New()
				c1 := vectorclock.New()
				c1.Increment("node-1")
				s.Put("k", "v1", c1)

				c2 := vectorclock.New()
				c2.Increment("node-2")
				s.Put("k", "v2", c2)

				entry, _ := s.Get("k")
				if len(entry.Values) != 2 {
					t.Errorf("expected 2 siblings, got %d", len(entry.Values))
				}
			},
		},
		{
			name: "put ignores stale clock",
			run: func(t *testing.T) {
				s := store.New()
				c1 := vectorclock.New()
				c1.Increment("node-1")
				c1.Increment("node-1")
				s.Put("k", "v1", c1)

				c2 := vectorclock.New()
				c2.Increment("node-1")
				s.Put("k", "v2", c2)

				entry, _ := s.Get("k")
				if entry.Values[0].Value != "v1" {
					t.Errorf("expected v1 (stale write ignored), got %s", entry.Values[0].Value)
				}
			},
		},
		{
			name: "delete removes key",
			run: func(t *testing.T) {
				s := store.New()
				s.Put("k", "v", vectorclock.New())
				s.Delete("k")
				_, ok := s.Get("k")
				if ok {
					t.Error("expected key to be deleted")
				}
			},
		},
		{
			name: "Keys returns all keys",
			run: func(t *testing.T) {
				s := store.New()
				s.Put("a", "1", vectorclock.New())
				s.Put("b", "2", vectorclock.New())
				s.Put("c", "3", vectorclock.New())
				keys := s.Keys()
				if len(keys) != 3 {
					t.Errorf("expected 3 keys, got %d", len(keys))
				}
			},
		},
		{
			name: "Len returns correct count",
			run: func(t *testing.T) {
				s := store.New()
				if s.Len() != 0 {
					t.Errorf("expected 0, got %d", s.Len())
				}
				s.Put("k", "v", vectorclock.New())
				if s.Len() != 1 {
					t.Errorf("expected 1, got %d", s.Len())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
