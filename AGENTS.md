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

4. **Run unit tests after every change** — run `go test -race ./...` after coding and ensure all tests pass before moving to the next step.

5. **E2E tests must use realistic latency** — never use sub-millisecond delays (like 10μs) in benchmarks or E2E tests. Simulate real-world conditions with latencies in the 100ms–5s range, and use concurrent goroutines to model production traffic patterns.

## Task Management Rules

1. **Checkboxes belong in `tracker.md` only** — task checkboxes (`- [ ]` / `- [x]`) must be placed in `system-design/*/tasks/tracker.md`, not in the daily task files (`YYYY-MM-DD.md`). Daily task files contain only task descriptions without checkbox markers.
