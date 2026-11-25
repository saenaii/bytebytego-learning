# Rate Limiter — Staff Engineer Mock Interview

> A practice transcript for explaining rate-limiter design at staff level. The candidate should make assumptions explicit, connect design choices to product requirements, and communicate trade-offs. The repository project is a lightweight, in-process Go demo; production-scale architecture discussed here is an extension, not a guarantee provided by that demo.

## Interview Setup

**Interviewer:** We’re going to design a rate limiter for a public API. I’m interested in how you clarify requirements, make trade-offs, and lead the design—not just which algorithm you can name. Please think aloud and call out assumptions.

**Candidate:** Sounds good. I’ll first establish the traffic and policy requirements, then propose the request path and state model. I’ll compare algorithm choices and finish with distribution, failure handling, and operations. I’ll keep hard guarantees separate from best-effort behavior.

## 1. Problem Framing and Requirements

**Interviewer:** What problem are we solving, and what would you clarify before designing it?

**Candidate:** We need to control how frequently a client can consume a protected resource. That can protect service capacity, reduce abuse, and allocate capacity fairly across customers. I’d clarify:

- What is the identity key: API key, account, user, IP, route, or a combination?
- Are limits per second, minute, or another window? Do clients need short bursts?
- Is the goal strict enforcement, approximate enforcement, or overload protection?
- What are peak request rate, number of active identities, latency budget, and availability target?
- Do policies vary by plan, endpoint, or region, and how quickly must policy changes take effect?
- Should a rejected request receive a retry hint, and are there contractual fairness requirements?

Until we have numbers, I’ll assume a multi-tenant API with per-account limits, high request volume, low added latency, and a tolerance for small cross-region discrepancies. I’d confirm that last assumption because it determines whether we can use regional counters or need stronger coordination.

**Interviewer:** What are the core functional and non-functional requirements?

**Candidate:** Functionally, the service needs to decide whether to admit a request under the applicable policy and return a useful rejection response. Operationally, it needs policy updates, health signals, and metrics for allowed and rejected traffic. Non-functionally, it should have bounded decision latency, scale with traffic and active keys, avoid becoming a single point of failure, and define behavior when its state store or policy service is unavailable. We should also avoid allowing callers to spoof the identity used for enforcement.

## 2. API and Request Flow

**Interviewer:** Sketch the API and the request flow.

**Candidate:** I’d usually put the decision in gateway middleware rather than require an application-level round trip. The logical decision API could be:

```text
checkLimit(identity, policy, cost) -> allowed, remaining, retryAfter
```

At the edge, trusted authentication produces the identity. The middleware resolves the policy, computes the cost (one request by default, possibly weighted for expensive operations), and asks the limiter for an atomic decision. If allowed, it forwards the request; otherwise it returns HTTP 429 with a retry hint where we can estimate one. The exact response contract should be consistent across endpoints.

I’d avoid trusting an arbitrary client-supplied user ID or forwarding header. For unauthenticated traffic, an IP-based key may be a fallback, but NATs and proxies make it an imperfect fairness boundary.

**Interviewer:** Would you call a centralized service on every request?

**Candidate:** It’s simple to reason about, but a synchronous central call adds network latency and creates a dependency on that service for every API request. I’d compare three models:

1. **Central atomic decision:** strongest shared enforcement, but adds a network hop and can bottleneck.
2. **Regional or sharded counters:** lower latency and better locality, with bounded inconsistency or more complex routing.
3. **Leased token allocation to gateways:** gateways make local decisions cheaply, while a coordinator grants token batches. This reduces per-request coordination but can temporarily skew usage and strand unused tokens.

The choice follows the consistency requirement. For overload protection, approximate regional enforcement is often adequate. For contractual quotas or costly operations, I’d use tighter coordination or reserve capacity centrally for the high-risk actions.

## 3. Algorithm Selection

**Interviewer:** Which rate-limiting algorithm would you choose?

**Candidate:** There isn’t one universally best algorithm; it depends on burst tolerance, smoothness, state cost, and fairness.

- A **token bucket** accumulates tokens up to a capacity and spends one or more tokens per request. It’s a strong default when short bursts are acceptable but average consumption should remain bounded. Capacity sets the burst allowance; refill rate sets the sustained rate.
- A **leaking bucket** can buffer accepted work and drain it at a steady rate, which smooths downstream output. The queue adds latency and requires a clear overflow policy. If the API is synchronous, we need to distinguish queueing work from merely permitting a request.
- A **fixed window** is easy and cheap, but clients can send nearly twice the nominal limit around a boundary. A **sliding window counter or log** reduces that boundary effect at additional state and computation cost.
- **Adaptive congestion control** can respond to changing capacity, but it is harder to explain and tune as a customer quota mechanism. The repository’s BBR-inspired implementation is an educational adaptive limiter; it should not be treated as a production BBR implementation or a substitute for explicit tenant policy without validating its behavior.

I’d start with token bucket for a general API limit, then validate burst and fairness requirements with traffic data. I’d use separate overload protection or concurrency limits when the protected resource is sensitive to in-flight work, because request rate alone does not bound concurrency or request cost.

**Interviewer:** How would you tune a token bucket?

**Candidate:** I’d define the sustained rate from the service budget and tenant entitlement, then choose capacity from the burst we can safely absorb. For example, if a tenant is entitled to 100 requests per second, a capacity of 100 permits a one-second burst, but only if downstream capacity can absorb that spike. I’d model the refill continuously or in appropriately granular increments, use a monotonic clock, clamp tokens at capacity, and make the check-and-decrement atomic. For weighted requests, consume the request’s cost and reject if fewer tokens remain.

**Interviewer:** What does the repository implement, and what should we not infer from it?

**Candidate:** The repository is a local Go service exposing `/allow`, `/health`, and `/stats`, with configurable Token Bucket, Leaking Bucket, and BBR-inspired options. The implementations are in-process and use synchronization for concurrent access. That is useful for learning and testing algorithm behavior. It does not provide shared state across replicas, distributed fairness, global quotas, or production-grade adaptive control. Running multiple instances would give each instance its own limiter state unless we add coordination or deliberately divide the budget.

## 4. State, Distribution, and Capacity

**Interviewer:** How would you scale this to multiple gateway instances?

**Candidate:** First I’d decide whether limits are global or regional and quantify acceptable error. If each gateway independently enforces the full quota, aggregate traffic can exceed the limit roughly in proportion to the number of gateways. A shared atomic counter avoids that over-admission but adds a hot-path dependency.

A common compromise is partitioning by identity so one shard owns the decision state for a key, or leasing token batches to gateways. With leases, the coordinator grants a bounded number of tokens and the gateway enforces locally. That gives an explicit upper bound on overshoot from outstanding leases, but it can waste capacity when gateways are unevenly loaded. I’d monitor lease utilization and rebalance based on observed demand.

For a distributed counter, I’d store enough state to compute refill or window expiration and perform the read-update decision atomically. I’d shard by a stable hash of the identity, size shards using active-key cardinality and request skew, and plan for hot tenants. A single large customer can dominate one key, so we may need sub-partitioned local budgets or dedicated treatment while preserving that tenant’s total limit.

**Interviewer:** How would you estimate capacity?

**Candidate:** I’d ask for peak decisions per second, active identities, state bytes per identity, and retention period. As a rough example, 2 million active identities at 100 bytes each is about 200 MB raw state before datastore overhead, replication, and indexes. At 500,000 decisions per second, one centralized decision path is unlikely to be a single small service; I’d benchmark the chosen store and network path, then shard and replicate based on measured throughput and latency. I’d include skew and failover headroom rather than divide evenly by average load. These are starting estimates, not a capacity commitment.

## 5. Consistency and Failure Modes

**Interviewer:** What happens if the state store is unavailable?

**Candidate:** That’s a product and risk decision. **Fail open** preserves API availability but may allow abuse or overload. **Fail closed** protects the service but can block legitimate traffic. I’d use different behavior by endpoint and risk tier: perhaps fail open with a conservative local emergency limit for low-risk reads, and fail closed for expensive or security-sensitive operations. The decision should be explicit in policy, and we need an alert when the fallback activates.

If we use local fallback buckets, I’d cap their capacity so every gateway cannot independently grant the full global quota without bound. We should also ensure retries do not amplify load during a store outage. Timeouts, bounded retries, circuit breaking, and load shedding matter as much as the data structure.

**Interviewer:** What about clock skew and duplicate requests?

**Candidate:** Use monotonic time for elapsed intervals within a process. For distributed windows, prefer datastore time or a well-defined time source and tolerate small skew in the policy. The limiter decision itself should be atomic, so concurrent requests for one identity cannot both spend the same token. If the client retries after an ambiguous timeout, it may consume quota twice. If that is unacceptable, add an idempotency key or make the operation naturally idempotent; that introduces storage and retention costs, so I wouldn’t add it to every ordinary rate-limit check by default.

**Interviewer:** How do you handle policy changes?

**Candidate:** Version policies and distribute updates with a defined propagation target. Gateways can cache policies with a bounded TTL and refresh asynchronously. For a stricter limit, we need to decide whether enforcement starts immediately or after propagation; promising instantaneous global changes while using cached regional state would be misleading. Roll out new policies gradually where possible, validate their effect, and keep a rollback path.

## 6. Fairness, Security, and Client Experience

**Interviewer:** How do you prevent a noisy neighbor from harming others?

**Candidate:** Use per-tenant keys and avoid a single global bucket as the only control. Add service-wide protection as a separate layer, then allocate tenant budgets within that envelope. Policy tiers should be explicit, and we should consider weighted costs for endpoints with very different resource consumption. We should also guard against high-cardinality key attacks: validate identity sources, cap state creation, expire inactive keys, and avoid using raw untrusted strings as unbounded storage keys.

**Interviewer:** What should clients see when throttled?

**Candidate:** A stable 429 response with a machine-readable error and, when meaningful, a retry delay. We should avoid giving false precision if distributed state makes the estimate approximate. Documentation should tell clients to use exponential backoff with jitter and not retry immediately in a tight loop. We can expose remaining quota when doing so is useful and safe, but headers should be consistent and treated as guidance if the enforcement is approximate.

## 7. Observability and Operations

**Interviewer:** What would you monitor?

**Candidate:** I’d monitor both system health and policy outcomes:

- Decision rate, allow/reject rate, and rejection ratio by route and policy tier.
- Added latency distributions, timeout rate, datastore latency, and error rate.
- Active key count, state memory, shard utilization, hot-key concentration, and eviction rate.
- Fallback activations, policy propagation delay, lease utilization, and estimated overshoot.
- Downstream saturation and request cost, so we can tell whether the limiter is protecting the resource it was designed to protect.

I’d avoid tenant IDs as unbounded metric labels. Use sampled or access-controlled diagnostics for per-tenant investigation, while aggregate metrics remain cardinality-bounded. Alerts should distinguish a sudden traffic spike from a limiter dependency failure or an accidental policy rollout.

**Interviewer:** How would you test the design before rollout?

**Candidate:** Unit-test algorithm invariants and boundary cases: full and empty buckets, refill, exact capacity, weighted cost, concurrent decisions, and clock behavior. Integration tests should exercise the HTTP response contract and configuration. Then load-test realistic key distributions, including skew and bursts, and inject state-store latency, timeouts, failover, and policy-update delays. I’d compare observed overshoot and decision latency to explicit targets. For rollout, shadow decisions first, then canary enforcement for a small tenant cohort, watch rejection and service health, and expand gradually with a fast rollback.

The repository includes unit and end-to-end tests for its local service; those validate that demo’s behavior. They do not by themselves prove distributed correctness, production capacity, or resilience under a shared-store outage.

## 8. Staff-Level Trade-Offs and Closing

**Interviewer:** What is the most important trade-off in your design?

**Candidate:** It’s the balance between enforcement accuracy and request-path latency and availability. A globally coordinated counter can enforce a shared quota more precisely, but every request inherits the coordination service’s latency and failure modes. Local token leases reduce latency and dependency pressure but permit bounded overshoot and can waste budget. I’d choose based on the business cost of overshoot versus the cost of rejecting or slowing legitimate traffic, and make that contract visible to API owners.

**Interviewer:** Summarize your proposed design in a minute.

**Candidate:** I’d start with an edge middleware that authenticates the caller, derives a trusted tenant key, resolves a versioned policy, and makes an atomic token-bucket decision. For the common case, I’d use regional or leased local budgets to keep the request path fast, with a centrally managed allocation that bounds overshoot. I’d add separate concurrency and service-wide overload controls where request rate is not enough. We’d define fail-open or fail-closed behavior per risk tier, return consistent 429 responses, and monitor decision latency, rejection rates, fallback use, skew, and downstream saturation. I would validate the numbers and traffic shape through shadowing and load tests before enforcing broadly. The repository’s in-process Go implementation is a useful local demonstration of algorithms and API behavior, but distributed state, policy propagation, and production resilience require additional design and validation.

## Interviewer Evaluation Guide

A strong staff-level performance should:

- Clarify identity, scope, traffic, burst, latency, availability, and consistency before selecting an algorithm.
- Explain why token bucket is a reasonable default while naming when a queueing bucket, window-based limiter, concurrency limit, or adaptive control is more appropriate.
- Distinguish per-process state from globally coordinated state and quantify the implications of replication, sharding, leases, and hot keys.
- Make failure behavior and the cost of fail-open versus fail-closed explicit.
- Address security, fairness, policy rollout, observability, testing, and staged rollout.
- Communicate assumptions and trade-offs plainly, and avoid claiming guarantees that the design has not established.
