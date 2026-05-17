#!/usr/bin/env python3
"""Era 4 (2023-2026) scaffold-gap audit.

Mirrors scripts/era1_scaffold_audit.py but filters for cards classified as
era 4 by classifyCardEra (discover, descend, battles, prototype, craft,
role token, finality counter, the ring).

Output: data/rules/era4_scaffold_audit.md  +  prints top-50 to stdout.
"""

from __future__ import annotations

import json
import re
from collections import Counter, defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DATASET = ROOT / "data" / "rules" / "ast_dataset.jsonl"
OUT = ROOT / "data" / "rules" / "era4_scaffold_audit.md"

ERA4_KW = ["discover", "descend", "battle", "prototype", "craft",
           "role token", "finality counter", "the ring"]
ERA3_KW = ["daybound", "nightbound", "disturb", "cleave", "decayed",
           "exploit", "companion", "mutate", "foretell", "learn",
           "ward", "perpetual", "conjure"]
ERA2_KW = ["partner", "experience counter", "eminence", "energy counter",
           "crew", "adapt", "amass", "afterlife", "spectacle", "riot"]

TARGET_ERA = 4


def classify_era(oracle_text: str, type_line: str) -> int:
    text = (oracle_text or "").lower()
    types = (type_line or "").lower()
    for kw in ERA4_KW:
        if kw in text or kw in types:
            return 4
    for kw in ERA3_KW:
        if kw in text:
            return 3
    for kw in ERA2_KW:
        if kw in text:
            return 2
    return 1


BUCKETED_KINDS = {
    "fateful_hour", "life_threshold",
    "threshold", "card_count_zone",
    "metalcraft", "morbid", "ferocious", "revolt", "delirium",
    "devotion",
    "you_attacked_this_turn",
    "you_control",
    "paid_optional_cost", "for_each",
    "etb_as", "enters_as", "enters_with",
    "did_prior_action",
    "mana_spent",
    "was_kicked",
    "hellbent", "raid", "attacked_this_turn",
    "spell_mastery", "gained_life_this_turn", "creature_died_this_turn",
    "no_spells_cast_last_turn", "two_plus_spells_cast_last_turn",
    "you_control_creature_power_ge",
    "etb_tapped_unless", "domain", "etb_if", "repeat_n",
    "lieutenant", "ki_counters_ge_2", "self_is_tapped",
    "attacked_or_blocked_this_combat", "coven", "self_has_counter",
    "didnt_attack_this_turn", "dealt_damage_to_opponent_this_turn",
    "no_mana_spent_to_cast",
    # Era 4 audit additions (dev/era2-era4-scaffolds branch).
    "landfall", "you_descended_this_turn",
    "it_was_a_creature", "no_creatures_on_battlefield",
}

RAW_KINDS = {"intervening_if", "as_long_as", "conditional", "raw", "if"}

# Era 4-flavored raw patterns: discover, descend, battle defeated, prototype,
# craft, role tokens, the ring, finality counters, plus the carry-over
# evergreens.
RAW_PATTERNS = [
    ("discover_n", re.compile(r"discover \d")),
    ("descended_turn", re.compile(r"descend(?:ed)? this turn|fully descended|you have descend")),
    ("battle_defeated", re.compile(r"defeated (?:this turn|a battle)|battle.*defeated|defeat (?:a )?battle")),
    ("battle_attack", re.compile(r"attacks? (?:a battle|battle)")),
    ("prototype_cast", re.compile(r"prototype|cast .* for its prototype")),
    ("craft_with", re.compile(r"craft with|crafted")),
    ("role_token", re.compile(r"role token|monster role|wicked role|young hero|virtuous role|sorcerer role|cursed role|royal role")),
    ("the_ring_tempts", re.compile(r"the ring tempts you|ring-bearer|your ring-bearer")),
    ("finality_counter", re.compile(r"finality counter")),
    ("bargain_perm", re.compile(r"bargain|you may sacrifice an artifact, enchantment, or token")),
    ("celebration_perm_etb", re.compile(r"celebration|two or more nonland permanents")),
    ("commit_a_crime", re.compile(r"commit(?:s|ted)? a crime|committed a crime this turn")),
    ("collect_evidence", re.compile(r"collect evidence \d|collected evidence")),
    ("plot_exile", re.compile(r"plot|cast it without paying its mana cost.*plot")),
    ("disguise_facedown", re.compile(r"disguise|face-down|turn.*face up.*for its disguise")),
    ("cloak_facedown", re.compile(r"cloak|cloaked|the top card.*face down")),
    ("offspring_token", re.compile(r"offspring|1/1 copy")),
    ("freerunning", re.compile(r"freerunning|its freerunning cost")),
    ("expend_n", re.compile(r"expend \d")),
    ("eerie_room", re.compile(r"eerie|whenever an enchantment.*enters")),
    ("survival_tapped", re.compile(r"survival|untapped creature you control")),
    ("flurry_second_spell", re.compile(r"flurry|second spell each turn")),
    ("manifest_dread", re.compile(r"manifest dread")),
    ("warp_cast", re.compile(r"warp|its warp cost")),
    # Era 4 scaffold additions (dev/era2-era4-scaffolds branch).
    ("planeswalker_etb_turn", re.compile(r"planeswalker (?:entered|enters) the battlefield.*this turn|planeswalker.*you've cast.*this turn")),
    ("artifact_etb_turn", re.compile(r"artifact (?:entered|enters) the battlefield.*this turn.*(?:under your control|you control)")),
    ("reveal_land_otherwise_hand", re.compile(r"if it's a land card.*(?:onto the battlefield|put it onto).*otherwise")),
    ("had_counters_on_it", re.compile(r"had (?:a |one or more |any |a \+1/\+1 |a death )?counters? on it")),
    ("you_cast_from_hand", re.compile(r"you cast it from your hand|you cast it from a graveyard|you cast it(?! from)")),
    ("still_on_battlefield", re.compile(r"it's on the battlefield|is still on the battlefield")),
    ("hellbent_no_hand", re.compile(r"hellbent|no cards in.*hand")),
    ("monarch", re.compile(r"the monarch|you('re| are) the monarch")),
    ("initiative", re.compile(r"the initiative|have initiative")),
    ("revolt_perm_left", re.compile(r"revolt|permanent.*left the battlefield.*this turn")),
    ("delirium_types", re.compile(r"delirium|four or more card types")),
    ("metalcraft_three_art", re.compile(r"metalcraft|three or more artifacts")),
    ("ferocious_power4", re.compile(r"ferocious|creature with power (?:four|4)")),
    ("formidable", re.compile(r"formidable|total power (?:eight|8)")),
    ("kicker_paid", re.compile(r"kicked|kicker (?:cost )?was paid|multikicker")),
    ("gained_life_turn", re.compile(r"(?:gained|gain) life.*this turn")),
    ("life_lost_turn", re.compile(r"(?:lost|lose) life.*this turn")),
    ("opponent_lost_life", re.compile(r"opponent.*(?:lost life|lose life|dealt damage).*this turn")),
    ("life_above", re.compile(r"(?:you have|your life total is) \d+ or more life")),
    ("life_below", re.compile(r"(?:you have|your life total is) \d+ or (?:less|fewer) life")),
    ("cast_spell_turn", re.compile(r"(?:cast a spell|cast a noncreature|cast an instant|you cast).*this turn")),
    ("second_spell_turn", re.compile(r"(?:second|third) (?:spell|creature|noncreature|instant|sorcery|artifact|enchantment)")),
    ("creature_etb_turn", re.compile(r"creature (?:entered|enters).*(?:this turn|battlefield)")),
    ("drawn_card_turn", re.compile(r"(?:drew|drawn|draw) a card.*this turn")),
    ("attacked_turn", re.compile(r"attacked this turn|you attacked|creature attacked")),
    ("sacrificed_turn", re.compile(r"sacrific.*this turn")),
    ("combat_damage", re.compile(r"combat damage.*(?:this turn|to a player|dealt)")),
    ("landfall_turn", re.compile(r"landfall|land.*entered|played a land.*this turn")),
    ("discarded_turn", re.compile(r"discard.*this turn")),
    ("enchanted_creature", re.compile(r"enchanted creature")),
    ("equipped_creature", re.compile(r"equipped creature")),
    ("any_player_phase", re.compile(r"(?:each player|each opponent).*(?:upkeep|end step)")),
    ("upkeep", re.compile(r"upkeep")),
    ("died_this_turn", re.compile(r"died this turn")),
    ("graveyard_creatures", re.compile(r"graveyard.*creature card|creatures in.*graveyard")),
    ("graveyard_card", re.compile(r"graveyard")),
    ("mana_spent_raw", re.compile(r"mana was (?:spent|paid)|amount of mana (?:spent|paid)|mana value of \S+ is \d+ or (?:greater|more)")),
    ("permanent_left_bf", re.compile(r"permanent left")),
    ("becomes_tapped", re.compile(r"becomes tapped|is tapped")),
    ("becomes_target", re.compile(r"becomes (?:the|a) target")),
    ("tokens_created", re.compile(r"tokens.*created this turn")),
    ("you_control_raw", re.compile(r"you control")),
]


def match_raw(text: str) -> str | None:
    t = text.lower()
    for name, rx in RAW_PATTERNS:
        if rx.search(t):
            return name
    return None


def walk(node, conds, trigs):
    if isinstance(node, dict):
        t = node.get("__ast_type__")
        if t == "Condition":
            conds.append(node)
        elif t == "Trigger":
            trigs.append(node)
        for v in node.values():
            walk(v, conds, trigs)
    elif isinstance(node, list):
        for x in node:
            walk(x, conds, trigs)


def main():
    cond_kinds = Counter()
    cond_kinds_bucketed = Counter()
    cond_kinds_unbucketed = Counter()
    trig_events = Counter()
    raw_buckets = Counter()
    raw_unbucketed_text = Counter()
    raw_unbucketed_examples = defaultdict(list)

    total_cards = 0
    target_cards = 0
    era_counts = Counter()

    with DATASET.open("r", encoding="utf-8") as f:
        for line in f:
            if not line.strip():
                continue
            row = json.loads(line)
            total_cards += 1
            era = classify_era(row.get("oracle_text", ""), row.get("type_line", ""))
            era_counts[era] += 1
            if era != TARGET_ERA:
                continue
            target_cards += 1
            ast = row.get("ast")
            if not ast:
                continue
            conds, trigs = [], []
            walk(ast, conds, trigs)
            name = row.get("name", "?")
            for c in conds:
                k = (c.get("kind") or "").lower()
                cond_kinds[k] += 1
                if k in BUCKETED_KINDS:
                    cond_kinds_bucketed[k] += 1
                elif k in RAW_KINDS:
                    args = c.get("args") or []
                    raw_txt = ""
                    if args and isinstance(args[0], str):
                        raw_txt = args[0]
                    bucket = match_raw(raw_txt) if raw_txt else None
                    if bucket:
                        raw_buckets[bucket] += 1
                        cond_kinds_bucketed[k] += 1
                    else:
                        cond_kinds_unbucketed[k] += 1
                        key = (raw_txt[:80] or "<empty>").lower()
                        raw_unbucketed_text[key] += 1
                        if len(raw_unbucketed_examples[key]) < 3:
                            raw_unbucketed_examples[key].append(name)
                else:
                    cond_kinds_unbucketed[k] += 1
            for tg in trigs:
                ev = (tg.get("event") or "").lower()
                trig_events[ev] += 1

    total_conds = sum(cond_kinds.values())
    total_bucket = sum(cond_kinds_bucketed.values())
    total_unbucket = sum(cond_kinds_unbucketed.values())

    lines = []
    lines.append(f"# Era {TARGET_ERA} (2023-2026) Scaffold-Gap Audit\n")
    lines.append(f"- Total cards in dataset: **{total_cards}**\n")
    lines.append(f"- Era distribution: " +
                 ", ".join(f"era{e}={era_counts[e]}" for e in sorted(era_counts)) + "\n")
    lines.append(f"- Era {TARGET_ERA} cards: **{target_cards}**\n")
    lines.append(f"- Era {TARGET_ERA} Condition nodes: **{total_conds}** "
                 f"(bucketed {total_bucket}, unbucketed {total_unbucket}, "
                 f"{100.0*total_unbucket/max(1,total_conds):.1f}% gap)\n")
    lines.append(f"- Era {TARGET_ERA} Trigger nodes: **{sum(trig_events.values())}**\n")

    lines.append("\n## Top unbucketed condition Kinds\n")
    for k, n in cond_kinds_unbucketed.most_common(60):
        lines.append(f"- `{k or '<empty>'}` × {n}")

    lines.append("\n## Top unbucketed raw-text fragments (kind in raw/intervening_if/as_long_as)\n")
    for txt, n in raw_unbucketed_text.most_common(60):
        ex = ", ".join(raw_unbucketed_examples[txt][:3])
        lines.append(f"- × {n}: `{txt}`  _(e.g. {ex})_")

    lines.append("\n## Bucketed condition Kinds (sanity)\n")
    for k, n in cond_kinds_bucketed.most_common(20):
        lines.append(f"- `{k}` × {n}")

    lines.append("\n## Top trigger events\n")
    for ev, n in trig_events.most_common(40):
        lines.append(f"- `{ev or '<empty>'}` × {n}")

    OUT.write_text("\n".join(lines))
    print("\n".join(lines))


if __name__ == "__main__":
    main()
