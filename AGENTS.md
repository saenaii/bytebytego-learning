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
