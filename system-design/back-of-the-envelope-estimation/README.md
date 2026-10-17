# Back-of-the-Envelope Estimation

Back-of-the-envelope estimation is a fast way to test whether a system design can meet its capacity, latency, storage, and availability requirements. The goal is not exact arithmetic; it is to make explicit assumptions, use sensible approximations, and identify the scale of the problem before choosing components.

## TL;DR

### 1. Use powers of two for data sizes

Distributed-system calculations become easier when data sizes are expressed with familiar binary magnitudes. As a rough mental model, $2^{10}$ is about 1 thousand, $2^{20}$ is about 1 million, $2^{30}$ is about 1 billion, $2^{40}$ is about 1 trillion, and $2^{50}$ is about 1 quadrillion. These correspond approximately to KB, MB, GB, TB, and PB respectively.

Remember to account for the size of each field, record overhead, indexes, replication, and retention when turning an object estimate into a storage estimate.

### 2. Keep latency orders of magnitude in mind

The relative gap between operations is more useful in an interview than memorizing every exact number. Cache and memory access are measured in nanoseconds, compression and small network transfers in microseconds, data-center round trips and disk operations in milliseconds, and cross-continent round trips in hundreds of milliseconds.

This leads to practical design rules:

- Prefer memory and sequential access over disk seeks on hot paths.
- Use simple compression when reducing network transfer costs is worth the CPU time.
- Keep latency-sensitive services and their dependencies close together.
- Treat cross-region calls as a major latency and availability boundary.

### 3. Translate availability into downtime

Availability is easier to reason about when converted from a percentage into an allowed outage budget. Approximate annual downtime is:

| Availability | Downtime per day | Downtime per month | Downtime per year |
| --- | ---: | ---: | ---: |
| 99% | 14.4 minutes | 7.31 hours | 3.65 days |
| 99.99% | 8.64 seconds | 4.38 minutes | 52.60 minutes |
| 99.999% | 864 milliseconds | 26.30 seconds | 5.26 minutes |
| 99.9999% | 86.4 milliseconds | 2.63 seconds | 31.56 seconds |

More nines require stronger redundancy, failover, monitoring, testing, and operational discipline. An SLA should therefore influence the architecture rather than appear only as a final number.

### 4. Estimate traffic from explicit assumptions

Start with a small set of stated inputs: total users, active-user percentage, actions per active user, object size, retention period, and a peak-to-average multiplier. Write down units for every value.

For example, suppose a service has 300 million monthly active users, 50% daily-active users, and two posts per daily-active user:

$$
\text{DAU} = 300\text{ million} \times 50\% = 150\text{ million}
$$

$$
\text{Average write QPS} = \frac{150\text{ million} \times 2}{86{,}400} \approx 3{,}500
$$

$$
\text{Peak write QPS} \approx 2 \times \text{average QPS} \approx 7{,}000
$$

If 10% of posts contain a 1 MB media object and media is retained for five years:

$$
\text{Daily media} = 150\text{ million} \times 2 \times 10\% \times 1\text{ MB} \approx 30\text{ TB/day}
$$

$$
\text{Five-year media} = 30\text{ TB} \times 365 \times 5 \approx 55\text{ PB}
$$

The values are illustrative, not production measurements. The important result is the order of magnitude: thousands of requests per second and petabytes of retained media imply different storage, partitioning, caching, and operational choices than a small single-server workload.

### 5. Use a repeatable interview process

- Clarify the scope and what must be estimated: QPS, peak QPS, storage, cache capacity, bandwidth, or server count.
- State assumptions before calculating and revisit them if the result seems implausible.
- Round aggressively. Replace awkward values such as 99,987 / 9.1 with approximately 100,000 / 10.
- Label every number with its unit so bytes, seconds, requests, and users are not mixed accidentally.
- Calculate average load first, then apply a stated peak factor.
- Add replication, indexes, metadata, compression effects, and headroom when the estimate is used to size infrastructure.
- Use the estimate to compare designs and expose bottlenecks; precision beyond the assumptions is usually false confidence.

## A practical checklist

1. Define the workload and retention window.
2. Write down user, request, object-size, and peak assumptions.
3. Convert daily totals into per-second rates.
4. Estimate storage, bandwidth, cache, and compute separately.
5. Apply replication and safety margins.
6. Compare the result with the rough capacity and latency of the proposed components.
7. Explain which assumptions would most change the design.

Source: [ByteByteGo — Back-of-the-Envelope Estimation](https://bytebytego.com/courses/system-design-interview/back-of-the-envelope-estimation)
