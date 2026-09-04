# Rate Limiter

A lightweight, in-process rate limiter service written in Go. It exposes an HTTP API that clients call to check whether a request should be allowed or rejected based on the configured rate-limiting algorithm.

## Key Takeaways (from ByteByteGo)

| Concept | What We Learned |
|---------|----------------|
| **Why rate limiting?** | Protect backend services from traffic spikes, prevent abuse, and ensure fair resource sharing across clients. Without it, a single noisy neighbor can bring down the entire system. |
| **Token Bucket** | Best when bursts are acceptable. A bucket holds up to `capacity` tokens that refill at a steady rate. Simple, memory-efficient, and the most widely used algorithm (e.g., AWS API Gateway). |
| **Leaking Bucket** | Best when output rate must be perfectly smooth. Requests queue into a FIFO; a background goroutine drains at a constant rate. Trade-off: added latency under burst. |
| **BBR (Adaptive)** | Best when traffic patterns are unpredictable. Estimates bandwidth and RTT from a sliding window, self-tunes the allow-rate. No manual capacity tuning required — the system adapts to its own throughput. |
| **Interface-driven design** | All three algorithms satisfy the same `RateLimiter` interface (`Allow() bool`, `Stats()`). Swapping algorithms is a one-line config change — callers never change. |
| **Thread-safety is non-negotiable** | Any in-process limiter must handle concurrent `Allow()` calls safely (`sync.Mutex`). A race condition in the limiter silently under- or over-counts, defeating its purpose. |

> **Bottom line:** There is no single "best" rate limiter. Token Bucket handles bursts, Leaking Bucket smooths them, and BBR adapts to unknown conditions. The right choice depends on whether your priority is burst tolerance, output smoothness, or self-tuning under changing load.

## Features

- **Three configurable rate-limiting algorithms:**
  - **Token Bucket** - Maintains a bucket of tokens that refill at a constant rate
  - **Leaking Bucket** - Uses a FIFO queue that processes items at a constant rate
  - **BBR (Bottleneck Bandwidth and Round-trip propagation time)** - Adaptive congestion-control-based rate limiter
- **Thread-safe** - All implementations support concurrent `Allow()` calls
- **HTTP API** - Simple RESTful endpoints for rate limiting, health checks, and stats
- **Docker support** - Multi-stage Docker build with docker-compose orchestration
- **Comprehensive testing** - Unit tests and end-to-end tests with ≥80% coverage

## Project Structure

```
system-design/rate-limiter/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go            # Configuration loading
│   │   └── config_test.go
│   ├── limiter/
│   │   ├── interface.go         # RateLimiter interface
│   │   ├── token_bucket.go      # Token Bucket implementation
│   │   ├── leaking_bucket.go    # Leaking Bucket implementation
│   │   ├── bbr.go               # BBR implementation
│   │   └── factory.go           # Factory function
│   └── server/
│       ├── server.go            # HTTP server and handlers
│       └── server_test.go
├── tests/
│   └── e2e/
│       └── e2e_test.go          # End-to-end tests
├── config.yaml                  # Default configuration
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Docker Compose definition
├── Makefile                     # Build and test targets
└── design.md                    # Design document
```

## Prerequisites

- Go 1.26 or later
- Docker and Docker Compose (for containerized deployment)
- Make (optional, for using Makefile targets)

## Quick Start

### Build and Run Locally

```bash
# Build the binary
make build

# Run the server
make run
```

### Run with Docker

```bash
# Start the service with docker-compose
docker-compose up
```

The server will be available at `http://localhost:8080`.

## Configuration

The service is configured via `config.yaml`. Example configuration:

```yaml
# Select one of: "token_bucket", "leaking_bucket", "bbr"
limiter_type: "token_bucket"

server:
  port: 8080

# Token Bucket settings
token_bucket:
  capacity: 10
  refill_rate: 1
  refill_interval: 1s

# Leaking Bucket settings
leaking_bucket:
  queue_size: 10
  leak_rate: 1

# BBR settings
bbr:
  gain: 2.0
  window_size: 10
```

## API Endpoints

### `GET /health`

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "limiter_type": "token_bucket"
}
```

### `GET /allow`

Check if a request is allowed.

**Response (Allowed - 200 OK):**
```json
{
  "allowed": true,
  "limiter_type": "token_bucket"
}
```

**Response (Rejected - 429 Too Many Requests):**
```json
{
  "allowed": false,
  "limiter_type": "token_bucket"
}
```

### `GET /stats`

Get current rate limiter statistics.

**Response (Token Bucket example):**
```json
{
  "limiter_type": "token_bucket",
  "tokens": 7.5,
  "capacity": 10,
  "refill_rate": 1.0,
  "refill_interval": "1s",
  "allowed": 42,
  "rejected": 5
}
```

## Testing

```bash
# Run all unit tests with race detection
make test

# Run end-to-end tests
make test-e2e

# Generate coverage report
make coverage
```

## Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build the Go binary |
| `test` | Run all unit tests with race detection |
| `test-e2e` | Run end-to-end tests |
| `coverage` | Generate coverage report |
| `run` | Build and run locally |
| `docker-build` | Build Docker image |
| `docker-run` | Run Docker container |
| `clean` | Remove build artifacts |

## Algorithms

### Token Bucket

A bucket holds up to `capacity` tokens. Each request consumes 1 token if available. Tokens are added back at a fixed rate. Supports burst traffic up to the bucket capacity.

### Leaking Bucket

A FIFO queue of fixed size. Incoming requests are added to the queue if space permits. A background goroutine processes items at a constant rate. Provides smooth output rate regardless of input burstiness.

### BBR

Adaptively estimates available bandwidth and round-trip time from recent request samples. Adjusts the allow rate dynamically based on network conditions. Self-tuning with no manual capacity configuration needed.

## License

This project is for educational purposes.