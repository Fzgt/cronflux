# Security Policy

## Supported versions

cronflux is pre-1.0. Security fixes are applied to the latest released minor
version.

| Version | Supported |
| ------- | --------- |
| 0.1.x   | ✅        |

## Reporting a vulnerability

Please **do not** open a public issue for security problems.

Instead, report privately using GitHub's
[private vulnerability reporting](https://github.com/Fzgt/cronflux/security/advisories/new)
on this repository. Include:

- a description of the issue and its impact,
- the version or commit affected,
- steps to reproduce, and a proof of concept if you have one.

You can expect an acknowledgement within a few days. Once a fix is ready we will
coordinate a release and credit you in the advisory unless you prefer to remain
anonymous.

## Scope notes

cronflux executes the commands configured for each job. Treat job definitions as
trusted input: anyone who can create jobs through the API can run commands on the
host. Deploy the HTTP API behind authentication and network controls
appropriate to your environment.
