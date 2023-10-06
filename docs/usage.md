# Usage

## Running the server

```sh
cronflux -addr :8080 -workers 8
```

With PostgreSQL:

```sh
cronflux -store postgres \
  -database-url 'postgres://cronflux:cronflux@localhost:5432/cronflux?sslmode=disable'
```

The schema is created automatically on startup.

## Defining jobs in a file

Point `-jobs` at a JSON file and cronflux upserts every job on startup.
Durations use Go's syntax (`30s`, `5m`, `1h`).

```json
{
  "jobs": [
    {
      "id": "extract",
      "name": "Extract",
      "spec": "0 * * * *",
      "command": ["./extract.sh"],
      "max_retries": 3,
      "backoff": { "base": "5s", "max": "5m", "factor": 2, "jitter": 0.2 },
      "enabled": true
    },
    {
      "id": "transform",
      "name": "Transform",
      "command": ["./transform.sh"],
      "depends_on": ["extract"],
      "enabled": true
    }
  ]
}
```

Here `transform` has no schedule of its own: it runs only after `extract`
succeeds, forming a two-step DAG.

## Cron syntax

```
┌─────────── minute        (0-59)
│ ┌───────── hour          (0-23)
│ │ ┌─────── day of month  (1-31)
│ │ │ ┌───── month         (1-12 or JAN-DEC)
│ │ │ │ ┌─── day of week   (0-6 or SUN-SAT, 0 = Sunday)
│ │ │ │ │
* * * * *
```

Supported in each field: `*`, single values, `a-b` ranges, `a,b,c` lists and
`*/n` or `a-b/n` steps. A leading sixth field adds seconds. Shortcuts:
`@yearly`, `@monthly`, `@weekly`, `@daily`, `@hourly` and `@every <duration>`.

Examples:

| Spec            | Meaning                            |
| --------------- | ---------------------------------- |
| `*/15 * * * *`  | every 15 minutes                   |
| `0 9 * * 1-5`   | 09:00 on weekdays                  |
| `0 0 1 * *`     | midnight on the 1st of every month |
| `@every 90s`    | every 90 seconds                   |
| `30 0 2 * * *`  | 02:00:30 every day (with seconds)  |

## Retries

When a run fails it is retried up to `max_retries` times. The delay before
attempt _n_ is `base * factor^n`, capped at `max`, optionally jittered. After
the last attempt the run is marked `dead`.

## Graceful shutdown

On `SIGINT`/`SIGTERM` cronflux stops accepting new HTTP connections, lets the
current tick finish and drains in-flight requests before exiting. Runs still
executing are left leased so they are redelivered rather than lost.
