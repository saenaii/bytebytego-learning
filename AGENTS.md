# AGENTS.md — Project Rules

## Go Coding Rules

1. **Use `any` instead of `interface{}`** — all new code must use `any` (Go 1.18+) rather than the legacy `interface{}` alias.

2. **Unit tests must use table-driven tests** — every test function should follow the standard Go table-driven test pattern:
   ```go
   func TestXxx(t *testing.T) {
       tests := []struct {
           name string
           ...
       }{
           {name: "case 1", ...},
           {name: "case 2", ...},
       }
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               ...
           })
       }
   }
   ```

3. **Keep cyclomatic complexity ≤ 15** — any function with cyclomatic complexity exceeding 15 must be refactored into smaller functions.

4. **Run unit tests only when Go files change** — run `go test -race ./...` after coding only if Go files were modified. Skip tests when changes are limited to documentation, config, markdown, or other non-Go files.

5. **E2E tests must use realistic latency** — never use sub-millisecond delays (like 10μs) in benchmarks or E2E tests. Simulate real-world conditions with latencies in the 100ms–5s range, and use concurrent goroutines to model production traffic patterns.

6. **E2E tests must generate a `REPORT.md`** — after every E2E test run, the test suite must write a human-friendly `REPORT.md` in the project root. The report must include:
   - Emoji and Markdown tables for readability
   - Total operations, success/failure counts, and success rate
   - Latency breakdown (Average, P50, P95, P99) for each operation type (PUT, GET)
   - Architecture summary (topology, replication config, conflict resolution strategy)

## Task Management Rules

1. **Checkboxes belong in `tracker.md` only** — task checkboxes (`- [ ]` / `- [x]`) must be placed in `system-design/*/tasks/tracker.md`, not in the daily task files (`YYYY-MM-DD.md`). Daily task files contain only task descriptions without checkbox markers.

2. **Record conversation summaries in `tracker.md`** — After each AI interaction, append a 1-sentence summary under today's date section in the relevant `tasks/tracker.md` for the current `system-design/*/` context. If no date section exists for today, create one (`## YYYY-MM-DD`). If multiple conversations occur in one day, append new records below the existing ones. If the working directory doesn't clearly map to a tracker file, ask the user which tracker to update.

## Implement Rule: Structural Prompt Normalizer for Live Demos

### 1. Trigger Condition
Whenever the user requests to **design, implement, or evaluate a technical live demo, system architecture, or algorithm visualizer** (e.g., Consistent Hashing, Rate Limiters, Distributed Locks, Distributed KV Stores, etc.), you MUST follow the **6-Step Prompt Normalization Protocol** before executing or generating code.

---

### 2. Execution Workflow
1. **Format First**: Reframe and expand the user's intent into the **6-Step Prompt Architecture** defined below.
2. **Confirm Alignment**: Present the normalized structure to the user (or acknowledge it silently if operating in fully automated agent mode).
3. **Execute & Deliver**: Proceed with the technical implementation strictly adhering to the normalized prompt specifications.

---

### 3. The 6-Step Prompt Architecture Template

```markdown
# Role & Target
You are a Senior Systems Architect. Implement a high-performance, runnable local demo for [System/Algorithm Name] demonstrating [Core Business Value].

# Architecture & Tech Stack
- **Language & Framework**: [e.g., Go / Node.js / Python]
- **Key Algorithms & Hash**: [e.g., MurmurHash3 for uniform distribution]
- **Components**:
  1. `Proxy/Gateway`: Core router logic & admin control APIs.
  2. `Worker Nodes`: Lightweight backend services.
  3. `Traffic Generator`: Automated client sending continuous key/request streams.

# Technical Specifications
1. **Core Data Structure & Algorithm**:
   - [Specification 1: e.g., Ring lookup using Binary Search O(log K)]
   - [Specification 2: e.g., Virtual Nodes mapping N=50 per real node]
2. **Dynamic Operations**:
   - `POST /admin/nodes` -> Scale out node
   - `DELETE /admin/nodes/{id}` -> Scale in node
3. **Metrics & Observability**:
   - `GET /admin/stats` -> Returns QPS per node, key distribution, and Load Standard Deviation.

# Live Demo & Dashboard Specifications
1. **UI/Dashboard Mode**: [e.g., Embedded Web Dashboard with Chart.js / Vue]
2. **Interactive Controls**: Real-time "Add Node" and "Remove Node" trigger buttons on UI.
3. **Visual Highlights**:
   - Real-time QPS/Key Distribution Bar Chart (Auto-refresh every 1s).
   - Dynamic Load Standard Deviation Meter (Highlights balance quality).

# Test & Verification Workflow
1. Run system with default node count via `docker-compose up`.
2. Generate continuous load (e.g., 100 QPS across uniform key space).
3. Scale Out -> Verify remapping ratio / load balance shift.
4. Scale In -> Verify smooth failover.
5. Key metrics (e.g., Standard deviation) must stay within acceptable threshold (e.g., ±15%).

# Output Deliverables
1. Complete executable code with clear domain separation.
2. `docker-compose.yml` for single-command start (`docker-compose up`).
3. `design.md` containing architectural diagrams, API contracts, and verification test logs.
