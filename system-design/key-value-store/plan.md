# Plan: Distributed Key-Value Store

## Architecture

```
client -> proxy (coordinator) -> cluster (node-1, node-2, ...)
```

- **Proxy**: Single entry point for clients. Uses consistent hashing to route requests to the correct node(s). Handles read repair and conflict resolution.
- **Node**: Each node is a Docker container running an HTTP server. Stores key-value pairs in memory with vector clocks for conflict tracking.
- **Consistent Hashing**: Ring-based with virtual nodes for even key distribution. Supports dynamic add/remove of nodes.

## Components

| Package | Responsibility |
|---|---|
| `internal/consistent` | Consistent hash ring with virtual nodes |
| `internal/vectorclock` | Vector clock for conflict detection and resolution |
| `internal/merkle` | Merkle tree for efficient data comparison between nodes |
| `internal/store` | In-memory key-value store with versioned entries |
| `internal/server` | HTTP servers for proxy and node |
| `internal/client` | HTTP client for proxy communication |
| `cmd/proxy` | Proxy binary entry point |
| `cmd/node` | Node binary entry point |
| `tests/e2e` | End-to-end integration tests |

## Implementation Steps

### Step 1: Scaffolding
- go.mod, Makefile, Dockerfile, docker-compose.yml, .gitignore, config.yaml

### Step 2: Consistent Hashing
- Ring with virtual nodes (default 3 per physical node)
- `Add(nodeID)`, `Remove(nodeID)`, `Get(key) -> nodeID`
- Use SHA-256 for hash function

### Step 3: Vector Clock
- Map[nodeID]uint64 for each clock
- `Increment(nodeID)`, `Merge(other)`, `Compare(other) -> {Before, After, Concurrent, Equal}`
- JSON serialization for HTTP transport

### Step 4: Merkle Tree
- Build tree from sorted key hashes
- `Root() -> hash`, `Diff(other) -> []Key`
- Use for anti-entropy between nodes

### Step 5: In-Memory Store
- `Get(key) -> (value, clock, ok)`
- `Put(key, value, clock)`
- `Keys() -> []string` for Merkle tree building
- Thread-safe with sync.RWMutex

### Step 6: Node Server
- HTTP handlers: `/get`, `/put`, `/sync/merkle`, `/sync/keys`, `/health`
- Registers with proxy on startup

### Step 7: Proxy Server
- HTTP handlers: `/get`, `/put`, `/admin/node`, `/health`, `/status`
- Routes requests using consistent hashing
- Routes writes to primary + N-1 replicas
- Routes reads to primary, compares vector clocks, resolves conflicts
- Node management (add/remove via admin API)

### Step 8: Client
- HTTP client with Get/ Put methods

### Step 9: E2E Tests
- Basic put/get flow
- Concurrent writes
- Node addition/removal during operations
- Latency and success rate measurement

### Step 10: Docker & Documentation
- Multi-stage Dockerfile for proxy and node
- docker-compose.yml with proxy + 2 nodes (expandable to 5)
- README.md with architecture, setup, and testing guide
