# Chaos Gauntlet Report

Generated: 2026-05-17T09:22:30-07:00

## Configuration

| Parameter | Value |
|-----------|-------|
| Oracle Corpus | 36656 cards |
| Legendary Creatures | 3433 |
| Total Games | 2000 |
| Seed | 777 |
| Permutations | 1 |
| Seats | 4 |
| Max Turns | 60 |
| Nightmare Boards | 10000 |

## Summary

### Chaos Games

| Metric | Count |
|--------|-------|
| Duration | 1m19.55s |
| Throughput | 25 games/sec |
| Crashes | 0 (in 0 games) |
| Invariant Violations | 620 (in 25 games) |
| Clean Games | 1975 |

### Nightmare Boards

| Metric | Count |
|--------|-------|
| Duration | 2.102s |
| Throughput | 4756 boards/sec |
| Crashes | 0 |
| Invariant Violations | 2 |
| Clean Boards | 9999 |

## Invariant Violations (Chaos Games)

### By Invariant

| Invariant | Count |
|-----------|-------|
| AttachmentConsistency | 16 |
| ZoneConservation | 172 |
| CardIdentity | 424 |
| ZoneCastGrantExpiry | 2 |
| CombatLegality | 4 |
| TriggerCompleteness | 2 |

### Violation Details (first 30)

#### Violation 1

- **Game**: 153 (seed 1530778, perm 0)
- **Invariant**: AttachmentConsistency
- **Turn**: 46, Phase=combat Step=end_of_combat
- **Commanders**: Borborygmos and Fblthp, Irma, Part-Time Mutant, Mai, Jaded Edge, Quicksilver, Brash Blur
- **Message**: AttachmentConsistency: "Hedron Blade" (seat 1) is attached to "creature token construct Token" which is not on any battlefield

<details>
<summary>Game State</summary>

```
Turn 46, Phase=combat Step=end_of_combat Active=seat1
Stack: 0 items, EventLog: 3607 events
  Seat 0 [LOST]: life=-28 library=76 hand=6 graveyard=7 exile=0 battlefield=0 cmdzone=1 mana=0
  Seat 1 [WON]: life=10 library=68 hand=0 graveyard=17 exile=0 battlefield=13 cmdzone=0 mana=3
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Hedron Blade (P/T 0/0, dmg=0)
    - Seeker of Insight (P/T 1/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Treasure Map // Treasure Cove (P/T 0/0, dmg=0) [T]
    - Stone Idol Generator (P/T 0/0, dmg=0) [T]
    - Sword of the Squeak (P/T 0/0, dmg=0)
    - Gate to Seatower (P/T 0/0, dmg=0) [T]
    - Irma, Part-Time Mutant (P/T 1/1, dmg=0) [T]
    - Skywise Teachings (P/T 0/0, dmg=0)
  Seat 2 [LOST]: life=0 library=82 hand=6 graveyard=9 exile=0 battlefield=0 cmdzone=1 mana=0
  Seat 3 [LOST]: life=-33 library=80 hand=0 graveyard=7 exile=0 battlefield=0 cmdzone=0 mana=0

```

</details>

<details>
<summary>Recent Events</summary>

```
[3587] priority_pass seat=0 source= target=seat0
[3588] stack_resolve seat=1 source=Stone Idol Generator target=seat0
[3589] gain_energy seat=1 source=Stone Idol Generator amount=1 target=seat0
[3590] blockers seat=0 source= target=seat0
[3591] damage seat=1 source=creature token construct Token amount=2 target=seat0
[3592] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3593] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3594] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3595] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3596] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3597] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3598] damage seat=1 source=Irma, Part-Time Mutant amount=1 target=seat0
[3599] damage seat=0 source=Khenra Spellspear // Gitaxian Spellstalker amount=2 target=seat1
[3600] sba_704_5a seat=0 source= amount=-28
[3601] destroy seat=0 source=Khenra Spellspear // Gitaxian Spellstalker
[3602] sba_704_5g seat=0 source=Khenra Spellspear // Gitaxian Spellstalker
[3603] zone_change seat=0 source=Khenra Spellspear // Gitaxian Spellstalker
[3604] sba_cycle_complete seat=-1 source=
[3605] seat_eliminated seat=0 source= amount=18
[3606] game_end seat=1 source=
```

</details>

#### Violation 2

- **Game**: 153 (seed 1530778, perm 0)
- **Invariant**: AttachmentConsistency
- **Turn**: 46, Phase=combat Step=end_of_combat
- **Commanders**: Borborygmos and Fblthp, Irma, Part-Time Mutant, Mai, Jaded Edge, Quicksilver, Brash Blur
- **Message**: AttachmentConsistency: "Hedron Blade" (seat 1) is attached to "creature token construct Token" which is not on any battlefield

<details>
<summary>Game State</summary>

```
Turn 46, Phase=combat Step=end_of_combat Active=seat1
Stack: 0 items, EventLog: 3607 events
  Seat 0 [LOST]: life=-28 library=76 hand=6 graveyard=7 exile=0 battlefield=0 cmdzone=1 mana=0
  Seat 1 [WON]: life=10 library=68 hand=0 graveyard=17 exile=0 battlefield=13 cmdzone=0 mana=3
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Hedron Blade (P/T 0/0, dmg=0)
    - Seeker of Insight (P/T 1/3, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Treasure Map // Treasure Cove (P/T 0/0, dmg=0) [T]
    - Stone Idol Generator (P/T 0/0, dmg=0) [T]
    - Sword of the Squeak (P/T 0/0, dmg=0)
    - Gate to Seatower (P/T 0/0, dmg=0) [T]
    - Irma, Part-Time Mutant (P/T 1/1, dmg=0) [T]
    - Skywise Teachings (P/T 0/0, dmg=0)
  Seat 2 [LOST]: life=0 library=82 hand=6 graveyard=9 exile=0 battlefield=0 cmdzone=1 mana=0
  Seat 3 [LOST]: life=-33 library=80 hand=0 graveyard=7 exile=0 battlefield=0 cmdzone=0 mana=0

```

</details>

<details>
<summary>Recent Events</summary>

```
[3587] priority_pass seat=0 source= target=seat0
[3588] stack_resolve seat=1 source=Stone Idol Generator target=seat0
[3589] gain_energy seat=1 source=Stone Idol Generator amount=1 target=seat0
[3590] blockers seat=0 source= target=seat0
[3591] damage seat=1 source=creature token construct Token amount=2 target=seat0
[3592] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3593] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3594] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3595] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3596] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3597] damage seat=1 source=creature token construct Token amount=6 target=seat0
[3598] damage seat=1 source=Irma, Part-Time Mutant amount=1 target=seat0
[3599] damage seat=0 source=Khenra Spellspear // Gitaxian Spellstalker amount=2 target=seat1
[3600] sba_704_5a seat=0 source= amount=-28
[3601] destroy seat=0 source=Khenra Spellspear // Gitaxian Spellstalker
[3602] sba_704_5g seat=0 source=Khenra Spellspear // Gitaxian Spellstalker
[3603] zone_change seat=0 source=Khenra Spellspear // Gitaxian Spellstalker
[3604] sba_cycle_complete seat=-1 source=
[3605] seat_eliminated seat=0 source= amount=18
[3606] game_end seat=1 source=
```

</details>

#### Violation 3

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 42, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 42, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2048 events
  Seat 0 [alive]: life=22 library=74 hand=0 graveyard=10 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Harvest Hand // Scrounged Scythe (P/T 2/2, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=9 library=82 hand=0 graveyard=3 exile=0 battlefield=15 cmdzone=0 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Petrified Field (P/T 0/0, dmg=0) [T]
    - Catalyst Stone (P/T 0/0, dmg=0)
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=28 library=80 hand=7 graveyard=5 exile=0 battlefield=5 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2028] play_land seat=1 source=Plains target=seat0
[2029] add_mana seat=1 source=Plains amount=1 target=seat0
[2030] pay_mana seat=1 source=Zephyr Boots amount=1 target=seat0
[2031] cast seat=1 source=Zephyr Boots amount=1 target=seat0
[2032] stack_push seat=1 source=Zephyr Boots target=seat0
[2033] priority_pass seat=2 source= target=seat0
[2034] priority_pass seat=3 source= target=seat0
[2035] priority_pass seat=0 source= target=seat0
[2036] stack_resolve seat=1 source=Zephyr Boots target=seat0
[2037] enter_battlefield seat=1 source=Zephyr Boots target=seat0
[2038] citys_blessing seat=1 source= amount=14 target=seat0
[2039] equip seat=1 source=Zephyr Boots amount=1 target=seat0
[2040] phase_step seat=1 source= target=seat0
[2041] declare_attackers seat=1 source= target=seat0
[2042] blockers seat=0 source= target=seat0
[2043] damage seat=1 source=Knight of Dawn amount=2 target=seat0
[2044] phase_step seat=1 source= target=seat0
[2045] zone_change seat=1 source=Elenda's Hierophant
[2046] discard seat=1 source=Elenda's Hierophant target=seat0
[2047] state seat=1 source= target=seat0
```

</details>

#### Violation 4

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 42, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 42, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2048 events
  Seat 0 [alive]: life=22 library=74 hand=0 graveyard=10 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Harvest Hand // Scrounged Scythe (P/T 2/2, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=9 library=82 hand=0 graveyard=3 exile=0 battlefield=15 cmdzone=0 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Petrified Field (P/T 0/0, dmg=0) [T]
    - Catalyst Stone (P/T 0/0, dmg=0)
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=28 library=80 hand=7 graveyard=5 exile=0 battlefield=5 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2028] play_land seat=1 source=Plains target=seat0
[2029] add_mana seat=1 source=Plains amount=1 target=seat0
[2030] pay_mana seat=1 source=Zephyr Boots amount=1 target=seat0
[2031] cast seat=1 source=Zephyr Boots amount=1 target=seat0
[2032] stack_push seat=1 source=Zephyr Boots target=seat0
[2033] priority_pass seat=2 source= target=seat0
[2034] priority_pass seat=3 source= target=seat0
[2035] priority_pass seat=0 source= target=seat0
[2036] stack_resolve seat=1 source=Zephyr Boots target=seat0
[2037] enter_battlefield seat=1 source=Zephyr Boots target=seat0
[2038] citys_blessing seat=1 source= amount=14 target=seat0
[2039] equip seat=1 source=Zephyr Boots amount=1 target=seat0
[2040] phase_step seat=1 source= target=seat0
[2041] declare_attackers seat=1 source= target=seat0
[2042] blockers seat=0 source= target=seat0
[2043] damage seat=1 source=Knight of Dawn amount=2 target=seat0
[2044] phase_step seat=1 source= target=seat0
[2045] zone_change seat=1 source=Elenda's Hierophant
[2046] discard seat=1 source=Elenda's Hierophant target=seat0
[2047] state seat=1 source= target=seat0
```

</details>

#### Violation 5

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 43, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 43, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 2142 events
  Seat 0 [alive]: life=16 library=74 hand=0 graveyard=10 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Harvest Hand // Scrounged Scythe (P/T 2/2, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=9 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=80 hand=7 graveyard=5 exile=0 battlefield=5 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2122] declare_attackers seat=2 source= target=seat0
[2123] blockers seat=0 source= target=seat0
[2124] damage seat=2 source=Kain, Traitorous Dragoon amount=2 target=seat0
[2125] trigger_fires seat=2 source=Kain, Traitorous Dragoon amount=2 target=seat0
[2126] stack_push seat=2 source=Kain, Traitorous Dragoon target=seat0
[2127] priority_pass seat=3 source= target=seat0
[2128] priority_pass seat=0 source= target=seat0
[2129] priority_pass seat=1 source= target=seat0
[2130] stack_resolve seat=2 source=Kain, Traitorous Dragoon target=seat0
[2131] parsed_effect_residual seat=2 source=Kain, Traitorous Dragoon target=seat0
[2132] damage seat=2 source=Giant Fly amount=2 target=seat0
[2133] damage seat=2 source=Infernal Pet amount=2 target=seat0
[2134] phase_step seat=2 source= target=seat0
[2135] stack_push seat=2 source=Desolation target=seat0
[2136] priority_pass seat=3 source= target=seat0
[2137] priority_pass seat=0 source= target=seat0
[2138] priority_pass seat=1 source= target=seat0
[2139] stack_resolve seat=2 source=Desolation target=seat0
[2140] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[2141] state seat=2 source= target=seat0
```

</details>

#### Violation 6

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 43, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 43, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 2142 events
  Seat 0 [alive]: life=16 library=74 hand=0 graveyard=10 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Harvest Hand // Scrounged Scythe (P/T 2/2, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=9 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=80 hand=7 graveyard=5 exile=0 battlefield=5 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2122] declare_attackers seat=2 source= target=seat0
[2123] blockers seat=0 source= target=seat0
[2124] damage seat=2 source=Kain, Traitorous Dragoon amount=2 target=seat0
[2125] trigger_fires seat=2 source=Kain, Traitorous Dragoon amount=2 target=seat0
[2126] stack_push seat=2 source=Kain, Traitorous Dragoon target=seat0
[2127] priority_pass seat=3 source= target=seat0
[2128] priority_pass seat=0 source= target=seat0
[2129] priority_pass seat=1 source= target=seat0
[2130] stack_resolve seat=2 source=Kain, Traitorous Dragoon target=seat0
[2131] parsed_effect_residual seat=2 source=Kain, Traitorous Dragoon target=seat0
[2132] damage seat=2 source=Giant Fly amount=2 target=seat0
[2133] damage seat=2 source=Infernal Pet amount=2 target=seat0
[2134] phase_step seat=2 source= target=seat0
[2135] stack_push seat=2 source=Desolation target=seat0
[2136] priority_pass seat=3 source= target=seat0
[2137] priority_pass seat=0 source= target=seat0
[2138] priority_pass seat=1 source= target=seat0
[2139] stack_resolve seat=2 source=Desolation target=seat0
[2140] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[2141] state seat=2 source= target=seat0
```

</details>

#### Violation 7

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 44, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 44, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 2207 events
  Seat 0 [alive]: life=16 library=74 hand=0 graveyard=10 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Harvest Hand // Scrounged Scythe (P/T 2/2, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=9 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2187] priority_pass seat=2 source= target=seat0
[2188] stack_resolve seat=3 source=Mutavault target=seat0
[2189] modification_effect seat=3 source=Mutavault target=seat0
[2190] parser_gap seat=3 source=Mutavault target=seat0
[2191] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2192] play_land seat=3 source=Swamp target=seat0
[2193] add_mana seat=3 source=Swamp amount=1 target=seat0
[2194] pay_mana seat=3 source=Mutavault amount=1 target=seat0
[2195] activate_ability seat=3 source=Mutavault target=seat0
[2196] stack_push seat=3 source=Mutavault target=seat0
[2197] priority_pass seat=0 source= target=seat0
[2198] priority_pass seat=1 source= target=seat0
[2199] priority_pass seat=2 source= target=seat0
[2200] stack_resolve seat=3 source=Mutavault target=seat0
[2201] modification_effect seat=3 source=Mutavault target=seat0
[2202] parser_gap seat=3 source=Mutavault target=seat0
[2203] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2204] phase_step seat=3 source= target=seat0
[2205] phase_step seat=3 source= target=seat0
[2206] state seat=3 source= target=seat0
```

</details>

#### Violation 8

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 44, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 44, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 2207 events
  Seat 0 [alive]: life=16 library=74 hand=0 graveyard=10 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Harvest Hand // Scrounged Scythe (P/T 2/2, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=9 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2187] priority_pass seat=2 source= target=seat0
[2188] stack_resolve seat=3 source=Mutavault target=seat0
[2189] modification_effect seat=3 source=Mutavault target=seat0
[2190] parser_gap seat=3 source=Mutavault target=seat0
[2191] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2192] play_land seat=3 source=Swamp target=seat0
[2193] add_mana seat=3 source=Swamp amount=1 target=seat0
[2194] pay_mana seat=3 source=Mutavault amount=1 target=seat0
[2195] activate_ability seat=3 source=Mutavault target=seat0
[2196] stack_push seat=3 source=Mutavault target=seat0
[2197] priority_pass seat=0 source= target=seat0
[2198] priority_pass seat=1 source= target=seat0
[2199] priority_pass seat=2 source= target=seat0
[2200] stack_resolve seat=3 source=Mutavault target=seat0
[2201] modification_effect seat=3 source=Mutavault target=seat0
[2202] parser_gap seat=3 source=Mutavault target=seat0
[2203] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2204] phase_step seat=3 source= target=seat0
[2205] phase_step seat=3 source= target=seat0
[2206] state seat=3 source= target=seat0
```

</details>

#### Violation 9

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 45, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 45, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 2259 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2239] declare_attackers seat=0 source= target=seat0
[2240] blockers seat=2 source= target=seat0
[2241] damage seat=0 source=Harvest Hand // Scrounged Scythe amount=2 target=seat2
[2242] damage seat=0 source=Spider-Bot amount=2 target=seat2
[2243] damage seat=0 source=Sokka, Bold Boomeranger amount=1 target=seat2
[2244] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat0
[2245] destroy seat=0 source=Harvest Hand // Scrounged Scythe
[2246] sba_704_5g seat=0 source=Harvest Hand // Scrounged Scythe
[2247] zone_change seat=0 source=Harvest Hand // Scrounged Scythe
[2248] stack_push seat=0 source=Harvest Hand // Scrounged Scythe target=seat0
[2249] priority_pass seat=1 source= target=seat0
[2250] priority_pass seat=2 source= target=seat0
[2251] priority_pass seat=3 source= target=seat0
[2252] stack_resolve seat=0 source=Harvest Hand // Scrounged Scythe target=seat0
[2253] parsed_effect_residual seat=0 source=Harvest Hand // Scrounged Scythe target=seat0
[2254] sba_cycle_complete seat=-1 source=
[2255] phase_step seat=0 source= target=seat0
[2256] pool_drain seat=0 source= amount=5 target=seat0
[2257] damage_wears_off seat=2 source=Blood-Chin Fanatic amount=2 target=seat0
[2258] state seat=0 source= target=seat0
```

</details>

#### Violation 10

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 45, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 11 extra real cards appeared (expected 394, found 405) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 45, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 2259 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=15 library=57 hand=7 graveyard=18 exile=1 battlefield=14 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Knight of Dawn (P/T 2/2, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2239] declare_attackers seat=0 source= target=seat0
[2240] blockers seat=2 source= target=seat0
[2241] damage seat=0 source=Harvest Hand // Scrounged Scythe amount=2 target=seat2
[2242] damage seat=0 source=Spider-Bot amount=2 target=seat2
[2243] damage seat=0 source=Sokka, Bold Boomeranger amount=1 target=seat2
[2244] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat0
[2245] destroy seat=0 source=Harvest Hand // Scrounged Scythe
[2246] sba_704_5g seat=0 source=Harvest Hand // Scrounged Scythe
[2247] zone_change seat=0 source=Harvest Hand // Scrounged Scythe
[2248] stack_push seat=0 source=Harvest Hand // Scrounged Scythe target=seat0
[2249] priority_pass seat=1 source= target=seat0
[2250] priority_pass seat=2 source= target=seat0
[2251] priority_pass seat=3 source= target=seat0
[2252] stack_resolve seat=0 source=Harvest Hand // Scrounged Scythe target=seat0
[2253] parsed_effect_residual seat=0 source=Harvest Hand // Scrounged Scythe target=seat0
[2254] sba_cycle_complete seat=-1 source=
[2255] phase_step seat=0 source= target=seat0
[2256] pool_drain seat=0 source= amount=5 target=seat0
[2257] damage_wears_off seat=2 source=Blood-Chin Fanatic amount=2 target=seat0
[2258] state seat=0 source= target=seat0
```

</details>

#### Violation 11

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 46, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 46, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2387 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2367] cast seat=1 source=Smile at Death // Smile at Death target=seat0
[2368] stack_push seat=1 source=Smile at Death // Smile at Death target=seat0
[2369] priority_pass seat=2 source= target=seat0
[2370] priority_pass seat=3 source= target=seat0
[2371] priority_pass seat=0 source= target=seat0
[2372] stack_resolve seat=1 source=Smile at Death // Smile at Death target=seat0
[2373] zone_change seat=1 source=Smile at Death // Smile at Death
[2374] resolve seat=1 source=Smile at Death // Smile at Death target=seat0
[2375] phase_step seat=1 source= target=seat0
[2376] declare_attackers seat=1 source= target=seat0
[2377] blockers seat=2 source= target=seat0
[2378] damage seat=1 source=Knight of Dawn amount=2 target=seat2
[2379] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat1
[2380] destroy seat=1 source=Knight of Dawn
[2381] sba_704_5g seat=1 source=Knight of Dawn
[2382] zone_change seat=1 source=Knight of Dawn
[2383] sba_cycle_complete seat=-1 source=
[2384] phase_step seat=1 source= target=seat0
[2385] damage_wears_off seat=2 source=Blood-Chin Fanatic amount=2 target=seat0
[2386] state seat=1 source= target=seat0
```

</details>

#### Violation 12

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 46, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 46, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2387 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=81 hand=0 graveyard=6 exile=0 battlefield=13 cmdzone=0 mana=0
    - Darksteel Citadel (P/T 0/0, dmg=0) [T]
    - Kain, Traitorous Dragoon (P/T 2/4, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0)
  Seat 3 [alive]: life=28 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2367] cast seat=1 source=Smile at Death // Smile at Death target=seat0
[2368] stack_push seat=1 source=Smile at Death // Smile at Death target=seat0
[2369] priority_pass seat=2 source= target=seat0
[2370] priority_pass seat=3 source= target=seat0
[2371] priority_pass seat=0 source= target=seat0
[2372] stack_resolve seat=1 source=Smile at Death // Smile at Death target=seat0
[2373] zone_change seat=1 source=Smile at Death // Smile at Death
[2374] resolve seat=1 source=Smile at Death // Smile at Death target=seat0
[2375] phase_step seat=1 source= target=seat0
[2376] declare_attackers seat=1 source= target=seat0
[2377] blockers seat=2 source= target=seat0
[2378] damage seat=1 source=Knight of Dawn amount=2 target=seat2
[2379] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat1
[2380] destroy seat=1 source=Knight of Dawn
[2381] sba_704_5g seat=1 source=Knight of Dawn
[2382] zone_change seat=1 source=Knight of Dawn
[2383] sba_cycle_complete seat=-1 source=
[2384] phase_step seat=1 source= target=seat0
[2385] damage_wears_off seat=2 source=Blood-Chin Fanatic amount=2 target=seat0
[2386] state seat=1 source= target=seat0
```

</details>

#### Violation 13

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 47, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 47, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 2463 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2443] priority_pass seat=3 source= target=seat0
[2444] priority_pass seat=0 source= target=seat0
[2445] priority_pass seat=1 source= target=seat0
[2446] stack_resolve seat=2 source=Blood-Chin Fanatic target=seat0
[2447] parsed_effect_residual seat=2 source=Blood-Chin Fanatic target=seat0
[2448] activated_ability_resolved seat=2 source=Blood-Chin Fanatic target=seat0
[2449] phase_step seat=2 source= target=seat0
[2450] declare_attackers seat=2 source= target=seat0
[2451] blockers seat=3 source= target=seat0
[2452] damage seat=2 source=Giant Fly amount=2 target=seat3
[2453] damage seat=2 source=Infernal Pet amount=2 target=seat3
[2454] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat3
[2455] phase_step seat=2 source= target=seat0
[2456] stack_push seat=2 source=Desolation target=seat0
[2457] priority_pass seat=3 source= target=seat0
[2458] priority_pass seat=0 source= target=seat0
[2459] priority_pass seat=1 source= target=seat0
[2460] stack_resolve seat=2 source=Desolation target=seat0
[2461] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[2462] state seat=2 source= target=seat0
```

</details>

#### Violation 14

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 47, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 47, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 2463 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=79 hand=7 graveyard=5 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2443] priority_pass seat=3 source= target=seat0
[2444] priority_pass seat=0 source= target=seat0
[2445] priority_pass seat=1 source= target=seat0
[2446] stack_resolve seat=2 source=Blood-Chin Fanatic target=seat0
[2447] parsed_effect_residual seat=2 source=Blood-Chin Fanatic target=seat0
[2448] activated_ability_resolved seat=2 source=Blood-Chin Fanatic target=seat0
[2449] phase_step seat=2 source= target=seat0
[2450] declare_attackers seat=2 source= target=seat0
[2451] blockers seat=3 source= target=seat0
[2452] damage seat=2 source=Giant Fly amount=2 target=seat3
[2453] damage seat=2 source=Infernal Pet amount=2 target=seat3
[2454] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat3
[2455] phase_step seat=2 source= target=seat0
[2456] stack_push seat=2 source=Desolation target=seat0
[2457] priority_pass seat=3 source= target=seat0
[2458] priority_pass seat=0 source= target=seat0
[2459] priority_pass seat=1 source= target=seat0
[2460] stack_resolve seat=2 source=Desolation target=seat0
[2461] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[2462] state seat=2 source= target=seat0
```

</details>

#### Violation 15

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 48, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 48, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 2530 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2510] priority_pass seat=2 source= target=seat0
[2511] stack_resolve seat=3 source=Mutavault target=seat0
[2512] modification_effect seat=3 source=Mutavault target=seat0
[2513] parser_gap seat=3 source=Mutavault target=seat0
[2514] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2515] pay_mana seat=3 source=Mutavault amount=1 target=seat0
[2516] activate_ability seat=3 source=Mutavault target=seat0
[2517] stack_push seat=3 source=Mutavault target=seat0
[2518] priority_pass seat=0 source= target=seat0
[2519] priority_pass seat=1 source= target=seat0
[2520] priority_pass seat=2 source= target=seat0
[2521] stack_resolve seat=3 source=Mutavault target=seat0
[2522] modification_effect seat=3 source=Mutavault target=seat0
[2523] parser_gap seat=3 source=Mutavault target=seat0
[2524] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2525] phase_step seat=3 source= target=seat0
[2526] phase_step seat=3 source= target=seat0
[2527] zone_change seat=3 source=Locthwain Lancer
[2528] discard seat=3 source=Locthwain Lancer target=seat0
[2529] state seat=3 source= target=seat0
```

</details>

#### Violation 16

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 48, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 48, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 2530 events
  Seat 0 [alive]: life=16 library=73 hand=0 graveyard=11 exile=1 battlefield=13 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0)
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=6 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2510] priority_pass seat=2 source= target=seat0
[2511] stack_resolve seat=3 source=Mutavault target=seat0
[2512] modification_effect seat=3 source=Mutavault target=seat0
[2513] parser_gap seat=3 source=Mutavault target=seat0
[2514] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2515] pay_mana seat=3 source=Mutavault amount=1 target=seat0
[2516] activate_ability seat=3 source=Mutavault target=seat0
[2517] stack_push seat=3 source=Mutavault target=seat0
[2518] priority_pass seat=0 source= target=seat0
[2519] priority_pass seat=1 source= target=seat0
[2520] priority_pass seat=2 source= target=seat0
[2521] stack_resolve seat=3 source=Mutavault target=seat0
[2522] modification_effect seat=3 source=Mutavault target=seat0
[2523] parser_gap seat=3 source=Mutavault target=seat0
[2524] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2525] phase_step seat=3 source= target=seat0
[2526] phase_step seat=3 source= target=seat0
[2527] zone_change seat=3 source=Locthwain Lancer
[2528] discard seat=3 source=Locthwain Lancer target=seat0
[2529] state seat=3 source= target=seat0
```

</details>

#### Violation 17

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 49, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 49, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 2572 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=2 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2552] play_land seat=0 source=Mountain target=seat0
[2553] add_mana seat=0 source=Mountain amount=1 target=seat0
[2554] phase_step seat=0 source= target=seat0
[2555] declare_attackers seat=0 source= target=seat0
[2556] trigger_fires seat=0 source=Wildfire Eternal target=seat0
[2557] stack_push seat=0 source=Wildfire Eternal target=seat0
[2558] priority_pass seat=1 source= target=seat0
[2559] priority_pass seat=2 source= target=seat0
[2560] priority_pass seat=3 source= target=seat0
[2561] stack_resolve seat=0 source=Wildfire Eternal target=seat0
[2562] zone_change seat=0 source=Mountain
[2563] zone_cast_grant_registered seat=0 source=Wildfire Eternal target=seat0
[2564] impulse_play seat=0 source=Wildfire Eternal target=seat0
[2565] blockers seat=2 source= target=seat0
[2566] damage seat=0 source=Spider-Bot amount=2 target=seat2
[2567] damage seat=0 source=Sokka, Bold Boomeranger amount=1 target=seat2
[2568] damage seat=0 source=Wildfire Eternal amount=1 target=seat2
[2569] phase_step seat=0 source= target=seat0
[2570] pool_drain seat=0 source= amount=10 target=seat0
[2571] state seat=0 source= target=seat0
```

</details>

#### Violation 18

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 49, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 12 extra real cards appeared (expected 394, found 406) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 49, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 2572 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=13 library=54 hand=7 graveyard=20 exile=1 battlefield=15 cmdzone=1 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
  Seat 2 [alive]: life=2 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2552] play_land seat=0 source=Mountain target=seat0
[2553] add_mana seat=0 source=Mountain amount=1 target=seat0
[2554] phase_step seat=0 source= target=seat0
[2555] declare_attackers seat=0 source= target=seat0
[2556] trigger_fires seat=0 source=Wildfire Eternal target=seat0
[2557] stack_push seat=0 source=Wildfire Eternal target=seat0
[2558] priority_pass seat=1 source= target=seat0
[2559] priority_pass seat=2 source= target=seat0
[2560] priority_pass seat=3 source= target=seat0
[2561] stack_resolve seat=0 source=Wildfire Eternal target=seat0
[2562] zone_change seat=0 source=Mountain
[2563] zone_cast_grant_registered seat=0 source=Wildfire Eternal target=seat0
[2564] impulse_play seat=0 source=Wildfire Eternal target=seat0
[2565] blockers seat=2 source= target=seat0
[2566] damage seat=0 source=Spider-Bot amount=2 target=seat2
[2567] damage seat=0 source=Sokka, Bold Boomeranger amount=1 target=seat2
[2568] damage seat=0 source=Wildfire Eternal amount=1 target=seat2
[2569] phase_step seat=0 source= target=seat0
[2570] pool_drain seat=0 source= amount=10 target=seat0
[2571] state seat=0 source= target=seat0
```

</details>

#### Violation 19

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 50, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 50, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2666 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=11 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2646] priority_pass seat=0 source= target=seat0
[2647] stack_resolve seat=1 source=Marina Vendrell // Marina Vendrell target=seat0
[2648] zone_change seat=1 source=Marina Vendrell // Marina Vendrell
[2649] resolve seat=1 source=Marina Vendrell // Marina Vendrell target=seat0
[2650] priority_pass seat=2 source= target=seat0
[2651] priority_pass seat=3 source= target=seat0
[2652] priority_pass seat=0 source= target=seat0
[2653] stack_resolve seat=1 source=Beatrix, Loyal General target=seat0
[2654] enter_battlefield seat=1 source=Beatrix, Loyal General target=seat0
[2655] cast seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2656] stack_push seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2657] priority_pass seat=2 source= target=seat0
[2658] priority_pass seat=3 source= target=seat0
[2659] priority_pass seat=0 source= target=seat0
[2660] stack_resolve seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2661] zone_change seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets
[2662] resolve seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2663] phase_step seat=1 source= target=seat0
[2664] phase_step seat=1 source= target=seat0
[2665] state seat=1 source= target=seat0
```

</details>

#### Violation 20

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 50, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 50, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2666 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=11 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=80 hand=1 graveyard=8 exile=0 battlefield=10 cmdzone=1 mana=0
    - Braids's Frightful Return (P/T 0/0, dmg=0)
    - Swamp (P/T 0/0, dmg=0) [T]
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Blood-Chin Fanatic (P/T 3/3, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2646] priority_pass seat=0 source= target=seat0
[2647] stack_resolve seat=1 source=Marina Vendrell // Marina Vendrell target=seat0
[2648] zone_change seat=1 source=Marina Vendrell // Marina Vendrell
[2649] resolve seat=1 source=Marina Vendrell // Marina Vendrell target=seat0
[2650] priority_pass seat=2 source= target=seat0
[2651] priority_pass seat=3 source= target=seat0
[2652] priority_pass seat=0 source= target=seat0
[2653] stack_resolve seat=1 source=Beatrix, Loyal General target=seat0
[2654] enter_battlefield seat=1 source=Beatrix, Loyal General target=seat0
[2655] cast seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2656] stack_push seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2657] priority_pass seat=2 source= target=seat0
[2658] priority_pass seat=3 source= target=seat0
[2659] priority_pass seat=0 source= target=seat0
[2660] stack_resolve seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2661] zone_change seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets
[2662] resolve seat=1 source=Toski, Bearer of Secrets // Toski, Bearer of Secrets target=seat0
[2663] phase_step seat=1 source= target=seat0
[2664] phase_step seat=1 source= target=seat0
[2665] state seat=1 source= target=seat0
```

</details>

#### Violation 21

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 51, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 51, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 2731 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=7 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2711] phase_step seat=2 source= target=seat0
[2712] declare_attackers seat=2 source= target=seat0
[2713] blockers seat=1 source= target=seat0
[2714] damage seat=2 source=Giant Fly amount=2 target=seat1
[2715] damage seat=2 source=Infernal Pet amount=2 target=seat1
[2716] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat1
[2717] damage seat=1 source=Beatrix, Loyal General amount=4 target=seat2
[2718] destroy seat=2 source=Blood-Chin Fanatic
[2719] sba_704_5g seat=2 source=Blood-Chin Fanatic
[2720] zone_change seat=2 source=Blood-Chin Fanatic
[2721] sba_cycle_complete seat=-1 source=
[2722] phase_step seat=2 source= target=seat0
[2723] stack_push seat=2 source=Desolation target=seat0
[2724] priority_pass seat=3 source= target=seat0
[2725] priority_pass seat=0 source= target=seat0
[2726] priority_pass seat=1 source= target=seat0
[2727] stack_resolve seat=2 source=Desolation target=seat0
[2728] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[2729] damage_wears_off seat=1 source=Beatrix, Loyal General amount=3 target=seat0
[2730] state seat=2 source= target=seat0
```

</details>

#### Violation 22

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 51, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 51, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 2731 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=7 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=78 hand=7 graveyard=6 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2711] phase_step seat=2 source= target=seat0
[2712] declare_attackers seat=2 source= target=seat0
[2713] blockers seat=1 source= target=seat0
[2714] damage seat=2 source=Giant Fly amount=2 target=seat1
[2715] damage seat=2 source=Infernal Pet amount=2 target=seat1
[2716] damage seat=2 source=Blood-Chin Fanatic amount=3 target=seat1
[2717] damage seat=1 source=Beatrix, Loyal General amount=4 target=seat2
[2718] destroy seat=2 source=Blood-Chin Fanatic
[2719] sba_704_5g seat=2 source=Blood-Chin Fanatic
[2720] zone_change seat=2 source=Blood-Chin Fanatic
[2721] sba_cycle_complete seat=-1 source=
[2722] phase_step seat=2 source= target=seat0
[2723] stack_push seat=2 source=Desolation target=seat0
[2724] priority_pass seat=3 source= target=seat0
[2725] priority_pass seat=0 source= target=seat0
[2726] priority_pass seat=1 source= target=seat0
[2727] stack_resolve seat=2 source=Desolation target=seat0
[2728] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[2729] damage_wears_off seat=1 source=Beatrix, Loyal General amount=3 target=seat0
[2730] state seat=2 source= target=seat0
```

</details>

#### Violation 23

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 52, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 52, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 2804 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=7 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2784] activate_ability seat=3 source=Mutavault target=seat0
[2785] stack_push seat=3 source=Mutavault target=seat0
[2786] priority_pass seat=0 source= target=seat0
[2787] priority_pass seat=1 source= target=seat0
[2788] priority_pass seat=2 source= target=seat0
[2789] stack_resolve seat=3 source=Mutavault target=seat0
[2790] modification_effect seat=3 source=Mutavault target=seat0
[2791] parser_gap seat=3 source=Mutavault target=seat0
[2792] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2793] cast seat=3 source=Professor Dellian Fel Emblem target=seat0
[2794] stack_push seat=3 source=Professor Dellian Fel Emblem target=seat0
[2795] priority_pass seat=0 source= target=seat0
[2796] priority_pass seat=1 source= target=seat0
[2797] priority_pass seat=2 source= target=seat0
[2798] stack_resolve seat=3 source=Professor Dellian Fel Emblem target=seat0
[2799] zone_change seat=3 source=Professor Dellian Fel Emblem
[2800] resolve seat=3 source=Professor Dellian Fel Emblem target=seat0
[2801] phase_step seat=3 source= target=seat0
[2802] phase_step seat=3 source= target=seat0
[2803] state seat=3 source= target=seat0
```

</details>

#### Violation 24

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 52, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 52, Phase=ending Step=cleanup Active=seat3
Stack: 0 items, EventLog: 2804 events
  Seat 0 [alive]: life=16 library=71 hand=0 graveyard=11 exile=2 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Spider-Bot (P/T 2/1, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
  Seat 1 [alive]: life=7 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2784] activate_ability seat=3 source=Mutavault target=seat0
[2785] stack_push seat=3 source=Mutavault target=seat0
[2786] priority_pass seat=0 source= target=seat0
[2787] priority_pass seat=1 source= target=seat0
[2788] priority_pass seat=2 source= target=seat0
[2789] stack_resolve seat=3 source=Mutavault target=seat0
[2790] modification_effect seat=3 source=Mutavault target=seat0
[2791] parser_gap seat=3 source=Mutavault target=seat0
[2792] activated_ability_resolved seat=3 source=Mutavault target=seat0
[2793] cast seat=3 source=Professor Dellian Fel Emblem target=seat0
[2794] stack_push seat=3 source=Professor Dellian Fel Emblem target=seat0
[2795] priority_pass seat=0 source= target=seat0
[2796] priority_pass seat=1 source= target=seat0
[2797] priority_pass seat=2 source= target=seat0
[2798] stack_resolve seat=3 source=Professor Dellian Fel Emblem target=seat0
[2799] zone_change seat=3 source=Professor Dellian Fel Emblem
[2800] resolve seat=3 source=Professor Dellian Fel Emblem target=seat0
[2801] phase_step seat=3 source= target=seat0
[2802] phase_step seat=3 source= target=seat0
[2803] state seat=3 source= target=seat0
```

</details>

#### Violation 25

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 53, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 53, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 2862 events
  Seat 0 [alive]: life=16 library=69 hand=0 graveyard=12 exile=3 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Celestial Prism (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=5 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2842] priority_pass seat=1 source= target=seat0
[2843] priority_pass seat=2 source= target=seat0
[2844] priority_pass seat=3 source= target=seat0
[2845] stack_resolve seat=0 source=Wildfire Eternal target=seat0
[2846] zone_change seat=0 source=A-Dragonborn Looter
[2847] zone_cast_grant_registered seat=0 source=Wildfire Eternal target=seat0
[2848] impulse_play seat=0 source=Wildfire Eternal target=seat0
[2849] blockers seat=1 source= target=seat0
[2850] damage seat=0 source=Spider-Bot amount=2 target=seat1
[2851] damage seat=0 source=Sokka, Bold Boomeranger amount=1 target=seat1
[2852] damage seat=0 source=Wildfire Eternal amount=1 target=seat1
[2853] damage seat=1 source=Beatrix, Loyal General amount=4 target=seat0
[2854] destroy seat=0 source=Spider-Bot
[2855] sba_704_5g seat=0 source=Spider-Bot
[2856] zone_change seat=0 source=Spider-Bot
[2857] sba_cycle_complete seat=-1 source=
[2858] phase_step seat=0 source= target=seat0
[2859] pool_drain seat=0 source= amount=9 target=seat0
[2860] damage_wears_off seat=1 source=Beatrix, Loyal General amount=2 target=seat0
[2861] state seat=0 source= target=seat0
```

</details>

#### Violation 26

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 53, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 13 extra real cards appeared (expected 394, found 407) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 53, Phase=ending Step=cleanup Active=seat0
Stack: 0 items, EventLog: 2862 events
  Seat 0 [alive]: life=16 library=69 hand=0 graveyard=12 exile=3 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Celestial Prism (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=5 library=51 hand=7 graveyard=22 exile=1 battlefield=17 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2842] priority_pass seat=1 source= target=seat0
[2843] priority_pass seat=2 source= target=seat0
[2844] priority_pass seat=3 source= target=seat0
[2845] stack_resolve seat=0 source=Wildfire Eternal target=seat0
[2846] zone_change seat=0 source=A-Dragonborn Looter
[2847] zone_cast_grant_registered seat=0 source=Wildfire Eternal target=seat0
[2848] impulse_play seat=0 source=Wildfire Eternal target=seat0
[2849] blockers seat=1 source= target=seat0
[2850] damage seat=0 source=Spider-Bot amount=2 target=seat1
[2851] damage seat=0 source=Sokka, Bold Boomeranger amount=1 target=seat1
[2852] damage seat=0 source=Wildfire Eternal amount=1 target=seat1
[2853] damage seat=1 source=Beatrix, Loyal General amount=4 target=seat0
[2854] destroy seat=0 source=Spider-Bot
[2855] sba_704_5g seat=0 source=Spider-Bot
[2856] zone_change seat=0 source=Spider-Bot
[2857] sba_cycle_complete seat=-1 source=
[2858] phase_step seat=0 source= target=seat0
[2859] pool_drain seat=0 source= amount=9 target=seat0
[2860] damage_wears_off seat=1 source=Beatrix, Loyal General amount=2 target=seat0
[2861] state seat=0 source= target=seat0
```

</details>

#### Violation 27

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 54, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 14 extra real cards appeared (expected 394, found 408) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 54, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2956 events
  Seat 0 [alive]: life=12 library=69 hand=0 graveyard=12 exile=3 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Celestial Prism (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=3 library=48 hand=7 graveyard=24 exile=1 battlefield=18 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2936] sba_cycle_complete seat=-1 source=
[2937] play_land seat=1 source=Plains target=seat0
[2938] add_mana seat=1 source=Plains amount=1 target=seat0
[2939] pay_mana seat=1 source=Pollen Remedy amount=1 target=seat0
[2940] cast seat=1 source=Pollen Remedy amount=1 target=seat0
[2941] stack_push seat=1 source=Pollen Remedy target=seat0
[2942] priority_pass seat=2 source= target=seat0
[2943] priority_pass seat=3 source= target=seat0
[2944] priority_pass seat=0 source= target=seat0
[2945] stack_resolve seat=1 source=Pollen Remedy target=seat0
[2946] zone_change seat=1 source=Pollen Remedy
[2947] resolve seat=1 source=Pollen Remedy target=seat0
[2948] phase_step seat=1 source= target=seat0
[2949] declare_attackers seat=1 source= target=seat0
[2950] blockers seat=0 source= target=seat0
[2951] damage seat=1 source=Beatrix, Loyal General amount=4 target=seat0
[2952] phase_step seat=1 source= target=seat0
[2953] zone_change seat=1 source=Protector of the Crown
[2954] discard seat=1 source=Protector of the Crown target=seat0
[2955] state seat=1 source= target=seat0
```

</details>

#### Violation 28

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 54, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 14 extra real cards appeared (expected 394, found 408) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 54, Phase=ending Step=cleanup Active=seat1
Stack: 0 items, EventLog: 2956 events
  Seat 0 [alive]: life=12 library=69 hand=0 graveyard=12 exile=3 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Celestial Prism (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=3 library=48 hand=7 graveyard=24 exile=1 battlefield=18 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=2 library=79 hand=2 graveyard=11 exile=0 battlefield=7 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
  Seat 3 [alive]: life=21 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2936] sba_cycle_complete seat=-1 source=
[2937] play_land seat=1 source=Plains target=seat0
[2938] add_mana seat=1 source=Plains amount=1 target=seat0
[2939] pay_mana seat=1 source=Pollen Remedy amount=1 target=seat0
[2940] cast seat=1 source=Pollen Remedy amount=1 target=seat0
[2941] stack_push seat=1 source=Pollen Remedy target=seat0
[2942] priority_pass seat=2 source= target=seat0
[2943] priority_pass seat=3 source= target=seat0
[2944] priority_pass seat=0 source= target=seat0
[2945] stack_resolve seat=1 source=Pollen Remedy target=seat0
[2946] zone_change seat=1 source=Pollen Remedy
[2947] resolve seat=1 source=Pollen Remedy target=seat0
[2948] phase_step seat=1 source= target=seat0
[2949] declare_attackers seat=1 source= target=seat0
[2950] blockers seat=0 source= target=seat0
[2951] damage seat=1 source=Beatrix, Loyal General amount=4 target=seat0
[2952] phase_step seat=1 source= target=seat0
[2953] zone_change seat=1 source=Protector of the Crown
[2954] discard seat=1 source=Protector of the Crown target=seat0
[2955] state seat=1 source= target=seat0
```

</details>

#### Violation 29

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 55, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 14 extra real cards appeared (expected 394, found 408) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 55, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 3009 events
  Seat 0 [alive]: life=12 library=69 hand=0 graveyard=12 exile=3 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Celestial Prism (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=3 library=48 hand=7 graveyard=24 exile=1 battlefield=18 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=2 library=78 hand=0 graveyard=12 exile=0 battlefield=9 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Isolated Watchtower (P/T 0/0, dmg=0) [T]
    - Abyssal Hunter (P/T 1/1, dmg=0)
  Seat 3 [alive]: life=17 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2989] stack_push seat=2 source=Abyssal Hunter target=seat0
[2990] priority_pass seat=3 source= target=seat0
[2991] priority_pass seat=0 source= target=seat0
[2992] priority_pass seat=1 source= target=seat0
[2993] stack_resolve seat=2 source=Abyssal Hunter target=seat0
[2994] tap seat=0 source=Abyssal Hunter target=seat0
[2995] enter_battlefield seat=2 source=Abyssal Hunter target=seat0
[2996] phase_step seat=2 source= target=seat0
[2997] declare_attackers seat=2 source= target=seat0
[2998] blockers seat=3 source= target=seat0
[2999] damage seat=2 source=Giant Fly amount=2 target=seat3
[3000] damage seat=2 source=Infernal Pet amount=2 target=seat3
[3001] phase_step seat=2 source= target=seat0
[3002] stack_push seat=2 source=Desolation target=seat0
[3003] priority_pass seat=3 source= target=seat0
[3004] priority_pass seat=0 source= target=seat0
[3005] priority_pass seat=1 source= target=seat0
[3006] stack_resolve seat=2 source=Desolation target=seat0
[3007] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[3008] state seat=2 source= target=seat0
```

</details>

#### Violation 30

- **Game**: 174 (seed 1740778, perm 0)
- **Invariant**: ZoneConservation
- **Turn**: 55, Phase=ending Step=cleanup
- **Commanders**: Sokka, Bold Boomeranger, Beatrix, Loyal General, Kain, Traitorous Dragoon, Tivash, Gloom Summoner
- **Message**: zone conservation suspicious: 14 extra real cards appeared (expected 394, found 408) — possible copy bug

<details>
<summary>Game State</summary>

```
Turn 55, Phase=ending Step=cleanup Active=seat2
Stack: 0 items, EventLog: 3009 events
  Seat 0 [alive]: life=12 library=69 hand=0 graveyard=12 exile=3 battlefield=14 cmdzone=0 mana=0
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Mountain (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Tyrite Sanctum (P/T 0/0, dmg=0) [T]
    - Sokka, Bold Boomeranger (P/T 1/1, dmg=0) [T]
    - Island (P/T 0/0, dmg=0) [T]
    - Strip Mine (P/T 0/0, dmg=0) [T]
    - Wildfire Eternal (P/T 1/4, dmg=0) [T]
    - Mountain (P/T 0/0, dmg=0) [T]
    - Celestial Prism (P/T 0/0, dmg=0)
  Seat 1 [alive]: life=3 library=48 hand=7 graveyard=24 exile=1 battlefield=18 cmdzone=0 mana=0
    - Plains (P/T 0/0, dmg=0) [T]
    - Wanderbrine Trapper (P/T 2/1, dmg=0) [T]
    - Forbidden Orchard (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Library of Alexandria (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Stalking Stones (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
    - Zephyr Boots (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Blessing (P/T 0/0, dmg=0)
    - Plains (P/T 0/0, dmg=0) [T]
    - Beatrix, Loyal General (P/T 4/4, dmg=0) [T]
    - Plains (P/T 0/0, dmg=0) [T]
  Seat 2 [alive]: life=2 library=78 hand=0 graveyard=12 exile=0 battlefield=9 cmdzone=1 mana=0
    - Bloodthirsty Ogre (P/T 3/1, dmg=0) [T]
    - Desolation (P/T 0/0, dmg=0)
    - Crystal Vein (P/T 0/0, dmg=0) [T]
    - Giant Fly (P/T 2/2, dmg=0) [T]
    - Infernal Pet (P/T 2/2, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Isolated Watchtower (P/T 0/0, dmg=0) [T]
    - Abyssal Hunter (P/T 1/1, dmg=0)
  Seat 3 [alive]: life=17 library=77 hand=7 graveyard=7 exile=0 battlefield=6 cmdzone=1 mana=0
    - Swamp (P/T 0/0, dmg=0) [T]
    - Mutavault (P/T 0/0, dmg=0) [T]
    - Gift of Fangs (P/T 0/0, dmg=0)
    - Phyrexian Scrapyard (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]
    - Swamp (P/T 0/0, dmg=0) [T]

```

</details>

<details>
<summary>Recent Events</summary>

```
[2989] stack_push seat=2 source=Abyssal Hunter target=seat0
[2990] priority_pass seat=3 source= target=seat0
[2991] priority_pass seat=0 source= target=seat0
[2992] priority_pass seat=1 source= target=seat0
[2993] stack_resolve seat=2 source=Abyssal Hunter target=seat0
[2994] tap seat=0 source=Abyssal Hunter target=seat0
[2995] enter_battlefield seat=2 source=Abyssal Hunter target=seat0
[2996] phase_step seat=2 source= target=seat0
[2997] declare_attackers seat=2 source= target=seat0
[2998] blockers seat=3 source= target=seat0
[2999] damage seat=2 source=Giant Fly amount=2 target=seat3
[3000] damage seat=2 source=Infernal Pet amount=2 target=seat3
[3001] phase_step seat=2 source= target=seat0
[3002] stack_push seat=2 source=Desolation target=seat0
[3003] priority_pass seat=3 source= target=seat0
[3004] priority_pass seat=0 source= target=seat0
[3005] priority_pass seat=1 source= target=seat0
[3006] stack_resolve seat=2 source=Desolation target=seat0
[3007] each_player_effect seat=2 source=Desolation amount=4 target=seat0
[3008] state seat=2 source= target=seat0
```

</details>

*... and 590 more violations not shown.*

## Invariant Violations (Nightmare Boards)

| Invariant | Count |
|-----------|-------|
| CardIdentity | 2 |

## Top Cards Correlated with Violations

Cards that appeared disproportionately in violation games vs clean games.
Only cards appearing in 3+ total games are shown.

| Rank | Card | Violation Games | Clean Games | Correlation |
|------|------|-----------------|-------------|-------------|
| 1 | Pollenbright Wings | 2 | 1 | 0.67 |
| 2 | Hamza, Might of the Yathan | 2 | 2 | 0.50 |
| 3 | Brood Butcher | 2 | 3 | 0.40 |
| 4 | Kraul Whipcracker | 1 | 2 | 0.33 |
| 5 | Thoughtflare | 1 | 2 | 0.33 |
| 6 | Wayfaring Temple | 1 | 2 | 0.33 |
| 7 | Vish Kal, Blood Arbiter | 2 | 4 | 0.33 |
| 8 | The Infamous Cruelclaw | 2 | 4 | 0.33 |
| 9 | Leafkin Avenger | 1 | 2 | 0.33 |
| 10 | Kumena, Tyrant of Orazca | 1 | 2 | 0.33 |

## Verdict: ISSUES FOUND

**622 total issues** across 2000 chaos games and 10000 nightmare boards.
- 0 crashes in chaos games
- 620 invariant violations in chaos games
- 0 crashes in nightmare boards
- 2 invariant violations in nightmare boards

Review the details above to identify which cards and interactions are problematic.
