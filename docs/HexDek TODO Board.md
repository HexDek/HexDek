---

kanban-plugin: board

---

## High Priority — Parser (100% Coverage Push)

*Empty — 100% coverage achieved (0 UnknownEffect across 31,963 cards)*


## High Priority — Engine

- [ ] **Remaining 276 commander handlers** — coverage at 447/652 files (750 registered names covering all 652 pool commanders). Template generator (`cmd/gen-handlers/main.go`) handles simple patterns. Pool now at 1292 decks (threshold lowered 100→80, 2026-05-05). #engine #per_card


## High Priority — Platform

- [ ] **Mobile full pass** — leaderboard, spectate, operator, meta pages need individual mobile audit at 375px. Deck drilldown done, rest pending. #ui #mobile
- [ ] **Mobile deck drilldown: curse data last** — on mobile, curse/genetic section should render below all other panels (currently positioned mid-page). #ui #mobile
- [ ] **Global glossary disclosure system** — every stat/label/metric across all pages tap-to-expand inline explanation. One shared component + glossary data source. Replaces FAQ concept. #ui #ux #accessibility
- [ ] **Curse Proficiency sigil** — cymatic SVG replacing curse section on deck page (circle → flower of life evolution, color-identity tinted). Part of "Hats" → "Curses" rebrand. #ui #design
- [ ] **Action button context boxes** — brief TLDR above gauntlet/test variant/etc buttons for neurodivergent UX clarity. #ui #ux #accessibility
- [ ] **"Consider Cutting" rationale** — each cut recommendation needs: what was detected (stats/pattern), why it's recommended (synergy gap, mana curve, etc), the resulting effect, and suggested swaps. #ui #deck #freya
- [ ] **Value Engine rationale** — explain WHY each value engine was identified for this deck (what cards/interactions trigger it, how the engine functions). #ui #deck #freya
- [ ] **Win Condition rationale** — show detection logic for each win-con (which cards form the line, what conditions are needed, how the combo resolves). #ui #deck #freya
- [ ] **Deck clone** — non-owners can clone a deck to their own collection for editing. Clone creates a copy under the cloner's owner dir, then they can rename/edit freely. #ui #platform
- [ ] **Ephemeral spectator rooms** — "Spectate Live" button on deck drilldown spawns a dedicated game instance for that deck. Keyed by deck_id (second viewer joins existing room, no duplicates). Room runs games continuously while viewers are connected; after last viewer leaves, finishes current game then tears down. Per-room speed dial. Max 8 concurrent rooms. Main fishtank `/spectate` untouched. Games feed into ELO/Heimdall/cardstats (rated). Spectated deck always seat 1 (camera focus). Bracket-matched opponents. Max 50 concurrent rooms, unlimited viewers per room. Backend: `POST /api/spectate/spawn {deck_id}` → `room_id`, WebSocket at `/ws/spectate/{room_id}`. Frontend: same Spectator component, room-aware. #platform #spectate #backend #ui
- [ ] **Reconnection countdown** — when WebSocket disconnects, show attempt number and countdown timer per reconnection attempt (currently just shows "DISCONNECTED — RECONNECTING"). #ui #ux
- [ ] **Magic link graceful flow** — user clicks email link → new tab catches the auth post and logs them in → tab auto-closes → original tab plays a console-style "logging in" feed animation → redirects to /operator. #ui #auth #ux


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

- [ ] **Equipment stacking intelligence** — hat currently scores equip at a flat 20 with no awareness of existing attachments. Needs: (1) prefer stacking multiple equipment on the same high-value creature over spreading thin, (2) prioritize commander as equip target (survives recast, Voltron payoff), (3) evaluate equipment synergies (Resurrection Orb on commander = recursive value engine). Currently in `greedy.go:scoreEffect` + `ChooseTarget`. #hat #equipment #voltron
- [ ] **Equipment-specific target scoring** — `ChooseTarget` for equip abilities picks first legal creature. Should score targets by: power/toughness, commander status, evasion keywords (flying/trample/unblockable), existing equipment count (diminishing returns vs stacking value), and equipment-creature synergy (e.g. deathtouch creature + equipment that grants first strike). #hat #equipment #targeting
- [ ] **Equipment recurrence awareness** — hat should recognize "equip → creature dies → re-equip" loops as value patterns (Skullclamp, Sword cycle). Equipment that generates advantage on connect (Sword of Feast and Famine) should be prioritized on evasive creatures. #hat #equipment #strategy
- [ ] **Graveyard recursion awareness (non-reanimator)** — all graveyard intelligence (intentional yard dumping, "let it die we can bring it back" logic, reanimation spell scanning) is gated behind `ArchetypeReanimator`. Decks with recursion effects (Szarekh, Meren, Karador) that aren't classified as reanimator don't get strategic graveyard play. Needs: detect recursion density in decklist regardless of archetype classification, apply yard-valuation logic proportionally. #hat #graveyard #strategy
- [ ] **Zone-cast grant strategic valuation** — Yggdrasil tracks zone-cast grants (flashback, escape, unearth, etc.) but doesn't factor them into discard/surveil decisions for non-reanimator decks. If a creature has Unearth, the hat should value it LESS in hand (can recover from yard) and MORE as a discard target. Same for flashback instants/sorceries. Currently only ArchetypeReanimator gets surveil-to-yard logic. #hat #graveyard #zonecast
- [ ] **Recursion-aware sacrifice evaluation** — hat should recognize when sacrificing a permanent with a recursion path (Persist, Undying, Unearth, etc.) is net-positive. Currently treats all sacrifices as pure loss unless ArchetypeReanimator. Sacrifice + Persist = free death trigger + creature returns. #hat #sacrifice #recursion

## Medium Priority — Engine

- [ ] **Temporal Pincer** — anon UUID cookie → session tracking → on login stitch all anon device UUIDs to authenticated profile. No PII, all UUIDs. Powers P&R via GraphQL. #infra #platform
- [ ] **BUG: Esika/Prismatic Bridge 9.2% WR** — systemic B5 combo execution issue. Bridge should flip and cast free spells every upkeep but combo assembly/sequencing not firing correctly. #engine #bug #per_card
- [ ] **BUG: Multiple B5 decks at 13-14% WR** — combo execution ceiling across several B5 commanders. Likely related to Esika issue — combo sequencer not recognizing all win-line piece states. #engine #bug #combo
- [ ] **34K corpus audit — DONE (initial run)** — 31,963 cards tested, 181 unique failures (99.4% card coverage). Full report: `data/corpus-audit-full-report.md`. Remaining work is handler coverage for the 181 failing cards. #engine #qa
- [ ] **Corpus audit: draw handler gaps (2,032 failures)** — conditional draw triggers not firing in test harness. Top cards: Solemn Simulacrum, Kindred Discovery, Veil of Summer, Keldon Raider, etc. Mix of test scaffolding issues (conditions not met) and missing handlers. #engine #handlers #draw
- [ ] **Corpus audit: lifegain/lifeloss gaps (1,612 failures)** — gain_life (1,101) + lose_life (511). Cards like Bloodchief Ascension, Grave Venerations, Senu. Triggered effects needing condition setup or handler wiring. #engine #handlers #life
- [ ] **Corpus audit: damage gaps (1,095 failures)** — damage effects parsed but not executing. Need to distinguish test harness setup issues from real handler gaps. #engine #handlers #damage
- [ ] **Corpus audit: discard/mill/buff gaps (534 failures)** — discard (305), mill (148), buff (81). Lower volume, mixed causes. #engine #handlers #misc
- [x] **Thor test harness: conditional trigger setup** — 14 new scaffold kinds in conditional_setup.go (gained-life, cast-spell, ETB, drawn-card, attacked, sacrificed, combat-damage, landfall, discarded, enchanted, opponent-lost-life, life-threshold, upkeep). (2026-05-07) #engine #qa #thor
- [x] **Muninn-Thor mismatch audit** — crossref.go: loads Muninn data + Thor failures, builds TP/FN/FP confusion matrix, computes recall/precision, writes markdown report. (2026-05-07) #engine #qa

## High Priority — Thor 2.0 (7174n1c, 2026-05-06) — ALL DONE

- [x] **Action traces** — `CardTraceRecorder` fluent builder with 8 phases (setup → panic), `TraceCollector` writes `.trace` files per card, failures-only mode, snapshot diff summaries. 27 tests. (2026-05-07) #thor #diagnostics
- [x] **Opponent auto-detect** — oracle text pattern analysis → 11 requirement categories → idempotent board-state enrichment for targeting effects. 30 tests. (2026-05-07) #thor #scaffolding
- [x] **Conditional trigger scaffolding** — 14 new scaffold kinds (gained-life, cast-spell, creature-ETB, drawn-card, attacked, sacrificed, combat-damage, landfall, discarded, enchanted-creature, opponent-lost-life, life-above/below-threshold, upkeep-phase). All three layers (detect/apply/trace) updated. 49 new tests. (2026-05-07) #thor #scaffolding
- [x] **Oracle Errata Pipeline** — `cmd/hexdek-oracle-sync/`: streaming Scryfall bulk download, field-level diff, markdown report, `--live`/`--dry-run`/`--diff-only`/`--report` modes. 27 tests. (2026-05-07) #thor #pipeline #infra
- [x] **Muninn cross-reference** — loads Muninn data, builds confusion matrix against Thor failures (TP/FN/FP), computes recall/precision, writes markdown report. 10 tests. (2026-05-07) #thor #qa


## Medium Priority — Platform

- [ ] BOINC-style distributed compute (desktop client → contribute games → earn credits) #distributed
- [ ] **Deterministic seed capture (anti-cheat Phase 1)** — surface the existing Heimdall seed ring buffer + JSONL flush as a cryptographic per-game seed contract; sign at game-start, verify on replay; required before spot-check + cauterize phases. Builds on the seed capture work already wired into all 3 game paths #anticheat
- [ ] Deterministic replay anti-cheat (cryptographic seed, spot-check 2-5%, auto-cauterize bad actors) — *enhanced by seed capture from Phase 1* #anticheat
- [ ] Statistical anomaly detection (per-contributor distribution tracking, 3σ flagging) #anticheat
- [ ] Credit economy (contribute compute → earn credits → spend on own deck testing) #economy
- [ ] Stream/narrator layer (game state → visual renderer → Twitch/OBS output) #stream
- [ ] **Stream/narrator OBS overlay** — concrete OBS browser-source build of the narrator layer: transparent-background spectator viewport, Ive three-act narrative caption strip, lower-third commander tags, configurable seat order. Targets streamer use over the open spectator feed. #stream #ui


## Low Priority — Hat Research (Bronze tier) — MOSTLY DONE

*Ref: `docs/architecture-hat-evolution.md` Levels 6-7 + Skunkworks.*
*Level 6 (Neural Position Evaluator), Level 7 (Self-Play Loop), Skunkworks (Tesla/Feynman/Lovelace/Ive/Watts) all complete 2026-05-02.*



## Low Priority

- [ ] **i18n — IN PROGRESS** — scaffolding shipped (i18n.js + locales/ + useT() hook + URL/navigator detection across 8 locales, commit 059a9d1, 2026-05-04). Content translation remaining: ~500 UI keys × 8+ languages still need professional translation; Scryfall localized card names integration also pending. #platform
- [ ] Multi-format support beyond Commander (Modern, Legacy deck ratings) #engine
- [ ] **Tournament prize pools** — hat-vs-hat bracket tournaments with cash prizes (1st/2nd/3rd/4th splits). Starcraft model: deckbuilding is the skill, hat execution is the layer. Legally sound as skill competition (no entry-fee model safest, donations-funded). Needs: bracket system, payout logic, age verification (18+), tax reporting (>$600). #platform #economy #future


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
