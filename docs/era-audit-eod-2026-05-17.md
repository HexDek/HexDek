# Era 1-4 Scaffold-Gap Audit — EOD 2026-05-17

Re-ran `scripts/era{1,2,3,4}_scaffold_audit.py` against
`data/rules/ast_dataset.jsonl` (49.8 MB, mtime 2026-05-16 17:55) to check
whether the day's merges moved the unbucketed condition counts.

## Headline

**No change.** All four EOD audit outputs are byte-identical to the
morning baselines in `data/rules/era*_scaffold_audit.md`
(`git diff --stat` against those four files: empty).

Expected: today's merge stream was per-card handlers
(`muninn-handlers-181-200`, `muninn-bulk-patterns-5`,
`sai-test-pollution`), hat eval-weight tuning (`hat-improvements`), and
UI work (`hot-cards-widget`, `visual-polish-round-7`). Nothing touched
`cmd/hexdek-thor/conditional_setup.go`, `detectConditionScaffold`, the
`BUCKETED_KINDS` set in the audit scripts, or the AST dataset itself.
Last scaffold-touching commits are all from the morning audit pass:

```
0cd6f32 thor/scaffolds: era 3 audit + 13 new condition scaffold kinds
6a60951 thor/scaffolds: era 2 + era 4 audits + 15 new condition scaffold kinds
7957992 thor/scaffolds: era 1 audit + 13 new condition scaffold kinds
```

## Counts vs morning baseline

| Era  | Cards  | Conditions | Bucketed | Unbucketed | Gap %  | Δ vs AM |
|------|--------|------------|----------|------------|--------|---------|
| 1    | 26,932 | 2,499      | 1,978    | 521        | 20.8%  | 0       |
| 2    |    537 |    75      |    75    |    0       |  0.0%  | 0       |
| 3    |    798 |    55      |    44    |   11       | 20.0%  | 0       |
| 4    |  3,696 |   514      |   409    |  105       | 20.4%  | 0       |
| **Σ**| **31,963** | **3,143** | **2,506** | **637**   | **20.3%** | **0** |

Trigger node totals (unchanged, reference only): era1 11,548 · era2 408
· era3 538 · era4 2,515 = 15,009.

The CLAUDE.md issue-log entry from 2026-05-08 cited "4,190 unbucketed
condition/trigger nodes across all 4 eras (33.9% of 12,363 total)".
The morning's three scaffold passes (commits above) closed the condition
side from that 33.9% combined figure down to the 20.3% condition-only
figure now standing, and EOD adds nothing further — the day's work was
elsewhere in the stack.

## Top remaining unbucketed kinds (unchanged, copied for one-glance review)

### Era 1 (521)
- `if` × 300
- `conditional` × 213
- `life_vs_half_starting` × 3
- `repeat_any_optional` × 3
- `life_threshold_both` × 1
- `life_delta_threshold` × 1

Top raw-text fragments still unbucketed (full list in
`data/rules/era1_scaffold_audit.md`): `you cast it` ×12,
`you cast it from your hand` ×9, `tribute wasn't paid` ×8,
`you have the city's blessing` ×6, `it has three or more +1/+1
counters on it` ×5, `it was bargained` ×5. Plus a long tail of
`{X} was spent to cast it` rows (5 distinct color pairs at ×2 each)
that would collapse under one `mana_color_spent_to_cast` scaffold.

### Era 3 (11)
- `if` × 6
- `conditional` × 5

All 11 are single-occurrence raw fragments (judgment counters,
saddled+excess-damage, decayed inversion, bargained — each appearing
on a single card).

### Era 4 (105)
- `if` × 65
- `conditional` × 38
- `life_delta_threshold` × 1
- `life_vs_half_starting` × 1

Highest-leverage raw cluster is `an opponent controls more lands than
you` ×5 (Claim Jumper / Knight of the White Orchid / Sunstar
Expansionist family — single scaffold would clear all five), and
`artifact entered the battlefield under your control this turn` ×2
(Mechan Shieldmate / Shipwreck Sentry — could fold into an
`artifact_etb_this_turn` scaffold paired with the existing `for_each`
infrastructure).

## False-positive / over-firing check

Scanned the bucketed `Kind` histograms in all four era audit files for
any kind whose count is implausibly high given the card pool. Nothing
looks suspicious:

- `paid_optional_cost` is the leader in every era (era1 561, era2 38,
  era3 16, era4 ~) — expected, since kicker / overload / bestow /
  surge / bargain all route through it.
- `if` and `conditional` counts inside the bucketed column reflect raw
  conditions that matched a `RAW_PATTERNS` regex; the underlying scaffold
  identified by the regex name is what actually fires at runtime, so a
  high `if` count here is not the same as a scaffold over-firing.
- New `was_kicked`, `hellbent`, `raid`, `spell_mastery`,
  `gained_life_this_turn`, `creature_died_this_turn`,
  `no_spells_cast_last_turn`, `two_plus_spells_cast_last_turn`,
  `you_control_creature_power_ge`, `etb_tapped_unless`, `domain`,
  `etb_if`, `repeat_n`, `lieutenant`, `ki_counters_ge_2`,
  `self_is_tapped`, `attacked_or_blocked_this_combat`, `coven`,
  `self_has_counter`, `didnt_attack_this_turn`,
  `dealt_damage_to_opponent_this_turn`, `no_mana_spent_to_cast` kinds
  added in the morning passes all sit in the expected single- to
  low-double-digit range. None are wildly inflated; no false-positive
  flags raised.

No goldilocks or Muninn signal from today is pointing at a bucketed
scaffold firing where it shouldn't.

## Next-pass leverage (carried forward unchanged)

If we want another bite at the era-1 gap, the highest-yield additions
would be:

1. `you_cast_it` / `you_cast_it_from_hand` (would close 21 of the 521).
2. `tribute_unpaid` (8 cards, mechanic-locked).
3. `city's_blessing` (6, ascend mechanic).
4. `mana_color_spent_to_cast` (covers the `{R}{R}`, `{G}{G}`, `{U}{U}`,
   `{R}`, `{U}` was-spent rows — ~12 conditions across one scaffold).
5. `bargained` (5 cards, mirrors the era-3 single-card entry).

Estimated condition reduction if items 1-5 land: ~52 era-1 + ~1 era-3
+ ~0 era-4 ≈ 53 (would take total gap from 637 → ~584, 20.3% → ~18.6%).
