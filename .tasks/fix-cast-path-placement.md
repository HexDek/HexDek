# Task: Fix Category A — Cast-Path Placement (19 cards)

## Problem
Thor places cards directly on the battlefield without going through the cast→stack→resolve path. Cards with "if you cast it" ETB conditions never trigger because `wasCast` is false.

## Cards affected (19)
The One Ring, Tiamat, Eon Frolicker, Skitterbeam Battalion, Geological Appraiser, Cyclone Summoner, Breaching Leviathan, Transcendent Dragon, Light-Paws Emperor's Voice, Wild Pair, Acererak the Archlich, Wistfulness, Gruff Triplets, Kodama of the East Tree, Preston the Vanisher, Frodo Sauron's Bane, Acclaimed Contender, Oracle of Bones, Rankle and Torbran

## Fix
In `cmd/hexdek-thor/goldilocks.go` (and/or `main.go`), when setting up a card for testing:

1. **Detect cast-conditional ETBs** — scan the card's AST for trigger conditions that check `wasCast` or contain "if you cast it" / "if you cast it from your hand" patterns. Look at how the AST represents this (likely a flag on the trigger node or an `if_intervening` clause).

2. **Cast-path placement** — when a card has cast-conditional triggers, instead of directly placing it on the battlefield:
   - Put the card in seat 0's hand
   - Push it onto the stack as a cast (set `wasCast=true` on the card/permanent)
   - Resolve it through the stack pipeline so it enters the battlefield via the normal ETB path
   - This should cause "if you cast it" triggers to fire

3. **Handle inverse conditions** — Preston the Vanisher triggers when something WASN'T cast ("if it wasn't cast"). For these, the existing direct placement is correct. Don't change the path for inverse-cast cards.

4. **Handle "from your hand" specifically** — Cyclone Summoner, Breaching Leviathan, Wild Pair require casting FROM HAND specifically. Make sure the card starts in hand before the cast.

5. **Trace entries** — emit: `[N] CAST_SETUP: casting "Card Name" from hand (wasCast=true)`

Look at:
- `internal/gameengine/` for how casting works (stack push, resolve, ETB)
- The existing `testGoldilocksCard` flow for where to inject the cast path
- How `wasCast` is tracked on permanents (likely a flag or the card's entry path)

Test with: `go run ./cmd/hexdek-thor --card "The One Ring" --trace` — should now show the ETB trigger firing.
