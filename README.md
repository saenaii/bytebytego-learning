# ByteByteGo Learning

A hands-on collection of system-design study notes and runnable distributed-systems examples inspired by [ByteByteGo](https://bytebytego.com/).

## Projects

### [System Design](./system-design/)

The main learning project in this repository. It currently includes:

- Scaling a web application from a single server to millions of users
- Rate limiting with token bucket, leaking bucket, and adaptive BBR algorithms
- Consistent hashing with virtual nodes
- A distributed key-value store with replication, quorum operations, vector clocks, and Merkle-tree repair

Each topic is self-contained and may include design notes, Go implementation code, Docker workflows, tests, and task tracking. See the [system-design README](./system-design/README.md) for the topic index and learning path.

## Getting started

Clone the repository and enter the project you want to explore:

```bash
git clone git@github.com:saenaii/bytebytego-learning.git
cd bytebytego-learning/system-design
```

Then choose a topic and follow its README. Runnable Go topics generally provide Makefile targets such as:

```bash
make test
make build
```

Some topics also support Docker Compose and end-to-end test workflows; use the topic-specific documentation for details.

## Repository conventions

- Keep each system-design topic self-contained under its project directory.
- Put topic-specific task checklists and conversation summaries in `tasks/tracker.md`.
- Prefer table-driven Go tests and run the validation commands described by the topic.
- Treat the examples as educational material rather than production-ready services.
