"""Spot-check tests for Wave 3 residual-wrapper promotions.

Wave 3 (``scripts/extensions/a_wave3_residual_promotions.py``) runs after
waves 1b and 2 and picks up next-tier residual phrases. These tests cover
one example from each pattern family.
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

_ROOT = Path(__file__).resolve().parents[1]
_SCRIPTS = _ROOT / "scripts"
sys.path.insert(0, str(_SCRIPTS))

from mtg_ast import CardAST, Modification, Static  # noqa: E402

_MODULE_PATH = _SCRIPTS / "extensions" / "a_wave3_residual_promotions.py"
_spec = importlib.util.spec_from_file_location("a_wave3", _MODULE_PATH)
_wave3 = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_wave3)

_hook = _wave3._wave3_post_parse_hook


def _run(raw_kind: str, raw_text: str, *extra_args) -> Modification:
    args = (raw_text,) + tuple(extra_args)
    wrapped = Modification(kind=raw_kind, args=args)
    ability = Static(modification=wrapped, raw="")
    card = CardAST(name="t", abilities=(ability,))
    out = _hook(card)
    return out.abilities[0].modification


# ---------- condition-side ----------

def test_this_way_chain_damage_prevented():
    m = _run("if_intervening_tail", "damage is prevented this way", "x")
    assert m.args[0].kind == "this_way_chain"
    assert m.args[0].args == ("damage_prevented",)


def test_this_way_chain_creature_destroyed():
    m = _run("if_intervening_tail", "a creature destroyed this way", "x")
    assert m.args[0].args == ("creature_destroyed",)


def test_cast_sorcery_speed():
    m = _run("if_intervening_tail", "you cast it any time a sorcery couldn't have been cast", "x")
    assert m.args[0].kind == "cast_timing_window"
    assert m.args[0].args == ("sorcery_speed",)


def test_cast_mode_foretold():
    m = _run("if_intervening_tail", "this spell was foretold", "x")
    assert m.args[0].kind == "cast_mode_check"
    assert m.args[0].args == ("foretold",)


def test_cast_mode_madness_paid():
    m = _run("if_intervening_tail", "this spell's madness cost was paid", "x")
    assert m.args[0].args == ("madness_cost_paid",)


def test_cast_revealed_card():
    m = _run("if_intervening_tail",
             "you revealed a dragon card or controlled a dragon as you cast this spell", "x")
    assert m.args[0].kind == "cast_revealed_or_controlled"
    assert m.args[0].args == ("dragon",)


def test_controls_count_spelled():
    m = _run("if_intervening_tail", "you control four or more creatures", "x")
    assert m.args[0].kind == "controls_count_threshold"
    assert m.args[0].args == (4, "creatures", "or_more")


def test_activated_count_spelled():
    m = _run("if_intervening_tail",
             "this ability has been activated four or more times this turn", "x")
    assert m.args[0].kind == "activation_count_check"
    assert m.args[0].args == (4, "or_more")


def test_this_would_zone_change_die():
    m = _run("if_intervening_tail", "~ would be put into a graveyard from anywhere", "x")
    assert m.args[0].kind == "this_would_zone_change"
    assert m.args[0].args == ("die_anywhere",)


def test_this_creature_would_die():
    m = _run("if_intervening_tail", "this creature would be destroyed", "x")
    assert m.args[0].args == ("destroyed",)


def test_card_would_be_milled_this_turn():
    m = _run("if_intervening_tail", "a card would be put into your graveyard from anywhere this turn", "x")
    assert m.args[0].kind == "card_would_be_put_in_graveyard"
    assert m.args[0].args == ("this_turn",)


def test_would_create_tokens():
    m = _run("if_intervening_tail", "an effect would create one or more tokens under your control", "x")
    assert m.args[0].kind == "would_create_tokens"


def test_forced_discard():
    m = _run("if_intervening_tail",
             "a spell or ability an opponent controls causes you to discard this card", "x")
    assert m.args[0].kind == "forced_discard_by_opponent"


def test_mana_spent_on_type():
    m = _run("if_intervening_tail", "that mana is spent on a creature spell", "x")
    assert m.args[0].kind == "mana_spent_on_type"
    assert m.args[0].args == ("a creature spell",)


def test_chosen_name_match():
    m = _run("if_intervening_tail", "that card has the chosen name", "x")
    assert m.args[0].kind == "chosen_name_match"


def test_damage_to_self_with_counter():
    m = _run("if_intervening_tail",
             "damage would be dealt to this creature while it has a +1/+1 counter on it", "x")
    assert m.args[0].kind == "damage_to_self_with_counter"
    assert m.args[0].args == ("+1/+1",)


def test_creature_has_keyword():
    m = _run("if_intervening_tail", "that creature has toxic", "x")
    assert m.args[0].kind == "creature_has_keyword"
    assert m.args[0].args == ("toxic",)


def test_doesnt_have_keyword():
    m = _run("if_intervening_tail", "it doesn't have suspend", "x")
    assert m.args[0].kind == "doesnt_have_keyword"
    assert m.args[0].args == ("suspend",)


def test_gift_not_promised():
    m = _run("if_intervening_tail", "the gift wasn't promised", "x")
    assert m.args[0].kind == "state_flag_check_negative"
    assert m.args[0].args == ("gift_promised",)


def test_card_property_instant_sorcery():
    m = _run("if_intervening_tail", "it's an instant or sorcery card", "x")
    assert m.args[0].kind == "card_type_check"
    assert m.args[0].args == ("type:instant_or_sorcery",)


def test_card_type_negative_land_card():
    m = _run("if_intervening_tail", "it isn't a land card", "x")
    assert m.args[0].kind == "card_type_check_negative"
    assert m.args[0].args == ("a land card",)


def test_land_tapped_for_mana():
    m = _run("if_intervening_tail", "a land is tapped for mana", "x")
    assert m.args[0].kind == "land_tapped_for_mana"


def test_land_was_nonbasic():
    m = _run("if_intervening_tail", "that land was nonbasic", "x")
    assert m.args[0].kind == "land_was_nonbasic"


def test_player_would_draw():
    m = _run("if_intervening_tail", "a player would draw a card", "x")
    assert m.args[0].args == ("any_player_draw_card",)


def test_would_lose_mana():
    m = _run("if_intervening_tail", "you would lose unspent mana", "x")
    assert m.args[0].args == ("lose_unspent_mana",)


def test_source_would_deal_damage():
    m = _run("if_intervening_tail", "a source you control would deal damage to a permanent or player", "x")
    assert m.args[0].kind == "source_would_deal_damage"
    assert m.args[0].args == ("a_permanent_or_player",)


def test_nontoken_opp_would_die():
    m = _run("if_intervening_tail", "a nontoken creature an opponent controls would die", "x")
    assert m.args[0].kind == "would_die_filter"
    assert m.args[0].args == ("nontoken_opp",)


def test_perm_dealt_would_die():
    m = _run("if_intervening_tail", "a permanent dealt damage this way would die this turn", "x")
    assert m.args[0].args == ("perm_dealt_this_way",)


# ---------- effect-side tokens ----------

def test_add_mana_per_token():
    m = _run("parsed_effect_residual", "add_{g}_per:creature you control")
    assert m.kind == "add_mana_per"
    assert m.args == ("{G}", "creature you control")


def test_choose_token_simple():
    m = _run("parsed_effect_residual", "choose:color")
    assert m.kind == "choose"
    assert m.args == ("color",)


def test_choose_token_typed():
    m = _run("parsed_effect_residual", "choose:type:creature")
    assert m.kind == "choose"
    assert m.args == ("type", "creature")


def test_choose_khans_or_dragons():
    m = _run("parsed_effect_residual", "choose khans or dragons")
    assert m.kind == "choose"
    assert m.args == ("faction", "khans_or_dragons")


def test_become_basic_token():
    m = _run("parsed_effect_residual", "become_basic:subj=target land;type=island")
    assert m.kind == "become_basic_land"
    assert m.args[0] == ("subj", "target land")
    assert m.args[1] == ("type", "island")


def test_opp_chooses_n_token():
    m = _run("parsed_effect_residual", "opponent_chooses:2")
    assert m.kind == "opp_chooses_n"
    assert m.args == (2,)


# ---------- static effects ----------

def test_creatures_gain_keyword_eot():
    m = _run("parsed_effect_residual", "creatures you control gain indestructible until end of turn")
    assert m.kind == "creatures_gain_keyword_eot"
    assert m.args == ("indestructible",)
    assert m.layer == "6"


def test_creatures_gain_two_keywords_eot():
    m = _run("parsed_effect_residual", "creatures you control gain flying and lifelink until end of turn")
    assert m.kind == "creatures_gain_keyword_eot"
    assert m.args == ("flying", "lifelink")


def test_all_creatures_have_keyword():
    m = _run("parsed_tail", "all creatures have haste")
    assert m.kind == "all_creatures_have_keyword"
    assert m.args == ("haste",)


def test_permanents_have_keyword_artifacts():
    m = _run("parsed_tail", "artifacts you control have hexproof")
    assert m.kind == "permanents_have_keyword"
    assert m.args == ("artifacts", "hexproof")


def test_opp_creatures_enter_tapped():
    m = _run("parsed_tail", "creatures your opponents control enter tapped")
    assert m.kind == "opp_creatures_enter_tapped"


def test_lands_enter_untapped():
    m = _run("parsed_tail", "lands you control enter untapped")
    assert m.kind == "lands_enter_untapped"


def test_each_player_one_spell():
    m = _run("parsed_tail", "each player can't cast more than one spell each turn")
    assert m.kind == "cast_limit_one_per_turn"
    assert m.args == ("each_player",)


def test_you_one_spell():
    m = _run("parsed_tail", "you can't cast more than one spell each turn")
    assert m.args == ("you",)


def test_chosen_type_added():
    m = _run("parsed_tail", "this creature is the chosen type in addition to its other types")
    assert m.kind == "chosen_type_added"
    assert m.layer == "4"


def test_ali_from_cairo():
    m = _run("parsed_tail",
             "damage that would reduce your life total to less than 1 reduces it to 1 instead")
    assert m.kind == "cant_lose_life_below"
    assert m.args == (1,)


def test_this_is_all_colors():
    m = _run("parsed_tail", "~ is all colors")
    assert m.kind == "this_is_all_colors"
    assert m.layer == "5"


def test_enters_prepared():
    m = _run("parsed_tail", "~ enters prepared")
    assert m.kind == "enters_prepared"


# ---------- zone returns / sacrifices / cant-be-X ----------

def test_return_at_eoc():
    m = _run("parsed_effect_residual", "return that creature to its owner's hand at end of combat")
    assert m.kind == "return_to_hand_at"
    assert m.args == ("end_of_combat",)


def test_destroy_at_eoc():
    m = _run("parsed_effect_residual", "destroy it (and this) at end of combat")
    assert m.kind == "destroy_at"
    assert m.args == ("end_of_combat",)


def test_damage_cant_be_prevented():
    m = _run("parsed_effect_residual", "the damage can't be prevented")
    assert m.kind == "damage_cant_be_prevented"


def test_cant_regen_this_way():
    m = _run("parsed_effect_residual", "a creature destroyed this way can't be regenerated")
    assert m.kind == "cant_regenerate_this_way"


def test_counters_var_on_it():
    m = _run("parsed_effect_residual", "counters_var on it")
    assert m.kind == "counters_var_on_this"
    assert m.args == ("it",)


def test_pile_split():
    m = _run("parsed_effect_residual", "put one pile into your hand and the other into your graveyard")
    assert m.kind == "pile_split_distribution"
    assert m.args == ("hand_vs_graveyard",)


def test_opp_separates():
    m = _run("parsed_effect_residual", "opp separates piles")
    assert m.kind == "opp_separates_piles"


def test_may_cast_copy():
    m = _run("parsed_effect_residual", "you may cast the copy")
    assert m.kind == "may_cast_copy"


def test_sacrifice_those_tokens():
    m = _run("parsed_effect_residual", "sacrifice those tokens")
    assert m.kind == "sacrifice_those_tokens"


def test_cost_reduction_var():
    m = _run("parsed_tail",
             "this spell costs {x} less to cast, where x is the greatest power among creatures you control")
    assert m.kind == "cost_reduction_var"
    assert m.args[0] == "x"
    assert "greatest power" in m.args[1]


def test_may_play_no_cost_eot():
    m = _run("parsed_effect_residual",
             "until end of turn, you may play that card without paying its mana cost")
    assert m.kind == "may_play_no_cost"
    assert m.args == ("eot",)


def test_may_play_until_next_turn():
    m = _run("parsed_effect_residual", "you may play that card until the end of your next turn")
    assert m.kind == "may_play_exiled"
    assert m.args == ("until_end_of_next_turn",)


def test_lose_life_per():
    m = _run("parsed_effect_residual", "you lose 2 life for each x")
    assert m.kind == "lose_life_per"
    assert m.args == (2, "x")


# ---------- safety ----------

def test_unknown_phrase_passes_through():
    m = _run("parsed_effect_residual", "this is not a wave 3 phrase at all")
    assert m.kind == "parsed_effect_residual"
