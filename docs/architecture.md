# Architecture

cronflux is a single process that materialises runs from schedules and hands
them to workers, backed by a pluggable store.

```
             +-------------------+
             |     HTTP API      |  REST + dashboard + /metrics
             +---------+---------+
                       |
        +--------------+--------------+
        |          Scheduler          |
        |  +-----------------------+  |
        |  |  dispatcher (cron +   |  |
        |  |  DAG materialisation) |  |
        |  +-----------+-----------+  |
        |              |              |
        |  +-----------v-----------+  |
        |  |     worker pool       |  |  claim -> execute -> ack/retry
        |  +-----------+-----------+  |
        +--------------|--------------+
                       |
             +---------v---------+
             |       Store       |  memory | postgres
             +-------------------+
```

## Components

### cron

The `cron` package parses schedules into per-field bitmasks and computes the
next activation with a calendar walk. It is dependency-free and importable on
its own.

### store

`store.Store` is the persistence contract. Two implementations satisfy it and
share a single conformance suite (`storetest`) so their behaviour cannot drift:

- **memory** — a mutex-guarded map, used by default and in tests.
- **postgres** — durable storage where `ClaimDue` uses
  `FOR UPDATE SKIP LOCKED`, letting many nodes claim disjoint work without
  blocking each other.

### scheduler

The scheduler runs a tick loop. On each tick it:

1. **Materialises** due runs. For cron jobs the dispatcher enqueues a pending
   run per elapsed fire time; the first sighting of a job schedules its next
   fire in the future so startup does not backfill.
2. **Drains** ready work. It claims a batch of runs via the store's lease
   mechanism and processes them across a bounded worker pool.
3. **Publishes** backlog and lag gauges for Prometheus.

### workers

A worker executes one run, records the outcome and, on failure, either
schedules a retry (a fresh pending run with an incremented attempt) or moves the
run to the dead-letter state once retries are exhausted. On success it advances
the DAG by enqueueing any dependents whose upstreams have all succeeded in the
same batch.

## Data model

- **Job** — the definition: id, schedule, command, retry policy, dependencies.
- **Run** — one execution attempt of a job: state, attempt number, lease and
  timestamps. A batch id groups the runs produced by a single DAG trigger.

See [design notes](design-notes.md) for the reasoning behind at-least-once
delivery and the lease model.
