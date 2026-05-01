---

kanban-plugin: board

---

## High Priority — Parser (100% Coverage Push)

- [ ] **P1: Reduce UnknownEffect** — remaining ~4,339 cards (13.6%). Long tail: parsed_tail (1,875), custom (1,039), ability_word (820) #parser
- [ ] **P2: spell_effect kind** (820 cards) — ability word labels whose trigger body failed re-parse #parser
- [ ] **P6: saga_chapter kind** (70 cards) — saga chapter text parsing + lore counter mechanics #parser #engine
- [ ] **P12: TurnFaceUp handler** — only missing effect type (1/43). Morph/disguise face-up, needs Phase 8 layer-1b framework #engine


## High Priority — Engine

- [ ] Duplicate prevention rules (`sba.go:894`) — parser doesn't surface "can't have more than N" state, needs phase-7 wiring #engine
- [ ] AI behavior policy (`internal/ai/autopilot.go`) — only advances phases, no card play or combat decisions. Non-functional stub. #engine #ai
- [ ] 20 missing Game Changer per-card handlers (Wave 7 in progress) — Humility, Teferi's Protection, Consecrated Sphinx, Narset, Braids, etc. #engine #per_card
- [ ] Layer 3 text-changing effects — not implemented per CONFIDENCE_MATRIX #engine #layers
- [ ] ~30 keyword abilities marked STUB — Annihilator, Afflict, Bushido, etc. per CONFIDENCE_MATRIX #engine #keywords
- [ ] Trigger doubling (`obeka_support.go:579`) — not implemented #engine
- [ ] Bracket-aware tournament grinder — switch from AssemblePod to AssembleBracketPod, config flag #engine #matchmaking


## High Priority — Platform

- [ ] Negative ELO shame badges — "MID" stamp at 0, escalating tiers for deep negative. Leaderboard bottom-10 wall of shame section #ui #fun
- [ ] Operator platform page/tab (operator profile, deck management, analytics dashboard) #ui #platform
- [ ] Friends system + player profiles — add friends, view each other's profiles/decks #ui #social
- [ ] Bracket-stratified leaderboard tabs — filter by B1-B5, separate rankings per bracket #ui
- [ ] Game Changer cards list on deck page — show which specific GC cards the deck runs #ui


## Medium Priority — Engine

- [ ] Dungeon tracking (`sba.go:958`) — SBA 704.5t not implemented, low priority unless Acererak enters meta #engine
- [ ] Battle/Siege mechanics (`sba.go:1054,1071`) — protector state not modeled #engine
- [ ] Speed mechanic (`sba.go:1138`) — future mechanic, implement when Speed cards land #engine
- [ ] Regenerate 3 failing test goldens (basking_rootwalla, shivan_dragon, thorn_lieutenant) — expected failures from parser progress, regen after conjunction-fix pass #test


## Medium Priority — Platform

- [ ] BOINC-style distributed compute (desktop client → contribute games → earn credits) #distributed
- [ ] Deterministic replay anti-cheat (cryptographic seed, spot-check 2-5%, auto-cauterize bad actors) #anticheat
- [ ] Statistical anomaly detection (per-contributor distribution tracking, 3σ flagging) #anticheat
- [ ] Credit economy (contribute compute → earn credits → spend on own deck testing) #economy
- [ ] Stream/narrator layer (game state → visual renderer → Twitch/OBS output) #stream


## Low Priority

- [ ] Concession diagnostics — track concession rate per commander, board state at scoop, turn of scoop. Muninn + Heimdall. #rating #analytics
- [ ] Multi-format support beyond Commander (Modern, Legacy deck ratings) #engine
- [ ] Mobile-friendly leaderboard #ui
- [ ] Donations page BOINC/ads buttons — "COMING SOON" placeholders (`Donations.jsx:109,119`) #ui
- [ ] Report analysis placeholder (`Report.jsx:332`) — feature not fully wired #ui
- [ ] Remove empty `internal/rules/` package or populate it #cleanup


## Done

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
