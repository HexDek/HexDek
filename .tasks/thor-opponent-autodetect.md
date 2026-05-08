# Task: Thor 2.0 — Opponent Auto-Detect

## Context
Thor currently tests cards in a single-seat vacuum. Cards that reference "opponent" in their oracle text (e.g. "whenever an opponent casts a spell", "target opponent discards", "each opponent loses 1 life") need an active adversarial seat to exercise their triggers. Currently these cards get a passive seat 1 with a Bear on the battlefield, but that seat never takes actions — so "whenever an opponent casts" never fires because no opponent ever casts.

## Task
Add automatic opponent detection and adversarial seat spawning to Thor's test setup:

1. **AST parsing for opponent references** — scan each card's AST for nodes that reference opponents:
   - Target filters containing "opponent" 
   - Trigger conditions containing "opponent" (e.g. "whenever an opponent casts/attacks/gains life/draws")
   - Effect targets that say "each opponent" or "target opponent"
   - Look at the AST structure in `internal/gameast/` to find how opponent references are represented

2. **Adversarial seat setup** — when a card references opponents, enhance the test board:
   - Seat 1 gets a hand with castable spells (for "opponent casts" triggers)
   - Seat 1 gets creatures that can attack (for "opponent attacks" triggers)  
   - Seat 1 gets life changes queued (for "opponent gains/loses life" triggers)
   - The adversarial seat takes one relevant action before assertions are checked

3. **Integration with existing test flow** — this hooks into `makeGameState` or the per-card setup in `testInteraction`/`testGoldilocksCard`. When the card under test has opponent references in its AST, the enhanced board is used instead of the default.

4. **No manual curation** — the whole point is that this is automatic. Parse the AST, detect opponent references, spawn the appropriate adversarial setup. No card-by-card configuration.

Look at:
- `cmd/hexdek-thor/main.go` and `cmd/hexdek-thor/goldilocks.go` for the current test setup
- `internal/gameast/` for AST node types and how opponent references appear
- The recently added `trace.go` for the trace infrastructure (your work should emit trace entries too)

Wire trace entries into your opponent actions so traces show: `[N] OPPONENT_ACTION: cast spell "Test Instant" from seat 1`
