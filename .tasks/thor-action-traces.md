# Task: Thor 2.0 Action Traces

## Context
Thor is the per-card interaction stress tester at `internal/tools/thor/`. It runs deterministic tests on every card (2496 tests across 104 cards in the gap set, 285K tests/sec full corpus). Currently when a test fails, we only see the final assertion ("board was set up but nothing changed" or "no events logged"). We need to see the FULL event chain step-by-step to diagnose WHERE the chain broke.

## Task
Add a `.trace` output mode to Thor that dumps per-card interaction execution traces:

1. **Trace capture** — during each test execution, record every significant event in order:
   - Board setup actions (add creature, set life, etc)
   - Trigger checks (which trigger conditions were evaluated, pass/fail)
   - Handler entry/exit (which per_card handler was called)
   - Effect resolution (what effect was attempted, what state changed)
   - Assertion result (what was expected vs observed)

2. **Output format** — when `--trace` flag is passed:
   - Write `data/thor-traces/{card_name_slug}.trace` for each FAILING test
   - Format: one event per line, timestamped with step number
   - Example:
     ```
     [001] SETUP: add_creature seat=0 name="Soldier Token" power=1 toughness=1
     [002] SETUP: set_zone card="Tolsimir, Midnights Light" zone=battlefield
     [003] TRIGGER_CHECK: "when ~ or another Wolf enters" -> condition=true
     [004] HANDLER_ENTER: tolsimir_midnights_light.go:OnETB
     [005] EFFECT_ATTEMPT: fight target="Soldier Token"
     [006] STATE_CHANGE: none (target already dead? or fight not resolved?)
     [007] ASSERT_FAIL: goldilocks_dead_effect expected board change, got delta=0
     ```

3. **Integration** — wire into existing Thor test loop in `internal/tools/thor/runner.go` (or wherever the main test execution lives). The trace collector should be a lightweight struct that accumulates entries and flushes on failure.

4. **CLI flag** — add `--trace` to the thor CLI. When enabled, traces are written for all failing tests. When disabled (default), zero overhead.

Look at the existing Thor code structure first, understand how tests execute, then add the trace layer. Keep it simple — append-only slice of trace entries per test, flush to file on failure.
