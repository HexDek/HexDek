# Thor 2.1 Corpus Audit — Era 2 (Modern 2003-2014)

_Generated 2026-05-08. Audit-only; no scaffold code added._

## Scope

- Era window: 2003-01-01 → 2014-12-31 (8th Edition through Khans of Tarkir block).
- Era inclusion proxy: a card's **latest printing** in `oracle-cards.json`
  has `released_at` inside the window. This UNDER-counts cards reprinted
  into 2015+ products (e.g. staple removal/burn reprinted in core sets) —
  those originally printed in Era 2 but with a post-2014 latest print are
  surfaced when the **next** era audit runs. Without an `all-printings` bulk
  the original-print date is not directly available; this audit treats the
  latest-printing window as a defensible proxy.
- Scaffold reference: 36 `conditionScaffoldKind` constants in
  `cmd/hexdek-thor/conditional_setup.go`.
- Targets: AST nodes whose condition `kind` is `intervening_if`,
  `as_long_as`, `conditional`, or `raw`. Structured trigger events
  (`creature_dies`, `etb`, etc.) are NOT scaffold targets — those are
  dispatched via `gameengine` triggers, not condition text matching.

## Summary

- Era 2 candidate names (latest-printing in window): **6925**
- Era 2 cards present in AST corpus: **6551**
- Cards with at least one condition-text node: **2115**
- Total condition-text nodes scanned: **2485**
  - Skipped — structured trigger raw text (handled by gameengine triggers, not scaffold layer): **2185**
  - Skipped — `Conditional` branch with structured kind (`type_match`, etc.) that the Go matcher doesn't route through `detectConditionScaffold`: **173**
- Matcher-eligible nodes (kind ∈ `intervening_if|as_long_as|conditional|raw`): **127**
- Bucketed (matched a known scaffold): **52**
- Flagged (no matching scaffold): **75**
- Coverage of matcher-eligible nodes: **40.9%**

## Bucketed breakdown

| Scaffold kind | Hits | Description |
|---|---:|---|
| `EnchantedCreature` | 20 | Aura conditions |
| `YouControlSubtype` | 14 | Tribal lord triggers |
| `CardInGraveyard` | 13 | Generic graveyard precondition |
| `CreatureCardsInGraveyard` | 3 | Threshold-style graveyard count |
| `LifeBelowThreshold` | 1 | Low-life ability |
| `CombatDamageDealt` | 1 | Combat damage condition |

### Sample bucketed cards (first 10 per kind)

#### `EnchantedCreature` (20 hits)

- **Fists of the Demigod** (static_condition) — `as long as enchanted creature is black, it gets +1/+1 and has wither`
- **Captivating Glance** (static_condition) — `if you win, gain control of enchanted creature. otherwise, that player gains control of enchanted creature`
- **Edge of the Divinity** (static_condition) — `as long as enchanted creature is white, it gets +1/+2`
- **Gift of the Deity** (static_condition) — `as long as enchanted creature is black, it gets +1/+1 and has deathtouch`
- **Scourge of the Nobilis** (static_condition) — `as long as enchanted creature is white, it gets +1/+1 and has lifelink`
- **Favor of the Overbeing** (static_condition) — `as long as enchanted creature is green, it gets +1/+1 and has vigilance`
- **Helm of the Ghastlord** (static_condition) — `as long as enchanted creature is blue, it gets +1/+1 and has "whenever this creature deals damage to an opponent, draw a`
- **Runes of the Deus** (static_condition) — `as long as enchanted creature is red, it gets +1/+1 and has double strike`
- **Clout of the Dominus** (static_condition) — `as long as enchanted creature is blue, it gets +1/+1 and has shroud`
- **Steel of the Godhead** (static_condition) — `as long as enchanted creature is white, it gets +1/+1 and has lifelink`

#### `YouControlSubtype` (14 hits)

- **Summit Apes** (static_condition) — `as long as you control a mountain, this creature has menace`
- **Grixis Grimblade** (static_condition) — `as long as you control another multicolored permanent, this creature gets +1/+1 and has deathtouch`
- **Esper Stormblade** (static_condition) — `as long as you control another multicolored permanent, this creature gets +1/+1 and has flying`
- **Skirk Outrider** (static_condition) — `as long as you control a beast, this creature gets +2/+2 and has trample`
- **Kithkin Greatheart** (static_condition) — `as long as you control a giant, this creature gets +1/+1 and has first strike`
- **Bant Sureblade** (static_condition) — `as long as you control another multicolored permanent, this creature gets +1/+1 and has first strike`
- **Konda's Hatamoto** (static_condition) — `as long as you control a legendary samurai, this creature gets +1/+2 and has vigilance`
- **Sejiri Merfolk** (static_condition) — `as long as you control a plains, this creature has first strike and lifelink`
- **Griffin Rider** (static_condition) — `as long as you control a griffin creature, this creature gets +3/+3 and has flying`
- **Naya Hushblade** (static_condition) — `as long as you control another multicolored permanent, this creature gets +1/+1 and has shroud`

#### `CardInGraveyard` (13 hits)

- **Increasing Ambition** (static_condition) — `if this spell was cast from a graveyard, instead search your library for two cards and put those cards into your hand. t`
- **Vexing Arcanix** (static_condition) — `if that card has the chosen name, that player puts it into their hand. otherwise, they put it into their graveyard and t`
- **Soldevi Excavations** (static_condition) — `if this land would enter, sacrifice an untapped island instead. if you do, put this land onto the battlefield. if you do`
- **Heart of Yavimaya** (static_condition) — `if this land would enter, sacrifice a forest instead. if you do, put this land onto the battlefield. if you don't, put i`
- **Balduvian Trading Post** (static_condition) — `if this land would enter, sacrifice an untapped mountain instead. if you do, put this land onto the battlefield. if you `
- **Kjeldoran Outpost** (static_condition) — `if this land would enter, sacrifice a plains instead. if you do, put this land onto the battlefield. if you don't, put i`
- **Neurok Familiar** (static_condition) — `if it's an artifact card, put it into your hand. otherwise, put it into your graveyard`
- **Candles of Leng** (static_condition) — `if it has the same name as a card in your graveyard, put it into your graveyard. otherwise, draw a card`
- **Lake of the Dead** (static_condition) — `if this land would enter, sacrifice a swamp instead. if you do, put this land onto the battlefield. if you don't, put it`
- **Zur's Weirding** (static_condition) — `if a player does, put that card into its owner's graveyard. otherwise, that player draws a card`

#### `CreatureCardsInGraveyard` (3 hits)

- **Enduring Renewal** (static_condition) — `if it's a creature card, put it into your graveyard. otherwise, draw a card`
- **Call of the Wild** (static_condition) — `if it's a creature card, put it onto the battlefield. otherwise, put it into your graveyard`
- **Impromptu Raid** (static_condition) — `if it isn't a creature card, put it into your graveyard. otherwise, put that card onto the battlefield`

#### `LifeBelowThreshold` (1 hits)

- **Phyrexian Unlife** (static_condition) — `as long as you have 0 or less life, all damage is dealt to you as though its source had infect`

#### `CombatDamageDealt` (1 hits)

- **Weathered Bodyguards** (static_condition) — `as long as this creature is untapped, all combat damage that would be dealt to you by unblocked creatures is dealt to th`

## Flagged clusters

_75 unbucketed condition-text nodes across 74 unique cards._

### Cluster: `could_not_classify` (15 nodes, 15 cards)

**Top raw texts:**

- (2×) `if it's a land card, the player puts it onto the battlefield. otherwise, the player casts it without paying its mana cost if able`
- (1×) `if it's a permanent card, you may put it onto the battlefield. if you do, repeat this process`
- (1×) `if you control more creatures than each other player, put two of those cards into your hand. otherwise, put one of them into your hand. then put the rest on the`
- (1×) `if you searched for a creature card that doesn't have that name, you may put it onto the battlefield under your control. then that player shuffles`
- (1×) `if you sacrifice a snow forest this way, this creature gains trample until end of turn. if you don't sacrifice a forest, sacrifice this creature and it deals 7 `
- (1×) `as long as a card exiled with this creature has flying, this creature has flying`
- (1×) `if you would begin your turn while this artifact is tapped, you may skip that turn instead. if you do, untap this artifact`
- (1×) `if you win, target player discards two cards. otherwise, that player discards a card`

**Sample cards:**

- **Primal Surge** (static_condition) — `if it's a permanent card, you may put it onto the battlefield. if you do, repeat this process`
- **Advice from the Fae** (static_condition) — `if you control more creatures than each other player, put two of those cards into your hand. otherwise, put one of them into your hand. then`
- **Sphinx Ambassador** (static_condition) — `if you searched for a creature card that doesn't have that name, you may put it onto the battlefield under your control. then that player sh`
- **Gargantuan Gorilla** (static_condition) — `if you sacrifice a snow forest this way, this creature gains trample until end of turn. if you don't sacrifice a forest, sacrifice this crea`
- **Death-Mask Duplicant** (static_condition) — `as long as a card exiled with this creature has flying, this creature has flying`
- **Time Vault** (static_condition) — `if you would begin your turn while this artifact is tapped, you may skip that turn instead. if you do, untap this artifact`
- **Pulling Teeth** (static_condition) — `if you win, target player discards two cards. otherwise, that player discards a card`
- **Wild Evocation** (static_condition) — `if it's a land card, the player puts it onto the battlefield. otherwise, the player casts it without paying its mana cost if able`
- **Lost in the Woods** (static_condition) — `if it's a forest card, remove that creature from combat. then put the revealed card on the bottom of your library`
- **Bronze Horse** (static_condition) — `as long as you control another creature, prevent all damage that would be dealt to this creature by spells that target it`
- **Scrapyard Mongrel** (static_condition) — `as long as you control an artifact, this creature gets +2/+0 and has trample`
- **Kaervek's Torch** (static_condition) — `as long as ~ is on the stack, spells that target it cost {2} more to cast`

### Cluster: `shares_creature_type` (12 nodes, 12 cards)

**Top raw texts:**

- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, this creature deals 2 damage to each creature`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, put a +1/+1 counter on this creature`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, each player discards their hand, then draws four cards`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, each creature you control gains flying until end of turn`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, you may play that card without paying its mana cost`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, each opponent mills three cards`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, this creature gets +1/+1 until end of turn`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, each opponent discards a card`

**Sample cards:**

- **Pyroclast Consul** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, this creature deals 2 damage to each creature`
- **Winnower Patrol** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, put a +1/+1 counter on this creature`
- **Sensation Gorger** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, each player discards their hand, then draws four cards`
- **Waterspout Weavers** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, each creature you control gains flying until end of turn`
- **Leaf-Crowned Elder** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, you may play that card without paying its mana cost`
- **Ink Dissolver** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, each opponent mills three cards`
- **Mudbutton Clanger** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, this creature gets +1/+1 until end of turn`
- **Squeaking Pie Grubfellows** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, each opponent discards a card`
- **Wolf-Skull Shaman** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, create a 2/2 green wolf creature token`
- **Wandering Graybeard** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, you gain 4 life`
- **Kithkin Zephyrnaut** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, this creature gets +2/+2 and gains flying and vigilance until`
- **Nightshade Schemers** (static_condition) — `if it shares a creature type with this creature, you may reveal it. if you do, each opponent loses 2 life`

### Cluster: `paired_soulbond` (10 nodes, 10 cards)

**Top raw texts:**

- (1×) `as long as this creature is paired with another creature, both creatures have lifelink`
- (1×) `as long as this creature is paired with another creature, each of those creatures gets +1/+1`
- (1×) `as long as this creature is paired with another creature, each of those creatures gets +2/+2`
- (1×) `as long as this creature is paired with another creature, both creatures have reach`
- (1×) `as long as this creature is paired with another creature, both creatures have hexproof`
- (1×) `as long as this creature is paired with another creature, both creatures have trample`
- (1×) `as long as this creature is paired with another creature, both creatures have deathtouch`
- (1×) `as long as this creature is paired with another creature, each of those creatures gets +4/+4`

**Sample cards:**

- **Nearheath Pilgrim** (static_condition) — `as long as this creature is paired with another creature, both creatures have lifelink`
- **Trusted Forcemage** (static_condition) — `as long as this creature is paired with another creature, each of those creatures gets +1/+1`
- **Druid's Familiar** (static_condition) — `as long as this creature is paired with another creature, each of those creatures gets +2/+2`
- **Geist Trappers** (static_condition) — `as long as this creature is paired with another creature, both creatures have reach`
- **Elgaud Shieldmate** (static_condition) — `as long as this creature is paired with another creature, both creatures have hexproof`
- **Pathbreaker Wurm** (static_condition) — `as long as this creature is paired with another creature, both creatures have trample`
- **Nightshade Peddler** (static_condition) — `as long as this creature is paired with another creature, both creatures have deathtouch`
- **Wolfir Silverheart** (static_condition) — `as long as this creature is paired with another creature, each of those creatures gets +4/+4`
- **Silverblade Paladin** (static_condition) — `as long as this creature is paired with another creature, both creatures have double strike`
- **Diregraf Escort** (static_condition) — `as long as this creature is paired with another creature, both creatures have protection from zombies`

### Cluster: `equipped_creature_state` (5 nodes, 5 cards)

**Top raw texts:**

- (2×) `as long as this creature is equipped, it gets +1/+1 and has flying`
- (1×) `as long as this creature is equipped, it gets +1/+1 and has vigilance`
- (1×) `as long as this creature is equipped, it gets +1/+1 and has first strike`
- (1×) `as long as this creature is equipped, each creature you control that's a soldier or a knight gets +1/+1`

**Sample cards:**

- **Skyhunter Cub** (static_condition) — `as long as this creature is equipped, it gets +1/+1 and has flying`
- **Leonin Den-Guard** (static_condition) — `as long as this creature is equipped, it gets +1/+1 and has vigilance`
- **Auriok Glaivemaster** (static_condition) — `as long as this creature is equipped, it gets +1/+1 and has first strike`
- **Kitesail Apprentice** (static_condition) — `as long as this creature is equipped, it gets +1/+1 and has flying`
- **Auriok Steelshaper** (static_condition) — `as long as this creature is equipped, each creature you control that's a soldier or a knight gets +1/+1`

### Cluster: `top_of_library` (4 nodes, 4 cards)

**Top raw texts:**

- (1×) `as long as the top card of your library is black, this creature and other vampire creatures you control get +2/+1 and have flying`
- (1×) `as long as the top card of your library is a creature card, creatures you control that share a color with that card get +1/+1`
- (1×) `as long as the top card of your library is an artifact or creature card, this creature has all activated abilities of that card`
- (1×) `as long as the top card of your library is a creature card, this creature gets +3/+3`

**Sample cards:**

- **Vampire Nocturnus** (static_condition) — `as long as the top card of your library is black, this creature and other vampire creatures you control get +2/+1 and have flying`
- **Crown of Convergence** (static_condition) — `as long as the top card of your library is a creature card, creatures you control that share a color with that card get +1/+1`
- **Skill Borrower** (static_condition) — `as long as the top card of your library is an artifact or creature card, this creature has all activated abilities of that card`
- **Mul Daya Channelers** (static_condition) — `as long as the top card of your library is a creature card, this creature gets +3/+3`

### Cluster: `permanent_state_self` (4 nodes, 4 cards)

**Top raw texts:**

- (1×) `as long as this creature is untapped, noncreature artifacts you control can't be enchanted, they have indestructible, and other players can't gain control of th`
- (1×) `as long as this creature is enchanted, it can attack as though it didn't have defender`
- (1×) `as long as this creature is untapped, all damage that would be dealt to you by artifacts is dealt to this creature instead`
- (1×) `as long as this creature is untapped, green creatures you control get +1/+1`

**Sample cards:**

- **Guardian Beast** (static_condition) — `as long as this creature is untapped, noncreature artifacts you control can't be enchanted, they have indestructible, and other players can'`
- **Pillar of War** (static_condition) — `as long as this creature is enchanted, it can attack as though it didn't have defender`
- **Martyrs of Korlis** (static_condition) — `as long as this creature is untapped, all damage that would be dealt to you by artifacts is dealt to this creature instead`
- **Juniper Order Advocate** (static_condition) — `as long as this creature is untapped, green creatures you control get +1/+1`

### Cluster: `equipped_creature_attachee` (3 nodes, 3 cards)

**Top raw texts:**

- (1×) `as long as equipped creature is a human or an angel, it has vigilance`
- (1×) `as long as equipped creature is a human, it gets an additional +1/+0`
- (1×) `as long as equipped creature is a human, it gets an additional +1/+1`

**Sample cards:**

- **Bladed Bracers** (static_condition) — `as long as equipped creature is a human or an angel, it has vigilance`
- **Silver-Inlaid Dagger** (static_condition) — `as long as equipped creature is a human, it gets an additional +1/+0`
- **Heavy Mattock** (static_condition) — `as long as equipped creature is a human, it gets an additional +1/+1`

### Cluster: `non_creature_aura` (3 nodes, 3 cards)

**Top raw texts:**

- (1×) `as long as enchanted artifact isn't a creature, it's an artifact creature with power and toughness each equal to its mana value`
- (1×) `as long as enchanted land is a basic mountain, goblin creatures get +0/+2`
- (1×) `as long as enchanted land is a basic mountain, goblin creatures get +1/+0`

**Sample cards:**

- **Animate Artifact** (static_condition) — `as long as enchanted artifact isn't a creature, it's an artifact creature with power and toughness each equal to its mana value`
- **Goblin Caves** (static_condition) — `as long as enchanted land is a basic mountain, goblin creatures get +0/+2`
- **Goblin Shrine** (static_condition) — `as long as enchanted land is a basic mountain, goblin creatures get +1/+0`

### Cluster: `quest_counters` (3 nodes, 3 cards)

**Top raw texts:**

- (1×) `as long as this enchantment has five or more quest counters on it, creatures you control get +2/+0`
- (1×) `as long as this enchantment has six or more quest counters on it, if you would draw a card, you may instead search your library for a card, put that card into y`
- (1×) `as long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step`

**Sample cards:**

- **Quest for the Goblin Lord** (static_condition) — `as long as this enchantment has five or more quest counters on it, creatures you control get +2/+0`
- **Archmage Ascension** (static_condition) — `as long as this enchantment has six or more quest counters on it, if you would draw a card, you may instead search your library for a card, `
- **Quest for Renewal** (static_condition) — `as long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step`

### Cluster: `hand_size_threshold` (3 nodes, 3 cards)

**Top raw texts:**

- (1×) `as long as you have four or more cards in hand, ~ has vigilance`
- (1×) `as long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike`
- (1×) `as long as you have seven or more cards in hand, this creature gets +2/+1 and has fear`

**Sample cards:**

- **Kiyomaro, First to Stand** (static_condition) — `as long as you have four or more cards in hand, ~ has vigilance`
- **Akki Underling** (static_condition) — `as long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike`
- **Deathmask Nezumi** (static_condition) — `as long as you have seven or more cards in hand, this creature gets +2/+1 and has fear`

### Cluster: `hand_size_compare` (3 nodes, 3 cards)

**Top raw texts:**

- (1×) `as long as you have more cards in hand than each opponent, this creature gets +2/+2 and has flying`
- (1×) `as long as you have more cards in hand than each opponent, this creature gets +1/+2 and has "whenever this creature deals combat damage, you gain 3 life."`
- (1×) `as long as you have more cards in hand than each opponent, this creature gets +3/+3`

**Sample cards:**

- **Secretkeeper** (static_condition) — `as long as you have more cards in hand than each opponent, this creature gets +2/+2 and has flying`
- **Descendant of Kiyomaro** (static_condition) — `as long as you have more cards in hand than each opponent, this creature gets +1/+2 and has "whenever this creature deals combat damage, you`
- **Okina Nightwatch** (static_condition) — `as long as you have more cards in hand than each opponent, this creature gets +3/+3`

### Cluster: `count_for_each` (2 nodes, 1 cards)

**Top raw texts:**

- (1×) `as long as ~ isn't attacking, its power and toughness are each equal to the number of forests you control`
- (1×) `as long as ~ is attacking, its power and toughness are each equal to the number of forests defending player controls`

**Sample cards:**

- **Gaea's Liege** (static_condition) — `as long as ~ isn't attacking, its power and toughness are each equal to the number of forests you control`

### Cluster: `monstrous` (2 nodes, 2 cards)

**Top raw texts:**

- (1×) `as long as this creature is monstrous, it has trample and can attack as though it didn't have defender`
- (1×) `as long as this creature is monstrous, it has reach`

**Sample cards:**

- **Colossus of Akros** (static_condition) — `as long as this creature is monstrous, it has trample and can attack as though it didn't have defender`
- **Swarmborn Giant** (static_condition) — `as long as this creature is monstrous, it has reach`

### Cluster: `opponent_state` (1 nodes, 1 cards)

**Top raw texts:**

- (1×) `as long as an opponent has 10 or less life, this creature gets +2/+1 and has intimidate`

**Sample cards:**

- **Guul Draz Vampire** (static_condition) — `as long as an opponent has 10 or less life, this creature gets +2/+1 and has intimidate`

### Cluster: `name_match` (1 nodes, 1 cards)

**Top raw texts:**

- (1×) `if that library contains exactly the chosen number of cards with the chosen name, ~ deals 8 damage to that player. then that player shuffles`

**Sample cards:**

- **Mindblaze** (static_condition) — `if that library contains exactly the chosen number of cards with the chosen name, ~ deals 8 damage to that player. then that player shuffles`

### Cluster: `exact_count_self` (1 nodes, 1 cards)

**Top raw texts:**

- (1×) `as long as you control exactly one creature, that creature gets +3/+1 and has lifelink`

**Sample cards:**

- **Homicidal Seclusion** (static_condition) — `as long as you control exactly one creature, that creature gets +3/+1 and has lifelink`

### Cluster: `target_pt_threshold` (1 nodes, 1 cards)

**Top raw texts:**

- (1×) `if target creature has toughness 5 or greater, it gets +4/-4 until end of turn. otherwise, it gets +4/-x until end of turn, where x is its toughness minus 1`

**Sample cards:**

- **Blood Lust** (static_condition) — `if target creature has toughness 5 or greater, it gets +4/-4 until end of turn. otherwise, it gets +4/-x until end of turn, where x is its t`

### Cluster: `phase_step_other` (1 nodes, 1 cards)

**Top raw texts:**

- (1×) `if you would draw a card during your draw step, instead you may skip that draw. if you do, until your next turn, you can't be attacked except by creatures with `

**Sample cards:**

- **Island Sanctuary** (static_condition) — `if you would draw a card during your draw step, instead you may skip that draw. if you do, until your next turn, you can't be attacked excep`

### Cluster: `blocks_or_blocked` (1 nodes, 1 cards)

**Top raw texts:**

- (1×) `as long as this creature is untapped, all damage that would be dealt to you by unblocked creatures is dealt to this creature instead`

**Sample cards:**

- **Veteran Bodyguard** (static_condition) — `as long as this creature is untapped, all damage that would be dealt to you by unblocked creatures is dealt to this creature instead`

### Top 30 unbucketed raw texts (cross-cluster)

| Count | Raw text |
|---:|---|
| 2 | `as long as this creature is equipped, it gets +1/+1 and has flying` |
| 2 | `if it's a land card, the player puts it onto the battlefield. otherwise, the player casts it without paying its mana cost if able` |
| 1 | `as long as an opponent has 10 or less life, this creature gets +2/+1 and has intimidate` |
| 1 | `if it shares a creature type with this creature, you may reveal it. if you do, this creature deals 2 damage to each creature` |
| 1 | `as long as the top card of your library is black, this creature and other vampire creatures you control get +2/+1 and have flying` |
| 1 | `as long as equipped creature is a human or an angel, it has vigilance` |
| 1 | `as long as enchanted artifact isn't a creature, it's an artifact creature with power and toughness each equal to its mana value` |
| 1 | `as long as this enchantment has five or more quest counters on it, creatures you control get +2/+0` |
| 1 | `as long as the top card of your library is a creature card, creatures you control that share a color with that card get +1/+1` |
| 1 | `if that library contains exactly the chosen number of cards with the chosen name, ~ deals 8 damage to that player. then that player shuffles` |
| 1 | `as long as you have four or more cards in hand, ~ has vigilance` |
| 1 | `as long as this creature is untapped, noncreature artifacts you control can't be enchanted, they have indestructible, and other players can't gain control of th` |
| 1 | `as long as enchanted land is a basic mountain, goblin creatures get +0/+2` |
| 1 | `as long as you control exactly one creature, that creature gets +3/+1 and has lifelink` |
| 1 | `if it shares a creature type with this creature, you may reveal it. if you do, put a +1/+1 counter on this creature` |
| 1 | `if it shares a creature type with this creature, you may reveal it. if you do, each player discards their hand, then draws four cards` |
| 1 | `if it's a permanent card, you may put it onto the battlefield. if you do, repeat this process` |
| 1 | `if you control more creatures than each other player, put two of those cards into your hand. otherwise, put one of them into your hand. then put the rest on the` |
| 1 | `as long as this creature is enchanted, it can attack as though it didn't have defender` |
| 1 | `as long as this creature is paired with another creature, both creatures have lifelink` |
| 1 | `if target creature has toughness 5 or greater, it gets +4/-4 until end of turn. otherwise, it gets +4/-x until end of turn, where x is its toughness minus 1` |
| 1 | `as long as this enchantment has six or more quest counters on it, if you would draw a card, you may instead search your library for a card, put that card into y` |
| 1 | `as long as enchanted land is a basic mountain, goblin creatures get +1/+0` |
| 1 | `as long as this creature is equipped, it gets +1/+1 and has vigilance` |
| 1 | `as long as this creature is paired with another creature, each of those creatures gets +1/+1` |
| 1 | `if you searched for a creature card that doesn't have that name, you may put it onto the battlefield under your control. then that player shuffles` |
| 1 | `as long as this creature is paired with another creature, each of those creatures gets +2/+2` |
| 1 | `as long as you have more cards in hand than each opponent, this creature gets +2/+2 and has flying` |
| 1 | `as long as this creature is paired with another creature, both creatures have reach` |
| 1 | `as long as this creature is equipped, it gets +1/+1 and has first strike` |

## Proposed new scaffold kinds

### `condScaffoldSharedType` (covers cluster `shares_creature_type`, 12 nodes / 12 cards)

'if it shares a creature type with this creature' (Hivemaster-style reveal). Plant a card on top of the relevant zone whose subtype overlaps with the source's subtype set.

**Representative raw texts:**

- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, this creature deals 2 damage to each creature`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, put a +1/+1 counter on this creature`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, each player discards their hand, then draws four cards`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, each creature you control gains flying until end of turn`
- (1×) `if it shares a creature type with this creature, you may reveal it. if you do, you may play that card without paying its mana cost`

### `condScaffoldEquippedSelf` (covers cluster `equipped_creature_state`, 5 nodes / 5 cards)

'as long as this creature is equipped' or 'as long as ~ is equipped'. Place a friendly Equipment with AttachedTo = source so the IsEquipped predicate succeeds.

**Representative raw texts:**

- (2×) `as long as this creature is equipped, it gets +1/+1 and has flying`
- (1×) `as long as this creature is equipped, it gets +1/+1 and has vigilance`
- (1×) `as long as this creature is equipped, it gets +1/+1 and has first strike`
- (1×) `as long as this creature is equipped, each creature you control that's a soldier or a knight gets +1/+1`

### `condScaffoldTopOfLibrary` (covers cluster `top_of_library`, 4 nodes / 4 cards)

'as long as the top card of your library is <predicate>'. Stack a card matching the predicate (color/type) onto the controller's library top.

**Representative raw texts:**

- (1×) `as long as the top card of your library is black, this creature and other vampire creatures you control get +2/+1 and have flying`
- (1×) `as long as the top card of your library is a creature card, creatures you control that share a color with that card get +1/+1`
- (1×) `as long as the top card of your library is an artifact or creature card, this creature has all activated abilities of that card`
- (1×) `as long as the top card of your library is a creature card, this creature gets +3/+3`

### `condScaffoldEquippedCreatureAttachee` (covers cluster `equipped_creature_attachee`, 3 nodes / 3 cards)

'as long as equipped creature is <X>' — sibling to EquippedSelf but the predicate is on the attached creature, not the source equipment. Place a creature matching the predicate and attach the source Equipment to it.

**Representative raw texts:**

- (1×) `as long as equipped creature is a human or an angel, it has vigilance`
- (1×) `as long as equipped creature is a human, it gets an additional +1/+0`
- (1×) `as long as equipped creature is a human, it gets an additional +1/+1`

### `condScaffoldEnchantedNonCreature` (covers cluster `non_creature_aura`, 3 nodes / 3 cards)

'as long as enchanted artifact/land/permanent ...' — extend EnchantedCreature to cover non-creature aura attachments (Animate Artifact, Goblin Caves, Spreading Seas).

**Representative raw texts:**

- (1×) `as long as enchanted artifact isn't a creature, it's an artifact creature with power and toughness each equal to its mana value`
- (1×) `as long as enchanted land is a basic mountain, goblin creatures get +0/+2`
- (1×) `as long as enchanted land is a basic mountain, goblin creatures get +1/+0`

### `condScaffoldQuestCounters` (covers cluster `quest_counters`, 3 nodes / 3 cards)

'as long as this enchantment has N or more quest counters on it'. Add N quest counters to the source permanent before resolution.

**Representative raw texts:**

- (1×) `as long as this enchantment has five or more quest counters on it, creatures you control get +2/+0`
- (1×) `as long as this enchantment has six or more quest counters on it, if you would draw a card, you may instead search your library for a card, put that card into y`
- (1×) `as long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step`

### `condScaffoldHandSizeThreshold` (covers cluster `hand_size_threshold`, 3 nodes / 3 cards)

'as long as you have N or more cards in hand'. Top up the controller's hand to N.

**Representative raw texts:**

- (1×) `as long as you have four or more cards in hand, ~ has vigilance`
- (1×) `as long as you have seven or more cards in hand, this creature gets +2/+1 and has first strike`
- (1×) `as long as you have seven or more cards in hand, this creature gets +2/+1 and has fear`

### `condScaffoldSelfPermanentState` (covers cluster `permanent_state_self`, 4 nodes / 4 cards)

Self-state predicates the existing matchers don't cover: 'is untapped', 'is enchanted', 'this artifact has' — place the source in the required runtime state.

**Representative raw texts:**

- (1×) `as long as this creature is untapped, noncreature artifacts you control can't be enchanted, they have indestructible, and other players can't gain control of th`
- (1×) `as long as this creature is enchanted, it can attack as though it didn't have defender`
- (1×) `as long as this creature is untapped, all damage that would be dealt to you by artifacts is dealt to this creature instead`
- (1×) `as long as this creature is untapped, green creatures you control get +1/+1`

### `condScaffoldPairedSoulbond` (covers cluster `paired_soulbond`, 10 nodes / 10 cards)

Soulbond preconditions ('as long as this creature is paired with another creature'). Scaffolding should put a friendly creature on the battlefield and mark the source as Paired = true with that partner.

**Representative raw texts:**

- (1×) `as long as this creature is paired with another creature, both creatures have lifelink`
- (1×) `as long as this creature is paired with another creature, each of those creatures gets +1/+1`
- (1×) `as long as this creature is paired with another creature, each of those creatures gets +2/+2`
- (1×) `as long as this creature is paired with another creature, both creatures have reach`
- (1×) `as long as this creature is paired with another creature, both creatures have hexproof`

### `condScaffoldHandSizeCompare` (covers cluster `hand_size_compare`, 3 nodes / 3 cards)

'as long as you have more cards in hand than each opponent' (Kiyomaro / Descendant of Kiyomaro). Top up controller's hand to exceed opponents' hand sizes.

**Representative raw texts:**

- (1×) `as long as you have more cards in hand than each opponent, this creature gets +2/+2 and has flying`
- (1×) `as long as you have more cards in hand than each opponent, this creature gets +1/+2 and has "whenever this creature deals combat damage, you gain 3 life."`
- (1×) `as long as you have more cards in hand than each opponent, this creature gets +3/+3`

### `condScaffoldMonstrous` (covers cluster `monstrous`, 2 nodes / 2 cards)

Theros 'as long as this creature is monstrous' / 'monstrosity' preconditions. Mark the source permanent's `Monstrous = true` flag (or its TurnCounters equivalent) before resolution. Currently absent from the scaffold catalogue.

**Representative raw texts:**

- (1×) `as long as this creature is monstrous, it has trample and can attack as though it didn't have defender`
- (1×) `as long as this creature is monstrous, it has reach`
