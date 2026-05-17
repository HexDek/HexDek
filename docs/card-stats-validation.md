# /api/card-stats/{commander} Validation

**Date:** 2026-05-17
**Branch:** dev/card-stats-validation
**Trigger:** Hot Cards widget (dev/hot-cards-widget) consumes this endpoint
**Audit target:** Is the win-rate the widget shows actually right?

## TL;DR

The arithmetic is right, but the metric is wrong, the serialization breaks the widget that consumes it, and one of the five sampled commanders returns nothing despite 200k games. Four discrepancies, ranked:

1. **JSON casing mismatch — Hot Cards widget can never render any data.** API emits PascalCase (`Games`, `Wins`, `CardName`); React reads snake_case (`s.games`, `s.wins`, `s.card_name`). Every value reads `undefined`, every `games >= 20` filter fails, both `04.CS CARD STATS` and `04.HC HOT CARDS` panels collapse to the "not enough data" empty state for every deck.
2. **Win-rate is biased to 100% by construction.** Losing seats are empty at game-end (loss = lose all permanents), so the recorder almost never sees a losing battlefield. 47/50 top cards for Queen Marchesa report WR = 100%, including basic Plains and Command Tower.
3. **`BoardPresence` is degenerate (always 1.0).** `OnBoardAtWin` is incremented only inside the `isWinner` branch, by the same amount as `Wins`. `BoardPresence = OnBoardAtWin / Wins` is therefore `1.0` for every row.
4. **`Sin, Spira's Punishment` returns 0 cards despite 203,039 games / ~49k wins.** Not a name-encoding bug — the leaderboard string matches byte-for-byte. The commander has no rows passing the `games >= 10` threshold.

Math, ordering, threshold, and commander-field semantics are all correct. The endpoint reports what the table contains; the table is what is wrong.

## Methodology

- Pulled `/api/card-stats/{commander}` for 5 commanders chosen by descending game count from `/api/live/elo`:
  - Queen Marchesa (216,140 g)
  - Varina, Lich Queen (203,852 g)
  - Sin, Spira's Punishment (202,904 g)
  - Sigarda, Font of Blessings (171,926 g)
  - Szarekh, the Silent King (158,968 g)
- Tried to reconstruct stats from `/api/games`. **Cannot:** the in-memory `gameHistory` is rehydrated from SQLite via `db.LoadGameSeats` (`internal/db/showmatch.go:201-220`), whose `SELECT` omits the `battlefield_cards` column. Every `final_seats[].battlefield` in `/api/games` is `[]`, including the winner's. The persisted JSON in `showmatch_game_seat.battlefield_cards` exists but is unreadable from any public route.
- Fell back to two consistency checks:
  - Field-level math against the formula in `db.LoadCardWinStats` (`internal/db/showmatch.go:293`).
  - Cross-endpoint agreement against `/api/card-stats/card/{cardName}/by-commander`.

## Results

### Per-commander summary

| Commander | Returned | WR=1.0 count | WR range | Games range | Notes |
|---|---|---|---|---|---|
| Queen Marchesa | 50 | 47 / 50 | 0.952 – 1.000 | 10 – 54 | Basic lands also report 100% |
| Varina, Lich Queen | 5 | 5 / 5 | 1.000 – 1.000 | all entries 100% wins |
| Sin, Spira's Punishment | **0** | — | — | — | 203k commander games, no `card_win_stats` rows ≥ 10 |
| Sigarda, Font of Blessings | 21 | 12 / 21 | 0.909 – 1.000 | 10 – 48 | Only commander where losing-seat snapshots leaked through |
| Szarekh, the Silent King | 50 | 0 / 50 | 0.920 – 0.992 | 11 – 266 | Highest absolute sample sizes; every row still WR > 90% |

### Field-level checks (all 5 commanders, 126 total rows)

| Check | Pass / Total |
|---|---|
| `WinRate == Wins / Games` (float exact) | 126 / 126 |
| `BoardPresence == OnBoardAtWin / Wins` (for `Wins > 0`) | 126 / 126 |
| `BoardPresence == 1.0` exactly | 126 / 126 |
| `OnBoardAtWin == Wins` exactly | 126 / 126 |
| `Games >= 10` (filter from `LoadCardWinStats`) | 126 / 126 |
| `Commander` field == path commander | 126 / 126 |
| Result sorted descending by `WinRate` | 5 / 5 commanders |

### Cross-endpoint check

`Trazyn the Infinite` × `Szarekh, the Silent King`:
- `/api/card-stats/Szarekh%2C%20the%20Silent%20King` row: `Games=125 Wins=124 WinRate=0.992`
- `/api/card-stats/card/Trazyn%20the%20Infinite/by-commander` row: `games=125 wins=124 win_rate=0.992`

Agreement, byte-for-byte.

## Discrepancy 1 — JSON casing breaks the consumer

`db.CardWinStat` (`internal/db/showmatch.go:50-58`) has no `json:` struct tags. `encoding/json` falls back to field names, so the wire format is:

```json
{"CardName": "Plains", "Commander": "Queen Marchesa", "Games": 54, "Wins": 53,
 "OnBoardAtWin": 53, "WinRate": 0.9814..., "BoardPresence": 1}
```

`hexdek/src/screens/DeckArchive.jsx:2025-2037` and `:2086-2100` (Hot Cards widget) both read:

```js
.filter(s => deckCardNames.has(s.card_name || s.name))
.filter(s => (s.games_included || s.games || 0) >= 20)
.map(s => {
  const games = s.games_included || s.games || 0
  const wins  = s.wins_when_included || s.wins || 0
  ...
})
```

None of `card_name`, `games`, `wins`, `games_included`, `wins_when_included` exist on the wire response. The first `filter` drops every row (no `card_name`), or — if it survived — the `games >= 20` filter drops everything anyway (`undefined || 0 = 0`). The widget always renders the "Not enough card-level data yet" empty state.

**Fix:** add JSON tags to `db.CardWinStat`. Recommended snake_case to match the rest of the API surface:

```go
type CardWinStat struct {
    CardName      string  `json:"card_name"`
    Commander     string  `json:"commander"`
    Games         int     `json:"games"`
    Wins          int     `json:"wins"`
    OnBoardAtWin  int     `json:"on_board_at_win"`
    WinRate       float64 `json:"win_rate"`
    BoardPresence float64 `json:"board_presence"`
}
```

Not done in this branch — out of scope for a validation pass, and the fix needs paired QA on every other consumer of `db.CardWinStat` (CSV exports, Heimdall reports, etc.) that may already depend on PascalCase.

## Discrepancy 2 — win-rate biased to 100%

The recorder in `internal/hexapi/showmatch.go:1822-1866` walks `g.FinalSeats[i].Battlefield` for every seat and emits `(card_name, commander, win=isWinner?1:0)` rows. The intent is "of N games where this card was on the seat's battlefield at game end, how often did the seat win?" — a board-presence-conditioned win rate.

In practice, **losing seats lose their permanents**: a seat reaches 0 life with all their creatures already dead, or gets decked with everything destroyed. By the time `FinalSeats` is built, `seat.Battlefield` is essentially empty for losers. The Sigarda result (12 / 21 at WR=1.0) is the only commander in the sample where any losing-seat rows leaked through, and even there the losses are single-digit per row.

Observable signal:

- Queen Marchesa-the-card on Queen Marchesa-the-deck: 43 games / 43 wins, WR = 100%. Commander-level WR is 56,961 / 216,140 ≈ 26.3% (the 4-player baseline).
- Plains × Queen Marchesa: 54 games / 53 wins, WR = 98.1%.

A card that is "on the battlefield at game end" essentially **is** the seat that won. The metric is `P(win | card on BF at game end)` — which is ~1.0 by construction, not the `P(win | card in deck)` that the Hot Cards widget assumes.

The Hot Cards lift formula `(wr − 25) × √games` is built around the 25% 4-player baseline. With the current data, `wr − 25` is ~75 for almost every card. If the widget ever rendered (see Discrepancy 1), every card in every deck would qualify as "hot."

**Two ways to read this:**
- *As designed:* the metric is honest about what it measures (BF-presence-conditioned WR) and the consumer just needs to use the right baseline. The right baseline is not 25% — it would be the empirical mean of all rows for that commander.
- *As the consumer assumes:* the metric should be "win rate given the card was in the deck list." That requires a separate counting path that increments `games` for every game where the card was in the seat's *decklist* (not BF), and `wins` only when the seat won. This already exists at `internal/hexapi/showmatch.go:1882-` (the "deck-list-based per-card aggregate" that drives `/api/cards/{name}/stats`). The Hot Cards widget is reading from the wrong endpoint.

Recommend the latter: rewire `getCardStatsByCommander` to a per-commander decklist-conditioned source. Out of scope for this branch.

## Discrepancy 3 — `BoardPresence` is tautologically 1.0

`internal/hexapi/showmatch.go:1853-1865`:

```go
for name := range bfSet {
    win := 0
    onBoard := 0
    if isWinner {
        win = 1
        onBoard = 1
    }
    cardStats = append(cardStats, db.CardWinStat{
        ...
        Wins:         win,
        OnBoardAtWin: onBoard,
    })
}
```

`onBoard` only fires inside `isWinner`, and it fires at the same time and amount as `win`. The UPSERT in `BatchUpsertCardWinStats` adds them as deltas. So for every (card, commander) row, `OnBoardAtWin == Wins` exactly. `BoardPresence = OnBoardAtWin / Wins = 1.0` for every row. Confirmed across 126/126 rows.

The metric was presumably intended to mean "of all games this card was in the deck, in what fraction did it actually reach the battlefield at game end" — but with the current accounting, the numerator and denominator are the same column. The field carries zero information and could be removed (or fixed by counting decklist-presence into the denominator).

## Discrepancy 4 — Sin, Spira's Punishment returns 0 rows

`Sin, Spira's Punishment` has 203,039 games and ~49k wins in `/api/live/elo`, identical commander string. Encoding-checked against three URL variants and confirmed the leaderboard string byte-for-byte (no smart-quote or whitespace mismatch). Despite ~49k commander wins, no `card_win_stats` row crosses the `games >= 10` threshold.

For comparison Szarekh has 158k commander games and the top card crosses 266 BF appearances. Sin has more games, fewer wins per game maybe, but the asymmetry doesn't explain a complete absence.

Possible causes (not investigated this pass):
- Sin's wins persist via a code path that doesn't reach `BatchUpsertCardWinStats` (gauntlet replays vs. live showmatch are two persistence paths).
- A recent commander whose `card_win_stats` rows began accumulating only after a backfill window.
- An ID/name normalization bug for legendary creatures with comma in the name (note Sigarda, Szarekh, Varina, Queen all parse fine; Sin is the outlier).

Flag for follow-up. Worth a SQL `SELECT count(*), max(games) FROM card_win_stats WHERE commander='Sin, Spira''s Punishment'` on DARKSTAR to confirm whether rows exist below threshold or not at all.

## What the audit could *not* check

- **Per-row truth** ("did Trazyn the Infinite actually win 124 of 125 games it appeared in for Szarekh?"). Requires access to the source `showmatch_game_seat.battlefield_cards` column, which is persisted but not exposed by any GET route. `LoadGameSeats` (`internal/db/showmatch.go:201`) omits the column from its `SELECT`.
- **Throughput** (whether `BatchUpsertCardWinStats` is being called for every completed game vs. a subset). Inferred from the sparse counts but not confirmed.

A direct DB query on DARKSTAR would close both gaps in 5 minutes.

## Recommended follow-ups (priority order)

1. **Add `json:` tags to `db.CardWinStat`.** Hot Cards widget is currently dark. (1-line diff, plus audit of other consumers.)
2. **Decide what `/api/card-stats/{commander}` actually means** and either rename it (e.g. `/api/card-stats/{commander}/board-presence`) or rewire it to decklist-conditioned counts. Hot Cards expects the latter.
3. **Remove `BoardPresence`** or rebuild it from a different denominator.
4. **Investigate Sin, Spira's Punishment** — query `card_win_stats` directly on DARKSTAR.
5. **Surface `battlefield_cards` on `/api/games/{id}`** so future validation passes don't have to ssh into the host. Trivial: add the column to the `LoadGameSeats` SELECT and pass through.
