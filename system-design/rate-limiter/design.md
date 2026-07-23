# Rate Limiter — Local Design Document

> **Design a local rate-limiter service with three configurable algorithms, containerized via Docker, and fully covered by unit tests and e2e tests.**

---

## 1. Overview

A lightweight, in-process rate limiter service written in Go. It exposes an HTTP API that clients call to check whether a request should be allowed or rejected based on the configured rate-limiting algorithm.

```
┌──────────────┐     HTTP (GET /allow)     ┌──────────────────────────────┐
│   Client(s)  │ ──────────────────────────▶│         Rate Limiter        │
│              │ ◀──────────────────────────│         HTTP Server         │
└──────────────┘    200 OK / 429 Too Many   │                              │
                                            │  ┌────────────────────────┐ │
                                            │  │   RateLimiter Factory  │ │
                                            │  └───────┬────────────────┘ │
                                            │          │                   │
                                            │  ┌───────┼───────┬─────────┐ │
                                            │  ▼       ▼       ▼         │ │
                                            │  TB      LB      BBR      │ │
                                            │  (token  (leaking (BBR    │ │
                                            │   bucket) bucket)  algo)  │ │
                                            │                              │
                                            │  config.yaml ──▶ Config      │
                                            └──────────────────────────────┘
```

- **Language:** Go 1.26 (latest stable)
- **Transport:** HTTP (standard `net/http`)
- **Container:** Multi-stage Docker build (golang:alpine → alpine)
- **Orchestration:** `docker-compose up` (single command)

---

## 2. Requirements

### 2.1 Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| FR1 | The service MUST support **3 rate-limiting algorithms**: Token Bucket, Leaking Bucket, BBR (Bottleneck Bandwidth and Round-trip propagation time) | P0 |
| FR2 | Each algorithm MUST be independently selectable via a configuration file (`config.yaml`) | P0 |
| FR3 | The service MUST expose an HTTP endpoint `GET /allow` that returns `200 OK` if the request is allowed, or `429 Too Many Requests` if it is rejected | P0 |
| FR4 | The service MUST expose a health-check endpoint `GET /health` returning `200 OK` | P1 |
| FR5 | The service MUST expose a stats endpoint `GET /stats` returning current algorithm state (e.g., available tokens, queue size, estimated bandwidth) | P1 |
| FR6 | All rate limiter implementations MUST be **thread-safe** (support concurrent `Allow()` calls) | P0 |
| FR7 | Configuration MUST be loaded from a YAML file at startup, defaulting to built-in defaults if the file is missing or partially specified | P1 |

### 2.2 Non-Functional Requirements

| ID | Requirement | Priority |
|----|-------------|----------|
| NFR1 | The code MUST be **readable** and **human-friendly** (clear naming, comments, no magic numbers) | P0 |
| NFR2 | All public functions and types MUST be covered by **unit tests** (goal: ≥ 80% coverage) | P0 |
| NFR3 | The service MUST pass an **end-to-end test** that sends a predefined sequence of requests and verifies the allowed/rejected counts match expected results | P0 |
| NFR4 | The service MUST start via a single command: `docker-compose up` | P0 |
| NFR5 | The Docker image MUST use a **multi-stage build**: golang:alpine for build → scratch/alpine for runtime | P0 |
| NFR6 | All code MUST reside under the `system-design/rate-limiter/` directory | P0 |

---

## 3. Architecture & Component Overview

```
system-design/rate-limiter/
├── cmd/
│   └── server/
│       └── main.go              # Entry point: parse config, build limiter, start HTTP server
├── internal/
│   ├── config/
│   │   ├── config.go            # Config struct, LoadConfig(path) function
│   │   └── config_test.go       # Unit tests for config loading
│   ├── limiter/
│   │   ├── interface.go         # RateLimiter interface definition
│   │   ├── token_bucket.go      # TokenBucket implementation
│   │   ├── token_bucket_test.go # Unit tests for TokenBucket
│   │   ├── leaking_bucket.go    # LeakingBucket implementation
│   │   ├── leaking_bucket_test.go
│   │   ├── bbr.go               # BBR implementation
│   │   ├── bbr_test.go
│   │   ├── factory.go           # Factory function: NewRateLimiter(config)
│   │   └── factory_test.go
│   └── server/
│       ├── server.go            # HTTP server: handlers, routes, middleware
│       └── server_test.go       # Unit tests using httptest
├── tests/
│   └── e2e/
│       └── e2e_test.go          # End-to-end tests (start server, send requests, verify)
├── config.yaml                  # Default configuration file
├── Dockerfile                   # Multi-stage Docker build
├── docker-compose.yml           # Docker Compose service definition
├── Makefile                     # Build, test, run targets
├── go.mod                       # Go module file
├── go.sum                       # Go module checksums
├── tasks/
│   ├── 2026-07-23.md            # Task breakdown for the day
│   └── tracker.md               # Task tracker with checkboxes
└── design.md                    # This document
```

### 3.1 Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| `cmd/server/main.go` | Parse CLI flags, load config, instantiate the limiter via factory, start the HTTP server |
| `internal/config/config.go` | Define `Config` struct; parse `config.yaml` using `gopkg.in/yaml.v3`; apply defaults for missing fields |
| `internal/limiter/interface.go` | Define the `RateLimiter` interface that all algorithms must implement |
| `internal/limiter/token_bucket.go` | Token bucket algorithm: maintain a fixed-capacity bucket that refills at a constant rate |
| `internal/limiter/leaking_bucket.go` | Leaking bucket algorithm: maintain a FIFO queue that leaks (processes) at a constant rate |
| `internal/limiter/bbr.go` | BBR algorithm: estimate bandwidth & RTT from recent samples; adaptively allow/reject |
| `internal/limiter/factory.go` | Factory function that returns the correct `RateLimiter` based on `LimiterType` in config |
| `internal/server/server.go` | HTTP server with routes; wraps `RateLimiter`; uses `sync.Mutex` for concurrent handler safety |
| `tests/e2e/e2e_test.go` | Integration test: build binary, start server, send predefined request sequences, assert results |

---

## 4. Algorithm Specifications

### 4.1 Token Bucket

**Concept:** A bucket holds up to `capacity` tokens. Each `Allow()` call consumes 1 token if available. Tokens are added back at a fixed `refillRate` per `refillInterval`. If the bucket is empty, the request is denied.

**Formulas:**

```
tokens = min(capacity, tokens + (elapsedTime / refillInterval) * refillRate)
if tokens >= 1:
    tokens -= 1
    return true (allowed)
else:
    return false (rejected)
```

**Data Structure:**

```go
type TokenBucket struct {
    mu             sync.Mutex
    capacity       float64       // Maximum number of tokens
    tokens         float64       // Current token count
    refillRate     float64       // Number of tokens added per refillInterval
    refillInterval time.Duration // Time between refills
    lastRefillTime time.Time     // Last time tokens were refilled
}
```

**Key Properties:**
- Burst: Up to `capacity` requests can pass instantaneously
- Steady-state: Average rate = `refillRate / refillInterval` per second
- Thread-safe via `sync.Mutex`

### 4.2 Leaking Bucket

**Concept:** A FIFO queue of fixed size `queueSize`. Incoming requests are added to the queue (if space permits). A background goroutine "leaks" (processes) items from the queue at a constant `leakRate` per second. If the queue is full, the request is rejected.

**Formulas:**

```
// On each Allow():
if queue length < queueSize:
    enqueue request
    return true (allowed)
else:
    return false (rejected)

// Background goroutine (runs every leakInterval):
dequeue 1 item and process it
```

**Data Structure:**

```go
type LeakingBucket struct {
    mu            sync.Mutex
    queueSize     int           // Maximum queue capacity
    queue         []time.Time   // FIFO queue of request timestamps
    leakRate      float64       // Number of items leaked per second
    leakInterval  time.Duration // Time between leaks (ceil(1s / leakRate) or 1ms if rate > 1000)
    lastLeakTime  time.Time     // Last time an item was leaked
    stopCh        chan struct{} // Signal to stop background goroutine
}
```

**Key Properties:**
- Smooth output: Requests are processed at a constant rate regardless of input burstiness
- Burst absorption: Short bursts are buffered in the queue
- Thread-safe via `sync.Mutex`

### 4.3 BBR (Bottleneck Bandwidth and Round-trip propagation time)

**Concept:** BBR estimates the available bandwidth (`bw`) and round-trip time (`rtt`) from recent request samples. It maintains a sliding window of `windowSize` samples. The allowed rate is calculated as:

```
estimatedBandwidth = min over sliding window of (bytes_delivered / delivery_rate)
minRTT = min over sliding window of (sample_rtt)
BDP = estimatedBandwidth * minRTT  // Bandwidth-Delay Product
allowedRate = gain * estimatedBandwidth
```

**Simplified Implementation:** For this local rate limiter, BBR monitors the inter-arrival time of requests and the processing time to dynamically adjust an internal "allow probability." When the system is underutilized (low RTT, high bandwidth), it allows more requests; when congested (high RTT, low bandwidth), it reduces the allow rate.

**Data Structure:**

```go
type BBR struct {
    mu               sync.Mutex
    maxBandwidth     float64       // Estimated max bandwidth (requests per second)
    minRTT           time.Duration // Estimated minimum RTT
    gain             float64       // Multiplicative gain factor (default: 2.0)
    windowSize       int           // Number of samples in the sliding window
    samples          []sample      // Ring buffer of recent samples
    nextSampleIndex  int           // Next write position in ring buffer
    allowed          int64         // Total allowed requests
    rejected         int64         // Total rejected requests
    allowProbability float64       // Current probability of allowing a request
}

type sample struct {
    bandwidth float64
    rtt       time.Duration
}
```

**Key Properties:**
- Adaptive: Adjusts dynamically to changing network/load conditions
- Self-tuning: No manual capacity configuration needed
- Thread-safe via `sync.Mutex`

---

## 5. Interface Definition

```go
// internal/limiter/interface.go

// RateLimiter defines the contract that all rate limiter implementations must satisfy.
type RateLimiter interface {
    // Allow checks if a request is allowed at the current time.
    // Returns true if the request should be processed, false if it should be rejected.
    Allow() bool

    // AllowAt checks if a request would be allowed at the specified time.
    // Useful for testing and simulation.
    AllowAt(t time.Time) bool

    // Name returns the name of the rate limiter algorithm.
    Name() string

    // Stats returns current statistics about the rate limiter state.
    Stats() map[string]interface{}
}
```

---

## 6. Configuration Specification

### 6.1 Config File Format (`config.yaml`)

```yaml
# Rate Limiter Configuration
# Select one of: "token_bucket", "leaking_bucket", "bbr"
limiter_type: "token_bucket"

# Server settings
server:
  port: 8080

# Token Bucket settings (used when limiter_type = "token_bucket")
token_bucket:
  capacity: 10          # Maximum number of tokens in the bucket
  refill_rate: 1        # Number of tokens added per refill_interval
  refill_interval: 1s   # Time between refills (e.g., 100ms, 1s, 1m)

# Leaking Bucket settings (used when limiter_type = "leaking_bucket")
leaking_bucket:
  queue_size: 10        # Maximum number of requests in the queue
  leak_rate: 1          # Number of requests processed per second

# BBR settings (used when limiter_type = "bbr")
bbr:
  gain: 2.0             # Multiplicative gain factor (> 1.0)
  window_size: 10       # Number of samples in the sliding window
```

### 6.2 Config Struct (Go)

```go
// internal/config/config.go

type Config struct {
    LimiterType   string         `yaml:"limiter_type"`
    Server        ServerConfig   `yaml:"server"`
    TokenBucket   *TokenBucketConfig `yaml:"token_bucket,omitempty"`
    LeakingBucket *LeakingBucketConfig `yaml:"leaking_bucket,omitempty"`
    BBR           *BBRConfig     `yaml:"bbr,omitempty"`
}

type ServerConfig struct {
    Port int `yaml:"port"`
}

type TokenBucketConfig struct {
    Capacity       int           `yaml:"capacity"`
    RefillRate     float64       `yaml:"refill_rate"`
    RefillInterval time.Duration `yaml:"refill_interval"`
}

type LeakingBucketConfig struct {
    QueueSize int     `yaml:"queue_size"`
    LeakRate  float64 `yaml:"leak_rate"`
}

type BBRConfig struct {
    Gain       float64 `yaml:"gain"`
    WindowSize int     `yaml:"window_size"`
}
```

### 6.3 Default Values

| Parameter | Default |
|-----------|---------|
| `server.port` | 8080 |
| `token_bucket.capacity` | 10 |
| `token_bucket.refill_rate` | 1.0 |
| `token_bucket.refill_interval` | 1s |
| `leaking_bucket.queue_size` | 10 |
| `leaking_bucket.leak_rate` | 1.0 |
| `bbr.gain` | 2.0 |
| `bbr.window_size` | 10 |

---

## 7. API Specification

### 7.1 `GET /health`

**Purpose:** Health check / liveness probe.

**Response:**
```json
{
  "status": "ok",
  "limiter_type": "token_bucket"
}
```

- Status Code: `200 OK`

### 7.2 `GET /allow`

**Purpose:** Check if a request is allowed according to the configured rate limiter.

**Response (Allowed):**
```json
{
  "allowed": true,
  "limiter_type": "token_bucket"
}
```
- Status Code: `200 OK`

**Response (Rejected):**
```json
{
  "allowed": false,
  "limiter_type": "token_bucket"
}
```
- Status Code: `429 Too Many Requests`

### 7.3 `GET /stats`

**Purpose:** Return current state of the rate limiter (useful for monitoring and debugging).

**Response (Token Bucket):**
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

**Response (Leaking Bucket):**
```json
{
  "limiter_type": "leaking_bucket",
  "queue_size": 10,
  "queue_length": 3,
  "leak_rate": 1.0,
  "allowed": 42,
  "rejected": 5
}
```

**Response (BBR):**
```json
{
  "limiter_type": "bbr",
  "max_bandwidth": 100.0,
  "min_rtt": "10ms",
  "gain": 2.0,
  "window_size": 10,
  "allow_probability": 0.85,
  "allowed": 42,
  "rejected": 5
}
```

- Status Code: `200 OK`

---

## 8. Testing Strategy

### 8.1 Unit Tests (per package)

| Package | Test File | Test Cases |
|---------|-----------|------------|
| `config` | `config_test.go` | Load valid YAML, load missing file, partial config with defaults, invalid YAML |
| `limiter/token_bucket` | `token_bucket_test.go` | Basic allow, basic deny (empty bucket), refill over time, burst capacity, concurrent access, edge cases (zero capacity, negative refill) |
| `limiter/leaking_bucket` | `leaking_bucket_test.go` | Basic allow (queue not full), queue full rejection, leak over time, concurrent access, edge cases (zero queue, negative leak rate) |
| `limiter/bbr` | `bbr_test.go` | Initial state, allow probability adjustment, adaptation to load, concurrent access, edge cases (zero gain, negative window) |
| `limiter/factory` | `factory_test.go` | Create token bucket, create leaking bucket, create BBR, unknown type error |
| `server` | `server_test.go` | Health endpoint, allow endpoint (allowed), allow endpoint (rejected), stats endpoint, concurrent requests |

### 8.2 End-to-End Tests

**Test Scenario 1: Token Bucket (capacity=5, refill_rate=1, refill_interval=1s)**

```
Time 0ms:  Send 5 requests → expect 5 allowed, 0 rejected
Time 0ms:  Send 1 more request → expect 0 allowed, 1 rejected
Time 1000ms: Send 1 request → expect 1 allowed, 0 rejected (1 token refilled)
Time 1000ms: Send 1 request → expect 0 allowed, 1 rejected (bucket empty again)
```

**Test Scenario 2: Leaking Bucket (queue_size=3, leak_rate=1)**

```
Time 0ms:  Send 3 requests → expect 3 allowed, 0 rejected
Time 0ms:  Send 1 request → expect 0 allowed, 1 rejected (queue full)
Time 1000ms: Send 1 request → expect 1 allowed, 0 rejected (1 item leaked, slot freed)
```

**Test Scenario 3: BBR (gain=2.0, window_size=5)**

```
Send 50 requests as fast as possible → verify that at most (gain * some baseline) are allowed
Send 10 requests with 100ms spacing → verify most are allowed (low load)
```

### 8.3 Coverage Requirement

- All unit tests must pass: `go test ./... -v -race -coverprofile=coverage.out`
- Coverage must be ≥ 80% on the `internal/` packages
- E2E tests must pass: `go test ./tests/e2e/... -v -count=1`

---

## 9. Error Handling & Edge Cases

| Condition | Behavior |
|-----------|----------|
| Config file not found | Use built-in defaults; log a warning |
| Invalid YAML syntax | Return error and exit |
| Unknown `limiter_type` | Return error from factory; program exits with non-zero code |
| Negative capacity/queue_size | Treat as zero; log a warning |
| Zero refill_rate/leak_rate | Allow no requests (effectively disabled); log a warning |
| Concurrent `Allow()` calls | Thread-safe via `sync.Mutex`; no race conditions |
| Time travel (system clock goes backward) | Use `max(0, elapsed)` to prevent negative token counts |
| Overflow protection | Cap token count at `capacity`; cap queue length at `queueSize` |
| Port already in use | Log error; exit |

---

## 10. Build & Deployment

### 10.1 Makefile Targets

| Target | Description | Command |
|--------|-------------|---------|
| `build` | Build the Go binary for current OS/arch | `go build -o bin/server ./cmd/server` |
| `test` | Run all unit tests with race detection | `go test ./... -v -race -count=1` |
| `test-e2e` | Run end-to-end tests | `go test ./tests/e2e/... -v -count=1` |
| `coverage` | Run tests and show coverage | `go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html` |
| `run` | Build and run locally | `go run ./cmd/server/ -config=config.yaml` |
| `docker-build` | Build Docker image | `docker build -t rate-limiter:latest .` |
| `docker-run` | Run Docker container | `docker run -p 8080:8080 -v $(PWD)/config.yaml:/app/config.yaml rate-limiter:latest` |
| `clean` | Remove build artifacts | `rm -rf bin/ coverage.out coverage.html` |

### 10.2 Dockerfile (Multi-Stage)

```dockerfile
# Stage 1: Build
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

### 10.3 docker-compose.yml

```yaml
version: "3.9"
services:
  rate-limiter:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml
    restart: unless-stopped
```

### 10.4 Startup Commands

```bash
# Start the service
docker-compose up

# Build and run locally (without Docker)
make run

# Run all tests
make test

# Run e2e tests (requires the service to be running)
make test-e2e
```

---

## 11. Development Workflow

1. **Understand this design document** — read through all sections carefully.
2. **Generate task breakdowns** — create task files in `./tasks/` with time-based names (e.g., `2026-07-23-scaffolding.md`).
3. **Track progress** — update `./tasks/tracker.md` with checkboxes as tasks are completed.
4. **Implement one component at a time** — start from the bottom of the dependency chain (config → interface → algorithms → factory → server → e2e tests → Docker).
5. **Write tests first or alongside** — every public function must have corresponding unit tests.
6. **Run the full test suite before marking a task complete.**
7. **Focus only on pending tasks** — do not rework completed tasks unless bugs are found.

---

## 12. Glossary

| Term | Definition |
|------|------------|
| **Token Bucket** | A rate-limiting algorithm that maintains a bucket of tokens that refill at a constant rate. Each request consumes one token. |
| **Leaking Bucket** | A rate-limiting algorithm that uses a FIFO queue and processes items at a constant rate. Bursts are buffered. |
| **BBR** | Bottleneck Bandwidth and Round-trip propagation time. An adaptive congestion-control-based rate limiter. |
| **BDP** | Bandwidth-Delay Product. The amount of data that can be in flight in a network path. |
| **RTT** | Round-Trip Time. The time it takes for a packet to travel from source to destination and back. |
| **Thread-safe** | Code that can be safely called from multiple goroutines concurrently without race conditions. |
| **E2E Test** | End-to-end test that exercises the full stack (HTTP server, rate limiter, config) to verify behavior. |
| **Multi-stage build** | A Docker build process that uses one stage for compilation and a separate minimal stage for the runtime image. |