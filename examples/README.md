# Examples

Runnable examples for cronflux.

| Example                          | What it shows                                       |
| -------------------------------- | --------------------------------------------------- |
| [cron-next](cron-next/)          | Compute upcoming fire times for a cron spec         |
| [backoff](backoff/)              | The retry delays an exponential policy produces      |
| [inmemory-dag](inmemory-dag/)    | Running a three-step DAG with the in-memory store    |

There is also [`jobs.json`](jobs.json), a sample job-definitions file you can
feed to the server:

```sh
cronflux -jobs examples/jobs.json
```

Run any example with `go run`:

```sh
go run ./examples/cron-next "0 9 * * 1-5"
go run ./examples/backoff
go run ./examples/inmemory-dag
```
