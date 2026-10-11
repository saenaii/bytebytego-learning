# Scale From Zero to Millions of Users

This chapter presents a practical path for evolving a simple web application into a highly available system that can serve millions of users. The central idea is to scale incrementally: identify the next bottleneck, introduce the smallest architectural change that addresses it, and keep each tier independently scalable.

## TL;DR

### 1. Start with a single server

Keep the web application, database, and cache on one machine while traffic is small. Clients resolve a domain through DNS and send HTTP requests to the server, which returns HTML or JSON.

### 2. Separate the web and data tiers

Move the database to its own server so the web tier and data tier can scale independently. Choose a relational database by default; consider NoSQL when the data is unstructured, extremely large, needs very low latency, or primarily requires serialization and retrieval.

### 3. Scale the web tier horizontally

Vertical scaling (more CPU, memory, or disk) is simple but has hardware, cost, and single-point-of-failure limits. Add multiple web servers behind a load balancer instead. The load balancer distributes traffic, supports failover, and allows healthy servers to be added as demand grows.

### 4. Replicate the database

Use a primary database for writes and one or more replicas for reads. Replication improves read throughput, reliability, and availability. If a replica fails, route reads elsewhere; if the primary fails, promote a replica, while accounting for replication lag and recovery of missing data.

### 5. Add caching and a CDN

Put frequently read, infrequently changed data in a cache to reduce database load and latency. A read-through flow checks the cache first, queries the database on a miss, stores the result, and returns it. Define expiration and eviction policies, plan for stale data and cache failures, and avoid treating volatile cache memory as durable storage.

Use a CDN for static assets such as images, video, CSS, and JavaScript. The CDN serves content from a nearby edge location and fetches missing objects from the origin. Choose TTLs carefully, account for cost, provide an origin fallback, and use invalidation or versioned URLs when assets change.

### 6. Make the web tier stateless

Do not keep session state on individual web servers. Store sessions and other shared state in durable or shared storage so any request can reach any server. Stateless servers remove the need for sticky sessions and make autoscaling, failover, and rolling changes much simpler.

### 7. Expand to multiple data centers

Use GeoDNS or another traffic-routing mechanism to direct users to a nearby healthy data center. Multi-region deployments reduce latency and improve availability, but require deliberate solutions for traffic redirection, cross-region data and cache synchronization, failover behavior, and consistent testing and deployment.

### 8. Decouple work with message queues

Place a durable message queue between producers and consumers for work that does not need to complete during the request. Producers can continue when consumers are unavailable, and each side can scale independently. For example, web servers can enqueue photo-processing jobs while worker processes consume them; scale workers based on queue depth.

### 9. Operate the system with observability and automation

Centralize logs so failures can be searched across servers. Collect host metrics such as CPU, memory, and disk I/O; aggregate tier-level health and performance; and track business metrics such as daily active users, retention, and revenue. Automate builds, tests, deployments, and other repetitive operational work.

### 10. Scale the data tier with sharding

When one database cannot handle the data volume or traffic, shard it across multiple servers using a partition key. A good key distributes data evenly and lets queries route directly to the correct shard. Sharding introduces trade-offs: resharding and data movement, hot or celebrity keys, and cross-shard joins that often require denormalization.

## End-state principles

- Keep the web tier stateless.
- Build redundancy into every tier.
- Cache aggressively where the data and consistency model allow it.
- Serve static assets through a CDN.
- Support multiple data centers and a tested failover path.
- Use queues to decouple independently scalable workloads.
- Shard the data tier when vertical scaling is no longer sufficient.
- Split tightly coupled components into independently deployable services as complexity grows.
- Monitor the system and automate its delivery and operation.

Scaling is iterative rather than a one-time redesign: measure the current bottleneck, improve that layer, and repeat as usage and requirements grow.

Source: [ByteByteGo — Scale From Zero To Millions Of Users](https://bytebytego.com/courses/system-design-interview/scale-from-zero-to-millions-of-users)
