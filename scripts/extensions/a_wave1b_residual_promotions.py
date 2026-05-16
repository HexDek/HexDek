#!/usr/bin/env python3
"""Wave 1b — promote known phrases sitting inside residual ``Modification``
stubs into typed-discriminator nodes.

Wave 1a (``a_wave1a_promotions.py``) targets phrases entering the parser as
``UnknownEffect`` leaves and rewrites them at parse time. By the time the
post-parse hooks run, ``parser._maybe_promote_unknown`` has already demoted
every surviving ``UnknownEffect`` into one of four catch-all kinds:

  - ``parsed_effect_residual``  (sole-effect leftover, args[0] = raw text)
  - ``parsed_tail``             (tail of a parsed compound effect)
  - ``untyped_effect``          (effect couldn't be categorized)
  - ``if_intervening_tail``     (text after an "if X, ..." condition)

A second post-parse pass can pick up phrases that are still recognizable
inside those wrappers. This module walks the AST one more time looking at
those four wrapper kinds, matches their string arg against a regex table,
and replaces the wrapper with a discriminator-tagged ``Modification``.

Targeted patterns and the audit counts that justified them
(``scripts/audit_unknown_effects.py`` plus a residual-aware drill from the
pre-Wave-1b dataset):

  cast_mode_check    ~135 nodes   "this spell was kicked / bargained / cast
                                  from a graveyard / had its additional cost
                                  paid / had five-or-more mana spent on it"
  state_flag_check    ~40 nodes   "you're the monarch / you become the
                                  monarch / there is no monarch / you have
                                  the city's blessing / you win"
  card_type_check     ~50 nodes   "it's a creature card / it's a land card
                                  / it was a creature card / it's a creature"
  prev_action_resp    ~80 nodes   "the player does/doesn't / a player does
                                  / they do/don't / you do/don't / you can't
                                  / they can't"
  cast_alt_path       ~17 nodes   "you cast a spell this way"
  count_threshold     ~10 nodes   "x is 5 or more"
  roll_table_row      ~87 nodes   plain "roll-table row" (already has a kind
                                  but the parsed_effect_residual wrapper
                                  hides it)
  choose_one_intro     ~9 nodes   bare "choose one"
  regenerate_creature ~12 nodes   bare "regenerate creature"

Estimated total promotions: ~440 stub nodes per corpus pass.

Non-goals:
  - No new AST classes — every promotion stays inside ``Modification``,
    just with a more specific ``kind`` and a small, fixed ``args`` tuple.
  - No grammar changes in ``scripts/parser.py``.
  - No regression on goldens — the walker only fires on residual wrapper
    kinds, never on already-typed effects (Damage, Draw, Buff, etc.).
"""

from __future__ import annotations

import dataclasses
import re
import sys
from pathlib import Path

_HERE = Path(__file__).resolve().parent
_SCRIPTS = _HERE.parent
if str(_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_SCRIPTS))

from mtg_ast import CardAST, Modification, UnknownEffect  # noqa: E402


# Wrapper kinds whose args[0] string we still want to inspect.
RESIDUAL_KINDS = frozenset({
    "parsed_effect_residual",
    "parsed_tail",
    "untyped_effect",
    "if_intervening_tail",
})


_REPLACEMENTS: list[tuple[re.Pattern, callable]] = []


def _phr(pattern: str):
    """Register a residual-string → typed-Modification replacement."""
    def deco(fn):
        _REPLACEMENTS.append((re.compile(pattern, re.I), fn))
        return fn
    return deco


# ---------------------------------------------------------------------------
# Cast-mode conditions (most-common if_intervening_tail family).
# ---------------------------------------------------------------------------

@_phr(r"^this spell was kicked$")
def _ph_cast_kicked(_m):
    return Modification(kind="cast_mode_check", args=("kicked",))


@_phr(r"^this spell was bargained$")
def _ph_cast_bargained(_m):
    return Modification(kind="cast_mode_check", args=("bargained",))


@_phr(r"^this spell was cast from (?:a|your) graveyard$")
def _ph_cast_from_grave(_m):
    return Modification(kind="cast_mode_check", args=("from_graveyard",))


@_phr(r"^this spell'?s additional cost was paid$")
def _ph_cast_additional_paid(_m):
    return Modification(kind="cast_mode_check", args=("additional_cost_paid",))


@_phr(r"^this spell was cast for its (kicker|frenzy|flashback|jump-start|escape|disturb|adventure|spectacle|surge|prowl|emerge|delve) cost$")
def _ph_cast_alt_cost(m):
    return Modification(kind="cast_mode_check", args=(m.group(1).lower(),))


@_phr(r"^(?:you cast (?:this spell|it) using its (kicker|frenzy|flashback|jump-start|escape|disturb|adventure|spectacle|surge|prowl|emerge|delve) (?:cost|ability))$")
def _ph_cast_alt_used(m):
    return Modification(kind="cast_mode_check", args=(m.group(1).lower(),))


@_phr(r"^you cast a spell this way$")
def _ph_cast_alt_path(_m):
    return Modification(kind="cast_alt_path", args=())


@_phr(r"^five or more mana was spent to cast that spell$")
def _ph_cast_five_or_more_mana(_m):
    return Modification(kind="cast_mode_check", args=("five_or_more_mana_spent",))


# ---------------------------------------------------------------------------
# State-flag conditions.
# ---------------------------------------------------------------------------

@_phr(r"^you'?re the monarch$")
def _ph_state_monarch_you(_m):
    return Modification(kind="state_flag_check", args=("monarch",))


@_phr(r"^you become the monarch$")
def _ph_state_become_monarch(_m):
    return Modification(kind="state_flag_check", args=("became_monarch",))


@_phr(r"^there is no monarch$")
def _ph_state_no_monarch(_m):
    return Modification(kind="state_flag_check", args=("no_monarch",))


@_phr(r"^you have the city'?s blessing$")
def _ph_state_city_blessing(_m):
    return Modification(kind="state_flag_check", args=("city_blessing",))


@_phr(r"^you win$")
def _ph_state_win(_m):
    return Modification(kind="state_flag_check", args=("you_win",))


# ---------------------------------------------------------------------------
# Card-type checks used as conditions ("if it's a creature card, ...").
# ---------------------------------------------------------------------------

_CARD_TYPES = (
    "creature card", "land card", "artifact card", "enchantment card",
    "instant card", "sorcery card", "planeswalker card", "battle card",
    "creature", "land", "artifact", "enchantment", "instant", "sorcery",
    "planeswalker", "nonland card", "nonland permanent",
)


@_phr(r"^(?:it'?s|it was|the card is|the card was) ((?:a|an) )?([a-z][a-z ]+?(?: card)?)$")
def _ph_card_type_check(m):
    type_tok = (m.group(2) or "").strip().lower()
    if type_tok in _CARD_TYPES:
        return Modification(kind="card_type_check", args=(type_tok,))
    return None  # decline, no match


# ---------------------------------------------------------------------------
# Pronoun-response "the player does / they don't / you can't" tails.
# ---------------------------------------------------------------------------

@_phr(r"^(the player|a player|that player|each player|they|you|it|that creature)\s+(does(?:n'?t)?|do(?:n'?t)?|did(?:n'?t)?|can'?t|cannot|will(?:n'?t)?|won'?t)$")
def _ph_prev_action_response(m):
    subject = m.group(1).lower().replace(" ", "_")
    verb = m.group(2).lower().replace("'", "")
    # Normalize the polarity.
    negative = verb in ("doesnt", "dont", "didnt", "cant", "cannot", "wont", "willnt")
    return Modification(
        kind="prev_action_response",
        args=(subject, "deny" if negative else "affirm"),
    )


# ---------------------------------------------------------------------------
# Threshold / numeric conditions of the form "X is N or more".
# ---------------------------------------------------------------------------

@_phr(r"^([a-z])\s+is\s+(\d+)\s+or\s+more$")
def _ph_count_threshold(m):
    return Modification(
        kind="count_threshold",
        args=(m.group(1), int(m.group(2)), "or_more"),
    )


@_phr(r"^([a-z])\s+is\s+(\d+)\s+or\s+less$")
def _ph_count_threshold_le(m):
    return Modification(
        kind="count_threshold",
        args=(m.group(1), int(m.group(2)), "or_less"),
    )


# ---------------------------------------------------------------------------
# Wave-1a-stragglers: phrases the post-hook should have caught but missed
# because they're now wrapped in parsed_effect_residual rather than living
# as bare UnknownEffect leaves.
# ---------------------------------------------------------------------------

@_phr(r"^roll-table row$")
def _ph_roll_table(_m):
    return Modification(kind="roll_table_row", args=())


@_phr(r"^choose one$")
def _ph_choose_one(_m):
    return Modification(kind="choose_one_intro", args=())


@_phr(r"^regenerate creature$")
def _ph_regenerate_creature(_m):
    return Modification(kind="regenerate", args=("creature",))


# ============================================================================
# Walker — replace residual-wrapper Modifications via the table above.
# ============================================================================

def _match_residual_string(raw: str):
    """Run the regex table against a raw residual string. Returns the
    builder result (a typed Modification) or None on no match."""
    if not raw:
        return None
    for pat, builder in _REPLACEMENTS:
        m = pat.match(raw)
        if not m:
            continue
        try:
            result = builder(m)
        except Exception:
            continue
        if result is not None:
            return result
    return None


def _try_promote_residual(mod: Modification):
    """If `mod` is a residual-wrapper Modification whose first arg is a
    string we know how to type, return a new Modification appropriately.

    Single-arg wrappers (``parsed_effect_residual``, ``parsed_tail``,
    ``untyped_effect``) are replaced wholesale with the typed Modification.

    Two-arg wrappers (``if_intervening_tail`` — shape ``(condition_str,
    consequent_str)``) are kept intact; only ``args[0]`` is replaced with a
    typed inner Modification, preserving ``args[1]`` so the consequent text
    is never lost. Callers downstream can still inspect ``args[1]`` for the
    consequent string or run their own phase to type it."""
    if mod.kind not in RESIDUAL_KINDS:
        return mod
    args = mod.args or ()
    if not args or not isinstance(args[0], str):
        return mod
    raw = args[0].strip().lower().rstrip(".")
    typed = _match_residual_string(raw)
    if typed is None:
        return mod
    if mod.kind == "if_intervening_tail":
        # Preserve the consequent string at args[1]; only type the condition.
        rest = args[1:] if len(args) > 1 else ()
        return Modification(kind="if_intervening_tail", args=(typed,) + rest, layer=mod.layer)
    # Single-arg residual wrapper: replace wholesale.
    return typed


def _walk_field(val):
    """Walk a dataclass field value, recursing into tuples/lists/dataclasses."""
    if isinstance(val, UnknownEffect):
        return val  # Wave-1a's hook owns these; we don't touch them.
    if isinstance(val, Modification):
        promoted = _try_promote_residual(val)
        if promoted is not val:
            return promoted
        return _walk_replace(val)
    if dataclasses.is_dataclass(val) and not isinstance(val, type):
        return _walk_replace(val)
    if isinstance(val, tuple):
        new_items = tuple(_walk_field(item) for item in val)
        if any(n is not o for n, o in zip(new_items, val)):
            return new_items
        return val
    if isinstance(val, list):
        new_items = [_walk_field(item) for item in val]
        if any(n is not o for n, o in zip(new_items, val)):
            return new_items
        return val
    return val


def _walk_replace(node):
    """Recurse through a frozen dataclass, returning a new instance with
    any residual Modifications inside promoted to typed discriminators."""
    if isinstance(node, Modification):
        promoted = _try_promote_residual(node)
        if promoted is not node:
            return promoted
        # Fall through to walk into its args.
    if not dataclasses.is_dataclass(node) or isinstance(node, type):
        return node
    changes = {}
    for f in dataclasses.fields(node):
        val = getattr(node, f.name)
        new_val = _walk_field(val)
        if new_val is not val:
            changes[f.name] = new_val
    if not changes:
        return node
    kwargs = {}
    for f in dataclasses.fields(node):
        if not f.init:
            continue
        kwargs[f.name] = changes.get(f.name, getattr(node, f.name))
    try:
        return type(node)(**kwargs)
    except Exception:
        return node


def _wave1b_post_parse_hook(card_ast: CardAST) -> CardAST:
    """Walk the CardAST and promote residual-wrapper Modifications."""
    new_abilities = []
    changed = False
    for ability in card_ast.abilities:
        new_ab = _walk_replace(ability)
        new_abilities.append(new_ab)
        if new_ab is not ability:
            changed = True
    if not changed:
        return card_ast
    kwargs = {}
    for f in dataclasses.fields(card_ast):
        if not f.init:
            continue
        if f.name == "abilities":
            kwargs[f.name] = tuple(new_abilities)
        else:
            kwargs[f.name] = getattr(card_ast, f.name)
    try:
        return CardAST(**kwargs)
    except Exception:
        return card_ast


POST_PARSE_HOOKS = [_wave1b_post_parse_hook]
