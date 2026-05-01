#!/usr/bin/env python3
"""Wave B parser phrase-coverage promotions (post-Wave A).

Named ``a_wave_b_*`` so it loads AFTER ``a_wave_a_promotions.py`` but
BEFORE all other extensions. This ensures Wave B rules preempt the
labeled-Modification stubs in later-loading catch-all extensions
(``partial_final.py``, ``unparsed_final_sweep.py``, etc.).

Goal: promote ~400+ high-frequency ``parsed_effect_residual`` phrases to
typed AST nodes, pushing structural coverage from ~72.6% toward 85%+.

Target families (by corpus frequency):
  - Bounce own creatures/permanents (~17 cards)
  - Untap specific targets (~35 cards)
  - Tap specific targets (~20 cards)
  - Discard-then-draw / loot patterns (~47 cards)
  - Bolster N (~11 cards)
  - Treasure token creation (~6 cards)
  - Controller loses life (~10 cards)
  - Detain (~9 cards)
  - Exile top of library (~31 cards)
  - Creatures get -N/-N debuff (~15 cards)
  - Target doesn't untap / stun (~17 cards)
  - Another target gains keyword (~49 cards)
  - Can't be blocked this turn (~52 cards)
  - Can't block this turn (~21 cards)
  - Draw additional card (~6 cards)
  - Untap lands (~12 cards)
  - Put card on bottom of library (~26 cards)
  - Return from graveyard (~20+ cards)
  - Goad target (~6 cards)

Non-goals:
  - No new AST node types — all promotions route through existing nodes.
  - No per-card snowflake handlers.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

_HERE = Path(__file__).resolve().parent
_SCRIPTS = _HERE.parent
if str(_SCRIPTS) not in sys.path:
    sys.path.insert(0, str(_SCRIPTS))

from mtg_ast import (  # noqa: E402
    Bounce, Buff, CounterMod, CreateToken, Damage, Destroy, Discard,
    Draw, Exile, Filter, GainLife, GrantAbility, LoseLife, Mill,
    Modification, Optional_, Reanimate, Recurse, Sacrifice, Scry,
    Sequence, Shuffle, TapEffect, Tutor, UntapEffect, UnknownEffect,
    EACH_OPPONENT, EACH_PLAYER, SELF, TARGET_ANY, TARGET_CREATURE,
    TARGET_OPPONENT, TARGET_PLAYER,
)


EFFECT_RULES: list[tuple[re.Pattern, callable]] = []


def _eff(pattern: str):
    def deco(fn):
        EFFECT_RULES.append((re.compile(pattern, re.I | re.S), fn))
        return fn
    return deco


_NUMS = {
    "a": 1, "an": 1, "one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
    "six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
}
_NUM_RE = r"(?:a|an|one|two|three|four|five|six|seven|eight|nine|ten|x|\d+)"


def _n(tok: str):
    t = (tok or "").strip().lower()
    if t in _NUMS:
        return _NUMS[t]
    if t.isdigit():
        return int(t)
    return t


def _parse_filter(s: str):
    """Lazy import parse_filter from parser."""
    from parser import parse_filter
    return parse_filter(s)


# ============================================================================
# GROUP 1: Bounce own creatures/permanents to hand
# ============================================================================
# "return a creature you control to its owner's hand" (7)
# "return a permanent you control to its owner's hand" (5)
# "return another creature you control to its owner's hand" (3)
# "return another permanent you control to its owner's hand" (1)

@_eff(r"^return (?:a|an|another|another target) ((?:[a-z]+ )?(?:creature|permanent))(?: you control)? to its owner'?s? hand(?:\.|$)")
def _bounce_own(m):
    base = m.group(1).lower()
    targeted = "target" in m.group(0).lower()
    return Bounce(target=Filter(base=base.split()[-1], you_control=True, targeted=targeted))


# ============================================================================
# GROUP 2: Untap specific targets
# ============================================================================
# "untap another target permanent" (6)
# "untap another target creature you control" (3)
# "untap another target permanent you control" (3)
# "untap this creature" variant (already has a rule but quoted version fails)

@_eff(r"^untap another target (creature|permanent|artifact|land)(?: you control)?(?:\.|$)")
def _untap_another_target(m):
    base = m.group(1).lower()
    you_ctrl = "you control" in m.group(0).lower()
    return UntapEffect(target=Filter(base=base, targeted=True,
                                      you_control=you_ctrl,
                                      extra=("other",)))


# "untap two target lands" (4)
# "untap up to N lands" (various)
# "untap x target lands" (2)
@_eff(r"^untap (?:up to )?(" + _NUM_RE + r") target (creatures?|permanents?|lands?|artifacts?)(?:\.|$)")
def _untap_n_target(m):
    n = _n(m.group(1))
    base = m.group(2).lower().rstrip("s")
    q = "up_to_n" if "up to" in m.group(0).lower() else "exact"
    return UntapEffect(target=Filter(base=base, quantifier=q, count=n, targeted=True))


# "untap up to N lands" (no "target")
@_eff(r"^untap up to (" + _NUM_RE + r") (creatures?|permanents?|lands?|artifacts?)(?:\.|$)")
def _untap_up_to_n(m):
    n = _n(m.group(1))
    base = m.group(2).lower().rstrip("s")
    return UntapEffect(target=Filter(base=base, quantifier="up_to_n", count=n, targeted=False))


# ============================================================================
# GROUP 3: Tap specific targets
# ============================================================================
# "tap enchanted permanent" (7)
# "tap enchanted creature" (variant)
@_eff(r"^tap (?:enchanted|equipped) (creature|permanent|artifact|land)(?:\.|$)")
def _tap_enchanted(m):
    base = m.group(1).lower()
    return TapEffect(target=Filter(base=f"enchanted_{base}", targeted=False))


# "tap another target creature" (4)
@_eff(r"^tap another target (creature|permanent|artifact|land)(?:\.|$)")
def _tap_another_target(m):
    base = m.group(1).lower()
    return TapEffect(target=Filter(base=base, targeted=True, extra=("other",)))


# ============================================================================
# GROUP 4: Discard-then-draw (loot/rummage patterns)
# ============================================================================
# "discard up to N cards, then draw that many cards" (7)
@_eff(r"^discard (?:up to )?(" + _NUM_RE + r") cards?,\s*then draw that many cards(?:\.|$)")
def _rummage_n(m):
    n = _n(m.group(1))
    return Sequence(items=(
        Discard(count=n, target=SELF),
        Draw(count="that_many", target=SELF),
    ))


# "discard any number of cards, then draw that many cards (plus one)?" (3)
@_eff(r"^discard any number of cards,?\s*then draw that many cards(?:\s+plus one)?(?:\.|$)")
def _rummage_any(m):
    return Sequence(items=(
        Discard(count="any", target=SELF),
        Draw(count="that_many", target=SELF),
    ))


# "target player draws N cards, then discards N cards" (looting variant)
@_eff(r"^target (?:player|opponent) draws? (" + _NUM_RE + r") cards?,\s*then discards? (" + _NUM_RE + r") cards?(?:\.|$)")
def _target_draw_then_discard(m):
    draw_n = _n(m.group(1))
    disc_n = _n(m.group(2))
    who = TARGET_OPPONENT if "opponent" in m.group(0).lower() else TARGET_PLAYER
    return Sequence(items=(
        Draw(count=draw_n, target=who),
        Discard(count=disc_n, target=who),
    ))


# "target player draws a card, then discards a card"
@_eff(r"^target (?:player|opponent) draws? a card,?\s*then discards? a card(?:\.|$)")
def _target_draw_discard_one(m):
    who = TARGET_OPPONENT if "opponent" in m.group(0).lower() else TARGET_PLAYER
    return Sequence(items=(
        Draw(count=1, target=who),
        Discard(count=1, target=who),
    ))


# "you may discard your hand and draw two cards"
@_eff(r"^you may discard your hand and draw (" + _NUM_RE + r") cards?(?:\.|$)")
def _may_discard_hand_draw(m):
    n = _n(m.group(1))
    return Optional_(body=Sequence(items=(
        Discard(count="all", target=SELF),
        Draw(count=n, target=SELF),
    )))


# ============================================================================
# GROUP 5: Bolster N
# ============================================================================
# "bolster 1" (6), "bolster 2" (4), "bolster 5" (1)
@_eff(r"^bolster (" + _NUM_RE + r")(?:\.|$)")
def _bolster(m):
    n = _n(m.group(1))
    return CounterMod(op="put", count=n, counter_kind="+1/+1",
                      target=Filter(base="creature", you_control=True,
                                    extra=("least_toughness",)))


# ============================================================================
# GROUP 6: Treasure token creation
# ============================================================================
# "you create a treasure token" (4)
# "create a treasure token" (various)
@_eff(r"^(?:you )?create (" + _NUM_RE + r" )?treasure tokens?(?:\.|$)")
def _create_treasure(m):
    n = _n(m.group(1)) if m.group(1) else 1
    return CreateToken(count=n, types=("treasure",), pt=None)


# ============================================================================
# GROUP 7: Controller loses life
# ============================================================================
# "its controller loses 2 life" (6)
# "its controller loses N life" (various)
@_eff(r"^its controller loses (\d+) life(?:\.|$)")
def _controller_loses_life(m):
    return LoseLife(amount=int(m.group(1)),
                    target=Filter(base="its_controller", targeted=False))


# "its controller loses N life and you gain N life" (1)
@_eff(r"^its controller loses (\d+) life and you gain (\d+) life(?:\.|$)")
def _controller_loses_you_gain(m):
    return Sequence(items=(
        LoseLife(amount=int(m.group(1)),
                 target=Filter(base="its_controller", targeted=False)),
        GainLife(amount=int(m.group(2))),
    ))


# ============================================================================
# GROUP 8: Detain
# ============================================================================
# "detain target creature an opponent controls" (4)
# "detain target nonland permanent an opponent controls" (2)
# "detain up to two target creatures your opponents control" (1)
# "detain each nonland permanent your opponents control" (1)
@_eff(r"^detain (?:up to (" + _NUM_RE + r") )?(?:target |each )?((?:nonland )?(?:creature|permanent)s?)(?:\s+(?:your opponents?|an opponent) controls?)?(?:\s+with [^.]+)?(?:\.|$)")
def _detain(m):
    return TapEffect(target=Filter(
        base="permanent",
        targeted="target" in m.group(0).lower(),
        opponent_controls=True,
        extra=("detain",),
    ))


# ============================================================================
# GROUP 9: Exile top card(s) of library
# ============================================================================
# "exile the top card of each player's library" (5)
# "exile the top card of each opponent's library" (4)
# "exile the top card of target player's library" (3)
# "exile the top card of that player's library" (5)
@_eff(r"^exile the top (?:card|(" + _NUM_RE + r") cards?) of (?:each (player|opponent)'s|target (player|opponent)'s|that player'?s?) library(?:\.|$)")
def _exile_top_library(m):
    n = _n(m.group(1)) if m.group(1) else 1
    text_low = m.group(0).lower()
    if "each player" in text_low:
        who = EACH_PLAYER
    elif "each opponent" in text_low:
        who = EACH_OPPONENT
    elif "target opponent" in text_low:
        who = TARGET_OPPONENT
    elif "target player" in text_low:
        who = TARGET_PLAYER
    else:
        who = Filter(base="that_player", targeted=False)
    return Exile(target=who)


# "target opponent exiles the top card of their library" (4)
# "target opponent exiles a card from their hand" (4)
@_eff(r"^target (?:opponent|player) exiles (?:the top card of their library|a card from their (?:hand|graveyard))(?:\.|$)")
def _target_player_exiles(m):
    text_low = m.group(0).lower()
    who = TARGET_OPPONENT if "opponent" in text_low else TARGET_PLAYER
    return Exile(target=who)


# "target player exiles a card from their graveyard" (4)
@_eff(r"^target player exiles a card from their graveyard(?:\.|$)")
def _target_exiles_from_gy(m):
    return Exile(target=TARGET_PLAYER)


# ============================================================================
# GROUP 10: Creatures get -N/-N debuff (opponent or own)
# ============================================================================
# "creatures your opponents control get -1/-1 until end of turn" (4)
# "creatures your opponents control get -2/-2 until end of turn" (3)
@_eff(r"^creatures (?:your opponents?|you) control get ([+-]\d+)/([+-]\d+) until end of turn(?:\.|$)")
def _creatures_debuff_eot(m):
    p = int(m.group(1))
    t = int(m.group(2))
    text_low = m.group(0).lower()
    opp = "opponent" in text_low
    return Buff(power=p, toughness=t,
                target=Filter(base="creature", quantifier="all",
                              opponent_controls=opp,
                              you_control=not opp),
                duration="until_end_of_turn")


# ============================================================================
# GROUP 11: Target doesn't untap (stun)
# ============================================================================
# "target creature doesn't untap during its controller's next untap step" (5)
# "that creature doesn't untap during its controller's next untap step" (4)
# "target creature an opponent controls doesn't untap ..." (3)
@_eff(r"^(target creature(?:\s+an opponent controls)?|that creature) doesn'?t untap during its controller'?s? (?:next )?untap step(?:\.|$)")
def _stun_creature(m):
    subj = m.group(1).lower()
    if "that creature" in subj:
        target = Filter(base="that_creature", targeted=False)
    elif "opponent" in subj:
        target = Filter(base="creature", targeted=True, opponent_controls=True)
    else:
        target = TARGET_CREATURE
    return TapEffect(target=target)


# ============================================================================
# GROUP 12: Another target creature gains keyword
# ============================================================================
# "another target creature gains haste until end of turn" (3)
# "another target attacking creature gains flying until end of turn" (4)
# "another target creature you control gains [kw] until end of turn" (various)
@_eff(r"^another target ((?:attacking |blocking )?(?:creature|permanent)(?:\s+(?:you control|without [a-z]+))?)\s+gains ([a-z ]+?) until end of turn(?:\.|$)")
def _another_target_gains_kw_eot(m):
    kw = m.group(2).strip()
    return GrantAbility(
        ability_name=kw,
        target=Filter(base="creature", targeted=True, extra=("other",)),
        duration="until_end_of_turn",
    )


# ============================================================================
# GROUP 13: Can't be blocked this turn
# ============================================================================
# "up to one target attacking creature can't be blocked this turn" (6)
# "another target attacking creature can't be blocked this turn" (3)
# "target attacking creature can't be blocked this turn" (2)
# "creatures you control can't be blocked this turn" (2)
# "target creature can't be blocked this turn" (2)
# "this creature can't be blocked this turn" (2)
@_eff(r"^(?:up to (" + _NUM_RE + r") )?(?:target |another target |another )?((?:attacking |blocking )?(?:creatures?|permanents?)(?:\s+you control)?)\s+can'?t be blocked this turn(?:\s+except by [^.]+)?(?:\.|$)")
def _cant_be_blocked_this_turn(m):
    return GrantAbility(
        ability_name="unblockable",
        target=Filter(base="creature", targeted="target" in m.group(0).lower()),
        duration="until_end_of_turn",
    )


# "this creature can't be blocked this turn" variants (with conditions)
@_eff(r"^(?:this creature|~) can'?t be blocked this turn(?:\s+except by [^.]+)?(?:\.|$)")
def _self_cant_be_blocked_this_turn(m):
    return GrantAbility(
        ability_name="unblockable",
        target=Filter(base="self", targeted=False),
        duration="until_end_of_turn",
    )


# ============================================================================
# GROUP 14: Can't block this turn
# ============================================================================
# "up to two target creatures can't block this turn" (3)
# "target creature can't block this turn" (various)
@_eff(r"^(?:up to (" + _NUM_RE + r") )?target (creatures?) can'?t block this turn(?:\.|$)")
def _target_cant_block_this_turn(m):
    return GrantAbility(
        ability_name="cant_block",
        target=Filter(base="creature", targeted=True),
        duration="until_end_of_turn",
    )


# "target creature blocks this turn if able" (3)
@_eff(r"^target creature blocks this turn if able(?:\.|$)")
def _must_block_this_turn(m):
    return GrantAbility(
        ability_name="must_block",
        target=TARGET_CREATURE,
        duration="until_end_of_turn",
    )


# ============================================================================
# GROUP 15: Draw additional card
# ============================================================================
# "draw an additional card" (4+)
@_eff(r"^draw an additional card(?:\.|$)")
def _draw_additional(m):
    return Draw(count=1, target=SELF)


# ============================================================================
# GROUP 16: Put card on bottom of library
# ============================================================================
# "put target card from your graveyard on the bottom of your library" (6)
@_eff(r"^put target card from (?:your|a|an opponent'?s?) graveyard on the bottom of (?:your|its owner'?s?|their) library(?:\.|$)")
def _tuck_from_gy(m):
    return Modification(kind="library_tuck", args=("graveyard",))


# "put the top card of your library on the bottom of your library" (1)
@_eff(r"^put the top card of your library on the bottom of your library(?:\.|$)")
def _tuck_top_card(m):
    return Modification(kind="library_tuck", args=("library_top",))


# ============================================================================
# GROUP 17: Return from graveyard
# ============================================================================
# "return another target creature card from your graveyard to your hand" (4)
@_eff(r"^return another target ([^.]+?) card from (?:your|a) graveyard to your hand(?:\.|$)")
def _recurse_another(m):
    query = _parse_filter("target " + m.group(1) + " card") or Filter(base="card")
    return Recurse(query=query, destination="hand")


# "return it from your graveyard to the battlefield" (4)
@_eff(r"^return (?:it|this card|this creature|~) from (?:your|a) graveyard to the battlefield(?:\.|$)")
def _reanimate_self_from_gy(m):
    return Reanimate(query=Filter(base="self", targeted=False))


# "return that card to the battlefield under your control" (8)
# "return that card to the battlefield under its owner's control" (various)
@_eff(r"^return (?:that card|it) to the battlefield under (?:your|its owner'?s?) control(?:\.|$)")
def _reanimate_that(m):
    return Reanimate(query=Filter(base="that_card", targeted=False))


# "return this card from your graveyard to the battlefield with a finality counter on it" (4)
@_eff(r"^return (?:this card|~) from (?:your|a) graveyard to the battlefield(?:\s+with [^.]+)?(?:\.|$)")
def _reanimate_self_with_counter(m):
    return Reanimate(query=Filter(base="self", targeted=False))


# ============================================================================
# GROUP 18: Goad
# ============================================================================
# "goad target creature" / "goad it" / "goad that creature"
@_eff(r"^goad (?:target creature|it|that creature|each creature (?:your opponents?|an opponent) controls?)(?:\.|$)")
def _goad_typed(m):
    text_low = m.group(0).lower()
    if "each" in text_low:
        target = Filter(base="creature", quantifier="all", opponent_controls=True)
    elif "target" in text_low:
        target = TARGET_CREATURE
    else:
        target = Filter(base="that_creature", targeted=False)
    return GrantAbility(
        ability_name="goad",
        target=target,
        duration="until_next_turn",
    )


# ============================================================================
# GROUP 19: You may [simple effects]
# ============================================================================
# "you may discard a card" (5)
@_eff(r"^you may discard (" + _NUM_RE + r") cards?(?:\.|$)")
def _may_discard(m):
    n = _n(m.group(1))
    return Optional_(body=Discard(count=n, target=SELF))


# "you may exile it" (5) / "you may exile target [filter]" (various)
@_eff(r"^you may exile (?:it|that card|that creature)(?:\.|$)")
def _may_exile_pronoun(m):
    return Optional_(body=Exile(target=Filter(base="that_thing", targeted=False)))


# "you may exile target card from a graveyard" (3)
@_eff(r"^you may exile target card from (?:a|your|an opponent'?s?) graveyard(?:\.|$)")
def _may_exile_from_gy(m):
    return Optional_(body=Exile(
        target=Filter(base="card", targeted=True),
    ))


# "you may untap this artifact" (4) / "you may untap it" (various)
@_eff(r"^you may untap (?:this artifact|this creature|this permanent|it|~)(?:\.|$)")
def _may_untap_self(m):
    return Optional_(body=UntapEffect(target=Filter(base="self", targeted=False)))


# "you may put a -1/-1 counter on target creature" (5)
@_eff(r"^you may put (" + _NUM_RE + r") ([+-]\d+/[+-]\d+) counters? on (target [^.]+?)(?:\.|$)")
def _may_put_counter(m):
    n = _n(m.group(1))
    kind = m.group(2)
    target = _parse_filter(m.group(3)) or TARGET_CREATURE
    return Optional_(body=CounterMod(op="put", count=n, counter_kind=kind, target=target))


# ============================================================================
# GROUP 20: Miscellaneous high-frequency promotions
# ============================================================================

# "you and target opponent each draw a card" (5)
@_eff(r"^you and target opponent each draw (" + _NUM_RE + r") cards?(?:\.|$)")
def _you_and_opp_draw(m):
    n = _n(m.group(1))
    return Sequence(items=(
        Draw(count=n, target=SELF),
        Draw(count=n, target=TARGET_OPPONENT),
    ))


# "target opponent reveals a card at random from their hand" (4)
@_eff(r"^target opponent reveals (" + _NUM_RE + r") cards? at random from their hand(?:\.|$)")
def _opp_reveal_random(m):
    from mtg_ast import Reveal
    n = _n(m.group(1))
    return Reveal(source="opponent_hand", actor="opponent", count=n)


# "shuffle the cards from your hand into your library, then draw that many cards" (4)
@_eff(r"^shuffle (?:the cards from )?your hand into your library,?\s*then draw that many cards(?:\.|$)")
def _shuffle_hand_draw(m):
    return Sequence(items=(
        Shuffle(target=SELF),
        Draw(count="that_many", target=SELF),
    ))


# "each player discards their hand, then draws [seven/that many] cards" (5)
@_eff(r"^each player discards their hand,?\s*then draws? (" + _NUM_RE + r"|that many) cards?(?:\.|$)")
def _wheel_effect(m):
    n = _n(m.group(1)) if m.group(1) != "that many" else "that_many"
    return Sequence(items=(
        Discard(count="all", target=EACH_PLAYER),
        Draw(count=n, target=EACH_PLAYER),
    ))


# "each player discards their hand then x" — Timetwister-style (5)
@_eff(r"^each player discards their hand(?:\.|$)")
def _each_discard_hand(m):
    return Discard(count="all", target=EACH_PLAYER)


# "look at that many cards from the top of your library" (4)
@_eff(r"^look at (?:that many|the top (" + _NUM_RE + r")) cards? (?:from (?:the top of )?)?(?:your|of your) library(?:\.|$)")
def _look_at_top(m):
    from mtg_ast import LookAt
    n = _n(m.group(1)) if m.group(1) else "that_many"
    return LookAt(target=SELF, zone="library_top_n", count=n)


# "untap this creature" — bare verb (the quoted "untap this creature." variant)
@_eff(r'^untap (?:this creature|this artifact|~)(?:\.")?(?:\.|$)')
def _untap_self(m):
    return UntapEffect(target=Filter(base="self", targeted=False))


# ============================================================================
# GROUP 21: Exile and return self (flicker/transform)
# ============================================================================
# "exile ~, then return it to the battlefield transformed under its owner's control" (6)
@_eff(r"^exile (?:~|this creature|this permanent),?\s*then return (?:it|~) to the battlefield(?:\s+transformed)?(?:\s+under (?:its owner'?s?|your) control)?(?:\.|$)")
def _flicker_self(m):
    return Sequence(items=(
        Exile(target=Filter(base="self", targeted=False)),
        Reanimate(query=Filter(base="self", targeted=False)),
    ))


# ============================================================================
# GROUP 22: Create typed tokens with keywords (complex patterns)
# ============================================================================
# "create a 3/2 colorless vehicle artifact token with crew 1" (3)
@_eff(r"^create (" + _NUM_RE + r" )?(\d+)/(\d+) ([a-z]+(?: [a-z]+)*?) (creature|artifact|enchantment) tokens? with ([^.]+?)(?:\.|$)")
def _create_token_with_keyword(m):
    n = _n(m.group(1)) if m.group(1) else 1
    pt = (int(m.group(2)), int(m.group(3)))
    types = tuple(m.group(4).strip().split()) + (m.group(5),)
    return CreateToken(count=n, pt=pt, types=types)


# "create that many N/N [color] [type] creature tokens" (3)
@_eff(r"^create that many (\d+)/(\d+) ([a-z]+(?: [a-z]+)*?) creature tokens?(?:\.|$)")
def _create_that_many_tokens(m):
    pt = (int(m.group(1)), int(m.group(2)))
    types = tuple(m.group(3).strip().split())
    return CreateToken(count="that_many", pt=pt, types=types)


# ============================================================================
# GROUP 23: Additional counter patterns
# ============================================================================
# "put a +1/+1 counter on this creature" / "put a +1/+1 counter on ~"
@_eff(r"^put (" + _NUM_RE + r") ([+-]\d+/[+-]\d+) counters? on (?:this creature|this permanent|~|it)(?:\.|$)")
def _put_counter_self(m):
    n = _n(m.group(1))
    kind = m.group(2)
    return CounterMod(op="put", count=n, counter_kind=kind,
                      target=Filter(base="self", targeted=False))


# "during your turn, put a +1/+1 counter on this creature" (4)
@_eff(r"^during your turn,\s*put (" + _NUM_RE + r") ([+-]\d+/[+-]\d+) counters? on (?:this creature|~)(?:\.|$)")
def _put_counter_self_during_turn(m):
    n = _n(m.group(1))
    kind = m.group(2)
    return CounterMod(op="put", count=n, counter_kind=kind,
                      target=Filter(base="self", targeted=False))


# "return that card to the battlefield under its owner's control with a +1/+1 counter on it" (3)
@_eff(r"^return (?:that card|it) to the battlefield under (?:its owner'?s?|your) control with (" + _NUM_RE + r") ([+-]\d+/[+-]\d+) counters? on it(?:\.|$)")
def _reanimate_with_counter(m):
    return Reanimate(query=Filter(base="that_card", targeted=False))


# ============================================================================
# GROUP 24: Various gain/loss patterns
# ============================================================================
# "you get that many {e}" — energy tokens (4)
@_eff(r"^you get (?:that many |(" + _NUM_RE + r") )?\{e\}(?:\{e\})*(?:\.|$)")
def _get_energy(m):
    return Modification(kind="energy_gain", args=(m.group(0),))


# "you may pay {e}{e}" — energy payment (4)
@_eff(r"^you may pay \{e\}(?:\{e\})*(?:\.|$)")
def _may_pay_energy(m):
    return Optional_(body=Modification(kind="energy_pay", args=(m.group(0),)))


# ============================================================================
# GROUP 25: Search library for specific permanent types
# ============================================================================
# "search your library for a rebel permanent card with mana value 3 or less,
#  put it onto the battlefield, then shuffle" (4)
@_eff(r"^search your library for (?:a|an) ([^,]+?) card(?:\s+with [^,]+)?,?\s*put (?:it|that card) onto the battlefield(?:\s+tapped)?,?\s*then shuffle(?:\.|$)")
def _tutor_to_battlefield(m):
    query = _parse_filter(m.group(1) + " card") or Filter(base="card")
    return Tutor(query=query, destination="battlefield", shuffle_after=True)


# ============================================================================
# GROUP 26: Miscellaneous verb-led patterns
# ============================================================================

# "you may tap any number of untapped creatures you control" (3)
@_eff(r"^you may tap any number of untapped creatures you control(?:\.|$)")
def _may_tap_any_number(m):
    return Optional_(body=TapEffect(target=Filter(
        base="creature", quantifier="any", you_control=True,
        extra=("untapped",),
    )))


# "you may cast target instant or sorcery card from your graveyard this turn" (3)
@_eff(r"^you may cast target ([^.]+?) card from your graveyard(?:\s+this turn)?(?:\.|$)")
def _may_cast_from_gy(m):
    return Optional_(body=Recurse(
        query=Filter(base="card", targeted=True),
        destination="cast_from_gy",
    ))


# "attach to target creature you control" (5)
@_eff(r"^attach (?:it |this equipment |this aura |~ )?to target ([^.]+?)(?:\.|$)")
def _attach_to_target(m):
    return Modification(kind="attach", args=(m.group(1),))


# "change target" (4) — redirect
@_eff(r"^change (?:the )?target(?:\.|$)")
def _change_target(m):
    return Modification(kind="change_target", args=())


# "gain control of <demonstrative> <type>" (5)
@_eff(r"^gain control of (?:target |that |it|the )(creature|permanent|artifact|enchantment|land)?[^.]*?(?:\.|$)")
def _gain_control(m):
    from mtg_ast import GainControl
    text_low = m.group(0).lower()
    if "target" in text_low:
        target = Filter(base=m.group(1) or "permanent", targeted=True)
    else:
        target = Filter(base="that_thing", targeted=False)
    return GainControl(target=target)


# "manifest dread, then attach this equipment to that creature" (4)
@_eff(r"^manifest dread(?:,?\s*then [^.]+)?(?:\.|$)")
def _manifest_dread(m):
    return Modification(kind="manifest_dread", args=(m.group(0),))


# ============================================================================
# GROUP 27: Conditional effects — "if [condition], [effect]"
# ============================================================================
# This is the single biggest remaining bucket (~880 instances, ~748 single-kind
# cards). The pattern "if <cond>, <effect>" is caught by a post-hook catch-all
# in a_wave_a_promotions.py that emits Modification(kind="conditional_effect").
# We preempt it by parsing the body via parse_effect() and wrapping in a typed
# Conditional node. If the body doesn't parse to typed, we return None to let
# the catch-all handle it (no regression).

@_eff(r"^if (.+?),\s+(.+?)(?:\.|$)")
def _conditional_typed(m):
    from mtg_ast import Conditional, Condition
    cond_text = m.group(1).strip()
    body_text = m.group(2).strip()
    # Avoid infinite recursion: only fire at the top level
    from parser import parse_effect, _parse_effect_depth
    if _parse_effect_depth[0] > 2:
        return None
    body = parse_effect(body_text)
    if body is None:
        return None
    # Only accept if the body is a typed node, not a Modification/UnknownEffect
    if isinstance(body, (Modification, UnknownEffect)):
        return None
    return Conditional(
        condition=Condition(kind="if", args=(cond_text,)),
        body=body,
    )


# "if [condition], [effect]. otherwise, [else_effect]"
@_eff(r"^if (.+?),\s+(.+?)\.\s+otherwise,?\s+(.+?)(?:\.|$)")
def _conditional_else_typed(m):
    from mtg_ast import Conditional, Condition
    cond_text = m.group(1).strip()
    body_text = m.group(2).strip()
    else_text = m.group(3).strip()
    from parser import parse_effect, _parse_effect_depth
    if _parse_effect_depth[0] > 2:
        return None
    body = parse_effect(body_text)
    else_body = parse_effect(else_text)
    if body is None or isinstance(body, (Modification, UnknownEffect)):
        return None
    if else_body is None or isinstance(else_body, (Modification, UnknownEffect)):
        return None
    return Conditional(
        condition=Condition(kind="if", args=(cond_text,)),
        body=body,
        else_body=else_body,
    )


# ============================================================================
# GROUP 28: "you may [verb effect]" — broad optional with typed body
# ============================================================================
# The broad "you may .+$" catch-all emits Modification(kind="optional_effect").
# We preempt for cases where the inner effect parses to a typed node.

@_eff(r"^you may (.+?)(?:\.|$)")
def _optional_typed(m):
    inner_text = m.group(1).strip()
    from parser import parse_effect, _parse_effect_depth
    if _parse_effect_depth[0] > 2:
        return None
    inner = parse_effect(inner_text)
    if inner is None or isinstance(inner, (Modification, UnknownEffect)):
        return None
    return Optional_(body=inner)


# ============================================================================
# GROUP 29: Investigate / create clue tokens
# ============================================================================
# "investigate" (59 instances as stub)
@_eff(r"^investigate(?:\.|$)")
def _investigate(m):
    return CreateToken(count=1, types=("clue",), pt=None)


# ============================================================================
# GROUP 30: Proliferate
# ============================================================================
# "proliferate" (41 instances as stub)
@_eff(r"^proliferate(?:\.|$)")
def _proliferate(m):
    return CounterMod(op="proliferate", count=1, counter_kind="any",
                      target=Filter(base="permanent", quantifier="all"))


# ============================================================================
# GROUP 31: Regenerate
# ============================================================================
# "regenerate ~" / "regenerate target creature" / "regenerate this creature"
@_eff(r"^regenerate (?:~|this creature|target creature)(?:\.|$)")
def _regenerate_self(m):
    text_low = m.group(0).lower()
    if "target" in text_low:
        target = TARGET_CREATURE
    else:
        target = Filter(base="self", targeted=False)
    return Modification(kind="regenerate_typed", args=(target,))


# ============================================================================
# GROUP 32: Become the monarch
# ============================================================================
# "you become the monarch" (36 instances)
@_eff(r"^you become the monarch(?:\.|$)")
def _become_monarch(m):
    return Modification(kind="become_monarch_typed", args=())


# ============================================================================
# GROUP 33: Scry N, then draw a card (common sequence)
# ============================================================================
@_eff(r"^scry (" + _NUM_RE + r"),?\s*then draw (" + _NUM_RE + r") cards?(?:\.|$)")
def _scry_then_draw(m):
    s = _n(m.group(1))
    d = _n(m.group(2))
    return Sequence(items=(
        Scry(count=s),
        Draw(count=d, target=SELF),
    ))


# ============================================================================
# GROUP 34: Transform self
# ============================================================================
# "transform ~" / "transform this creature" (166 instances)
@_eff(r"^transform (?:~|this creature|this permanent)(?:\.|$)")
def _transform_self(m):
    return Modification(kind="transform_self_typed", args=())


# ============================================================================
# GROUP 35: Flip a coin
# ============================================================================
@_eff(r"^flip a coin(?:\.|$)")
def _flip_coin(m):
    return Modification(kind="flip_coin_typed", args=())


# ============================================================================
# GROUP 36: Venture into the dungeon
# ============================================================================
@_eff(r"^venture into the dungeon(?:\.|$)")
def _venture(m):
    return Modification(kind="venture_dungeon_typed", args=())


# ============================================================================
# GROUP 37: Additional typed combat / keyword rules
# ============================================================================

# "~ gains [keyword(s)] until end of turn" — self gains keyword(s)
@_eff(r"^(?:~|this creature|this permanent) gains? ([a-z, ]+?) until end of turn(?:\.|$)")
def _self_gains_kw_eot(m):
    return GrantAbility(
        ability_name=m.group(1).strip(),
        target=Filter(base="self", targeted=False),
        duration="until_end_of_turn",
    )


# "target creature gets -N/-0 until end of turn" (debuff without toughness)
@_eff(r"^target (creature|permanent) gets ([+-]\d+)/([+-]\d+) until end of turn(?:\.|$)")
def _debuff_target_eot(m):
    return Buff(power=int(m.group(2)), toughness=int(m.group(3)),
                target=Filter(base=m.group(1), targeted=True),
                duration="until_end_of_turn")


# "each creature gets -N/-N until end of turn"
@_eff(r"^each creature gets ([+-]\d+)/([+-]\d+) until end of turn(?:\.|$)")
def _each_creature_debuff_eot(m):
    return Buff(power=int(m.group(1)), toughness=int(m.group(2)),
                target=Filter(base="creature", quantifier="all"),
                duration="until_end_of_turn")


# "other creatures get -N/-N until end of turn"
@_eff(r"^other creatures get ([+-]\d+)/([+-]\d+) until end of turn(?:\.|$)")
def _other_creature_debuff_eot(m):
    return Buff(power=int(m.group(1)), toughness=int(m.group(2)),
                target=Filter(base="creature", quantifier="all", extra=("other",)),
                duration="until_end_of_turn")


# ============================================================================
# GROUP 38: Reveal top card(s) of library
# ============================================================================
@_eff(r"^reveal the top (" + _NUM_RE + r") cards? of your library(?:\.|$)")
def _reveal_top(m):
    from mtg_ast import Reveal
    n = _n(m.group(1))
    return Reveal(source="library_top", actor="controller", count=n)


# "reveal the top card of your library"
@_eff(r"^reveal the top card of your library(?:\.|$)")
def _reveal_top_one(m):
    from mtg_ast import Reveal
    return Reveal(source="library_top", actor="controller", count=1)


# ============================================================================
# GROUP 39: Gain life equal to / lose life equal to
# ============================================================================
@_eff(r"^you gain life equal to [^.]+(?:\.|$)")
def _gain_life_equal(m):
    return GainLife(amount="var")


@_eff(r"^you lose life equal to [^.]+(?:\.|$)")
def _lose_life_equal(m):
    return LoseLife(amount="var")


@_eff(r"^target (?:player|opponent) loses life equal to [^.]+(?:\.|$)")
def _target_loses_life_equal(m):
    who = TARGET_OPPONENT if "opponent" in m.group(0).lower() else TARGET_PLAYER
    return LoseLife(amount="var", target=who)


# ============================================================================
# GROUP 40: Exile all [filter] / destroy all [filter]
# ============================================================================
@_eff(r"^exile all ([^.]+?)(?:\.|$)")
def _exile_all(m):
    target = _parse_filter("all " + m.group(1)) or Filter(base="permanent", quantifier="all")
    return Exile(target=target)


@_eff(r"^exile each ([^.]+?)(?:\.|$)")
def _exile_each(m):
    target = _parse_filter("each " + m.group(1)) or Filter(base="permanent", quantifier="each")
    return Exile(target=target)


# ============================================================================
# GROUP 41: Sacrifice a creature / permanent
# ============================================================================
@_eff(r"^sacrifice (" + _NUM_RE + r") (creatures?|permanents?|artifacts?|enchantments?|lands?)(?:\.|$)")
def _sacrifice_n(m):
    n = _n(m.group(1))
    base = m.group(2).lower().rstrip("s")
    return Sacrifice(query=Filter(base=base, you_control=True), actor="self")


# "sacrifice a creature" (bare)
@_eff(r"^sacrifice (?:a|an) ([a-z ]+?)(?:\.|$)")
def _sacrifice_one(m):
    base = m.group(1).strip().rstrip("s")
    if base in ("creature", "permanent", "artifact", "enchantment", "land",
                "token", "food", "treasure", "clue", "blood"):
        return Sacrifice(query=Filter(base=base, you_control=True), actor="self")
    return None


# ============================================================================
# GROUP 42: Put cards from hand on top/bottom of library
# ============================================================================
@_eff(r"^put (" + _NUM_RE + r") cards? from (?:your|their) hand on (?:top|bottom) of (?:your|their|its owner'?s?) library(?:\.|$)")
def _put_hand_to_library(m):
    return Modification(kind="hand_to_library", args=(m.group(0),))


# ============================================================================
# GROUP 43: Damage to each opponent / each player
# ============================================================================
@_eff(r"^(?:~|this creature|this permanent) deals (\d+) damage to each (?:opponent|player)(?:\.|$)")
def _self_dmg_each(m):
    text_low = m.group(0).lower()
    target = EACH_OPPONENT if "opponent" in text_low else EACH_PLAYER
    return Damage(amount=int(m.group(1)), target=target)


# "~ deals damage to each opponent equal to [X]"
@_eff(r"^(?:~|this creature) deals damage (?:equal )?to each (?:opponent|player) equal to [^.]+(?:\.|$)")
def _self_dmg_each_var(m):
    text_low = m.group(0).lower()
    target = EACH_OPPONENT if "opponent" in text_low else EACH_PLAYER
    return Damage(amount="var", target=target)


# ============================================================================
# GROUP 44: Copy target spell
# ============================================================================
@_eff(r"^copy target (?:instant or sorcery )?spell(?:\.|$)")
def _copy_spell(m):
    from mtg_ast import CopySpell
    return CopySpell(target=Filter(base="spell", targeted=True))


# "copy it" / "copy that spell"
@_eff(r"^copy (?:it|that spell)(?:\.|$)")
def _copy_that_spell(m):
    from mtg_ast import CopySpell
    return CopySpell(target=Filter(base="that_spell", targeted=False))


# ============================================================================
# GROUP 45: Return from exile / put onto battlefield from exile
# ============================================================================
@_eff(r"^return (?:the exiled cards?|all cards exiled (?:with|by) (?:~|this [a-z]+)) to (?:the battlefield|their owners?'? hands?)(?:\.|$)")
def _return_exiled(m):
    return Reanimate(query=Filter(base="exiled_card", targeted=False))


# ============================================================================
# GROUP 46: Target creature's power/toughness become N
# ============================================================================
@_eff(r"^target creature'?s? (?:base )?power and toughness (?:become|each become|are each) (\d+)/(\d+)(?:\s+until [^.]+)?(?:\.|$)")
def _set_pt(m):
    return Buff(power=int(m.group(1)), toughness=int(m.group(2)),
                target=TARGET_CREATURE,
                duration="set_base")


# ============================================================================
# GROUP 47: Amass N
# ============================================================================
@_eff(r"^amass (?:[a-z]+ )?(\d+)(?:\.|$)")
def _amass(m):
    return CreateToken(count=1, types=("zombie_army",), pt=(0, 0))


# ============================================================================
# GROUP 48: Populate
# ============================================================================
@_eff(r"^populate(?:\.|$)")
def _populate(m):
    return CreateToken(count=1, types=("copy_token",), pt=None)


# ============================================================================
# GROUP 49: Learn
# ============================================================================
@_eff(r"^learn(?:\.|$)")
def _learn(m):
    return Modification(kind="learn_typed", args=())


# ============================================================================
# GROUP 50: "each opponent" effect patterns
# ============================================================================
@_eff(r"^each opponent discards? (" + _NUM_RE + r") cards?(?:\.|$)")
def _each_opp_discard(m):
    n = _n(m.group(1))
    return Discard(count=n, target=EACH_OPPONENT)


@_eff(r"^each opponent sacrifices? (?:a|an) ([a-z ]+?)(?:\.|$)")
def _each_opp_sac(m):
    base = m.group(1).strip().rstrip("s")
    return Sacrifice(query=Filter(base=base), actor="each_opponent")


@_eff(r"^each opponent mills? (" + _NUM_RE + r") cards?(?:\.|$)")
def _each_opp_mill(m):
    n = _n(m.group(1))
    return Mill(count=n, target=EACH_OPPONENT)


# ============================================================================
# GROUP 51: Token creation — "create X [color] [type] tokens"
# ============================================================================
# "create X treasure tokens" (no pt)
@_eff(r"^create (" + _NUM_RE + r") (treasure|clue|food|blood|map|powerstone|incubator) tokens?(?:\.|$)")
def _create_utility_tokens(m):
    n = _n(m.group(1))
    ttype = m.group(2).lower()
    return CreateToken(count=n, types=(ttype,), pt=None)


# "create a [N/N] [color] [type] creature token"
# This extends the existing rule to handle more complex type lists
@_eff(r"^create (" + _NUM_RE + r" )?(\d+)/(\d+) ([a-z]+(?: [a-z]+)*?) tokens?(?:\.|$)")
def _create_generic_tokens(m):
    n = _n(m.group(1)) if m.group(1) else 1
    pt = (int(m.group(2)), int(m.group(3)))
    types = tuple(m.group(4).strip().split())
    return CreateToken(count=n, pt=pt, types=types)


# ============================================================================
# GROUP 52: "then shuffle" standalone tail
# ============================================================================
@_eff(r"^then shuffle(?:\.|$)")
def _then_shuffle(m):
    return Shuffle(target=SELF)


# ============================================================================
# GROUP 53: "that player" verb effects
# ============================================================================
# "that player discards N cards / a card at random" (4+)
@_eff(r"^that player discards? (" + _NUM_RE + r") cards?(?:\s+at random)?(?:\.|$)")
def _that_player_discards(m):
    n = _n(m.group(1))
    random = "random" in m.group(0).lower()
    return Discard(count=n,
                   target=Filter(base="that_player", targeted=False),
                   chosen_by="random" if random else "discarder")


# "that player discards a card at random"
@_eff(r"^that player discards a card at random(?:\.|$)")
def _that_player_discards_random(m):
    return Discard(count=1,
                   target=Filter(base="that_player", targeted=False),
                   chosen_by="random")


# "that player mills a card" (3)
@_eff(r"^that player mills? (" + _NUM_RE + r") cards?(?:\.|$)")
def _that_player_mills(m):
    n = _n(m.group(1))
    return Mill(count=n, target=Filter(base="that_player", targeted=False))


# "that player mills a card" (bare)
@_eff(r"^that player mills a card(?:\.|$)")
def _that_player_mills_one(m):
    return Mill(count=1, target=Filter(base="that_player", targeted=False))


# "that player discards two cards" (3)
@_eff(r"^that player discards (" + _NUM_RE + r") cards(?:\.|$)")
def _that_player_discards_n(m):
    n = _n(m.group(1))
    return Discard(count=n,
                   target=Filter(base="that_player", targeted=False),
                   chosen_by="discarder")


# ============================================================================
# GROUP 54: Discard a card at random
# ============================================================================
@_eff(r"^discard (" + _NUM_RE + r") cards? at random(?:\.|$)")
def _discard_random(m):
    n = _n(m.group(1))
    return Discard(count=n, target=SELF, chosen_by="random")


# "discard a card at random"
@_eff(r"^discard a card at random(?:\.|$)")
def _discard_one_random(m):
    return Discard(count=1, target=SELF, chosen_by="random")


# ============================================================================
# GROUP 55: Return enchanted/equipped creature to hand
# ============================================================================
@_eff(r"^return (?:enchanted|equipped) (creature|permanent) to its owner'?s? hand(?:\.|$)")
def _bounce_enchanted(m):
    base = m.group(1).lower()
    return Bounce(target=Filter(base=f"enchanted_{base}", targeted=False))


# ============================================================================
# GROUP 56: Each creature blocking gets debuff
# ============================================================================
@_eff(r"^each creature blocking (?:it|~|this creature) gets ([+-]\d+)/([+-]\d+) until end of turn(?:\.|$)")
def _blocking_debuff(m):
    return Buff(power=int(m.group(1)), toughness=int(m.group(2)),
                target=Filter(base="creature", quantifier="all",
                              extra=("blocking_this",)),
                duration="until_end_of_turn")


# ============================================================================
# GROUP 57: Exile a card from a graveyard
# ============================================================================
@_eff(r"^exile (" + _NUM_RE + r" )?cards? from (?:a|your|target (?:player|opponent)'?s?|each (?:player|opponent)'?s?) graveyard(?:\.|$)")
def _exile_from_gy(m):
    return Exile(target=Filter(base="card", targeted=False))


# ============================================================================
# GROUP 58: Put creature on top/bottom of library
# ============================================================================
@_eff(r"^put (?:~|this creature|this permanent|target [^.]+) on (?:the )?(top|bottom) of (?:its owner'?s?|your) library(?:\.|$)")
def _tuck_to_library(m):
    return Bounce(target=Filter(base="self", targeted=False),
                  to="top_of_library" if m.group(1) == "top" else "bottom_of_library")


# ============================================================================
# GROUP 59: "you may have it deal N damage"
# ============================================================================
@_eff(r"^you may have (?:it|this creature|~) deal (\d+) damage to (?:any target|target [^.]+)(?:\.|$)")
def _may_self_deals(m):
    return Optional_(body=Damage(amount=int(m.group(1)), target=TARGET_ANY))


# ============================================================================
# GROUP 60: More exile patterns
# ============================================================================
# "exile that many cards from the top of your library" (3)
@_eff(r"^exile (?:that many|the top (" + _NUM_RE + r")) cards? from (?:the top of )?your library(?:\.|$)")
def _exile_top_self(m):
    return Exile(target=SELF)


# ============================================================================
# GROUP 61: "its owner shuffles their graveyard into their library" (3)
# ============================================================================
@_eff(r"^its owner shuffles their graveyard into their library(?:\.|$)")
def _owner_shuffles_gy(m):
    return Shuffle(target=Filter(base="its_owner", targeted=False))


# ============================================================================
# GROUP 62: Various draw + secondary effect sequences
# ============================================================================
# "draw a card, then scry N" (3)
@_eff(r"^draw (" + _NUM_RE + r") cards?,\s*then scry (" + _NUM_RE + r")(?:\.|$)")
def _draw_then_scry(m):
    d = _n(m.group(1))
    s = _n(m.group(2))
    return Sequence(items=(
        Draw(count=d, target=SELF),
        Scry(count=s),
    ))


# "draw a card and reveal it" (3)
@_eff(r"^draw (" + _NUM_RE + r") cards? and reveal (?:it|them)(?:\.|$)")
def _draw_and_reveal(m):
    from mtg_ast import Reveal
    n = _n(m.group(1))
    return Sequence(items=(
        Draw(count=n, target=SELF),
        Reveal(source="hand", actor="controller", count=n),
    ))


# ============================================================================
# GROUP 63: Defender / combat keywords as effects
# ============================================================================
# "defending player gets a poison counter" (3)
@_eff(r"^defending player gets (" + _NUM_RE + r") poison counters?(?:\.|$)")
def _defender_poison(m):
    n = _n(m.group(1))
    return CounterMod(op="put", count=n, counter_kind="poison",
                      target=Filter(base="defending_player", targeted=False))


# "that player gets N poison counters" (3)
@_eff(r"^that player gets (" + _NUM_RE + r") poison counters?(?:\.|$)")
def _that_player_poison(m):
    n = _n(m.group(1))
    return CounterMod(op="put", count=n, counter_kind="poison",
                      target=Filter(base="that_player", targeted=False))


# ============================================================================
# GROUP 64: Return colored creature you control
# ============================================================================
@_eff(r"^return (?:a|an) ([a-z]+(?: or [a-z]+)?) (creature|permanent) you control to its owner'?s? hand(?:\.|$)")
def _bounce_colored_own(m):
    return Bounce(target=Filter(base=m.group(2), you_control=True, targeted=False))


# ============================================================================
# GROUP 65: Target creature this creature gains keyword
# ============================================================================
@_eff(r"^(?:~|this creature) gains? ([a-z ]+?) until (?:end of turn|your next turn)(?:\.|$)")
def _self_gains_kw_eot_v2(m):
    kw = m.group(1).strip()
    # Avoid matching "~ gains +N/+N" which is a buff
    if re.match(r'^[+-]', kw):
        return None
    return GrantAbility(
        ability_name=kw,
        target=Filter(base="self", targeted=False),
        duration="until_end_of_turn",
    )


# "~ gets +N/+N until end of turn" with variable amounts
@_eff(r"^(?:~|this creature) gets ([+-]\d+)/([+-]\d+) until end of turn(?:\.|$)")
def _self_gets_buff_eot(m):
    return Buff(power=int(m.group(1)), toughness=int(m.group(2)),
                target=Filter(base="self", targeted=False),
                duration="until_end_of_turn")


# "this creature gets +1/-1 or -1/+1 until end of turn" (3)
@_eff(r"^(?:~|this creature) gets \+(\d+)/-(\d+) or -(\d+)/\+(\d+) until end of turn(?:\.|$)")
def _self_gets_choice_buff(m):
    from mtg_ast import Choice
    return Choice(options=(
        Buff(power=int(m.group(1)), toughness=-int(m.group(2)),
             target=Filter(base="self", targeted=False)),
        Buff(power=-int(m.group(3)), toughness=int(m.group(4)),
             target=Filter(base="self", targeted=False)),
    ), pick=1)


# ============================================================================
# GROUP 66: Return from exile
# ============================================================================
@_eff(r"^return all cards exiled with (?:it|~|this [a-z]+) to (?:the battlefield|their owners?'? (?:hands?|control))(?:\.|$)")
def _return_exiled_with(m):
    return Reanimate(query=Filter(base="exiled_card", targeted=False))


# ============================================================================
# GROUP 67: "target creature you control explores" (3)
# ============================================================================
@_eff(r"^(?:target creature you control|~|this creature) explores(?:\.|$)")
def _explores(m):
    return Modification(kind="explore_typed", args=())
