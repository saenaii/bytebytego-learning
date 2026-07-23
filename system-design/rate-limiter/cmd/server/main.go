package main

import (
	"flag"
	"log"
	"os"

	"github.com/bytebytego-learning/rate-limiter/internal/config"
	"github.com/bytebytego-learning/rate-limiter/internal/limiter"
	"github.com/bytebytego-learning/rate-limiter/internal/server"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	port := flag.Int("port", 0, "Port to listen on (overrides config)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Determine port: CLI flag overrides config
	serverPort := cfg.Server.Port
	if *port > 0 {
		serverPort = *port
	}
	if serverPort == 0 {
		serverPort = 8080
	}

	// Create rate limiter from config
	rateLimiter, err := limiter.NewRateLimiter(cfg)
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}

	log.Printf("Using rate limiter: %s", rateLimiter.Name())

	// Create and start the server
	srv := server.NewServer(rateLimiter, serverPort)

	if err := srv.Start(); err != nil {
		log.Printf("Server stopped: %v", err)
		os.Exit(1)
	}
}
