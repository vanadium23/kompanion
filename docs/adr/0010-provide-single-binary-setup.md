# 10. Provide single binary setup

Date: 2025-03-02

## Status

Accepted

## Context

Not all software are good to be distributed via docker. Some users may prefer to just run a single binary, and have something like `curl && ./run`. But currently we need to be migrations and web folders to be placed with binaries.

## Decision

Embed static files inside binary and adjust code to it. Also, we need to adjust CI to place the binary in release.

We have several options to implement embeded FS:
- `go:embed`
- `go-bindata`
- `go-rice`

We use `go:embed`, because it will not introduce new dependencies.


## Consequences

- (+) Single binary is enough, but still requires postgresql
- (-) Binary will be larger

