package metrics

import "sync"

type Snapshot struct {
	Shorten   int64 `json:"shorten"`
	Resolve   int64 `json:"resolve"`
	Redirect  int64 `json:"redirect"`
	CacheHit  int64 `json:"cache_hit"`
	CacheMiss int64 `json:"cache_miss"`
	Errors    int64 `json:"errors"`
	Requests  int64 `json:"requests"`
}

type Metrics struct {
	mu    sync.RWMutex
	value Snapshot
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch name {
	case "shorten":
		m.value.Shorten++
	case "resolve":
		m.value.Resolve++
	case "redirect":
		m.value.Redirect++
	case "cache_hit":
		m.value.CacheHit++
	case "cache_miss":
		m.value.CacheMiss++
	case "errors":
		m.value.Errors++
	case "requests":
		m.value.Requests++
	}
}

func (m *Metrics) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}
