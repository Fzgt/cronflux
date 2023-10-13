# HTTP API reference

All responses are JSON. The base URL is the configured listen address.

## Jobs

### `GET /api/jobs`

List all jobs.

```sh
curl localhost:8080/api/jobs
```

### `POST /api/jobs`

Create or replace a job. The body is a job object. Durations (`timeout`,
`backoff.base`, `backoff.max`) are expressed in **nanoseconds** on the wire.

```sh
curl -X POST localhost:8080/api/jobs -d '{
  "id": "report",
  "name": "Nightly report",
  "spec": "0 2 * * *",
  "command": ["./report.sh"],
  "max_retries": 2,
  "enabled": true
}'
```

Returns `201 Created` with the stored job, or `400` if the id is missing or the
cron spec is invalid.

### `GET /api/jobs/{id}`

Fetch one job. `404` if it does not exist.

### `DELETE /api/jobs/{id}`

Delete a job. Returns `204 No Content`, or `404` if it does not exist.

### `POST /api/jobs/{id}/trigger`

Enqueue an immediate run in a fresh batch. Returns `202 Accepted` with the
created run.

## Runs

### `GET /api/runs`

List runs, newest first. Query parameters:

| Param    | Description                          |
| -------- | ------------------------------------ |
| `job`    | filter by job id                     |
| `batch`  | filter by batch id                   |
| `state`  | filter by state                      |
| `limit`  | maximum rows (default 100)           |
| `offset` | rows to skip                         |

```sh
curl 'localhost:8080/api/runs?job=report&state=failed&limit=20'
```

### `GET /api/runs/{id}`

Fetch one run. `404` if it does not exist.

## Run states

| State       | Meaning                                     |
| ----------- | ------------------------------------------- |
| `pending`   | waiting to be claimed                       |
| `running`   | leased to a worker                          |
| `succeeded` | completed successfully                      |
| `failed`    | an attempt failed; a retry was scheduled    |
| `dead`      | retries exhausted                           |
| `skipped`   | an upstream dependency did not succeed      |

## Operational endpoints

| Endpoint    | Description                              |
| ----------- | ---------------------------------------- |
| `GET /`     | web dashboard                            |
| `GET /healthz` | liveness probe                        |
| `GET /readyz`  | readiness probe (checks the store)    |
| `GET /metrics` | Prometheus metrics                    |
