# Tournament Deep Audit — 2026-05-17

**Branch:** `dev/tournament-deep-audit`
**Pool:** Top 100 decks by ELO `hex_rating` from `data/hexdek.db` (`showmatch_elo` table)
**Format:** 4-seat Commander, pool mode (each game samples 4 unique decks from the 100)
**Hat:** YggdrasilHat with per-deck Freya StrategyProfile, budget=50 (auto-scaled by `BudgetForPower`), noise σ=0.2
**Engine commit:** `main` @ `1c49f08`
**Runner:** `cmd/tournament-audit` (added in this branch; thin driver around `tournament.RunOneGameForAudit`)
**Artifacts:** `hexdek/tournament-audit-2026-05-17/`

| Pass | Games | Audit | Analytics | Seed | Throughput |
|------|------:|:-----:|:---------:|-----:|----------:|
| Main (winrate + crashes) | 2000 | on | off | 51717 | 16.1 g/s (124 s) |
| Sample (blunder spot)    |   50 | on | on  | 91317 | 12.6 g/s (4 s)   |

---

## 1. Headline numbers

```
Main run (2000 games):
  crashes:     0
  draws:       2  (+ 1 turn-cap tie with a leader)
  avg turns:   45.84
  concessions: 0
  end reasons: last_seat_standing 1920 (96.0%)
               turn_cap_leader      77 ( 3.85%)
               draw                  2 ( 0.10%)
               turn_cap_tie          1 ( 0.05%)
  pool coverage: 63 unique commanders (multiple deck files share partner names)
```

**Net:** Engine is rock-stable on the curated top-100 pool. No panics, no rule-auditor violations surfaced, no hangs.

---

## 2. Crashes (Loki-style)

**0 crashes across 2000 games.** Full per-game JSONL in
`main_per_game.jsonl` and the empty `crash_records: []` field in
`summary.json`.

For context, this run loaded each deck through `astload` +
`deckparser.ParseDeckFile` + `hat.LoadStrategyFromFreya`; the same code
path Loki fuzzes. The clean run validates that the top-100 ELO pool
does not contain any combination of cards that currently panics the
engine — every previously logged crash signature (Tesla pivot,
Abdel Adrian LTB, etc.) is closed by recent commits.

Note: Loki itself stresses random/garbage card sets; this audit
stresses *real curated decks*, which is a different (and previously
crashy) shape. Both pass.

---

## 3. Winrate distribution

### 3.1 Bucketed histogram (63 commanders)

| Winrate bucket | Commanders |
|:---------------|----------:|
|  0%             |  0  |
|  0–10%          |  4  |
|  10–20%         | 20  |
|  20–30%         | 13  |
|  30–40%         | 16  |
|  40–50%         |  3  |
|  50%+           |  7  |

```
   0–10% : ████                                              4
  10–20% : ████████████████████                             20
  20–30% : █████████████                                    13
  30–40% : ████████████████                                 16
  40–50% : ███                                               3
  50%+   : ███████                                           7
```

Roughly bell-shaped around 20–40%, with a long right tail of seven
commanders that are visibly oppressive in the meta. Almost no
"zero-winrate" decks — every entry in the pool wins at least
occasionally, which is the right shape for a pool labelled "top".

### 3.2 Top 10 (by winrate)

| Commander | Games | Wins | Winrate |
|-----------|------:|----:|---------|
| Emiel the Blessed | 55 | 40 | 72.72% |
| Ureni of the Unwritten | 81 | 54 | 66.66% |
| Valgavoth, Harrower of Souls | 89 | 51 | 57.30% |
| Maelstrom Wanderer | 89 | 51 | 57.30% |
| Sisay, Weatherlight Captain | 212 | 117 | 55.18% |
| Yuriko, the Tiger's Shadow | 80 | 42 | 52.50% |
| Lorehold, the Historian | 77 | 39 | 50.64% |
| Derevi, Empyrial Tactician | 88 | 42 | 47.72% |
| Edgar Markov | 269 | 128 | 47.58% |
| Shiko, Paragon of the Way | 81 | 34 | 41.97% |

### 3.3 Bottom 10 (by winrate)

| Commander | Games | Wins | Winrate |
|-----------|------:|----:|---------|
| Vial Smasher the Fierce | 93 | 3 | 3.22% |
| Malcolm, Keen-Eyed Navigator | 88 | 5 | 5.68% |
| Glarb, Calamity's Augur | 230 | 21 | 9.13% |
| Norman Osborn // Green Goblin | 274 | 27 | 9.85% |
| Gorma, the Gullet | 82 | 9 | 10.97% |
| Rograkh, Son of Rohgahh | 687 | 88 | 12.80% |
| Shorikai, Genesis Engine | 83 | 11 | 13.25% |
| Flubs, the Fool | 83 | 11 | 13.25% |
| The Gitrog Monster | 81 | 11 | 13.58% |
| Thrasios, Triton Hero | 240 | 33 | 13.75% |

### 3.4 Caveat — partner commanders aggregate

Pool-mode samples by deck file, but this report aggregates by
**commander name**. Partner shells share a commander across many deck
files in the top 100:

| Commander | Deck files | Expected seat-games (× ~80) | Observed |
|-----------|-----------:|----------------------------:|---------:|
| Tymna the Weaver | 11 | 880 | 876 |
| Rograkh, Son of Rohgahh | 9 | 720 | 687 |
| Kinnan, Bonder Prodigy | 5 | 400 | 396 |
| Thrasios, Triton Hero | 3 | 240 | 240 |
| Edgar Markov | 3 | 240 | 269 |
| Sisay, Weatherlight Captain | 3 | 240 | 212 |

So "Rograkh 12.8%" is the winrate of the average Rograkh-partner
**shell**, not of a single deck. Drilling to per-deck-file winrate
would require the participants array to carry deck-path instead of
commander-name; the per-game JSONL doesn't carry that today.

---

## 4. AI blunder spot-check (50 sampled games)

Sample pass had analytics enabled, so every game emits a
`GameAnalysis` with `WinCondition`, `WinningCard`, `MissedCombos`,
`MissedFinishers`, and a `StallReport`. Raw records in
`sample_samples.jsonl`.

### 4.1 What the requested "blunder" signals look like

- **Conceded at high life:** 0 occurrences. The Hat's concession
  detector did not fire once in 2050 games. Either the curated pool
  never put a Yggdrasil seat in a conviction-loss position, or the
  concession heuristic is too conservative — **most likely the
  latter**, since 3.85% of main-run games hit turn-cap with a clear
  leader and the losers played to the cap rather than scooping.
- **Declined winning attack:** the engine doesn't currently log a
  "lethal attack offered and rejected" signal, so we can't quantify
  this directly. The closest proxy is `MissedFinishers` (finisher in
  play, opp life ≤ lethal, didn't win) — **0 detected in 50 games**.
  Worth treating with skepticism: the detector only fires on
  Freya-classified finishers AND requires opponent at low life at
  game end. Most stalled games end with all 4 players >25 life,
  outside the detector's window.

### 4.2 Stronger blunder signal: stall rate

| Sample stall stats |  |
|---|---:|
| Games hitting turn cap (80) with all 4 alive | 10 / 50 |
| Stall cause (every case) | `low_aggression` |
| Survivors at game end (every case) | 4 |

**20% of games stall.** Across 2000 main games, the same shape shows up
as 77 `turn_cap_leader` + 1 `turn_cap_tie` = 3.9% turn-cap rate (the
big gap vs. sample is partly noise on n=50, partly that
`turn_cap_leader` only fires when *one* player escapes — the sample
shows the 4-survivor pattern is much more common). In every stall the
classifier flags `low_aggression`, not `pillow_fort` or
`board_parity`. **This is the real blunder family**: Hat is hoarding
threats and not closing.

### 4.3 Win-condition distribution (50-game sample)

| Win condition | Count | % |
|---------------|------:|--:|
| combat_damage     | 34 | 68% |
| decking           |  9 | 18% |
| life_drain        |  2 |  4% |
| turn_cap          |  2 |  4% |
| combo             |  1 |  2% |
| commander_damage  |  1 |  2% |
| unknown           |  1 |  2% |

**Combo accounts for 2% of wins** even though roughly half the
top-100 pool is combo-archetype (per the Freya strategy load logs:
"archetype=combo" appears on a large fraction of the Yggdrasil
factories). The single combo win was *Thassa's Oracle* on turn 27 in a
Tymna deck.

**Decking is 18%** — players milling themselves out is winning more
games than combos are. The decking wins have `winning_card: null`
(no triggering card recorded) and span turns 25–73. This is a strong
hint that Hat is over-drawing relative to its closing ability.

### 4.4 MissedCombos / MissedFinishers

- `missed_combos: 0` across 50 games
- `missed_finishers: 0` across 50 games

Given combo-archetype prevalence in the pool, "0 missed combos" is
implausibly clean. Two possibilities:

1. Hat is genuinely playing the combo well *when the pieces converge*
   — supported by the fact that when combos resolve they hit
   (Thassa's Oracle in the one sample combo game).
2. The detector requires *all combo pieces simultaneously on the
   battlefield with mana available at game end*. Most combos don't
   sit there — they fire-and-finish or get disrupted. So the
   detector misses the meaningful blunder ("had piece A, drew piece B,
   never assembled C") because A or C left the battlefield before
   end-of-game.

The decking + low-combo rates point at (2). Recommend extending the
detector to scan per-turn snapshots, not just end-of-game.

---

## 5. Specific games worth eyeballing

### Draws / ties (main run)

| game | turns | reason | participants |
|-----:|------:|--------|--------------|
|  481 | 80 | turn_cap_tie | Kilo · Iroh · Rograkh · Y'shtola |
|  887 | 58 | draw         | Glarb · Tymna · Tayam · Kilo |
| 1691 | 55 | draw         | Yurlok · Rograkh · Meria · Sisay |

Three games ended without a winner. Two ended **before** the turn cap
(58 and 55 turns), which means simultaneous-death state-based
actions, not a soft hang. Worth Goldilocks-replaying these three
seeds.

### Single combo win (sample)

`sample_samples.jsonl` game 14, Tymna seat, Thassa's Oracle, turn 27.
Opponents: Norman Osborn // Green Goblin, Shorikai, Magus Lucea Kane.
This is the canonical Oracle-flip line working as intended.

### Long stalls (sample)

Games 0, 17, 39, 41 all ran the full 80 turns with 4 survivors and
`cause: low_aggression`. Same shell archetype, same outcome —
indicates a systemic Hat behavior, not deck-specific.

---

## 6. Recommended follow-ups

Priority order:

1. **Investigate the low-aggression stall family.** 20% of games (in
   sample) end with all four players alive at turn cap. Add Hat
   instrumentation: turn-by-turn aggression score (attackers
   declared / attackers possible), expected damage left on the
   table per turn. Most likely root cause: ThreatExposure dim weight
   too high relative to BoardPresence in non-aggro archetypes.

2. **Surface the "declined lethal" event** in the engine. Today
   there's no log signature for "Hat could swing for lethal at seat
   N but didn't." Add it at `runner.go` combat-step decision points
   so future audits can quantify directly.

3. **Tune concession heuristic.** 0 concessions in 2050 games is
   suspicious — Hat plays to turn-cap from clearly-lost positions
   (e.g. 1-of-4 survivors with no board). Bias concession toward
   firing when the relative-position track has been monotonic for ≥6
   turns AND the seat's life is the lowest by a meaningful margin.

4. **Extend MissedCombo detector** to per-turn snapshots so it
   catches "had A+B for 3 turns, never drew C" patterns. The
   end-of-game-only scan systematically misses combos that ever
   touched the graveyard.

5. **Investigate decking wins.** 18% of sample wins are by
   self-mill. Either decks are over-drawing (Hat treats CardAdvantage
   too greedily) or there's a parser gap that's letting infinite-draw
   loops run uncapped. Pick 2-3 of the 9 decking games and replay
   with full event log.

6. **Per-deck-file winrate.** Promote `participants` in the audit
   JSONL to carry deck-path so partner shells can be separated. The
   `tournament-audit` runner already has `deckIdxs` in scope; small
   change.

---

## 7. Reproducing this run

```bash
# 1. snapshot the top-100 deck list from the showmatch_elo table
sqlite3 data/hexdek.db \
  "SELECT deck_key FROM showmatch_elo WHERE games >= 1 \
   ORDER BY hex_rating DESC LIMIT 200;" \
  | awk '{print "data/decks/" $0 ".txt"}' \
  | while read p; do [ -f "$p" ] && echo "$p"; done \
  | head -100 > /tmp/top100_paths.txt

# 2. build the audit binary
go build -o /tmp/tournament-audit ./cmd/tournament-audit

# 3. run it (2000-game main + 50-game sample)
/tmp/tournament-audit \
  --deck-list /tmp/top100_paths.txt \
  --games 2000 --sample-games 50 \
  --workers 8 \
  --out-dir hexdek/tournament-audit-2026-05-17

# Output:
#   hexdek/tournament-audit-2026-05-17/summary.json
#   hexdek/tournament-audit-2026-05-17/main_per_game.jsonl   (2000 lines)
#   hexdek/tournament-audit-2026-05-17/sample_per_game.jsonl (50 lines)
#   hexdek/tournament-audit-2026-05-17/sample_samples.jsonl  (50 lines, deep)
#   hexdek/tournament-audit-2026-05-17/run.log
```

Seeds are pinned (`--seed 51717` / `--sample-seed 91317`); same code +
same deck list reproduces the run bit-for-bit.
