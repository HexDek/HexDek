# CR Parser Audit — R37

**Date:** 2026-05-17
**CR file:** `data/rules/MagicCompRules-20260227.txt` (effective 2026-02-27)
**Branch:** dev/cr-parser-audit
**Scope:** Every `§XXX.X[a-z]?` citation in `internal/**/*.go` vs. sections present in the CR text.

## Methodology note

There is **no Go package that parses the CR** — HexDek has no `internal/rules/` directory. The CR text is referenced solely through human-authored comments (`§614`, `§903.9b`, etc.) in engine source. For this audit the "parsed corpus" is the set of section/subrule identifiers extractable directly from the CR text (`grep -ohE '^[0-9]{3}(\.[0-9]+[a-z]?)?'`). A reference is "missing" when the engine cites an identifier the current CR file does not contain (typically renumbered, removed, or yet-to-be-written sections), and "orphan" when a CR section is present but never cited.

Extraction commands:
```
grep -rohE "§[0-9]{3}(\.[0-9]+[a-z]?)?" internal/ --include='*.go' | sort | uniq -c | sort -rn
grep -ohE "^[0-9]{3}(\.[0-9]+[a-z]?)?" data/rules/MagicCompRules-20260227.txt | sort -u
```

## (a) Totals

| Metric | Count |
|---|---|
| Total `§XXX.X` citations in `internal/` (with duplicates) | **2 970** |
| Unique CR identifiers cited by engine | **724** |
| CR identifiers present in MagicCompRules-20260227.txt | **3 266** |
| Cited identifiers that match the CR corpus | **675** (93.2 % of citations) |
| Cited identifiers MISSING from the CR corpus | **49** (6.8 %) |
| CR identifiers never cited by engine (ORPHANS) | **2 591** (79.3 % of corpus) |

## (b) Present (sample, by top-level section)

Hit-coverage breakdown — top-level sections where engine citations actually land in the CR corpus:

| Section | Hits | What lives there |
|---|---|---|
| 702 | 292 | Keyword ability rules |
| 701 | 64 | Keyword actions |
| 704 | 30 | State-based actions |
| 613 | 21 | Layer system |
| 603 | 18 | Triggered abilities |
| 903 | 11 | Commander rules |
| 601 | 11 | Casting spells |
| 712 | 10 | Sagas |
| 205 | 10 | Type lines |
| 616 | 8 | Replacement-effect interaction |
| 608 | 8 | Resolving spells/abilities |
| 800 | 7 | Multiplayer general |
| 509 | 7 | Declare blockers |
| 614 | 6 *(of 78 citations — see Missing)* | Replacement effects |
| 117 | 6 | Timing and priority |

## (c) MISSING — engine cites but CR text does not contain

49 unique identifiers. Two are not real CR identifiers (`702.17x`, `702.19x` are placeholder range-notation in `keywords_stubs_tail.go` comment header — drop from the "broken" tally → **47 genuine misses**).

### High-impact misses (cited 5+ times)

| § | Count | Where engine cites it | Likely status |
|---|---|---|---|
| **706.10** + 706.10a/b/c/f | 27 + 13 = 40 | `resolve.go` (token-copy creation), `keywords_batch3_replicate_test.go` (Replicate) | **Renumbered.** Token-copy rules now live elsewhere in §707 area; current §706 is "Rolling a Die" (only goes to 706.8). Largest single drift in the engine. |
| **726.3a** | 17 | mulligan-related code | Subrule never existed; current 726.3 has no subrules. Citation should be `§726.3`. |
| **716.5** | (low; counted separately) | `keywords_class.go` (Class card static abilities) | Section 716 stops at 716.4. Likely renumbered when Class card rules were reorganized. |

### All 49 missing identifiers

```
112.6k    116.1b    212.3f    310.5b    400.10b   510.5     602.2d
700.5b    701.34c   701.39b   701.51d   701.62c   701.71    701.71a
701.71b   702.107b  702.117b  702.11k   702.137b  702.144b  702.146c
702.176b  702.176c  702.17x*  702.181b  702.183b  702.192   702.192a
702.192b  702.192c  702.19x*  702.21c   702.27b   702.27c   702.34b
702.34c   702.36e   702.74b   702.74c   702.84b   702.84c   702.93c
706.10    706.10a   706.10b   706.10c   706.10f   716.5     726.3a
```
*`702.17x` and `702.19x` are placeholder notation in `keywords_stubs_tail.go`, not real citations.*

The pattern in 702-space is **stub-subrule overreach** — engine code anticipates an additional `b`/`c` subrule on a keyword that has only `a` in the current CR (or no subrules at all). Examples: 702.34b/c (Banding), 702.74b/c (Fading), 702.84b/c, 702.27b/c. These either reference text removed in a later CR update or describe behavior the engine added speculatively.

The 701-space misses (701.34c, 701.39b, 701.51d, 701.62c, 701.71*) follow the same pattern in keyword actions.

## (d) ORPHAN sections — CR has, engine never cites (2 591)

Coverage gaps grouped by top-level section (highest = least covered):

| Section | Orphans | Topic | Why uncovered |
|---|---|---|---|
| 702 | 463 | Keyword abilities | Long tail of obscure keywords (e.g. Banding-with-other, esoteric subrules) |
| 701 | 226 | Keyword actions | Same — subrule depth not all reached |
| 712 | 49 | Sagas | Only chapter mechanics partly modeled |
| 107 | 47 | Numbers and symbols | Mostly editorial / glossary-adjacent |
| 113 | 41 | Abilities | Subrule depth not exhausted |
| 903 | 39 | Commander | Variant subrules (1v1, partner permutations) |
| 901 | 36 | Planechase | Variant not implemented |
| 508 | 35 | Declare attackers step | Sub-clauses for restrictions/requirements |
| 118 | 35 | Costs | Modal/alternative-cost edge cases |
| 614 | 32 | Replacement effects | (Top-cited section overall — orphans are obscure sub-subrules) |
| 700 | 31 | General game | |
| 603 | 30 | Triggered abilities | |
| 111 | 30 | Tokens | |
| 103 | 30 | Starting the game | |
| 801 | 29 | Multiplayer limited range | Variant not implemented |

Most orphans are **expected** — HexDek targets 4-player Commander and skips Planechase, Conspiracy, Vanguard, Archenemy, etc. The orphan count is not a defect signal on its own; it becomes interesting only when an orphan section is in a rule area HexDek claims to model.

**Suspicious orphan clusters** (would benefit from a deeper audit pass):
- §508 (Declare Attackers) — 35 orphans against only ~6 attacker-related citations in the engine; combat-restriction subrules may be underused.
- §118 (Costs) — 35 orphans; alternative-cost handling may be shallow.
- §603 (Triggered Abilities) — 30 orphans, despite this being core engine territory.

## (e) Top 10 most-referenced CR sections

| Rank | § | Count | Topic |
|---|---|---|---|
| 1 | **§614** | 78 | Replacement effects |
| 2 | **§613** | 38 | Layer system (continuous effects) |
| 3 | **§601.2f** | 38 | Final cost determination & legality on cast |
| 4 | **§903.9b** | 33 | Commander zone-change redirection |
| 5 | **§706.10** | 27 | Token-copy creation — **MISSING from current CR** |
| 6 | **§704.6c** | 26 | SBA: zero-toughness creature destruction |
| 7 | **§726.2** | 22 | Restart-game card retention |
| 8 | **§800.4a** | 18 | Multiplayer leave-game cleanup |
| 9 | **§707.2** | 18 | Face-down spells |
| 10 | **§704.5a** | 18 | SBA: 0-life loss |
| 10 | **§101.4** | 18 | APNAP order |

The dominance of §614 (replacements), §613 (layers), §601.2f (cast legality), and §903.9b (commander redirect) reflects exactly where engine complexity concentrates — these are also the four areas with the largest forensic write-ups in CLAUDE.md.

## Recommendations

1. **Fix `§706.10*` citations** (40 occurrences) — the largest documented/code drift in the engine. Re-cite against current CR or note the comment is intentionally vintage. `resolve.go` and `keywords_batch3_replicate_test.go` are the impacted files.
2. **Fix `§726.3a` → `§726.3`** (17 occurrences) — subrule never existed.
3. **Fix `§716.5`** — Class static-ability rule, current CR caps at 716.4. Likely now under a different paragraph.
4. **Sweep `702.{N}b`/`702.{N}c` stub-subrule citations** — engine anticipates depth the CR doesn't have. Either drop the subletter or document them as speculative.
5. **No CR parser package exists.** If the engine wants to enforce citation validity (e.g., a `//go:generate` check that every `§XXX.X` comment resolves to a section in the bundled CR text), this audit's grep pipeline is the seed for that check. Recommend wiring it into `cmd/hexdek-judge` or a `go vet` pass.

---

*Generated 2026-05-17 from r37 main (a94d7aa) against CR effective 2026-02-27.*
