---

kanban-plugin: board

---

## High Priority — Engine


## High Priority — Platform

- [ ] Web leaderboard page (show deck ratings, confidence intervals, matchup data) #ui
- [ ] Deck drilldown UI: Freya curve/ratio analysis + Heimdall game analytics (ELO, card performance, matchups) — APIs already wired #ui
- [ ] Operator platform page/tab (operator profile, deck management, analytics dashboard — non-engine, UI + platform) #ui #platform
- [ ] BOINC-style distributed compute (desktop client → contribute games → earn credits) #distributed
- [ ] Deterministic replay anti-cheat (cryptographic seed, spot-check 2-5%, auto-cauterize bad actors) #anticheat
- [ ] Statistical anomaly detection (per-contributor distribution tracking, 3σ flagging) #anticheat
- [ ] Credit economy (contribute compute → earn credits → spend on own deck testing) #economy
- [ ] Stream/narrator layer (game state → visual renderer → Twitch/OBS output) #stream


## Low Priority

- [ ] Concession diagnostics first: track concession rate per commander, board state at scoop, turn of scoop — collect data before designing discount system. Falls under Muninn + Heimdall. #rating #analytics
- [ ] Multi-format support beyond Commander (future: Modern, Legacy deck ratings) #engine
- [ ] Mobile-friendly leaderboard #ui


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



%% kanban:settings
```
{"kanban-plugin":"board","list-collapse":[false,false,false,false,false]}
```
%%
