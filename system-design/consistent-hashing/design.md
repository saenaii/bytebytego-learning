# Design and Implement a Consistent Hashing Demo

## 1. Role & Target
A Senior Systems Architect demonstration of **Consistent Hashing with virtual nodes**
(ByteByteGo pattern): a runnable, local demo proving that dynamic scale-out / scale-in
redistributes traffic evenly while migrating the **minimum possible number of keys** —
unlike naive `key % N` hashing, which rehashes ~all keys on every topology change.

## 2. Architecture & Tech Stack
- **Language & Framework**: Go 1.26 (stdlib only, no external deps)
- **Hash Function**: MD5 (truncated to 64-bit) — deterministic, well-distributed
- **Key Algorithm**: Consistent hash ring with **N=50 virtual nodes per physical node**;
  lookup via **binary search over the sorted vnode ring → O(log K)**, K = nodes × 50
- **Deployment**: `docker-compose up` (single command), Alpine runtime images

### Components
```
                   ┌──────────────┐        ┌──────────────┐
                   │   Traffic    │        │   Browser    │
                   │  Generator   │        │  Dashboard   │
                   └──────┬───────┘        └──────┬───────┘
                          │ PUT/GET /kv/{key}     │ GET / (page) + /stats + admin
                          ▼                       ▼
                   ┌─────────────────────────────────────┐
                   │  Proxy (hash ring + web dashboard)  │
                   │  ring.Get(key) → binary search      │
                   └──┬─────────┬─────────┬──────────────┘
                      ▼         ▼         ▼
                ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
                │ node-1 │ │ node-2 │ │ node-3 │ │ node-4 │ │ node-5 │
                └────────┘ └────────┘ └────────┘ └────────┘ └────────┘
                        (node-4/5 run in standby: healthy but unregistered)
```
1. **Proxy** (`cmd/proxy`): owns the `internal/ring.Ring`, routes `/kv/{key}` to the
   responsible worker, serves the embedded dashboard, exposes admin + metrics APIs.
2. **Worker Node** (`internal/node`): HTTP KV store with **simulated 100ms–5s latency**;
   reports metrics via `/stats`, key set via `/keys`.
3. **Traffic Generator** (`cmd/traffic`): continuous random-key traffic at 100 req/s
   over a 10k key space.
4. **Dashboard** (embedded in proxy binary via `go:embed`): browser UI, no build step.

### Why virtual nodes?
- With few physical nodes, plain hashing concentrates keys onto the smallest arc (bad balance).
- 50 vnodes/node smooths the ring: each physical node owns ~50 uniformly spread arcs,
  keeping load standard deviation within ±15% of the mean.
- On scale changes only the affected vnodes' arcs move — not the whole key space.

## 3. Technical Specifications

### 3.1 Core Data Structure (`internal/ring`)
| Item | Detail |
|---|---|
| Ring | `map[uint64]string` vnode hash → nodeID + `[]uint64` sorted vnode hashes |
| Vnode key | `"<nodeID>#<i>"`, i ∈ [0, 50) |
| Add / Remove | Insert/delete 50 vnode points, re-sort → O(K log K), K = nodes × 50 |
| Get(key) | `binary search` first vnode ≥ hash(key), wrap-around → **O(log K)** |
| Concurrency | `sync.RWMutex`; concurrent lookups allowed |

### 3.2 API Contracts
| Method | Path | Purpose |
|---|---|---|
| `GET` | `/` | Embedded HTML dashboard (Chart.js) |
| `PUT` | `/kv/{key}` | Store `{"value":"..."}` → routed to responsible node |
| `GET` | `/kv/{key}` | Fetch value from responsible node |
| `POST` | `/admin/nodes/{id}` | Enable a **standby** worker (add to ring) |
| `DELETE` | `/admin/nodes/{id}` | Disable a worker (remove from ring) |
| `GET` | `/stats` | Per-node QPS, key count, % load, load std-dev |
| `GET` | `/keys` (node) | Node's stored key set — measures **actual key migration** |

**Standby model for interactivity:** docker-compose starts all 5 nodes; node-4/5 boot
healthy but stay **unregistered**. The dashboard's Add/Remove buttons call the admin API,
so scale-out/in happens live in the browser with zero terminal commands. Routing is purely
in-proxy, so rebalancing is immediate.

### 3.3 Metrics & Observability (`GET /stats`)
- QPS + total requests per node (PUT/GET split), key count, distribution %
- **Load standard deviation** (balance quality meter)
- Per-node latency avg/p50/p95/p99 collected by the traffic generator

## 4. Live Demo & Dashboard Specifications
1. **Dashboard mode**: embedded HTML page served by the proxy at `http://localhost:8080`
   — vanilla JS + Chart.js, zero build step, works offline of any CLI.
2. **Interactive controls**:
   - **Add Node** → `POST /admin/nodes/node-4` (then node-5) → ring rebalances live
   - **Remove Node** → `DELETE /admin/nodes/{id}` → traffic reroutes instantly
3. **Visual highlights** (auto-refresh every 1s):
   - Real-time request/QPS bar chart per node
   - Key distribution bar chart + ring visualization (vnode arcs colored per node)
   - **Load Standard Deviation meter** (green ≤10% / yellow ≤15% / red >15%)
   - Live migration counter: keys moved since last topology change

## 5. Test & Verification Workflow
| Phase | Action | Acceptance |
|---|---|---|
| Unit | `make test` (table-driven, `-race`) | Balance ±15%; migration 18–32% on 3→4; only removed node's keys move on scale-in |
| E2E | `make test-e2e` | Simulated latency 100ms–5s; 3 nodes → add → remove; success rate ≥ 99% |
| Report | E2E writes `REPORT.md` | Ops, success/failure, P50/P95/P99 per op type, architecture summary |

**Expected results**
- **Scale out (3→4 nodes):** ~25% of keys migrate (vs ~75% naive rehash); std-dev/mean ≤ 15%.
- **Scale in:** only the removed node's keys re-home; zero traffic to a disabled node.

## 6. Output Deliverables
1. Go source: `internal/ring`, `internal/node`, `cmd/{proxy,node,traffic}`, `internal/dashboard`
2. `docker-compose.yml` — one-command start with 5 nodes (2 standby for interactive demo)
3. `design.md` + auto-generated `REPORT.md` (per E2E run, per AGENTS.md)
4. **`http://localhost:8080` dashboard** — the primary live-demo surface
