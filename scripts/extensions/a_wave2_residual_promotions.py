#!/usr/bin/env python3
"""Wave 2 — second-tier promotion of residual-wrapper ``Modification`` stubs.

Wave 1b (``a_wave1b_residual_promotions.py``) walks the four residual wrapper
kinds left behind by ``parser._maybe_promote_unknown`` and rewrites the most
common phrase families into typed ``Modification`` discriminators. Wave 2
runs *after* wave 1b, against whatever residuals it left behind, and targets
the next-tier patterns surfaced by ``scripts/audit_unknown_effects.py``
plus a residual-aware drill of the post-wave-1b dataset:

  controls_permanent       70 nodes   "you control a/an X" — basic land,
                                      subtype, color-creature, permanent type
  state_flag_check (ext)   35 nodes   "you have a full party / you've
                                      completed a dungeon / evidence was
                                      collected / the gift was promised /
                                      a dragon was beheld"
  would_event_check        26 nodes   "you would draw a card / lose the
                                      game / put one or more counters on..."
  damage_would_be_dealt    12 nodes   "damage would be dealt to you /
                                      this creature / target / a player"
  counters_would_be_put    12 nodes   "one or more +1/+1 counters would be
                                      put on a creature you control"
  colored_mana_spent       12 nodes   "{B} / {U} / ... was spent to cast
                                      this spell"
  nth_resolve              14 nodes   "this is the {second|third|...} time
                                      this ability has resolved this turn"
  nth_time_short            9 nodes   "it's the {second|third} time"
  cast_from_zone (ext)      5 nodes   "this spell was cast from exile /
                                      your hand / the command zone" — extends
                                      wave-1b ``cast_mode_check``
  cast_during_phase         5 nodes   "you cast this spell during your
                                      {main|combat|end} phase"
  creature_event_turn       5 nodes   "a creature died/entered this turn"
  card_type_check (ext)    18 nodes   adds "it's legendary / tapped /
                                      attacking" and "it isn't X" tails
  controls_count_thr       10 nodes   "you control N or more X" /
                                      "you control a creature with
                                      power N or greater"
  no_depletion              5 nodes   "there are no depletion counters
                                      on this land"

  animate_subject         117 nodes   parser tokens emitted as
                                      "animate:subj=...;pt=...;descr=...;type=..."
  becomes_creature         67 nodes   "becomes_creature:type=...;pt=..."
  bare_pt                  50 nodes   "bare p/t N/N"
  type_change              34 nodes   "type_change:subj=...;to=..."
  opp_choose_pile          28 nodes   "opp_choose:pile" and friends
  you_choose_pile           8 nodes   "you_choose:..."
  extra_land_per_turn      18 nodes   bare "extra land per turn"
  gift_promised            17 nodes   "the gift was promised"
  change_target            16 nodes   "change target"
  play_exiled_eot          15 nodes   "play exiled cards until end of turn"
  may_play_exiled           9 nodes   "you may play that card this turn"
  cant_block                8 nodes   "~ can't block" tail
  counters_var_on_this      8 nodes   "counters_var on this creature"
  skip_next_turn            8 nodes   "skip next turn"
  repeat_process            8 nodes   "repeat this process"
  trample_alike             8 nodes   "you may have this creature assign
                                      its combat damage as though it
                                      weren't blocked"
  incubate_n                7 nodes   "incubate X"
  triggers_once             7 nodes   "this ability triggers only once"
  may_pay_amount            7 nodes   "you may pay {X}"
  cant_be_blocked_eot       7 nodes   "it can't be blocked this turn"
  must_attack_eot           7 nodes   "target_creature_must_attack_eot"
  experience_counter        6 nodes   "you get an experience counter"
  lose_half_life            6 nodes   "that player loses half their life,
                                      rounded up"
  end_the_turn              6 nodes   bare "end the turn"
  may_cast_exiled           5 nodes   "you may cast this card for as long
                                      as it remains exiled"
  roll_die                  5 nodes   "roll a d20" / "roll a d6"
  cost_reduction_static     5 nodes   "this spell costs {N} less to cast"
  self_damage_power         7 nodes   "target creature deals damage to
                                      itself equal to its power"

Estimated total promotions: ~700 stub nodes per corpus pass.

Non-goals (same as wave 1b):
  - No new AST classes — promotions stay inside ``Modification``.
  - No grammar changes in ``scripts/parser.py``.
  - No regression on goldens — walker only fires on residual wrapper kinds.

Ordering: wave 1b's filename sorts before wave 2's, so wave 1b's hook runs
first; the wave 2 walker sees only the residuals wave 1b couldn't type.
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


# ---------------------------------------------------------------------------
# Condition-side promotions (mostly if_intervening_tail args[0]).
# ---------------------------------------------------------------------------

_BASIC_LANDS = ("plains", "island", "swamp", "mountain", "forest", "wastes")
_CREATURE_SUBTYPES = (
    "wizard", "cleric", "warrior", "rogue", "knight", "soldier", "goblin",
    "elf", "zombie", "vampire", "dragon", "sliver", "merfolk", "human",
    "spirit", "ally", "angel", "demon", "giant", "dwarf", "samurai",
    "ninja", "pirate", "monk", "shaman", "druid", "bird", "cat", "dog",
    "snake", "wolf", "wurm", "beast", "elemental", "horror", "phoenix",
)
_PERMANENT_TYPES = (
    "artifact", "enchantment", "creature", "land", "planeswalker",
    "instant", "sorcery", "tribal", "battle", "saga", "vehicle", "equipment",
    "aura", "token", "legendary creature", "legendary artifact",
    "snow land", "snow permanent", "nonland permanent", "nontoken creature",
)
_PROTECTION_FILTERS = (
    "white creature", "blue creature", "black creature", "red creature",
    "green creature", "colorless creature", "multicolored creature",
)


# Register the more-specific count / power thresholds BEFORE the generic
# "you control a X" pattern, so they win the match.

@_phr(r"^you control a creature with power (\d+) or (greater|more|less)$")
def _ph_controls_power_threshold(m):
    rel = "or_less" if m.group(2).lower() == "less" else "or_more"
    return Modification(
        kind="controls_power_threshold",
        args=(int(m.group(1)), rel),
    )


@_phr(r"^you control (\d+) or more ([a-z][a-z0-9\- ]+?)$")
def _ph_controls_count_threshold(m):
    return Modification(
        kind="controls_count_threshold",
        args=(int(m.group(1)), m.group(2).strip().lower(), "or_more"),
    )


@_phr(r"^you control (a|an|another|each) ([a-z][a-z0-9'\- ]+?)$")
def _ph_controls_permanent(m):
    quant = m.group(1).lower()
    obj = m.group(2).strip().lower()
    # Decline if the tail still has structural words the more-specific
    # patterns above should have caught — keeps us from swallowing
    # "creature with power N or greater" etc.
    if " with " in obj or " or " in obj:
        return None
    if obj in _BASIC_LANDS:
        return Modification(kind="controls_basic_land", args=(obj,))
    if obj in _CREATURE_SUBTYPES:
        return Modification(kind="controls_creature_subtype", args=(obj,))
    if obj in _PROTECTION_FILTERS:
        return Modification(kind="controls_filtered_creature", args=(obj,))
    if obj in _PERMANENT_TYPES:
        return Modification(kind="controls_permanent_type", args=(obj,))
    # Generic catch-all so the wrapper still gets typed.
    return Modification(kind="controls_permanent", args=(quant, obj))


# Extend wave-1b cast_mode_check with more zones / phases / colored-mana spent.

@_phr(r"^this spell was cast from (exile|your hand|the command zone|your graveyard)$")
def _ph_cast_from_zone(m):
    zone = m.group(1).lower()
    key = {
        "exile": "from_exile",
        "your hand": "from_hand",
        "the command zone": "from_command_zone",
        "your graveyard": "from_graveyard",
    }[zone]
    return Modification(kind="cast_mode_check", args=(key,))


@_phr(r"^you cast this spell during your (main|combat|end|beginning|precombat|postcombat) phase$")
def _ph_cast_during_phase(m):
    return Modification(kind="cast_during_phase", args=(m.group(1).lower(),))


@_phr(r"^\{([wubrgc])\}(?:/\{[wubrgc]\})?\s+was spent to cast this spell$")
def _ph_colored_mana_spent(m):
    return Modification(
        kind="colored_mana_spent_check",
        args=(m.group(1).upper(),),
    )


@_phr(r"^mana from a treasure was spent to cast this spell$")
def _ph_treasure_mana_spent(_m):
    return Modification(kind="colored_mana_spent_check", args=("treasure",))


# State-flag extensions — wave 1b covered monarch / city's blessing / you win;
# wave 2 adds the rest of the per-set state flags that show up as conditions.

_STATE_FLAGS_EXACT = {
    "you have a full party": "full_party",
    "you've completed a dungeon": "completed_dungeon",
    "you have completed a dungeon": "completed_dungeon",
    "evidence was collected": "evidence_collected",
    "the gift was promised": "gift_promised",
    "a dragon was beheld": "dragon_beheld",
    "the day is day": "day_is_day",
    "the day is night": "day_is_night",
    "you have the initiative": "initiative",
    "you took the initiative": "took_initiative",
    "you weren't the starting player": "not_starting_player",
    "you were the starting player": "starting_player",
    "no one does": "no_one_does",
    "planeswalk gets more votes": "vote_planeswalk_wins",
    "chaos gets more votes or the vote is tied": "vote_chaos_or_tied",
}


@_phr(r"^(.+)$")
def _ph_state_flag_extended(m):
    raw = m.group(1).lower().strip().rstrip(".")
    key = _STATE_FLAGS_EXACT.get(raw)
    if key is None:
        return None
    return Modification(kind="state_flag_check", args=(key,))


# Resolve-count / activation-count thresholds.

_ORDINALS = {
    "second": 2, "third": 3, "fourth": 4, "fifth": 5,
    "sixth": 6, "seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10,
}


@_phr(r"^this is the (second|third|fourth|fifth|sixth|seventh|eighth|ninth|tenth) time this ability has resolved this turn$")
def _ph_nth_resolve(m):
    return Modification(
        kind="resolve_count_check",
        args=(_ORDINALS[m.group(1).lower()],),
    )


@_phr(r"^it'?s the (second|third|fourth|fifth|sixth|seventh|eighth|ninth|tenth) time$")
def _ph_nth_time_short(m):
    return Modification(
        kind="nth_time_short",
        args=(_ORDINALS[m.group(1).lower()],),
    )


@_phr(r"^this ability has been activated (\d+) or more times this turn$")
def _ph_nth_activated(m):
    return Modification(
        kind="activation_count_check",
        args=(int(m.group(1)), "or_more"),
    )


# Same-turn event checks.

@_phr(r"^a creature (died|entered the battlefield|entered) this turn$")
def _ph_creature_event_turn(m):
    ev = m.group(1).lower()
    key = "died" if ev == "died" else "entered"
    return Modification(kind="creature_event_this_turn", args=(key,))


# "would" replacement conditions.

@_phr(r"^you would (draw a card)(?: while your library has no cards in it)?$")
def _ph_would_draw(m):
    return Modification(kind="would_event_check", args=("draw_card",))


@_phr(r"^you would lose the game$")
def _ph_would_lose(_m):
    return Modification(kind="would_event_check", args=("lose_game",))


@_phr(r"^you would gain life$")
def _ph_would_gain_life(_m):
    return Modification(kind="would_event_check", args=("gain_life",))


@_phr(r"^damage would be dealt to (you|this creature|target creature you control|a player|any of you|a creature you control)$")
def _ph_damage_would_be_dealt(m):
    subj = m.group(1).lower().replace(" ", "_")
    return Modification(kind="damage_would_be_dealt", args=(subj,))


@_phr(r"^one or more (\+1/\+1|\-1/\-1)?\s*counters? would be put on (a creature you control|target creature|this creature|a permanent you control)$")
def _ph_counters_would_be_put(m):
    ctype = (m.group(1) or "any").lower()
    subj = m.group(2).lower().replace(" ", "_")
    return Modification(kind="counters_would_be_put", args=(ctype, subj))


# Card-type extensions (wave 1b handled "it's a creature card" etc.; wave 2
# extends to short status / supertype / subtype tails).

_CARD_STATUS_EXACT = {
    "legendary": "supertype:legendary",
    "tapped": "status:tapped",
    "untapped": "status:untapped",
    "attacking": "status:attacking",
    "blocking": "status:blocking",
    "blocked": "status:blocked",
    "an artifact": "type:artifact",
    "an artifact creature": "type:artifact_creature",
    "a vampire": "subtype:vampire",
    "a wizard": "subtype:wizard",
    "a knight": "subtype:knight",
    "a goblin": "subtype:goblin",
    "a zombie": "subtype:zombie",
    "a human": "subtype:human",
    "a soldier": "subtype:soldier",
    "a beast": "subtype:beast",
    "a dragon": "subtype:dragon",
    "a sliver": "subtype:sliver",
    "an elf": "subtype:elf",
    "a cleric": "subtype:cleric",
    "a warrior": "subtype:warrior",
    "a rogue": "subtype:rogue",
    "a spirit": "subtype:spirit",
}


@_phr(r"^it'?s (legendary|tapped|untapped|attacking|blocking|blocked|an artifact|an artifact creature|a vampire|a wizard|a knight|a goblin|a zombie|a human|a soldier|a beast|a dragon|a sliver|an elf|a cleric|a warrior|a rogue|a spirit)$")
def _ph_card_status_short(m):
    key = _CARD_STATUS_EXACT[m.group(1).lower()]
    return Modification(kind="card_type_check", args=(key,))


@_phr(r"^it isn'?t (a creature|an artifact|legendary|a land|a creature card)$")
def _ph_card_isnt(m):
    obj = m.group(1).lower()
    return Modification(kind="card_type_check_negative", args=(obj,))


@_phr(r"^there are no depletion counters on this land$")
def _ph_no_depletion(_m):
    return Modification(kind="no_depletion_counters", args=())


# ---------------------------------------------------------------------------
# Effect-side promotions (parsed_effect_residual, parsed_tail).
# ---------------------------------------------------------------------------

# Structured tokens emitted as raw text by upstream extensions
# (type_changes.py, choices.py, old_templating.py). Parse the inner
# key=value pairs so downstream tools can read them as args.

def _parse_kv(payload: str) -> tuple[tuple[str, str], ...]:
    """Parse `key1=v1;key2=v2;key3=v3` payload into a tuple of pairs."""
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


@_phr(r"^animate:(.+)$")
def _ph_animate_token(m):
    return Modification(kind="animate_subject", args=_parse_kv(m.group(1)))


@_phr(r"^becomes_creature:(.+)$")
def _ph_becomes_creature_token(m):
    return Modification(kind="becomes_creature_subject", args=_parse_kv(m.group(1)))


@_phr(r"^it_becomes_creature:(.+)$")
def _ph_it_becomes_creature_token(m):
    return Modification(kind="becomes_creature_subject", args=_parse_kv(m.group(1)))


@_phr(r"^type_change:(.+)$")
def _ph_type_change_token(m):
    return Modification(kind="type_change_subject", args=_parse_kv(m.group(1)))


@_phr(r"^opp_choose:(.+)$")
def _ph_opp_choose_token(m):
    body = m.group(1).strip().lower()
    return Modification(kind="opp_choose", args=(body,))


@_phr(r"^you_choose:(.+)$")
def _ph_you_choose_token(m):
    body = m.group(1).strip().lower()
    return Modification(kind="you_choose", args=(body,))


# Bare P/T set.

@_phr(r"^bare p/t (\d+|x)/(\d+|x)$")
def _ph_bare_pt(m):
    def _val(s):
        return int(s) if s.isdigit() else "x"
    return Modification(
        kind="becomes_pt",
        args=(_val(m.group(1)), _val(m.group(2))),
        layer="7b",
    )


# Effect tokens that already encode a concept — just type them.

@_phr(r"^extra land per turn$")
def _ph_extra_land_per_turn(_m):
    return Modification(kind="extra_land_per_turn", args=(1,))


@_phr(r"^you may play (?:that card|it) this turn$")
def _ph_may_play_this_turn(_m):
    return Modification(kind="may_play_exiled", args=("this_turn",))


@_phr(r"^you may play (?:that card|it) for as long as it remains exiled$")
def _ph_may_play_while_exiled(_m):
    return Modification(kind="may_play_exiled", args=("while_exiled",))


@_phr(r"^you may cast (?:that card|this card|it) for as long as it remains exiled$")
def _ph_may_cast_while_exiled(_m):
    return Modification(kind="may_cast_exiled", args=("while_exiled",))


@_phr(r"^play exiled cards until end of turn$")
def _ph_play_exiled_eot(_m):
    return Modification(kind="play_exiled_cards", args=("eot",))


@_phr(r"^end the turn$")
def _ph_end_the_turn(_m):
    return Modification(kind="end_the_turn", args=())


@_phr(r"^skip next turn$")
def _ph_skip_next_turn(_m):
    return Modification(kind="skip_turn", args=("next",))


@_phr(r"^skip your next (draw|untap|combat|main|end) (step|phase)$")
def _ph_skip_phase(m):
    return Modification(
        kind="skip_phase",
        args=(m.group(1).lower(), m.group(2).lower()),
    )


@_phr(r"^change target$")
def _ph_change_target(_m):
    return Modification(kind="change_target", args=())


@_phr(r"^it can'?t be blocked this turn$")
def _ph_cant_be_blocked_eot(_m):
    return Modification(kind="cant_be_blocked", args=("eot",))


@_phr(r"^~ can'?t block$")
def _ph_cant_block_tail(_m):
    return Modification(kind="cant_block", args=())


@_phr(r"^target_creature_must_attack_eot$")
def _ph_must_attack_eot(_m):
    return Modification(kind="must_attack", args=("eot",))


@_phr(r"^this ability triggers only once$")
def _ph_triggers_once(_m):
    return Modification(kind="triggers_once", args=())


@_phr(r"^this spell costs \{(\d+|x)\} less to cast$")
def _ph_cost_reduction_static(m):
    v = m.group(1).lower()
    v = int(v) if v.isdigit() else v
    return Modification(kind="cost_reduction_static", args=(v,))


@_phr(r"^you may pay \{(x|\d+)\}$")
def _ph_may_pay_amount(m):
    v = m.group(1).lower()
    v = int(v) if v.isdigit() else v
    return Modification(kind="may_pay_amount", args=(v,))


@_phr(r"^repeat this process$")
def _ph_repeat_process(_m):
    return Modification(kind="repeat_process", args=())


@_phr(r"^the game is a draw$")
def _ph_game_is_draw(_m):
    return Modification(kind="game_is_draw", args=())


@_phr(r"^you get an experience counter$")
def _ph_experience_counter(_m):
    return Modification(kind="experience_counter", args=(1,))


@_phr(r"^you may have this creature assign its combat damage as though it weren'?t blocked$")
def _ph_trample_alike(_m):
    return Modification(kind="combat_damage_as_unblocked", args=())


@_phr(r"^counters_var on this creature$")
def _ph_counters_var_on_this(_m):
    return Modification(kind="counters_var_on_this", args=())


@_phr(r"^incubate (x|\d+)$")
def _ph_incubate_n(m):
    v = m.group(1).lower()
    v = int(v) if v.isdigit() else v
    return Modification(kind="incubate", args=(v,))


@_phr(r"^that player loses half their life,? rounded (up|down)$")
def _ph_lose_half_life(m):
    return Modification(kind="lose_half_life", args=(m.group(1).lower(),))


@_phr(r"^target creature deals damage to itself equal to its power$")
def _ph_self_damage_power(_m):
    return Modification(kind="self_damage_equal_to_power", args=())


@_phr(r"^roll a d(\d+)$")
def _ph_roll_die(m):
    return Modification(kind="roll_die", args=(int(m.group(1)),))


# ============================================================================
# Walker — replace residual-wrapper Modifications via the table above.
# Structurally identical to wave 1b; kept here to keep the two passes
# independent (wave 1b's table is closed, wave 2's can grow without
# touching it).
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


def _wave2_post_parse_hook(card_ast: CardAST) -> CardAST:
    """Walk the CardAST and promote residual-wrapper Modifications wave 1b missed."""
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


POST_PARSE_HOOKS = [_wave2_post_parse_hook]
