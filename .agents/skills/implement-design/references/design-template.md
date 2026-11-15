# <Topic> — Design Document

> One sentence on what will be built, for whom, and why.

**Status:** draft | approved | implemented
**Last updated:** YYYY-MM-DD

## 1. Source and scope

- Source note: `<relative path or URL>`
- Course or reference link: `<url>`
- In scope: `<bullets>`
- Out of scope: `<bullets>`

<!--
Keep the scope short and testable. Anything listed as out of scope must not
appear later in the document.
-->

## 2. Assumptions and open questions

| # | Item | Type | Resolution |
| ---: | --- | --- | --- |
| 1 | `<assumption or question>` | assumption / question | `<answer or owner>` |

<!--
Never invent a requirement. If the source is silent, record it here. A blocking
open question must be resolved before implementation starts.
-->

## 3. Architecture

```text
<client> -> <service> -> <datastore>
                     \-> <cache or queue>
```

- Topology: `<single process, compose services, workers>`
- Communication: `<HTTP, gRPC, queue>`
- Trust boundaries: `<what is public, what is internal>`

## 4. Components and responsibilities

| Component | Responsibility |
| --- | --- |
| `<path or package>` | `<single responsibility>` |

## 5. API contract

### `METHOD /path`

Request:

```json
{}
```

Response `200`:

```json
{}
```

Errors: `<status codes and meanings>`

<!--
Include every endpoint the implementation must expose, including health,
metrics, and administrative routes.
-->

## 6. Data model and storage

```sql
<schema, or the equivalent for the chosen store>
```

- Keys and uniqueness: `<constraints>`
- Indexes: `<why each exists>`
- Retention and growth: `<estimate>`
- Migration strategy: `<how schema changes are applied>`

## 7. Algorithms and complexity

| Operation | Approach | Time | Space |
| --- | --- | --- | --- |
| `<operation>` | `<algorithm>` | `<O(...)>` | `<O(...)>` |

<!--
State the invariant each algorithm preserves. If a data structure is chosen over
an alternative, name the alternative and why it lost.
-->

## 8. Request flows

### `<Flow name>`

1. `<step>`
2. `<step>`

<!--
Cover the happy path, the idempotent or duplicate path, and the failure path for
each flow that matters.
-->

## 9. Configuration and secrets

| Setting | Source | Default | Notes |
| --- | --- | --- | --- |
| `<name>` | `config.json` / env | `<value>` | `<why>` |

- Secrets come from environment variables only and are never committed.
- Startup behavior when a required value is missing: `<fail closed / default>`.

## 10. Observability

- Metrics: `<names and meaning>`
- Logs: `<what is logged, what is never logged>`
- Dashboard: `<embedded, external, or none>`

## 11. Failure modes and trade-offs

| Failure | Impact | Mitigation |
| --- | --- | --- |
| `<dependency down>` | `<user-visible effect>` | `<behavior>` |

Accepted trade-offs:

- `<trade-off>`: `<what is given up and why that is acceptable>`

## 12. Testing and acceptance criteria

| Criterion | Verification |
| --- | --- |
| `<behavior>` | `<test name or command>` |

- Unit tests: `<scope, table-driven where applicable>`
- Integration tests: `<scope and required services>`
- End-to-end or benchmark tests: `<latency profile and report file>`

## 13. Deployment

- Runtime: `<image base, non-root, static binary>`
- Services: `<compose services and ports>`
- Start command: `<command>`
- Health check: `<endpoint>`

## 14. Verification log

| Date | Command | Result |
| --- | --- | --- |
| YYYY-MM-DD | `<command>` | `<pass, fail, or skipped and why>` |

<!--
Append entries; never rewrite history. Unrun checks are recorded as skipped with
the reason, never as passed.
-->
