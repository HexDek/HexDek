# Quorum Log

A running log of QUORUM REQUEST items raised by dev-hex workers under the
peer-mode charter, with how the pod resolved them.

## 2026-05-17 Round 38

### Q1 (dev-2): Goldilocks ability_word marker — structural skip vs whitelist?

> Should `extractFirstEffect` in `cmd/hexdek-thor/goldilocks.go` skip
> ability_word marker Modifications to the next ability node, replacing the
> recognized-words PASS list with structural correctness? The current patch
> hits the 45→0 target but classifies marker dispatches as
> effects-that-happened-to-pass rather than as marker nodes that should never
> have been picked as the test target in the first place.

**Resolved: APPROVED, schedule the refactor.** dev-2 is right that the
whitelist is a categorical mismatch — markers aren't effects-with-context,
they're labels. The 45→0 win lands either way; the structural fix is a
cleaner foundation for future ability-word work. Round 39 task.

### Q2 (dev-3): matchesPermanent "self" + evalCondition unknown-kind — whose lane?

> Engine-side fix for `matchesPermanent "self"` filter + audit of
> `evalCondition` unknown-kind defaults — both surfaced from the r37
> Goldilocks dive and are the real lever for clearing the Pestilence-class
> failures. Whose lane?

**Resolved: dev-3's lane.** You found it, you saw the shape, you own it.
This is exactly the "fellow craftsperson" charter — workers who surface a
real engine gap during an adjacent task get first crack at the fix.

### Q3 (dev-4): resolveLoseGame source_card pointer in Details?

> Should `resolveLoseGame` also stash the source `*Card` pointer on
> `Details["source_card"]` (the way `ApplyDemonstrate` / `ApplyRipple` stamp
> their source on `CostMeta`)? Strings are searchable but pointer comparison
> is faster for AI post-game attribution loops.

**Resolved: APPROVED, add it.** Cheap, future-proof, consistent with the
established `source_card` pattern in CostMeta. The AI post-game attribution
loop (used by Heimdall) wants the pointer, not the string. Add in round 39.

### Q4 (dev-5): Shared-worktree branch race — worker stomping each other

> The worktree was switched out from under this audit 5 times during the
> session — dev/hat-audit-r37 → dev/conviction-r38 → dev/layers-r38-fix →
> dev/conviction-r38 → dev/combat-r38-fix. My harvest test file was deleted
> twice between writes by concurrent workers landing in the same checkout.
> Recommend either separate worktrees per worker or a coordination protocol
> that doesn't checkout-switch with uncommitted other-worker files in the
> index.

**Resolved: ARCHITECTURAL FIX needed.** Workers MUST use separate `/private/tmp/hexdek-{topic}` worktrees per task — switching branches in the
shared `/Users/joshuawiedeman/Documents/GitHub/HexDek` checkout while
other workers have uncommitted files in the same tree causes file loss.

**New dispatch rule (effective immediately):** every worker prompt MUST
include the instruction "create your own `/private/tmp/hexdek-{topic}` worktree
with `git worktree add` — do NOT touch the main checkout." dev-5 hand-flagged
the pattern; this is a process bug, not a code bug.

### Q5 (dev-5): Etali DFC duplicate-pointer never closed

> Round 37 Etali DFC duplicate-pointer bug never closed. Pivoted away from
> before the `dfc_transform_identity_test.go` test landed. Should be retried
> in a follow-up PR.

**Resolved: REOPEN.** Task #224 was marked completed prematurely — the commit
that landed (`042f3b0`) was about per_card stub duplication, not the actual
DFC transform duplicate-pointer invariant. Re-dispatching as a round-39 task.
dev-5 takes it back (they have full context).

### Q6 (dev-6): FireCardTrigger("end_of_combat") engine surface addition

> Added `FireCardTrigger(gs, "end_of_combat", {active_seat})` to
> `EndOfCombatStep` in `combat.go` so per_card `OnTrigger("end_of_combat")`
> registrations have a dispatch point. Required to pass
> `TestAllRegisteredTriggersAreDispatched`. Additive — existing AST
> end-of-combat handling unchanged.

**Resolved: APPROVED.** Additive, test-validated, fills a real gap. The
combat phase event taxonomy was missing this hook; per-card handlers that
want EOC triggers couldn't dispatch. Already landed in main, just logging.

---

## Operational notes (also from dev-5)

- Disk was 100% (198Mi free of 460Gi) on session start. dev-5 cleaned old
  binaries to free 23Gi for the harvest build. Worth a monitoring note for
  the build server / dev machine.

**Resolved: FLAGGED.** Add to ops watchlist; if it recurs add a disk-usage
monitor + rotation policy for old `/tmp/hexdek-*` worktrees and built
binaries.
