# Chaos Gauntlet Report

Generated: 2026-05-17T10:36:47-07:00

## Configuration

| Parameter | Value |
|-----------|-------|
| Oracle Corpus | 36656 cards |
| Legendary Creatures | 3433 |
| Total Games | 1000 |
| Seed | 42 |
| Permutations | 1 |
| Seats | 4 |
| Max Turns | 60 |
| Nightmare Boards | 10000 |

## Summary

### Chaos Games

| Metric | Count |
|--------|-------|
| Duration | 38.37s |
| Throughput | 26 games/sec |
| Crashes | 0 (in 0 games) |
| Invariant Violations | 480 (in 15 games) |
| Clean Games | 985 |

### Nightmare Boards

| Metric | Count |
|--------|-------|
| Duration | 2.263s |
| Throughput | 4419 boards/sec |
| Crashes | 0 |
| Invariant Violations | 2 |
| Clean Boards | 9999 |

## Invariant Violations (Chaos Games)

### By Invariant

| Invariant | Count |
|-----------|-------|
| AttachmentConsistency | 6 |
| CardIdentity | 392 |
| ZoneConservation | 78 |
| TriggerCompleteness | 4 |

### Violation Details (first 30)

#### Violation 1

- **Game**: 108 (seed 1080043, perm 0)
- **Invariant**: AttachmentConsistency
- **Turn**: 54, Phase=combat Step=end_of_combat
- **Commanders**: Tersa Lightshatter, Yargle, Glutton of Urborg, A-Phylath, World Sculptor, Sauron, Lord of the Rings
- **Message**: AttachmentConsistency: "Inventor's Axe" (seat 2) is attached to "creature token green elf warrior Token" which is not on any battlefield

<details>
<summary>Game State</summary>

```
Turn 54, Phase=combat Step=end_of_combat Active=seat2
Stack: 0 items, EventLog: 2662 events
  Seat 0 [LOST]: life=-19 library=72 hand=7 graveyard=10 exile=0 battlefield=0 cmdzone=1 mana=0
  Seat 1 [LOST]: life=-6 library=80 hand=5 graveyard=7 exile=0 battlefield=0 cmdzone=0 mana=0
  Seat 2 [WON]: life=7 library=77 hand=6 graveyard=5 exile=1 battlefield=10 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Inventor's Axe (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Dust Bowl (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - A-Phylath, World Sculptor (P/T 5/5, dmg=0)
  Seat 3 [LOST]: life=-23 library=74 hand=4 graveyard=5 exile=0 battlefield=0 cmdzone=1 mana=0

```

</details>

<details>
<summary>Recent Events</summary>

```
[2642] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2643] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2644] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2645] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2646] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2647] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2648] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2649] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2650] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2651] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2652] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2653] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2654] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2655] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2656] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2657] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2658] sba_704_5a seat=0 source= amount=-19
[2659] sba_cycle_complete seat=-1 source=
[2660] seat_eliminated seat=0 source= amount=38
[2661] game_end seat=2 source=
```

</details>

#### Violation 2

- **Game**: 108 (seed 1080043, perm 0)
- **Invariant**: AttachmentConsistency
- **Turn**: 54, Phase=combat Step=end_of_combat
- **Commanders**: Tersa Lightshatter, Yargle, Glutton of Urborg, A-Phylath, World Sculptor, Sauron, Lord of the Rings
- **Message**: AttachmentConsistency: "Inventor's Axe" (seat 2) is attached to "creature token green elf warrior Token" which is not on any battlefield

<details>
<summary>Game State</summary>

```
Turn 54, Phase=combat Step=end_of_combat Active=seat2
Stack: 0 items, EventLog: 2662 events
  Seat 0 [LOST]: life=-19 library=72 hand=7 graveyard=10 exile=0 battlefield=0 cmdzone=1 mana=0
  Seat 1 [LOST]: life=-6 library=80 hand=5 graveyard=7 exile=0 battlefield=0 cmdzone=0 mana=0
  Seat 2 [WON]: life=7 library=77 hand=6 graveyard=5 exile=1 battlefield=10 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Inventor's Axe (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Dust Bowl (P/T 0/0, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - A-Phylath, World Sculptor (P/T 5/5, dmg=0)
  Seat 3 [LOST]: life=-23 library=74 hand=4 graveyard=5 exile=0 battlefield=0 cmdzone=1 mana=0

```

</details>

<details>
<summary>Recent Events</summary>

```
[2642] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2643] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2644] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2645] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2646] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2647] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2648] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2649] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2650] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2651] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2652] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2653] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2654] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2655] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2656] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2657] damage seat=2 source=creature token green elf warrior Token amount=1 target=seat0
[2658] sba_704_5a seat=0 source= amount=-19
[2659] sba_cycle_complete seat=-1 source=
[2660] seat_eliminated seat=0 source= amount=38
[2661] game_end seat=2 source=
```

</details>

#### Violation 3

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 31, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 31, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1216 events
  Seat 0 [alive]: life=35 library=82 hand=3 graveyard=6 exile=0 battlefield=9 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=32 library=85 hand=1 graveyard=9 exile=0 battlefield=4 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=11 library=80 hand=3 graveyard=3 exile=0 battlefield=9 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1196] zone_change seat=1 source=Conduit Pylons
[1197] activate_ability seat=1 source=Adric, Mathematical Genius target=seat0
[1198] stack_push seat=1 source=Adric, Mathematical Genius target=seat0
[1199] priority_pass seat=2 source= target=seat0
[1200] priority_pass seat=3 source= target=seat0
[1201] priority_pass seat=0 source= target=seat0
[1202] stack_resolve seat=1 source=Adric, Mathematical Genius target=seat0
[1203] counter_spell_fizzle seat=1 source=generic_counter target=seat0
[1204] activated_ability_resolved seat=1 source=Adric, Mathematical Genius target=seat0
[1205] sacrifice seat=1 source=Bane, Lord of Darkness target=seat1
[1206] zone_change seat=1 source=Bane, Lord of Darkness
[1207] activate_ability seat=1 source=Adric, Mathematical Genius target=seat0
[1208] stack_push seat=1 source=Adric, Mathematical Genius target=seat0
[1209] priority_pass seat=2 source= target=seat0
[1210] priority_pass seat=3 source= target=seat0
[1211] priority_pass seat=0 source= target=seat0
[1212] stack_resolve seat=1 source=Adric, Mathematical Genius target=seat0
[1213] counter_spell_fizzle seat=1 source=generic_counter target=seat0
[1214] activated_ability_resolved seat=1 source=Adric, Mathematical Genius target=seat0
[1215] state seat=1 source= target=seat0
```

</details>

#### Violation 4

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 31, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 31, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1216 events
  Seat 0 [alive]: life=35 library=82 hand=3 graveyard=6 exile=0 battlefield=9 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=32 library=85 hand=1 graveyard=9 exile=0 battlefield=4 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=11 library=80 hand=3 graveyard=3 exile=0 battlefield=9 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1196] zone_change seat=1 source=Conduit Pylons
[1197] activate_ability seat=1 source=Adric, Mathematical Genius target=seat0
[1198] stack_push seat=1 source=Adric, Mathematical Genius target=seat0
[1199] priority_pass seat=2 source= target=seat0
[1200] priority_pass seat=3 source= target=seat0
[1201] priority_pass seat=0 source= target=seat0
[1202] stack_resolve seat=1 source=Adric, Mathematical Genius target=seat0
[1203] counter_spell_fizzle seat=1 source=generic_counter target=seat0
[1204] activated_ability_resolved seat=1 source=Adric, Mathematical Genius target=seat0
[1205] sacrifice seat=1 source=Bane, Lord of Darkness target=seat1
[1206] zone_change seat=1 source=Bane, Lord of Darkness
[1207] activate_ability seat=1 source=Adric, Mathematical Genius target=seat0
[1208] stack_push seat=1 source=Adric, Mathematical Genius target=seat0
[1209] priority_pass seat=2 source= target=seat0
[1210] priority_pass seat=3 source= target=seat0
[1211] priority_pass seat=0 source= target=seat0
[1212] stack_resolve seat=1 source=Adric, Mathematical Genius target=seat0
[1213] counter_spell_fizzle seat=1 source=generic_counter target=seat0
[1214] activated_ability_resolved seat=1 source=Adric, Mathematical Genius target=seat0
[1215] state seat=1 source= target=seat0
```

</details>

#### Violation 5

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 32, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 32, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1269 events
  Seat 0 [alive]: life=35 library=82 hand=3 graveyard=6 exile=0 battlefield=9 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=33 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=80 hand=3 graveyard=3 exile=0 battlefield=9 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1249] priority_pass seat=3 source= target=seat0
[1250] priority_pass seat=1 source= target=seat0
[1251] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1252] priority_pass seat=3 source= target=seat0
[1253] priority_pass seat=0 source= target=seat0
[1254] priority_pass seat=1 source= target=seat0
[1255] phase_step seat=2 source= target=seat0
[1256] priority_pass seat=3 source= target=seat0
[1257] priority_pass seat=0 source= target=seat0
[1258] priority_pass seat=1 source= target=seat0
[1259] stack_resolve seat=2 source=Syr Ginger, the Meal Ender target=seat0
[1260] enter_battlefield seat=2 source=Syr Ginger, the Meal Ender target=seat0
[1261] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1262] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1263] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1264] priority_pass seat=2 source= target=seat0
[1265] priority_pass seat=3 source= target=seat0
[1266] priority_pass seat=1 source= target=seat0
[1267] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1268] state seat=2 source= target=seat0
```

</details>

#### Violation 6

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 32, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 32, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1269 events
  Seat 0 [alive]: life=35 library=82 hand=3 graveyard=6 exile=0 battlefield=9 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=33 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=80 hand=3 graveyard=3 exile=0 battlefield=9 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1249] priority_pass seat=3 source= target=seat0
[1250] priority_pass seat=1 source= target=seat0
[1251] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1252] priority_pass seat=3 source= target=seat0
[1253] priority_pass seat=0 source= target=seat0
[1254] priority_pass seat=1 source= target=seat0
[1255] phase_step seat=2 source= target=seat0
[1256] priority_pass seat=3 source= target=seat0
[1257] priority_pass seat=0 source= target=seat0
[1258] priority_pass seat=1 source= target=seat0
[1259] stack_resolve seat=2 source=Syr Ginger, the Meal Ender target=seat0
[1260] enter_battlefield seat=2 source=Syr Ginger, the Meal Ender target=seat0
[1261] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1262] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1263] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1264] priority_pass seat=2 source= target=seat0
[1265] priority_pass seat=3 source= target=seat0
[1266] priority_pass seat=1 source= target=seat0
[1267] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1268] state seat=2 source= target=seat0
```

</details>

#### Violation 7

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 33, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 33, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 1365 events
  Seat 0 [alive]: life=30 library=82 hand=3 graveyard=6 exile=0 battlefield=9 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=33 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1345] stack_push seat=3 source=Viscerid Deepwalker target=seat0
[1346] priority_pass seat=0 source= target=seat0
[1347] priority_pass seat=1 source= target=seat0
[1348] priority_pass seat=2 source= target=seat0
[1349] stack_resolve seat=3 source=Viscerid Deepwalker target=seat0
[1350] buff seat=0 source=Viscerid Deepwalker amount=1 target=seat0
[1351] activated_ability_resolved seat=3 source=Viscerid Deepwalker target=seat0
[1352] phase_step seat=3 source= target=seat0
[1353] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1354] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1355] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1356] priority_pass seat=3 source= target=seat0
[1357] priority_pass seat=1 source= target=seat0
[1358] priority_pass seat=2 source= target=seat0
[1359] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1360] declare_attackers seat=3 source= target=seat0
[1361] blockers seat=0 source= target=seat0
[1362] damage seat=3 source=Canoptek Tomb Sentinel amount=4 target=seat0
[1363] phase_step seat=3 source= target=seat0
[1364] state seat=3 source= target=seat0
```

</details>

#### Violation 8

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 33, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 33, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 1365 events
  Seat 0 [alive]: life=30 library=82 hand=3 graveyard=6 exile=0 battlefield=9 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=33 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1345] stack_push seat=3 source=Viscerid Deepwalker target=seat0
[1346] priority_pass seat=0 source= target=seat0
[1347] priority_pass seat=1 source= target=seat0
[1348] priority_pass seat=2 source= target=seat0
[1349] stack_resolve seat=3 source=Viscerid Deepwalker target=seat0
[1350] buff seat=0 source=Viscerid Deepwalker amount=1 target=seat0
[1351] activated_ability_resolved seat=3 source=Viscerid Deepwalker target=seat0
[1352] phase_step seat=3 source= target=seat0
[1353] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1354] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1355] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1356] priority_pass seat=3 source= target=seat0
[1357] priority_pass seat=1 source= target=seat0
[1358] priority_pass seat=2 source= target=seat0
[1359] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1360] declare_attackers seat=3 source= target=seat0
[1361] blockers seat=0 source= target=seat0
[1362] damage seat=3 source=Canoptek Tomb Sentinel amount=4 target=seat0
[1363] phase_step seat=3 source= target=seat0
[1364] state seat=3 source= target=seat0
```

</details>

#### Violation 9

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 34, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 34, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 1429 events
  Seat 0 [alive]: life=30 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=30 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1409] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1410] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1411] priority_pass seat=1 source= target=seat0
[1412] priority_pass seat=2 source= target=seat0
[1413] priority_pass seat=3 source= target=seat0
[1414] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1415] phase_step seat=0 source= target=seat0
[1416] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1417] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1418] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1419] priority_pass seat=1 source= target=seat0
[1420] priority_pass seat=2 source= target=seat0
[1421] priority_pass seat=3 source= target=seat0
[1422] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1423] per_card_handler seat=0 source=Ezuri, Claw of Progress target=seat0
[1424] declare_attackers seat=0 source= target=seat0
[1425] blockers seat=2 source= target=seat0
[1426] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat2
[1427] phase_step seat=0 source= target=seat0
[1428] state seat=0 source= target=seat0
```

</details>

#### Violation 10

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 34, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 battlefield

<details>
<summary>Game State</summary>

```
Turn 34, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 1429 events
  Seat 0 [alive]: life=30 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=82 hand=4 graveyard=9 exile=0 battlefield=3 cmdzone=1 mana=0
    - Sceptre of Eternal Glory (P/T 0/0, dmg=0)
    - Wanderlight Spirit (P/T 2/3, dmg=0) [T]
    - Adric, Mathematical Genius (P/T 1/1, dmg=0)
  Seat 2 [alive]: life=30 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1409] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1410] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1411] priority_pass seat=1 source= target=seat0
[1412] priority_pass seat=2 source= target=seat0
[1413] priority_pass seat=3 source= target=seat0
[1414] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1415] phase_step seat=0 source= target=seat0
[1416] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1417] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1418] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1419] priority_pass seat=1 source= target=seat0
[1420] priority_pass seat=2 source= target=seat0
[1421] priority_pass seat=3 source= target=seat0
[1422] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1423] per_card_handler seat=0 source=Ezuri, Claw of Progress target=seat0
[1424] declare_attackers seat=0 source= target=seat0
[1425] blockers seat=2 source= target=seat0
[1426] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat2
[1427] phase_step seat=0 source= target=seat0
[1428] state seat=0 source= target=seat0
```

</details>

#### Violation 11

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 35, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 35, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1482 events
  Seat 0 [alive]: life=30 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=30 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1462] play_land seat=1 source=Island target=seat0
[1463] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1464] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1465] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1466] priority_pass seat=1 source= target=seat0
[1467] priority_pass seat=2 source= target=seat0
[1468] priority_pass seat=3 source= target=seat0
[1469] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1470] add_mana seat=1 source=Island amount=1 target=seat0
[1471] phase_step seat=1 source= target=seat0
[1472] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1473] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1474] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1475] priority_pass seat=1 source= target=seat0
[1476] priority_pass seat=2 source= target=seat0
[1477] priority_pass seat=3 source= target=seat0
[1478] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1479] phase_step seat=1 source= target=seat0
[1480] pool_drain seat=1 source= amount=1 target=seat0
[1481] state seat=1 source= target=seat0
```

</details>

#### Violation 12

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 35, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 35, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1482 events
  Seat 0 [alive]: life=30 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=30 library=84 hand=1 graveyard=9 exile=0 battlefield=6 cmdzone=0 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Syr Ginger, the Meal Ender (P/T 3/1, dmg=0)
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1462] play_land seat=1 source=Island target=seat0
[1463] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1464] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1465] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1466] priority_pass seat=1 source= target=seat0
[1467] priority_pass seat=2 source= target=seat0
[1468] priority_pass seat=3 source= target=seat0
[1469] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1470] add_mana seat=1 source=Island amount=1 target=seat0
[1471] phase_step seat=1 source= target=seat0
[1472] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1473] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1474] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1475] priority_pass seat=1 source= target=seat0
[1476] priority_pass seat=2 source= target=seat0
[1477] priority_pass seat=3 source= target=seat0
[1478] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1479] phase_step seat=1 source= target=seat0
[1480] pool_drain seat=1 source= amount=1 target=seat0
[1481] state seat=1 source= target=seat0
```

</details>

#### Violation 13

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 36, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 36, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1530 events
  Seat 0 [alive]: life=30 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=31 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1510] draw seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore amount=1 target=seat0
[1511] cast seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1512] stack_push seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1513] priority_pass seat=3 source= target=seat0
[1514] priority_pass seat=0 source= target=seat0
[1515] priority_pass seat=1 source= target=seat0
[1516] stack_resolve seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1517] zone_change seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore
[1518] resolve seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1519] phase_step seat=2 source= target=seat0
[1520] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1521] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1522] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1523] priority_pass seat=2 source= target=seat0
[1524] priority_pass seat=3 source= target=seat0
[1525] priority_pass seat=1 source= target=seat0
[1526] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1527] phase_step seat=2 source= target=seat0
[1528] pool_drain seat=2 source= amount=4 target=seat0
[1529] state seat=2 source= target=seat0
```

</details>

#### Violation 14

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 36, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 36, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1530 events
  Seat 0 [alive]: life=30 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=31 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=79 hand=2 graveyard=3 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Canoptek Tomb Sentinel (P/T 4/3, dmg=0)
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1510] draw seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore amount=1 target=seat0
[1511] cast seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1512] stack_push seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1513] priority_pass seat=3 source= target=seat0
[1514] priority_pass seat=0 source= target=seat0
[1515] priority_pass seat=1 source= target=seat0
[1516] stack_resolve seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1517] zone_change seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore
[1518] resolve seat=2 source=Uthros, Titanic Godcore // Uthros, Titanic Godcore target=seat0
[1519] phase_step seat=2 source= target=seat0
[1520] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1521] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1522] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1523] priority_pass seat=2 source= target=seat0
[1524] priority_pass seat=3 source= target=seat0
[1525] priority_pass seat=1 source= target=seat0
[1526] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1527] phase_step seat=2 source= target=seat0
[1528] pool_drain seat=2 source= amount=4 target=seat0
[1529] state seat=2 source= target=seat0
```

</details>

#### Violation 15

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 37, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 37, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 1633 events
  Seat 0 [alive]: life=24 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=31 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1613] phase_step seat=3 source= target=seat0
[1614] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1615] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1616] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1617] priority_pass seat=3 source= target=seat0
[1618] priority_pass seat=1 source= target=seat0
[1619] priority_pass seat=2 source= target=seat0
[1620] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1621] declare_attackers seat=3 source= target=seat0
[1622] blockers seat=0 source= target=seat0
[1623] damage seat=3 source=Canoptek Tomb Sentinel amount=4 target=seat0
[1624] damage seat=3 source=Viscerid Deepwalker amount=5 target=seat0
[1625] damage seat=0 source=Qumulox amount=6 target=seat3
[1626] destroy seat=3 source=Canoptek Tomb Sentinel
[1627] sba_704_5g seat=3 source=Canoptek Tomb Sentinel
[1628] zone_change seat=3 source=Canoptek Tomb Sentinel
[1629] sba_cycle_complete seat=-1 source=
[1630] phase_step seat=3 source= target=seat0
[1631] damage_wears_off seat=0 source=Qumulox amount=4 target=seat0
[1632] state seat=3 source= target=seat0
```

</details>

#### Violation 16

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 37, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 37, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 1633 events
  Seat 0 [alive]: life=24 library=81 hand=2 graveyard=6 exile=0 battlefield=11 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 6/5, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=31 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1613] phase_step seat=3 source= target=seat0
[1614] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1615] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1616] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1617] priority_pass seat=3 source= target=seat0
[1618] priority_pass seat=1 source= target=seat0
[1619] priority_pass seat=2 source= target=seat0
[1620] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1621] declare_attackers seat=3 source= target=seat0
[1622] blockers seat=0 source= target=seat0
[1623] damage seat=3 source=Canoptek Tomb Sentinel amount=4 target=seat0
[1624] damage seat=3 source=Viscerid Deepwalker amount=5 target=seat0
[1625] damage seat=0 source=Qumulox amount=6 target=seat3
[1626] destroy seat=3 source=Canoptek Tomb Sentinel
[1627] sba_704_5g seat=3 source=Canoptek Tomb Sentinel
[1628] zone_change seat=3 source=Canoptek Tomb Sentinel
[1629] sba_cycle_complete seat=-1 source=
[1630] phase_step seat=3 source= target=seat0
[1631] damage_wears_off seat=0 source=Qumulox amount=4 target=seat0
[1632] state seat=3 source= target=seat0
```

</details>

#### Violation 17

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 38, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 38, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 1691 events
  Seat 0 [alive]: life=24 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=21 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1671] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1672] priority_pass seat=1 source= target=seat0
[1673] priority_pass seat=2 source= target=seat0
[1674] priority_pass seat=3 source= target=seat0
[1675] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1676] phase_step seat=0 source= target=seat0
[1677] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1678] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1679] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1680] priority_pass seat=1 source= target=seat0
[1681] priority_pass seat=2 source= target=seat0
[1682] priority_pass seat=3 source= target=seat0
[1683] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1684] per_card_handler seat=0 source=Ezuri, Claw of Progress target=seat0
[1685] declare_attackers seat=0 source= target=seat0
[1686] blockers seat=2 source= target=seat0
[1687] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat2
[1688] damage seat=0 source=Qumulox amount=7 target=seat2
[1689] phase_step seat=0 source= target=seat0
[1690] state seat=0 source= target=seat0
```

</details>

#### Violation 18

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 38, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 38, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 1691 events
  Seat 0 [alive]: life=24 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=81 hand=4 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=21 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1671] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1672] priority_pass seat=1 source= target=seat0
[1673] priority_pass seat=2 source= target=seat0
[1674] priority_pass seat=3 source= target=seat0
[1675] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1676] phase_step seat=0 source= target=seat0
[1677] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1678] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1679] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1680] priority_pass seat=1 source= target=seat0
[1681] priority_pass seat=2 source= target=seat0
[1682] priority_pass seat=3 source= target=seat0
[1683] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1684] per_card_handler seat=0 source=Ezuri, Claw of Progress target=seat0
[1685] declare_attackers seat=0 source= target=seat0
[1686] blockers seat=2 source= target=seat0
[1687] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat2
[1688] damage seat=0 source=Qumulox amount=7 target=seat2
[1689] phase_step seat=0 source= target=seat0
[1690] state seat=0 source= target=seat0
```

</details>

#### Violation 19

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 39, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 39, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1706 events
  Seat 0 [alive]: life=24 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=21 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1686] blockers seat=2 source= target=seat0
[1687] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat2
[1688] damage seat=0 source=Qumulox amount=7 target=seat2
[1689] phase_step seat=0 source= target=seat0
[1690] state seat=0 source= target=seat0
[1691] turn_start seat=1 source= target=seat0
[1692] untap_done seat=1 source=Island target=seat0
[1693] add_mana seat=1 source=Island amount=1 target=seat0
[1694] draw seat=1 source=Regna, the Redeemer amount=1 target=seat0
[1695] phase_step seat=1 source= target=seat0
[1696] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1697] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1698] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1699] priority_pass seat=1 source= target=seat0
[1700] priority_pass seat=2 source= target=seat0
[1701] priority_pass seat=3 source= target=seat0
[1702] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1703] phase_step seat=1 source= target=seat0
[1704] pool_drain seat=1 source= amount=1 target=seat0
[1705] state seat=1 source= target=seat0
```

</details>

#### Violation 20

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 39, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 39, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1706 events
  Seat 0 [alive]: life=24 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=21 library=83 hand=1 graveyard=10 exile=0 battlefield=5 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1686] blockers seat=2 source= target=seat0
[1687] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat2
[1688] damage seat=0 source=Qumulox amount=7 target=seat2
[1689] phase_step seat=0 source= target=seat0
[1690] state seat=0 source= target=seat0
[1691] turn_start seat=1 source= target=seat0
[1692] untap_done seat=1 source=Island target=seat0
[1693] add_mana seat=1 source=Island amount=1 target=seat0
[1694] draw seat=1 source=Regna, the Redeemer amount=1 target=seat0
[1695] phase_step seat=1 source= target=seat0
[1696] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1697] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1698] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1699] priority_pass seat=1 source= target=seat0
[1700] priority_pass seat=2 source= target=seat0
[1701] priority_pass seat=3 source= target=seat0
[1702] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1703] phase_step seat=1 source= target=seat0
[1704] pool_drain seat=1 source= amount=1 target=seat0
[1705] state seat=1 source= target=seat0
```

</details>

#### Violation 21

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 40, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 40, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1748 events
  Seat 0 [alive]: life=24 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1728] stack_resolve seat=2 source=Copper Carapace target=seat0
[1729] enter_battlefield seat=2 source=Copper Carapace target=seat0
[1730] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1731] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1732] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1733] priority_pass seat=2 source= target=seat0
[1734] priority_pass seat=3 source= target=seat0
[1735] priority_pass seat=1 source= target=seat0
[1736] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1737] phase_step seat=2 source= target=seat0
[1738] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1739] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1740] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1741] priority_pass seat=2 source= target=seat0
[1742] priority_pass seat=3 source= target=seat0
[1743] priority_pass seat=1 source= target=seat0
[1744] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1745] phase_step seat=2 source= target=seat0
[1746] pool_drain seat=2 source= amount=3 target=seat0
[1747] state seat=2 source= target=seat0
```

</details>

#### Violation 22

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 40, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 40, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1748 events
  Seat 0 [alive]: life=24 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=11 library=78 hand=2 graveyard=4 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1728] stack_resolve seat=2 source=Copper Carapace target=seat0
[1729] enter_battlefield seat=2 source=Copper Carapace target=seat0
[1730] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1731] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1732] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1733] priority_pass seat=2 source= target=seat0
[1734] priority_pass seat=3 source= target=seat0
[1735] priority_pass seat=1 source= target=seat0
[1736] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1737] phase_step seat=2 source= target=seat0
[1738] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1739] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1740] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1741] priority_pass seat=2 source= target=seat0
[1742] priority_pass seat=3 source= target=seat0
[1743] priority_pass seat=1 source= target=seat0
[1744] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1745] phase_step seat=2 source= target=seat0
[1746] pool_drain seat=2 source= amount=3 target=seat0
[1747] state seat=2 source= target=seat0
```

</details>

#### Violation 23

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 41, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 41, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 1819 events
  Seat 0 [alive]: life=19 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=11 library=77 hand=2 graveyard=4 exile=0 battlefield=12 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]
    - Thought Harvester (P/T 2/4, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1799] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1800] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1801] priority_pass seat=3 source= target=seat0
[1802] priority_pass seat=1 source= target=seat0
[1803] priority_pass seat=2 source= target=seat0
[1804] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1805] phase_step seat=3 source= target=seat0
[1806] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1807] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1808] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1809] priority_pass seat=3 source= target=seat0
[1810] priority_pass seat=1 source= target=seat0
[1811] priority_pass seat=2 source= target=seat0
[1812] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1813] declare_attackers seat=3 source= target=seat0
[1814] blockers seat=0 source= target=seat0
[1815] damage seat=3 source=Viscerid Deepwalker amount=2 target=seat0
[1816] damage seat=3 source=Bonded Construct amount=2 target=seat0
[1817] phase_step seat=3 source= target=seat0
[1818] state seat=3 source= target=seat0
```

</details>

#### Violation 24

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 41, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 41, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 1819 events
  Seat 0 [alive]: life=19 library=80 hand=2 graveyard=6 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 7/6, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=11 library=77 hand=2 graveyard=4 exile=0 battlefield=12 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]
    - Thought Harvester (P/T 2/4, dmg=0)

```

</details>

<details>
<summary>Recent Events</summary>

```
[1799] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1800] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1801] priority_pass seat=3 source= target=seat0
[1802] priority_pass seat=1 source= target=seat0
[1803] priority_pass seat=2 source= target=seat0
[1804] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1805] phase_step seat=3 source= target=seat0
[1806] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1807] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1808] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1809] priority_pass seat=3 source= target=seat0
[1810] priority_pass seat=1 source= target=seat0
[1811] priority_pass seat=2 source= target=seat0
[1812] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1813] declare_attackers seat=3 source= target=seat0
[1814] blockers seat=0 source= target=seat0
[1815] damage seat=3 source=Viscerid Deepwalker amount=2 target=seat0
[1816] damage seat=3 source=Bonded Construct amount=2 target=seat0
[1817] phase_step seat=3 source= target=seat0
[1818] state seat=3 source= target=seat0
```

</details>

#### Violation 25

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 42, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 42, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 1879 events
  Seat 0 [alive]: life=19 library=79 hand=2 graveyard=7 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 8/7, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=8 library=77 hand=2 graveyard=5 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[1859] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1860] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1861] priority_pass seat=1 source= target=seat0
[1862] priority_pass seat=2 source= target=seat0
[1863] priority_pass seat=3 source= target=seat0
[1864] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1865] per_card_handler seat=0 source=Ezuri, Claw of Progress target=seat0
[1866] declare_attackers seat=0 source= target=seat0
[1867] blockers seat=3 source= target=seat0
[1868] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat3
[1869] damage seat=0 source=Qumulox amount=4 target=seat3
[1870] damage seat=3 source=Thought Harvester amount=2 target=seat0
[1871] destroy seat=3 source=Thought Harvester
[1872] sba_704_5g seat=3 source=Thought Harvester
[1873] zone_change seat=3 source=Thought Harvester
[1874] sba_cycle_complete seat=-1 source=
[1875] phase_step seat=0 source= target=seat0
[1876] pool_drain seat=0 source= amount=4 target=seat0
[1877] damage_wears_off seat=0 source=Qumulox amount=2 target=seat0
[1878] state seat=0 source= target=seat0
```

</details>

#### Violation 26

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 42, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 42, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 1879 events
  Seat 0 [alive]: life=19 library=79 hand=2 graveyard=7 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 8/7, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=80 hand=5 graveyard=12 exile=0 battlefield=1 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=8 library=77 hand=2 graveyard=5 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[1859] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1860] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1861] priority_pass seat=1 source= target=seat0
[1862] priority_pass seat=2 source= target=seat0
[1863] priority_pass seat=3 source= target=seat0
[1864] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1865] per_card_handler seat=0 source=Ezuri, Claw of Progress target=seat0
[1866] declare_attackers seat=0 source= target=seat0
[1867] blockers seat=3 source= target=seat0
[1868] damage seat=0 source=Ezuri, Claw of Progress amount=3 target=seat3
[1869] damage seat=0 source=Qumulox amount=4 target=seat3
[1870] damage seat=3 source=Thought Harvester amount=2 target=seat0
[1871] destroy seat=3 source=Thought Harvester
[1872] sba_704_5g seat=3 source=Thought Harvester
[1873] zone_change seat=3 source=Thought Harvester
[1874] sba_cycle_complete seat=-1 source=
[1875] phase_step seat=0 source= target=seat0
[1876] pool_drain seat=0 source= amount=4 target=seat0
[1877] damage_wears_off seat=0 source=Qumulox amount=2 target=seat0
[1878] state seat=0 source= target=seat0
```

</details>

#### Violation 27

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 43, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 43, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1903 events
  Seat 0 [alive]: life=19 library=79 hand=2 graveyard=7 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 8/7, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=79 hand=5 graveyard=12 exile=0 battlefield=2 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=8 library=77 hand=2 graveyard=5 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[1883] play_land seat=1 source=Swamp target=seat0
[1884] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1885] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1886] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1887] priority_pass seat=1 source= target=seat0
[1888] priority_pass seat=2 source= target=seat0
[1889] priority_pass seat=3 source= target=seat0
[1890] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1891] add_mana seat=1 source=Swamp amount=1 target=seat0
[1892] phase_step seat=1 source= target=seat0
[1893] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1894] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1895] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1896] priority_pass seat=1 source= target=seat0
[1897] priority_pass seat=2 source= target=seat0
[1898] priority_pass seat=3 source= target=seat0
[1899] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1900] phase_step seat=1 source= target=seat0
[1901] pool_drain seat=1 source= amount=2 target=seat0
[1902] state seat=1 source= target=seat0
```

</details>

#### Violation 28

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 43, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 43, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 1903 events
  Seat 0 [alive]: life=19 library=79 hand=2 graveyard=7 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 8/7, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=79 hand=5 graveyard=12 exile=0 battlefield=2 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=22 library=82 hand=1 graveyard=10 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=8 library=77 hand=2 graveyard=5 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[1883] play_land seat=1 source=Swamp target=seat0
[1884] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1885] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1886] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1887] priority_pass seat=1 source= target=seat0
[1888] priority_pass seat=2 source= target=seat0
[1889] priority_pass seat=3 source= target=seat0
[1890] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1891] add_mana seat=1 source=Swamp amount=1 target=seat0
[1892] phase_step seat=1 source= target=seat0
[1893] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1894] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1895] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1896] priority_pass seat=1 source= target=seat0
[1897] priority_pass seat=2 source= target=seat0
[1898] priority_pass seat=3 source= target=seat0
[1899] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1900] phase_step seat=1 source= target=seat0
[1901] pool_drain seat=1 source= amount=2 target=seat0
[1902] state seat=1 source= target=seat0
```

</details>

#### Violation 29

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 44, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 44, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1938 events
  Seat 0 [alive]: life=19 library=79 hand=2 graveyard=7 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 8/7, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=79 hand=5 graveyard=12 exile=0 battlefield=2 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=23 library=81 hand=1 graveyard=11 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=8 library=77 hand=2 graveyard=5 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[1918] draw seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator amount=1 target=seat0
[1919] cast seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1920] stack_push seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1921] priority_pass seat=3 source= target=seat0
[1922] priority_pass seat=0 source= target=seat0
[1923] priority_pass seat=1 source= target=seat0
[1924] stack_resolve seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1925] zone_change seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator
[1926] resolve seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1927] phase_step seat=2 source= target=seat0
[1928] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1929] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1930] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1931] priority_pass seat=2 source= target=seat0
[1932] priority_pass seat=3 source= target=seat0
[1933] priority_pass seat=1 source= target=seat0
[1934] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1935] phase_step seat=2 source= target=seat0
[1936] pool_drain seat=2 source= amount=4 target=seat0
[1937] state seat=2 source= target=seat0
```

</details>

#### Violation 30

- **Game**: 170 (seed 1700043, perm 0)
- **Invariant**: CardIdentity
- **Turn**: 44, Phase=ending Step=cleanup
- **Commanders**: Ezuri, Claw of Progress, The Master of Keys, Syr Ginger, the Meal Ender, Katara, Water Tribe's Hope
- **Message**: CardIdentity: card "Adric, Mathematical Genius" (ptr 0xc00c88b600) appears in both seat 1 hand and seat 1 graveyard

<details>
<summary>Game State</summary>

```
Turn 44, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 1938 events
  Seat 0 [alive]: life=19 library=79 hand=2 graveyard=7 exile=0 battlefield=12 cmdzone=0 mana=0
    - Forest (P/T 0/0, dmg=0) [T]
    - Mana Prism (P/T 0/0, dmg=0)
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Bounty Board (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Ezuri, Claw of Progress (P/T 3/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Forest (P/T 0/0, dmg=0) [T]
    - Qumulox (P/T 8/7, dmg=0) [T]
    - Lurking Predators (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=28 library=79 hand=5 graveyard=12 exile=0 battlefield=2 cmdzone=1 mana=0
    - Island (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=23 library=81 hand=1 graveyard=11 exile=0 battlefield=6 cmdzone=1 mana=0
    - War Room (P/T 0/0, dmg=0) [T]
    - Inventors' Fair (P/T 0/0, dmg=0) [T]
    - Gallifrey Council Chamber (P/T 0/0, dmg=0) [T]
    - Grafted Wargear (P/T 0/0, dmg=0)
    - Vesuva (P/T 0/0, dmg=0) [T]
    - Copper Carapace (P/T 0/0, dmg=0)
  Seat 3 [alive]: life=8 library=77 hand=2 graveyard=5 exile=0 battlefield=11 cmdzone=1 mana=0
    - Hidden Grotto (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Icatian Javelineers (P/T 1/1, dmg=0) [T]
    - Rikku, Resourceful Guardian (P/T 2/3, dmg=0) [T]
    - Serendib Sorcerer (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Squeeze (P/T 0/0, dmg=0)
    - Kor Halberd (P/T 0/0, dmg=0)
    - Viscerid Deepwalker (P/T 2/3, dmg=0) [T]
    - Bonded Construct (P/T 2/1, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[1918] draw seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator amount=1 target=seat0
[1919] cast seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1920] stack_push seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1921] priority_pass seat=3 source= target=seat0
[1922] priority_pass seat=0 source= target=seat0
[1923] priority_pass seat=1 source= target=seat0
[1924] stack_resolve seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1925] zone_change seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator
[1926] resolve seat=2 source=Jhoira, Ageless Innovator // Jhoira, Ageless Innovator target=seat0
[1927] phase_step seat=2 source= target=seat0
[1928] trigger_evaluated seat=0 source=Ezuri, Claw of Progress
[1929] stack_push seat=0 source=Ezuri, Claw of Progress target=seat0
[1930] triggered_ability seat=0 source=Ezuri, Claw of Progress target=seat0
[1931] priority_pass seat=2 source= target=seat0
[1932] priority_pass seat=3 source= target=seat0
[1933] priority_pass seat=1 source= target=seat0
[1934] stack_resolve seat=0 source=Ezuri, Claw of Progress target=seat0
[1935] phase_step seat=2 source= target=seat0
[1936] pool_drain seat=2 source= amount=4 target=seat0
[1937] state seat=2 source= target=seat0
```

</details>

*... and 450 more violations not shown.*

## Invariant Violations (Nightmare Boards)

| Invariant | Count |
|-----------|-------|
| CardIdentity | 2 |

## Top Cards Correlated with Violations

Cards that appeared disproportionately in violation games vs clean games.
Only cards appearing in 3+ total games are shown.

| Rank | Card | Violation Games | Clean Games | Correlation |
|------|------|-----------------|-------------|-------------|
| 1 | Coveted Prize | 2 | 1 | 0.67 |
| 2 | Zurgo Bellstriker | 2 | 2 | 0.50 |
| 3 | Luminarch Aspirant | 2 | 2 | 0.50 |
| 4 | Loafing Giant | 2 | 2 | 0.50 |
| 5 | Piece It Together | 2 | 3 | 0.40 |
| 6 | Tomik, Distinguished Advokist | 2 | 3 | 0.40 |
| 7 | Axebane Stag | 2 | 3 | 0.40 |
| 8 | Consuming Ferocity | 2 | 3 | 0.40 |
| 9 | Sinuous Benthisaur | 2 | 3 | 0.40 |
| 10 | Jeweled Spirit | 2 | 3 | 0.40 |

## Verdict: ISSUES FOUND

**482 total issues** across 1000 chaos games and 10000 nightmare boards.
- 0 crashes in chaos games
- 480 invariant violations in chaos games
- 0 crashes in nightmare boards
- 2 invariant violations in nightmare boards

Review the details above to identify which cards and interactions are problematic.
