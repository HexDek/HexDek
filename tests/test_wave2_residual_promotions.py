"""Spot-check tests for Wave 2 residual-wrapper promotions.

The wave-2 post-parse hook in
``scripts/extensions/a_wave2_residual_promotions.py`` runs after wave 1b and
picks up the next-tier residual phrases. These tests exercise one example
from each pattern family so a future regex regression trips a specific signal.
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

_ROOT = Path(__file__).resolve().parents[1]
_SCRIPTS = _ROOT / "scripts"
sys.path.insert(0, str(_SCRIPTS))

from mtg_ast import CardAST, Modification, Static  # noqa: E402

_MODULE_PATH = _SCRIPTS / "extensions" / "a_wave2_residual_promotions.py"
_spec = importlib.util.spec_from_file_location("a_wave2", _MODULE_PATH)
_wave2 = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_wave2)

_hook = _wave2._wave2_post_parse_hook


def _run(raw_kind: str, raw_text: str, *extra_args) -> Modification:
    args = (raw_text,) + tuple(extra_args)
    wrapped = Modification(kind=raw_kind, args=args)
    ability = Static(modification=wrapped, raw="")
    card = CardAST(name="t", abilities=(ability,))
    out = _hook(card)
    return out.abilities[0].modification


# ---------- condition-side (if_intervening_tail args[0]) ----------

def test_controls_basic_land():
    m = _run("if_intervening_tail", "you control a swamp", "x")
    assert m.kind == "if_intervening_tail"
    assert m.args[0].kind == "controls_basic_land"
    assert m.args[0].args == ("swamp",)
    assert m.args[1] == "x"


def test_controls_creature_subtype():
    m = _run("if_intervening_tail", "you control a wizard", "x")
    assert m.args[0].kind == "controls_creature_subtype"
    assert m.args[0].args == ("wizard",)


def test_controls_permanent_type():
    m = _run("if_intervening_tail", "you control an artifact", "x")
    assert m.args[0].kind == "controls_permanent_type"
    assert m.args[0].args == ("artifact",)


def test_controls_count_threshold():
    m = _run("if_intervening_tail", "you control 4 or more creatures", "x")
    assert m.args[0].kind == "controls_count_threshold"
    assert m.args[0].args == (4, "creatures", "or_more")


def test_controls_power_threshold():
    m = _run("if_intervening_tail", "you control a creature with power 4 or greater", "x")
    assert m.args[0].kind == "controls_power_threshold"
    assert m.args[0].args == (4, "or_more")


def test_cast_from_zone_exile():
    m = _run("if_intervening_tail", "this spell was cast from exile", "x")
    assert m.args[0].kind == "cast_mode_check"
    assert m.args[0].args == ("from_exile",)


def test_cast_during_phase_main():
    m = _run("if_intervening_tail", "you cast this spell during your main phase", "x")
    assert m.args[0].kind == "cast_during_phase"
    assert m.args[0].args == ("main",)


def test_colored_mana_spent():
    m = _run("if_intervening_tail", "{B} was spent to cast this spell", "x")
    assert m.args[0].kind == "colored_mana_spent_check"
    assert m.args[0].args == ("B",)


def test_treasure_mana_spent():
    m = _run("if_intervening_tail", "mana from a treasure was spent to cast this spell", "x")
    assert m.args[0].kind == "colored_mana_spent_check"
    assert m.args[0].args == ("treasure",)


def test_state_flag_full_party():
    m = _run("if_intervening_tail", "you have a full party", "x")
    assert m.args[0].kind == "state_flag_check"
    assert m.args[0].args == ("full_party",)


def test_state_flag_completed_dungeon():
    m = _run("if_intervening_tail", "you've completed a dungeon", "x")
    assert m.args[0].kind == "state_flag_check"
    assert m.args[0].args == ("completed_dungeon",)


def test_state_flag_evidence_collected():
    m = _run("if_intervening_tail", "evidence was collected", "x")
    assert m.args[0].args == ("evidence_collected",)


def test_state_flag_gift_promised():
    m = _run("if_intervening_tail", "the gift was promised", "x")
    assert m.args[0].args == ("gift_promised",)


def test_resolve_count_check_second():
    m = _run("if_intervening_tail", "this is the second time this ability has resolved this turn", "x")
    assert m.args[0].kind == "resolve_count_check"
    assert m.args[0].args == (2,)


def test_nth_time_short():
    m = _run("if_intervening_tail", "it's the third time", "x")
    assert m.args[0].kind == "nth_time_short"
    assert m.args[0].args == (3,)


def test_activation_count_check():
    m = _run("if_intervening_tail", "this ability has been activated 4 or more times this turn", "x")
    assert m.args[0].kind == "activation_count_check"
    assert m.args[0].args == (4, "or_more")


def test_creature_event_this_turn_died():
    m = _run("if_intervening_tail", "a creature died this turn", "x")
    assert m.args[0].kind == "creature_event_this_turn"
    assert m.args[0].args == ("died",)


def test_would_event_draw():
    m = _run("if_intervening_tail", "you would draw a card", "x")
    assert m.args[0].kind == "would_event_check"
    assert m.args[0].args == ("draw_card",)


def test_would_event_lose_game():
    m = _run("if_intervening_tail", "you would lose the game", "x")
    assert m.args[0].args == ("lose_game",)


def test_damage_would_be_dealt():
    m = _run("if_intervening_tail", "damage would be dealt to you", "x")
    assert m.args[0].kind == "damage_would_be_dealt"
    assert m.args[0].args == ("you",)


def test_counters_would_be_put():
    m = _run("if_intervening_tail", "one or more +1/+1 counters would be put on a creature you control", "x")
    assert m.args[0].kind == "counters_would_be_put"
    assert m.args[0].args == ("+1/+1", "a_creature_you_control")


def test_card_type_status_legendary():
    m = _run("if_intervening_tail", "it's legendary", "x")
    assert m.args[0].kind == "card_type_check"
    assert m.args[0].args == ("supertype:legendary",)


def test_card_type_status_tapped():
    m = _run("if_intervening_tail", "it's tapped", "x")
    assert m.args[0].args == ("status:tapped",)


def test_card_type_negative():
    m = _run("if_intervening_tail", "it isn't a creature", "x")
    assert m.args[0].kind == "card_type_check_negative"
    assert m.args[0].args == ("a creature",)


def test_no_depletion_counters():
    m = _run("if_intervening_tail", "there are no depletion counters on this land", "x")
    assert m.args[0].kind == "no_depletion_counters"


# ---------- effect-side (parsed_effect_residual / parsed_tail) ----------

def test_animate_token_kv_parse():
    m = _run(
        "parsed_effect_residual",
        "animate:subj=this vehicle;pt=;descr=artifact;type=creature",
    )
    assert m.kind == "animate_subject"
    assert m.args[0] == ("subj", "this vehicle")
    assert m.args[2] == ("descr", "artifact")
    assert m.args[3] == ("type", "creature")


def test_becomes_creature_token():
    m = _run(
        "parsed_effect_residual",
        "becomes_creature:type:the creature type of your choice;pt:",
    )
    assert m.kind == "becomes_creature_subject"
    assert m.args[0] == ("type", "the creature type of your choice")


def test_type_change_token():
    m = _run(
        "parsed_effect_residual",
        "type_change:subj=target land;to=the basic land type of your choice",
    )
    assert m.kind == "type_change_subject"
    assert dict(m.args)["subj"] == "target land"


def test_opp_choose_token():
    m = _run("parsed_effect_residual", "opp_choose:pile")
    assert m.kind == "opp_choose"
    assert m.args == ("pile",)


def test_you_choose_token():
    m = _run("parsed_effect_residual", "you_choose:you choose one of them")
    assert m.kind == "you_choose"


def test_bare_pt_set():
    m = _run("parsed_effect_residual", "bare p/t 4/4")
    assert m.kind == "becomes_pt"
    assert m.args == (4, 4)
    assert m.layer == "7b"


def test_extra_land_per_turn():
    m = _run("parsed_effect_residual", "extra land per turn")
    assert m.kind == "extra_land_per_turn"
    assert m.args == (1,)


def test_may_play_exiled_this_turn():
    m = _run("parsed_effect_residual", "you may play that card this turn")
    assert m.kind == "may_play_exiled"
    assert m.args == ("this_turn",)


def test_may_play_exiled_while_remains():
    m = _run("parsed_effect_residual", "you may play that card for as long as it remains exiled")
    assert m.kind == "may_play_exiled"
    assert m.args == ("while_exiled",)


def test_may_cast_exiled_while_remains():
    m = _run("parsed_effect_residual", "you may cast this card for as long as it remains exiled")
    assert m.kind == "may_cast_exiled"


def test_play_exiled_eot():
    m = _run("parsed_effect_residual", "play exiled cards until end of turn")
    assert m.kind == "play_exiled_cards"
    assert m.args == ("eot",)


def test_end_the_turn():
    m = _run("parsed_effect_residual", "end the turn")
    assert m.kind == "end_the_turn"


def test_skip_next_turn():
    m = _run("parsed_effect_residual", "skip next turn")
    assert m.kind == "skip_turn"
    assert m.args == ("next",)


def test_skip_phase():
    m = _run("parsed_effect_residual", "skip your next draw step")
    assert m.kind == "skip_phase"
    assert m.args == ("draw", "step")


def test_change_target():
    m = _run("parsed_effect_residual", "change target")
    assert m.kind == "change_target"


def test_cant_be_blocked_eot():
    m = _run("parsed_effect_residual", "it can't be blocked this turn")
    assert m.kind == "cant_be_blocked"
    assert m.args == ("eot",)


def test_cant_block_tail():
    m = _run("parsed_tail", "~ can't block")
    assert m.kind == "cant_block"


def test_must_attack_eot():
    m = _run("parsed_effect_residual", "target_creature_must_attack_eot")
    assert m.kind == "must_attack"
    assert m.args == ("eot",)


def test_triggers_once():
    m = _run("parsed_tail", "this ability triggers only once")
    assert m.kind == "triggers_once"


def test_cost_reduction_static():
    m = _run("parsed_tail", "this spell costs {1} less to cast")
    assert m.kind == "cost_reduction_static"
    assert m.args == (1,)


def test_may_pay_amount():
    m = _run("parsed_effect_residual", "you may pay {x}")
    assert m.kind == "may_pay_amount"
    assert m.args == ("x",)


def test_repeat_process():
    m = _run("parsed_effect_residual", "repeat this process")
    assert m.kind == "repeat_process"


def test_experience_counter():
    m = _run("parsed_effect_residual", "you get an experience counter")
    assert m.kind == "experience_counter"
    assert m.args == (1,)


def test_combat_damage_as_unblocked():
    m = _run("parsed_effect_residual",
             "you may have this creature assign its combat damage as though it weren't blocked")
    assert m.kind == "combat_damage_as_unblocked"


def test_counters_var_on_this():
    m = _run("parsed_effect_residual", "counters_var on this creature")
    assert m.kind == "counters_var_on_this"


def test_incubate_n():
    m = _run("parsed_effect_residual", "incubate 2")
    assert m.kind == "incubate"
    assert m.args == (2,)


def test_lose_half_life():
    m = _run("parsed_effect_residual", "that player loses half their life, rounded up")
    assert m.kind == "lose_half_life"
    assert m.args == ("up",)


def test_self_damage_equal_to_power():
    m = _run("parsed_effect_residual", "target creature deals damage to itself equal to its power")
    assert m.kind == "self_damage_equal_to_power"


def test_roll_die_d20():
    m = _run("parsed_effect_residual", "roll a d20")
    assert m.kind == "roll_die"
    assert m.args == (20,)


def test_game_is_draw():
    m = _run("parsed_effect_residual", "the game is a draw")
    assert m.kind == "game_is_draw"


# ---------- safety: untouched kinds pass through ----------

def test_unknown_phrase_passes_through():
    m = _run("parsed_effect_residual", "this is not a known wave-2 phrase whatsoever")
    assert m.kind == "parsed_effect_residual"
    assert m.args[0] == "this is not a known wave-2 phrase whatsoever"
