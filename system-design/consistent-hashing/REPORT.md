# 📊 Consistent Hashing — E2E Report

## ✅ Test Summary

| Metric | Value |
|--------|-------|
| Total Operations | 879 |
| Successful | 879 |
| Failed | 0 |
| **Success Rate** | **100.00%** |
| Topology Changes | 5 |
| Final Active Nodes | 3 |

## ⏱️ Latency Breakdown

| Metric | PUT | GET |
|--------|-----|-----|
| Count | 359 | 520 |
| Average | 2557 ms | 2668 ms |
| P50 | 2472 ms | 2750 ms |
| P95 | 4771 ms | 4745 ms |
| P99 | 4921 ms | 4937 ms |

## 📈 Key Migration & Balance

| Phase | Result |
|-------|--------|
| Steady state (3 nodes) | load std-dev / mean = **6.12%** (threshold ≤ 15%) |
| Scale out 3→4 | **21.3%** of keys migrated (expected 18–32%; naive rehash ≈75%); std-dev = **9.44%** |
| Scale in 4→3 | exactly **318** keys re-homed (only the removed node's keys); zero traffic to disabled node |
| Post-scale-in (info) | std-dev = **24.05%** — converges as live traffic rewrites orphaned keys (no data-migration engine in this demo) |

## 🏗️ Architecture Summary

- **Topology**: 1 Proxy + 5 Worker Nodes (3 active + 2 healthy standby), in-memory KV
- **Hash Ring**: consistent hashing, MD5 truncated to 64-bit, **50 virtual nodes per physical node**
- **Lookup**: binary search over the sorted vnode ring → **O(log K)**, K = nodes × 50
- **Replication**: none — each key has exactly one owner (routing is purely in-proxy)
- **Conflict Resolution**: N/A (single writer per key via ring ownership)
- **Simulated Latency**: 100ms–5s per worker request (realistic network conditions)
- **Standby Model**: node-4/5 boot healthy but unregistered; enabled live via 'POST /admin/nodes/{id}'
