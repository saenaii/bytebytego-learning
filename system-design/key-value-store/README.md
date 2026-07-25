# Distributed Key-Value Store

A distributed key-value store simulated via Docker containers. Implements consistent hashing (50 virtual nodes), vector clocks, and Merkle trees for conflict resolution with quorum-based reads/writes (N=3, W=2, R=2).

## Architecture

```
client -> proxy (coordinator) -> cluster (node-1, node-2, node-3)
```

- **Proxy**: Routes requests via consistent hashing, quorum reads/writes, sibling return on concurrent conflicts
- **Nodes**: In-memory KV stores with vector clock versioning + periodic Merkle tree anti-entropy
- **Conflict Resolution**: Vector clocks on read (returns siblings), Merkle tree anti-entropy in background

## Project Structure

```
system-design/key-value-store/
├── cmd/{proxy,node}/main.go   # Entry points
├── internal/
│   ├── consistent/            # Consistent hash ring (50 vnodes)
│   ├── vectorclock/           # Vector clock with Compare/Merge
│   ├── merkle/                # Merkle tree for anti-entropy
│   ├── store/                 # In-memory store with siblings support
│   ├── server/                # HTTP servers (proxy + node)
│   └── client/                # HTTP client
├── tests/e2e/                 # E2E + chaos tests
├── docker-compose.yml          # 4 containers (proxy + 3 nodes)
├── Dockerfile
├── REPORT.md                   # Auto-generated performance report
├── plan.md                     # Implementation plan
└── design.md                   # Design document
```

## Quick Start

```bash
# Run locally
go run ./cmd/proxy/            # Terminal 1
go run ./cmd/node/ -addr=:8081 -id=node-1  # Terminal 2
go run ./cmd/node/ -addr=:8082 -id=node-2  # Terminal 3
go run ./cmd/node/ -addr=:8083 -id=node-3  # Terminal 4

# Run with Docker
docker compose up -d

# Tests
make test       # Unit tests
make test-e2e   # E2E tests (generates REPORT.md)
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/kv/{key}` | Retrieve value (returns siblings if concurrent) |
| PUT | `/kv/{key}` | Store value (body: `{"value":"..."}`) |
| GET | `/health` | Health check |
| GET | `/status` | Cluster status |
| POST | `/admin/node` | Add/remove nodes |

## Features

- **Consistent Hashing**: 50 virtual nodes per physical node for even distribution
- **Quorum**: N=3, W=2, R=2 for fault-tolerant reads/writes
- **Vector Clocks**: Per-node version tracking, returns siblings on concurrent conflicts
- **Merkle Tree Anti-Entropy**: Background goroutine syncs replicas periodically
- **Dynamic Membership**: Add/remove nodes up to max 5
- **Chaos Testing**: Network partition simulation with eventual consistency healing
- **Performance Report**: Auto-generated `REPORT.md` with latency breakdown
