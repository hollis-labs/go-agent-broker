## go-agent-broker

`go-agent-broker` is the intent-based agent-router primitive shared across the Hollis Labs portfolio. Given the user's input plus session context (mode, scope tier, reflex match), `broker.Decide` returns a `Decision{ AgentProfile, Reason, Confidence }` that callers use to route the turn to the right agent profile (chat handles it directly, dispatch to `worker`, dispatch to `planner`, etc.).

This is a sibling repo following the same convention as `go-providers`, `go-modelsdev`, `go-mcp`, and `go-otel` — a narrow, dependency-free Go module the host imports via `replace` until first publish.

## Status

Scaffold (CW-20260502-0005). The `NewModeBroker` impl is a no-op that mirrors the host's pre-broker dispatch behavior — it returns `default-chat` for `chat`, `worker` for `work`, `planner` for `plan`. The deterministic v1 implementation (composing classifier + reflex + scope-tier signals) lands in a follow-up sub-ticket per `decisions.nanite.architecture.agent_broker_v1`.

## Usage

```go
import "github.com/hollis-labs/go-agent-broker/broker"

b := broker.NewModeBroker()
decision, err := b.Decide(ctx, broker.Input{
    UserText:    "summarize the meeting notes",
    SessionMode: "chat",
})
// decision.AgentProfile == "default-chat"
```

`Decision.Reason` is a human-readable string surfaced in `event_log` so future sessions (and the v2 decision review) can see why the broker chose what it did.
