package metrics

import "testing"

func TestInc(t *testing.T) {
	tests := []struct {
		name   string
		metric string
		field  func(Snapshot) int64
	}{
		{name: "shorten", metric: "shorten", field: func(s Snapshot) int64 { return s.Shorten }},
		{name: "resolve", metric: "resolve", field: func(s Snapshot) int64 { return s.Resolve }},
		{name: "redirect", metric: "redirect", field: func(s Snapshot) int64 { return s.Redirect }},
		{name: "cache hit", metric: "cache_hit", field: func(s Snapshot) int64 { return s.CacheHit }},
		{name: "cache miss", metric: "cache_miss", field: func(s Snapshot) int64 { return s.CacheMiss }},
		{name: "errors", metric: "errors", field: func(s Snapshot) int64 { return s.Errors }},
		{name: "requests", metric: "requests", field: func(s Snapshot) int64 { return s.Requests }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &Metrics{}
			collector.Inc(tt.metric)
			if got := tt.field(collector.Snapshot()); got != 1 {
				t.Fatalf("metric = %d, want 1", got)
			}
		})
	}
}
