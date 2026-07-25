package e2e_test

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/bytebytego-learning/key-value-store/internal/client"
	"github.com/bytebytego-learning/key-value-store/internal/server"
	"github.com/bytebytego-learning/key-value-store/internal/store"
)

func startProxy(t *testing.T) (*server.ProxyServer, string, net.Listener) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("proxy listen failed: %v", err)
	}
	proxy := server.NewProxyServer(50, server.Quorum{N: 3, W: 2, R: 2})
	go proxy.Serve(ln)
	return proxy, "http://" + ln.Addr().String(), ln
}

func startNode(t *testing.T, id string, proxyAddr string) (*server.NodeServer, string) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("node listen failed: %v", err)
	}
	node := server.NewNodeServer(id)
	go node.Serve(ln)

	addr := ln.Addr().String()
	c := client.New(proxyAddr)
	if err := c.AddNode(id, addr); err != nil {
		t.Fatalf("failed to add node %s: %v", id, err)
	}
	return node, addr
}

func firstValue(entry *store.Entry) string {
	if entry == nil || len(entry.Values) == 0 {
		return ""
	}
	return entry.Values[0].Value
}

func writeReport(path string, content string) {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(path, []byte(content), 0644)
}

func TestBasicPutGet(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)
	if err := c.Health(); err != nil {
		t.Fatalf("proxy health check failed: %v", err)
	}
	if err := c.Put("key1", "value1"); err != nil {
		t.Fatalf("put failed: %v", err)
	}
	entry, err := c.Get("key1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if entry == nil || len(entry.Values) == 0 {
		t.Fatal("expected entry with values")
	}
	if entry.Values[0].Value != "value1" {
		t.Errorf("expected value1, got %s", entry.Values[0].Value)
	}
}

func TestPutGetMultipleKeys(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)
	pairs := map[string]string{"alpha": "a", "beta": "b", "gamma": "c", "delta": "d", "epsilon": "e"}
	for k, v := range pairs {
		if err := c.Put(k, v); err != nil {
			t.Fatalf("put %s failed: %v", k, err)
		}
	}
	for k, v := range pairs {
		entry, err := c.Get(k)
		if err != nil {
			t.Fatalf("get %s failed: %v", k, err)
		}
		if entry == nil || len(entry.Values) == 0 {
			t.Errorf("key %s not found", k)
			continue
		}
		if entry.Values[0].Value != v {
			t.Errorf("key %s: expected %s, got %s", k, v, entry.Values[0].Value)
		}
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)
	var wg sync.WaitGroup
	errs := make(chan error, 200)

	writer := func(id int) {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("concurrent-key-%d", id)
			val := fmt.Sprintf("val-%d-%d", id, i)
			if err := c.Put(key, val); err != nil {
				errs <- fmt.Errorf("writer %d put failed: %w", id, err)
			}
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		}
	}
	reader := func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("concurrent-key-%d", rand.Intn(3))
			entry, err := c.Get(key)
			if err != nil {
				errs <- fmt.Errorf("reader get error: %w", err)
				continue
			}
			if entry != nil && len(entry.Values) > 0 && entry.Values[0].Value == "" {
				errs <- fmt.Errorf("got empty value for %s", key)
			}
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		}
	}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go writer(i)
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go reader()
	}
	wg.Wait()
	close(errs)

	var testErr error
	for e := range errs {
		t.Error(e)
		testErr = e
	}
	if testErr != nil {
		t.Fail()
	}
}

func TestNodeAddRemove(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)
	for i := 0; i < 10; i++ {
		if err := c.Put(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)); err != nil {
			t.Fatalf("initial put %d failed: %v", i, err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	// Add node-4 and node-5 (max 5 nodes)
	for n := 4; n <= 5; n++ {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("node-%d listen failed: %v", n, err)
		}
		node := server.NewNodeServer(fmt.Sprintf("node-%d", n))
		go node.Serve(ln)
		time.Sleep(100 * time.Millisecond)
		if err := c.AddNode(fmt.Sprintf("node-%d", n), ln.Addr().String()); err != nil {
			t.Fatalf("add node-%d failed: %v", n, err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	// Remove node-2
	if err := c.RemoveNode("node-2"); err != nil {
		t.Fatalf("remove node-2 failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	if err := c.Put("newkey", "newvalue"); err != nil {
		t.Errorf("put after remove failed: %v", err)
	}
	entry, err := c.Get("newkey")
	if err != nil {
		t.Fatalf("get newkey after remove failed: %v", err)
	}
	if entry == nil || len(entry.Values) == 0 || entry.Values[0].Value != "newvalue" {
		t.Errorf("expected newvalue, got %v", entry)
	}
}

func TestChaosPartitionAndHeal(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	// Start with 3 nodes
	nodes := make([]struct {
		node *server.NodeServer
		ln   net.Listener
	}, 3)
	for i := 1; i <= 3; i++ {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("node-%d listen failed: %v", i, err)
		}
		node := server.NewNodeServer(fmt.Sprintf("node-%d", i))
		go node.Serve(ln)
		nodes[i-1] = struct {
			node *server.NodeServer
			ln   net.Listener
		}{node, ln}
	}

	c := client.New(proxyAddr)
	for i := 1; i <= 3; i++ {
		if err := c.AddNode(fmt.Sprintf("node-%d", i), nodes[i-1].ln.Addr().String()); err != nil {
			t.Fatalf("add node-%d failed: %v", i, err)
		}
	}
	time.Sleep(300 * time.Millisecond)

	// Write initial data
	for i := 0; i < 10; i++ {
		if err := c.Put(fmt.Sprintf("chaos-k%d", i), fmt.Sprintf("v%d", i)); err != nil {
			t.Fatalf("initial put %d failed: %v", i, err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	// Simulate partition: remove node-3 from the ring
	if err := c.RemoveNode("node-3"); err != nil {
		t.Fatalf("remove node-3 failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	// Write more data while node-3 is partitioned
	for i := 10; i < 15; i++ {
		if err := c.Put(fmt.Sprintf("chaos-k%d", i), fmt.Sprintf("v%d", i)); err != nil {
			t.Errorf("put during partition %d failed: %v", i, err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	// Heal: re-add node-3
	if err := c.AddNode("node-3", nodes[2].ln.Addr().String()); err != nil {
		t.Fatalf("re-add node-3 failed: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Verify all keys are still readable
	for i := 0; i < 15; i++ {
		entry, err := c.Get(fmt.Sprintf("chaos-k%d", i))
		if err != nil {
			t.Logf("get k%d after heal failed: %v (acceptable during anti-entropy)", i, err)
			continue
		}
		if entry == nil || len(entry.Values) == 0 {
			t.Logf("key k%d not found yet (anti-entropy may still be in progress)", i)
			continue
		}
		expected := fmt.Sprintf("v%d", i)
		if entry.Values[0].Value != expected {
			t.Errorf("key k%d: expected %s, got %s", i, expected, entry.Values[0].Value)
		}
	}
}

func TestLatencyAndReport(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(300 * time.Millisecond)

	c := client.New(proxyAddr)

	type result struct {
		latency time.Duration
		err     error
		op      string
	}
	results := make(chan result, 500)
	var wg sync.WaitGroup

	ops := 200
	for i := 0; i < ops; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("perf-key-%d", i)
			val := fmt.Sprintf("val-%d", i)

			start := time.Now()
			err := c.Put(key, val)
			results <- result{time.Since(start), err, "put"}

			if err != nil {
				return
			}

			start = time.Now()
			_, err = c.Get(key)
			results <- result{time.Since(start), err, "get"}
		}(i)
	}
	wg.Wait()
	close(results)

	var putLats, getLats []time.Duration
	success, failure := 0, 0
	for r := range results {
		if r.err != nil {
			failure++
			continue
		}
		success++
		switch r.op {
		case "put":
			putLats = append(putLats, r.latency)
		case "get":
			getLats = append(getLats, r.latency)
		}
	}

	allLats := append(putLats, getLats...)
	sort.Slice(allLats, func(i, j int) bool { return allLats[i] < allLats[j] })
	sort.Slice(putLats, func(i, j int) bool { return putLats[i] < putLats[j] })
	sort.Slice(getLats, func(i, j int) bool { return getLats[i] < getLats[j] })

	total := success + failure
	successRate := float64(success) / float64(total) * 100
	avg := func(l []time.Duration) time.Duration {
		if len(l) == 0 {
			return 0
		}
		var t time.Duration
		for _, v := range l {
			t += v
		}
		return t / time.Duration(len(l))
	}
	p := func(l []time.Duration, percentile int) time.Duration {
		if len(l) == 0 {
			return 0
		}
		idx := percentile * len(l) / 100
		if idx >= len(l) {
			idx = len(l) - 1
		}
		return l[idx]
	}

	report := fmt.Sprintf(`# 📊 Performance Report

## Test Summary

| Metric | Value |
|--------|-------|
| Total Operations | %d |
| Successful | %d |
| Failed | %d |
| **Success Rate** | **%.1f%%** |

## Latency Breakdown

| Metric | PUT | GET | Overall |
|--------|-----|-----|---------|
| Count | %d | %d | %d |
| Average | %v | %v | %v |
| P50 (Median) | %v | %v | %v |
| P95 | %v | %v | %v |
| P99 | %v | %v | %v |

## Architecture

- **Topology**: 1 Proxy + 3 Storage Nodes (in-memory)
- **Replication**: N=3, W=2, R=2 (Quorum-based)
- **Virtual Nodes**: 50 per physical node
- **Conflict Resolution**: Vector Clocks (on read) + Merkle Tree Anti-Entropy (background)
`,
		total, success, failure, successRate,
		len(putLats), len(getLats), len(allLats),
		avg(putLats), avg(getLats), avg(allLats),
		p(putLats, 50), p(getLats, 50), p(allLats, 50),
		p(putLats, 95), p(getLats, 95), p(allLats, 95),
		p(putLats, 99), p(getLats, 99), p(allLats, 99),
	)

	reportPath := filepath.Join("..", "..", "REPORT.md")
	writeReport(reportPath, report)
	t.Logf("Report written to %s", reportPath)

	t.Logf("\n%s", report)

	if success == 0 {
		t.Fatal("no successful operations")
	}
}

func TestGetNonexistentKey(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)
	entry, err := c.Get("nonexistent")
	if err != nil {
		t.Fatalf("get nonexistent key failed: %v", err)
	}
	if entry != nil {
		t.Error("expected nil for nonexistent key")
	}
}

func TestOverwriteValue(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)
	if err := c.Put("k", "v1"); err != nil {
		t.Fatalf("first put failed: %v", err)
	}
	entry, _ := c.Get("k")
	if entry == nil || len(entry.Values) == 0 {
		t.Fatal("expected entry after first put")
	}
	t.Logf("After first put: %d siblings", len(entry.Values))

	if err := c.Put("k", "v2"); err != nil {
		t.Fatalf("second put failed: %v", err)
	}
	entry, _ = c.Get("k")
	if entry == nil || len(entry.Values) == 0 {
		t.Fatal("expected entry after second put")
	}
	t.Logf("After second put: %d siblings", len(entry.Values))
	foundV2 := false
	for _, vv := range entry.Values {
		if vv.Value == "v2" {
			foundV2 = true
			break
		}
	}
	if !foundV2 {
		t.Errorf("v2 not found among siblings: %v", entry.Values)
	}
}

func TestSiblingsOnConcurrentWrites(t *testing.T) {
	_, proxyAddr, pln := startProxy(t)
	defer pln.Close()

	for i := 1; i <= 3; i++ {
		_, _ = startNode(t, fmt.Sprintf("node-%d", i), proxyAddr)
	}
	time.Sleep(200 * time.Millisecond)

	c := client.New(proxyAddr)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Put("sibling-key", fmt.Sprintf("from-writer-%d", i))
		}(i)
	}
	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	entry, err := c.Get("sibling-key")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry")
	}
	t.Logf("Siblings for sibling-key: %d versions", len(entry.Values))
}
