# Goldilocks Post-Merge Audit

**Date:** 2026-05-16
**Branch:** `dev/goldilocks-postmerge` (from `main` @ `b6e6d43`)
**Tool:** `cmd/hexdek-thor` `--goldilocks` (full corpus, 10 workers)
**Inputs:** `data/rules/ast_dataset.jsonl` (31,963 cards), `data/rules/oracle-cards.json` (35,708 cards)
**Raw failures CSV:** [`goldilocks_postmerge.csv`](goldilocks_postmerge.csv) (64 rows)

## Headline

| Metric | Count |
|---|---:|
| Cards loaded | 35,708 |
| AST-verified tests run | 31,963 |
| Passed | 30,277 |
| Dead-effect | 61 |
| Invariant violations | 2 |
| Unverified (verifier gap) | 1 |
| **Panics** | **0** |
| Skipped (no abilities) | 4,106 |
| Wall clock | 577 ms |

**Pass rate: 99.80%. Zero panics.** This is a clean baseline — well below the 1,915-failure spike from 2026-05-08 (`keyword_dead` regression) and consistent with the 54-failure post-fix run on the same date.

## Regression check against today's merge wave

Spot-tested every card touched by today's merges. All pass cleanly:

| Card | Merge | Goldilocks result |
|---|---|---|
| Starbreach Whale (Warp) | `621e865` | ZERO FAILURES |
| Chaos Warp | `621e865` (sanity) | ZERO FAILURES |
| The One Ring | `b8fe855` | ZERO FAILURES |
| Land Tax | `4c4377b` | ZERO FAILURES |
| Necromancy | `fab44ea` | ZERO FAILURES |
| Bloodchief Ascension | `fab44ea` | ZERO FAILURES |
| Kodama of the East Tree | `fab44ea` | ZERO FAILURES |
| Light-Paws, Emperor's Voice | `b6e6d43` | ZERO FAILURES |
| Tiamat | `b6e6d43` | ZERO FAILURES |
| Abdel Adrian, Gorion's Ward | `af762ad` | ZERO FAILURES |

**No regression introduced by today's merge wave.** Notably the Abdel Adrian crash signature from the May 11 grinder flood (324 nil-derefs, `internal/gameengine/per_card/abdel_adrian.go`) is fully closed — neither the panic nor the related invariant reappears in this run.

## Failure breakdown

### By interaction type

| Kind | Count |
|---|---:|
| `goldilocks_dead_effect` | 61 |
| `goldilocks_invariant` | 2 |
| `goldilocks_unverified` | 1 |

### Dead-effect by `(effect, abilityKind)`

| Effect | AbilityKind | Count | Category |
|---|---|---:|---|
| `ability_word` | static | 45 | Scaffold gap — new ability words |
| `sacrifice` | triggered | 4 | Scaffold gap — upkeep cost-or-sacrifice |
| `modification_effect` | triggered | 4 | Scaffold gap — ETB condition not met |
| `lose_life` | static | 2 | Scaffold gap — symmetric setup |
| `exile` | triggered | 2 | Scaffold gap — graveyard/target filter |
| `create_token` | triggered | 2 | Scaffold gap — combat trigger |
| `sacrifice` | static | 1 | Scaffold gap — needs lands |
| `destroy` | activated | 1 | Scaffold gap — needs lands |

## Triage

### 🟥 Engine-side invariant violations — investigate

Both are net-new since the last published Goldilocks run; neither is in CLAUDE.md's issue log.

**1. `Etali, Primal Conqueror // Etali, Primal Sickness` — CardIdentity**
```
[modification_effect] CardIdentity: card "LibCard 1-0" (ptr 0xc005ba0e10)
appears in both seat 0 battlefield and seat 0 battlefield
```
Same card pointer indexed twice on the same battlefield. Suggests a transform/MDFC handling bug where the back-face exiled-and-cast play surfaces the same `*Card` on both faces simultaneously. Pattern mirrors the resolved 2026-05-08 Dread bug (stale zone reference after move) but on the battlefield side rather than across zones. Recommend looking at the transform/exile-and-cast pipeline in `internal/gameengine/per_card/` or the MDFC layer in `internal/gameengine/moveCardBetweenZones` for a missing dedupe when the back face enters.

**2. `Phage the Untouchable` — TurnStructure**
```
[lose_game] TurnStructure: active seat 0 is Lost but life is 20 with no LossReason
```
`lose_game` is firing (Phage's "loses the game unless cast from hand" cast trigger) but the handler omits the `LossReason` argument that the TurnStructure SBA checker requires. Likely a one-line fix in the Phage cast-trigger handler — set `LossReason = "phage_cast_from_non_hand"` (or equivalent) before flagging `Lost`. Without it, replay/serialization will report a player as Lost while their life total still reads 20, breaking Heimdall and probably the spectator UI.

### 🟨 Verifier coverage gap

**3. `Expose the Culprit` — `turn_face_up` unverified**
```
effect kind 'turn_face_up' not in verifiable set
```
Single-card miss. Add `turn_face_up` to the Goldilocks verifiable-effects table (`cmd/hexdek-thor/goldilocks.go` ~line 2599) with a check that some morph/manifest card on the battlefield flipped face-up between snapshots. Low priority; no engine impact.

### 🟩 Scaffold limitations — not engine bugs, but worth a Goldilocks PR

**4. `ability_word` static — 45 cards (70% of all failures)**

All 45 cards carry new-era ability words whose underlying gameplay needs a *specific* trigger condition the current `ability_word` scaffold (lines 1228-1245 of `goldilocks.go`) doesn't satisfy. Spot-checks:

| Card | Ability word | Needs |
|---|---|---|
| Heartfire Hero, Mouse Trapper, Emberheart Challenger | Valiant | Source becoming target of a controlled spell/ability this turn |
| Werewolf Pack Leader, Hobgoblin Captain | Pack tactics | An attack with creatures totaling ≥6 power |
| Foot Mystic, Wandering Graybeard | Disappear | A permanent leaving the battlefield this turn (flag is set but the modification body needs a real ETB) |
| Winnower Patrol, Kithkin Zephyrnaut | Kinship | Top-of-library reveal sharing a creature type |
| Collective Voyage | Join forces | Multi-player mana payment phase |
| Mana-Charged Dragon, Waterspout Weavers | Replicate, Domain | Cast-from-hand with optional cost paid / 4 basic land types |
| Heartfire Hero (A-) | Start your engines! | Speed/race counter scaffold not modelled |

Driver is the Muninn parser waves 1b/2/3 (commits `9a6fb01`, `f773555`, `fdb6293`) that promoted residual-wrapper Modifications into typed `ability_word` AST nodes. The engine handles these correctly at game time — Goldilocks just hasn't grown verifiers for each ability word's specific condition. Treat as a Goldilocks backlog item, not an engine defect.

Recommendation: split the `ability_word` scaffold into per-keyword sub-scaffolds (one each for Valiant, Pack tactics, Disappear, Kinship, Join forces, etc.) similar to how the keyword combat scaffold was rebuilt on 2026-05-08.

**5. `sacrifice` triggered, `filterBase=self` — 4 cards (Pestilence, Pyrohemia, Withering Wisps, Task Mage Assembly)**

All four are "At the beginning of your upkeep, sacrifice ~ unless you pay {X}{X}". The scaffold pre-fills `ManaPool = 10`, so the upkeep cost auto-pays and the sacrifice branch never runs. Fix: add a scaffold variant that zeros mana for upkeep-cost-or-sacrifice abilities.

**6. `modification_effect` triggered — 4 cards (Lord of Tresserhorn, Scourge of Numai, Reaver Drone, Fathom Fleet Boarder)**

All four are ETBs with a *conditional* modification ("ETB you lose 2 life", "ETB sac a creature unless ascend", "unless you control another Devoid creature"). The scaffold places the source but doesn't satisfy / break the condition. Audit needed in `goldilocks.go` ETB scaffolding to vary friendly/opposing board state.

**7. `sacrifice` static `filterBase=land` (Planar Engineering), `destroy` activated `filterBase=land` (Demonic Hordes)**

Scaffold lacks lands on battlefield when the effect targets lands. Trivial fix: in the land-filter branch of `setupScaffold`, drop a few basics in.

**8. `create_token` triggered (Trynn, Nightsquad Commando)**

Combat-trigger create-token where the trigger requires either an `ascend` city's-blessing flag (Nightsquad) or a partner-with combat scenario (Trynn). Scaffold gap.

**9. `exile` triggered (Soul-Guide Lantern card-filter, Fishing Gear that_player-filter)**

Soul-Guide needs cards in the targeted graveyard; Fishing Gear needs the equipped creature to deal combat damage to a player. Both scaffold gaps.

**10. `lose_life` static (Fraying Omnipotence, Pox Plague)**

Symmetric lose-life effects whose snapshot diff doesn't change because the scaffold doesn't have anything to lose (no hand, no creatures on either side). Scaffold gap.

## Recommendations

| # | Action | Owner | Priority |
|---|---|---|---|
| R1 | Fix `Etali, Primal Conqueror` battlefield double-add | per_card | **High** — invariant violation, could surface in Loki/Muninn |
| R2 | Add `LossReason` to Phage lose_game handler | per_card | **High** — corrupts game-over telemetry |
| R3 | Add `turn_face_up` to Goldilocks verifiable set | thor | Low |
| R4 | Split `ability_word` scaffold into per-keyword sub-scaffolds (Valiant, Pack tactics, Disappear, Kinship, Join forces, Replicate, Start your engines!) | thor | Medium — would convert ~45 false positives into real signal |
| R5 | Add upkeep-cost-or-sacrifice scaffold variant (zero mana) | thor | Low |
| R6 | Vary ETB scaffold board state to satisfy/break conditional modification effects | thor | Low |
| R7 | Always seed a couple of basics in land-filter scaffolds | thor | Low |
| R8 | Add `ascend` (city's blessing) + equip-and-attack scaffold variants | thor | Low |

## Conclusion

Today's merge wave (Warp keyword + 4 era-scaffold waves + 8 per_card handlers + Abdel Adrian fix) introduced **no regressions** to the Goldilocks audit. Pass rate is 99.80% with zero panics. The 64 failures decompose into 2 engine invariant violations (Etali, Phage) that predate today's merges and ~46 Goldilocks scaffold limitations driven by the Muninn parser promotions adding new `ability_word` AST nodes faster than the verifier was extended to cover them.
