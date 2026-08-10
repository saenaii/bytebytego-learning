package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytebytego-learning/consistent-hashing/internal/node"
)

func main() {
	addr := flag.String("addr", ":8081", "node listen address")
	id := flag.String("id", "node-1", "node identifier")
	proxyURL := flag.String("proxy", "http://localhost:8080", "proxy base URL")
	publicAddr := flag.String("public-addr", "", "address advertised to the proxy (defaults to the listen address)")
	register := flag.Bool("register", true, "join the ring on boot (false = healthy standby)")
	latency := flag.Duration("latency", 100*time.Millisecond, "base simulated processing latency")
	jitter := flag.Duration("jitter", 4900*time.Millisecond, "max additional random latency")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("node %s listen failed: %v", *id, err)
	}

	srv := node.New(*id, node.WithLatency(*latency, *jitter))
	serverDone := make(chan error, 1)
	go func() { serverDone <- srv.Serve(ln) }()

	advertised := *publicAddr
	if advertised == "" {
		advertised = "http://" + ln.Addr().String()
	}

	client := &http.Client{Timeout: 5 * time.Second}
	if !registerWithProxy(client, *proxyURL, *id, advertised) {
		log.Fatalf("node %s: could not reach proxy at %s", *id, *proxyURL)
	}
	if *register && !enableOnRing(client, *proxyURL, *id) {
		log.Fatalf("node %s: could not join the ring", *id)
	}
	log.Printf("node %s ready at %s (register=%v, latency=%s..%s)", *id, advertised, *register, *latency, *latency+*jitter)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serverDone:
		log.Fatalf("node %s server exited: %v", *id, err)
	case <-sig:
		log.Printf("node %s shutting down", *id)
	}
}

func registerWithProxy(client *http.Client, proxyURL, id, addr string) bool {
	body, _ := json.Marshal(map[string]string{"id": id, "addr": addr})
	for i := 0; i < 30; i++ {
		resp, err := client.Post(proxyURL+"/register", "application/json", bytes.NewReader(body))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func enableOnRing(client *http.Client, proxyURL, id string) bool {
	for i := 0; i < 30; i++ {
		req, err := http.NewRequest(http.MethodPost, proxyURL+"/admin/nodes/"+id, nil)
		if err != nil {
			return false
		}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
				return true
			}
		}
		time.Sleep(2 * time.Second)
	}
	return false
}
