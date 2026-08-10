package main

import (
	"flag"
	"log"
	"net"

	"github.com/bytebytego-learning/consistent-hashing/internal/proxy"
)

func main() {
	addr := flag.String("addr", ":8080", "proxy listen address")
	vnodes := flag.Int("vnodes", 50, "virtual nodes per physical node")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("proxy listen failed: %v", err)
	}

	p := proxy.New(proxy.Config{VNodes: *vnodes})
	log.Fatal(p.Serve(ln))
}
