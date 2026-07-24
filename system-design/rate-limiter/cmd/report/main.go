package main

import (
	"log"

	"github.com/bytebytego-learning/rate-limiter/internal/report"
)

func main() {
	if err := report.Generate("."); err != nil {
		log.Fatalf("Failed to generate report: %v", err)
	}
	log.Println("Report generated: REPORT.md")
}
