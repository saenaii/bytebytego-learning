# Worked Example: URL Shortener

This is a reference for what a completed topic looks like. It is not an instruction set. Do not copy its stack, ports, directory name, or domain rules into another topic.

## Topic shape

```text
system-design/design-a-url-shortner/
├── README.md                 study note and source material
├── design.md                 implementation contract
├── config.json               non-secret runtime settings
├── docker-compose.yml        proxy, PostgreSQL, Redis
├── Dockerfile                multi-stage, non-root runtime
├── Makefile                  build and test shortcuts
├── cmd/proxy/main.go         wiring and startup
├── internal/                 config, shortcode, service, repository, handler, metrics
├── web/static/               embedded dashboard assets
└── tasks/tracker.md          interaction log
```

## What the design decided

| Decision | Choice | Reason recorded in `design.md` |
| --- | --- | --- |
| Short code | Base62 of the database ID | Collision-free and simple; enumeration accepted for the demo |
| Idempotency | `UNIQUE(normalized_url, mode)` | Repeated submissions return the existing record |
| Read path | Cache-aside in Redis with a TTL | Redirects dominate traffic |
| Redirect | `302 Found` | Keeps the service in the request path for metrics |
| Duplicate writes | Transaction plus unique-conflict retry | Concurrent submissions stay idempotent |
| Secrets | Environment only | `config.json` stays non-secret |

## How the workflow applied

1. **Discover** — read `AGENTS.md`, the topic note, and `git status`; detected Go from `go.mod` and the neighboring layout.
2. **Design** — turned the study note into `design.md` with scope, APIs, schema, flows, failure modes, and acceptance criteria.
3. **Plan** — mapped each acceptance criterion to a test or smoke check.
4. **Implement** — followed the repository layout; no web framework; local dashboard assets.
5. **Validate** — `gofmt`, `go vet`, `go test -race ./...`, `docker compose config`, then health and redirect smoke checks; results went into the verification log.
6. **Close out** — tracker entry and index update only where the repository required them.

## What was topic-specific

| Fact | Why it must not leak |
| --- | --- |
| Go with `net/http`, `pgx`, `go-redis` | Other topics may use another language or no database |
| PostgreSQL schema and `BIGSERIAL` | Domain-specific model |
| Base62, normalization, query blacklist | URL-shortener semantics |
| Port `8080` and `/shorturl/{code}` | Route and deployment choice |
| Embedded dashboard with local Chart.js | This demo's UI requirement |
| Directory spelling `design-a-url-shortner` | Preserve only for this topic |

## What generalized into the skill

- The study note is the source; `design.md` is the contract.
- The design is written before code and re-checked after implementation.
- Folder layout, commands, and tracker rules come from the repository, not from memory.
- Acceptance criteria map one-to-one to tests or commands.
- Unrun checks are reported as unrun.
- Unrelated work is preserved, and nothing is committed without being asked.
