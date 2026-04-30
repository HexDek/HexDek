# Tool - Valkyrie

> Last updated: 2026-04-29
> Source: `cmd/mtgsquad-valkyrie/`

Deck regression runner. Loads every saved deck from `data/decks/`, plays real Commander games with [GreedyHat](Greedy%20Hat.md) opponents, reports issues.

## What It Catches

```mermaid
flowchart TD
    Run[Valkyrie run] --> Load[Load every .txt deck<br/>under data/decks/]
    Load --> Game[Play real game<br/>GreedyHat all seats]
    Game --> Issues{Detect}
    Issues --> I1[Crashes / panics]
    Issues --> I2[Parser gaps<br/>UnknownEffect nodes]
    Issues --> I3[Invariant violations]
    Issues --> I4[Empty-option turns<br/>hat had nothing to do]
    Issues --> I5[Zero-mana games<br/>broken color identity]
    Issues --> I6[Commanders never<br/>offered as castable]
    I1 --> Report[Aggregate report]
    I2 --> Report
    I3 --> Report
    I4 --> Report
    I5 --> Report
    I6 --> Report
```

## Why It's Distinct from Loki

- [Loki](Tool%20-%20Loki.md) = random decks, finds invariant violations
- [Thor](Tool%20-%20Thor.md) = exhaustive per-card, finds effect bugs
- **Valkyrie** = real saved decks, finds regressions specific to "decks Josh + 7174n1c actually play" — catches issues that random decks miss because the random decks never assemble that specific cohort

## Usage

```bash
go run ./cmd/mtgsquad-valkyrie
go run ./cmd/mtgsquad-valkyrie --decks data/decks/lyon --games 10
go run ./cmd/mtgsquad-valkyrie --verbose --fail-fast
```

## Output

Per-deck breakdown of failure categories. Designed as a CI smoke test against the curated portfolio (32 decks across 9 folders, see [Engine Overview](Engine%20Overview.md)).

## Related

- [Tool - Loki](Tool%20-%20Loki.md)
- [Tool - Thor](Tool%20-%20Thor.md)
- [Tool - Tournament](Tool%20-%20Tournament.md)
