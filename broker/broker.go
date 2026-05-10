// Package broker is the intent-based agent-router primitive used by host
// applications (e.g. nanite) to decide which agent profile should handle a
// user turn — chat handles it directly, dispatch to a `worker`, dispatch to
// a `planner`, etc.
//
// The package ships two implementations:
//
//   - `ModeBroker` (CW-20260502-0005 scaffold) — no-op session-mode-only
//     pass-through preserved for parity with pre-broker dispatch behavior.
//   - `DeterministicBroker` (CW-20260509-0045, v1) — priority-ordered rule
//     set composing reflex / classified-mode / scope-tier × execution-pattern
//     signals per `decisions.nanite.architecture.agent_broker_v1`.
//
// The v2 peer-agent escape hatch is explicitly out of scope; the `Broker`
// interface accommodates it but v1 ships deterministic-only and lights up
// telemetry (`Decision.Reason`) so v2 can be justified by data.
package broker

import "context"

// Input is the per-turn signal bundle a broker consults to make a decision.
// All fields are zero-value-safe — callers that only populate a subset get
// well-defined (degenerate) behavior. Adding fields in the future is
// non-breaking under the same contract.
//
// Field projections (string / float64 shapes rather than concrete classify
// or reflex types) keep the broker module dependency-free: the host
// projects its internal classifier / reflex output into these primitives.
type Input struct {
	// UserText is the raw user-facing prompt for the turn.
	UserText string

	// SessionMode is the active workspace mode slug ("chat", "plan",
	// "work"). This is the persistent per-session mode, NOT the per-turn
	// classified mode. Empty means unknown.
	//
	// Read by `ModeBroker` (no-op pass-through). The deterministic v1 impl
	// does not consult SessionMode directly — its rules key off `Mode`
	// (per-turn classified) instead, but SessionMode remains in the input
	// for telemetry / future rule extensions.
	SessionMode string

	// Mode is the per-turn classified mode produced by the host's mode
	// classifier (`classify.ClassifyMode` in nanite): one of "chat",
	// "plan", "work". Empty means the host did not classify (the
	// deterministic broker treats this as no-signal and falls through).
	Mode string

	// ModeConfidence is the classifier's confidence in `Mode`, in [0, 1].
	// The deterministic broker's mode rules apply confidence thresholds
	// (>= 0.85 for slash / imperative, >= 0.7 for action-verb).
	ModeConfidence float64

	// ScopeTier is the host's classify.ScopeTier projection ("trivial",
	// "small", "medium", "large", "open"). Empty means no classification.
	ScopeTier string

	// ExecutionPattern is the host's classify.ExecutionPattern projection
	// ("inline", "subagent", "background"). Empty means no classification.
	ExecutionPattern string

	// ReflexMatchID is the matched reflex's ID when the host's reflex
	// matcher fired upstream. Empty means no match.
	ReflexMatchID string

	// ReflexAgentSlug is the agent slug the matched reflex specified. The
	// deterministic broker's highest-priority rule routes to this slug
	// when ReflexMatchID is set AND ReflexAgentSlug is non-empty.
	ReflexAgentSlug string

	// ReflexConfidence is the matched reflex's confidence in [0, 1]. When
	// reflex routing fires, the broker's `Decision.Confidence` inherits
	// this value.
	ReflexConfidence float64
}

// Decision is the broker's per-turn output. AgentProfile is the slug the
// caller should route to — empty string means "chat handles it directly,"
// non-empty means "dispatch via task_execute with this slug." Reason is a
// human-readable explanation surfaced in the host's event_log for
// telemetry and v2 decision review. Confidence is in [0, 1].
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

// AgentProfile slug constants the broker emits. Hosts may dispatch other
// slugs by reflex / configuration; these are the v1 default routes.
const (
	ProfileWorker  = "worker"
	ProfilePlanner = "planner"
	// ProfileChat is the empty AgentProfile — chat handles the turn
	// directly without dispatch. Defined as a named constant so callers
	// can compare against `decision.AgentProfile == broker.ProfileChat`
	// rather than `== ""`.
	ProfileChat = ""
)

// Confidence thresholds applied by the deterministic broker's mode rules.
// Match the tiers `classify.ClassifyMode` produces in nanite (slash 1.0,
// imperative 0.85, action-verb 0.7, default 0.0).
const (
	ModeConfidenceHigh = 0.85
	ModeConfidenceLow  = 0.7
)

// Scope tier and execution pattern wire strings. These match the canonical
// `String()` projections of `classify.ScopeTier` / `classify.ExecutionPattern`
// in nanite. Defined here so the broker compiles without depending on
// nanite/internal/classify.
const (
	TierTrivial = "trivial"
	TierSmall   = "small"
	TierMedium  = "medium"
	TierLarge   = "large"
	TierOpen    = "open"

	PatternInline     = "inline"
	PatternSubagent   = "subagent"
	PatternBackground = "background"
)

// Mode wire strings. Match `classify.ModeResult.Suggested` values.
const (
	ModeChat = "chat"
	ModePlan = "plan"
	ModeWork = "work"
)

// ModeBroker is the no-op v1 scaffold impl from CW-20260502-0005. It maps
// SessionMode → AgentProfile using the host's pre-broker behavior so wiring
// it in produces no semantic change. Retained for parity testing during the
// CW-20260509-0046 call-site cutover; new callers should use the
// `DeterministicBroker` returned by `New()`.
type ModeBroker struct{}

// NewModeBroker returns the no-op mode-only broker.
func NewModeBroker() *ModeBroker { return &ModeBroker{} }

// Decide implements Broker for ModeBroker.
func (*ModeBroker) Decide(_ context.Context, input Input) (Decision, error) {
	switch input.SessionMode {
	case ModeWork:
		return Decision{AgentProfile: ProfileWorker, Reason: "mode=work", Confidence: 1.0}, nil
	case ModePlan:
		return Decision{AgentProfile: ProfilePlanner, Reason: "mode=plan", Confidence: 1.0}, nil
	default:
		return Decision{AgentProfile: "default-chat", Reason: "mode=chat", Confidence: 1.0}, nil
	}
}

// Compile-time assertion that *ModeBroker satisfies Broker.
var _ Broker = (*ModeBroker)(nil)
