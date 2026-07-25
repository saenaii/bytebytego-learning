package store

import (
	"sync"

	"github.com/bytebytego-learning/key-value-store/internal/vectorclock"
)

type VersionedValue struct {
	Value string             `json:"value"`
	Clock *vectorclock.Clock `json:"clock"`
}

type Entry struct {
	Values []VersionedValue `json:"values"`
}

type Store struct {
	mu   sync.RWMutex
	data map[string]*Entry
}

func New() *Store {
	return &Store{
		data: make(map[string]*Entry),
	}
}

func (s *Store) Get(key string) (*Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[key]
	if !ok {
		return nil, false
	}
	return copyEntry(e), true
}

func (s *Store) Put(key, value string, clock *vectorclock.Clock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.data[key]
	if !ok {
		s.data[key] = &Entry{
			Values: []VersionedValue{{Value: value, Clock: clock.Copy()}},
		}
		return
	}

	allBefore := true
	var kept []VersionedValue
	for _, vv := range existing.Values {
		rel := vv.Clock.Compare(clock)
		switch rel {
		case vectorclock.Before:
		case vectorclock.Equal, vectorclock.After:
			allBefore = false
			kept = append(kept, vv)
		case vectorclock.Concurrent:
			allBefore = false
			kept = append(kept, vv)
		}
	}
	if allBefore {
		existing.Values = []VersionedValue{{Value: value, Clock: clock.Copy()}}
		return
	}
	kept = append(kept, VersionedValue{Value: value, Clock: clock.Copy()})
	existing.Values = kept
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

func (s *Store) All() map[string]*Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*Entry, len(s.data))
	for k, v := range s.data {
		result[k] = copyEntry(v)
	}
	return result
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func copyEntry(e *Entry) *Entry {
	cp := &Entry{Values: make([]VersionedValue, len(e.Values))}
	for i, vv := range e.Values {
		cp.Values[i] = VersionedValue{Value: vv.Value, Clock: vv.Clock.Copy()}
	}
	return cp
}
