---
name: implement-design
description: Turn a topic markdown file into an implementation-ready design.md, then build the runnable project in this repository following its existing folder structure, conventions, and AGENTS.md rules.
---

# Implement Design

Build a topic in this repository in two phases: synthesize an implementation-ready `design.md`, then implement it. The design is the contract; the code follows it.

## When to use

Use this skill when the user asks to:

- implement or build a topic under `system-design/`
- turn a topic `README.md`, study note, spec, or other markdown file into `design.md` and working code
- continue or finish an implementation from an existing `design.md`

Do not use it for:

- writing or rewriting study notes — use the `system-design` skill
- mock interviews — use the `mock-interview` skill
- algorithm walkthroughs under `coding-interview/`, which are not runnable services
- commits or pushes unless the user explicitly requests them

## Inputs and modes

| Input | Behavior |
| --- | --- |
| An explicit path to any `.md` | Use it as the source; infer the topic directory from its parent |
| A topic directory | Prefer `design.md`; fall back to `README.md` |
| `system-design/<topic>/design.md` | Treat as the contract and skip design synthesis |
| `system-design/<topic>/README.md` | Treat as the source note; synthesize `design.md` first |
| `coding-interview/*` | Stop: this skill targets runnable system-design topics |

Resolve the topic directory first and preserve its exact spelling, including typos.

Determine the mode:

- **greenfield** — no code in the topic directory: synthesize the design and implement.
- **existing** — code already present: verify it against `design.md`, fix gaps, and change only what the request or the design requires. Never regenerate the project wholesale.

## Phase 0 — Discover

1. Read the root `AGENTS.md` and any `AGENTS.md` inside the topic.
2. Read the topic `README.md`, `design.md` when present, and `tasks/tracker.md`.
3. Run `git status` and list modified and untracked files. Preserve unrelated work.
4. Detect the stack from repository markers (`go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`) and from neighboring topics. Do not assume a language.
5. Read the repository's 6-step prompt-normalization rule in `AGENTS.md` when it exists and reuse its headings: role and target, architecture and stack, technical specifications, live demo and dashboard, test and verification, output deliverables.
6. Stop and ask only when a wrong guess would be costly: unclear topic, missing source, a destructive rewrite, or a multi-service build.

## Phase 1 — Write design.md

Source precedence:

1. An explicit markdown file named by the user.
2. `design.md`, which is the contract and needs no synthesis.
3. The topic `README.md`.
4. Other markdown files in the topic.

Rules:

- Extract only what the source supports: problem scope, scale assumptions, APIs, data model, algorithms, request flows, failure modes, and trade-offs.
- Never invent requirements silently. Record every gap as an explicit `Assumption` or `Open question`.
- Do not copy stack choices, ports, directory names, or domain semantics from another topic.
- Follow `references/design-template.md` for the section structure.
- When `design.md` exists and the source is another file, show a section-level diff summary and get confirmation before replacing it. An explicit "rewrite it" from the user overrides this gate.
- Preserve user-authored decisions and accepted trade-offs when merging.
- Record the source link, revision date, and any course URL in the `Source and scope` section.
- Blocking open questions must be answered before Phase 3.

For greenfield work, the design must be complete enough that implementation needs no further product decisions.

## Phase 2 — Plan the build

1. Derive the file tree from the design and mirror the nearest implemented topic's layout.
2. List the packages or modules, endpoints, configuration, and persistence changes.
3. Map every acceptance criterion to a test or a verification command.
4. Identify risky or ambiguous items and resolve them before coding.

Show the plan and get approval for anything that spans multiple packages or services.

## Phase 3 — Implement

- Follow the existing repository layout rather than inventing a new one. See the conventions table below.
- Apply the language rules in `AGENTS.md`. When Go is used, the non-negotiables are `any` instead of `interface{}`, table-driven tests, and cyclomatic complexity at or below 15.
- Keep browser assets local and embedded when the design calls for a dashboard.
- Add no web framework, dependency, build pipeline, or service that the design does not name.
- Keep code and design aligned. If implementation reveals a design flaw, update the design first and say so; do not let them drift.
- Do not touch unrelated topics, files, or pre-existing user changes.

## Phase 4 — Validate

Derive commands from the repository instead of hardcoding them:

| Repository signal | Commands |
| --- | --- |
| `AGENTS.md` or `Makefile` names checks | Use those first |
| `go.mod` | `gofmt -w .`, `go vet ./...`, `go test -race ./...` |
| `package.json` | `npm test`, `npm run lint` |
| `pyproject.toml` | `pytest`, `ruff check` |
| `docker-compose.yml` | `docker compose config`, `up --build -d`, then smoke-test documented endpoints |

Additional rules:

- Run tests only when source files for that stack changed; skip them for documentation-only edits.
- When the topic has benchmarks or end-to-end tests, use realistic latency in the 100 ms to 5 s range and write `REPORT.md` as `AGENTS.md` requires.
- Write the verification log into `design.md`: date, exact command, and outcome.
- Report every command that could not run and its exact failure. Never claim an unrun check passed.

## Phase 5 — Close out

1. Append a one-sentence summary under today's date in the topic `tasks/tracker.md`, creating the date section when missing. Checkboxes belong only there.
2. Update `system-design/README.md` only when a new topic was created.
3. Run the definition of done below.
4. Re-read `design.md` and confirm the implementation matches it, section by section.

## Repository conventions

Read these from `AGENTS.md` at runtime. Use this table as the fallback when it is missing or silent.

| Rule | Value |
| --- | --- |
| Topic root | `system-design/<requested-directory>/`; preserve exact spelling |
| Required docs | `README.md`, `design.md`, `tasks/tracker.md` |
| Typical code layout | `cmd/<binary>/`, `internal/<package>/`, `web/static/`, `docker-compose.yml`, `Dockerfile`, `Makefile`, `config.json` |
| Index | Update `system-design/README.md` only for a new topic |
| Tracker | One sentence per interaction; checkboxes only here |
| Secrets | Environment variables only, never `config.json`; fail closed when unset |

## Definition of done

- `design.md` has no empty sections and no unresolved blocking questions.
- Every acceptance criterion maps to a passing test or a recorded verification command.
- Build, lint, and tests pass, or failures are reported with exact output.
- `docker compose config` is valid and the documented endpoints respond when a compose file exists.
- No secrets are committed and no unrelated files changed.
- Tracker entry exists and the index is updated when required.
- No drift between `design.md` and the implementation.

## Anti-patterns

- Hardcoding a language, port, directory name, or command from another topic.
- Copying domain semantics from a neighboring project.
- Rewriting an existing `design.md` without showing the diff.
- Regenerating code when the request was to verify or extend it.
- Implementing features the design does not describe.
- Claiming checks passed without running them.
- Committing, pushing, or reverting unrelated work.

A worked example lives in `references/example-topic.md`. It is a reference, not an instruction set; do not copy its stack or domain choices.
