# cron-next

Prints the next five activation times for a cron expression using the
`github.com/Fzgt/cronflux/cron` package.

```sh
go run ./examples/cron-next "0 9 * * 1-5"
```

Output:

```
next fire times for "0 9 * * 1-5":
  2026-07-20T09:00:00+10:00
  2026-07-21T09:00:00+10:00
  ...
```

Omit the argument to use the default `*/15 * * * *`.
