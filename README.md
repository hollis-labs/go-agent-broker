## go-agent-broker

`go-agent-broker` is the intent-based agent-router primitive shared across the Hollis Labs portfolio. Given the user's input plus session context (mode, scope tier, reflex match), `broker.Decide` returns a `Decision{ AgentProfile, Reason, Confidence }` that callers use to route the turn to the right agent profile (chat handles it directly, dispatch to `worker`, dispatch to `planner`, etc.).

This is a sibling repo following the same convention as `go-providers`, `go-modelsdev`, `go-mcp`, and `go-otel` — a narrow, dependency-free Go module the host imports via `replace` until first publish.

## Status

- **v0.1.0** (CW-20260502-0005) — scaffold. `NewModeBroker` is a no-op session-mode pass-through preserved for parity testing.
- **v0.2.0** (CW-20260509-0045) — deterministic v1 (`broker.New()`). Priority-ordered rule set composing reflex / classified-mode / scope-tier × execution-pattern signals per `decisions.nanite.architecture.agent_broker_v1`.

The v2 peer-agent escape hatch is explicitly out of scope; the `Broker` interface accommodates it but v1 ships deterministic-only and lights up telemetry (`Decision.Reason`) so v2 can be justified by data.

## Usage

### v1 deterministic (production wiring)

```go
import "github.com/hollis-labs/go-agent-broker/broker"

b := broker.New()
decision, err := b.Decide(ctx, broker.Input{
    UserText:         "implement the foo refactor",
    Mode:             broker.ModeWork,    // from classify.ClassifyMode
    ModeConfidence:   0.85,                // from classify.ClassifyMode
    ScopeTier:        broker.TierMedium,   // from classify.Classify
    ExecutionPattern: broker.PatternInline,
    // ReflexMatchID/Slug populated when reflex matcher fired upstream.
})
// decision.AgentProfile == "worker"
// decision.Reason       == "mode=work"
// decision.Confidence   == 0.85
```

### Rule priority (v1 deterministic)

1. **Reflex match** with non-empty `ReflexAgentSlug` → that slug. Confidence inherits from `ReflexConfidence`.
2. **Mode = "work" with confidence ≥ 0.85** → `worker`. `Reason="mode=work"`.
3. **Mode = "work" with confidence ≥ 0.7** → `worker`. `Reason="action-verb-work"` (logged distinctly so v2 telemetry can isolate the band).
4. **Mode = "plan" with confidence ≥ 0.85 AND ScopeTier = "open"** → `planner`. Otherwise plan-mode → chat (planning conversation).
5. **ScopeTier = "open" AND ExecutionPattern = "subagent"** → `planner` (mode-independent).
6. **Default** → chat handles it directly (`AgentProfile == ""`).

### v0.1.0 no-op (deprecated)

```go
b := broker.NewModeBroker() // no-op session-mode pass-through; CW-20260502-0005 scaffold
```

`Decision.Reason` is a human-readable string surfaced in `event_log` so future sessions (and the v2 decision review) can see why the broker chose what it did.
