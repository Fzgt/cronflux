# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0]

Initial release.

### Added

- Cron parser supporting the five-field syntax, an optional seconds field, the
  `@`-descriptors and `@every <duration>`, exposed as a standalone `cron`
  package.
- Exponential backoff with a cap and optional jitter (`backoff` package).
- Job and run domain model with a dependency graph, topological sort and cycle
  detection.
- Pluggable `store.Store` with in-memory and PostgreSQL backends sharing a
  conformance suite; PostgreSQL claims work with `FOR UPDATE SKIP LOCKED`.
- Scheduler with at-least-once delivery, lease-based redelivery, retries,
  dead-lettering and DAG-aware dispatch.
- HTTP API for jobs and runs, health and readiness probes, an embedded web
  dashboard and Prometheus metrics.
- Single static binary with flag/environment configuration, a job-definitions
  file loader and graceful shutdown.

[Unreleased]: https://github.com/Fzgt/cronflux/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Fzgt/cronflux/releases/tag/v0.1.0
