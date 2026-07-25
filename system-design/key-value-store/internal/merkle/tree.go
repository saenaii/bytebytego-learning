package merkle

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
)

type Tree struct {
	mu     sync.RWMutex
	root   string
	leaves map[string]string
}

func New() *Tree {
	return &Tree{
		leaves: make(map[string]string),
	}
}

func leafHash(key, value string) string {
	h := sha256.Sum256([]byte(key + ":" + value))
	return fmt.Sprintf("%x", h)
}

func internalHash(left, right string) string {
	if left > right {
		left, right = right, left
	}
	h := sha256.Sum256([]byte(left + right))
	return fmt.Sprintf("%x", h)
}

func (t *Tree) Update(key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.leaves[key] = leafHash(key, value)
	t.rebuild()
}

func (t *Tree) Remove(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.leaves, key)
	t.rebuild()
}

func (t *Tree) rebuild() {
	if len(t.leaves) == 0 {
		t.root = ""
		return
	}

	keys := make([]string, 0, len(t.leaves))
	for k := range t.leaves {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hashes := make([]string, 0, len(keys))
	for _, k := range keys {
		hashes = append(hashes, t.leaves[k])
	}

	for len(hashes) > 1 {
		var next []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				next = append(next, internalHash(hashes[i], hashes[i+1]))
			} else {
				next = append(next, hashes[i])
			}
		}
		hashes = next
	}
	t.root = hashes[0]
}

func (t *Tree) Root() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

func (t *Tree) Diff(other *Tree) []string {
	allKeys := make(map[string]bool)

	t.mu.RLock()
	for k := range t.leaves {
		allKeys[k] = true
	}
	myLeaves := make(map[string]string, len(t.leaves))
	for k, v := range t.leaves {
		myLeaves[k] = v
	}
	t.mu.RUnlock()

	other.mu.RLock()
	for k := range other.leaves {
		allKeys[k] = true
	}
	otherLeaves := make(map[string]string, len(other.leaves))
	for k, v := range other.leaves {
		otherLeaves[k] = v
	}
	other.mu.RUnlock()

	var diff []string
	for k := range allKeys {
		if myLeaves[k] != otherLeaves[k] {
			diff = append(diff, k)
		}
	}
	sort.Strings(diff)
	return diff
}

func (t *Tree) Keys() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	keys := make([]string, 0, len(t.leaves))
	for k := range t.leaves {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
