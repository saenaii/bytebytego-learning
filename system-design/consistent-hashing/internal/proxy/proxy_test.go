package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bytebytego-learning/consistent-hashing/internal/node"
)

func startTestNode(t *testing.T, id string) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := node.New(id)
	go srv.Serve(ln)
	t.Cleanup(func() { ln.Close() })
	return "http://" + ln.Addr().String()
}

func startTestProxy(t *testing.T) (*Proxy, string) {
	t.Helper()
	p := New(Config{VNodes: 50, Logger: log.New(testWriter{t}, "proxy-test: ", 0)})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go p.Serve(ln)
	t.Cleanup(func() { ln.Close() })
	return p, "http://" + ln.Addr().String()
}

type testWriter struct{ t *testing.T }

func (w testWriter) Write(b []byte) (int, error) {
	w.t.Log(strings.TrimSpace(string(b)))
	return len(b), nil
}

func registerAt(t *testing.T, proxyURL, id, addr string) {
	t.Helper()
	resp, err := http.Post(proxyURL+"/register", "application/json",
		strings.NewReader(fmt.Sprintf(`{"id":%q,"addr":%q}`, id, addr)))
	if err != nil {
		t.Fatalf("register %s: %v", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register %s: status %d", id, resp.StatusCode)
	}
}

func adminOp(t *testing.T, method, proxyURL, id string) int {
	t.Helper()
	req, err := http.NewRequest(method, proxyURL+"/admin/nodes/"+id, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, id, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func putKey(t *testing.T, proxyURL, key, value string) int {
	t.Helper()
	body := fmt.Sprintf(`{"value":%q}`, value)
	req, err := http.NewRequest(http.MethodPut, proxyURL+"/kv/"+key, strings.NewReader(body))
	if err != nil {
		t.Fatalf("put %s: %v", key, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("put %s: %v", key, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func getKeyStatus(t *testing.T, proxyURL, key string) (int, string) {
	t.Helper()
	resp, err := http.Get(proxyURL + "/kv/" + key)
	if err != nil {
		t.Fatalf("get %s: %v", key, err)
	}
	defer resp.Body.Close()
	var body struct {
		Value string `json:"value"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Value
}

func fetchStats(t *testing.T, proxyURL string) Stats {
	t.Helper()
	resp, err := http.Get(proxyURL + "/stats")
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	defer resp.Body.Close()
	var st Stats
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	return st
}

func TestRegisterAndRoute(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	addr := startTestNode(t, "node-1")
	registerAt(t, proxyURL, "node-1", addr)
	if code := adminOp(t, http.MethodPost, proxyURL, "node-1"); code != http.StatusOK {
		t.Fatalf("enable node-1: status %d", code)
	}

	if code := putKey(t, proxyURL, "user-42", "hello"); code != http.StatusOK {
		t.Errorf("put: status %d, want 200", code)
	}
	code, value := getKeyStatus(t, proxyURL, "user-42")
	if code != http.StatusOK || value != "hello" {
		t.Errorf("get: status=%d value=%q, want 200/hello", code, value)
	}
}

func TestAdminErrors(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	addr := startTestNode(t, "node-1")
	registerAt(t, proxyURL, "node-1", addr)
	if code := adminOp(t, http.MethodPost, proxyURL, "node-1"); code != http.StatusOK {
		t.Fatalf("initial enable node-1: %d", code)
	}

	tests := []struct {
		name   string
		method string
		id     string
		want   int
	}{
		{name: "enable unknown node", method: http.MethodPost, id: "ghost", want: http.StatusNotFound},
		{name: "enable twice", method: http.MethodPost, id: "node-1", want: http.StatusConflict},
		{name: "disable inactive node", method: http.MethodDelete, id: "ghost", want: http.StatusNotFound},
		{name: "method not allowed", method: http.MethodGet, id: "node-1", want: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := adminOp(t, tt.method, proxyURL, tt.id); code != tt.want {
				t.Errorf("got %d, want %d", code, tt.want)
			}
		})
	}
}

func TestKVBeforeAndAfterTopologyChange(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	addr1 := startTestNode(t, "node-1")
	registerAt(t, proxyURL, "node-1", addr1)
	if code := adminOp(t, http.MethodPost, proxyURL, "node-1"); code != http.StatusOK {
		t.Fatalf("enable node-1: %d", code)
	}

	putKey(t, proxyURL, "user-7", "v1")
	code, value := getKeyStatus(t, proxyURL, "user-7")
	if code != http.StatusOK || value != "v1" {
		t.Fatalf("before change: status=%d value=%q", code, value)
	}

	if code := adminOp(t, http.MethodDelete, proxyURL, "node-1"); code != http.StatusOK {
		t.Fatalf("remove node-1: %d", code)
	}
	if code, _ := getKeyStatus(t, proxyURL, "user-7"); code != http.StatusServiceUnavailable {
		t.Errorf("after removal: status %d, want 503 (empty ring)", code)
	}
}

func TestMigrationCounterOnScaleOut(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("node-%d", i)
		registerAt(t, proxyURL, id, startTestNode(t, id))
		if code := adminOp(t, http.MethodPost, proxyURL, id); code != http.StatusOK {
			t.Fatalf("enable %s: %d", id, code)
		}
	}

	for i := 0; i < 1500; i++ {
		if code := putKey(t, proxyURL, fmt.Sprintf("user-%d", i), "x"); code != http.StatusOK {
			t.Fatalf("seed put %d: %d", i, code)
		}
	}

	addr4 := startTestNode(t, "node-4")
	registerAt(t, proxyURL, "node-4", addr4)
	if code := adminOp(t, http.MethodPost, proxyURL, "node-4"); code != http.StatusOK {
		t.Fatalf("enable node-4: %d", code)
	}

	st := fetchStats(t, proxyURL)
	if st.Migration.TrackedKeys != 1500 {
		t.Errorf("tracked keys: got %d, want 1500", st.Migration.TrackedKeys)
	}
	if st.Migration.MigrationPct < 18 || st.Migration.MigrationPct > 32 {
		t.Errorf("migration %.1f%% out of 18-32%% range", st.Migration.MigrationPct)
	}
	if st.ActiveCount != 4 {
		t.Errorf("active count: got %d, want 4", st.ActiveCount)
	}
	if len(st.Standby) != 0 {
		t.Errorf("standby: got %v, want none", st.Standby)
	}

	addr5 := startTestNode(t, "node-5")
	registerAt(t, proxyURL, "node-5", addr5)
	st = fetchStats(t, proxyURL)
	if len(st.Standby) != 1 || st.Standby[0] != "node-5" {
		t.Errorf("standby after node-5 register: %v, want [node-5]", st.Standby)
	}
}

func TestStatsBalanceAndFailureThreshold(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("node-%d", i)
		registerAt(t, proxyURL, id, startTestNode(t, id))
		if code := adminOp(t, http.MethodPost, proxyURL, id); code != http.StatusOK {
			t.Fatalf("enable %s: %d", id, code)
		}
	}
	for i := 0; i < 3000; i++ {
		putKey(t, proxyURL, fmt.Sprintf("user-%d", i), "x")
	}
	st := fetchStats(t, proxyURL)
	if st.LoadStdDevPct > 15 {
		t.Errorf("load std-dev %.2f%% > 15%%: counts=%v", st.LoadStdDevPct, st.Nodes)
	}
	if st.TotalKeys != 3000 {
		t.Errorf("total keys: got %d, want 3000", st.TotalKeys)
	}
	active := 0
	for _, n := range st.Nodes {
		if n.Active {
			active++
			if n.LoadPct <= 0 || n.LoadPct > 60 {
				t.Errorf("node %s load %.1f%% out of plausible range", n.ID, n.LoadPct)
			}
		}
	}
	if active != 3 {
		t.Errorf("active nodes: got %d, want 3", active)
	}
}

func TestDashboardAndHealth(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	resp, err := http.Get(proxyURL + "/")
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	body := make([]byte, 200)
	n, _ := resp.Body.Read(body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("dashboard status: %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("dashboard content type: %s", resp.Header.Get("Content-Type"))
	}
	if !strings.Contains(string(body[:n]), "Consistent Hashing") {
		t.Error("dashboard does not reference Consistent Hashing")
	}

	resp, err = http.Get(proxyURL + "/health")
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	defer resp.Body.Close()
	var h map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&h)
	if h["status"] != "ok" {
		t.Errorf("health status: %v", h["status"])
	}
}

func TestStatsQPSNonNegative(t *testing.T) {
	_, proxyURL := startTestProxy(t)
	id := "node-1"
	registerAt(t, proxyURL, id, startTestNode(t, id))
	adminOp(t, http.MethodPost, proxyURL, id)
	putKey(t, proxyURL, "user-1", "v")
	time.Sleep(100 * time.Millisecond)
	st1 := fetchStats(t, proxyURL)
	putKey(t, proxyURL, "user-2", "v")
	time.Sleep(100 * time.Millisecond)
	st2 := fetchStats(t, proxyURL)
	if st2.TotalRequests < st1.TotalRequests {
		t.Errorf("requests decreased: %d -> %d", st1.TotalRequests, st2.TotalRequests)
	}
	for _, n := range st2.Nodes {
		if n.QPS < 0 {
			t.Errorf("node %s negative qps: %.2f", n.ID, n.QPS)
		}
	}
}