# Task Tracker

## 2026-11-15

- Added the `design-a-url-shortner` topic with a ByteByteGo-based URL-shortener TL;DR and updated the system-design index.
- Created the reusable implementation skill, rewrote the design contract, and implemented the Go, Redis, PostgreSQL, Docker, dashboard, and table-driven test demo.
- Audited the demo for secrets (only local throwaway credentials) and fixed embedded dashboard assets returning 404 by adding a static-file fallback with table-driven tests.
- Documented the enumeration risk of sequential codes in `design.md` §5.1 and specified a keyed Feistel permutation instead of raw `Base62(id)`; rejected the nanosecond-timestamp approach with capacity math.
- Reverted the `design.md` §5.1 ID-generation design, including the Feistel/key-management references in §1, §2, §6, §7, and §10, after deciding ID generation is out of scope; the demo keeps `Base62(id)`.
