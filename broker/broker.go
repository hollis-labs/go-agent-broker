// Package broker is the intent-based agent-router primitive used by host
// applications (e.g. nanite) to decide which agent profile should handle a
// user turn — chat handles it directly, dispatch to a `worker`, dispatch to
// a `planner`, etc.
//
// The package ships a no-op `ModeBroker` (CW-20260502-0005 scaffold) that
// mirrors the host's pre-broker dispatch behavior so the call-site seam can
// be wired without a semantic change. The deterministic v1 implementation
// — composing classifier + reflex + scope-tier signals per
// `decisions.nanite.architecture.agent_broker_v1` — lands in a follow-up.
package broker

import "context"

// Input is the per-turn signal bundle a broker consults to make a decision.
// The no-op `ModeBroker` only reads SessionMode; the v1 deterministic impl
// will additionally read UserText, ScopeTier, and ReflexMatchID. Adding
// fields is non-breaking — callers that don't populate them get the zero
// value.
type Input struct {
	// UserText is the raw user-facing prompt for the turn.
	UserText string
	// SessionMode is the active workspace mode slug ("chat", "plan",
	// "work"). Empty string means unknown — the broker falls back to chat.
	SessionMode string
	// ScopeTier is the host's classify.ScopeTier projection ("trivial",
	// "small", "medium", "large", "open"). Empty string means the host did
	// not classify (broker treats as unknown).
	ScopeTier string
	// ReflexMatchID is the matched reflex's ID when the host's reflex
	// matcher fired upstream. Empty string means no reflex match.
	ReflexMatchID string
}

// Decision is the broker's per-turn output. AgentProfile is the slug the
// caller should route to (empty string means "chat handles it directly").
// Reason is a human-readable explanation surfaced in the host's event_log
// for telemetry and v2 decision review. Confidence is in [0, 1].
type Decision struct {
	AgentProfile string
	Reason       string
	Confidence   float64
}

// Broker is the narrow surface host call-sites consult before dispatch.
// Implementations must be safe for concurrent use.
type Broker interface {
	Decide(ctx context.Context, input Input) (Decision, error)
}

// ModeBroker is the no-op v1 scaffold impl. It maps SessionMode → AgentProfile
// using the host's pre-broker behavior so wiring it in produces no semantic
// change. The v1 deterministic impl per agent_broker_v1 replaces this with
// a priority-ordered rule set.
type ModeBroker struct{}

// NewModeBroker returns the no-op mode-only broker.
func NewModeBroker() *ModeBroker { return &ModeBroker{} }

// Decide implements Broker.
func (*ModeBroker) Decide(_ context.Context, input Input) (Decision, error) {
	switch input.SessionMode {
	case "work":
		return Decision{AgentProfile: "worker", Reason: "mode=work", Confidence: 1.0}, nil
	case "plan":
		return Decision{AgentProfile: "planner", Reason: "mode=plan", Confidence: 1.0}, nil
	default:
		return Decision{AgentProfile: "default-chat", Reason: "mode=chat", Confidence: 1.0}, nil
	}
}
