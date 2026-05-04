---

kanban-plugin: board

---

## High Priority — Parser (100% Coverage Push)

*Empty — 100% coverage achieved (0 UnknownEffect across 31,963 cards)*


## High Priority — Engine

- [ ] **Remaining 276 commander handlers** — coverage at 447/652 files (681 registered names). Most remaining are 1-2 deck count. Template generator (`cmd/gen-handlers/main.go`) handles simple patterns. #engine #per_card
- [ ] **BUG: Ajani Nacatl Pariah 74% WR** — handler is correct; high WR was because opponents couldn't deal with PW. PW threat scoring fix (below) should normalize this. Re-check after grinder reset. #engine #bug #hat


## High Priority — Platform

- [ ] **Amiibo display on deck page** — show per-deck DNA pool: generation count, best fitness, 7 personality params (radar chart), 20 DimStats weight corrections (heatmap), fitness sparkline over generations. Force graph or 3D brain visualization for evolved weight topology. (wiedeman/7174n1c 2026-05-04) #ui #amiibo #design
- [ ] Negative ELO shame badges — "MID" stamp at 0, escalating tiers for deep negative. Leaderboard bottom-10 wall of shame section #ui #fun
- [ ] Operator platform page/tab (operator profile, deck management, analytics dashboard) #ui #platform
- [ ] Friends system + player profiles — lightweight "pub" model: see each other's decks/ELO, no feed/notifications. Add via search or deck page #ui #social
- [x] Bracket-stratified leaderboard tabs — filter by B1-B5, separate rankings per bracket + band labels (2026-05-04) #ui
- [x] Game Changer cards list on deck page — GC card names persisted to strategy.json + "GAME CHANGERS" panel with art thumbnails on deck drilldown (2026-05-04) #ui

### Legality Flag (7174n1c — 2026-05-02)

- [x] **Persist legality in strategy JSON** — Freya already runs 5 checks (card count, color identity, singleton, banned list, commander legality) but doesn't write `LegalityReport` to `.strategy.json`. Add `legality` field to output (2026-05-04) #engine #freya
- [x] **Legality badge on deck cards** — DeckList card tiles: green ✓ / red ✗ next to bracket. Legality read from strategy.json via enrichDeckSummary (2026-05-04) #ui #legality
- [x] **Legality section on deck info panel** — DeckArchive sidebar: new KV row `LEGALITY` with LEGAL/ILLEGAL status. If illegal, expandable violation list panel (2026-05-04) #ui #legality
- [x] **Legality filter on deck browse** — ALL/✓LEGAL/✗ILLEGAL filter tags on DECKS page, reads `d.legal` from API (2026-05-04) #ui #legality
- [ ] **Fix Meglin phantom metadata** — 3 Meglin decks (`brudiclad`, `lich_jarads_rats`, `sisay_trix`) missing commander card + bracket. Re-import with COMMANDER: header or manual fix #data

### Freya → UI Wiring Pass (2026-05-02 audit)

*14 fields computed by Freya, written to strategy.json, but never displayed by the frontend.*

**Deck identity & strength:**
- [x] **Star cards display** — `star_cards`: deck's best cards (Lovelace Composer scoring). Show as highlighted cards on deck page, art thumbnails with ★ score (2026-05-04) #ui #wiring
- [x] **Cuttable cards display** — `cuttable_cards`: weakest cards / upgrade candidates. Show in a "Consider Cutting" section, compact thumbnails (2026-05-04) #ui #wiring
- [x] **Power percentile badge** — `power_percentile`: power ranking vs all analyzed decks. Show as "Top X%" in deck info KV row (2026-05-04) #ui #wiring
- [x] **Commander synergy score** — `commander_synergy`: 0-1 float, how well the 99 supports the commander. Show as percentage in deck info KV row (2026-05-04) #ui #wiring
- [x] **Commander themes** — `commander_themes`: keyword themes the commander cares about. Display as tags on deck page (2026-05-04) #ui #wiring

**Tactical intel:**
- [x] **Vulnerable-to warnings** — `vulnerable_to`: deck weaknesses (e.g. "graveyard hate", "board wipes"). Show in strategy section as caution tags (2026-05-04) #ui #wiring
- [x] **Meta matchups** — `meta_matchups`: archetype matchup grid with favored/neutral/unfavored ratings + reason tooltips. Reason field added to strategy JSON (2026-05-04) #ui #wiring
- [x] **Mana base grade** — `mana_base_grade`: letter grade for mana base quality. Show in deck info panel KV row (2026-05-04) #ui #wiring
- [x] **Keepable hand %** — `keepable_hand_pct`: estimated % of opening hands worth keeping. Show in deck info KV row (2026-05-04) #ui #wiring
- [x] **Interaction profile** — `interaction_avg_cmc` + `cheap_interaction`: KV rows for avg CMC + count at ≤2 CMC in deck info sidebar (2026-05-04) #ui #wiring

**Card-level data:**
- [ ] **Card roles grid** — `card_roles`: per-card role tag (ramp/draw/removal/combo/etc). Powers the Ive card grid grouped by Freya role #ui #wiring #design
- [x] **Finisher cards callout** — `finisher_cards`: the actual win-condition cards. "Win Conditions" panel with art thumbnails (2026-05-04) #ui #wiring
- [x] **Color demand heatmap** — already implemented: Color Balance panel shows production vs demand bars reading `color_demand` from strategy.json (2026-05-04) #ui #wiring

**Discovery data:**
- [x] **Emergent synergies display** — `emergent_synergies`: Huginn-discovered card interactions, tier badges, observation count. Panel with card pairs + effect patterns (2026-05-04) #ui #wiring

**Freya output gaps (compute exists, not written to strategy.json):**
- [x] **Persist legality report** — already tracked above in Legality Flag section (2026-05-04) #engine #freya
- [x] **Persist curve warnings** — `CurveWarnings` from FreyaReport, structural mana curve issues (2026-05-04) #engine #freya
- [x] **Persist color mismatch warnings** — `ColorMismatch` from FreyaReport, under/over-represented colors (2026-05-04) #engine #freya
- [x] **Persist combo notes** — `ComboNotes` from FreyaReport, partial combo piece warnings (2026-05-04) #engine #freya

### UX Overhaul (Ive/Jobs/Watts Quorum — 2026-05-02)

**Navigation restructure:**
- [ ] Reduce nav from 8 tabs to 5-6 — PUBLIC: DECKS, RANKINGS, SPECTATE; AUTH adds: MY DECKS + contextual IMPORT #ui #nav
- [ ] Rename "DASH" → "MY DECKS" — possessive language signals ownership, not admin tooling #ui #nav
- [ ] Fold PLAY, FORGE, REPORT into contextual access — not top-level nav. Surface where needed (e.g. FORGE inside deck drilldown, REPORT inside game end) #ui #nav
- [ ] Consolidate ABOUT into footer — not worth a nav slot #ui #nav

**Home / Splash page:**
- [ ] Embed fishtank on home/splash page — merge splash + spectate into one attention trap. Live game visible immediately on landing #ui #home
- [ ] "Upload My Deck" CTA prominent on home page and browse/deck pages — primary conversion action #ui #home

**Deck pages — "tangible object" design:**
- [ ] Commander color-identity page theming — CSS vars `--page-wash`, `--accent` derived from commander color identity (e.g. Grixis = deep blue-black gradient with red accent). Every deck page feels unique #ui #deck #design
- [ ] Full-bleed commander art on deck pages — hero image, not a thumbnail. The commander IS the page #ui #deck #design
- [ ] Card grid view as default — art thumbnails grouped by Freya role (ramp, draw, removal, combo pieces, etc). Text list as toggle for data people #ui #deck
- [ ] Deck personality blurb prominent — Freya's 2-3 sentence archetype/strategy description front and center, not buried in a tab #ui #deck
- [ ] Commander theming on deck pages — visual design system where each deck page feels like holding a real object, not browsing a database row #ui #deck #design

**Deck library:**
- [ ] Deck library as visual shelf — card objects (commander art + name + bracket badge), not table rows. Browse = flip through a collection #ui #deck #library

**Search:**
- [ ] Universal search bar — one search field: decks, commanders, cards, players. Contextual results. Always accessible #ui #search

**Auth flow:**
- [ ] One-tap auth (Google/Discord) — triggered contextually (first deck upload, first friend add), not gatekept at the door #ui #auth

**Sharing:**
- [ ] Share = link — no login needed to view a shared deck. Shareable URL on every deck page #ui #social


## High Priority — Learning Loop (Observability)

*Ref: `docs/architecture-learning-loop.md` + `docs/architecture-hat-evolution.md`*

### Phase 1: Heimdall Package + Seed Capture (1 week) — BLOCKER for all below

- [x] **Extract Heimdall to `internal/heimdall`** — Observer struct, GameSeed, Observation, DeadTrigger, CoTriggerPair, PivotEvent, HealthPulse types. HuginnSink/MuninnSink/TelemetrySink interfaces. Nil-safe (2026-05-02) #engine #heimdall
- [x] **Seed ring buffer + disk flush** — `RecordSeed()` ring buffer (1000 cap), flushes to `data/heimdall/seeds.jsonl`. `RecordObservation()` routes to Huginn/Muninn. `RecordCrash()` immediate. `Flush()` for graceful shutdown (2026-05-02) #engine #heimdall
- [x] **Wire Heimdall into grinder** — `runOneGameFast()` calls `RecordSeed()` after ELO update, outside mutex. Uses `ClassifyKillWithMaxTurns`. Hoisted deckKeys for crash recovery access (2026-05-02) #engine #heimdall
- [x] **Wire Heimdall into fishtank** — `runOneGame()` calls `RecordSeed()` after game end, before persist channel send (2026-05-02) #engine #heimdall
- [x] **Wire Heimdall into gauntlet** — `RunGauntlet()` calls `RecordSeed()` after ELO update. Extracted gameSeed as named variable (2026-05-02) #engine #heimdall
- [x] **Wire crash recovery into Heimdall** — `RecordCrash()` in all three panic recovery blocks (grinder, fishtank, gauntlet). `Shutdown()` method calls `Flush()` for graceful shutdown (2026-05-02) #engine #heimdall
- [x] **`ClassifyKill()` helper** — reads `Seat.LossReason` first, falls back to heuristic (poison counters, commander damage map, life, mill). `ClassifyKillWithMaxTurns` variant for timeout detection (2026-05-02) #engine #heimdall
- [x] **Parser gap flag in engine** — `parser_gap` event + Flags counter at 4 unhandled paths: UnknownEffect, default effect, conditional_effect, modification_effect in resolve.go/resolve_helpers.go. `emitPartial` in per_card/helpers.go also emits parser_gap. Feeds into Muninn via tournament runner (2026-05-02) #engine #heimdall

### Phase 2: Muninn + Huginn Wiring (3-4 days) — Depends: Phase 1

- [x] **Muninn: RecordParserGaps()** — `muninnAdapter` converts gaps to count map, delegates to `PersistParserGaps`. Wired as real Heimdall sink in NewShowmatch (2026-05-02) #engine #muninn
- [x] **Muninn: RecordDeadTriggers()** — adapter converts heimdall.DeadTrigger to muninn format, new `PersistDeadTriggersRaw()` for direct writes (2026-05-02) #engine #muninn
- [x] **Muninn: RecordCrash() wiring** — adapter delegates to `PersistCrashLogs` with stack trace + deck keys. Immediate write (2026-05-02) #engine #muninn
- [x] **Muninn: invariant_violations.json** — `InvariantViolation` struct (rule, message, game_seed, turn, timestamp), `PersistInvariantViolations()` + reader. Atomic append-write pattern (2026-05-02) #engine #muninn
- [x] **Muninn: regression_failures.json** — `RegressionFailure` struct (test_name, expected, got, game_seed, timestamp), `PersistRegressionFailures()` + reader. Same pattern (2026-05-02) #engine #muninn
- [x] **Huginn: IngestCoTriggers()** — `huginnAdapter` converts heimdall.CoTriggerPair to huginn.RawObservation, `PersistRawObservationsRaw()` for direct writes. Wired as real Heimdall sink in NewShowmatch (2026-05-02) #engine #huginn
- [x] **Batch replay system** — `ReadSeeds()` from JSONL, `ReplayWithObservation()` mirrors runOneGameFast setup (deterministic RNG, deck loading, TakeTurn loop), extracts ParserGaps + CoTriggers, routes to Observer. `BatchReplay()` with per-game panic recovery. `ReplayContext` with lazy deck cache (2026-05-02) #engine #heimdall

### Phase 3: Huginn → Freya Pipe (2-3 days) — Depends: Phase 2

- [x] **Huginn Tier 3 export** — `FreyaInteraction` type, `exportTier3ForFreya()` called at end of `Ingest()`, writes all Tier 3 patterns expanded by card pairs to `data/huginn/tier3_for_freya.json`. `ReadTier3ForFreya()` exported reader (2026-05-02) #engine #huginn
- [x] **Freya reads Tier 3 file** — `findEmergentSynergies()` in hexdek-freya reads `tier3_for_freya.json` first (highest confidence), then learned interactions. Deduped by sorted card pair keys (2026-05-02) #engine #freya
- [x] **Validate loop closure** — deployed to DARKSTAR, grinder running 33K g/min. 44K seeds captured in 2min, 1187/1196 amiibo pools created, evolution triggering (gen 1+). Heimdall→seeds.jsonl + Amiibo persistence confirmed live (2026-05-02) #engine #integration


## High Priority — Hat Intelligence

*Ref: `docs/architecture-hat-evolution.md` Levels 2-3*

### Level 2: Combo Sequencer (2-4 weeks) — Depends: Freya combo packages exist (DONE)

- [x] **ComboConstraint struct** — `PiecesNeeded`, `ZonesAccepted` (string zones matching engine convention), `ManaRequired`, `SequenceOrder`, `NeedsProtection`. `ComboAssessment` result type with Executable/Assembling/BestLine/NextAction/MissingPiece (2026-05-02) #engine #hat #combo
- [x] **Combo line scanner** — `ComboSequencer.Evaluate()` builds zone index (hand/battlefield/graveyard), checks each piece against accepted zones, selects best line by completion ratio. 13 tests (2026-05-02) #engine #hat #combo
- [x] **Mana feasibility check** — sums `ManaCostOf` for uncast pieces, checks against `AvailableManaEstimate`. Rejects lines where mana insufficient (2026-05-02) #engine #hat #combo
- [x] **Combo sequencer integration** — wired into ChooseCastFromHand (COMBO-CAST short-circuit), ChooseActivation (COMBO-ACTIVATE), ChooseTarget (COMBO-TUTOR for MissingPiece). All constructors init from StrategyProfile (2026-05-02) #engine #hat #combo
- [x] **Tutor-to-combo targeting** — Assembling state detects tutor in hand (oracle text "search your library" or AST `tutor` effect), `MissingPiece` field populated for tutor targeting (2026-05-02) #engine #hat #combo
- [x] **Combo execution loop** — Execute plan combo-casts bypass normal eval, SequenceOrder via NextAction, activation support for battlefield pieces (2026-05-02) #engine #hat #combo

### Level 2.5: Hat State Machine (2-3 weeks) — Depends: Combo Sequencer

- [x] **GamePlan enum + PlanState struct** — 6 states (Develop/Assemble/Execute/Disrupt/Pivot/Defend) with `String()`. PlanState tracks ComboReady, ComboTotal, ThreatLevel, TurnsSincePlan (2026-05-02) #engine #hat #statemachine
- [x] **Transition logic** — `PlanState.Evaluate()`: Execute on combo ready, Assemble on combo-1+tutor, Disrupt on threat>0.7, Pivot after 5 turns assembling, Defend/Pivot timeout to Develop after 3 turns (2026-05-02) #engine #hat #statemachine
- [x] **Plan-biased evaluation weights** — `PlanWeightMultipliers()` returns per-dimension multipliers: Execute=ComboProximity 2.5x, Assemble=CardAdvantage 1.6x, Disrupt=ThreatExposure 1.8x, Defend=LifeResource 1.8x, Pivot=BoardPresence 1.6x. Applied in `rescaleWeights` via `PlanMultiplier` field (2026-05-02) #engine #hat #statemachine
- [x] **Wire PlanState into YggdrasilHat** — evaluated on upkeep via ObserveEvent, combo assessment + threat level fed in, plan transitions logged, evaluator PlanMultiplier set per turn. Reset on game_start (2026-05-02) #engine #hat #statemachine

### Level 3: Genetic Amiibo (2-3 weeks) — Depends: Heimdall seed capture

- [x] **AmiiboDNA struct** — 5 evolvable floats (Aggression, ComboPat, ThreatParanoia, ResourceGreed, PoliticalMemory) + DeckKey, Generation, GamesPlayed, Fitness (2026-05-02) #engine #hat #amiibo
- [x] **AmiiboPool per deck** — population of 8, fitness-proportional roulette `SelectForGame()`, EMA fitness update in `RecordResult()`, `InitPool()` with random params + 0.25 baseline fitness. Stored `*rand.Rand` with `SetRNG()` for deserialization (2026-05-02) #engine #hat #amiibo
- [x] **Evolution step** — `evolve()`: sort by fitness, kill bottom 2, clone top 2, gaussian mutate (σ=0.05), clamp [0,1], reset game counter. Auto-triggers at 100 games per deck (2026-05-02) #engine #hat #amiibo
- [x] **NewYggdrasilHatWithDNA()** — DNA field on YggdrasilHat, constructor maps 5 params: Aggression→attack thresholds (±0.15), ComboPat→combo hold multiplier (±40%), ThreatParanoia→ThreatExposure weight (±40%), ResourceGreed→CardAdvantage/BoardPresence balance (±30%), PoliticalMemory→detente threshold + grudge decay (2-6 turns, 0.08-0.22 rate). DNA [0,1] centered at 0.5=neutral (2026-05-02) #engine #hat #amiibo
- [x] **Wire Amiibo into grinder** — all 3 game paths (grinder/fishtank/gauntlet): lazy pool creation via `getOrCreateAmiiboPool()`, `SelectForGame()` → DNA copy → `NewYggdrasilHatWithDNA()`, `RecordResult()` post-game. Dedicated `amiiboMu` mutex, no deadlock with `sm.mu` (2026-05-02) #engine #hat #amiibo
- [x] **Amiibo API endpoints** — `GET /api/decks/{owner}/{id}/amiibo` returns deck_key, game_count, population array (8 members with generation, games_played, fitness, 5 personality traits). Snapshot under amiiboMu, 404 if no pool (2026-05-02) #api #amiibo
- [x] **Amiibo persistence** — `SavePool`/`LoadPool`/`SaveAllPools`/`LoadAllPools` in `amiibo.go`. Pools loaded at startup, saved on shutdown + after each evolution step (goroutine, no hot-path blocking). `data/amiibo/{deck_key}.json` (2026-05-02) #engine #hat #amiibo
- [x] **Amiibo test suite** — 10 tests: init, selection, fitness update, evolution trigger, clamp, save/load single, save/load all, empty dir, nonexistent dir (2026-05-02) #engine #test #amiibo
- [x] **Heimdall test suite** — 11 tests: seed buffer + flush, auto-flush at capacity, observation routing (Huginn/Muninn sinks), crash recording, nil sinks, ClassifyKill (LossReason + heuristic fallback + nil gs + timeout), concurrent seed recording (2026-05-02) #engine #test #heimdall


## High Priority — Telemetry

### Phase 6: GA4 Health Pulse (2-3 days) — Depends: Heimdall package

- [x] **GA4 client** — `internal/telemetry` package. Env-gated (`HEXDEK_GA4_MEASUREMENT_ID` + `HEXDEK_GA4_API_SECRET`), nil-safe, fire-and-forget POST, 5s timeout (2026-05-02) #infra #telemetry
- [x] **Health pulse every 60s** — `telemetryAdapter` implements TelemetrySink, `Pulse()` on Observer forwards. `runHealthPulse()` goroutine reads sm.stats + Muninn data files for gap/crash/trigger counts, sends top 5 gap cards (2026-05-02) #infra #telemetry
- [x] **Client-side gtag.js** — added to all 4 HTML pages (index, leaderboard, drilldown, import). Custom events: device_register, deck_import (premade/paste), game_start. Measurement ID placeholder `G-HEXDEK` (2026-05-02) #ui #telemetry


## High Priority — Calibration

- [x] **Floor calibration decks** — 3 decks at `data/decks/calibration/`: Golos 5c (99 basics), Child of Alara 5c (99 basics), Isamaru mono-W (99 Plains). Ready for grinder baseline run (2026-05-02) #rating #calibration
- [ ] **Ceiling calibration** — after Amiibo ships: identify best B5 combo deck after 10K+ generations. This is the upper bound reference point. *Depends: Amiibo* #rating #calibration


## Medium Priority — Hat Advanced (Silver tier)

*Ref: `docs/architecture-hat-evolution.md` Levels 4-5. Next quarter.*

### Level 4: Staged Decision Architecture (1-2 months) — Depends: Combo Sequencer + State Machine

- [ ] **Mjolnir/Gungnir/Ragnarok routing** — formalize 3-tier decision dispatch. Mjolnir (budget-0 heuristic, 90%), Gungnir (SAT+eval+UCB1, 9%), Ragnarok (MCTS, 1%). Route based on decision complexity + confidence. #engine #hat #staged
- [x] **Watts confidence threshold dial** — bracket-aware: B1=0.3 (first good-enough), B3=0.6 (moderate), B5=0.9 (near-optimal only). Same code path, different sensitivity. #engine #hat #staged
- [x] **Shannon entropy tracking** — model opponent hands as probability distributions. Tutor=near-zero entropy. 3-card draw=high entropy. Held mana=interaction probability. Feed into threat assessment. #engine #hat #staged

### Level 5: Information-Set MCTS (1-2 months) — Depends: Staged Architecture

- [x] **IS-MCTS implementation** — `determinize()` shuffles opponent hands, `multiRolloutForCard()` runs 3 determinized rollouts per candidate, picks best-on-average. Wired into ChooseCastFromHand. (2026-05-02) #engine #hat #mcts
- [x] **MCTS trigger conditions** — fires at genuine uncertainty points via budget system. 90% of decisions use heuristic path, IS-MCTS only for high-impact cast decisions. (2026-05-02) #engine #hat #mcts


## Medium Priority — Engine

- [ ] **N-card combo line detection** — Huginn currently tracks pairwise co-triggers. Expand to 3-5 card lines (e.g. Dramatic Reversal + Isochron Scepter + mana rock). Either N-tuple tracking or chained pairwise inference. (7174n1c 2026-05-04) #engine #huginn #combo
- [ ] **Muninn persist batching** — coalesce per-game read-modify-write (parser_gaps.json, dead_triggers.json) into periodic batch flush. Currently re-reads + re-writes full file per game, expensive as files grow. #engine #performance
- [ ] **Temporal Pincer** — anon UUID cookie → session tracking → on login stitch all anon device UUIDs to authenticated profile. No PII, all UUIDs. Powers P&R via GraphQL. #infra #platform


## Medium Priority — Platform

- [ ] BOINC-style distributed compute (desktop client → contribute games → earn credits) #distributed
- [ ] Deterministic replay anti-cheat (cryptographic seed, spot-check 2-5%, auto-cauterize bad actors) — *enhanced by seed capture from Phase 1* #anticheat
- [ ] Statistical anomaly detection (per-contributor distribution tracking, 3σ flagging) #anticheat
- [ ] Credit economy (contribute compute → earn credits → spend on own deck testing) #economy
- [ ] Stream/narrator layer (game state → visual renderer → Twitch/OBS output) #stream


## Low Priority — Hat Research (Bronze tier)

*Ref: `docs/architecture-hat-evolution.md` Levels 6-7 + Skunkworks. No timeline — requires Silver tier data.*

### Level 6: Neural Position Evaluator — DONE (2026-05-02)

- [x] **Game state tensor encoding** — 92-dim StateVector (22 features × 4 seats + 4 global), normalized to [0,1], perspective-rotated. (2026-05-02) #research #neural
- [x] **Value network training pipeline** — PyTorch script: 92→256→128→64→1 MLP, ReduceLROnPlateau, early stopping, CUDA 4090 + CPU fallback, pivot-weighted MSE loss. (2026-05-02) #research #neural
- [x] **Neural eval integration** — 80/20 heuristic/neural blend via `evalPosition()`, auto-loads `data/training/model.json`, graceful nil degradation. (2026-05-02) #research #neural

### Level 7: Self-Play Loop — DONE (2026-05-02)

- [x] **Self-play training loop** — SelfPlayManager: 10K sample threshold → PyTorch training → hot-reload NeuralEvaluator into running hats. Atomic goroutine safety. 5-minute cooldown. (2026-05-02) #research #selfplay
- [ ] **Genetic→Neural distillation** — Amiibo explores parameter space cheaply, neural net distills into general model, model feeds better Amiibo starting points. *Deferred — needs more training data first* #research #selfplay

### Skunkworks Named Concepts — Mostly DONE

- [x] **Tesla Causal Graphs** — `ExtractPivot()` finds max relative swing turn. `LabelSamplesWithPivot()` enriches training data with pivot distance. `EvalSnapshotCollector` in all 3 game paths. (2026-05-02) #research #skunkworks
- [x] **Feynman Oracle** — 8 invariant checks (§704.5a/c/f/v, zone accounting, winner count, turn bounds, negative counters). Runs post-game in all 3 paths. hasCantLoseEffect() false-positive fix. (2026-05-02) #research #skunkworks
- [x] **Lovelace Composer Intent** — star card +0.20 boost, commander theme keyword matching +0.12 (oracle text + type line). Deck identity → signature card weighting → thematic play priority (2026-05-02) #research #skunkworks
- [x] **Ive Three-Act Spectator** — narrative arc generation (setup/conflict/resolution from Tesla causal pivots). ComposeNarrative() broadcasts to spectators on showmatch end. (2026-05-02) #research #skunkworks #ui
- [x] **Watts Soul Layer** — bracket-aware confidence threshold. B1=warm/casual, B5=cold/optimal. Implemented via applyBracketDial() + selectAmongTop(). #research #skunkworks


## Low Priority

- [ ] **i18n** — internationalize hexdek.dev for global audience. Scryfall has localized card names for 11 print languages. 500 UI keys, 50 languages, <$200 translation cost. #platform
- [ ] Multi-format support beyond Commander (Modern, Legacy deck ratings) #engine
- [ ] Mobile-friendly leaderboard #ui
- [ ] Donations page BOINC/ads buttons — "COMING SOON" placeholders (`Donations.jsx:109,119`) #ui
- [ ] Report analysis placeholder (`Report.jsx:332`) — feature not fully wired #ui


## Done

- [x] **Holy documentation pass** — 8 new architecture docs (Genetic Amiibo, Hat State Machine, Neural Evaluator, Self-Play Loop, Shannon Entropy, Tesla Causal Pivots, Feynman Oracle, Ive Spectator) + Learning Loop pipeline doc. Fixed 3 stale docs (ARCHITECTURE.md, YggdrasilHat.md, README.md). Updated TODO board with all skunkworks completions. 47→47 current docs. (2026-05-04) #docs
- [x] **Grinder memory leak fix** — Heimdall `obsBuf` unbounded growth. Observations were dispatched to Huginn/Muninn sinks then redundantly appended to a buffer that was never read. Removed dead buffer. (2026-05-04) #engine #performance
- [x] **Depression concession removal** — Score-based conviction concession was too aggressive (hat scooped at 38 life). Removed entirely; everyone fights to the death. Engine's SBA cap + loop detector handles actual infinite loops. (2026-05-04) #engine #hat
- [x] **Feynman Oracle false positive fixes** — 704.5a: hasCantLoseEffect() for Platinum Angel. Zone accounting: §800.4a cards_left_game tracking. Zombie Army token "token" type fix for IsToken(). (2026-05-04) #engine #hat
- [x] **Bracket filter leaderboard** — B1-B5 filter tabs + band labels on leaderboard, live via WebSocket. Desktop table + mobile cards. (2026-05-04) #ui
- [x] **Jhoira of the Ghitu suspend** — proper `OnActivated` handler, picks highest-CMC nonland from hand, calls `SuspendCard(gs, seat, card, 4)`. Removed stub (2026-05-01) #engine
- [x] **Lich's Mastery life observers** — `OnTrigger("life_gained")` draws cards, `OnTrigger("life_lost")` exiles permanents/hand/graveyard. Added `FireCardTrigger("life_lost")` to `resolveLoseLife` (2026-05-01) #engine
- [x] **Ulrich transform events** — `FireCardTrigger("transform")` added to `TransformPermanent()`. Back-face fight trigger on transform to Uncontested Alpha, front-face +4/+4 on transform back (2026-05-01) #engine
- [x] **Wayward Servant ETB observer** — `OnTrigger("token_created")`+`OnTrigger("permanent_etb")`: Zombie ETB drains opponents 1, gains controller 1 (2026-05-01) #engine
- [x] **Coat of Arms layer 7** — `RegisterContinuousEffect` layer 7c: each creature +N/+N where N = other creatures sharing a creature type, all battlefields (2026-05-01) #engine
- [x] **Concession diagnostics** — `ConcessionRecord` in Muninn: commander, turn, board power, life, hand size, opponents alive. `PersistConcessions()`, `SortedConcessions()`, `hexdek-muninn --concessions` flag (2026-05-01) #rating #analytics
- [x] **Dungeon tracking (704.5t)** — enhanced SBA infers completion from `dungeon_level` vs max rooms (7/4/4 for standard dungeons), cleans up flags (2026-05-01) #engine
- [x] **Battle/Siege protectors (704.5w/x)** — full SBA: auto-assigns first living opponent as protector, reassigns on protector death, sacrifices if no opponents. Siege controller=protector reset (2026-05-01) #engine
- [x] **Speed mechanic (704.5z)** — SBA checks permanent types + `start_your_engines` flag, sets `speed=1` on seats without it (2026-05-01) #engine
- [x] **Layer 3 text-changing effects** — STUB: LayerText=3 slot exists in pipeline (layers.go:429), no effects register. Intentional no-op per CONFIDENCE_MATRIX — no portfolio deck uses Magical Hack / Trait Doctoring. Deferred until meta demand. (2026-05-01) #engine #layers
- [x] **opponentLikelyHasWrath expansion** — replaced boolean with `wrathProbability()` returning graded float64. Factors: hand size (0.04/card), mana availability, opponent colors (W+0.15, B+0.10, R+0.05), archetype (Control+0.15), cast cadence (nothing cast + full hand + mana = +0.10), prior wrath history from 3rd Eye cardsSeen (+0.20). Cap 0.95 (2026-05-01) #engine #evaluator
- [x] **Partner-aware mulligan adjustment** — detects partner pair (CommanderNames >= 2), collects hand colors from lands, counts enablers per commander's colors. Mulligans if one commander has 0 enablers and hand lacks star cards / combo pieces (2026-05-01) #engine
- [x] **Transform recognition in cardHeuristic** — IsMDFC() flexibility bonus +0.10, back-face-is-land +0.10, back-face cheaper and castable +0.08. Scores both faces of DFC cards (2026-05-01) #engine
- [x] **Tournament runner post-game stat emission** — SeatStats struct (15 fields: life, board, hand, graveyard, mana sources, spells, creatures, conceded, etc). Emitted per-game in PostGameStats without event log. `CountManaRocksAndLands` exported for cross-package use (2026-05-01) #engine #analytics
- [x] **PartnerSynergy evaluator dimension** — partner pair on-field bonus, 4-color coverage scoring, complementary role detection (draw/attack/tutor/removal), tax penalty for repeated deaths. 13 archetype weight profiles. Tests (2026-05-01) #engine #evaluator
- [x] **ActivationTempo evaluator dimension** — non-mana activated ability scoring, untapped vs tapped weighting, repeatable (no-tap) engines boosted, high-impact activation bonus, opponent-relative comparison. Tests (2026-05-01) #engine #evaluator
- [x] **ToolboxBreadth evaluator dimension** — tutors in hand, modal spells, MDFC flexibility, non-mana board activations, tutor-target-aware bonus from Freya profile. Tests (2026-05-01) #engine #evaluator
- [x] **Threat trajectory prediction** — forward-looking opponent power projection: board power + deployment potential (hand × mana) + spell cadence bonus. Clamps to [-2, 0]. Tests (2026-05-01) #engine #evaluator
- [x] **UCB1 exploration factor per archetype/turn** — `refreshExplorationFactor()` replaces hardcoded √2. Aggro/Tribal=1.0, Combo=1.8, Control=1.6, Stax=1.2. Early game +0.3, late game -0.3 decay. Floor 0.5. Cached per turn (2026-05-01) #engine #evaluator
- [x] **Dynamic evaluator weight rescaling** — `rescaleWeights()` adjusts all 16 dimension weights by game stage (early boosts mana/card, late boosts combo/threat/board) and position (behind boosts combo/toolbox, ahead boosts card/mana/life). Tests (2026-05-01) #engine #evaluator
- [x] **Light mode toggle** — `[data-theme="light"]` CSS vars, toggle button on all pages (drilldown, leaderboard, import, game), localStorage persistence (2026-05-01) #ui
- [x] **Curve analysis UI** — mana curve bar chart + color balance demand/supply visualization in Freya Analysis tab. ManaCurve + ColorBalance added to strategy.json (2026-05-01) #ui #analytics
- [x] **Deck page auto-refresh on Freya push** — SSE endpoint `/api/decks/{owner}/{id}/events` broadcasts `freya_complete` event when analysis finishes. Drilldown page auto-reloads data via EventSource (2026-05-01) #ui #infra
- [x] **Exile-cast pattern** — Dauthi Voidwalker (replacement effect via graveyard_enter trigger), Emry (ZoneCastGrant from graveyard + ExileOnResolve), Urza (free cast from exile + EOT cleanup) (2026-05-01) #engine
- [x] **Clone/copy handlers** — already implemented: Phantasmal Image (DeepCopy + sac-on-target flag), Riku (creature token copy + spell stack copy). Stale TODO removed (2026-05-01) #engine
- [x] **ELO-bracket correlation re-check** — 485K games, rho=-0.49 (inverted). B1 avg 1649, B5 avg 1315 in standard ELO. HexELO preserves bracket ordering via seeded starts. Drift outliers: 921/1196 decks (2026-05-01) #engine #rating
- [x] **Soulshift keyword** — CR §702.46 implemented: CheckSoulshift fires on creature death, returns highest-CMC Spirit with mana value ≤ N from graveyard to hand. CONFIDENCE_MATRIX updated to SOLID. 0 STUB keywords remain (2026-05-01) #engine #keywords
- [x] **Opponent graveyard threat tracking** — 12th evaluator dimension `scoreOpponentGraveyard`: reanimation targets, flashback/escape spells, enablers on battlefield, known GY-abuse commanders. Weighted per archetype (2026-05-01) #engine #evaluator
- [x] **Color-aware mana sequencing** — ChooseLandToPlay now scans hand spells for color pip demand, boosts lands producing needed colors with deficit-aware weighting. Near-castable spells (CMC ≤ available+1) get 2x priority (2026-05-01) #engine #evaluator
- [x] **Jeska's Will** — exile-play permission via RegisterZoneCastGrant for each exiled card + end-of-turn DelayedTrigger cleanup. Last GC clause gap closed (2026-05-01) #engine
- [x] **Panoptic Mirror** — imprint persistence via sync.Map (Isochron Scepter pattern), upkeep creates StackItem{IsCopy: true} + InvokeResolveHook. Last GC clause gap closed (2026-05-01) #engine
- [x] **P1: 100% parser coverage** — 0 UnknownEffect across 31,963 cards. `final_13_cards.py` extension covers forced-attack, variable P/T, as-enters, monstrous triggers, composite ETB, search-library, card-put-into-zone (2026-05-01) #parser
- [x] **Graveyard-leave observer hook** — `graveyard_leave` event fires from MoveCard when fromZone=="graveyard". Tormod creates 2/2 Zombie, Imotekh creates 2x 2/2 Necron Warriors on artifact leave (2026-05-01) #engine
- [x] **Land-tap hook** — `land_tapped_for_mana` event fires from AddManaFromPermanent when source is land. Caged Sun doubles controller's mana, Gauntlet of Power doubles all players' mana (2026-05-01) #engine
- [x] **Cast observer → Door of Destinies** — `spell_cast` already fires; wired Door of Destinies charge counter increment on controller cast (2026-05-01) #engine
- [x] **Hat: Poison/infect threat awareness** — `poisonReceivedFrom` 3rd Eye tracking, `poisonPenalty` in ThreatExposure evaluator (9 counters=0.8, 7=0.5, 5=0.2), infect/toxic creature detection boosts threat score, assessAllThreats poison proximity (2026-05-02) #engine #hat #evaluator
- [x] **Hat: Planeswalker threat scoring** — `estimatePWUltimateCost()` parses oracle text for ultimate, `PWLoyaltyThreat` in seatThreat struct, loyalty-scaled removal scoring (can-ult=+5.0, 1-2 away=+3.5, base=+2.0), assessAllThreats PW proximity (2026-05-02) #engine #hat #evaluator
- [x] **Hat: Alternate wincon awareness** — poison/mill/commander damage penalties in ThreatExposure. Mill: library<10=0.8 penalty when opponent has mill permanents. Commander: 18+ damage=0.6, 15+=0.3. Feeds state machine Disrupt transitions (2026-05-02) #engine #hat #evaluator
- [x] **BUG: Rosnakht 2.5% WR** — battle cry keyword flag in ETB handler (ensures ApplyBattleCry recognition), heroic trigger via OnTrigger("spell_cast") creating 0/1 Kobold tokens (2026-05-02) #engine #bug
- [x] **BUG: Tazri 13% WR** — cost reduction via CountParty() in ScanCostModifiers (0-4 reduction), activated ability: top 6, reveal up to 2 Cleric/Rogue/Warrior/Wizard/Ally, hand, rest to bottom (2026-05-02) #engine #bug
- [x] **Bracket-aware tournament grinder** — switched AssemblePod → AssembleBracketPod in showmatch, populates Bracket field from ELO cache (2026-05-01) #engine #matchmaking
- [x] **Keyword stubs audit** — CONFIDENCE_MATRIX reconciled 2026-04-30, all 30+ keywords now SOLID. Only Soulshift remains (rare). Stale TODO cleaned (2026-05-01) #engine #keywords
- [x] **Cleanup: internal/rules/** — removed empty package, test goldens item stale (no files exist) (2026-05-01) #cleanup
- [x] **P12: TurnFaceUp effect handler** — 44th effect type wired in resolve.go, calls existing dfc.go:TurnFaceUp, megamorph +1/+1 counter support, 3 tests (2026-05-01) #engine
- [x] **P6: Saga ETB + chapter tagging** — saga_chapter detection moved before extension loop (all chapters tagged correctly), initSagaLoreCounters on ETB sets saga_final_chapter + first lore counter, 2 tests (2026-05-01) #parser #engine
- [x] **P2: Ability word extension** — already wired via load_extensions() STATIC_PATTERNS → EXT_STATIC_PATTERNS. Stale TODO: 820 cards resolved, coverage at 99.96% (2026-05-01) #parser
- [x] **SBA 704.5r counter limit enforcement** — parses "can't have more than N counters" from AST raw text, trims excess, 3 tests (2026-05-01) #engine
- [x] **AI autopilot behavior policy** — plays lands, taps mana, casts highest-CMC affordable spells, alpha-strikes in combat (2026-05-01) #engine #ai
- [x] **ELO reset #2** — cleared 1196 entries + 367 games + 4585 card stats, grinder resampling with 447 handlers active (2026-05-01) #rating
- [x] **Handler coverage push** — 66→447 handler files across waves 1-14, 6 dev hexes. Template generator `cmd/gen-handlers/main.go` (201 auto-gen + 228 manual) (2026-05-01) #engine #per_card
- [x] **53/53 GC per-card handlers** — all Game Changers have registered handlers (2026-04-30) #engine #per_card
- [x] **Spectator UI scroll fix** — overflow-anchor + CSS containment + rAF log scroll (2026-05-01) #ui
- [x] **Turn bar redesign** — left-anchor commander, right-anchor perms, kill blinker, ellipsis overflow (2026-05-01) #ui
- [x] **Stax lock wiring** — Null Rod/Ouphe/Totem via StaxCheck, Grand Abolisher cast block (2026-04-30) #engine
- [x] **CreateToken hook** — token_created trigger fires, Chatterfang + Anointed Procession + Pitiless Plunderer (2026-04-30) #engine
- [x] **Dual-track ELO** — standard + HexELO (bracket-weighted) computed every game (2026-04-30) #rating
- [x] **HexELO drift detection** — /api/live/elo/drift endpoint, outlier tagging (2026-04-30) #rating
- [x] **Loss reason display** — spectator UI shows GG reason (cmdr dmg, life, poison, etc) (2026-05-01) #ui
- [x] **Partner commander casting priority** — +0.20 base, +0.45 when partner on board (2026-05-01) #engine
- [x] **Sacrifice-as-value overhaul** — drain/draw/ramp payoffs, 1.5x fodder multiplier (2026-05-01) #engine
- [x] **Sandbagging exemption** — aristocrats/combo/enchantress/artifacts at 30% penalty (2026-05-01) #engine
- [x] **Reanimate activation awareness** — activationHeuristic graveyard-to-battlefield scoring (2026-05-01) #engine
- [x] Fix Tergrid recursive trigger crash (depth guard + total trigger cap) #engine
- [x] Fix Obeka wrong ability resolution #engine
- [x] Fix DFC commander name mismatch #engine
- [x] Fix compound type filter for cast triggers #engine
- [x] Giga Quorum v1 (30,597 games, 18 decks, 7m11s) #tournament
- [x] Giga Quorum v2 (30,595 games, 18 decks, 12m, trigger-capped) #tournament
- [x] Universal zone-change system (MoveCard) — 0 regressions across 64K tests #engine
- [x] Trigger dispatch audit — 8 dead triggers found, 7 fixed #engine
- [x] ELO rating system (standard K=32, multiplayer pairwise) #rating
- [x] Tune trigger cap — tested 2000/5000, converged at depth 15 (2000 is optimal) #engine
- [x] Giga Quorum v3 (30,595 games, 5000 cap, identical to v2 — confirms convergence) #tournament
- [x] Crash investigation: all 5 "crashes" are 90s wall-clock timeouts, not panics. Eshki in 5/5 pods. #engine
- [x] TrueSkill rating system — multiplayer Bayesian (μ, σ), pairwise decomposition, wired into tournament + round-robin #rating
- [x] Content-addressable deck hashing (SHA256 sorted card list) #rating
- [x] Giga Quorum v4 (30,595 games, TrueSkill enabled) — Yuriko #1, Coram #2, Oloro drops to #5 #tournament
- [x] YggdrasilHat political AI — unified hat with 8-dim evaluator, threat scoring, grudge tracking, budget system #hat
- [x] Giga Quorum v5 (30,598 games, Yggdrasil budget=50) — Shalai #1 win, Varina #1 TS, Soraya #2→#16, politics reshuffles meta #tournament
- [x] Rivalry tracker (per-deck matchup W/L, canonical key pairing, cross-run merging) #analytics
- [x] Killer-victim elimination tracking (threat graph, backward event log scan, kingmaker detection) #analytics
- [x] Bayesian prior inheritance for deck versioning (μ carries, σ inflates by card delta) #rating
- [x] Deck versioning DAG (content-addressable SHA256, lineage, HEAD leaderboard) #schema
- [x] Matchmaking scheduler (rating-aware pod assembly, info gain scoring) #matchmaking
- [x] Deck import from Moxfield URL (auto-register, auto-hash, auto-rate) #ui
- [x] Bug/Suggestion report (red footer button → form → JSON flat-file persistence) #ui #platform
- [x] Footer statusbar (bug report link, donate link, user status) #ui
- [x] About page (project overview, philosophy, tech stack, no-ads stance) #ui #platform
- [x] Donations page (monthly COGs breakdown, donation tracker bar, philosophy) #ui #platform
- [x] User profile page (display name, owner name for deck filtering) #ui #platform
- [x] Splash page GitHub + docs links #ui
- [x] W/L color fix (wins green, losses red across DeckList, DeckArchive, gauntlet) #ui
- [x] Parser Wave 4: 73% → 86.42% (+4,193 cards, 125 new rules) #parser
- [x] Engine resolve stubs: 163/163 mod handlers promoted (0 remaining) #engine
- [x] Per-card handlers: 17 commander staples (Sol Ring, Force of Will, Smothering Tithe, etc.) #engine
- [x] Per-card handlers: Necrotic Ooze, Bolas's Citadel, Food Chain, Underworld Breach improved #engine
- [x] Spell-copy tracking — `Card.IsCopy` bool + SBA 704.5e copy cleanup #engine #copy
- [x] Layer 7d P/T switching — RegisterPTSwitch, RegisterDoranSiegeTower #engine #layers
- [x] Reflexive triggers — QueueReflexiveTrigger via DelayedTrigger system #engine #triggers
- [x] Damage distribution (601.2d) — distributeDamage + DamageDistributor interface #engine
- [x] Monarch system — BecomeMonarch wired via court cards #engine
- [x] Annihilator/Afflict/Rampage/Bushido keywords — AST-aware N extraction #engine #keywords
- [x] Elimination logging — `>>>` death entries with loss reason (life/poison/cmdr/mill) #engine #spectator
- [x] Custom brutalist sliders — 3px track, 12x12 square thumb, var(--ok) fill #ui #design
- [x] Web leaderboard — sortable table, search, mobile sort bar, clickable rows, confidence dots #ui
- [x] ELO confidence badges on deck list + drilldown #ui #rating
- [x] Deck drilldown mana curve/color pie (client-side fallback when no Freya data) #ui
- [x] Wire Forge gauntlet backend — game count selector, RUN button, progress bar, results #ui
- [x] Scryfall card art prefetcher (`cmd/hexdek-artfetch/`) + RAM cache warming at startup #infra
- [x] Bracket System v2 — WotC-aligned 5-tier (Exhibition/Core/Upgraded/Optimized/cEDH) + 53 Game Changers scoring #engine #rating
- [x] Bracket-aware matchmaking — AssembleBracketPod with soft ±1 weighting #matchmaking



%% kanban:settings
```
{"kanban-plugin":"board","list-collapse":[false,false,false,false,false]}
```
%%
