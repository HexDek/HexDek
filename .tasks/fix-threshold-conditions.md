# Task: Fix Category C+E — Life/Counter/Threshold Conditions (12 cards)

## Problem C: Turn-state tracking (6 cards)
Cards that check "if you gained life this turn" or "if you put a counter on a creature this turn" need those events to have happened earlier in the turn.

Cards: Lathiel the Bounteous Dawn, Witch of the Moors, Angel of Destiny (15+ more life than starting), Star Charter ("gained or lost life this turn"), Lasting Tarfire ("put a counter on a creature this turn"), Lord Jyscal Guado (same as Lasting Tarfire)

## Problem E: Board state thresholds (6 cards)
Cards that need N permanents, specific subtypes, or specific configurations.

Cards: Twilight Prophet (ascend — 10+ permanents), Acclaimed Contender ("if you control another Knight"), Great Hall of Biblioplex (needs 5 mana to activate), Birthing Ritual ("if you control a creature"), Smirking Spelljacker ("if a card is exiled with it"), Frodo Sauron's Bane (needs activated abilities resolved in sequence)

## Fix for Category C
In the conditional scaffolding registry:

1. **"Gained life this turn"** — before the trigger check, call `gameengine.GainLife(gs, seat0, 3)` or equivalent. This sets the internal "gained life this turn" flag.
2. **"Put a counter this turn"** — add a +1/+1 counter to a creature on seat 0's battlefield before the trigger fires.
3. **"15+ more life"** — set seat 0 life to 55 (40 starting + 15 more than default).
4. **"Lost life this turn"** — deal 1 damage to seat 0 or call LoseLife.
5. **Trace:** `[N] CONDITION_SETUP: "gained life this turn" → GainLife(seat0, 3)`

## Fix for Category E
1. **Ascend (10+ permanents)** — place 10 vanilla permanents on seat 0's battlefield before testing Twilight Prophet. Set the city's blessing flag if the engine tracks it separately.
2. **Subtype control** — for Acclaimed Contender, place a Knight creature on seat 0's battlefield.
3. **Exile association** — for Smirking Spelljacker, put a card in exile linked to the Spelljacker (imprint-style).
4. **Creature control** — for Birthing Ritual, ensure seat 0 has a creature (should already be the case in most setups, but verify).
5. **Trace:** `[N] CONDITION_SETUP: "ascend" → placed 10 permanents on seat 0`

Look at:
- `cmd/hexdek-thor/conditional_setup.go` — extend the condition→setup registry
- `internal/gameengine/` for how life gain tracking, counter placement, and ascend work
- The existing conditional scaffolding patterns for how conditions are detected and actions dispatched

Test with: `go run ./cmd/hexdek-thor --card "Witch of the Moors" --trace` and `--card "Twilight Prophet" --trace`
