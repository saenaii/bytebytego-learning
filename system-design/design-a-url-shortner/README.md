# Design a URL Shortener

## One-sentence summary

A URL shortener maps a long URL to a compact, unique alias and redirects alias requests to the original URL with low latency and high availability.

## TL;DR

- The core product has two operations: create a short URL from a long URL, then redirect a short URL to its long URL.
- A useful baseline assumes 100 million new URLs per day, about 1,160 writes per second, and a 10:1 read-to-write ratio, or roughly 11,600 redirects per second.
- Over ten years, that workload produces about 365 billion mappings. With an average 100-byte URL, the raw URL data is roughly 36.5 TB before indexes, replication, and operational overhead.
- Store the durable `shortURL -> longURL` mapping in a relational database. Use a cache on the redirect-heavy read path to avoid hitting the database for every click.
- Base62 encoding turns a unique numeric ID into a short code using digits, lowercase letters, and uppercase letters. Seven Base62 characters can represent about 3.5 trillion values, enough for the estimated workload.
- The web tier should be stateless and sit behind a load balancer so it can scale horizontally. Add rate limiting to protect URL creation and consider analytics, replication, and sharding as the service grows.

## Scope and assumptions

The service supports two primary use cases:

1. Given a long URL, return a much shorter URL.
2. Given a short URL, redirect the client to the original URL.

For the baseline design, shortened URLs are not deleted or updated. Codes may contain `0-9`, `a-z`, and `A-Z`, giving an alphabet of 62 characters. The design prioritizes availability, scalability, fault tolerance, and fast redirects.

## Capacity estimates

| Estimate | Calculation | Approximate result |
| --- | --- | ---: |
| New URLs per day | Given workload | 100 million |
| Write QPS | 100 million / 86,400 seconds | 1,160 |
| Read QPS | 1,160 × 10 reads per write | 11,600 |
| Ten-year records | 100 million × 365 × 10 | 365 billion |
| Raw URL storage | 365 billion × 100 bytes | 36.5 TB |

These are planning estimates, not exact production limits. Indexes, database replication, metadata, cache entries, and headroom must be added before sizing infrastructure.

## API design

### Create a short URL

```http
POST /api/v1/data/shorten
Content-Type: application/json

{"longUrl": "https://www.example.com/a/very/long/path"}
```

The response contains the generated short URL.

### Redirect from a short URL

```http
GET /api/v1/shortUrl
```

The service resolves the short code and returns an HTTP redirect whose `Location` header contains the original long URL. A production API would normally place the short code in the path, for example `GET /{shortCode}`.

## Data model

A relational database provides durable storage for the mapping:

| Column | Role |
| --- | --- |
| `id` | Unique numeric identifier used by the code-generation strategy |
| `shortURL` | Unique short code or alias |
| `longURL` | Original destination URL |

The database should enforce uniqueness for the short code. It can also index the long URL if the product wants repeated submissions of the same long URL to return the existing short URL instead of creating a second mapping.

## Generating short codes

### Hash plus collision resolution

Hash the long URL, take a short prefix, and check whether the resulting code already exists. If there is a collision, modify the input with a predefined value and hash again until a free code is found. A Bloom filter can reduce unnecessary database checks, but the database remains the source of truth.

This approach provides a fixed-length code and does not require a numeric ID generator. Its drawbacks are collision handling, repeated checks, and the inability to predict the next available code.

### Base62 conversion

Generate a globally unique numeric ID and encode it using the 62-character alphabet. The smallest code length that supports the estimated 365 billion records is seven characters:

| Code length | Capacity |
| ---: | ---: |
| 6 | 56.8 billion |
| 7 | 3.5 trillion |

Base62 avoids collisions when IDs are unique and makes the next code easy to derive. That predictability can expose approximate creation order, so an implementation may need an ID-obfuscation strategy if enumeration is a security concern.

For this baseline design, Base62 is the simpler choice:

1. Check whether the long URL already has a mapping.
2. If it does, return the existing short URL.
3. Otherwise, obtain a globally unique ID.
4. Convert the ID to Base62.
5. Save the ID, short code, and long URL in one database record.
6. Return the short URL.

## Redirect flow

Redirects are much more frequent than writes, so the hot path uses a cache:

```text
Client
  -> Load balancer
  -> Stateless web server
       -> Cache lookup for short code
          -> cache hit: return long URL
          -> cache miss: query database, populate cache, return long URL
  -> HTTP redirect to destination
```

An unknown short code should return an appropriate not-found response rather than redirecting. The cache needs an expiration policy and a fallback for cache failure; the database remains the durable source.

## 301 versus 302 redirects

- **301 Permanent Redirect:** Browsers and intermediaries can cache the redirect, reducing future load on the shortener. This is useful when minimizing service traffic is the priority.
- **302 Temporary Redirect:** Requests continue to pass through the shortener, making click tracking and source analytics easier. This adds more read traffic to the service.

The choice depends on whether lower redirect traffic or richer analytics is more important. Cache-control behavior should be explicit.

## Scaling and reliability

- **Stateless web tier:** Keep session and mapping state outside web servers so any server can handle any request.
- **Load balancing:** Distribute traffic across web servers and replace unhealthy instances without taking the service offline.
- **Database replication:** Use replicas to improve read capacity and availability; define how promotion and replication lag are handled.
- **Database sharding:** Partition mappings when one database cannot hold or serve the workload. Choose a key that keeps distribution balanced and preserves efficient lookups.
- **Rate limiting:** Limit URL-creation requests by IP, account, or another policy to reduce abuse and protect capacity.
- **Analytics:** Record click events asynchronously when product metrics such as click count, time, and source matter. Keep analytics work off the redirect critical path where possible.
- **Observability:** Track redirect latency, cache hit rate, invalid-code rate, creation failures, database health, and traffic by region.

## Key trade-offs

| Decision | Benefit | Cost or risk |
| --- | --- | --- |
| Hash plus collision resolution | Fixed-length codes; no ID generator | Collision checks and retry logic |
| Base62 from unique IDs | Simple, collision-free code generation | Requires globally unique IDs; IDs may be predictable |
| 301 redirect | Lower repeat load on the shortener | Less visibility into clicks |
| 302 redirect | Better analytics and control | More requests reach the shortener |
| Cache mappings | Faster redirects and lower database load | Stale entries, eviction, and cache-failure handling |
| Relational database | Durable mapping with familiar constraints | May require replication and sharding at larger scale |

## Final design

Use a load balancer in front of stateless web servers. Generate short codes by converting globally unique IDs to Base62, and store the resulting mapping in a replicated relational database. Cache short-code lookups for the read-heavy redirect path. Add rate limiting, analytics, monitoring, and sharding as traffic, abuse risk, and product requirements increase.

Source: [ByteByteGo — Design a URL Shortener](https://bytebytego.com/courses/system-design-interview/design-a-url-shortener)

## Local demo

The topic includes a runnable Go proxy backed by PostgreSQL and Redis. Start the complete stack with:

```bash
docker-compose up --build
```

Open [http://localhost:8080](http://localhost:8080) for the embedded dashboard. The dashboard supports both URL-normalization modes, resolving codes, following redirects, and viewing local Chart.js metrics without a frontend build or CDN.

Useful API examples:

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/article?b=2&a=1","mode":"full"}'

curl http://localhost:8080/api/v1/resolve/<short-code>
curl -i http://localhost:8080/shorturl/<short-code>
```

For unit tests and static checks:

```bash
go test -race ./...
go vet ./...
```
