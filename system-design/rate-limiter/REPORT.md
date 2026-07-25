# Rate Limiter — Test Report

_Generated: 2026-07-25T11:22:47+08:00_

## 🔧 Rate Limiter View

200 sequential requests per limiter (capacity=10, refill=10/s).

| ⚙️ Limiter | 📏 Cap | ✅ Allowed | ❌ Rejected |
|---|---|---|---|
| 🪣 Token Bucket | 10 | 10 | 190 |
| 💧 Leaking Bucket | 10 | 10 | 190 |
| 📈 BBR | N/A | 97 | 103 |

## 👤 User View

200 concurrent requests with simulated processing delay (avg 200ms, max 5s, exponential).

| ⚙️ Limiter | 📊 Total | ✅ 200 | ❌ 429 | ✅ Success Rate | 🎯 p95 | 🎯 p99 |
|---|---|---|---|---|---|---|
| 🪣 Token Bucket | 200 | 32 | 168 | 16.0% | 554.022583ms | 927.043667ms |
| 💧 Leaking Bucket | 200 | 20 | 180 | 10.0% | 574.5235ms | 998.021417ms |
| 📈 BBR | 200 | 111 | 89 | 55.5% | 538.409959ms | 803.055708ms |

## 📊 Benchmark (Apple M4)

1M `Allow()` calls per limiter.

| ⚙️ Limiter | 🚀 Ops/sec | ⏱️ ns/op | 💾 B/op | 📦 allocs/op | 🧠 Heap (MiB) |
|---|---|---|---|---|---|
| 🪣 Token Bucket | 18.2M | 55 | 0 | 0 | 0.6 |
| 💧 Leaking Bucket | 20.0M | 50 | 0 | 0 | 0.7 |
| 📈 BBR | 11.9M | 84 | 0 | 0 | 0.7 |

## 🔥 Attacker Mindset Analysis

### 🚨 Above Threshold (3× capacity at once)

| ⚙️ Limiter | 📏 Cap | ✅ Allowed | ❌ Rejected | 🛡️ Verdict |
|---|---|---|---|---|
| 🪣 Token Bucket | 10 | 10 | 20 | ✅ Correctly capped |
| 💧 Leaking Bucket | 10 | 10 | 20 | ✅ Correctly capped |
| 📈 BBR | N/A | 11 | 19 | ⚠️ Probabilistic, not precise |

### ⚡ Burst vs Steady (2s each)

| ⚙️ Limiter | 🔄 Steady allow % | 💥 Burst allow % | 📝 Behavior |
|---|---|---|---|
| 🪣 Token Bucket | 19.8% | 2.0% | ✅ Steady gets more throughput |
| 💧 Leaking Bucket | 19.8% | 1.9% | ✅ Steady gets more throughput |
| 📈 BBR | 46.9% | 49.2% | ⚖️ Nearly equal (probabilistic) |

### 🔄 Double Traffic / Reset Attack

Drain → idle 2s → burst 3× capacity.

| ⚙️ Limiter | ✅ Allowed after idle | 🔓 Double traffic? |
|---|---|---|
| 🪣 Token Bucket | 10 | ✅ No (capped at capacity) |
| 💧 Leaking Bucket | 10 | ✅ No (capped at capacity) |
| 📈 BBR | 12 | ⚠️ Permissive after idle |

## 📝 Tradeoff Summary

| **Feature** | **🪣 Token Bucket** | **💧 Leaking Bucket** | **📈 BBR** |
|---|---|---|---|
| **Burst handling** | ✅ Excellent | ⚠️ Queue fills up | ⚖️ Adaptive |
| **Precision** | ✅ Exact | ✅ Exact | ⚠️ Probabilistic |
| **Memory** | 💾 Low (~0 B) | 💾 Queue-size dependent | 💾 Low (~0 B) |
| **Latency** | ⚡ ~79 ns | ⚡ ~75 ns | 🐢 ~134 ns |
| **Best for** | Predictable rate limits | Smooth traffic shaping | Unknown/varying traffic |

