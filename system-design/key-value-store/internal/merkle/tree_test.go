package merkle_test

import (
	"testing"

	"github.com/bytebytego-learning/key-value-store/internal/merkle"
)

func TestTree(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "empty tree has empty root",
			run: func(t *testing.T) {
				tree := merkle.New()
				if tree.Root() != "" {
					t.Errorf("expected empty root, got %s", tree.Root())
				}
			},
		},
		{
			name: "update changes root hash",
			run: func(t *testing.T) {
				tree := merkle.New()
				tree.Update("key1", "value1")
				r1 := tree.Root()
				if r1 == "" {
					t.Fatal("expected non-empty root")
				}
				tree.Update("key2", "value2")
				r2 := tree.Root()
				if r2 == r1 {
					t.Error("expected root to change after adding key")
				}
			},
		},
		{
			name: "remove changes root hash",
			run: func(t *testing.T) {
				tree := merkle.New()
				tree.Update("key1", "value1")
				tree.Update("key2", "value2")
				r1 := tree.Root()
				tree.Remove("key1")
				r2 := tree.Root()
				if r2 == r1 {
					t.Error("expected root to change after removal")
				}
			},
		},
		{
			name: "diff finds added keys",
			run: func(t *testing.T) {
				a := merkle.New()
				a.Update("key1", "value1")
				b := merkle.New()
				b.Update("key1", "value1")
				b.Update("key2", "value2")
				diff := a.Diff(b)
				if len(diff) != 1 || diff[0] != "key2" {
					t.Errorf("expected [key2], got %v", diff)
				}
			},
		},
		{
			name: "diff finds modified keys",
			run: func(t *testing.T) {
				a := merkle.New()
				a.Update("key1", "value1")
				b := merkle.New()
				b.Update("key1", "value2")
				diff := a.Diff(b)
				if len(diff) != 1 || diff[0] != "key1" {
					t.Errorf("expected [key1], got %v", diff)
				}
			},
		},
		{
			name: "diff finds removed keys",
			run: func(t *testing.T) {
				a := merkle.New()
				a.Update("key1", "value1")
				a.Update("key2", "value2")
				b := merkle.New()
				b.Update("key1", "value1")
				diff := a.Diff(b)
				if len(diff) != 1 || diff[0] != "key2" {
					t.Errorf("expected [key2], got %v", diff)
				}
			},
		},
		{
			name: "identical trees have empty diff",
			run: func(t *testing.T) {
				a := merkle.New()
				a.Update("key1", "value1")
				a.Update("key2", "value2")
				b := merkle.New()
				b.Update("key1", "value1")
				b.Update("key2", "value2")
				diff := a.Diff(b)
				if len(diff) != 0 {
					t.Errorf("expected empty diff, got %v", diff)
				}
			},
		},
		{
			name: "Keys returns sorted keys",
			run: func(t *testing.T) {
				tree := merkle.New()
				tree.Update("z", "1")
				tree.Update("a", "2")
				tree.Update("m", "3")
				keys := tree.Keys()
				expected := []string{"a", "m", "z"}
				if len(keys) != len(expected) {
					t.Fatalf("expected %v, got %v", expected, keys)
				}
				for i := range keys {
					if keys[i] != expected[i] {
						t.Errorf("position %d: expected %s, got %s", i, expected[i], keys[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
