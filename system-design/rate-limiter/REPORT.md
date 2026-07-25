# Rate Limiter — Test Report

_Generated: 2026-07-25T10:45:59+08:00_

## ✅ E2E Comparison Report

200 requests per limiter (capacity=10, refill=10/s); 1M `Allow()` calls for benchmark.

| ⚙️ Limiter | 📏 Cap | ✅ Allowed | ❌ Rejected | ⏱️ ns/op | 🧠 Mem (MiB) |
|---|---|---|---|---|---|
| 🪣 Token Bucket | 10 | 10 | 190 | 18.2M (55 ns) | 0.6 MiB |
| 💧 Leaking Bucket | 10 | 10 | 190 | 18.5M (54 ns) | 0.6 MiB |
| 📈 BBR | N/A | 106 | 94 | 13.5M (74 ns) | 0.6 MiB |

## 🔥 Attacker Mindset Analysis

### 🚨 Above Threshold (3× capacity at once)

| ⚙️ Limiter | 📏 Cap | ✅ Allowed | ❌ Rejected | 🛡️ Verdict |
|---|---|---|---|---|
| 🪣 Token Bucket | 10 | 10 | 20 | ✅ Correctly capped |
| 💧 Leaking Bucket | 10 | 10 | 20 | ✅ Correctly capped |
| 📈 BBR | N/A | 16 | 14 | ⚠️ Probabilistic, not precise |

### ⚡ Burst vs Steady (2s each)

| ⚙️ Limiter | 🔄 Steady allow % | 💥 Burst allow % | 📝 Behavior |
|---|---|---|---|
| 🪣 Token Bucket | 19.8% | 2.0% | ✅ Steady gets more throughput |
| 💧 Leaking Bucket | 19.8% | 1.9% | ✅ Steady gets more throughput |
| 📈 BBR | 44.8% | 49.9% | ⚖️ Nearly equal (probabilistic) |

### 🔄 Double Traffic / Reset Attack

Drain → idle 2s → burst 3× capacity.

| ⚙️ Limiter | ✅ Allowed after idle | 🔓 Double traffic? |
|---|---|---|
| 🪣 Token Bucket | 10 | ✅ No (capped at capacity) |
| 💧 Leaking Bucket | 10 | ✅ No (capped at capacity) |
| 📈 BBR | 19 | ⚠️ Permissive after idle |

## 📊 Benchmark (Apple M4)

1M `Allow()` calls per limiter.

| ⚙️ Limiter | 🚀 Ops/sec | ⏱️ ns/op | 💾 B/op | 📦 allocs/op | 🧠 Heap (MiB) |
|---|---|---|---|---|---|
| 🪣 Token Bucket | 18.2M | 55 | 0 | 0 | 0.6 |
| 💧 Leaking Bucket | 18.5M | 54 | 0 | 0 | 0.6 |
| 📈 BBR | 13.5M | 74 | 16 | 0 | 0.6 |

## 📝 Tradeoff Summary

| **Feature** | **🪣 Token Bucket** | **💧 Leaking Bucket** | **📈 BBR** |
|---|---|---|---|
| **Burst handling** | ✅ Excellent | ⚠️ Queue fills up | ⚖️ Adaptive |
| **Precision** | ✅ Exact | ✅ Exact | ⚠️ Probabilistic |
| **Memory** | 💾 Low (~0 B) | 💾 Queue-size dependent | 💾 Low (~0 B) |
| **Latency** | ⚡ ~79 ns | ⚡ ~75 ns | 🐢 ~134 ns |
| **Best for** | Predictable rate limits | Smooth traffic shaping | Unknown/varying traffic |

