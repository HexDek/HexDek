#!/usr/bin/env python3
"""Wave 3 — third-tier promotion of residual-wrapper ``Modification`` stubs.

Wave 2 (``a_wave2_residual_promotions.py``) typed ~700 stub nodes via a
39-pattern table covering "you control a X" conditions, state-flag tails,
"would" events, structured tokens (``animate:`` / ``becomes_creature:`` /
``type_change:`` / ``opp_choose:`` / ``you_choose:``), and a long
effect-side list (``bare p/t`` / ``end the turn`` / ``incubate X`` etc).

Wave 3 picks up the next tier — the families a residual-aware audit of
the post-wave-2 corpus surfaces. It runs as the third post-parse hook
(``a_wave3_…`` sorts after ``a_wave2_…``), so it only sees residuals
the prior two waves couldn't type.

Targeted families and audited counts (post-wave-2 corpus, 31,963 cards):

  Condition-side (mostly ``if_intervening_tail`` args[0]):
    this_way_chain              38 nodes   "damage is prevented this way",
                                           "that spell is countered this way",
                                           "a creature card was/is exiled
                                           this way", "it unspecializes this
                                           way", "you didn't put a card into
                                           your hand this way", "a permanent's
                                           ability is countered this way",
                                           "a player is dealt damage this way",
                                           "a creature destroyed this way"
    controls_count_spelled      18 nodes   "you control {four|eight|...} or
                                           more {creatures|lands|...}" — the
                                           spelled-out twin of wave-2's
                                           controls_count_threshold (which
                                           only handled \\d+ digits)
    cast_sorcery_speed          10 nodes   "you cast it any time a sorcery
                                           couldn't have been cast"
    would_zone_change            8 nodes   "~ would be put into a graveyard
                                           from anywhere" / "would be
                                           destroyed" / "would be exiled"
    opp_chooses_n_token          7 nodes   pre-encoded "opponent_chooses:N"
    would_create_tokens          6 nodes   "an effect would create one or
                                           more tokens under your control"
    card_would_be_milled         6 nodes   "a card would be put into your
                                           graveyard from anywhere [this turn]"
    damage_to_self_with_counter  5 nodes   "damage would be dealt to this
                                           creature while it has a +1/+1
                                           counter on it"
    forced_discard               5 nodes   "a spell or ability an opponent
                                           controls causes you to discard
                                           this card"
    mana_spent_on_type           5 nodes   "that mana is spent on a creature
                                           spell"
    creature_has_keyword         5 nodes   "that creature has toxic"
    activated_spelled            4 nodes   "this ability has been activated
                                           {four|...} or more times this turn"
                                           — extends wave-2 (digits only)
    cast_revealed_card           4 nodes   "you revealed a dragon card or
                                           controlled a dragon as you cast
                                           this spell"
    chosen_name_match            4 nodes   "that card has the chosen name"
    cast_mode_alt_cost_paid      4 nodes   "this spell's madness/delve/...
                                           cost was paid" — extends wave-1b
                                           cast_mode_check
    doesnt_have_keyword          4 nodes   "it doesn't have suspend"
    cast_mode_foretold           3 nodes   "this spell was foretold"
    card_type_negative_ext       3 nodes   "it isn't a land card" — extends
                                           wave-2 card_type_check_negative
    card_property_instant_sorc   3 nodes   "it's an instant or sorcery card"
    gift_not_promised            3 nodes   "the gift wasn't promised"
    land_tapped_for_mana         3 nodes   "a land is tapped for mana"
    land_was_nonbasic            3 nodes   "that land was nonbasic"
    player_would_draw            3 nodes   "a player would draw a card"
    would_lose_mana              3 nodes   "you would lose unspent mana"
    source_would_deal_damage     4 nodes   "a source you control would deal
                                           damage to a permanent or player"
    this_creature_would_die      3 nodes   "this creature would be destroyed"
    nontoken_opp_would_die       3 nodes   "a nontoken creature an opponent
                                           controls would die"
    perm_dealt_would_die         3 nodes   "a permanent dealt damage this way
                                           would die this turn"

  Effect-side (parsed_effect_residual, parsed_tail):
    add_mana_per_token          58 nodes   "add_{G}_per:creature you control"
    creatures_gain_keyword_eot  31 nodes   "creatures you control gain
                                           indestructible until end of turn"
    choose_token                27 nodes   "choose:color" / "choose:number" /
                                           "choose:type:creature" /
                                           "choose:direction"
    cost_reduction_var          21 nodes   "this spell costs {X} less to
                                           cast, where x is …"
    permanents_have_keyword     15 nodes   "artifacts you control have
                                           hexproof" / "other permanents …
                                           have hexproof"
    pile_split                  12 nodes   "put one/that pile into your hand
                                           and the other into your graveyard"
    become_basic_token          12 nodes   "become_basic:subj=target land;
                                           type=island"
    lose_life_per               11 nodes   "you lose 2 life for each X"
    opp_separates                6 nodes   "opp separates piles"
    counters_var_on              6 nodes   "counters_var on it" — extends
                                           wave-2 counters_var_on_this
    all_creatures_have_keyword   6 nodes   "all creatures have haste"
    cant_regen_this_way          6 nodes   "a creature destroyed this way
                                           can't be regenerated"
    destroy_at_eoc               5 nodes   "destroy it (and this) at end of
                                           combat"
    sacrifice_those_tokens       5 nodes   bare "sacrifice those tokens"
    may_cast_copy                5 nodes   "you may cast the copy"
    each_player_one_spell        5 nodes   "each player can't cast more than
                                           one spell each turn"
    may_play_no_cost_eot         4 nodes   "until end of turn, you may play
                                           that card without paying its mana
                                           cost"
    may_play_until_next          4 nodes   "you may play that card until the
                                           end of your next turn"
    damage_cant_be_prevented     4 nodes   "the damage can't be prevented"
    chosen_type_added            4 nodes   "this creature is the chosen type
                                           in addition to its other types"
    opp_creatures_enter_tapped   4 nodes   "creatures your opponents control
                                           enter tapped"
    return_at_eoc                4 nodes   "return that/it/them to its/their
                                           owner's hand at end of combat"
    this_is_all_colors           4 nodes   "~ is all colors"
    you_one_spell                3 nodes   "you can't cast more than one
                                           spell each turn"
    lands_enter_untapped         3 nodes   "lands you control enter untapped"
    ali_from_cairo               3 nodes   "damage that would reduce your
                                           life total to less than 1 reduces
                                           it to 1 instead"
    enters_prepared              3 nodes   "~ enters prepared"

Estimated total promotions: ~440 stub nodes per corpus pass.

Non-goals (same as waves 1b / 2):
  - No new AST classes; promotions stay inside ``Modification``.
  - No grammar changes in ``scripts/parser.py``.
  - No regression on already-typed nodes — walker only fires on the four
    residual wrapper kinds.
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


_SPELLED_NUM = {
    "one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
    "six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
}


# ---------------------------------------------------------------------------
# "...this way" effect-chain conditions.
# ---------------------------------------------------------------------------

_THIS_WAY_SUBJECTS = {
    "damage is prevented": "damage_prevented",
    "that spell is countered": "spell_countered",
    "a creature card was exiled": "creature_card_exiled",
    "a creature card is exiled": "creature_card_exiled",
    "it unspecializes": "unspecialized",
    "you didn't put a card into your hand": "card_not_drawn",
    "a permanent's ability is countered": "ability_countered",
    "a player is dealt damage": "player_dealt_damage",
    "a creature destroyed": "creature_destroyed",
    "a card was milled": "card_milled",
    "a card is exiled": "card_exiled",
    "an opponent draws a card": "opponent_drew_card",
    "life is lost": "life_lost",
    "life is gained": "life_gained",
}


@_phr(r"^(.+) this way$")
def _ph_this_way_chain(m):
    subj = m.group(1).strip().lower().rstrip(",")
    key = _THIS_WAY_SUBJECTS.get(subj)
    if key is None:
        return None
    return Modification(kind="this_way_chain", args=(key,))


# ---------------------------------------------------------------------------
# Cast-timing window + cast-mode extensions (extend wave 1b cast_mode_check).
# ---------------------------------------------------------------------------

@_phr(r"^you cast it any time a sorcery couldn'?t have been cast$")
def _ph_cast_sorcery_speed(_m):
    return Modification(kind="cast_timing_window", args=("sorcery_speed",))


@_phr(r"^this spell was foretold$")
def _ph_cast_mode_foretold(_m):
    return Modification(kind="cast_mode_check", args=("foretold",))


@_phr(r"^this spell'?s (madness|delve|escape|jump-start|flashback|kicker|surge|adventure|disturb|prowl|emerge|spectacle|frenzy) cost was paid$")
def _ph_cast_mode_alt_cost_paid(m):
    return Modification(
        kind="cast_mode_check",
        args=(f"{m.group(1).lower()}_cost_paid",),
    )


@_phr(r"^you revealed an? ([a-z][a-z0-9 ]*?) card or controlled an? \1 as you cast this spell$")
def _ph_cast_revealed_card(m):
    return Modification(
        kind="cast_revealed_or_controlled",
        args=(m.group(1).strip().lower(),),
    )


# ---------------------------------------------------------------------------
# Spelled-out "you control N or more X" — wave-2 only handled \d+ digits.
# ---------------------------------------------------------------------------

@_phr(r"^you control (one|two|three|four|five|six|seven|eight|nine|ten) or more ([a-z][a-z0-9\- ]+?)$")
def _ph_controls_count_spelled(m):
    return Modification(
        kind="controls_count_threshold",
        args=(_SPELLED_NUM[m.group(1).lower()], m.group(2).strip().lower(), "or_more"),
    )


@_phr(r"^this ability has been activated (one|two|three|four|five|six|seven|eight|nine|ten) or more times this turn$")
def _ph_activated_spelled(m):
    return Modification(
        kind="activation_count_check",
        args=(_SPELLED_NUM[m.group(1).lower()], "or_more"),
    )


# ---------------------------------------------------------------------------
# "Would" zone-change conditions.
# ---------------------------------------------------------------------------

@_phr(r"^~ would be (put into a graveyard from anywhere|destroyed|exiled)$")
def _ph_this_would_zone_change(m):
    ev = m.group(1).lower()
    key = {
        "put into a graveyard from anywhere": "die_anywhere",
        "destroyed": "destroyed",
        "exiled": "exiled",
    }[ev]
    return Modification(kind="this_would_zone_change", args=(key,))


@_phr(r"^this creature would be (destroyed|exiled|put into a graveyard)$")
def _ph_this_creature_would_die(m):
    return Modification(kind="this_would_zone_change", args=(m.group(1).lower().replace(" ", "_"),))


@_phr(r"^a nontoken creature an opponent controls would die$")
def _ph_nontoken_opp_would_die(_m):
    return Modification(kind="would_die_filter", args=("nontoken_opp",))


@_phr(r"^a permanent dealt damage this way would die this turn$")
def _ph_perm_dealt_would_die(_m):
    return Modification(kind="would_die_filter", args=("perm_dealt_this_way",))


@_phr(r"^a card would be put into your graveyard from anywhere(?: this turn)?$")
def _ph_card_would_be_milled(m):
    this_turn = "this turn" in m.group(0).lower()
    return Modification(
        kind="card_would_be_put_in_graveyard",
        args=(("this_turn",) if this_turn else ("any_time",)),
    )


@_phr(r"^(?:an effect would create one or more tokens under your control|one or more (?:creature )?tokens would be created under your control)$")
def _ph_would_create_tokens(_m):
    return Modification(kind="would_create_tokens", args=("under_your_control",))


@_phr(r"^a player would draw a card$")
def _ph_player_would_draw(_m):
    return Modification(kind="would_event_check", args=("any_player_draw_card",))


@_phr(r"^you would lose unspent mana$")
def _ph_would_lose_mana(_m):
    return Modification(kind="would_event_check", args=("lose_unspent_mana",))


@_phr(r"^a source you control would deal damage to (a permanent or player|a creature|an opponent|any target)$")
def _ph_source_would_deal_damage(m):
    return Modification(
        kind="source_would_deal_damage",
        args=(m.group(1).lower().replace(" ", "_"),),
    )


@_phr(r"^damage would be dealt to this creature while it has a (\+1/\+1|\-1/\-1) counter on it$")
def _ph_damage_to_self_with_counter(m):
    return Modification(
        kind="damage_to_self_with_counter",
        args=(m.group(1).lower(),),
    )


# ---------------------------------------------------------------------------
# Misc condition extensions.
# ---------------------------------------------------------------------------

@_phr(r"^a spell or ability an opponent controls causes you to discard this card$")
def _ph_forced_discard(_m):
    return Modification(kind="forced_discard_by_opponent", args=())


@_phr(r"^that mana is spent on (a creature spell|a noncreature spell|an instant or sorcery spell|an? [a-z]+ spell)$")
def _ph_mana_spent_on_type(m):
    target = m.group(1).lower()
    return Modification(kind="mana_spent_on_type", args=(target,))


@_phr(r"^that card has the chosen name$")
def _ph_chosen_name_match(_m):
    return Modification(kind="chosen_name_match", args=())


@_phr(r"^that creature has ([a-z]+)$")
def _ph_creature_has_keyword(m):
    return Modification(kind="creature_has_keyword", args=(m.group(1).lower(),))


@_phr(r"^it (?:doesn'?t|does not) have ([a-z]+)$")
def _ph_doesnt_have_keyword(m):
    return Modification(kind="doesnt_have_keyword", args=(m.group(1).lower(),))


@_phr(r"^the gift wasn'?t promised$")
def _ph_gift_not_promised(_m):
    return Modification(kind="state_flag_check_negative", args=("gift_promised",))


@_phr(r"^it'?s an instant or sorcery card$")
def _ph_card_property_instant_sorcery(_m):
    return Modification(kind="card_type_check", args=("type:instant_or_sorcery",))


@_phr(r"^it isn'?t (a land card|a creature card|a planeswalker)$")
def _ph_card_type_negative_ext(m):
    return Modification(kind="card_type_check_negative", args=(m.group(1).lower(),))


@_phr(r"^a land is tapped for mana$")
def _ph_land_tapped_for_mana(_m):
    return Modification(kind="land_tapped_for_mana", args=())


@_phr(r"^that land was nonbasic$")
def _ph_land_was_nonbasic(_m):
    return Modification(kind="land_was_nonbasic", args=())


# ---------------------------------------------------------------------------
# Pre-encoded effect tokens (analogous to wave-2 animate:/becomes_creature:/…).
# ---------------------------------------------------------------------------

def _parse_kv(payload: str) -> tuple[tuple[str, str], ...]:
    out: list[tuple[str, str]] = []
    for part in payload.split(";"):
        part = part.strip()
        if not part:
            continue
        if "=" in part:
            k, _, v = part.partition("=")
        elif ":" in part:
            k, _, v = part.partition(":")
        else:
            continue
        out.append((k.strip().lower(), v.strip().lower()))
    return tuple(out)


@_phr(r"^add_(\{[wubrgc]\})_per:(.+)$")
def _ph_add_mana_per_token(m):
    return Modification(
        kind="add_mana_per",
        args=(m.group(1).upper(), m.group(2).strip().lower()),
    )


@_phr(r"^choose:([a-z]+)(?::(.+))?$")
def _ph_choose_token(m):
    category = m.group(1).lower()
    detail = (m.group(2) or "").strip().lower() or None
    if detail:
        return Modification(kind="choose", args=(category, detail))
    return Modification(kind="choose", args=(category,))


@_phr(r"^choose khans or dragons$")
def _ph_choose_khans_dragons(_m):
    return Modification(kind="choose", args=("faction", "khans_or_dragons"))


@_phr(r"^become_basic:(.+)$")
def _ph_become_basic_token(m):
    return Modification(kind="become_basic_land", args=_parse_kv(m.group(1)))


@_phr(r"^opponent_chooses:(\d+)$")
def _ph_opp_chooses_n_token(m):
    return Modification(kind="opp_chooses_n", args=(int(m.group(1)),))


# ---------------------------------------------------------------------------
# Static / replacement / restriction effects (parsed_tail, parsed_effect_residual).
# ---------------------------------------------------------------------------

_CREATURES_GAIN_KW_RE = re.compile(
    r"^creatures you control gain "
    r"([a-z][a-z+, /-]*?) until end of turn$",
    re.I,
)


@_phr(_CREATURES_GAIN_KW_RE.pattern)
def _ph_creatures_gain_keyword_eot(m):
    keywords = tuple(
        k.strip().lower()
        for k in re.split(r",\s*(?:and\s+)?|\s+and\s+", m.group(1))
        if k.strip()
    )
    return Modification(
        kind="creatures_gain_keyword_eot",
        args=keywords,
        layer="6",
    )


@_phr(r"^all creatures have ([a-z][a-z, ]*)$")
def _ph_all_creatures_have_keyword(m):
    keywords = tuple(
        k.strip().lower()
        for k in re.split(r",\s*(?:and\s+)?|\s+and\s+", m.group(1))
        if k.strip()
    )
    return Modification(
        kind="all_creatures_have_keyword",
        args=keywords,
        layer="6",
    )


@_phr(r"^(other permanents|artifacts|enchantments|creatures|lands|nonland permanents) you control have ([a-z][a-z, ]*)$")
def _ph_permanents_have_keyword(m):
    subj = m.group(1).strip().lower().replace(" ", "_")
    keywords = tuple(
        k.strip().lower()
        for k in re.split(r",\s*(?:and\s+)?|\s+and\s+", m.group(2))
        if k.strip()
    )
    return Modification(
        kind="permanents_have_keyword",
        args=(subj,) + keywords,
        layer="6",
    )


@_phr(r"^creatures your opponents control enter tapped$")
def _ph_opp_creatures_enter_tapped(_m):
    return Modification(kind="opp_creatures_enter_tapped", args=())


@_phr(r"^lands you control enter untapped$")
def _ph_lands_enter_untapped(_m):
    return Modification(kind="lands_enter_untapped", args=())


@_phr(r"^each player can'?t cast more than one spell each turn$")
def _ph_each_player_one_spell(_m):
    return Modification(kind="cast_limit_one_per_turn", args=("each_player",))


@_phr(r"^you can'?t cast more than one spell each turn$")
def _ph_you_one_spell(_m):
    return Modification(kind="cast_limit_one_per_turn", args=("you",))


@_phr(r"^this creature is the chosen type in addition to its other types$")
def _ph_chosen_type_added(_m):
    return Modification(kind="chosen_type_added", args=(), layer="4")


@_phr(r"^damage that would reduce your life total to less than 1 reduces it to 1 instead$")
def _ph_ali_from_cairo(_m):
    return Modification(kind="cant_lose_life_below", args=(1,))


@_phr(r"^~ is all colors$")
def _ph_this_is_all_colors(_m):
    return Modification(kind="this_is_all_colors", args=(), layer="5")


@_phr(r"^~ enters prepared$")
def _ph_enters_prepared(_m):
    return Modification(kind="enters_prepared", args=())


# ---------------------------------------------------------------------------
# Effect-side: zone returns / sacrifices / costs / cant-be-X.
# ---------------------------------------------------------------------------

@_phr(r"^return (?:that creature|it|them|those creatures) to (?:its|their) owner'?s? hand at end of combat$")
def _ph_return_at_eoc(_m):
    return Modification(kind="return_to_hand_at", args=("end_of_combat",))


@_phr(r"^destroy (?:it|them|that creature|those creatures)(?: \(and this\))? at end of combat$")
def _ph_destroy_at_eoc(_m):
    return Modification(kind="destroy_at", args=("end_of_combat",))


@_phr(r"^the damage can'?t be prevented$")
def _ph_damage_cant_be_prevented(_m):
    return Modification(kind="damage_cant_be_prevented", args=())


@_phr(r"^a creature destroyed this way can'?t be regenerated$")
def _ph_cant_regen_this_way(_m):
    return Modification(kind="cant_regenerate_this_way", args=())


@_phr(r"^counters_var on (it|this creature)$")
def _ph_counters_var_on(m):
    subj = "it" if m.group(1).lower() == "it" else "this"
    return Modification(kind="counters_var_on_this", args=(subj,))


@_phr(r"^put (?:one|that) pile into your hand and the other into your graveyard$")
def _ph_pile_split(_m):
    return Modification(kind="pile_split_distribution", args=("hand_vs_graveyard",))


@_phr(r"^opp separates piles$")
def _ph_opp_separates(_m):
    return Modification(kind="opp_separates_piles", args=())


@_phr(r"^you may cast the copy$")
def _ph_may_cast_copy(_m):
    return Modification(kind="may_cast_copy", args=())


@_phr(r"^sacrifice those tokens$")
def _ph_sacrifice_those_tokens(_m):
    return Modification(kind="sacrifice_those_tokens", args=())


@_phr(r"^this spell costs \{(x|\d+)\} less to cast,? where x is (.+)$")
def _ph_cost_reduction_var(m):
    v = m.group(1).lower()
    v = int(v) if v.isdigit() else v
    return Modification(
        kind="cost_reduction_var",
        args=(v, m.group(2).strip().lower().rstrip(".")),
    )


@_phr(r"^until end of turn,? you may play (?:that card|it) without paying its mana cost$")
def _ph_may_play_no_cost_eot(_m):
    return Modification(kind="may_play_no_cost", args=("eot",))


@_phr(r"^you may play that card until the end of your next turn$")
def _ph_may_play_until_next(_m):
    return Modification(kind="may_play_exiled", args=("until_end_of_next_turn",))


@_phr(r"^you lose (\d+|x) life for each ([a-z][a-z0-9'\- ]*)$")
def _ph_lose_life_per(m):
    v = m.group(1).lower()
    v = int(v) if v.isdigit() else v
    return Modification(
        kind="lose_life_per",
        args=(v, m.group(2).strip().lower()),
    )


# ============================================================================
# Walker — structurally identical to waves 1b / 2.
# ============================================================================

def _match_residual_string(raw: str):
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
        rest = args[1:] if len(args) > 1 else ()
        return Modification(kind="if_intervening_tail", args=(typed,) + rest, layer=mod.layer)
    return typed


def _walk_field(val):
    if isinstance(val, UnknownEffect):
        return val
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
    if isinstance(node, Modification):
        promoted = _try_promote_residual(node)
        if promoted is not node:
            return promoted
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


def _wave3_post_parse_hook(card_ast: CardAST) -> CardAST:
    """Walk the CardAST and promote residual-wrapper Modifications waves 1b/2 missed."""
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


POST_PARSE_HOOKS = [_wave3_post_parse_hook]
