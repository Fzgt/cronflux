# Design notes

Some of the decisions behind cronflux, and the trade-offs they carry.

## At-least-once, not exactly-once

Runs are delivered to workers with a **lease**. A worker claims a run, which
sets its state to `running` and stamps a lease expiry. If the worker finishes it
acknowledges the outcome; if it crashes, the lease eventually expires and the
run becomes claimable again.

This gives **at-least-once** delivery: a run can execute more than once (for
example if a worker completes the work but dies before acknowledging). Exactly
-once would require distributed coordination we deliberately avoid. Instead the
contract is simple: **executors should be idempotent** where duplicate
execution would be harmful.

## Why `SELECT … FOR UPDATE SKIP LOCKED`

The PostgreSQL backend claims work with a CTE that selects due rows
`FOR UPDATE SKIP LOCKED` and flips them to `running` in one statement. `SKIP
LOCKED` means concurrent claimers step over rows already locked by another
transaction instead of blocking on them, so N workers (or N nodes) pull
disjoint batches with no coordination and no thundering herd. It is the standard
way to build a queue on top of PostgreSQL.

## Materialisation and the priming tick

The dispatcher tracks each job's next fire time in memory. The **first** time it
sees a job it schedules the next fire in the future rather than firing
immediately, so restarting the server does not backfill a burst of runs for
every past slot. Enqueueing is idempotent per `(job, scheduled_for)` so a
restart that re-seeds the cache cannot double-materialise a slot.

The cost is that missed slots while the process is down are not backfilled.
For a scheduler this is usually the desired behaviour; a job that must catch up
can be triggered manually.

## DAG semantics

A cron fire creates a **batch**. When a job's run succeeds, the scheduler looks
for jobs that depend on it and enqueues them into the same batch, but only once
**all** of their dependencies have a succeeded run in that batch. Enqueueing is
serialised so two parents completing at once cannot create duplicate children.
This keeps the DAG logic in the scheduler, where it is unit-testable, and out of
the store.

## Cron representation

Each field is compiled to a 64-bit mask of permitted values, and `Next` walks
the calendar field by field from most to least significant. Bitmasks make the
per-tick "does this fire now?" check trivial, and the walk naturally handles
awkward cases such as `31` in short months and February 29th, bounded by a
five-year search horizon so an impossible spec terminates.

## Things intentionally left out

- **No persistence of the fire-time cache.** It is rebuilt on startup.
- **No exactly-once.** See above.
- **No leader election.** With the PostgreSQL backend every node is symmetric;
  `SKIP LOCKED` handles contention.
