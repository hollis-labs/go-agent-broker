# go-agent-broker

A dependency-free Go primitive that routes one user turn to the agent profile
that should handle it: `Broker.Decide` maps an `Input` (user text plus mode,
scope tier, execution pattern and an optional reflex match) to a `Decision`
carrying a profile slug, a reason and a confidence. It does not classify the
input, dispatch the turn, or run anything — callers project their own
classifier and reflex outputs into `Input`'s primitive fields.

## Start Here

- `broker/doc.go` is the package contract and describes both implementations.
- `broker/broker.go` owns the `Broker` interface, `Input`, `Decision` and the
  named profile / tier / pattern / mode constants.
- `broker/deterministic.go` owns the priority-ordered rule set `New()` returns.
- `examples/deterministic/main.go` is the runnable call-site shape.
- `CHANGELOG.md` records release history and the v0.2.0 retraction.

## Commands

```bash
go test ./...
go vet ./...
```

## Boundaries

This module was absorbed into `agentkit` as `agentkit/broker` at agentkit
v0.1.0 and has not changed since v0.2.1 (2026-05-10). New work belongs in
`agentkit`; change this repo only to serve external consumers pinned to this
import path.

`Input` uses primitive-typed fields on purpose so the module stays
dependency-free. Do not add a dependency to accept a richer classifier or
reflex type — project it at the call site instead.

`v0.2.0` is retracted in `go.mod` because its release notes carried internal
references. Leave the retraction in place.
