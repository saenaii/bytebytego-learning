# Rate Limiter — Test Report

_Generated: 2026-07-24T21:24:37+08:00_

## ✅ E2E Comparison Report

200 requests per limiter (capacity=10, refill=10/s).

| ⚙️ Limiter | 📊 Total | ✅ Allowed | ❌ Rejected |
|---|---|---|---|
| 🪣 Token Bucket | 200 | 10 | 190 |
| 💧 Leaking Bucket | 200 | 10 | 190 |
| 📈 BBR | 200 | 97 | 103 |

## 🔥 Attacker Mindset Analysis

### 🚨 Above Threshold (3× capacity at once)

| ⚙️ Limiter | ✅ Allowed | ❌ Rejected | 🛡️ Verdict |
|---|---|---|---|
| 🪣 Token Bucket | 10 | 20 | ✅ Correctly capped |
| 💧 Leaking Bucket | 10 | 20 | ✅ Correctly capped |
| 📈 BBR | 15 | 15 | ⚠️ Probabilistic, not precise |

### ⚡ Burst vs Steady (2s each)

| ⚙️ Limiter | 🔄 Steady allow % | 💥 Burst allow % | 📝 Behavior |
|---|---|---|---|
| 🪣 Token Bucket | 20.7% | 2.0% | ✅ Steady gets more throughput |
| 💧 Leaking Bucket | 19.6% | 1.9% | ✅ Steady gets more throughput |
| 📈 BBR | 58.1% | 51.3% | ⚖️ Nearly equal (probabilistic) |

### 🔄 Double Traffic / Reset Attack

Drain → idle 2s → burst 3× capacity.

| ⚙️ Limiter | ✅ Allowed after idle | 🔓 Double traffic? |
|---|---|---|
| 🪣 Token Bucket | 10 | ✅ No (capped at capacity) |
| 💧 Leaking Bucket | 10 | ✅ No (capped at capacity) |
| 📈 BBR | 16 | ⚠️ Permissive after idle |

## 📊 Benchmark (Apple M4)

1M `Allow()` calls per limiter.

| ⚙️ Limiter | 🚀 Ops/sec | ⏱️ ns/op | 💾 B/op | 📦 allocs/op |
|---|---|---|---|---|
| 🪣 Token Bucket | 12.7M | 79 | 0 | 0 |
| 💧 Leaking Bucket | 13.0M | 77 | 0 | 0 |
| 📈 BBR | 7.2M | 138 | 0 | 0 |

## 📝 Tradeoff Summary

| **Feature** | **🪣 Token Bucket** | **💧 Leaking Bucket** | **📈 BBR** |
|---|---|---|---|
| **Burst handling** | ✅ Excellent | ⚠️ Queue fills up | ⚖️ Adaptive |
| **Precision** | ✅ Exact | ✅ Exact | ⚠️ Probabilistic |
| **Memory** | 💾 Low (~0 B) | 💾 Queue-size dependent | 💾 Low (~0 B) |
| **Latency** | ⚡ ~79 ns | ⚡ ~75 ns | 🐢 ~134 ns |
| **Best for** | Predictable rate limits | Smooth traffic shaping | Unknown/varying traffic |

