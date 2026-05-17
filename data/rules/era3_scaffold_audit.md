# Era 3 (2020-2022) Scaffold-Gap Audit

- Total cards in dataset: **31963**

- Era distribution: era1=26932, era2=537, era3=798, era4=3696

- Era 3 cards: **798**

- Era 3 Condition nodes: **55** (bucketed 44, unbucketed 11, 20.0% gap)

- Era 3 Trigger nodes: **538**


## Top unbucketed condition Kinds

- `if` × 6
- `conditional` × 5

## Top unbucketed raw-text fragments (kind in raw/intervening_if/as_long_as)

- × 1: `you cast it`  _(e.g. Lutri, the Spellchaser)_
- × 1: `if they do, tap this creature and create a 1/1 blue fish creature token with "th`  _(e.g. Reservoir Kraken)_
- × 1: `this creature has two or fewer judgment counters on it`  _(e.g. Faithbound Judge // Sinner's Judgment)_
- × 1: `as long as this creature has three or more judgment counters on it, it can attac`  _(e.g. Faithbound Judge // Sinner's Judgment)_
- × 1: `there are three or more judgment counters on it`  _(e.g. Faithbound Judge // Sinner's Judgment)_
- × 1: `you've cast both a creature spell and a noncreature spell this turn`  _(e.g. Eshki Dragonclaw)_
- × 1: `if excess damage was dealt this way, note that excess damage, then you get a one`  _(e.g. Mephit's Enthusiasm)_
- × 1: `if ~ is saddled and a creature was dealt damage this way, that creature perpetua`  _(e.g. Switchgrass Grazer)_
- × 1: `it didn't have decayed`  _(e.g. Wilhelt, the Rotcleaver)_
- × 1: `if ~ was bargained, that card perpetually gains "this spell costs {2} more to ca`  _(e.g. Talion's Throneguard)_
- × 1: `a creature or planeswalker an opponent controlled was dealt excess damage this t`  _(e.g. Rith, Liberated Primeval)_

## Bucketed condition Kinds (sanity)

- `paid_optional_cost` × 16
- `for_each` × 7
- `if` × 6
- `conditional` × 6
- `did_prior_action` × 6
- `delirium` × 1
- `creature_died_this_turn` × 1
- `etb_tapped_unless` × 1

## Top trigger events

- `etb` × 130
- `phase` × 53
- `attack` × 43
- `combat_damage_player` × 28
- `cast_filtered` × 25
- `mutates` × 25
- `turned_face_up` × 20
- `exploits_creature` × 20
- `die` × 15
- `enter_or_attack` × 13
- `cast_any` × 9
- `specialize_creature` × 8
- `you_attack` × 7
- `beginning_of_ordinal_step` × 7
- `becomes_target` × 6
- `cast_spell` × 6
- `another_typed_enters` × 5
- `etb_as` × 5
- `when_you_do` × 4
- `ally_exploits` × 4
- `sacrifice_filtered` × 4
- `dealt_damage` × 3
- `group_combat_damage_player` × 3
- `one_or_more_typed_event` × 3
- `gain_life` × 3
- `creature_dies` × 3
- `another_typed_dies` × 3
- `tribe_you_control_etb` × 2
- `sacrifices` × 2
- `remove_counter` × 2
- `on_cast_creature` × 2
- `card_put_into_zone` × 2
- `cast_color_spell` × 2
- `self_and` × 2
- `becomes_blocked` × 2
- `unlock_door` × 2
- `self_combat_damage` × 2
- `etb_or_another` × 2
- `you_whenever` × 2
- `another_creature_enters` × 2