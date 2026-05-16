"""Spot-check tests for Wave 1b residual-wrapper promotions.

The wave-1b post-parse hook in
``scripts/extensions/a_wave1b_residual_promotions.py`` walks ``Modification``
leaves whose ``kind`` is one of the residual-wrapper kinds
(``parsed_effect_residual``, ``parsed_tail``, ``untyped_effect``,
``if_intervening_tail``) and, when ``args[0]`` matches a known phrase,
replaces it with a typed discriminator Modification.

These tests build small synthetic ASTs and assert the hook fires correctly,
so a future regression in the walker (e.g. broken regex, lost args[1] on
two-arg wrappers) trips a specific red signal here.
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

_ROOT = Path(__file__).resolve().parents[1]
_SCRIPTS = _ROOT / "scripts"
sys.path.insert(0, str(_SCRIPTS))

from mtg_ast import CardAST, Modification, Static  # noqa: E402

_MODULE_PATH = _SCRIPTS / "extensions" / "a_wave1b_residual_promotions.py"
_spec = importlib.util.spec_from_file_location("a_wave1b", _MODULE_PATH)
_wave1b = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_wave1b)

_hook = _wave1b._wave1b_post_parse_hook


def _run(raw_kind: str, raw_text: str, *extra_args) -> Modification:
    """Wrap raw_text in a residual Modification, run the hook, return the
    (possibly promoted) effect Modification on the only ability."""
    args = (raw_text,) + tuple(extra_args)
    wrapped = Modification(kind=raw_kind, args=args)
    ability = Static(modification=wrapped, raw="")
    card = CardAST(name="t", abilities=(ability,))
    out = _hook(card)
    return out.abilities[0].modification


def test_cast_mode_check_kicked():
    m = _run("if_intervening_tail", "this spell was kicked", "draw a card")
    # if_intervening_tail wrapper preserved, args[0] now typed, args[1] preserved
    assert m.kind == "if_intervening_tail"
    assert isinstance(m.args[0], Modification)
    assert m.args[0].kind == "cast_mode_check"
    assert m.args[0].args == ("kicked",)
    assert m.args[1] == "draw a card"


def test_cast_mode_check_bargained():
    m = _run("if_intervening_tail", "this spell was bargained", "consequent")
    assert m.args[0].kind == "cast_mode_check"
    assert m.args[0].args == ("bargained",)


def test_cast_mode_check_from_graveyard():
    m = _run("if_intervening_tail", "this spell was cast from a graveyard", "x")
    assert m.args[0].kind == "cast_mode_check"
    assert m.args[0].args == ("from_graveyard",)


def test_cast_mode_check_additional_cost_paid():
    m = _run("if_intervening_tail", "this spell's additional cost was paid", "x")
    assert m.args[0].kind == "cast_mode_check"
    assert m.args[0].args == ("additional_cost_paid",)


def test_state_flag_monarch():
    m = _run("if_intervening_tail", "you're the monarch", "x")
    assert m.args[0].kind == "state_flag_check"
    assert m.args[0].args == ("monarch",)


def test_state_flag_city_blessing():
    m = _run("if_intervening_tail", "you have the city's blessing", "x")
    assert m.args[0].kind == "state_flag_check"
    assert m.args[0].args == ("city_blessing",)


def test_state_flag_you_win():
    m = _run("if_intervening_tail", "you win", "x")
    assert m.args[0].kind == "state_flag_check"
    assert m.args[0].args == ("you_win",)


def test_card_type_check_creature_card():
    m = _run("if_intervening_tail", "it's a creature card", "x")
    assert m.args[0].kind == "card_type_check"
    assert m.args[0].args == ("creature card",)


def test_card_type_check_land_card():
    m = _run("if_intervening_tail", "it's a land card", "x")
    assert m.args[0].kind == "card_type_check"
    assert m.args[0].args == ("land card",)


def test_prev_action_affirm():
    m = _run("if_intervening_tail", "the player does", "x")
    assert m.args[0].kind == "prev_action_response"
    assert m.args[0].args == ("the_player", "affirm")


def test_prev_action_deny():
    m = _run("if_intervening_tail", "the player doesn't", "x")
    assert m.args[0].kind == "prev_action_response"
    assert m.args[0].args == ("the_player", "deny")


def test_prev_action_you_cant():
    m = _run("if_intervening_tail", "you can't", "x")
    assert m.args[0].kind == "prev_action_response"
    assert m.args[0].args == ("you", "deny")


def test_count_threshold_or_more():
    m = _run("if_intervening_tail", "x is 5 or more", "do thing")
    assert m.args[0].kind == "count_threshold"
    assert m.args[0].args == ("x", 5, "or_more")


def test_cast_alt_path():
    m = _run("if_intervening_tail", "you cast a spell this way", "x")
    assert m.args[0].kind == "cast_alt_path"


def test_roll_table_row_single_arg():
    # Single-arg residual: whole wrapper is replaced with typed kind.
    m = _run("parsed_effect_residual", "roll-table row")
    assert m.kind == "roll_table_row"


def test_choose_one_intro_single_arg():
    m = _run("parsed_effect_residual", "choose one")
    assert m.kind == "choose_one_intro"


def test_no_match_passes_through():
    # A phrase not in our table should leave the wrapper untouched.
    m = _run("parsed_effect_residual", "some unrelated effect text")
    assert m.kind == "parsed_effect_residual"
    assert m.args == ("some unrelated effect text",)


def test_typed_modifications_not_touched():
    # An already-typed Modification (kind not in RESIDUAL_KINDS) must not
    # be matched/replaced even if its args[0] happens to look matchable.
    m = _run("draw", "this spell was kicked")  # 'draw' is not residual
    assert m.kind == "draw"
    assert m.args == ("this spell was kicked",)
