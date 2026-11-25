# Rate Limiter — Staff Engineer Mock Interview

> Use this as an English speaking practice script, not a speech to memorize. The candidate’s answers are intentionally concise and conversational; in a real interview, pause for the interviewer’s response and adapt. The local Go project in this repository is an in-process demo. The production choices below are design proposals, not features the demo already provides.

## A common interview rhythm

There is no single required script for a system-design interview. A useful rhythm is:

1. Clarify the goal and constraints.
2. State assumptions and choose a starting design.
3. Sketch the request flow and explain the main components.
4. Follow up on scale, consistency, failures, and trade-offs.
5. Cover rollout and operations, then summarize.

Treat this as a map, not a checklist. If the interviewer redirects or gives new constraints, update the design out loud. A clear decision with stated assumptions is usually stronger than listing every possible option.

## The interview

### Opening: frame the problem

**Interviewer:** Let’s design a rate limiter for a public API. I’m interested in your reasoning, not just the algorithm. Where would you start?

**Candidate:** Sure. Before I pick an algorithm, I’d like to pin down what we’re limiting and how strict the limit needs to be. Is the main goal to protect the service from overload, enforce customer quotas, or both?

**Interviewer:** Both. What else would you clarify?

**Candidate:** I’d ask who the limit applies to—an account, API key, user, IP address, or some combination. I’d also want peak requests per second, active customer count, burst tolerance, the latency budget, and whether the limit must be global across regions. Those answers change the design quite a bit.

For now, I’ll assume this is a multi-tenant API. We need per-account limits, low added latency, and a small amount of cross-region overshoot is acceptable for ordinary traffic. For expensive operations, we can use a stricter policy. Does that sound like a reasonable starting point?

**Interviewer:** Yes. Assume about 500,000 decisions per second at peak, with a p99 budget of 5 milliseconds for the limiter.

**Candidate:** I’d first confirm whether that 5-millisecond p99 covers the limiter path end to end. If it does, I’d treat same-AZ Redis as a plausible starting point, not a guarantee. As a rough budget, we might reserve about 0.5 ms for network round trips, 1 ms for gateway work, 0.5 ms for Redis plus the script, and 0.1 ms for serialization. That leaves roughly 2.9 ms for queueing, variance, and headroom. Those are planning numbers; we’d measure them in our environment.

A synchronous cross-region check would make that budget risky and leave little room for tail latency. If 5 ms is a hard target across regions, I’d keep decisions regional and use bounded token leases from a coordinator. That trades exact global enforcement for lower request latency and a measurable overshoot bound.

### Build or buy: establish the boundary

**Interviewer:** Would you build this, or use something that already exists?

**Candidate:** My default would be to check the gateway and managed API products we already operate. If they support our identity model, weighted limits, policy updates, observability, and the consistency we need, I’d prefer using them. That reduces the amount of infrastructure we own.

I’d build a separate limiter only if we have a real gap—for example, custom per-customer policies across several gateways, a latency or cost target the existing product can’t meet, or enforcement behavior we need to control. I’d compare the operational cost too: on-call ownership, upgrades, capacity planning, and incident response. Do we already have a gateway product in place?

**Interviewer:** Assume we have a gateway, but its built-in limits are local to each gateway instance. We need a shared account quota.

**Candidate:** Then I’d first check whether the gateway can call a shared decision service or receive centrally allocated token leases. If it can’t, that’s a concrete reason to add a shared component. I’d keep the policy and decision interfaces portable so we can revisit the buy-versus-build choice later.

### Draw the first design

**Interviewer:** Sketch your initial design.

**Candidate:** Let me draw the request path first. I’ll keep the first version simple and call out where coordination happens.

```text
Client
  │
  ▼
API Gateway ── authenticate and derive trusted account ID
  │
  ├── resolve cached, versioned policy
  │
  ▼
Rate-limit decision
  ├── MVP: atomic token-bucket check in shared Redis (Lua)
  └── later option: gateway-local tokens leased by a coordinator
  │
  ├── allowed ──► backend
  └── denied  ──► HTTP 429 + retry guidance

Policy service ──► versioned policy updates ──► gateways
Metrics and logs ◄──── gateway and decision service
```

The gateway should derive the account ID from trusted authentication, not from an arbitrary header. It resolves the policy, asks for an atomic decision, and either forwards the request or returns a 429. I’d include a request cost, defaulting to one, so an expensive endpoint can consume more quota than a cheap read.

**Interviewer:** Why Redis and Lua? Isn’t that just another dependency on every request?

**Candidate:** It is, and that’s the main trade-off. For an MVP, I’d choose a Redis-backed atomic token bucket because it gives us shared per-account state and a compact implementation. A Lua script can refill, check, and decrement in one atomic operation. I’d set a strict timeout and measure the full gateway-to-Redis path against the 5 millisecond p99 target.

I wouldn’t assume Redis will meet the target at 500,000 decisions per second. We’d benchmark with realistic key skew and failover settings. If the network hop or datastore cost misses our target, I’d move the common path to gateway-local tokens allocated in bounded batches. That lowers per-request coordination, but it introduces some overshoot and can leave tokens stranded on lightly used gateways.

**Interviewer:** What would the decision API look like?

**Candidate:** Logically, something like `check(account, policy, cost)` returning `allowed`, `remaining`, and an optional `retryAfter`. The gateway owns the HTTP response. I’d keep the response consistent across routes, and I’d only return a retry estimate when the algorithm can support one accurately enough.

### Pick an algorithm

**Interviewer:** Why a token bucket? What are the alternatives?

**Candidate:** Token bucket is my starting point because it gives us two understandable controls: a sustained refill rate and a maximum burst. It fits an API where brief bursts are okay but sustained usage should stay bounded.

A leaking bucket is useful when we need to smooth work into a downstream system, but queueing adds latency and may not fit a synchronous API. A fixed window is cheaper, but traffic can bunch up around a window boundary. Sliding windows reduce that boundary problem, at some extra state or computation cost. I’d use a concurrency limit as well if the real constraint is in-flight work rather than requests per second.

The repo also has a BBR-inspired adaptive option. I’d describe that as an educational adaptive limiter, not a production BBR implementation or a replacement for explicit customer quotas.

**Interviewer:** How would you set the bucket values?

**Candidate:** I’d set the refill rate from the customer entitlement and our service budget. Then I’d set capacity from the burst we can safely absorb. For example, a rate of 100 requests per second with a capacity of 100 allows about a one-second burst—but I’d only choose that capacity if the backend can handle it. For costly requests, I’d charge more than one token.

**Interviewer:** What if one account sends a huge amount of traffic?

**Candidate:** We need to protect both that tenant’s quota and the shared service. I’d have per-account limits plus a separate service-wide overload guard. There’s a specific issue with a shared Redis Cluster counter: one key maps to one hash slot, and that slot is served by one shard. So a very large tenant can saturate that shard even if the rest of the cluster is mostly idle. A Lua script keeps the check atomic, but it runs on that node’s command execution path and blocks other commands there while it runs; keeping the script short matters.

A dedicated shard can isolate that tenant from others, but it doesn’t split the hot key itself. To spread the work, we could divide the tenant’s budget across sub-keys or lease token batches to gateways. Both approaches need a coordinator or allocation rule to keep the total bounded, and both can allow some overshoot or strand unused budget. I’d choose between them based on the tenant’s traffic and the quota’s accuracy requirement.

### Scale, consistency, and failure behavior

**Interviewer:** At 500,000 decisions per second, how would you think about capacity and cost?

**Candidate:** I’d estimate state and request volume, but I wouldn’t size the cluster from the request rate alone. We need active account count, state size and expiry, key skew, replication overhead, and the cost of failover headroom. I’d benchmark Redis operations per shard with the actual script and traffic distribution.

For cost, I’d compare at least two models: shared decisions on every request, and local decisions using leased token batches. The first spends more on network and datastore operations; the second trades some unused lease capacity and bounded overshoot for fewer calls. We can model the monthly decision count, datastore and network capacity, replication, and operational ownership. I’d use provider pricing and benchmark results for actual dollar estimates rather than invent a number here.

**Interviewer:** How accurate does the quota need to be across regions?

**Candidate:** For ordinary protection, I’d start with regional enforcement or token leases and document the maximum expected overshoot. If this is a contractual hard quota, I’d use stronger coordination for that policy, even if it adds latency. I’d avoid claiming a globally exact limit while each region is making independent decisions.

**Interviewer:** What happens if Redis is unavailable?

**Candidate:** I’d make the behavior explicit by endpoint and risk. For low-risk reads, we might fail open with a small local emergency bucket to protect the backend. For expensive or abuse-sensitive operations, we might fail closed. Either way, we need bounded timeouts, no retry storm, and an alert when fallback starts. The right choice depends on the cost of admitting excess traffic versus rejecting legitimate users.

If we later use leases, a gateway can keep spending its remaining lease during a coordinator outage, but it must stop when the lease expires. That preserves a known bound; it doesn’t provide unlimited availability.

**Interviewer:** And policy changes?

**Candidate:** I’d version policies and define a propagation target. Gateways can cache them, but we should be clear that stricter limits may not take effect everywhere instantly. I’d start with shadow evaluation, then canary a small group, watch rejection rates and backend health, and keep rollback simple.

### Test the design

**Interviewer:** How would you test it before launch?

**Candidate:** I’d cover three layers. First, unit tests for the algorithm boundaries: empty and full buckets, refill, weighted cost, and concurrent requests. Second, load tests with realistic account distributions, including a hot tenant and burst traffic, ramping toward 500,000 decisions per second while measuring p99 latency and rejection accuracy. Third, failure tests: Redis failover, network partition, clock skew, and delayed policy updates.

For each failure, I’d verify the promised behavior, not just that the process stays up. For example: do we fail open or closed as intended, how much can a leased quota overshoot, and how quickly do new policies take effect? I’d run those tests with production-like topology before we call the latency or availability target met.

### MVP and operating the system

**Interviewer:** What would you ship first, and what would you defer?

**Candidate:** For v1, I’d support authenticated per-account token buckets, a shared atomic decision path, a small set of policy fields, consistent 429 responses, and basic metrics. I’d keep the initial deployment in one region if the product allows it.

I’d defer adaptive BBR-style control, a general-purpose policy language, cross-region exact quotas, and tenant-specific custom dashboards. Those add tuning or operational complexity before we’ve learned the real traffic shape. I’d revisit them when product requirements or measurements show the simple version is insufficient.

**Interviewer:** Who owns policies and what happens when a customer says they were throttled incorrectly?

**Candidate:** I’d make ownership part of the design. Product or account operations should own the customer-facing entitlement, while the platform team owns enforcement and safe defaults. Policy changes should be audited and reversible. Support needs a way to inspect a customer’s effective policy and recent aggregate decisions without exposing sensitive data or adding unbounded metric labels.

For an appeal, we should be able to trace which policy version was applied and why a request was rejected. We should also define a response target for investigating false throttling. If quotas are part of a customer SLA, the contract should say whether enforcement is exact, regional, or approximate and how incidents are handled.

**Interviewer:** What would you monitor?

**Candidate:** I’d watch decision latency and errors, allow and reject rates, Redis saturation, active keys, hot-key skew, policy propagation delay, fallback activations, and downstream saturation. I’d keep tenant IDs out of high-cardinality metrics; for investigations, I’d use access-controlled logs or sampled diagnostics. The goal is to tell whether we’re protecting the backend without unexpectedly blocking healthy customers.

### Close with a decision

**Interviewer:** Give me the short version of your design.

**Candidate:** Given what we know, I’d first reuse the gateway for authentication, routing, and policy caching. For v1, I’d use a shared Redis-backed atomic token bucket for per-account quotas, with a strict latency budget and a benchmark gate before launch. I’d start in one region, shadow the decisions, then canary enforcement.

If the shared call misses the p99 target or becomes too expensive, I’d move high-volume policies to bounded token leases at the gateways and accept a measured amount of overshoot. I’d keep a separate service-wide overload control, define fail-open or fail-closed behavior by risk, and make policy ownership and customer appeals operationally clear. The existing repo demo helps explain the algorithms and local API; it doesn’t implement this distributed production design.

## What the interviewer is listening for

A strong staff-level answer should make the interviewer feel that the candidate can lead a decision, not just list components. Look for whether the candidate:

- Clarifies the quota identity, burst behavior, latency target, and required accuracy.
- Checks whether an existing gateway or managed product already solves the problem.
- Chooses a default design under stated assumptions, then names what evidence would change it.
- Uses a diagram to make the request path and coordination points easy to follow.
- Separates a useful v1 from features that can wait.
- Connects cost, reliability, policy ownership, and customer impact to the technical design.
- Distinguishes facts in the repository demo from assumptions about a production system.

## Speaking practice: avoid sounding scripted

Don’t memorize the candidate’s paragraphs. Practice the moves instead:

- Start with a short answer: “My default would be X, assuming Y.”
- Explain one or two reasons, then name the trade-off.
- Invite correction: “Does that match the requirement you had in mind?”
- When new information arrives, update the design: “Given that constraint, I’d change X to Y.”
- Use the diagram as a guide; don’t narrate every box mechanically.

Useful phrases:

- “Let me check the requirement before I choose an algorithm.”
- “I’ll make an assumption for now, and you can correct me.”
- “My starting point would be …”
- “The trade-off is …”
- “I’d validate that with … before calling it a guarantee.”
- “Given that new constraint, I’d change …”
