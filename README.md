# go-agent-broker

`go-agent-broker` is a small Go primitive for routing a per-turn user input to the agent profile that should handle it. Given a `broker.Input` (the raw user text plus session context — mode, scope tier, execution pattern, an optional reflex match), `Broker.Decide` returns a `Decision{ AgentProfile, Reason, Confidence }` that the caller uses to route the turn — chat handles it directly, dispatch to a `worker`, dispatch to a `planner`, etc.

The package ships two implementations of `Broker`:

- `DeterministicBroker` (`broker.New()`) — recommended. Applies a priority-ordered rule set over reflex / classified-mode / scope-tier × execution-pattern signals.
- `ModeBroker` (`broker.NewModeBroker()`) — minimal fallback. Maps `SessionMode` directly to an agent profile.

Both are pure functions over `Input` (no mutable state) and safe for concurrent use. The `Broker` interface is the stable surface; richer implementations can replace either of the shipped impls without re-plumbing call-sites.

## Status

Pre-1.0 (`v0.x`). The `Broker` interface, `Input`, `Decision`, the named constants, and both shipped implementations are stable in shape; minor breaks may still happen between `v0.x` releases. See [`CHANGELOG.md`](CHANGELOG.md) for per-release detail and pin a version in your `go.mod`.

Documentation: [pkg.go.dev/github.com/hollis-labs/go-agent-broker](https://pkg.go.dev/github.com/hollis-labs/go-agent-broker).

## Install

```bash
go get github.com/hollis-labs/go-agent-broker
```

## Usage

The recommended path is `broker.New()` (`DeterministicBroker`):

```go
package main

import (
    "context"
    "fmt"

    "github.com/hollis-labs/go-agent-broker/broker"
)

func main() {
    b := broker.New()

    decision, err := b.Decide(context.Background(), broker.Input{
        UserText:         "implement the foo refactor",
        Mode:             broker.ModeWork,           // from your mode classifier
        ModeConfidence:   0.85,                      // confidence in [0, 1]
        ScopeTier:        broker.TierMedium,         // optional; from your scope classifier
        ExecutionPattern: broker.PatternInline,      // optional; from your pattern classifier
        // ReflexMatchID / ReflexAgentSlug populated when a reflex matcher fires upstream.
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("profile=%s reason=%s confidence=%.2f\n",
        decision.AgentProfile, decision.Reason, decision.Confidence)
    // profile=worker reason=mode=work confidence=0.85
}
```

`Decision.Reason` is a human-readable string suitable for surfacing in a caller's telemetry / decision-review surfaces so future readers can see why the broker chose what it did.

For a runnable end-to-end example covering each rule firing in priority order, see [`examples/deterministic/`](examples/deterministic).

### Rule priority (`DeterministicBroker`)

First match wins:

1. **Reflex match** with non-empty `ReflexAgentSlug` → that slug. Confidence inherits from `ReflexConfidence`. `Reason="reflex:<id>"`.
2. **`Mode = "work"` with confidence ≥ 0.85** → `worker`. `Reason="mode=work"`.
3. **`Mode = "work"` with confidence ≥ 0.7** → `worker`. `Reason="action-verb-work"` (logged distinctly so telemetry can isolate the lower-confidence band).
4. **`Mode = "plan"` with confidence ≥ 0.85 AND `ScopeTier = "open"`** → `planner`. `Reason="mode=plan,tier=open"`. Plan-mode at any other tier falls through to chat (planning conversation, not full decomposition).
5. **`ScopeTier = "open"` AND `ExecutionPattern = "subagent"`** → `planner` (mode-independent). `Reason="tier=open,pattern=subagent"`.
6. **Default** → chat handles the turn directly. `AgentProfile=""`, `Reason="default-chat-handle"`, `Confidence=0`.

### Minimal `ModeBroker` fallback

```go
b := broker.NewModeBroker()
d, _ := b.Decide(ctx, broker.Input{SessionMode: "work"})
// d.AgentProfile == "worker"
```

`ModeBroker` only consults `SessionMode` — `"work"` → `"worker"`, `"plan"` → `"planner"`, anything else → `"default-chat"` — and always returns `Confidence=1.0`. Useful when callers do not yet have classifier output to feed the deterministic broker.

## API Overview

Package `github.com/hollis-labs/go-agent-broker/broker`:

- `Input` — per-turn signal bundle. Fields: `UserText`, `SessionMode`, `Mode`, `ModeConfidence`, `ScopeTier`, `ExecutionPattern`, `ReflexMatchID`, `ReflexAgentSlug`, `ReflexConfidence`. All fields are zero-value-safe; adding fields is non-breaking.
- `Decision` — broker output: `AgentProfile` (slug; empty means "chat handles it directly"), `Reason` (human-readable), `Confidence` (`[0, 1]`).
- `Broker` — interface with a single `Decide(ctx, Input) (Decision, error)` method. Implementations must be safe for concurrent use.
- `New() *DeterministicBroker` — priority-ordered rule-set implementation.
- `NewModeBroker() *ModeBroker` — minimal `SessionMode` → profile pass-through.
- Profile slugs: `ProfileWorker`, `ProfilePlanner`, `ProfileChat` (empty).
- Confidence thresholds: `ModeConfidenceHigh` (0.85), `ModeConfidenceLow` (0.7).
- Scope tier strings: `TierTrivial`, `TierSmall`, `TierMedium`, `TierLarge`, `TierOpen`.
- Execution-pattern strings: `PatternInline`, `PatternSubagent`, `PatternBackground`.
- Mode strings: `ModeChat`, `ModePlan`, `ModeWork`.

## Dependencies

None. The module is stdlib-only.

## Testing

```bash
go test ./...
```

No environment variables, fixtures, or external services are needed.

For runnable end-to-end examples, see the [`examples/`](examples/) directory.

## License

[MIT](LICENSE) © Hollis Labs.
