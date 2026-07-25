package main

import (
	"flag"
	"log"
	"net"

	"github.com/bytebytego-learning/key-value-store/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	virtualNodes := flag.Int("vnodes", 50, "virtual nodes per physical node")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("proxy listen failed: %v", err)
	}

	proxy := server.NewProxyServer(*virtualNodes, server.Quorum{N: 3, W: 2, R: 2})
	log.Fatal(proxy.Serve(ln))
}
