# Consistent Hashing

A runnable local demo of **Consistent Hashing with virtual nodes** — proves that dynamic scale-out / scale-in redistributes traffic evenly while migrating the minimum possible number of keys, unlike naive `key % N` hashing.

## Key Takeaways (from ByteByteGo)

| Concept | What We Learned |
|---------|----------------|
| **The `key % N` problem** | When a node is added or removed, almost every key maps to a different bucket. In a 5-node cluster, adding a 6th node forces ~83% of keys to migrate — catastrophic for caches and session stores. |
| **Consistent hashing ring** | Keys and nodes are hashed onto a ring (0 → 2^64). A key is assigned to the first node found clockwise. On node removal, only that node's keys migrate — everything else stays put. |
| **Virtual nodes (50 per physical)** | With few physical nodes, plain hashing concentrates keys on the smallest arc. Adding 50 virtual nodes per physical node smooths the ring: load standard deviation stays within ±15% of the mean. |
| **Scale-out key migration** | Adding a node to a 3-node cluster moves only ~25% of keys (vs ~75% with naive rehash). Only the arcs adjacent to the new node's vnodes are affected. |
| **Binary search O(log K)** | The sorted vnode ring enables O(log K) lookup via binary search. With 200 vnodes (4 physical × 50), each lookup touches ~8 comparisons — negligible overhead. |
| **Standby node pattern** | Nodes boot healthy but stay unregistered. The admin API toggles them in/out of the ring, enabling live interactive demos without restarting containers. |

> **Bottom line:** Consistent hashing is the backbone of distributed caches (Memcached, DynamoDB), CDNs, and sharded databases. Virtual nodes turn a theoretically sound idea into a practically balanced one — the ring must be "thick" enough to absorb topology changes gracefully.

## Architecture

```
client -> proxy (hash ring + dashboard) -> worker nodes (node-1..5)
```

- **Proxy**: Owns the hash ring, routes `/kv/{key}` via `ring.Get(key)`, serves embedded dashboard, exposes admin + metrics APIs
- **Worker Nodes**: In-memory KV stores with simulated 100ms–5s latency; report metrics via `/stats`
- **Traffic Generator**: Continuous random-key traffic at 100 req/s over a 10k key space
- **Dashboard**: Embedded HTML page served by the proxy (`go:embed`), vanilla JS + Chart.js

## Project Structure

```
system-design/consistent-hashing/
├── cmd/
│   ├── proxy/           # Proxy entry point (ring + dashboard + admin)
│   └── traffic/         # Traffic generator
├── internal/
│   ├── ring/            # Consistent hash ring (50 vnodes, binary search)
│   ├── node/            # Worker node HTTP server
│   └── dashboard/       # Embedded HTML dashboard (go:embed)
├── tests/
│   └── e2e/             # End-to-end + chaos tests
├── docker-compose.yml   # 5 nodes (2 standby for interactive demo)
├── Dockerfile
├── Makefile
├── design.md            # Design document
└── REPORT.md            # Auto-generated performance report
```

## Quick Start

```bash
# Build and run
make build && make run

# Or with Docker
docker compose up

# Open dashboard
open http://localhost:8080

# Tests
make test        # Unit tests (table-driven, -race)
make test-e2e    # E2E tests (generates REPORT.md)
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Embedded HTML dashboard |
| PUT | `/kv/{key}` | Store value, routed via consistent hashing |
| GET | `/kv/{key}` | Fetch value from responsible node |
| POST | `/admin/nodes/{id}` | Add a standby node to the ring |
| DELETE | `/admin/nodes/{id}` | Remove a node from the ring |
| GET | `/stats` | Per-node QPS, key count, load stddev |

## Metrics

| Metric | Target |
|--------|--------|
| Key distribution stddev | ≤ 15% of mean |
| Scale-out key migration | ~25% (3→4 nodes) |
| Lookup latency | O(log K) via binary search |
| E2E success rate | ≥ 99% |

## License

This project is for educational purposes.
