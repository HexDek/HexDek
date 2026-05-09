---

kanban-plugin: board

---

## High Priority — Parser (100% Coverage Push)

*Empty — 100% coverage achieved (0 UnknownEffect across 31,963 cards)*


## High Priority — Engine

- [x] **Remaining 276 commander handlers** — generator improved (fuzzy slug resolution, hand-edit preservation, deterministic number extraction). NO_AST 73→0, 195/196 unhandled pool commanders templated. PR #1 merged 2026-05-09. #engine #per_card
- [x] **Pool-coverage handler batch** — 13 hand-written handlers selected by deck-pool inclusion: Arcades, Tifa Lockhart, Veyran, Shadow the Hedgehog, Eriette, Inalla, Eddie Brock//Venom, Tiamat, Mayael, Giada, Ghyrson Starn, Choco. PR #30 merged 2026-05-09. #engine #per_card
- [x] **Era 1 unification (FDN/DSK/BLB/OTJ/MKM)** — 12 templated commanders promoted from stub to custom logic: Mabel, Aesi, Bristly Bill, Byrke, Aminatou, Queen Marchesa, Kardur, The Swarmweaver, Rendmaw, Kona, Ezrim, Prime Speaker Zegana. PR #32 merged 2026-05-09. #engine #per_card
- [x] **Era 2 unification (CMM/LCI/WOE/MOM/ONE)** — 11 commanders promoted: Sliver Gravemother, Yenna, Felothar, Mondrak, Solphim, Drivnod, Zopandrel, Inalla (reprint), Mayael (reprint), Saheeli, plus one more. 18 new tests. PR #31 merged 2026-05-09. #engine #per_card
- [x] **Era 3 unification (SNC/BRO/DMU/NEO/CLB)** — 12 commanders promoted: Jetmir, Falco Spara, Lord Xander, Hidetsugu, plus 8 more. PR #33 merged 2026-05-09. #engine #per_card
- [x] **Era 4 unification (STX/MH2/AFR/MID/VOW/C19-C21)** — 7 templated commanders promoted from stub to custom logic: Galazeth Prismari, Lier (Disciple of the Drowned), Toxrill (Corrosive), Asmoranomardicadaistinaculdacar, Jadzi (Oracle of Arcavios), Silverquill (Disputant), Quandrix (Proof). Tiamat / Veyran / Acererak / Kalamax overlapped with prior batches. PR #34 merged 2026-05-09. #engine #per_card
- [x] **Era 5 unification (IKO/ZNR/KHM/CMR/C13–C18/pre-Modern)** — 13 templated commanders promoted from stub: Marchesa the Black Rose, Karador, Derevi, Mairsil, Yasharn, Charix, Kalamax, Chainer, Ruric Thar, Selenia, Yurlok, Sakashima, Araumi. PR #35 merged 2026-05-09. #engine #per_card
- [x] **Drop dead `gen_*` stubs** — 32 generator stubs removed where custom handlers fully supersede them. PR #36 merged 2026-05-09. #engine #per_card #cleanup

## Tracking — Cost-Unenforced Activated Abilities

> Per_card handlers that implement the *effect* of an activated ability but do not gate on the activation cost (mana, tap, sacrifice, exile, life payment, energy, etc.). When `InvokeActivatedHook` is called from a non-`ActivateAbility` path (test fixture, AI evaluator, replay rebuild), the AST cost dispatch is skipped — so these are best-effort but not provably correct under arbitrary AI activation. Once the engine learns a unified activated-cost gate, every open entry can drop in-handler `seat.ManaPool < N` checks.
>
> See `docs/structure-audit-report.md` for the generic gap caveat.

### Done — Era 5 unification (2026-05-09, PR #35)

- [x] **Selenia, Dark Angel** — "Pay 2 life" (only cost). Custom handler enforces.
- [x] **Chainer, Dementia Master** — "{B}{B}{B}, Pay 3 life" (life portion not in mana cost AST). Handler enforces 3-life payment.
- [x] **Araumi of the Dead Tide** — "{T}, Exile cards from your graveyard equal to the number of opponents you have" (the exile-from-GY portion is not in standard cost AST). Handler enforces graveyard-card exile.
- [x] **Mairsil, the Pretender** — ETB exile-and-tag cost is the resolution effect; activation copies bypass cage-counter validation. Handler tags caged cards for the engine-side hook to verify.

### Done — Era 1 + Era 2 cost gates (2026-05-09, dev/cost-unenforced-era1)

- [x] **Bristly Bill, Spine Sower** — `{3}{G}{G}` activated double now gates on `seat.ManaPool >= 5` via shared `payManaFromPool`.
- [x] **Ezrim, Agency Chief** — `{1}, Sacrifice an artifact` activated grant implemented; gates on mana + finds an artifact-to-sac before paying. Defaults to lifelink (combat-math winner).
- [x] **The Jolly Balloon Man** — `{1}, {T}, Activate only as a sorcery` copy-creature implemented with sorcery-speed + tap + mana gates. Refunds mana when no legal target found.
- [x] **Commander Mustard** — `{2}{R}{W}` activation pays 4 from pool, sets `mustard_soldier_attack_ping` flag, queues end-of-turn cleanup delayed trigger.
- [x] **The Master of Keys** — ETB now reads X from `gs.Flags["_master_of_keys_x_<seat>"]` (Walking Ballista pattern). Counters + mill no-op when X is absent so non-cost paths can't get free value.
- [x] **Kardur, Doomscourge** — goad flag now expires via a `next_upkeep` delayed trigger keyed to the controller's seat.

### Open — Era 1 (residual gaps not in scope for cost-gate PR)

- [ ] **Giada, Font of Hope** — `{T}: Add {W}` mana ability with "spend only on Angel" restriction: tap cost engine-enforced, restriction not gated by per_card. (Mana-system pipeline change.)
- [ ] **Aminatou, Veil Piercer** — enchantments-in-hand miracle grant: not wired into cast path. (Cast-pipeline change.)

### Done — Era 2 cost gates (2026-05-09, dev/cost-unenforced-era1)

- [x] **Sliver Gravemother** — Encore now adds the missing "Activate only as a sorcery" gate (`isSorcerySpeed`) on top of existing mana + graveyard-cost enforcement.
- [x] **Yenna, Redtooth Regent** — `{2}, {T}` copy now gates on sorcery-speed + untapped-source + ≥2 mana, paid only after a legal target is found (no cost burn on misclick).
- [x] **Felothar the Steadfast** — added `!src.Tapped` precondition on top of existing mana gate.
- [x] **Mondrak / Solphim / Drivnod / Zopandrel** — already gated on `seat.ManaPool >= 4` and two-other-creature availability; verified, no change needed in this PR.
- [x] **Inalla, Archmage Ritualist** — already gated on ≥5 untapped wizards; verified, no change needed in this PR.
- [x] **Mayael the Anima** — `{3}{R}{G}{W}, {T}` now gates on `seat.ManaPool >= 6` and `!src.Tapped`, taps Mayael on success.
- [x] **Saheeli, Radiant Creator** — already gates via `PayEnergy({E}{E}{E})`; verified, no change needed in this PR.

### Notes

> Colored mana enforcement (`{G}{G}` vs `{R}{R}`) is still engine-wide TODO — `seat.ManaPool` is a single integer, so the per-card gates above only verify generic cost sums. This is a defensive layer for non-engine callers (test fixtures, AI evaluator probes, replay rebuilds); the real gate runs through `ActivateAbility` when callers reach the engine path.

### Open — Other audited gaps

- [ ] **Azami, Lady of Scrolls** — "Tap an untapped Wizard you control" (tap-another cost). The "Complete" entry in the structure audit is misleading — only the draw effect is wired; the tap-a-Wizard cost is unpaid.
- [ ] **Bilbo, Birthday Celebrant** — life-payment activation for Birthday rolls.
- [ ] **Captain America, First Avenger** — discard-a-card shield-throw activation.
- [ ] **Ellie, Vengeful Hunter** — "Pay 2 life, Sacrifice another creature" — both portions unenforced in template.
- [ ] **Erebos, God of the Dead** — "{1}{B}, Pay 2 life" draw activation — life portion unenforced.
- [ ] **Phenax, God of Deception** — granted "{T}: Target player mills X" with no cost-enforcement on the granted ability.
- [ ] **Tasigur, the Golden Fang** — "Exile four cards from your graveyard" delve cost.
- [ ] **Ghen, Arcanum Weaver** — sacrifice-an-enchantment cost on the recursion activation.
- [ ] **Shilgengar, Sire of Famine** — sacrifice-a-creature drain cost.

#engine #per_card #cost-enforcement


## High Priority — Platform

- [x] **Mobile full pass** — fishtank telemetry hidden, curse section pushed to bottom, meta page padding trimmed for 375px. PR #3 merged 2026-05-09. #ui #mobile
- [x] **Mobile deck drilldown: curse data last** — flexbox `order: 999` on `.archive-curse-section`. PR #3 merged 2026-05-09. #ui #mobile
- [x] **Global glossary disclosure system** — ~50 entries, GlossaryTerm component (tap-to-expand, ARIA), wired into 6 pages. PR #4 merged 2026-05-09. #ui #ux #accessibility
- [ ] ~~**Curse Proficiency sigil**~~ — NIXED (radar plot is fine). #ui #design
- [x] **Action button context boxes** — ContextBox upgraded with persistent dismissal, 19 context boxes across all action surfaces. PR #8 merged 2026-05-09. #ui #ux #accessibility
- [x] **"Consider Cutting" rationale** — structured Go backend + expandable React panel. PR #10 merged 2026-05-09. #ui #deck #freya
- [x] **Value Engine rationale** — trigger/how-it-works/key-pieces per chain. PR #10 merged 2026-05-09. #ui #deck #freya
- [x] **Win Condition rationale** — combo pieces, conditions, resolution path. PR #10 merged 2026-05-09. #ui #deck #freya
- [x] **Deck clone** — rate limit 10/hr, cloned_from metadata, clone_log, two-step confirm, sign-in-to-clone UX. PR #7 merged 2026-05-09. #ui #platform
- [x] **Ephemeral spectator rooms** — RoomManager (856 lines), SpectateRoom.jsx (591 lines), full API, lifecycle tests. PR #5 merged 2026-05-09. #platform #spectate #backend #ui
- [x] **Reconnection countdown** — exponential backoff (1s-30s), 10 attempts, manual retry, ReconnectBanner component. PR #9 merged 2026-05-09. #ui #ux
- [x] **Magic link graceful flow** — BroadcastChannel cross-tab auth, console-style login animation, auto-close. PR #9 merged 2026-05-09. #ui #auth #ux


## High Priority — Learning Loop (Observability) — ALL PHASES DONE

*Ref: `docs/architecture-learning-loop.md` + `docs/architecture-hat-evolution.md`*
*Phase 1 (Heimdall), Phase 2 (Muninn+Huginn), Phase 3 (Huginn→Freya) all complete 2026-05-02. Loop validated live on DARKSTAR.*


## High Priority — Hat Intelligence — ALL LEVELS DONE

*Ref: `docs/architecture-hat-evolution.md` Levels 2-3*
*Level 2 (Combo Sequencer), Level 2.5 (State Machine), Level 3 (Genetic Curse) all complete 2026-05-02.*


## High Priority — Telemetry — DONE

*GA4 Health Pulse (Phase 6) complete 2026-05-02. Server-side + client-side telemetry wired.*


## High Priority — Calibration — DONE

*Floor + Ceiling calibration complete. Rating normalization (0-100 percentile) available via `NormalizeRating()`.*


## Medium Priority — Hat Advanced (Silver tier) — ALL LEVELS DONE

*Ref: `docs/architecture-hat-evolution.md` Levels 4-5.*
*Level 4 (Staged Decision Architecture), Level 5 (IS-MCTS) complete 2026-05-02/04.*


## Medium Priority — Hat Decision Making

- [x] **Equipment stacking intelligence** — commander Voltron routing (2x per-stack bonus), concentration kicker, Skullclamp/Sword-cycle recurrence, connect-payoff targeting. +480 lines, 20 tests. PR #2 merged 2026-05-09. #hat #equipment #voltron
- [x] **Equipment-specific target scoring** — fixed nil-equipment bug in ChooseTarget, stack-top recovery + single-equipment fallback. PR #2 merged 2026-05-09. #hat #equipment #targeting
- [x] **Equipment recurrence awareness** — Skullclamp pattern (tokens/1-toughness), Sword-cycle pattern (evasion premium), pump-rider synergy. PR #2 merged 2026-05-09. #hat #equipment #strategy
- [x] **Graveyard recursion awareness (non-reanimator)** — already shipped 2026-05-08 (`hasGraveyardRecursionValue` + `hasGraveyardRecursionEnabler`). Verified 2026-05-09. #hat #graveyard #strategy
- [x] **Zone-cast grant strategic valuation** — `hasActiveGraveyardCastGrant` + `hasGraveyardRecursionPotential` for Underworld Breach-style effects. PR #6 merged 2026-05-09. #hat #graveyard #zonecast
- [x] **Recursion-aware sacrifice evaluation** — new `SacrificeChooser` interface, persist/undying/unearth net-positive scoring, commander penalty. PR #6 merged 2026-05-09. #hat #sacrifice #recursion

## Medium Priority — Engine

- [x] **Temporal Pincer** — implemented 2026-05-04. SQLite schema, REST handlers, frontend wiring. #infra #platform
- [x] **BUG: Esika/Prismatic Bridge 9.2% WR** — `tryCastCommander` never set `CastingBackFace=true` for MDFC commanders. Bridge never reached battlefield. PR #15 merged 2026-05-09. #engine #bug #per_card
- [x] **BUG: Multiple B5 decks at 13-14% WR** — combo sequencer treated commanders in command zone as "missing" pieces + MDFC face aliases unrecognized. Command-zone awareness + DFC alias indexing + commander tax costing. PR #18 merged 2026-05-09. #engine #bug #combo
- [x] **34K corpus audit — DONE (initial run)** — 31,963 cards tested, 181 unique failures (99.4% card coverage). #engine #qa
- [x] **Corpus audit: draw handler gaps** — verified 2,032/2,032 PASS (was tests-run count, not failures). All draw tests already passing after keyword_dead fix 2026-05-08. #engine #handlers #draw
- [x] **Corpus audit: lifegain/lifeloss gaps** — verified stale (1,101 gain_life + 511 lose_life = 1,612 was tests-run, not failures). All passing. Re-ran `hexdek-thor --corpus-audit` 2026-05-09: 18,934 tests, 0 fail, 0 panic. #engine #handlers #life
- [x] **Corpus audit: damage gaps (1,095)** — verified stale via the same 18,934-test/0-fail Thor run that cleared discard/mill/buff and lifegain/lifeloss; original 1,095 was a tests-run count, not failures (2026-05-09). #engine #handlers #damage
- [x] **Corpus audit: discard/mill/buff gaps** — verified 498 discard / 303 mill / 2,109 buff all 100% pass across every era (2026-05-09). Original 305/148/81 figures were tests-run counts misread as failures, same shape as the draw stale-result. Report in `data/corpus-misc-verification.md`. #engine #handlers #misc
- [x] **Thor test harness: conditional trigger setup** — 14 new scaffold kinds in conditional_setup.go (gained-life, cast-spell, ETB, drawn-card, attacked, sacrificed, combat-damage, landfall, discarded, enchanted, opponent-lost-life, life-threshold, upkeep). (2026-05-07) #engine #qa #thor
- [x] **Muninn-Thor mismatch audit** — crossref.go: loads Muninn data + Thor failures, builds TP/FN/FP confusion matrix, computes recall/precision, writes markdown report. (2026-05-07) #engine #qa

## High Priority — Thor 2.0 (7174n1c, 2026-05-06) — ALL DONE

- [x] **Action traces** — `CardTraceRecorder` fluent builder with 8 phases (setup → panic), `TraceCollector` writes `.trace` files per card, failures-only mode, snapshot diff summaries. 27 tests. (2026-05-07) #thor #diagnostics
- [x] **Opponent auto-detect** — oracle text pattern analysis → 11 requirement categories → idempotent board-state enrichment for targeting effects. 30 tests. (2026-05-07) #thor #scaffolding
- [x] **Conditional trigger scaffolding** — 14 new scaffold kinds (gained-life, cast-spell, creature-ETB, drawn-card, attacked, sacrificed, combat-damage, landfall, discarded, enchanted-creature, opponent-lost-life, life-above/below-threshold, upkeep-phase). All three layers (detect/apply/trace) updated. 49 new tests. (2026-05-07) #thor #scaffolding
- [x] **Oracle Errata Pipeline** — `cmd/hexdek-oracle-sync/`: streaming Scryfall bulk download, field-level diff, markdown report, `--live`/`--dry-run`/`--diff-only`/`--report` modes. 27 tests. (2026-05-07) #thor #pipeline #infra
- [x] **Muninn cross-reference** — loads Muninn data, builds confusion matrix against Thor failures (TP/FN/FP), computes recall/precision, writes markdown report. 10 tests. (2026-05-07) #thor #qa
- [x] **Thor CLI wiring** — `thorFeatures` struct, 4 CLI flags (--trace, --trace-failures-only, --opponent-detect, --scaffold), conditional scaffolding wired into testInteraction. 6 integration tests. PR #12 merged 2026-05-09. #thor #cli


## Medium Priority — Platform

- [x] BOINC-style distributed compute — `cmd/hexdek-contrib` Go binary, WebSocket protocol, chunked work dispatch, spot-checking, contributor credits. 2,260 lines, tests. PR #17 merged 2026-05-09. #distributed
- [x] **Deterministic seed capture (anti-cheat Phase 1)** — `internal/seedcontract`: HMAC-SHA256, domain-separated prefixes, layered Seal/Sign/Verify, `DeriveContractKey`. Wired into all 3 game paths. 14 tests. PR #14 merged 2026-05-09. #anticheat
- [x] Deterministic replay anti-cheat — Phase 2 schema (PR #20), spot-check scheduler + cauterize service (PR #23), spot-check Service orchestrator (PR #28). All merged 2026-05-09. #anticheat
- [x] Statistical anomaly detection — `StatisticalAuditor` wired into game loop, per-contributor 3σ flagging on win rate/game length/turn variance. PR #19 merged 2026-05-09. #anticheat
- [x] Credit economy — `internal/credits` package: free tier (3 gauntlets/day), credits for compute, 402 paywall, concurrency-safe ledger, gauntlet endpoint gated. 14 tests. PR #21 merged 2026-05-09. #economy
- [x] Stream/narrator layer — game state → visual renderer → Twitch/OBS output. Shipped via StreamOverlay + NarratorOverlay. #stream
- [x] **Stream/narrator OBS overlay** — `deriveStreamPhase` three-act narrative captions (OPENING/MIDGAME/CLIMAX/CURTAIN), configurable seat order via `?seats=`, `?phase=off` toggle. PR #16 merged 2026-05-09. #stream #ui


## Low Priority — Hat Research (Bronze tier) — MOSTLY DONE

*Ref: `docs/architecture-hat-evolution.md` Levels 6-7 + Skunkworks.*
*Level 6 (Neural Position Evaluator), Level 7 (Self-Play Loop), Skunkworks (Tesla/Feynman/Lovelace/Ive/Watts) all complete 2026-05-02.*



## Tracking — Effect-Correct, Cost-Unenforced

*Cards where the per-card handler covers the effect but doesn't enforce the activation cost. Batch these for custom handler work.*

### Era 4 unification (STX, MH2, AFR, MID, VOW, C19-C21) — surfaced 2026-05-09

- [ ] **Asmoranomardicadaistinaculdacar** (MH2) — alt-cast cost `{B/R}` "as long as you've discarded a card this turn" not enforced; the cast pipeline doesn't expose the per-card alt-cost gate. #engine #per_card #cost_unenforced #era4
- [ ] **Galazeth Prismari** (STX) — `{T}: Add one mana of any color. Spend this mana only to cast an instant or sorcery spell.` — the spend-restriction isn't enforced in the mana system; the tap-add path is generic. #engine #per_card #cost_unenforced #era4
- [ ] **Lier, Disciple of the Drowned** (MID) — graveyard-cast permission for instants/sorceries and the "would be put into graveyard → exile instead" replacement aren't on the per-card hook path. Uncounterable mark is enforced via `CostMeta` when the stack item is in trigger ctx. #engine #per_card #cost_unenforced #era4
- [ ] **Jadzi, Oracle of Arcavios** (STX) — magecraft top-of-library reveal + cast-for-`{1}` alt cost not modeled; nonland cards route to hand as a stand-in. The DFC back-face Journey to the Oracle isn't wired. #engine #per_card #cost_unenforced #era4
- [ ] **Quandrix, the Proof** (STX/C21) — from-command-zone detection is a best-effort flag check; if the cast pipeline doesn't stamp `from_command_zone` on the new permanent, the X-counter distribute clause silently skips. #engine #per_card #cost_unenforced #era4
- [ ] **Silverquill, the Disputant** (STX/C21) — same from-command-zone detection gap; the optional "attach Aura/Equipment to Silverquill" rider is not modeled (Aura/Equipment attach isn't on the per-card hook path). #engine #per_card #cost_unenforced #era4
- [ ] **Toxrill, the Corrosive** (VOW) — `{U}{B}, Remove a slime counter from a creature: Draw a card.` — mana cost paid through the engine's mana system, but the sorcery-speed gate is approximated by phase-string match. Slime counter removal is enforced. #engine #per_card #cost_unenforced #era4



## Low Priority

- [x] **i18n — catalog translated** — 52 UI keys translated into 7 locales (es, de, fr, pt, ja, ko, zh). MTG terminology localized per community convention. Remaining: migrate ~450 hardcoded JSX strings to useT() + Scryfall localized card names. PR #11 merged 2026-05-09. #platform
- [ ] **Spectate background gradient morph** — when active commander / seat changes, the color-identity background gradient should slow-morph (CSS transition ~1-2s ease) to the next color instead of snapping instantly. #ui #spectator #polish
- [ ] Multi-format support beyond Commander (Modern, Legacy deck ratings) #engine
- [ ] **Tournament prize pools** — hat-vs-hat bracket tournaments with cash prizes (1st/2nd/3rd/4th splits). Starcraft model: deckbuilding is the skill, hat execution is the layer. Legally sound as skill competition (no entry-fee model safest, donations-funded). Needs: bracket system, payout logic, age verification (18+), tax reporting (>$600). #platform #economy #future


## Done — Session 2026-05-09 (QA/Cleanup Pass)

- [x] **Drop 32 dead `gen_*` stubs** — generator stubs removed where custom handlers fully supersede them. Eliminates double-firing from stacked registrations. PR #36 merged 2026-05-09. #engine #per_card #cleanup
- [x] **TODO board reconciliation** — wave 3 sync aligning board with merged era passes and QA results. PR #37 merged 2026-05-09. #docs
- [x] **AddCounter normalization** — 6 custom commander handlers migrated from manual `perm.Counters[kind]++` to `perm.AddCounter()` (nil-safe, floors at 0). PR #38 merged 2026-05-09. #engine #refactor
- [x] **Test backfill: Era 1 + Era 4 handlers** — 36 new tests across era-pass handlers that shipped without test coverage. PR #39 merged 2026-05-09. #engine #tests
- [x] **Build hygiene cleanup** — 35 staticcheck fixes (unused params, unnecessary type conversions, unreachable code). PR #40 merged 2026-05-09. #engine #cleanup
- [x] **Handler collision resolution** — 45 gen/custom double-firing conflicts resolved (custom wins, gen_ removed). Fixed **Nicol Bolas double-discard bug** (two ETB handlers causing opponents to discard twice). Wired **Grunn the Lonely King** (handler existed but was never registered). PR #41 merged 2026-05-09. #engine #per_card #bug

*6 PRs (#36-41), ~150 files touched. Key impact: Bolas double-discard was a gameplay-visible bug affecting every pod with Bolas.*


## Done — Session 2026-05-08 (Engine Deep Audit + Structural Systems)

- [x] **Deep sweep: 34 missing FireCardTrigger dispatch points** — dead triggers (targeted, card_exiled, zone_change, scry, proliferate, etc.) across 15 engine files. CI lint test `TestAllRegisteredTriggersAreDispatched` prevents future dead triggers. (2026-05-08) #engine #dispatch
- [x] **CheckEnd after 54 lethal damage sites** — 42 per_card handlers silently ignoring lethal opponent damage. Games could continue past death. Critical correctness fix. (2026-05-08) #engine #bug #critical
- [x] **AddCounter centralization** — 30 manual `perm.Counters[kind]++` sites migrated to `perm.AddCounter()` (nil-safe, floors at 0). (2026-05-08) #engine #refactor
- [x] **MoveResult return type** — `MoveCard` now returns `MoveResult{FinalZone, Permanent}` instead of bare string. Callers can access created permanents. (2026-05-08) #engine
- [x] **ExileLinked infrastructure (CR §406.7)** — `ExileLinked()` / `ReturnLinkedExile()` for O-Ring/Fiend Hunter/Knowledge Pool patterns. `Card.ExiledByTimestamp` + `Permanent.LinkedExile` fields. (2026-05-08) #engine
- [x] **ZoneCastPermission duration/expiry** — Duration, GrantTurn, SourceTimestamp, SpendAnyColor fields. `ExpireZoneCastGrants` at end-of-turn cleanup. Fixed bug where impulse draw grants (Prosper, Narset, Urza, etc.) persisted forever. 11 handlers upgraded. (2026-05-08) #engine #bug #critical
- [x] **Adventure ZoneCastGrant wiring (CR §715.4)** — `CastAdventure` now calls `RegisterZoneCastGrant` so creature half is castable from exile. Was previously half-broken. (2026-05-08) #engine #bug
- [x] **Prepared mechanic (§702.168)** — `Permanent.Prepared` bool + `Unprepare()` helper. 3 handlers: Abigale (upgraded from flags), Tam Observant Sequencer (landfall→draw+life), Lluwen Exchange Student (ETB/activate→Pest token). (2026-05-08) #engine #strixhaven
- [x] **Paradigm mechanic** — `GameState.ParadigmExile` tracking + `ResolveParadigmCopies` at first main phase. 5 handlers for all Strixhaven paradigm cards (Decorum Dissertation, Echocasting Symposium, Germination Practicum, Improvisation Capstone, Restoration Seminar). (2026-05-08) #engine #strixhaven
- [x] **9 new tests** — adventure grant registration, Prepared field lifecycle, paradigm exile tracking, ResolveParadigmCopies copy-cast. (2026-05-08) #engine #tests

*10 commits, ~2,500 lines across ~100 files. Key impact: ZoneCast duration fix + CheckEnd fix affect ~15% of commander pool each.*


## Done — Session 2026-05-05 Evening (UI/UX Sprint)

- [x] **Narrator enrichment** — 15 new event kinds surfaced (untap, tap, discard, scry, surveil, bounce, flicker, equip, etc), coalescing (consecutive same-seat events merge), dedup (cast suppresses ETB), per-seat color-coded log borders, turn separators. LogEntry enriched with source/targets/amount/count for animation rigging. (2026-05-05) #ui #spectator
- [x] **Card popup redesign** — click-to-open fullscreen overlay with actual Scryfall card image, MTG color-identity tinted backdrop (W=gold, U=blue, B=black, R=red, G=green), dimmed blur behind. Replaced DIY reconstructed card popup. VIEW CARD PAGE button. (2026-05-05) #ui #design #cards
- [x] **Volcmap pure gradient** — removed ISO contour lines from heatmap, pure gradient now. (2026-05-05) #ui #spectator
- [x] **HexELO display swap** — spectator seat display + ELO sidebar now show hex_rating/hex_delta instead of TrueSkill. (2026-05-05) #ui #spectator
- [x] **Card names linkable in log** — all card names in narrator/raw log wrapped in CardLink (hover preview + click to card page). 19 verb patterns covered. (2026-05-05) #ui #spectator
- [x] **Play View CTA removed** — redundant with fishtank embed. (2026-05-05) #ui
- [x] **SYS.BUILD version removed** — removed from header on desktop + mobile. (2026-05-05) #ui
- [x] **Mobile header 2-row layout** — search + theme + lang + auth share one row below nav. (2026-05-05) #ui #mobile
- [x] **Lang select dropdown** — proper `<select>` with 2-char codes (EN, ES, JA, etc) instead of cycling. (2026-05-05) #ui
- [x] **Search input reduced** — 30% smaller trigger, expandable overlay handles full search. (2026-05-05) #ui
- [x] **Tape bar hidden on mobile** — decorative info strip between nav and content removed at mobile breakpoint. (2026-05-05) #ui #mobile
- [x] **Amiibo → Curse rename** — 322 references across 25 files renamed (Go, React, docs, API). `data/amiibo/` → `data/curse/` on production. (2026-05-05) #engine #ui #legal
- [x] **Rankings + Meta merged** — single page at `/leaderboard` with RANKINGS/META tab toggle. META removed from main nav. `/meta` redirects. (2026-05-05) #ui #nav
- [x] **MY DECKS / ALL DECKS toggle** — always visible for logged-in users. Default to MY DECKS. Prefix matching for owner slug mismatch (joshua→josh). (2026-05-05) #ui
- [x] **Spectator telemetry hidden on mobile** — fishtank-only view on mobile, full telemetry on desktop/tablet. (2026-05-05) #ui #mobile
- [x] **Card page scrim overlay** — 14px blur frosted glass, heavy gradient opacity for text readability over art. (2026-05-05) #ui #design
- [x] **Recent games panel (FT.D)** — last 20 games in spectator sidebar with winner, turns, end reason, REPORT → link. Live via WebSocket. (2026-05-05) #ui #spectator
- [x] **Multi-scroll-pane fix** — removed independent `overflow: auto` from deck archive, leaderboard, dashboard, spectator, card page. Single page scroll via AppShell. (2026-05-05) #ui
- [x] **Nav control height normalization** — all appbar controls (search, theme, lang, auth) set to 24px height via shared CSS. Removed inline style overrides. (2026-05-05) #ui
- [x] **Username removed from header** — only LOGOUT button remains. (2026-05-05) #ui
- [x] **Accessibility contrast** — mana pip drop shadows, hero tags dark frosted backdrop, ADD FRIEND dark backdrop + blur, summary text shadow. (2026-05-05) #ui #a11y
- [x] **Curse data moved below tutor targets** — in deck drilldown. (2026-05-05) #ui
- [x] **2-layer footer** — top row: live connection status (green/yellow/red dot + state text) + OPEN SOURCE motto + user email. Bottom row: ABOUT/BUG/DONATE/DISCORD links. (2026-05-05) #ui
- [x] **Data consistency fix** — new `/api/owner/{owner}/stats` + `/api/owner/{owner}/games` endpoints querying SQLite directly. deck_key backfill on startup for old seat rows. Operator profile now shows real aggregate stats. (2026-05-05) #api #data
- [x] **Curse lifetime total** — added `TotalGames` field (never resets), seeded from `GenCount*100+GameCount` for existing pools. Frontend displays lifetime instead of cycle counter. (2026-05-05) #engine #curse
- [x] **Deck ownership enforcement** — DELETE/PUT/PATCH require `X-HexDek-Owner` header matching deck owner. 403 for non-owners. Frontend sends header automatically. (2026-05-05) #api #security
- [x] **Custom display names** — import flow persists name to deck_meta. PATCH rename for owners. Default display: commander name. (2026-05-05) #ui #api
- [x] **Deck hero button sizing** — FRIEND/SHARE/COMPARE normalized to matching padding + gap. (2026-05-05) #ui
- [x] **Operator profile owner slug fix** — prefix matching against ELO data resolves joshua→josh mismatch. (2026-05-05) #ui #bug
- [x] **Card page hero mobile fix** — hero body `position: relative` on mobile, card renders at 200px/160px. (2026-05-05) #ui #mobile

## Done — Session 2026-05-05 Day

- [x] **Deck import flow** — unified ImportModal with 3 input modes (paste/URL/file), real-time card validation against corpus, inline error surfacing, Freya progress indicator, success redirect with toast. Auth-gated. Replaces old piecemeal hooks. (2026-05-05) #ui #import
- [x] **Card performance tracking** — `GET /api/card-stats/card/{cardName}` + `/by-commander`. Per-card win rate, inclusion rate, top commanders, bracket distribution. CardPage frontend panel added. (2026-05-05) #engine #analytics #cards
- [x] **Ceiling calibration** — `cmd/hexdek-ceiling` CLI: scans B5 decks, cross-refs ELO + Curse fitness, runs focused gauntlet with accelerated evolution. `NormalizeRating()` maps ELO to 0-100 (floor=0, ceiling=100). `SkillPercentile()` letter grades. (2026-05-05) #rating #calibration
- [x] **Hat evaluator P/T migration** — 23 call sites in yggdrasil.go + poker.go migrated from raw `p.Power()`/`p.Toughness()` to `gs.PowerOf(p)`/`gs.ToughnessOf(p)`. Evaluator now respects Layer 7 continuous effects. (2026-05-05) #engine #layers #hat
- [x] **Layer 3 text-changing handlers** — full framework: 5 handlers (Swirl the Mists, Mind Bend, Magical Hack, Trait Doctoring, Painter's Servant), AST-driven registration, 14 tests. (2026-05-05) #engine #layers
- [x] **Expand layer dispatch** — Caged Sun + Gauntlet of Power → Layer 7c continuous effects. March of the Machines → Layer 4 (type-changing) + Layer 7b (P/T = CMC). (2026-05-05) #engine #layers
- [x] **BUG: MDFC permanent_types battlefield entry** — added proper `"battlefield"` case to `moveToZone`. EnsureBattlefieldFrontFace for MDFC type correction, CardCanEnterBattlefield gate, proper Permanent wrapper with ETB triggers. Eliminates ~80% of zone_accounting Feynman violations. 5 tests. (2026-05-05) #engine #bug #mdfc
- [x] **Genetic→Neural distillation** — Curse→neural feedback loop: HarvestHighFitness (top-quartile DNA), EnrichWithDNA (fitness-weighted training samples), SeedDNAFromManifest (warm-start new pools from archetype centroids), TryCycle (non-blocking 30min cooldown, reseed underperformers). 17 tests. (2026-05-05) #research #selfplay
- [x] **N-card combo line detection** — Huginn N-tuple pipeline fully wired: `DetectCoTriggerNTuples()` → `PersistRawNTuples()` → `IngestNTuples()` → `tier3_ntuples_for_freya.json`. CLI ingest/prune/stats/list commands added. Freya reads both pairwise and N-tuple exports. (2026-05-05) #engine #huginn #combo
- [x] **Muninn persist batching** — tournament runner wired to `Batcher` in all 3 paths (Run, runPool, runLazyPool). `feedBatcher()` streams parser gaps, crashes, concessions, dead triggers per-game. Auto-flush every 30s/100 games. `persistPostTournament()` for non-Muninn data. (2026-05-05) #engine #performance
- [x] **Feynman outlier fixes** — zone accounting: asymmetric tolerance `diff < -3 || diff > 20` (copy/clone positive diffs normal, missing cards = real bugs). game_end: turn-capped games (≥80 turns) downgraded to "info". 4 new tests. (2026-05-05) #engine #hat #feynman
- [x] **Gauntlet pool fix** — threshold 100→80, diagnostic logging for filtered decks (parse/noCmd/small/banned breakdown). Pool 1200→1292 decks. All 7174n1c decks now in engine. (2026-05-05) #engine #bug
- [x] **Tesla ExtractPivot panic fix** — guard against `winnerSeat=-1` (turn-cap draw games with no winner). Was crashing server during gauntlet runs. (2026-05-05) #engine #bug
- [x] **Tabs on deck drilldown** — Analysis / Deck List / Achievements tab layout on deck pages (2026-05-05) #ui #deck
- [x] **Gauntlet button moved under Freya** — gauntlet trigger relocated with error state display for decks not in engine pool (2026-05-05) #ui
- [x] **Win lines collapse** — 8 visible + "show more" toggle for long win-condition lists (2026-05-05) #ui #deck
- [x] **Import auth gate** — sign-in required before deck import (2026-05-05) #ui #auth
- [x] **Mobile hero layout** — card-centric 62/38 split for deck drilldown on mobile (2026-05-05) #ui #mobile
- [x] **Header art blur** — backdrop-filter 4px on deck page header art (2026-05-05) #ui #deck #design
- [x] **Card tap-to-popup on mobile** — popup first, CTA to full page (2026-05-05) #ui #mobile
- [x] **Compare page mobile fix** — stacked heroes, 1fr 100px 1fr stats grid for 375px viewport (2026-05-05) #ui #mobile


## Done — Session 2026-05-04 Day

- [x] **Curse display on deck page** — `CurseDisplay` component renders per-deck DNA pool with 7-axis personality radar + 20-cell DimStats weight heatmap, generation count, best fitness, fitness sparkline (commit bf0f73d, 2026-05-04) #ui #curse #design
- [x] **Curse fitness sparkline polish** — switched sparkline to per-generation best across last 20 generations (commit ea249a4, 2026-05-04) #ui #curse
- [x] **BUG: CursePanel `fitnessByRank` variable** — panel hardened against null / partial DNA snapshots; tile rendering survives empty-pool + missing-fitness shapes (commit 8dd7e72, 2026-05-04) #ui #curse #bug
- [x] Negative ELO shame badges — MID/DOWN BAD/COOKED/PACK IT UP/UNINSTALL ladder + Wall of Shame bottom-10 panel (2026-05-04) #ui #fun
- [x] **Achievement badges** — milestone badges (first 10/100/1K users), rare/commendable action badges (first blood, comeback from <5 life, perfect sweep, etc). Earned-badge showcase on deck pages (commit ba6db99, 2026-05-04) #ui #badges #design
- [x] **Volcano map smooth transition** — rAF-based heatmap interpolation, CSS transitions on seat-art opacity/filter (2026-05-04) #ui #spectator
- [x] Operator platform page/tab — `/operator` profile page with deck shelf, match history, friends panel (commit e4d61b1, 2026-05-04) #ui #platform
- [x] Friends system + player profiles — `/friends` pub-model page with search, add, browse; bidirectional friend list (commit a609816, 2026-05-04) #ui #social
- [x] Bracket-stratified leaderboard tabs — filter by B1-B5, separate rankings per bracket + band labels (2026-05-04) #ui
- [x] Game Changer cards list on deck page — GC card names persisted to strategy.json + "GAME CHANGERS" panel with art thumbnails (2026-05-04) #ui
- [x] **BUG: Ajani Nacatl Pariah 74% WR** — PW threat scoring fix shipped (commit 217927f, 2026-05-04) #engine #bug #hat
- [x] **BUG: MDFC permanent_types deep fix** — parser-side fix: deckparser extracts back-face metadata for all MDFCs (commit 26b88ed, 2026-05-04) #engine #bug #mdfc
- [x] **Mjolnir/Gungnir/Ragnarok routing** — formal 3-tier decision dispatcher (commit 1fc145f, 2026-05-04) #engine #hat #staged
- [x] Mobile-friendly leaderboard — 375px responsive pass (commit 136fa58, 2026-05-04) #ui
- [x] Donations page BOINC/ads buttons — real BOINC card + Support Dev card (commit 2f45e1a, 2026-05-04) #ui
- [x] Report analysis placeholder — mocked timeline replaced with real game data (commit cc2bf2d, 2026-05-04) #ui

### Legality Flag (7174n1c — 2026-05-02) — DONE 2026-05-04

- [x] **Persist legality in strategy JSON** — `legality` field in Freya output (2026-05-04) #engine #freya
- [x] **Legality badge on deck cards** — green ✓ / red ✗ next to bracket (2026-05-04) #ui #legality
- [x] **Legality section on deck info panel** — LEGAL/ILLEGAL status + expandable violation list (2026-05-04) #ui #legality
- [x] **Legality filter on deck browse** — ALL/✓LEGAL/✗ILLEGAL filter tags (2026-05-04) #ui #legality
- [x] **Fix Meglin phantom metadata** — `resolveDeckMetadata` fallback to COMMANDER header + Freya bracket (commit 622b889, 2026-05-04) #data

### Freya → UI Wiring Pass — DONE 2026-05-04

*14 fields computed by Freya wired to frontend:*
- [x] Star cards, cuttable cards, power percentile, commander synergy, commander themes #ui #wiring
- [x] Vulnerable-to warnings, meta matchups, mana base grade, keepable hand %, interaction profile #ui #wiring
- [x] Card roles grid, finisher cards callout, color demand heatmap, emergent synergies #ui #wiring
- [x] Persist: legality report, curve warnings, color mismatch, combo notes #engine #freya

### UX Overhaul (Ive/Jobs/Watts Quorum) — DONE 2026-05-04

- [x] Nav restructure (8→5-6 tabs), DASH→MY DECKS, contextual access, ABOUT→footer #ui #nav
- [x] Fishtank embed on home, "Upload My Deck" CTA #ui #home
- [x] Commander color-identity theming, full-bleed art, card grid default, personality blurb #ui #deck #design
- [x] Deck library as visual shelf #ui #deck #library
- [x] Universal search bar #ui #search
- [x] One-tap auth (contextual modal) #ui #auth
- [x] Share = link (clipboard + OG meta + CardLink wiring) #ui #social


## Done — Session 2026-05-04 Night

*All commits land on the same day; this group captures the back-half push beyond the daytime ship list above.*

- [x] **Card Page (`/cards/:cardName`)** — dedicated screen with Scryfall art, mana cost, oracle text, type line, set + rarity, plus per-card Freya appearance stats. Lazy-mounted route in App.jsx. (commits 4349bd9 + f4ae8b0, 2026-05-04) #ui #cards
- [x] **Card Popup component** — hover/tap preview with art + stats, attached to deck-list card names; trigger uses help-cursor and tolerates touch (commit e32f147, 2026-05-04) #ui #cards
- [x] **Card search + detail API** — `GET /api/cards/search?q=` + `GET /api/cards/{name}` backed by an in-memory index; oracle loader bumped to capture Scryfall `set` field (commit a37e3bd, 2026-05-04) #api #cards
- [x] **CardLink component + universal wiring** — single helper renders `<Link to="/cards/:cardName">` with stopPropagation; `linkifyAction` parses log strings against engine-templated patterns. Wired into DeckArchive/CardRolesGrid/SearchBar/Spectator/GameBoard (commit 3a253d0, 2026-05-04) #ui #cards
- [x] **Temporal Pincer** — anonymous UUID cookie → session tracking → on auth, stitch all anon device UUIDs to authenticated profile. SQLite schema, REST handlers, frontend wiring (commits b27f86f + b988507 + eb2dd26, 2026-05-04) #infra #platform
- [x] **Mobile responsiveness pass (375px)** — CardPage, CardPopup, search overlay, profile flag, fishtank embed (commit 8748b52, 2026-05-04) #ui #mobile
- [x] **MDFC fix v1: tryPlayLand back-face swap** — `SwapToBackFace` in tryPlayLand (commit fa018fd, 2026-05-04) #engine #mdfc
- [x] **MDFC fix v2: SwapToBackFace at every battlefield-entry path** — broader resolve-time entry points (commits 29c768d + 13ed4b3, 2026-05-04) #engine #mdfc
- [x] **game_end fix** — turn-cap leader-determination marks below-life survivors `Lost` (commit c2282f6, 2026-05-04) #engine #bug
- [x] **zone_accounting fix** — defensive sweep + library/command_zone arms (commit fb100ce, 2026-05-04) #engine #zones
- [x] **34+ new commander handlers** — 10 manual + 12 manual + 14 MDFC aliases (commits f464b43 + 49c882f + d1294ea, 2026-05-04) #engine #per_card

*Items not yet committed:*
- Spectator narrator — `hexdek/src/components/NarratorOverlay.jsx` exists and is imported by Spectator.jsx, but the file itself isn't committed yet.


## Done — Older

- [x] **Holy documentation pass** — 8 new architecture docs + Learning Loop pipeline doc. Fixed 3 stale docs. 47→47 current docs. (2026-05-04) #docs
- [x] **Grinder memory leak fix** — Heimdall `obsBuf` unbounded growth removed (2026-05-04) #engine #performance
- [x] **Depression concession removal** — Score-based conviction concession removed entirely (2026-05-04) #engine #hat
- [x] **Feynman Oracle false positive fixes** — hasCantLoseEffect(), §800.4a cards_left_game, Zombie Army token fix (2026-05-04) #engine #hat
- [x] **Bracket filter leaderboard** — B1-B5 filter tabs + band labels, live via WebSocket (2026-05-04) #ui
- [x] **Floor calibration decks** — 3 decks (Golos/Child of Alara/Isamaru) at `data/decks/calibration/` (2026-05-02) #rating #calibration
- [x] Watts confidence threshold dial + Shannon entropy tracking (2026-05-04) #engine #hat #staged
- [x] IS-MCTS implementation + trigger conditions (2026-05-02) #engine #hat #mcts
- [x] Heimdall Phase 1-3 (seed capture, Muninn+Huginn wiring, Huginn→Freya pipe) all complete (2026-05-02) #engine #heimdall
- [x] Hat Level 2 (Combo Sequencer) + Level 2.5 (State Machine) + Level 3 (Genetic Curse) all complete (2026-05-02) #engine #hat
- [x] GA4 Health Pulse (server + client telemetry) (2026-05-02) #infra #telemetry
- [x] Neural Position Evaluator (Level 6) + Self-Play Loop (Level 7) (2026-05-02) #research #neural #selfplay
- [x] Tesla Causal Graphs + Feynman Oracle + Lovelace Composer + Ive Spectator + Watts Soul Layer (2026-05-02) #research #skunkworks
- [x] Jhoira suspend, Lich's Mastery, Ulrich transform, Wayward Servant ETB (2026-05-01) #engine
- [x] Coat of Arms layer 7, Concession diagnostics, Dungeon tracking, Battle/Siege protectors, Speed mechanic (2026-05-01) #engine
- [x] Layer 3 text-changing stub (intentional no-op — no meta demand) (2026-05-01) #engine #layers
- [x] opponentLikelyHasWrath expansion → `wrathProbability()` graded float64 (2026-05-01) #engine #evaluator
- [x] Partner-aware mulligan, Transform recognition, Tournament stats emission (2026-05-01) #engine
- [x] PartnerSynergy + ActivationTempo + ToolboxBreadth evaluator dimensions (2026-05-01) #engine #evaluator
- [x] Threat trajectory prediction, UCB1 exploration factor, Dynamic weight rescaling (2026-05-01) #engine #evaluator
- [x] Light mode toggle, Curve analysis UI, Deck page auto-refresh on Freya push (2026-05-01) #ui
- [x] Exile-cast pattern (Voidwalker/Emry/Urza), Clone/copy handlers (2026-05-01) #engine
- [x] ELO-bracket correlation re-check (485K games), Soulshift keyword (2026-05-01) #engine
- [x] Opponent graveyard threat tracking, Color-aware mana sequencing (2026-05-01) #engine #evaluator
- [x] Jeska's Will, Panoptic Mirror (2026-05-01) #engine
- [x] P1: 100% parser coverage, Graveyard-leave hook, Land-tap hook, Cast observer (2026-05-01) #engine
- [x] Hat: Poison/infect awareness, Planeswalker threat scoring, Alternate wincon awareness (2026-05-02) #engine #hat
- [x] BUG: Rosnakht 2.5% WR, BUG: Tazri 13% WR (2026-05-02) #engine #bug
- [x] Bracket-aware tournament grinder, Keyword stubs audit, Cleanup: internal/rules/ (2026-05-01) #engine
- [x] P12: TurnFaceUp, P6: Saga ETB, P2: Ability word extension (2026-05-01) #parser #engine
- [x] SBA 704.5r counter limit, AI autopilot, ELO reset #2 (2026-05-01) #engine
- [x] Handler coverage push 66→447, 53/53 GC handlers (2026-05-01/04-30) #engine #per_card
- [x] Spectator UI scroll fix, Turn bar redesign, Loss reason display (2026-05-01) #ui
- [x] Stax lock wiring, CreateToken hook (2026-04-30) #engine
- [x] Dual-track ELO, HexELO drift detection (2026-04-30) #rating
- [x] Partner commander casting priority, Sacrifice-as-value overhaul, Sandbagging exemption, Reanimate activation awareness (2026-05-01) #engine
- [x] Fix Tergrid crash, Obeka, DFC name mismatch, compound type filter #engine
- [x] Giga Quorum v1-v5, TrueSkill, Content-addressable hashing (2026-04-30/05-01) #tournament #rating
- [x] YggdrasilHat political AI, Rivalry tracker, Killer-victim tracking (2026-04-30) #hat #analytics
- [x] Bayesian prior inheritance, Deck versioning DAG, Matchmaking scheduler (2026-04-30) #rating #matchmaking
- [x] Deck import from Moxfield, Bug/Suggestion report, Footer statusbar, About page, Donations page, User profile, Splash links #ui #platform
- [x] W/L color fix, Parser Wave 4, Engine resolve stubs 163/163, Per-card staples (17 + 4 improved) #engine #ui
- [x] Spell-copy tracking, Layer 7d P/T switching, Reflexive triggers, Damage distribution, Monarch system #engine
- [x] Annihilator/Afflict/Rampage/Bushido keywords, Elimination logging #engine
- [x] Custom brutalist sliders, Web leaderboard, ELO confidence badges, Deck drilldown curve/pie #ui
- [x] Wire Forge gauntlet backend, Scryfall art prefetcher, Bracket System v2, Bracket-aware matchmaking #engine #ui #infra
- [x] Universal zone-change system (MoveCard) — 0 regressions across 64K tests #engine
- [x] Trigger dispatch audit — 8 dead triggers found, 7 fixed #engine



%% kanban:settings
```
{"kanban-plugin":"board","list-collapse":[false,false,false,false,false]}
```
%%
