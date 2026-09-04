# Distributed Key-Value Store

A distributed key-value store simulated via Docker containers. Implements consistent hashing (50 virtual nodes), vector clocks, and Merkle trees for conflict resolution with quorum-based reads/writes (N=3, W=2, R=2).

## Key Takeaways (from ByteByteGo)

| Concept | What We Learned |
|---------|----------------|
| **Consistent hashing for data placement** | Keys are mapped onto a hash ring; each key lives on the next N physical nodes clockwise. On node add/remove, only the affected arcs move — the rest of the cluster is untouched. Virtual nodes (50 per physical) keep distribution within ±15% stddev. |
| **Quorum (N=3, W=2, R=2)** | With replication factor 3, writing to 2 and reading from 2 guarantees reading the latest write. W+R > N ensures overlap. Tuning W and R lets you trade consistency for latency: W=1 is fast but stale reads are possible. |
| **Vector clocks detect concurrency** | Each write increments the vector clock. On conflict (two concurrent writes with no causal dominance), the store returns **siblings** instead of silently overwriting. The client (or application) resolves the conflict — the database stays honest. |
| **Merkle trees for anti-entropy** | Each node maintains a Merkle tree of its keys. Background repair compares trees between replicas; only differing subtrees are transferred. This makes sync O(log N) in data size, not O(N). |
| **Eventual consistency is a design choice** | When the network heals, anti-entropy merges all replicas. The system is temporarily inconsistent during partitions, but converges to a single value. This is the CAP theorem in action: choose AP (availability + partition tolerance) and accept eventual consistency. |
| **Proxy as coordinator** | The proxy owns the ring, enforces quorum, and returns siblings on conflict. Separating routing from storage keeps nodes stateless and simple — each node only needs a local KV store and a Merkle tree. |

> **Bottom line:** A distributed KV store is a composition of well-known building blocks: consistent hashing for placement, quorum for consistency tuning, vector clocks for conflict detection, and Merkle trees for background repair. No single component is complex — the challenge is making them work together correctly under failure.

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
