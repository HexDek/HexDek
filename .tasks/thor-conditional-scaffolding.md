# Task: Thor 2.0 — Conditional Trigger Scaffolding

## Context
The #1 source of Thor failures (2,032 draw gaps, 1,612 life gaps, 1,095 damage gaps) is the Goldilocks test setting up a board but NOT meeting the card's trigger condition. Example: "Whenever a creature you control dies, draw a card" — Goldilocks places the card on the battlefield but nothing dies, so the trigger never fires, and the test reports "no events logged for draw effect."

This is a TEST HARNESS issue, not an engine bug. The card's handler might be perfectly correct — Thor just never gives it the right conditions.

## Task
Upgrade Goldilocks board setup to auto-generate trigger condition actions from AST:

1. **Parse trigger conditions** — for each card's triggered abilities, extract the condition clause:
   - "Whenever a creature dies" → condition: creature must die
   - "Whenever you gain life" → condition: gain life event needed
   - "Whenever you cast a spell" → condition: cast event needed
   - "Whenever a creature enters the battlefield" → condition: ETB event needed
   - "At the beginning of your upkeep" → condition: upkeep phase needed
   - "Whenever an opponent discards" → condition: opponent discard needed
   - "Whenever you attack" → condition: declare attackers needed

2. **Setup action registry** — map each condition type to a board setup action:
   ```
   creature_dies → add a 1/1 token, then destroy it
   gain_life → gs.SetLife(seat, life+1) 
   cast_spell → push a fake spell onto stack and resolve
   creature_etb → add a creature to battlefield via proper ETB path
   upkeep → advance phase to upkeep
   opponent_discards → force discard from seat 1
   attack → declare attack with existing creature
   ```

3. **Integration into Goldilocks** — in `testGoldilocksCard`, after building the goldilocks state but BEFORE taking the snapshot and firing:
   - Parse the trigger condition from the card's AST
   - Look up the setup action in the registry
   - Execute the setup action (this should cause the trigger to fire)
   - THEN check if the effect resolved

4. **Trace integration** — emit trace entries for setup actions:
   ```
   [N] CONDITION_SETUP: "whenever a creature dies" → destroying token "Setup Victim"
   [N+1] TRIGGER_FIRE: creature_dies triggered on "Card Under Test"
   ```

5. **Fallback** — if the AST condition can't be parsed or mapped, fall back to existing behavior (no setup action, just place the card). Don't crash on unrecognized conditions.

Look at:
- `cmd/hexdek-thor/goldilocks.go` — the `makeGoldilocksState` and `testGoldilocksCard` functions
- `internal/gameast/` — how trigger conditions are represented in the AST
- `cmd/hexdek-thor/trace.go` — emit trace entries for your setup actions
- The existing `setupForEffect` function in goldilocks.go — this sets up for the EFFECT side, your work sets up for the CONDITION side
