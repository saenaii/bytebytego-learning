package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/bytebytego-learning/key-value-store/internal/server"
)

func main() {
	addr := flag.String("addr", ":8081", "node listen address")
	id := flag.String("id", "node-1", "node identifier")
	proxyURL := flag.String("proxy", "http://localhost:8080", "proxy URL")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("node listen failed: %v", err)
	}

	node := server.NewNodeServer(*id)
	go registerWithProxyLoop(*proxyURL, *id, ln.Addr().String())
	log.Fatal(node.Serve(ln))
}

func registerWithProxyLoop(proxyURL, nodeID, nodeAddr string) {
	for {
		if err := registerWithProxy(proxyURL, nodeID, nodeAddr); err != nil {
			log.Printf("Failed to register with proxy: %v, retrying in 3s", err)
			time.Sleep(3 * time.Second)
			continue
		}
		return
	}
}

func registerWithProxy(proxyURL, nodeID, nodeAddr string) error {
	body := map[string]string{
		"action":    "add",
		"node_id":   nodeID,
		"node_addr": nodeAddr,
	}
	data, _ := json.Marshal(body)
	resp, err := http.Post(proxyURL+"/admin/node", "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("proxy returned %d", resp.StatusCode)
	}
	log.Printf("Registered with proxy: %s as %s", proxyURL, nodeID)
	return nil
}
