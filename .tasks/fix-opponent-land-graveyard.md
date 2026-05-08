# Task: Fix Category B+D — Opponent Land Seeding + Graveyard State (13 cards)

## Problem B: Opponent board comparison (7 cards)
Cards that check "if an opponent controls more lands than you" need the opponent seat to actually have more lands. Thor's opponent seat currently has no lands.

Cards: Land Tax, Knight of the White Orchid, Claim Jumper, Loyal Warhound, Aerial Surveyor, Ghitu Journeymage ("if you control another Wizard"), Lux Artillery (30+ energy counters)

## Problem D: Graveyard state requirements (6 cards)
Cards that need creatures in graveyard or "a creature died this turn" state.

Cards: Ichorid, Oversold Cemetery (4+ creature cards in GY), Necromancy (reanimate target in GY), Compy Swarm ("if a creature died this turn"), Tombstone Stairwell (creatures in each GY), Grave Scrabbler

## Fix for Category B
In the conditional scaffolding or opponent auto-detect code:

1. **Detect land-comparison conditions** — AST contains "opponent controls more lands" or similar comparative clause
2. **Seed opponent lands** — when detected, give seat 1 six Plains (more than seat 0's typical 3-4). This satisfies the "more lands than you" check.
3. **Subtype conditions** — for Ghitu Journeymage ("if you control another Wizard"), place a Wizard creature on seat 0's battlefield alongside the test card.
4. **Trace:** `[N] CONDITION_SETUP: "opponent controls more lands" → seeded 6 Plains on seat 1`

## Fix for Category D
1. **Detect graveyard conditions** — AST trigger checks graveyard count or "died this turn"
2. **Populate graveyard** — put 4+ creature cards into seat 0's graveyard for cards like Oversold Cemetery
3. **"Died this turn" flag** — for Compy Swarm, actually destroy a creature (place a token, destroy it) so the game state records a death event this turn
4. **Reanimate target** — for Necromancy, ensure there's a creature card in the graveyard to target
5. **Trace:** `[N] CONDITION_SETUP: "4+ creatures in graveyard" → populated GY with 4 creature cards`

Look at:
- `cmd/hexdek-thor/conditional_setup.go` — the registry you're extending
- `cmd/hexdek-thor/opponent_autodetect.go` — for opponent seat manipulation
- `internal/gameengine/` for how graveyard population and "died this turn" tracking works

Test with: `go run ./cmd/hexdek-thor --card "Land Tax" --trace` and `--card "Oversold Cemetery" --trace`
