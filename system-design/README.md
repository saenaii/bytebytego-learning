# System Design Learning

Runnable system-design examples and study notes inspired by [ByteByteGo](https://bytebytego.com/). Each topic combines a concise explanation of the core ideas with an implementation or design artifact that can be explored locally.

## Topics

| Topic | What it covers |
| --- | --- |
| [Back-of-the-Envelope Estimation](./back-of-the-envelope-estimation/) | Estimating QPS, storage, latency, availability, and capacity from explicit assumptions. |
| [Scale From Zero to Millions of Users](./scale-from-zero-to-millions-of-users/) | The incremental path from a single server to a globally distributed, observable, sharded system. |
| [Rate Limiter](./rate-limiter/) | Token bucket, leaking bucket, and adaptive BBR rate-limiting algorithms in Go. |
| [Consistent Hashing](./consistent-hashing/) | Hash rings, virtual nodes, traffic distribution, and low-migration scale-out/scale-in. |
| [Distributed Key-Value Store](./key-value-store/) | Consistent hashing, replication, quorum reads/writes, vector clocks, and Merkle-tree repair. |

## Repository layout

Each topic is self-contained and typically includes:

- `README.md` — key takeaways, architecture, and usage notes.
- `design.md` — design decisions and system contracts, where applicable.
- `cmd/` and `internal/` — Go implementation code, where applicable.
- `tests/` — end-to-end tests, where applicable.
- `docker-compose.yml` and `Makefile` — local build and run workflows, where applicable.
- `tasks/` — implementation task tracking.

## Getting started

Choose a topic and follow its README. For runnable Go projects, the common workflow is:

```bash
cd <topic-directory>
make test
make build
```

Some topics also provide Docker-based workflows and end-to-end tests. Check the topic-specific README and Makefile for the available commands.

## Learning path

The topics can be approached in this order:

1. Start with [Scale From Zero to Millions of Users](./scale-from-zero-to-millions-of-users/) for the overall architecture progression.
2. Learn [Back-of-the-Envelope Estimation](./back-of-the-envelope-estimation/) to quantify scale and capacity before selecting components.
3. Study [Rate Limiter](./rate-limiter/) for protecting services under load.
4. Explore [Consistent Hashing](./consistent-hashing/) for scalable data and traffic placement.
5. Finish with the [Distributed Key-Value Store](./key-value-store/) to see these ideas composed into a distributed system.
