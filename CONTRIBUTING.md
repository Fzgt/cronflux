# Contributing

Thanks for taking the time to contribute! This project is small enough that
you can hold most of it in your head — please skim the
[architecture doc](docs/architecture.md) before diving in.

## Getting started

```sh
git clone https://github.com/Fzgt/cronflux
cd cronflux
make all      # gofmt check, go vet, golangci-lint, tests
```

You need Go 1.23 or newer. `golangci-lint` is used for linting; install it from
<https://golangci-lint.run>.

## Development loop

- `make test` — run the unit tests.
- `make race` — run them with the race detector.
- `make cover` — print total coverage.
- `make lint` — run golangci-lint.
- `make build` — build the binary into `bin/`.

The PostgreSQL integration tests are behind a build tag and need a database:

```sh
export CRONFLUX_TEST_DATABASE_URL='postgres://cronflux:cronflux@localhost:5432/cronflux_test?sslmode=disable'
go test -tags integration ./internal/store/postgres/...
```

## Pull requests

- Keep each PR focused on one thing.
- Add or update tests for behaviour changes; new store methods should be covered
  by the shared `storetest` suite so both backends stay in lockstep.
- Run `make all` before pushing — CI runs the same checks.
- Update `CHANGELOG.md` and the relevant docs when user-facing behaviour
  changes.

## Commit messages

Conventional-commit prefixes (`feat:`, `fix:`, `docs:`, …) are appreciated but
not required. Write in the imperative and explain the _why_ when it is not
obvious.

## Reporting bugs

Open an issue with the job spec, the command you ran and the observed vs
expected behaviour. See [SECURITY.md](SECURITY.md) for security-sensitive
reports.
