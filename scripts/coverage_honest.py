#!/usr/bin/env python3
"""Dual-metric coverage reporter — separates "structurally typed AST" cards
from "custom-stub" cards to give an honest picture of what the parser
actually produces vs. what's been tagged for future engine work.

The "100% GREEN" number from `parser.py` is real in the sense that every
card returns an AST without parse errors. But some of that AST is:
  - typed nodes (Damage, Buff, Tutor, Destroy, etc.) that an engine can
    execute directly
  - Modification(kind="custom", args=("slug",)) stubs that say "this card
    has an effect; look up the slug in the runtime registry"
  - per-card handlers from per_card.py that intentionally emit
    custom(slug) placeholders for snowflake cards

This report splits them so external readers see the honest two numbers:
  - STRUCTURAL: % of cards whose AST is primarily typed engine-executable nodes
  - STUB: % of cards whose AST contains one or more custom(slug) markers

Usage: python3 scripts/coverage_honest.py
"""

from __future__ import annotations

import json
import sys
from collections import Counter
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import parser as p
from mtg_ast import (
    Modification, Static, Activated, Triggered, Keyword,
    Sequence, Choice, Optional_, Conditional, UnknownEffect,
)


ROOT = Path(__file__).resolve().parents[1]
REPORT = ROOT / "data" / "rules" / "coverage_honest.md"


# Modification kinds that are intentional stubs (engine-needs-slug-lookup)
# ── Genuinely opaque stubs ──────────────────────────────────────────
# These kinds mean "the parser recognized the text but couldn't
# decompose it into engine-executable semantics."  A runtime
# engine would need a hand-coded resolver keyed by slug or
# card name to actually play these.
_STUB_KINDS = {
    "custom",                    # per-card handler placeholder — snowflake cards
    "spell_effect",              # fallback: oracle text becomes one opaque Effect
    "ability_word",              # ability-word label whose trigger body failed re-parse
    "saga_chapter",              # chapter text preserved verbatim (engine must interpret)
    "saga_chapter_final_tail",   # final-chapter cleanup text
    "class_level_band",          # class level header text
    "orphan_choice",             # orphaned "choose target X" from modal body
    "modal_header_orphan",       # bare "choose one" header without parsed body
    "modal_bullet_effect",       # modal bullet whose body is opaque
    "modal_bullet",              # modal bullet text
    "inline_modal_with_bullets", # "choose one -- bullet" in single line
    "parsed_tail",               # catch-all residual text fragments
    "perpetually",               # Alchemy-specific perpetual modifier
    "perpetual_mod",             # Alchemy perpetual modification
    "villainous_choice",         # villain's choice mechanic (complex branching)
    "unknown",                   # truly unrecognized text
}

# ── Parameterized / flag-based kinds ───────────────────────────────
# These kinds carry enough semantic information in (kind, args) that
# the engine can dispatch them WITHOUT a hand-coded per-card resolver.
# They are NOT stubs — they are structured data the engine reads
# directly.  Examples:
#   timing_restriction("activate only as a sorcery") → set sorcery timing
#   etb_with_counters(2, "+1/+1") → enter with 2 +1/+1 counters
#   restriction("unblockable") → evasion flag
#   aura_buff(2, 2) → enchanted creature gets +2/+2
#
# If a Modification.kind is NOT in _STUB_KINDS, it is treated as
# structural (engine-executable).  The full list of structural kinds
# is maintained here for documentation:
_STRUCTURAL_MOD_KINDS = {
    # -- Timing / activation constraints --
    "timing_restriction",       # activate only as a sorcery / during X
    "activation_restriction",   # activate only once / only during X
    "once_per_turn",            # do this only once each turn
    "once_per_turn_may",        # once during each of your turns, you may ...
    "trigger_restriction",      # this ability triggers only once
    "cast_restriction",         # cast only during X
    "cast_from_gy",             # cast from graveyard (flashback-like)
    "modes_repeatable",         # choose same mode more than once
    "activation_rights",        # any player may activate
    # -- Combat constraints --
    "restriction",              # unblockable / cant_block / uncounterable
    "combat_restriction",       # can't attack/block
    "must_attack",              # attacks each combat if able
    "extra_block",              # can block additional creature
    "block_only_filter",        # can block only creatures with X
    "immunity",                 # can't be countered/targeted
    # -- ETB / state-based --
    "etb_tapped",               # enters the battlefield tapped
    "etb_with_counters",        # enters with N counters (type parameterized)
    "etb_p1p1_counter",         # enters with a +1/+1 counter
    "enters_prepared",          # enters prepared
    # -- Buff / debuff --
    "aura_buff",                # enchanted creature gets +P/+T
    "conditional_buff_self",    # gets +N/+N as long as X
    "conditional_buff_target",  # target gets +N/+N as long as X
    "conditional_debuff_self",  # gets -N/-N as long as X
    "counter_scoped_anthem",    # each creature with counter has X
    "counter_anthem",           # counter-scoped anthem variant
    # -- Untap / tap manipulation --
    "no_untap",                 # doesn't untap during untap step
    "no_untap_self",            # self doesn't untap
    "aura_no_untap",            # enchanted creature doesn't untap
    "optional_skip_untap",      # may choose not to untap
    "optional_skip_untap_self", # may choose not to untap (self)
    "stun_target",              # doesn't untap during next untap step
    "pronoun_tap",              # tap it
    # -- Cost modification --
    "additional_cost",          # as an additional cost, sacrifice/discard
    "cost_reduction",           # costs {N} less to cast
    "cost_reduce_self",         # this spell costs less
    "variable_cost_reduce",     # costs {X} less where X = ...
    "mana_restriction",         # spend this mana only to ...
    "mana_retention",           # don't lose this mana as phases end
    # -- Keyword / ability grants --
    "self_keyword",             # ~ has [keyword]
    "static_keyword_self",      # self-keyword static
    "pronoun_grant",            # it gains [keyword]
    "ability_grant",            # verbatim ability text grant
    "grant_flash_self",         # may cast as though it had flash
    "exert",                    # may exert as it attacks
    "during_turn_self_static",  # during your turn, has X
    # -- Tribal --
    "tribal_anthem",            # tribal anthem
    "tribal_anthem_get",        # tribal creatures get +P/+T
    "tribal_anthem_have",       # tribal creatures have keyword
    "tribal_anthem_keyword",    # tribal keyword grant
    "tribal_keyword",           # tribal keyword
    "tribal_etb_trig",          # tribal ETB trigger
    "tribal_cost_red",          # tribal cost reduction
    "static_creatures",         # creatures you control have X
    # -- Aura / equipment --
    "aura_keyword",             # aura grants keyword
    "equip_keyword",            # equipment grants keyword
    "aura_trigger",             # aura trigger
    "aura_upkeep_trigger",      # aura upkeep trigger
    "aura_eot_trigger",         # aura end-of-turn trigger
    # -- Library / zone manipulation --
    "library_bottom",           # put on bottom of library
    "fetch_land_tail",          # put land onto battlefield
    "play_those_this_turn",     # play those cards this turn
    "play_exiled_card_this_turn", # play exiled card this turn
    # -- Continuation / chaining --
    "copy_retarget",            # choose new targets for copy
    "chain_copy",               # copy chain effect
    "self_copy_retarget",       # self copy retarget
    "then_if_rider",            # then-if continuation
    "then_clause",              # then-clause continuation
    "when_you_do",              # when-you-do chained trigger
    "when_you_do_p1p1",         # when you do, put +1/+1 counter
    # -- Regen / destroy modifiers --
    "no_regen_tail",            # they can't be regenerated
    "no_regen_tail_it",         # it can't be regenerated
    # -- Control change --
    "temp_control",             # gain control until end of turn
    # -- Duration --
    "until_next_turn",          # until your next turn
    # -- Damage --
    "painland_tail",            # this land deals N damage to you
    # -- Hand size --
    "no_max_hand",              # no maximum hand size
    # -- Type manipulation --
    "still_type",               # it's still a [type]
    "is_also",                  # is also a [type]
    "still_a",                  # still a [type] variant
    "is_every_creature_type",   # is every creature type
    "token_type_def",           # token type definition
    "pronoun_verb",             # it does X (parameterized)
    # -- Conditional statics --
    "conditional_static",       # as long as X / if X (condition text preserved)
    "replacement_static",       # replacement effect (text preserved)
    # -- Delayed triggers --
    "delayed_trigger",          # at beginning of next end step / upkeep
    "delayed_eot_trigger",      # delayed end-of-turn trigger
    # -- Reanimation tails --
    "reanimate_that_card_tail", # return that card to battlefield
    "reanimate_it_tail",        # return it to battlefield
    # -- Named card restrictions --
    "chosen_name_uncastable",   # chosen name can't be cast
    "pithing_needle_chosen",    # chosen name can't activate
    # -- Opponent interaction --
    "opp_choice_card_pick",     # you choose a card from opponent
    # -- Redirect --
    "pariah_redirect",          # redirect damage
    "worship_life_loss",        # life loss prevention
}


def count_effects(effect) -> tuple[int, int]:
    """Walk a (possibly nested) Effect node. Return (structural_count, unknown_count)."""
    if effect is None:
        return (0, 0)
    kind = getattr(effect, "kind", None)
    if kind == "unknown":
        return (0, 1)
    if kind == "sequence":
        s, u = 0, 0
        for item in effect.items:
            ss, uu = count_effects(item)
            s += ss; u += uu
        return (s, u)
    if kind == "choice":
        s, u = 0, 0
        for opt in effect.options:
            ss, uu = count_effects(opt)
            s += ss; u += uu
        return (s, u)
    if kind == "optional":
        return count_effects(effect.body)
    if kind == "conditional":
        s1, u1 = count_effects(effect.body)
        s2, u2 = count_effects(effect.else_body) if effect.else_body else (0, 0)
        return (s1 + s2, u1 + u2)
    if kind is not None:
        return (1, 0)  # typed leaf effect
    return (0, 0)


def classify_card(ast) -> str:
    """Return one of 'structural', 'stub', 'mixed', 'vanilla'.

    - 'vanilla': no abilities (token reminder-only, blank cards, etc.)
    - 'structural': every ability maps to a typed AST node (Damage, Buff, etc.)
      or a keyword. No Modification with a stub kind.
    - 'stub': at least one ability is a Modification with a stub kind,
      OR contains an UnknownEffect.
    - 'mixed': has BOTH typed nodes and stubs.
    """
    if not ast.abilities:
        return "vanilla"

    has_structural = False
    has_stub = False

    for ab in ast.abilities:
        if isinstance(ab, Keyword):
            has_structural = True
            continue
        if isinstance(ab, Static):
            mod = ab.modification
            if mod is None:
                continue
            if mod.kind in _STUB_KINDS:
                has_stub = True
            else:
                has_structural = True
            continue
        if isinstance(ab, Triggered):
            s, u = count_effects(ab.effect)
            if s > 0:
                has_structural = True
            if u > 0:
                has_stub = True
            continue
        if isinstance(ab, Activated):
            s, u = count_effects(ab.effect)
            if s > 0:
                has_structural = True
            if u > 0:
                has_stub = True
            continue

    if has_structural and has_stub:
        return "mixed"
    if has_structural:
        return "structural"
    if has_stub:
        return "stub"
    return "vanilla"


def main():
    p.load_extensions()
    cards = json.loads((ROOT / "data" / "rules" / "oracle-cards.json").read_text())
    real = [c for c in cards if p.is_real_card(c)]

    buckets = Counter()
    per_card_handled = 0

    for c in real:
        ast = p.parse_card(c)
        cls = classify_card(ast)
        buckets[cls] += 1
        if c["name"] in p.PER_CARD_HANDLERS:
            per_card_handled += 1

    total = len(real)
    pct = lambda n: f"{100 * n / total:.2f}%"

    structural = buckets["structural"]
    mixed = buckets["mixed"]
    stub = buckets["stub"]
    vanilla = buckets["vanilla"]

    engine_ready = structural + vanilla  # vanilla is "no work needed, nothing to execute"
    engine_partial = mixed
    engine_stubs = stub

    report = f"""# Honest Coverage Report

**Parser status: 100% GREEN** (every card returns an AST without parse errors).

But GREEN is two things, and the distinction matters for what the runtime engine
will actually be able to execute. This report splits them.

## Three honest numbers

| Category | Cards | % | What it means |
|---|---:|---:|---|
| **Structural** | {structural:,} | {pct(structural)} | Every ability maps to a typed AST node (Damage, Buff, Tutor, Destroy, etc.) that the engine can execute directly. |
| **Mixed** | {mixed:,} | {pct(mixed)} | Some abilities are typed, others are stubs waiting for engine-side custom resolvers. Playable but incomplete. |
| **Stub** | {stub:,} | {pct(stub)} | AST contains only stub Modifications (`custom(slug)` or similar placeholders). Card is recognized; engine needs a hand-coded resolver. |
| **Vanilla** | {vanilla:,} | {pct(vanilla)} | No oracle text (vanilla creatures, tokens with no abilities). Trivially executable. |

## Per-card handler stats

- Per-card handlers in `per_card.py`: **{len(p.PER_CARD_HANDLERS):,}** named cards
- Of those, cards that actually hit the handler (i.e., are in the oracle dump): **{per_card_handled:,}**

Per-card handlers are intentionally emitting stub placeholders for snowflake
cards. They are NOT the same as structural coverage — they're a work queue
for the runtime engine's custom-resolver dispatch.

## The honest framing

- **"100% GREEN" = 100% of cards parse without error.** This is real.
- **"Engine-executable today" = Structural + Vanilla = {engine_ready:,} ({pct(engine_ready)}).**
  For these cards, the AST is fully typed and a runtime interpreter can execute
  them based on the node types alone.
- **"Engine work owed" = Stub + Mixed = {engine_stubs + engine_partial:,} ({pct(engine_stubs + engine_partial)}).**
  These cards parse, but the runtime engine would need custom-resolver code
  keyed by slug or by card name to actually play them.

## What to show externally

When describing this project honestly:

> "The parser reaches syntactic coverage of every printed Magic card ({total:,} cards,
> 100%). Of those, {pct(engine_ready)} produce a fully-typed AST that a runtime
> engine can execute from the node types alone. The remaining {pct(engine_stubs + engine_partial)}
> are recognized but carry stub modifications that will need hand-coded resolvers
> in the engine layer. This is the parser — the runtime engine is the next build."

That framing is both impressive and accurate. "Parsed every magic card" is
legitimately a thing no public FOSS project has cleanly accomplished. But
"can play every magic card" is not yet true, and this report preserves the
distinction.
"""

    REPORT.write_text(report)

    print(f"\n{'═' * 60}")
    print(f"  Honest coverage — {total:,} cards")
    print(f"{'═' * 60}\n")
    print(f"  Structural (typed AST throughout): {structural:>6,}  {pct(structural)}")
    print(f"  Mixed (some typed, some stubs):    {mixed:>6,}  {pct(mixed)}")
    print(f"  Stub (all stubs, needs resolver):  {stub:>6,}  {pct(stub)}")
    print(f"  Vanilla (no oracle text):          {vanilla:>6,}  {pct(vanilla)}")
    print()
    print(f"  Engine-executable today:           {engine_ready:>6,}  {pct(engine_ready)}")
    print(f"  Engine work owed:                  {engine_stubs + engine_partial:>6,}  {pct(engine_stubs + engine_partial)}")
    print()
    print(f"  Per-card handlers: {len(p.PER_CARD_HANDLERS):,} registered, {per_card_handled:,} in oracle pool")
    print(f"\n  → {REPORT}")


if __name__ == "__main__":
    main()
