# Graveyard — Buried Tools & Features

> Deprecated components that served their purpose and were retired. Documented here so future contributors know what existed, why it was built, and why it was cut.

## Buried Tools

| Tool | Lived | Died | Cause of Death |
|---|---|---|---|
| **Loki** | 2026-03 → 2026-05-06 | Fishtank + Feynman + Muninn | Random chaos testing without diagnostic context. The fishtank runs 24/7 random pods with full Heimdall telemetry; Feynman checks all 20 invariants on every action in every game; Muninn logs crashes with full card context. Loki found bugs but couldn't explain them. |
| **Parity** | 2026-03 → 2026-05-06 | Migration complete | Go↔Python engine equivalence verifier. Built to ensure the Go rewrite matched Python output byte-for-byte. Python engine no longer maintained — nothing to compare against. |
| **GreedyHat** | 2026-02 → 2026-04-26 | YggdrasilHat | First-gen player AI. Pure heuristic, no search, no politics, no combo awareness. Kept briefly for Parity byte-equivalence testing, retired when Parity retired. |
| **PokerHat** | 2026-03 → 2026-04-26 | YggdrasilHat | Second-gen player AI. Added bluff/fold logic and opponent modeling. Superseded by Yggdrasil's unified UCB1 + eval + politics in a single brain. |
| **OctoHat** | 2026-04 → 2026-04-26 | Never shipped | Test-only 8-seat hat experiment. 7174n1c confirmed it was test scaffolding, not a real implementation. |
| **MCTSHat** | 2026-04 → 2026-04-26 | Absorbed into YggdrasilHat | Monte Carlo tree search as standalone hat. Budget dial in Yggdrasil replaced the need for a separate MCTS wrapper — budget≥200 activates rollouts natively. |

## Buried Features

| Feature | Lived | Died | Cause of Death |
|---|---|---|---|
| **Python reference engine** | 2025 → 2026-04 | Go engine is sole implementation | Original prototype. Go rewrite achieved full parity then surpassed it in performance (10,000x). No reason to maintain two implementations. |
| **Certstream (phase 1b)** | 2025 → 2026-01-19 | calidog.io stopped sending data | Real-time certificate transparency feed. Service died upstream. Replaced by phase_1c direct CT log aggregator connecting to Google/Let's Encrypt servers. |
| **Quorum persona (Garfield/Rosewater/Venters)** | 2026-03 → 2026-04-15 | IP exposure risk | MTG Wizard channel used real WotC employee names as AI personas. Hard-renamed to Attacker/Defender/Builder archetypes before public release. |
| **Amiibo (hat naming)** | 2026-03 → 2026-05-05 | Nintendo IP risk | Hat personality system was called "amiibo." Renamed to "Curse Proficiency" with cymatic sigil visual. 322 references across 25 files renamed. |

## Buried Docs (consolidated or pending)

| Doc | Status | Action |
|---|---|---|
| `Greedy Hat.md` | Deprecated content | Flag as legacy, point here |
| `Poker Hat.md` | Deprecated content | Flag as legacy, point here |
| `MCTS and Yggdrasil.md` | Overlaps with `YggdrasilHat.md` | Merge MCTS explainer into YggdrasilHat as a "Search Algorithm" section |
| `architecture-learning-loop.md` (top-level DRAFT) | Superseded by `architecture/Learning Loop.md` | Delete — design spec is stale, production doc is authoritative |
| `architecture-hat-evolution.md` (top-level DRAFT) | Roadmap doc from design session | Keep as historical record of the plan, but flag as snapshot-in-time |

Note: `Tool - Freya.md` and `Freya Strategy Analyzer.md` are NOT duplicates — the former is an intentional quick-reference page that points to the full deep dive. Layered docs pattern, not redundancy.

## Why Document Deprecations

Code gets deleted. Git blame shows what changed but not *why something was retired*. This file is the institutional memory for architectural decisions that removed things — so nobody rebuilds Loki wondering "why didn't they just do random chaos testing?" The answer is always here.
