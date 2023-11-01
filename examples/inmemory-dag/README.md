# inmemory-dag

Runs a three-step DAG — `extract` → `transform` → `load` — using the in-memory
store and prints the order the jobs executed in. Each job runs only after its
upstream succeeds.

```sh
go run ./examples/inmemory-dag
# execution order: [extract transform load]
```
