// Package generator produces continuous random-key traffic against the proxy,
// mimicking real-world load: rate-limited, concurrent, and latency-measured.
package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Config controls the load profile.
type Config struct {
	ProxyURL  string
	Rate      int           // requests per second
	Keyspace  int           // number of distinct keys
	PutRatio  float64       // fraction of PUT ops in [0,1]
	Duration  time.Duration // 0 = run until Stop
	MaxConcurrent int       // bound on in-flight requests
	Timeout       time.Duration
}

// Result is one completed request.
type Result struct {
	Op      string // "put" or "get"
	Latency time.Duration
	Err     error
}

// Generator emits Result values for every completed request.
type Generator struct {
	cfg     Config
	client  *http.Client
	results chan Result
	stop    chan struct{}
	done    chan struct{}
	once    sync.Once
}

// New creates a generator with the given config.
func New(cfg Config) *Generator {
	if cfg.Rate <= 0 {
		cfg.Rate = 100
	}
	if cfg.Keyspace <= 0 {
		cfg.Keyspace = 10000
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 200
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	return &Generator{
		cfg:     cfg,
		client:  &http.Client{Timeout: cfg.Timeout},
		results: make(chan Result, 4096),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start launches the load loop in the background.
func (g *Generator) Start() {
	go g.run()
}

// Stop terminates the load loop and waits for it to finish.
func (g *Generator) Stop() {
	g.once.Do(func() { close(g.stop) })
	<-g.done
}

// Results returns the stream of completed request results.
func (g *Generator) Results() <-chan Result {
	return g.results
}

func (g *Generator) run() {
	interval := time.Second / time.Duration(g.cfg.Rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sem := make(chan struct{}, g.cfg.MaxConcurrent)
	var wg sync.WaitGroup
	var deadline <-chan time.Time
	if g.cfg.Duration > 0 {
		t := time.NewTimer(g.cfg.Duration)
		defer t.Stop()
		deadline = t.C
	}

loop:
	for {
		select {
		case <-ticker.C:
			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				g.results <- g.oneRequest()
			}()
		case <-deadline:
			break loop
		case <-g.stop:
			break loop
		}
	}
	// Drain in-flight requests before closing the results stream.
	wg.Wait()
	close(g.done)
	close(g.results)
}

func (g *Generator) oneRequest() Result {
	key := fmt.Sprintf("user-%d", rand.Intn(g.cfg.Keyspace))
	if rand.Float64() < g.cfg.PutRatio {
		return g.put(key)
	}
	return g.get(key)
}

func (g *Generator) put(key string) Result {
	body, _ := json.Marshal(map[string]any{"value": "v-" + key})
	req, _ := http.NewRequest(http.MethodPut, g.cfg.ProxyURL+"/kv/"+key, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := g.client.Do(req)
	if err != nil {
		return Result{Op: "put", Latency: time.Since(start), Err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return Result{Op: "put", Latency: time.Since(start), Err: fmt.Errorf("put %s: status %d", key, resp.StatusCode)}
	}
	return Result{Op: "put", Latency: time.Since(start)}
}

func (g *Generator) get(key string) Result {
	start := time.Now()
	resp, err := g.client.Get(g.cfg.ProxyURL + "/kv/" + key)
	if err != nil {
		return Result{Op: "get", Latency: time.Since(start), Err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Result{Op: "get", Latency: time.Since(start)}
	}
	if resp.StatusCode >= 500 {
		return Result{Op: "get", Latency: time.Since(start), Err: fmt.Errorf("get %s: status %d", key, resp.StatusCode)}
	}
	return Result{Op: "get", Latency: time.Since(start)}
}

// OpStats summarizes latencies for one operation type.
type OpStats struct {
	Count int
	Avg   time.Duration
	P50   time.Duration
	P95   time.Duration
	P99   time.Duration
}

// Summary aggregates a stream of results.
type Summary struct {
	Total       int
	Success     int
	Failure     int
	SuccessRate float64
	Puts        int
	Gets        int
	Put         OpStats
	Get         OpStats
}

func summarizeOp(latencies []time.Duration) OpStats {
	if len(latencies) == 0 {
		return OpStats{}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	percentile := func(p int) time.Duration {
		idx := p * len(latencies) / 100
		if idx >= len(latencies) {
			idx = len(latencies) - 1
		}
		return latencies[idx]
	}
	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	return OpStats{
		Count: len(latencies),
		Avg:   sum / time.Duration(len(latencies)),
		P50:   percentile(50),
		P95:   percentile(95),
		P99:   percentile(99),
	}
}

// Summarize computes a Summary from a completed results stream.
func Summarize(results []Result) Summary {
	s := Summary{}
	var putLats, getLats []time.Duration
	for _, r := range results {
		s.Total++
		if r.Err != nil {
			s.Failure++
			continue
		}
		s.Success++
		switch r.Op {
		case "put":
			s.Puts++
			putLats = append(putLats, r.Latency)
		case "get":
			s.Gets++
			getLats = append(getLats, r.Latency)
		}
	}
	if s.Total > 0 {
		s.SuccessRate = float64(s.Success) / float64(s.Total) * 100
	}
	s.Put = summarizeOp(putLats)
	s.Get = summarizeOp(getLats)
	return s
}
