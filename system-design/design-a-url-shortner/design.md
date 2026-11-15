# URL Shortener — Local Design Document

> Build a runnable URL-shortener demo in Go with a stateless HTTP proxy, PostgreSQL persistence, Redis caching, an embedded offline dashboard, and table-driven unit tests.

## 1. Goals and constraints

- Go 1.26 with `net/http`; no web framework.
- PostgreSQL stores durable URL mappings and generates unique numeric IDs.
- Redis provides cache-aside lookups and dashboard counters.
- Docker Compose starts the proxy, PostgreSQL, and Redis with one command.
- The runtime image is Alpine-based and runs as a non-root user.
- The dashboard is embedded in the Go binary and loads all browser assets locally.
- Short codes are Base62 encodings of database-generated numeric IDs; ID generation and code obfuscation are out of scope for this demo.
- The current repository contains uncommitted documentation changes; implementation must not overwrite unrelated work.

The service supports URL creation and redirection. Short URLs are not updated or deleted in this demo.

## 2. Architecture

```text
Browser / API client
        |
        v
Go proxy and embedded dashboard :8080
        |
        +--> PostgreSQL: durable URL mappings
        |
        +--> Redis: redirect cache and metrics
```

### Components

| Component | Responsibility |
| --- | --- |
| `cmd/proxy` | Load configuration, connect to dependencies, wire services, and start HTTP. |
| `internal/config` | Parse JSON configuration and apply safe defaults. |
| `internal/shortcode` | Validate URLs, normalize query strings, and encode numeric IDs using Base62. |
| `internal/service` | Coordinate idempotent shortening and cache-aside resolution. |
| `internal/repository` | Define storage interfaces and implement PostgreSQL/Redis adapters. |
| `internal/handler` | Expose JSON APIs, redirect routes, health, metrics, and the embedded dashboard. |
| `internal/metrics` | Provide concurrency-safe counters for dashboard and operational visibility. |
| `web/static` | Embedded HTML, CSS, JavaScript, and local Chart.js bundle. |

## 3. API contract

### `GET /`

Returns the embedded dashboard.

### `GET /health`

Returns `200 OK` when PostgreSQL and Redis are reachable. Returns `503 Service Unavailable` with dependency details when either dependency is unavailable.

### `POST /api/v1/shorten`

Request:

```json
{
  "url": "https://example.com/path?a=2&b=1",
  "mode": "full"
}
```

`mode` is either `full` or `path_only`; an omitted mode uses the configured default, which is `full`.

Successful response:

```json
{
  "short_url": "http://localhost:8080/shorturl/abc123",
  "short_code": "abc123",
  "original_url": "https://example.com/path?a=2&b=1",
  "normalized_url": "https://example.com/path?a=2&b=1",
  "mode": "full",
  "existing": false
}
```

If the normalized URL and mode already exist, return the existing record with `existing: true`.

### `GET /api/v1/resolve/{shortCode}`

Returns the mapping as JSON. Unknown codes return `404 Not Found`.

### `GET /shorturl/{shortCode}`

Resolves the code and returns `302 Found` with a `Location` header pointing to the original URL. The temporary redirect keeps the service in the request path for observable repeat visits.

### `GET /api/v1/metrics`

Returns counters for shorten requests, resolve requests, redirects, cache hits, cache misses, errors, and the current in-process request total.

## 4. URL normalization

Use `net/url` and reject anything that is not an absolute `http` or `https` URL with a host.

- `full`: preserve the query string after removing configured blacklist keys and sorting keys and duplicate values deterministically.
- `path_only`: remove `RawQuery` before lookup and storage.
- Preserve the URL fragment because the requirement only removes query parameters; note that browsers do not send fragments to the server.
- Use `(normalized_url, mode)` as the idempotency key.
- Use `url.Values.Encode()` after sorting values so equivalent query parameter order produces one record.

The default configuration contains an empty blacklist. Example blacklisted keys can include tracking parameters such as `utm_campaign`; users can replace the list in `config.json`.

## 5. Storage model

PostgreSQL initializes this table on startup:

```sql
CREATE TABLE IF NOT EXISTS urls (
    id BIGSERIAL PRIMARY KEY,
    short_code TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    normalized_url TEXT NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('full', 'path_only')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (normalized_url, mode)
);
```

The `id` is an internal surrogate key and `short_code` is its Base62 encoding. ID generation and code obfuscation are out of scope for this demo.

Redis keys:

- `url-shortener:code:{shortCode}` stores the original URL with the configured TTL.
- `url-shortener:metrics:{name}` stores optional shared counters.

The Go metrics collector remains available when Redis is unavailable so the dashboard can still report local process activity; health continues to expose dependency failure.

## 6. Request flows

### Shorten

1. Decode and validate the JSON request.
2. Select the requested mode or configured default.
3. Normalize the URL.
4. Query PostgreSQL for `(normalized_url, mode)`.
5. Return the existing mapping when found.
6. Otherwise reserve a sequence ID, Base62-encode it, and insert the mapping.
7. Return the generated short URL.

The repository uses a transaction for ID allocation and insertion. A unique-conflict retry reads the winning existing record, making concurrent duplicate submissions idempotent.

### Resolve

1. Check Redis for the short code.
2. On a hit, increment cache-hit metrics and return the URL.
3. On a miss, query PostgreSQL.
4. Populate Redis with the configured TTL.
5. Increment cache-miss metrics and return the URL, or return `404` if absent.

### Redirect

Resolve the code, record redirect metrics, and send `302 Found`. Invalid or missing codes return JSON errors instead of redirecting.

## 7. Configuration

`config.json` is parsed with the standard library:

```json
{
  "server": { "port": 8080 },
  "postgres": {
    "host": "postgres",
    "port": 5432,
    "database": "urlshortener",
    "user": "urlshortener",
    "password": "urlshortener"
  },
  "redis": { "address": "redis:6379", "password": "", "db": 0 },
  "cache_ttl": "1h",
  "default_mode": "full",
  "query_blacklist": ["utm_campaign", "utm_medium", "utm_source"]
}
```

## 8. Docker deployment

`docker-compose.yml` runs:

- `proxy` on host port `8080`.
- `postgres` using `postgres:16-alpine` with a named data volume.
- `redis` using `redis:7-alpine` with a named data volume.

The proxy waits for dependency health checks. `Dockerfile` uses `golang:1.26-alpine` to compile a static binary and `alpine:3.19` for the non-root runtime image.

## 9. Dashboard

The embedded dashboard provides:

- URL shortening form with `full` and `path_only` modes.
- Display of the short URL, normalized URL, and whether the mapping already existed.
- Resolve form for the original URL.
- Direct redirect link at `/shorturl/{shortCode}`.
- Health/dependency status.
- Offline Chart.js charts for operation counts and cache hit/miss rates.

Chart.js is committed as a local browser asset; no CDN or frontend build is required at runtime.

## 10. Testing and acceptance

All tests use table-driven cases where applicable. Unit tests use fakes for repositories and caches, so normal test execution does not require Docker.

Required coverage:

- Base62 boundaries and deterministic output.
- URL validation, blacklist filtering, query sorting, duplicate values, and both modes.
- Existing-record idempotency and new-record creation.
- Service cache hit/miss behavior.
- API validation, JSON responses, not-found behavior, redirects, health, and metrics.
- Concurrent access under the race detector.

Validation:

```bash
gofmt -w .
go test -race ./...
go vet ./...
docker compose config
docker compose up --build -d
curl http://localhost:8080/health
curl http://localhost:8080/
```

No separate e2e Go package is required for this demo; manual Compose smoke checks avoid triggering the repository’s specialized e2e `REPORT.md` requirements.
