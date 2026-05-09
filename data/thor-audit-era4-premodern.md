# Thor 2.1 Corpus Audit — Era 4 (Pre-Modern/Legacy: pre-2003)

Audit-only scan against `cmd/hexdek-thor/conditional_setup.go` —
for every pre-2003 card (Alpha through Onslaught block), every triggered
ability is matched against the 19 trigger slugs in `triggerConditionActions`,
and every condition node (intervening_if / static.condition / effect.Conditional)
is matched against the `conditionScaffoldKind` rules in `detectConditionScaffold`.
Static vanilla abilities (plain keywords, anthems) are excluded.

Era boundary: `released_at < 2003-01-01` from `data/rules/oracle-cards.json`.

## Summary

- Era 4 cards (released before 2003-01-01, present in AST corpus): **3427**
- Cards with at least one conditional/triggered ability: **1066**
- Cards with at least one ability bucketed by an existing scaffold: **820**
- Cards with at least one *flagged* (unmatched) ability: **246**
- Cards fully bucketed (no flagged abilities): **820**

- Total ability/condition nodes scanned: **1275**
- Bucketed nodes: **1015** (79.6%)
- Flagged nodes: **260** (20.4%)

## Top scaffold kinds by frequency (top 15)

| Rank | Scaffold | Hits | Examples |
|---:|---|---:|---|
| 1 | `TRIG:upkeep` | 189 | Aether Rift, Afiya Grove, Aku Djinn, Ana Sanctuary, Anurid Scavenger |
| 2 | `TRIG:creature_etb` | 178 | Abduction, Abyssal Horror, Accursed Centaur, Aether Charge, Aether Flash |
| 3 | `TRIG:attacks` | 164 | Abomination, Abyssal Nightstalker, Aisling Leprechaun, Alaborn Zealot, Alley Grifters |
| 4 | `TRIG:creature_dies` | 84 | Abduction, Abu Ja'far, Academy Rector, Alabaster Dragon, Aphetto Vulture |
| 5 | `TRIG:combat_damage` | 78 | Avenging Druid, Backfire, Balshan Beguiler, Bellowing Fiend, Binding Agony |
| 6 | `COND:STRUCTURED:paid_optional_cost` | 53 | Academy Rector, Auspicious Ancestor, Avenging Druid, Bone Dancer, Customs Depot |
| 7 | `COND:STRUCTURED:threshold` | 46 | Aboshan's Desire, Anurid Barkripper, Aven Warcraft, Battlewise Aven, Bloodcurdler |
| 8 | `TRIG:cast_spell` | 41 | Aether Barrier, Auspicious Ancestor, Aven Shrine, Bazaar of Wonders, Bog Gnarr |
| 9 | `TRIG:ltb` | 40 | Acidic Dagger, Ancestral Knowledge, Brood of Cockroaches, Corrosion, Delusions of Mediocrity |
| 10 | `TRIG:end_step` | 26 | Aggression, Asmira, Holy Avenger, Blistering Firecat, Blood Hound, Crystal Golem |
| 11 | `COND:STRUCTURED:for_each` | 26 | Ashuza's Breath, Bend or Break, Cleansing, Delaying Shield, Equipoise |
| 12 | `TRIG:opponent_cast` | 25 | Aether Sting, Freyalise's Charm, Havoc, Hidden Spider, Ichneumon Druid |
| 13 | `COND:STRUCTURED:domain` | 9 | Draco, Kavu Scout, Magnigoth Treefolk, Ordered Migration, Planar Despair |
| 14 | `TRIG:begin_combat` | 7 | Battering Ram, Goblin Flotilla, Greater Werewolf, Heat Stroke, Johan |
| 15 | `COND:STRUCTURED:did_prior_action` | 5 | Jeweled Torque, Pedantic Learning, Purgatory, Shifty Doppelganger, Standstill |

## Bucketed breakdown (all scaffolds)

Each row is a scaffold key — `TRIG:` for trigger slugs (from `classifyTrigger`),
`COND:` for condition scaffold kinds (from `detectConditionScaffold`).
`STRUCTURED:` rows are AST condition kinds.

| Scaffold | Hits | Examples |
|---|---:|---|
| `TRIG:upkeep` | 189 | Aether Rift, Afiya Grove, Aku Djinn, Ana Sanctuary, Anurid Scavenger |
| `TRIG:creature_etb` | 178 | Abduction, Abyssal Horror, Accursed Centaur, Aether Charge, Aether Flash |
| `TRIG:attacks` | 164 | Abomination, Abyssal Nightstalker, Aisling Leprechaun, Alaborn Zealot, Alley Grifters |
| `TRIG:creature_dies` | 84 | Abduction, Abu Ja'far, Academy Rector, Alabaster Dragon, Aphetto Vulture |
| `TRIG:combat_damage` | 78 | Avenging Druid, Backfire, Balshan Beguiler, Bellowing Fiend, Binding Agony |
| `COND:STRUCTURED:paid_optional_cost` | 53 | Academy Rector, Auspicious Ancestor, Avenging Druid, Bone Dancer, Customs Depot |
| `COND:STRUCTURED:threshold` | 46 | Aboshan's Desire, Anurid Barkripper, Aven Warcraft, Battlewise Aven, Bloodcurdler |
| `TRIG:cast_spell` | 41 | Aether Barrier, Auspicious Ancestor, Aven Shrine, Bazaar of Wonders, Bog Gnarr |
| `TRIG:ltb` | 40 | Acidic Dagger, Ancestral Knowledge, Brood of Cockroaches, Corrosion, Delusions of Mediocrity |
| `TRIG:end_step` | 26 | Aggression, Asmira, Holy Avenger, Blistering Firecat, Blood Hound, Crystal Golem |
| `COND:STRUCTURED:for_each` | 26 | Ashuza's Breath, Bend or Break, Cleansing, Delaying Shield, Equipoise |
| `TRIG:opponent_cast` | 25 | Aether Sting, Freyalise's Charm, Havoc, Hidden Spider, Ichneumon Druid |
| `COND:STRUCTURED:domain` | 9 | Draco, Kavu Scout, Magnigoth Treefolk, Ordered Migration, Planar Despair |
| `TRIG:begin_combat` | 7 | Battering Ram, Goblin Flotilla, Greater Werewolf, Heat Stroke, Johan |
| `COND:STRUCTURED:did_prior_action` | 5 | Jeweled Torque, Pedantic Learning, Purgatory, Shifty Doppelganger, Standstill |
| `COND:STRUCTURED:self_is_tapped` | 5 | Bottomless Vault, Dwarven Hold, Hollow Trees, Icatian Store, Sand Silos |
| `COND:CardInGraveyard` | 3 | Lotus Vale, Scorched Ruins, Wand of Denial |
| `TRIG:opponent_discards` | 3 | Mangara's Blessing, Metrognome, Sand Golem |
| `TRIG:draw_step` | 3 | Grafted Skullcap, Heightened Awareness, Wave of Terror |
| `TRIG:sacrifice` | 3 | Last Laugh, Orcish Mine, Task Mage Assembly |
| `TRIG:discard` | 3 | Confessor, Pitchstone Wall, Telekinetic Bonds |
| `COND:STRUCTURED:if` | 2 | Bogardan Phoenix, Soul Echo |
| `COND:CreatureCardsInGraveyard` | 2 | Bloodline Shaman, Zoologist |
| `COND:STRUCTURED:no_creatures_on_battlefield` | 2 | Pestilence, Task Mage Assembly |
| `TRIG:draw_card` | 1 | Fasting |
| `COND:STRUCTURED:you_control` | 1 | Ceta Sanctuary |
| `TRIG:lose_life` | 1 | Oath of Lim-Dûl |

## Flagged clusters (proposed new scaffold kinds)

Each cluster groups trigger events or condition texts that no existing scaffold
matched. The cluster tag is derived from the raw AST event name or condition text.
Below each table is a card sample with the raw event and ability context.

---

### `trigger:AnyPlayerUpkeep` — 42 hits, 42 unique cards

**Pattern:** `event="phase" phase="player"` — fires at the beginning of **each player's**
upkeep/end step. The existing `upkeep` scaffold only primes for the active player's
upkeep; these cards fire for every player in turn order (symmetrical group triggers).

**Proposed kind:** `any_player_phase` — no new game state required; prime by advancing
phase for both seats.

Examples:

- **Ancient Runes** (ability[0] triggered) — `"At the beginning of each player's upkeep, this enchantment deals damage to that player equal to the number of artifacts they control."`
- **Citadel of Pain** (ability[0] triggered) — `"At the beginning of each player's end step, this enchantment deals X damage to that player, where X is the number of untapped lands they control."`
- **Ley Line** (ability[0] triggered) — `"At the beginning of each player's upkeep, that player may put a +1/+1 counter on target creature they control."`
- **Mana Cache** (ability[0] triggered) — `"At the beginning of each player's end step, put a charge counter on this enchantment for each unused mana that player has."`
- **Antagonism**, **Barbed Wire**, **Bottomless Pit**, **Copper Tablet**, **Freyalise's Winds**

---

### `trigger:DelayedTrigger_NextTurn` — 34 hits, 34 unique cards

**Pattern:** `event="phase" phase="next_turn"` — a "draw a card at the beginning of the
next turn's upkeep" delayed trigger. Extremely common in Mirage/Visions/Weatherlight
as a drawback offset for instants and auras.

**Proposed kind:** `delayed_draw_next_upkeep` — prime by scheduling a draw event for
the upcoming upkeep step.

Examples:

- **Bone Harvest** (ability[1] triggered) — `"Draw a card at the beginning of the next turn's upkeep."`
- **Telim'Tor's Edict** (ability[1] triggered) — `"Draw a card at the beginning of the next turn's upkeep."`
- **Jolt** (ability[1] triggered) — `"Draw a card at the beginning of the next turn's upkeep."`
- **Vampirism** (ability[1] triggered) — ETB aura delayed draw
- **Ritual of Steel**, **Carrier Pigeons**, **Clairvoyance**, **Blessed Wine**, **Aleatory**

---

### `trigger:ETBAsChoice` — 32 hits, 32 unique cards

**Pattern:** `event="etb_as"` — the card's ETB resolves with a **modal choice** (`As ~
enters, choose a [type/color/player]`). These are distinct from simple ETBs because
they require a selection before the permanent enters. Different from `etb_as`
condition scaffold — this is a trigger event, not an intervening-if.

**Proposed kind:** `etb_modal_choice` — already partially handled by
`condScaffoldETBAs`; the trigger-event form needs explicit priming for the choose
action so the engine commits a deterministic selection.

Examples:

- **Circle of Solace** (ability[0] triggered) — `"As this enchantment enters, choose a creature type."`
- **Harsh Judgment** (ability[0] triggered) — `"As this enchantment enters, choose a color."`
- **Callous Oppressor** (ability[0] triggered) — `"As this creature enters, choose a creature type an opponent controls."`
- **Belbe's Portal** (ability[0] triggered) — `"As this artifact enters, choose a creature type."`
- **Alloy Golem**, **Chameleon Spirit**, **Camato Scout**, **Chimeric Idol**, **Crumbling Sanctuary**

---

### `trigger:BecomesTapped` — 16 hits, 16 unique cards

**Pattern:** `event="becomes_tapped"` — fires whenever a specific permanent
(usually enchanted creature/artifact/land) becomes tapped. Classic Mirage/Ice Age
enchantment designs. No existing scaffold for becomes-tapped events.

**Proposed kind:** `becomes_tapped` — prime by placing the enchanted/target permanent
in untapped state; fireTriggerEvent taps it.

Examples:

- **Insolence** (ability[0] triggered) — `"Whenever enchanted creature becomes tapped, this Aura deals 2 damage to that creature's controller."`
- **Relic Bind** (ability[0] triggered) — `"Whenever enchanted artifact becomes tapped, choose one — You gain 1 life; or this enchantment deals 1 damage to target player."`
- **Lifetap** (ability[0] triggered) — `"Whenever a Forest an opponent controls becomes tapped, you gain 1 life."`
- **Seizures** (ability[0] triggered) — `"Whenever enchanted creature becomes tapped, this Aura deals 3 damage to that creature's controller."`
- **Roots of Life**, **Betrayal**, **Artifact Possession**, **Haunting Wind**, **Relic Ward**

---

### `trigger:BecomesTarget` — 14 hits, 14 unique cards

**Pattern:** `event="becomes_target"` — fires when the creature/permanent becomes the
target of a spell or ability. Classic "ward-like" pre-modern mechanic for
self-protection or retaliation triggers. Appears across Mirage, Mercadian Masques,
Odyssey blocks.

**Proposed kind:** `becomes_target` — prime by putting a targeting spell on the stack
pointing at srcPerm.

Examples:

- **Tar Pit Warrior** (ability[0] triggered) — `"When this creature becomes the target of a spell or ability, sacrifice it."`
- **Cursed Monstrosity** (ability[1] triggered) — `"Whenever this creature becomes the target of a spell or ability, sacrifice it unless you pay {2}."`
- **Cephalid Illusionist** (ability[1] triggered) — `"Whenever this creature becomes the target of a spell or ability, mill three cards."`
- **Skulking Fugitive** (ability[0] triggered) — `"When this creature becomes the target of a spell or ability, sacrifice it."`
- **Forsaken Wastes**, **Fugitive Druid**, **Retromancer**, **Cephalid Aristocrat**

---

### `trigger:BeginningOfOrdinalStep` — 16 hits, 16 unique cards

**Pattern:** `event="beginning_of_ordinal_step"` — fires at the beginning of a specific
non-upkeep step (combat, end step, draw, etc.) that isn't covered by the existing
`begin_combat` or `end_step` slugs. Includes `beginning of combat`, `beginning of
your end step`, `beginning of your draw step` variants.

**Proposed kind:** Already proposed in Era 1 and Era 2 reports as
`condScaffoldBeginningOfOrdinalStep`. This is the trigger-event form of the same
pattern; the condition form is already scaffolded. The trigger event form needs a
matching registry entry.

Examples:

- **Arcum's Whistle** (ability[0] triggered) — beginning of upkeep/combat variant
- **Elkin Lair** (ability[0] triggered) — each player's draw step
- **Cycle of Life** (ability[0] triggered) — beginning of your end step
- **False Memories** (ability[0] triggered) — beginning of draw step
- **Infinite Authority**, **Mindstab Thrull**, **Petra Sphinx**, **Prismatic Boon**

---

### `trigger:UntilEOTDelayed` — 13 hits, 13 unique cards

**Pattern:** Two sub-patterns collapsed:
1. `event="until_eot_trigger"` (4 cards) — instant-speed spells that grant "until end
   of turn, whenever X happens, Y" delayed triggered abilities.
2. `event="phase" phase="next_cleanup"` (9 cards) — instant/aura effects that resolve
   "at the beginning of the next cleanup step" (cards cast as if they had flash;
   common Ice Age / Mirage design).

**Proposed kind:** `until_eot_delayed` — prime by entering the end-of-turn step with
the delayed trigger in scope.

Examples (until_eot_trigger):
- **Spiritualize** — `"Until end of turn, whenever target creature deals damage, you gain that much life."`
- **Bubbling Muck** — `"Until end of turn, whenever a player taps a Swamp for mana, that player adds an additional {B}."`
- **Gaze of Pain**, **False Cure**

Examples (next_cleanup):
- **Spider Climb**, **Mystic Veil**, **Parapet**, **Grave Servitude**, **Soar**, **Cunning**, **Relic Ward**

---

### `trigger:OpponentUpkeep` — 8 hits, 8 unique cards

**Pattern:** `event="phase" phase="opponent"` — fires at the beginning of each
**opponent's** upkeep. Distinguished from `any_player_phase` (phase="player") by
targeting opponents specifically rather than all players.

**Proposed kind:** `opponent_upkeep` — prime by advancing to seat 1's upkeep step.

Examples:

- **Rackling** (ability[0] triggered) — `"At the beginning of each opponent's upkeep, this creature deals X damage to that player, where X is 3 minus the number of cards in their hand."`
- **Paupers' Cage** — `"At the beginning of each opponent's upkeep, if that player has two or fewer cards in hand, this enchantment deals 2 damage to them."`
- **Misers' Cage**, **Wheel of Torture**, **Psychic Allergy**, **Dark Suspicions**, **Iron Maiden**, **Malignant Growth**

---

### `trigger:LandPlay` — 11 hits, 11 unique cards

**Pattern:** Three related land-event slugs collapsed into one cluster:
- `event="player_land_play"` (4 cards) — whenever a player plays a land
- `event="tapped_for_mana"` (4 cards) — whenever a land is tapped for mana
- `event="any_player_tap_land"` (3 cards) — whenever any player taps a land for mana

All three are land-interaction triggers that have no equivalent in the existing
scaffold. Note: `event="cycle"` (cycling — see below) is separate.

**Proposed kind:** `land_play_or_tap` — prime by placing lands in both seats and
advancing to main phase where land-for-mana activation can be simulated.

Examples (player_land_play):
- **Pangosaur** — `"Whenever a player plays a land, return this creature to its owner's hand."`
- **Overburden** — `"Whenever a player puts a nontoken creature onto the battlefield, that player returns a land they control to its owner's hand."`

Examples (tapped_for_mana):
- **Storm Cauldron** — `"Whenever a land is tapped for mana, return it to its owner's hand."`
- **Mana Web** — `"Whenever a land an opponent controls is tapped for mana, tap all lands that player controls that could produce any type of mana the tapped land could produce."`
- **Snowfall**, **Elvish Guidance**, **Scald**, **Price of Glory**, **Overabundance**, **Chaos Moon**

---

### `trigger:Cycling` — 5 hits, 5 unique cards

**Pattern:** `event="cycle"` (3 cards) and `event="any_cycle"` (2 cards) — fires when
a cycling cost is paid. The existing `condScaffoldCycled` handles the condition form;
this is the trigger-event form requiring a cycling action to be simulated.

**Proposed kind:** Already proposed as `condScaffoldCycled` in Tier 2B. The trigger
event form needs a registry entry in `triggerConditionActions`.

Examples:

- **Death Pulse** — `"Cycling {1}{B}{B}. When you cycle this card, target creature gets -4/-4 until end of turn."`
- **Sunfire Balm** — `"Cycling {1}{W}. When you cycle this card, prevent the next 4 damage..."`
- **Complicate** — `"Cycling {2}{U}. When you cycle this card, you may counter target spell..."`
- **Fleeting Aven** — `"Whenever a player cycles a card, return this creature to its owner's hand."`
- **Withering Hex** — `"Whenever a player cycles a card, put a plague counter on this Aura."`

---

### `trigger:Phasing` — 4 hits, 4 unique cards

**Pattern:** `event="self_phase_inout"` — fires when the permanent phases in or out.
Pre-Modern phasing mechanic (Ice Age/Mirage). No existing scaffold.

**Proposed kind:** `self_phase_inout` — prime by toggling the permanent's phased state
using the engine's phase-out path.

Examples:

- **Shimmering Efreet** — `"Phasing. Whenever this phases out, you may choose new targets for any spells or abilities targeting it."`
- **Ertai's Familiar** — `"Phasing. Whenever this phases in, untap target land."`
- **Warping Wurm** — `"Phasing. Whenever this phases in, put a +1/+1 counter on it."`
- **Teferi's Imp** — `"Phasing. Whenever this phases in or out, discard a card at random."`

---

### `trigger:BecomesTargetByAlly` — 3 hits, 3 unique cards

**Pattern:** `event="ally_targeted_by_opp"` — fires when you or a permanent you
control becomes the target of a spell or ability **an opponent controls**. Subset of
`becomes_target` specifically filtering for opponent-sourced targeting.

**Proposed kind:** `ally_targeted_by_opp` — prime with opponent targeting spell on
a friendly permanent.

Examples:

- **Rayne, Academy Chancellor** — `"Whenever you or a permanent you control becomes the target of a spell or ability an opponent controls, you may draw a card."`
- **Mossdog** — `"Whenever this creature becomes the target of a spell or ability an opponent controls, put a +1/+1 counter on it."`
- **Cloud Cover** — `"Whenever another permanent you control becomes the target of a spell or ability an opponent controls, return that permanent to its owner's hand."`

---

### `trigger:CounterThreshold` — 2 hits, 2 unique cards

**Pattern:** `event="counter_threshold"` — fires when the counter count on a permanent
reaches a specific number. Unique to early-Magic tide-counter mechanics (Homarid,
Tidal Influence). No equivalent in modern scaffolding.

**Proposed kind:** `counter_threshold` — prime by placing exactly N counters on srcPerm
(where N is the threshold encoded in the trigger args).

Examples:

- **Homarid** — `"At the beginning of your upkeep, put a tide counter on Homarid. When Homarid has exactly one tide counter on it, it gets -1/-1. When it has exactly three tide counters, it gets +1/+1 and has islandwalk. When it has four or more, return it to its owner's hand."`
- **Tidal Influence** — Same pattern on an enchantment.

---

### `trigger:OppLandfall` — 2 hits, 2 unique cards

**Pattern:** `event="opp_landfall"` — fires when an opponent plays a land. Landfall
for the opponent side, no current scaffold.

**Proposed kind:** `opp_landfall` — prime by playing a land from seat 1's hand.

Examples:

- **Dirtcowl Wurm** — `"Whenever an opponent plays a land, put a +1/+1 counter on Dirtcowl Wurm."`
- **Hidden Stag** — `"Whenever an opponent plays a land, if this permanent is an enchantment, it becomes a 3/2 Elk creature that is still an enchantment."`

---

### `trigger:MiscWhen` — 10 hits, 10 unique cards

**Pattern:** Two catch-all event slugs from the parser:
- `event="misc_when"` (6 cards) — one-off trigger conditions the parser couldn't
  canonicalize: regeneration events, token-creation chaining, specific activated
  ability triggers
- `event="misc_whenever_a"` (4 cards) — "whenever a [thing] happens" triggers the
  parser gave up on: creature-goes-to-graveyard from anywhere, land-to-graveyard,
  nontoken-creature-enters variants

These do NOT cluster into a single proposed scaffold — each card's mechanics is
unique enough to require per-card handler coverage rather than a generic scaffold.

Examples (misc_when):
- **Matopi Golem** — `"When it regenerates this way, put a -1/-1 counter on it."`
- **Soldevi Sentry** — `"When it regenerates this way, that opponent may draw a card."`
- **Skeleton Scavengers** — activated ability counting +1/+1 counters
- **Splintering Wind**, **Soulgorger Orgg**, **Sandals of Abdallah**

Examples (misc_whenever_a):
- **Mortuary** — `"Whenever a creature is put into your graveyard from the battlefield, put that card on top of your library."`
- **Grim Feast** — creature-to-any-graveyard trigger
- **Pedantic Learning** — land-to-graveyard-from-library trigger
- **Bomb Squad** — ability cascade trigger

---

### `trigger:SmallSingletons` — 14 hits, 14 unique cards

Low-frequency flagged events that don't cluster into multi-card patterns:

| Event | Count | Cards |
|---|---:|---|
| `phase\|their_next` | 2 | Nafs Asp, Sabertooth Cobra (poison delayed effect) |
| `nontoken_creature_event` | 2 | Dual Nature, Purgatory (nontoken creature ETB/dies) |
| `phase\|combat_on` | 2 | Fight or Flight, Web of Inertia (during combat) |
| `phase\|that_turn` | 2 | Final Fortune, Oracle en-Vec (end of that turn) |
| `you_whenever` | 2 | Hidden Stag (form-change), Rowen (reveal basic land) |
| `any_cycle` | 2 | Fleeting Aven, Withering Hex (any player cycles) |
| `nontoken_ally_event` | 1 | Remembrance (nontoken creature dies, yours) |
| `tap_for_mana` | 1 | Savage Firecat (cumulative upkeep triggered at tap) |
| `each_upkeep` | 1 | Power Struggle (each player's upkeep, controller contest) |
| `cumulative_upkeep_unpaid` | 1 | Heart of Bogardan (if cumulative upkeep goes unpaid) |
| `becomes_state` | 1 | Steam Vines (land becomes tapped — same as becomes_tapped) |
| `phase\|of_your` | 1 | Carpet of Flowers (each of your turns) |
| `phase\|of_that` | 1 | Ertai's Meddling (at end of that turn) |
| `targets_chosen` | 1 | Psychic Battle (when targets are chosen for a spell) |
| `all_trigger` | 1 | Mob Mentality (when all creatures attack) |
| `you_action` | 1 | Juju Bubble (when you gain life) |
| `self_and_another` | 1 | Rotlung Reanimator (this or another Cleric dies) |
| `land_tapped_for_mana` | 1 | Chaos Moon (when a mountain is tapped for mana) |
| `phase\|chosen_player` | 1 | Energy Vortex (each chosen player's upkeep) |
| `conditional_state` | 2 | Afiya Grove, Keldon Battlewagon (state-triggered removal) |
| `upkeep` (unexpected slug) | 1 | Drought (misclassified cumulative upkeep) |

---

### Flagged conditions — proposed new scaffold kinds

All 20 flagged condition nodes came from static ability `kind="conditional"` (raw text
clauses). Unlike the trigger flagged clusters, these are **individual card mechanics**
— no multi-card cluster hit more than 4 nodes.

#### `CounterExactCount` — 4 nodes, 2 cards

**Pattern:** `"as long as there is/are exactly N tide counter(s) on this"` — the
permanent has a counter-based state machine where each exact count activates a
different ability. Standard boolean scaffolds don't cover "exactly N" semantics.

**Proposed kind:** `counter_exact_count` — prime by placing exactly N counters of the
specified type on srcPerm, where N is parsed from the condition text.

Cards: **Homarid** (2 abilities), **Tidal Influence** (2 abilities)

---

#### `SelfUntappedCondition` — 3 nodes, 3 cards

**Pattern:** `"as long as this [artifact/creature] is untapped, [effect]"` — static
grant conditional on the source's own tapped state. Foundational lockdown mechanic
for artifacts (Static Orb, Spectral Guardian).

**Proposed kind:** `self_untapped_condition` — already partially covered by
`condScaffoldSelfIsTapped` (which handles the *tapped* direction). The *untapped*
direction needs a mirror scaffold: ensure srcPerm is untapped before snapshot.

Cards: **Static Orb**, **Watchdog**, **Spectral Guardian**

---

#### `RevealedCardType` — 2 nodes, 2 cards

**Pattern:** `"if it's a [creature/land/nonland] card, [A]. otherwise, [B]"` — revealed
or top-of-library card type branch. Standard conditional-reveal mechanic from Odyssey.

**Proposed kind:** `revealed_card_type_branch` — already partially covered by
`condScaffoldCardInGraveyard`; needs extension for top-of-library reveal context.

Cards: **Search for Survivors**, **Zoologist** (note: Bloodline Shaman is a separate case)

---

#### `RevealedLandType` — 2 nodes, 2 cards

**Pattern:** `"if that card is a land card, [destroy/debuff]"` — random discard reveal
type check (Chaos Harlequin, Paroxysm). Random discard reveals a card and branches
on whether it's a land.

**Proposed kind:** `revealed_land_type_check` — prime by placing a non-land card on top
of the library or in hand before the reveal action resolves.

Cards: **Chaos Harlequin**, **Paroxysm**

---

#### `Singletons (one card each)` — 10 nodes, 10 cards

| Pattern | Card | Raw text |
|---|---|---|
| Self-untap skip step | **Fasting** | `"if you would begin your draw step, you may skip that step instead"` |
| Opponent flying check | **Escaped Shapeshifter** | `"as long as an opponent controls a creature with flying not named ~, this creature has flying"` |
| Attack count in snow lands | **Snowblind** | `"if that creature is attacking, X is the number of snow lands defending player controls"` |
| Return-to-hand conditional | **Puppet Master** | `"if that card is returned to its owner's hand this way, you may pay..."` |
| No shell counters check | **Roc Hatchling** | `"as long as this creature has no shell counters on it, it gets +3/+2 and has flying"` |
| Revealed card is land (gain life) | **Prophecy** | `"if it's a land, you gain 1 life. then that player shuffles"` |
| All permanent colors | **Spirit of Resistance** | `"as long as you control a permanent of each color, prevent all damage"` |
| Color enchanted permanent | **Essence Leak** | `"as long as enchanted permanent is red or green, it has..."` |
| While-attacking damage prevention | **Camel** | `"as long as this creature is attacking, prevent all damage Deserts would deal to it"` |
| Physical flip mechanic | **Chaos Orb** | `"if this artifact turns over completely at least once during the flip, destroy all..."` |
| ETB replacement land sacrifice | **Lotus Vale** / **Scorched Ruins** | `"if this land would enter, sacrifice two untapped lands"` (these are bucketed via `CardInGraveyard` but are actually ETB-replacement conditions) |

---

## Proposed new scaffold kinds (synthesis)

Distilled from the flagged clusters above. Each row maps a working name to the
cluster keys carrying its events/clauses. Use as the starting point for new
`condScaffold*` constants or `triggerConditionActions` entries.

| Proposed kind | Description | Source clusters | Est. hit count |
|---|---|---|---:|
| **AnyPlayerPhase** | Phase fires for each player in turn (not just active player) | `phase\|player` trigger | 42 |
| **DelayedDrawNextUpkeep** | "Draw a card at the beginning of the next turn's upkeep" delayed trigger | `phase\|next_turn` trigger | 34 |
| **ETBModalChoice** | ETB resolves with a modal choose action (choose a type/color/player) | `etb_as` trigger event | 32 |
| **BecomesTapped** | Trigger fires when a specific permanent becomes tapped | `becomes_tapped` trigger | 16 |
| **BeginningOfOrdinalStep** (trigger-event form) | beginning of combat/draw/end step as trigger event (not condition) | `beginning_of_ordinal_step` trigger | 16 |
| **BecomesTarget** | Trigger fires when permanent becomes the target of a spell or ability | `becomes_target` trigger | 14 |
| **UntilEOTDelayed** | Delayed trigger that fires until end of turn or at next cleanup | `until_eot_trigger`, `phase\|next_cleanup` | 13 |
| **LandPlayOrTap** | Player plays/taps a land for mana trigger | `player_land_play`, `tapped_for_mana`, `any_player_tap_land` | 11 |
| **MiscWhenOnce** | Parser gave up — per-card handlers required | `misc_when`, `misc_whenever_a` | 10 |
| **OpponentUpkeep** | Phase fires at each opponent's upkeep | `phase\|opponent` | 8 |
| **Cycling** (trigger-event form) | Cycling event fired as trigger (not condition scaffold) | `cycle`, `any_cycle` | 5 |
| **Phasing** | Permanent phases in or out trigger | `self_phase_inout` | 4 |
| **AllyTargetedByOpp** | Ally permanent becomes target of opponent spell/ability | `ally_targeted_by_opp` | 3 |
| **OppLandfall** | Opponent plays a land trigger | `opp_landfall` | 2 |
| **CounterThreshold** | Counter count reaches exact threshold | `counter_threshold` trigger | 2 |
| **CounterExactCount** (condition) | Static: "as long as exactly N counters" condition | condition raw text | 4 |
| **SelfUntappedCondition** | Static: source is untapped grants ability | condition raw text | 3 |
| **RevealedCardTypeBranch** | Revealed/top card type branch (creature vs. non-creature) | condition raw text | 2 |
| **RevealedLandTypeCheck** | Random discard reveal checks if card is a land | condition raw text | 2 |

## Notes on pre-Modern mechanical diversity

As expected, the pre-2003 era has significantly more **trigger-event diversity** than
later eras. Key observations:

1. **Phase variants dominate flagged space.** Of 260 flagged nodes, 99 (38%) are
   phase-trigger variants (`phase|player`, `phase|next_turn`, `phase|opponent`,
   `phase|next_cleanup`, etc.). The current scaffold collapses phase triggers to
   `upkeep` / `end_step` / `begin_combat` but early Magic used phase triggers as a
   general delayed-effect mechanism.

2. **High bucketing rate despite diversity.** 79.6% of nodes bucket cleanly — the
   existing 19 trigger slugs plus threshold/for_each/domain/paid_optional_cost cover
   the backbone of pre-Modern card design. Core triggers (upkeep, ETB, attacks,
   combat damage, dies) dominated even in 1993-2002.

3. **`etb_as` is a trigger event in pre-2003 (32 hits) but a condition scaffold in
   modern cards.** Two different AST positions for the same game concept — ETB modal
   choice. The trigger-event form needs a dedicated registry entry; the condition form
   is already scaffolded by `condScaffoldETBAs`.

4. **Condition flagging is nearly zero.** Only 20 out of 260 flagged nodes are
   condition-type (7.7%). Pre-2003 cards used structured condition kinds (threshold,
   for_each, domain) extensively and rarely emitted raw "as long as" clauses that
   escape detection. The 20 flagged conditions are all bespoke one-off mechanics
   (Chaos Orb, Static Orb, Homarid tide counters, Camel/Deserts).

5. **No new ability words.** Threshold and Domain (Invasion block) are the only
   structured ability words in this era, both already scaffolded. Flashback, Madness,
   Threshold all route through existing scaffolds. Phasing, Banding, Rampage, Trample,
   Cumulative Upkeep, Echo — all handled as static or have no triggerable condition.
