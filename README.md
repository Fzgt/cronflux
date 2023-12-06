# cronflux

[![CI](https://github.com/Fzgt/cronflux/actions/workflows/ci.yml/badge.svg)](https://github.com/Fzgt/cronflux/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Fzgt/cronflux.svg)](https://pkg.go.dev/github.com/Fzgt/cronflux)
[![Go Report Card](https://goreportcard.com/badge/github.com/Fzgt/cronflux)](https://goreportcard.com/report/github.com/Fzgt/cronflux)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A small distributed job & cron scheduler written in Go.

cronflux runs cron-style and DAG-dependent jobs with at-least-once delivery,
exponential-backoff retries and a pluggable store. It ships as a single static
binary with an HTTP API, a web dashboard and Prometheus metrics.

## Features

- **Cron & intervals** — standard five-field cron, an optional seconds field,
  the `@hourly`/`@daily`/… descriptors and `@every 30s`.
- **At-least-once delivery** — runs are leased to workers; if a worker dies the
  lease expires and the run is redelivered.
- **Exponential-backoff retries** — per-job retry policy with a cap and optional
  jitter, then a dead-letter state.
- **DAG dependencies** — a job can depend on others; dependents run only after
  every upstream in the batch succeeds.
- **Pluggable store** — an in-memory backend for development and a PostgreSQL
  backend (using `SELECT … FOR UPDATE SKIP LOCKED`) for durable, multi-node use.
- **HTTP API + dashboard** — inspect jobs and runs over REST or in the browser.
- **Prometheus metrics** — run counts, durations, retries, backlog and lag.
- **Operable** — single static binary, structured logs, graceful shutdown.

## Install

```sh
go install github.com/Fzgt/cronflux/cmd/cronflux@latest
```

Or build from source:

```sh
git clone https://github.com/Fzgt/cronflux
cd cronflux
make build   # produces ./bin/cronflux
```

## Quickstart

```sh
# Start with the in-memory store on :8080.
cronflux

# In another shell, register a job that runs every 30 seconds.
curl -X POST localhost:8080/api/jobs -d '{
  "id": "heartbeat",
  "name": "Heartbeat",
  "spec": "@every 30s",
  "command": ["echo", "tick"],
  "enabled": true
}'

# Trigger it immediately and watch the runs.
curl -X POST localhost:8080/api/jobs/heartbeat/trigger
curl localhost:8080/api/runs
```

Open <http://localhost:8080/> for the dashboard and
<http://localhost:8080/metrics> for Prometheus metrics.

## Configuration

Flags override environment variables, which override the defaults.

| Flag             | Env                     | Default    | Description                        |
| ---------------- | ----------------------- | ---------- | ---------------------------------- |
| `-addr`          | `CRONFLUX_ADDR`         | `:8080`    | HTTP listen address                |
| `-store`         | `CRONFLUX_STORE`        | `memory`   | `memory` or `postgres`             |
| `-database-url`  | `CRONFLUX_DATABASE_URL` |            | PostgreSQL connection URL          |
| `-workers`       | `CRONFLUX_WORKERS`      | `4`        | Number of worker goroutines        |
| `-tick`          | `CRONFLUX_TICK`         | `1s`       | Scheduler tick interval            |
| `-lease`         | `CRONFLUX_LEASE`        | `30s`      | Run lease duration                 |
| `-jobs`          | `CRONFLUX_JOBS`         |            | Path to a job-definitions file     |
| `-log-level`     | `CRONFLUX_LOG_LEVEL`    | `info`     | `debug`, `info`, `warn` or `error` |

## Documentation

- [Architecture](docs/architecture.md)
- [Usage guide](docs/usage.md)
- [Design notes](docs/design-notes.md)
- [HTTP API reference](docs/api-reference.md)

## License

[MIT](LICENSE) © Chris Sun
