package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytebytego-learning/consistent-hashing/internal/generator"
)

func main() {
	proxyURL := flag.String("proxy", "http://localhost:8080", "proxy base URL")
	rate := flag.Int("rate", 100, "requests per second")
	keyspace := flag.Int("keyspace", 10000, "distinct key space")
	putRatio := flag.Float64("put-ratio", 0.4, "fraction of PUT operations")
	duration := flag.Duration("duration", 0, "run duration (0 = run until interrupted)")
	flag.Parse()

	cfg := generator.Config{
		ProxyURL:  *proxyURL,
		Rate:      *rate,
		Keyspace:  *keyspace,
		PutRatio:  *putRatio,
		Duration:  *duration,
	}
	g := generator.New(cfg)
	g.Start()

	log.Printf("traffic generator started: %d req/s over %d keys (proxy=%s)", *rate, *keyspace, *proxyURL)

	var all []generator.Result
	interim := time.NewTicker(5 * time.Second)
	defer interim.Stop()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	var stop <-chan time.Time
	if *duration > 0 {
		t := time.NewTimer(*duration)
		defer t.Stop()
		stop = t.C
	}

loop:
	for {
		select {
		case r, ok := <-g.Results():
			if !ok {
				break loop
			}
			all = append(all, r)
		case <-interim.C:
			s := generator.Summarize(all)
			fmt.Printf("[%s] ops=%d success=%.2f%% put-p50=%v get-p50=%v\n",
				time.Now().Format("15:04:05"), s.Total, s.SuccessRate, s.Put.P50, s.Get.P50)
		case <-stop:
			break loop
		case <-sig:
			break loop
		}
	}

	g.Stop()
	for r := range g.Results() {
		all = append(all, r)
	}

	s := generator.Summarize(all)
	fmt.Printf("\n=== Traffic generator summary ===\n")
	fmt.Printf("Total ops:  %d\n", s.Total)
	fmt.Printf("Success:    %d (%.2f%%)\n", s.Success, s.SuccessRate)
	fmt.Printf("Failed:     %d\n", s.Failure)
	fmt.Printf("Puts:       %d (avg %v, p50 %v, p95 %v, p99 %v)\n", s.Put.Count, s.Put.Avg, s.Put.P50, s.Put.P95, s.Put.P99)
	fmt.Printf("Gets:       %d (avg %v, p50 %v, p95 %v, p99 %v)\n", s.Get.Count, s.Get.Avg, s.Get.P50, s.Get.P95, s.Get.P99)
}
