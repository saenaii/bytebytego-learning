# 📊 Performance Report

## Test Summary

| Metric | Value |
|--------|-------|
| Total Operations | 400 |
| Successful | 400 |
| Failed | 0 |
| **Success Rate** | **100.0%** |

## Latency Breakdown

| Metric | PUT | GET | Overall |
|--------|-----|-----|---------|
| Count | 200 | 200 | 400 |
| Average | 165.662359ms | 4.19796ms | 84.93016ms |
| P50 (Median) | 167.9155ms | 1.23625ms | 83.797291ms |
| P95 | 214.904667ms | 31.264083ms | 207.32ms |
| P99 | 221.418709ms | 43.507583ms | 219.280166ms |

## Architecture

- **Topology**: 1 Proxy + 3 Storage Nodes (in-memory)
- **Replication**: N=3, W=2, R=2 (Quorum-based)
- **Virtual Nodes**: 50 per physical node
- **Conflict Resolution**: Vector Clocks (on read) + Merkle Tree Anti-Entropy (background)
