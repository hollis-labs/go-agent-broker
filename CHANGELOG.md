# Changelog

All notable changes to `go-agent-broker` are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.2.0 — 2026-05-10

Adds the `DeterministicBroker` implementation alongside scaffold-era
`ModeBroker`, expands `Input` with the additional signals the
deterministic broker consults, and ships the public-release surface
(README rewrite, CHANGELOG, godoc package doc, runnable example,
`.gitignore`).

### Added

- `DeterministicBroker` — priority-ordered `Broker` implementation
  composing reflex / classified-mode / scope-tier × execution-pattern
  signals. Constructor: `broker.New()`. Six rules in priority order:
  reflex override, high-confidence work mode, action-verb work mode,
  high-confidence plan mode at open tier, open tier × subagent pattern,
  default chat fallback.
- `Input` fields: `Mode`, `ModeConfidence`, `ExecutionPattern`,
  `ReflexAgentSlug`, `ReflexConfidence`. All fields are zero-value-safe
  and non-breaking for callers that don't populate them.
- Named constants for emitted profile slugs (`ProfileWorker`,
  `ProfilePlanner`, `ProfileChat`), confidence thresholds
  (`ModeConfidenceHigh`, `ModeConfidenceLow`), scope tier wire strings
  (`TierTrivial` … `TierOpen`), execution-pattern wire strings
  (`PatternInline`, `PatternSubagent`, `PatternBackground`), and
  per-turn mode wire strings (`ModeChat`, `ModePlan`, `ModeWork`).
- `broker/doc.go` — dedicated package-level godoc rendered by pkg.go.dev,
  covering both `Broker` implementations and the named constants.
- `examples/deterministic/main.go` — runnable end-to-end demo of
  `broker.New()` covering each rule firing in priority order. Verified
  via `go run ./examples/deterministic`.
- `CHANGELOG.md` (this file).
- `.gitignore` covering Go build artifacts and internal-tooling files.
- Test coverage for every `DeterministicBroker` rule: each rule firing,
  threshold boundaries, plan-mode tier carve-out, mid-band-confidence
  rule-5 fallback, and multi-rule-applicable inputs (priority-order
  assertions).

### Changed

- `Input.SessionMode` semantics clarified — `ModeBroker` still pass-throughs
  on `SessionMode`, but `DeterministicBroker` keys off `Input.Mode`
  (per-turn classified mode) instead. `SessionMode` is retained on `Input`
  for telemetry and future rule extensions.
- `README.md` rewritten end-to-end to a public open-source shape — adds
  Install / Status / godoc-link / API-overview / Dependencies / Testing
  sections; documents the deterministic-broker rule priority as the
  primary path; documents the `ModeBroker` fallback secondarily.
- Godoc-rendered package-doc surfaces (`broker/broker.go`,
  `broker/deterministic.go`) rephrased for public readers; internal
  ticket and architecture-decision references removed.

## v0.1.0 — 2026-05-09

Initial scaffold release.

### Added

- `broker` package with the `Broker` interface, `Input`, `Decision`,
  `ModeBroker`, and `NewModeBroker` constructor.
- `ModeBroker.Decide` maps `SessionMode` directly to an agent profile —
  `"work"` → `"worker"`, `"plan"` → `"planner"`, all else (including
  empty and unknown values) → `"default-chat"` — with `Confidence=1.0`
  and a `Reason` of `"mode=<value>"`.
- Tests covering the three mode → profile mappings and that additional
  `Input` fields (`UserText`, `ScopeTier`, `ReflexMatchID`) are
  accepted but ignored by `ModeBroker`.
- `LICENSE` (MIT, Hollis Labs) and an initial `README.md` (rewritten
  in v0.2.0).
- `go.mod` at module path `github.com/hollis-labs/go-agent-broker`,
  Go 1.26.1, no external dependencies.
