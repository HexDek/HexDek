# Task: Thor 2.0 — Muninn Cross-Reference

## Context
Thor finds failures in synthetic isolation. Muninn logs failures from live tournament games. These two sources don't always agree:
- Some cards fail Thor but work fine in real games (false positives — Thor's setup is wrong)
- Some cards work in Thor but fail in real games (false negatives — Thor doesn't test the right interaction)

Cross-referencing identifies both, letting us prioritize real bugs over test harness issues.

## Task
Build a cross-reference report that diffs Thor results against Muninn's gap data:

1. **Load Muninn data** — read Muninn's persistent gap files:
   - Look in `data/` for Muninn output files (likely `muninn-gap-cards.txt`, `muninn-gap-thor-report.md`, or similar)
   - Parse card names that have live-game failures logged by Muninn
   - Also check `internal/tools/muninn/` or `cmd/hexdek-muninn/` for the data format

2. **Load Thor results** — read the most recent Thor corpus audit report:
   - Parse `data/corpus-audit-full-report.md` or similar for cards that fail Thor
   - Group by failure type (draw, life, damage, goldilocks, etc.)

3. **Cross-reference** — produce three lists:
   - **Both fail** (high priority): cards that fail in Thor AND have Muninn live-game gaps. These are confirmed real bugs.
   - **Thor-only failures** (lower priority): cards that fail Thor but have no Muninn entry. Likely test harness issues.
   - **Muninn-only failures** (concerning): cards that PASS Thor but fail in real games. Thor has a blind spot for these — needs investigation.

4. **Output** — write `data/thor-muninn-crossref.md`:
   ```markdown
   # Thor ↔ Muninn Cross-Reference — {date}
   
   ## Both Fail (confirmed bugs) — {count}
   | Card | Thor Failure | Muninn Context |
   
   ## Thor-Only (likely harness issues) — {count}  
   | Card | Thor Failure | Notes |
   
   ## Muninn-Only (Thor blind spots) — {count}
   | Card | Muninn Gap | Why Thor Misses It |
   ```

5. **CLI** — this can be a standalone script or integrated into Thor as a `--crossref` flag. Whichever is simpler given the existing architecture. If standalone, put it in `cmd/hexdek-thor/crossref.go` or a small script in `scripts/`.

Look at:
- `data/muninn-gap-cards.txt` — list of cards with Muninn-logged gaps
- `data/muninn-gap-thor-report.md` — Thor results on Muninn gap cards (already exists from earlier run)
- `data/corpus-audit-full-report.md` — full corpus audit results (on DARKSTAR)
- `cmd/hexdek-muninn/` — understand Muninn's data format
