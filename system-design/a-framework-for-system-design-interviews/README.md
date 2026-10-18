# A Framework for System Design Interviews

## One-sentence summary

A strong system design interview is a collaborative conversation: clarify the problem, agree on a simple high-level design, explore the most important details, and close by discussing trade-offs, failure modes, and future scale.

## TL;DR

- Do not rush to a solution. Ask questions until the product requirements, scale, constraints, and success criteria are clear.
- Propose a high-level blueprint and get the interviewer’s agreement before spending time on implementation details.
- Deep dive into the components that matter most for the stated requirements, such as latency, data modeling, fanout, consistency, or availability.
- Think aloud, state assumptions, compare alternatives, and invite feedback throughout the interview.
- Finish with bottlenecks, failure handling, observability, rollout concerns, and the next scale curve.

## The four-step framework

### 1. Understand the problem and establish scope

Start by turning an ambiguous product prompt into a concrete engineering problem. Clarify:

- Which features are in scope and which are intentionally excluded.
- Who uses the system and what the main user flows are.
- User count, traffic, data size, growth expectations, and latency targets.
- Existing services or technology constraints that can simplify the design.
- Important quality attributes such as availability, consistency, privacy, or cost.

If the interviewer asks you to make assumptions, write them down. Assumptions are part of the design contract and give you something to revisit when requirements change.

### 2. Propose a high-level design and get buy-in

Draw a simple blueprint before opening a component-level discussion. Depending on the problem, it may include clients, DNS, a load balancer, API services, caches, databases, a CDN, and a message queue.

Walk through one or two important use cases end to end. For example, a news feed may have separate publishing and feed-retrieval flows. Showing the request path helps expose missing components and edge cases.

Explain why each major component exists, then ask for feedback. Treat the interviewer as a teammate and confirm that the blueprint addresses the agreed requirements before continuing.

### 3. Deep dive into the most important parts

Use the interviewer’s feedback and the requirements to choose the deep-dive topics. Do not attempt to specify every component equally. Prioritize the parts that demonstrate the key engineering decisions, such as:

- Data model, partition key, indexes, and access patterns.
- Read/write flow, cache behavior, and consistency trade-offs.
- Fanout, asynchronous processing, queues, and backpressure.
- Latency budgets, capacity estimates, and likely bottlenecks.
- Failure handling, replication, retries, idempotency, and recovery.

The right depth depends on the problem. API contracts and schemas may be useful for a focused game backend, while a broad search-engine design may need more attention on scale and system boundaries.

### 4. Wrap up with trade-offs and operations

Reserve time to summarize the design and discuss what happens beyond the happy path:

- Where are the current bottlenecks and how would you improve them?
- What happens when a server, dependency, network link, or data store fails?
- Which metrics, logs, and alerts show that the system is healthy?
- How would you deploy, roll back, migrate, or test the system safely?
- What changes are needed for the next scale curve, such as ten times more users?

There is always room for improvement. A thoughtful limitation is more useful than claiming the design is perfect.

## How the pieces fit together

The framework is a loop rather than four isolated presentations:

```text
Clarify requirements
        ↓
Agree on a high-level design
        ↓
Estimate and deep dive into the critical path
        ↓
Review trade-offs, failures, and operations
        ↺ revise assumptions or the design when new constraints appear
```

Back-of-the-envelope estimation belongs after the first blueprint, when there is enough context to estimate the important flows but before committing to an over-detailed design.

## Key concepts

| Concept | Simple explanation | Why it matters |
| --- | --- | --- |
| Scope | The features and constraints the design must address | Prevents solving the wrong problem |
| Assumption | A value or condition supplied by the candidate when requirements are unknown | Makes reasoning explicit and revisable |
| High-level design | The main components and request/data flows | Creates shared context before the deep dive |
| Deep dive | Detailed analysis of the components most affected by the requirements | Shows technical judgment without wasting time |
| Buy-in | Agreement from the interviewer that the direction is reasonable | Keeps the discussion collaborative and aligned |
| Trade-off | A deliberate balance between competing goals such as latency, cost, and consistency | Demonstrates practical design thinking |

## Time allocation for a 45-minute interview

Use this only as a rough guide; the problem’s scope and the interviewer’s priorities should determine the actual split.

| Step | Suggested time |
| --- | ---: |
| Understand the problem and establish scope | 3–10 minutes |
| Propose the high-level design and get buy-in | 10–15 minutes |
| Deep dive | 10–25 minutes |
| Wrap up | 3–5 minutes |

Keep checking the clock. A detailed explanation of one component is not useful if it leaves no time to explain the overall design or its failure modes.

## Do and do not

### Do

- Ask for clarification instead of silently guessing.
- Communicate your reasoning and label assumptions.
- Offer alternative approaches when they reveal a meaningful trade-off.
- Design the most critical components first.
- Ask for feedback and hints when the discussion gets stuck.
- Prepare for common system design problems before the interview.

### Do not

- Jump into a detailed solution before understanding the requirements.
- Over-engineer features that are not needed by the stated scope.
- Spend the opening minutes on one component’s low-level details.
- Think silently for long periods; share the decision process.
- Treat the design as finished immediately after presenting it.

## Practical checklist

1. Restate the problem and confirm the main user flows.
2. Ask about scale, latency, availability, consistency, and constraints.
3. Write down assumptions and define what is out of scope.
4. Draw the simplest high-level architecture that fits the requirements.
5. Walk through a concrete request or data flow and get buy-in.
6. Estimate capacity, then deep dive into the highest-risk components.
7. Explain trade-offs, failures, monitoring, deployment, and future scale.
8. Summarize the design and invite final feedback.

Source: [ByteByteGo — A Framework For System Design Interviews](https://bytebytego.com/courses/system-design-interview/a-framework-for-system-design-interviews)
