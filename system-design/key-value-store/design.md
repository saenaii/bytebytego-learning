# Design & Implement a Distributed Key-Value Store in Local

## High-Level Architecture
- Topology: Docker-based local cluster simulating a distributed system. Run 3 initial storage node containers + 1 proxy container in `docker-compose.yml`.
- Routing Flow: `client` -> `proxy/coordinator` -> `cluster nodes`.

## System Requirements & Technical Details
- **Data Distribution**: Consistent Hashing with **Virtual Nodes** (e.g., 50 vnodes per physical node) to handle dynamic node addition/removal gracefully.
- **Operations**: Support `GET(key)` and `PUT(key, value)`. 
- **Replication & Quorum**: Replication Factor $N=3$. Configurable Quorum ($W=2, R=2$) for read/write operations.
- **Conflict Resolution & Sync**:
  - **Online Path (Vector Clock)**: Use Vector Clock to detect concurrent writes. Return siblings/multi-version when concurrent conflict occurs during `GET`.
  - **Offline/Repair Path (Merkle Tree & Anti-Entropy)**: Each node maintains an in-memory Merkle Tree. Periodically run Anti-Entropy between replicas to heal stale or partitioned data.
- **Storage**: In-memory storage per node. Keys and values are lightweight strings. Data should be cleared after test execution.
- **Deployment**: Single `docker-compose up` command to launch the entire environment.

## Chaos Engineering & Testing Requirements
- **Unit & Integration Testing**: Coverage for Vector Clock merge logic, Merkle Tree diff calculation, and Hash Ring partitioning.
- **E2E & Chaos Tests (Attacker / Chaos View)**:
  - Simulate network partition (e.g., isolating 1 node) and node failure/rejoin (max 5 nodes dynamically).
  - High concurrency test: Concurrent `PUT`/`GET` on the same keys during network instability.
  - Assertion: Verify eventual consistency (Anti-Entropy heals all replicas once the network heals).
- **Benchmark & Observability Report**:
  - Generate an E2E metrics report covering CPU, Memory, I/O, Request Success Rate, and P95/P99 latency under chaos scenarios.

## Programming Language
- Go (Golang)

## Dev-Flow
1. Analyze requirements and generate a `plan.md` outlining the package structure, interface definitions, and state machine transitions.
2. Implement core components step-by-step: Hash Ring -> Vector Clock -> Node Store -> Coordinator/Proxy -> Merkle Anti-Entropy.
3. Write UT and Chaos E2E test suite.
4. Generate `README.md`, Dockerfiles, and `docker-compose.yml`.
5. Ensure all tests pass cleanly via automated scripts.
