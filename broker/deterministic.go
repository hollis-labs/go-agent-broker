package broker

import "context"

// DeterministicBroker is the v1 priority-ordered rule-set impl
// (CW-20260509-0045) per `decisions.nanite.architecture.agent_broker_v1`.
//
// Rule priority (highest first; first match wins):
//
//  1. Reflex match with a non-empty AgentSlug → that slug. Confidence
//     inherits from `Input.ReflexConfidence`. `Reason="reflex:<id>"`.
//  2. Mode = "work" with confidence ≥ 0.85 → "worker". `Reason="mode=work"`.
//  3. Mode = "work" with confidence ≥ 0.7  → "worker". `Reason="action-verb-work"`.
//  4. Mode = "plan" with confidence ≥ 0.85 AND ScopeTier = "open" → "planner".
//     `Reason="mode=plan,tier=open"`. Plan-mode at any other tier falls
//     through to chat (planning conversation, not full decomposition).
//  5. ScopeTier = "open" AND ExecutionPattern = "subagent" → "planner"
//     (regardless of mode). `Reason="tier=open,pattern=subagent"`.
//  6. Default → chat handles it. `Reason="default-chat-handle"`.
//
// The struct is concurrency-safe (it has no mutable state — Decide is a
// pure function over Input).
type DeterministicBroker struct{}

// New returns the v1 deterministic broker. Use this for production wiring
// in CW-20260509-0046; `NewModeBroker` is retained only for the no-op
// parity scaffold from CW-20260502-0005.
func New() *DeterministicBroker { return &DeterministicBroker{} }

// Decide implements Broker for the deterministic v1 rule set.
func (*DeterministicBroker) Decide(_ context.Context, in Input) (Decision, error) {
	// Rule 1 — reflex override. Highest priority because the user's reflex
	// catalog is an explicit, hand-authored routing decision; the rest of
	// the rule set is statistical inference over the message text.
	if in.ReflexMatchID != "" && in.ReflexAgentSlug != "" {
		return Decision{
			AgentProfile: in.ReflexAgentSlug,
			Reason:       "reflex:" + in.ReflexMatchID,
			Confidence:   in.ReflexConfidence,
		}, nil
	}

	// Rule 2 — high-confidence work mode → worker.
	// Slash (/work) and imperative-work phrases produce ModeConfidence 1.0
	// or 0.85 respectively in the host's classifier; both clear the high
	// threshold and dispatch with full confidence.
	if in.Mode == ModeWork && in.ModeConfidence >= ModeConfidenceHigh {
		return Decision{
			AgentProfile: ProfileWorker,
			Reason:       "mode=work",
			Confidence:   in.ModeConfidence,
		}, nil
	}

	// Rule 3 — action-verb work mode → worker (logged distinctly).
	// Action-verb classification (confidence 0.7) is weaker than imperative
	// phrasing but still dispatches; the distinct `Reason` string lets v2
	// telemetry analysis isolate this band when tuning the threshold or
	// deciding whether the peer-agent escape hatch is justified.
	if in.Mode == ModeWork && in.ModeConfidence >= ModeConfidenceLow {
		return Decision{
			AgentProfile: ProfileWorker,
			Reason:       "action-verb-work",
			Confidence:   in.ModeConfidence,
		}, nil
	}

	// Rule 4 — high-confidence plan mode AT TierOpen → planner.
	// Plan mode at any tier other than Open is a planning *conversation*
	// (the chat agent is fine for that); only Open-tier plan inputs are
	// large enough to warrant a planner subagent decomposition.
	if in.Mode == ModePlan && in.ModeConfidence >= ModeConfidenceHigh && in.ScopeTier == TierOpen {
		return Decision{
			AgentProfile: ProfilePlanner,
			Reason:       "mode=plan,tier=open",
			Confidence:   in.ModeConfidence,
		}, nil
	}

	// Rule 5 — TierOpen × PatternSubagent → planner (mode-independent).
	// This catches inputs that classify as chat-mode but the scope/pattern
	// classifier flags as decomposition-shaped (e.g. "research the foo
	// landscape exhaustively" — chat-mode, but Open + Subagent).
	if in.ScopeTier == TierOpen && in.ExecutionPattern == PatternSubagent {
		return Decision{
			AgentProfile: ProfilePlanner,
			Reason:       "tier=open,pattern=subagent",
			Confidence:   confidenceForTierPattern(in),
		}, nil
	}

	// Rule 6 — default. Chat handles the turn directly; no dispatch.
	// Confidence is 0 to signal "no rule matched, this is the fallback" —
	// telemetry analysis on `confidence=0,reason=default-chat-handle`
	// tells us how often we lean on the default and is the primary input
	// to the v2 peer-agent decision.
	return Decision{
		AgentProfile: ProfileChat,
		Reason:       "default-chat-handle",
		Confidence:   0,
	}, nil
}

// confidenceForTierPattern picks the broker's confidence for the rule-5
// (TierOpen × PatternSubagent) firing. The host's scope/pattern classifier
// does not emit a per-classification confidence the way ClassifyMode does,
// so the broker uses a fixed mid-band confidence (0.75) to flag this as
// "rules-derived, not as strong as a high-confidence mode signal but
// stronger than the default fallback."
//
// If the host did populate `ModeConfidence` (e.g. "research the foo" had
// no mode signal so ModeConfidence is 0), we keep the 0.75 floor — the
// rule-5 firing is the load-bearing signal here, not the mode classifier.
func confidenceForTierPattern(_ Input) float64 {
	return 0.75
}

// Compile-time assertion that *DeterministicBroker satisfies Broker.
var _ Broker = (*DeterministicBroker)(nil)
