# Thor 2.1 Corpus Audit — Era 1 (Post-Modern, 2015-01-01 to 2024-12-31)

Audit-only scan against `cmd/hexdek-thor/conditional_setup.go` —
for every Era 1 card, every triggered ability is matched against the
18 trigger slugs in `triggerConditionActions`, and every condition node
(intervening_if / static.condition / nested Conditional) is matched
against the 35 `conditionScaffoldKind` rules in `detectConditionScaffold`.
Static vanilla abilities (keywords, plain anthems) are excluded.

## Summary

- Era 1 cards (released 2015-01-01..2024-12-31, present in AST corpus): **17910**
- Cards with at least one conditional/triggered ability: **8105**
- Cards with at least one ability bucketed by an existing scaffold: **5852**
- Cards with at least one *flagged* (unmatched) ability: **3049**
- Cards fully bucketed (no flagged abilities): **5056**

- Total ability/condition nodes scanned: **10208**
- Bucketed nodes: **6927** (67.9%)
- Flagged nodes: **3281** (32.1%)

## Top scaffold kinds by frequency (top 15)

| Rank | Scaffold | Hits | Examples |
|---:|---|---:|---|
| 1 | `TRIG:creature_etb` | 2930 | A Girl and Her Dogs, A Killer Among Us, A-Acererak the Archlich, A-Baleful Beholder, A-Cauldron Familiar |
| 2 | `TRIG:attacks` | 1098 | A Girl and Her Dogs, A-Acererak the Archlich, A-Akki Ronin, A-Ancestral Katana, A-Ardent Dustspeaker |
| 3 | `TRIG:cast_spell` | 690 | A-Devoted Grafkeeper // A-Departed Soulkeeper, A-Dragon's Rage Channeler, A-Leyline of Resonance, A-Master of Winds, A-Mentor's Guidance |
| 4 | `COND:STRUCTURED:paid_optional_cost` | 365 | A-Akki Ronin, A-Ardent Dustspeaker, A-Civil Servant, A-Elderfang Ritualist, A-Elven Bow |
| 5 | `TRIG:upkeep` | 344 | A Girl and Her Dogs, A-Minsc & Boo, Timeless Heroes, A-The One Ring, A-Uurg, Spawn of Turg, A-Visions of Phyrexia |
| 6 | `TRIG:end_step` | 306 | A-Alrund, God of the Cosmos // A-Hakka, Whispering Raven, A-Cosmos Elixir, A-Death-Priest of Myrkul, A-Geology Enthusiast, A-Iridescent Hornbeetle |
| 7 | `COND:STRUCTURED:for_each` | 158 | A-Ocelot Pride, A-Rowan, Scholar of Sparks // A-Will, Scholar of Frost, Acererak the Archlich, Aclazotz, Deepest Betrayal // Temple of the Dead, Aetherspouts |
| 8 | `TRIG:creature_dies` | 130 | A-Skemfar Avenger, A-Skyclave Shadowcat, Abattoir Ghoul, Ajani's Last Stand, Akki Ember-Keeper |
| 9 | `COND:STRUCTURED:did_prior_action` | 112 | A-Leyline of Resonance, Aether Refinery, Agrus Kos, Eternal Soldier, Akoum Flameseeker, Alania, Divergent Storm |
| 10 | `TRIG:sacrifice` | 81 | A-Forge Boss, Albiorix, Goose Tyrant // Wild Goose Chase, Ashad, the Lone Cyberman, Bag of Devouring, Blood Aspirant |
| 11 | `TRIG:ltb` | 78 | A-Circuit Mender, Admonition Angel, Alpine Guide, Angel of Serenity, Angelic Sleuth |
| 12 | `TRIG:gain_life` | 73 | A-Vampire Scrivener, Ageless Entity, Ajani's Pridemate, Amalia Benavides Aguirre, Angelic Accord |
| 13 | `COND:STRUCTURED:if` | 72 | A-Cosmos Elixir, A-Ocelot Pride, All-Star Kicker, Amped Raptor, Archfiend of the Dross |
| 14 | `TRIG:discard` | 59 | A-Mishra, Excavation Prodigy, A-Urza, Powerstone Prodigy, Aclazotz, Deepest Betrayal // Temple of the Dead, Ajani's Last Stand, All-Seeing Arbiter |
| 15 | `TRIG:draw_card` | 56 | A-Orcish Bowmasters, A-Queza, Augur of Agonies, A-Wizard Class, Burlfist Oak, Chasm Skulker |

## Bucketed breakdown (all scaffolds)

Each row is a scaffold key — `TRIG:` for trigger slugs (from
`classifyTrigger`), `COND:` for condition scaffold kinds (from
`detectConditionScaffold`). `STRUCTURED:` rows are AST condition kinds
the scaffold detector intentionally skips because they go through other engine paths.

| Scaffold | Hits | Examples |
|---|---:|---|
| `TRIG:creature_etb` | 2930 | A Girl and Her Dogs, A Killer Among Us, A-Acererak the Archlich, A-Baleful Beholder, A-Cauldron Familiar |
| `TRIG:attacks` | 1098 | A Girl and Her Dogs, A-Acererak the Archlich, A-Akki Ronin, A-Ancestral Katana, A-Ardent Dustspeaker |
| `TRIG:cast_spell` | 690 | A-Devoted Grafkeeper // A-Departed Soulkeeper, A-Dragon's Rage Channeler, A-Leyline of Resonance, A-Master of Winds, A-Mentor's Guidance |
| `COND:STRUCTURED:paid_optional_cost` | 365 | A-Akki Ronin, A-Ardent Dustspeaker, A-Civil Servant, A-Elderfang Ritualist, A-Elven Bow |
| `TRIG:upkeep` | 344 | A Girl and Her Dogs, A-Minsc & Boo, Timeless Heroes, A-The One Ring, A-Uurg, Spawn of Turg, A-Visions of Phyrexia |
| `TRIG:end_step` | 306 | A-Alrund, God of the Cosmos // A-Hakka, Whispering Raven, A-Cosmos Elixir, A-Death-Priest of Myrkul, A-Geology Enthusiast, A-Iridescent Hornbeetle |
| `COND:STRUCTURED:for_each` | 158 | A-Ocelot Pride, A-Rowan, Scholar of Sparks // A-Will, Scholar of Frost, Acererak the Archlich, Aclazotz, Deepest Betrayal // Temple of the Dead, Aetherspouts |
| `TRIG:creature_dies` | 130 | A-Skemfar Avenger, A-Skyclave Shadowcat, Abattoir Ghoul, Ajani's Last Stand, Akki Ember-Keeper |
| `COND:STRUCTURED:did_prior_action` | 112 | A-Leyline of Resonance, Aether Refinery, Agrus Kos, Eternal Soldier, Akoum Flameseeker, Alania, Divergent Storm |
| `TRIG:sacrifice` | 81 | A-Forge Boss, Albiorix, Goose Tyrant // Wild Goose Chase, Ashad, the Lone Cyberman, Bag of Devouring, Blood Aspirant |
| `TRIG:ltb` | 78 | A-Circuit Mender, Admonition Angel, Alpine Guide, Angel of Serenity, Angelic Sleuth |
| `TRIG:gain_life` | 73 | A-Vampire Scrivener, Ageless Entity, Ajani's Pridemate, Amalia Benavides Aguirre, Angelic Accord |
| `COND:STRUCTURED:if` | 72 | A-Cosmos Elixir, A-Ocelot Pride, All-Star Kicker, Amped Raptor, Archfiend of the Dross |
| `TRIG:discard` | 59 | A-Mishra, Excavation Prodigy, A-Urza, Powerstone Prodigy, Aclazotz, Deepest Betrayal // Temple of the Dead, Ajani's Last Stand, All-Seeing Arbiter |
| `TRIG:draw_card` | 56 | A-Orcish Bowmasters, A-Queza, Augur of Agonies, A-Wizard Class, Burlfist Oak, Chasm Skulker |
| `COND:STRUCTURED:etb_tapped_unless` | 49 | Abandoned Campground, Archway of Innovation, Arena of Glory, Argoth, Sanctum of Nature, Barad-dûr |
| `COND:STRUCTURED:delirium` | 40 | A-Dragon's Rage Channeler, Backwoods Survivalists, Bloodbraid Marauder, Convert to Slime, Deathcap Cultivator |
| `COND:CardInGraveyard` | 21 | A-Lier, Disciple of the Drowned, A-Syndicate Infiltrator, Aven Heartstabber, Deadly Cover-Up, Dearly Departed |
| `COND:STRUCTURED:threshold` | 21 | Billowing Shriekmass, Cabal Initiate, Centaur Chieftain, Cephalid Inkmage, Cephalid Sage |
| `COND:STRUCTURED:etb_if` | 17 | Ascendant Packleader, Blacksnag Buzzard, Cackling Slasher, Cindering Cutthroat, Epochrasite |
| `COND:STRUCTURED:spell_mastery` | 15 | Calculated Dismissal, Dark Dabbling, Dark Petition, Exquisite Firecraft, Fiery Impulse |
| `TRIG:lose_life` | 14 | A-Vampire Scrivener, Bloodthirsty Conqueror, Exquisite Blood, Gonti's Machinations, Marina Vendrell's Grimoire |
| `COND:EnchantedCreature` | 13 | Combat Research, Dog Umbra, Equestrian Skill, Face of Divinity, Hope Against Hope |
| `COND:STRUCTURED:raid` | 10 | Admiral's Order, Firecannon Blast, Goblin Boarders, Heartless Pillage, Protean Raider |
| `COND:STRUCTURED:two_plus_spells_cast_last_turn` | 9 | Breakneck Rider // Neck Breaker, Convicted Killer // Branded Howler, Gatstaf Arsonists // Gatstaf Ravagers, Hermit of the Natterknolls // Lone Wolf of the Natterknolls, Kessig Forgemaster // Flameheart Werewolf |
| `COND:STRUCTURED:no_spells_cast_last_turn` | 9 | Breakneck Rider // Neck Breaker, Convicted Killer // Branded Howler, Gatstaf Arsonists // Gatstaf Ravagers, Hermit of the Natterknolls // Lone Wolf of the Natterknolls, Kessig Forgemaster // Flameheart Werewolf |
| `COND:STRUCTURED:hellbent` | 8 | Cutthroat il-Dal, Demonfire, Gathan Raiders, Infernal Tutor, Necromancer's Familiar |
| `TRIG:draw_step` | 8 | Avaricious Dragon, Bandit's Talent, Lord Skitter's Blessing, Mana Vault, Midnight Oil |
| `COND:STRUCTURED:metalcraft` | 8 | Etched Champion, Galvanic Blast, Indomitable Archangel, Jor Kadeen, the Prevailer, Puresteel Paladin |
| `COND:STRUCTURED:it_was_a_creature` | 8 | Enduring Courage, Enduring Curiosity, Enduring Friendship, Enduring Innocence, Enduring Tenacity |
| `COND:STRUCTURED:you_control` | 6 | Boundary Lands Ranger, Council's Deliberation, Historian of Zhalfir, Scholar of Stars, Settlement Blacksmith |
| `COND:CreatureCardsInGraveyard` | 6 | Cairn Wanderer, Gray Slaad // Entropic Decay, Kellan, Daring Traveler // Journey On, Necrotic Ooze, Nylea, Keen-Eyed |
| `COND:STRUCTURED:repeat_n` | 6 | Carnage Interpreter, Jadelight Spelunker, Myojin of Cryptic Dreams, Progenitor Exarch, Storm King's Thunder |
| `COND:STRUCTURED:domain` | 5 | Gaea's Might, Herd Migration, Leyline Binding, Tromp the Domains, Yavimaya Sojourner |
| `COND:STRUCTURED:you_control_creature_power_ge` | 5 | Beastbond Outcaster, Colossal Majesty, Hunter's Talent, Ornery Dilophosaur, Stampede Rider |
| `COND:STRUCTURED:lieutenant` | 5 | Angelic Field Marshal, Skyhunter Strike Force, Stormsurge Kraken, Thunderfoot Baloth, Tyrant's Familiar |
| `COND:STRUCTURED:life_threshold` | 4 | Caduceus, Staff of Hermes, Phial of Galadriel, Serra Ascendant, Twinblade Paladin |
| `COND:STRUCTURED:life_vs_half_starting` | 4 | Anya, Merciless Angel, Bane, Lord of Darkness, Bhaal, Lord of Murder, Myrkul, Lord of Bones |
| `COND:STRUCTURED:ferocious` | 4 | Frontier Mastodon, Temur Battle Rage, Wild Slash, Winds of Qal Sisma |
| `COND:STRUCTURED:you_descended_this_turn` | 3 | Child of the Volcano, Deep Goblin Skulltaker, Stalactite Stalker |
| `COND:STRUCTURED:was_kicked` | 3 | Keldon Strike Team, Phyrexian Warhorse, Sergeant-at-Arms |
| `COND:STRUCTURED:fateful_hour` | 3 | Courageous Resolve, Spell Snuff, Thraben Doomsayer |
| `COND:OpponentMoreLands` | 3 | Scholar of New Horizons, Stoic Farmer, Verge Rangers |
| `COND:STRUCTURED:gained_life_this_turn` | 3 | A-Ocelot Pride, Crested Sunmare, Ocelot Pride |
| `COND:CastSpellThisTurn` | 3 | Fortune Teller's Talent, Piston-Fist Cyclops, The Wandering Emperor |
| `COND:DiscardedThisTurn` | 3 | Asmoranomardicadaistinaculdacar, Oko, the Ringleader, Thought-Stalker Warlock |
| `COND:STRUCTURED:self_has_counter` | 3 | Arixmethes, Slumbering Isle, Scuttlegator, Skyclave Sentinel |
| `COND:YouControlSubtype:token` | 2 | Briarbridge Tracker, Druid of the Spade |
| `COND:STRUCTURED:coven` | 2 | Might of the Old Ways, Ritual of Hope |
| `COND:STRUCTURED:you_attacked_this_turn` | 2 | Nightsquad Commando, Trynn, Champion of Freedom |
| `COND:STRUCTURED:life_delta_threshold` | 2 | Leyline of Hope, Righteous Valkyrie |
| `COND:Ferocious` | 2 | Bristlepack Sentry, Drowsing Tyrannodon |
| `COND:STRUCTURED:no_mana_spent_to_cast` | 2 | Boromir, Warden of the Tower, Vexing Bauble |
| `COND:STRUCTURED:creature_died_this_turn` | 2 | Bulette, Sabertooth Mauler |
| `COND:STRUCTURED:no_creatures_on_battlefield` | 2 | Porphyry Nodes, Pyrohemia |
| `COND:YouControlSubtype:red` | 2 | Battle Brawler, Ember Weaver |
| `COND:STRUCTURED:didnt_attack_this_turn` | 2 | Curious Obsession, See Red |
| `COND:STRUCTURED:morbid` | 2 | Brimstone Volley, Grim Reaper's Sprint |
| `COND:DrawnCardThisTurn` | 1 | Darkblade Agent |
| `COND:YouControlSubtype:goblin` | 1 | Snarling Warg |
| `COND:STRUCTURED:life_threshold_both` | 1 | Blood Baron of Vizkopa |
| `COND:YouControlSubtype:name` | 1 | Noble Banneret |
| `COND:YouControlSubtype:attack` | 1 | Backstreet Bruiser |
| `TRIG:opponent_discards` | 1 | Waste Not |
| `COND:LandfallThisTurn` | 1 | Shoreline Scout |
| `COND:YouControlSubtype:chandra` | 1 | Renegade Firebrand |
| `COND:YouControlSubtype:cleric` | 1 | Relic Vial |
| `COND:Monarch` | 1 | Dawnglade Regent |
| `COND:LifeLostThisTurn` | 1 | Essence Channeler |
| `COND:OpponentLostLife` | 1 | Vampire Socialite |
| `COND:STRUCTURED:dealt_damage_to_opponent_this_turn` | 1 | Dunerider Outlaw |
| `COND:YouControlSubtype:rowan` | 1 | Rowan's Battleguard |
| `COND:YouControlSubtype:demon` | 1 | Unholy Annex // Ritual Chamber |
| `COND:YouControlSubtype:dinosaur` | 1 | Thrash of Raptors |
| `COND:Delirium` | 1 | Keen-Eyed Curator |
| `COND:YouControlSubtype:army` | 1 | Grond, the Gatebreaker |
| `COND:YouControlSubtype:jace` | 1 | Jace's Sentinel |
| `COND:YouControlSubtype:transformed` | 1 | Oculus Whelp |
| `COND:YouControlSubtype:raccoon` | 1 | Take Out the Trash |
| `COND:SpellMastery` | 1 | Ghitu Lavarunner |
| `COND:YouControlSubtype:yanling` | 1 | Moon-Eating Dog |
| `COND:YouControlSubtype:nissa` | 1 | Guardian of the Great Conduit |
| `COND:CombatDamageDealt` | 1 | Grisly Sigil |
| `COND:YouControlSubtype:dovin` | 1 | Dovin's Automaton |
| `COND:STRUCTURED:landfall` | 1 | Groundswell |
| `COND:YouControlSubtype:tezzeret` | 1 | Tezzeret's Strider |
| `COND:YouControlSubtype:dragon` | 1 | Dragonloft Idol |
| `COND:YouControlSubtype:cost` | 1 | Descendants' Path |
| `COND:YouControlSubtype:green` | 1 | Hungering Yeti |
| `COND:YouControlSubtype:teferi` | 1 | Teferi's Sentinel |
| `COND:YouControlSubtype:domri` | 1 | Charging War Boar |
| `COND:YouControlSubtype:wizard` | 1 | Expedition Diviner |

## Flagged clusters (proposed new scaffold kinds)

Each cluster groups condition/trigger texts that no existing scaffold matched.
The cluster tag is a heuristic from `classifyFlaggedCondition` /
`classifyFlaggedTrigger`. Below each table row is a 3-card sample with the
raw condition text and the ability slot it came from.

### `trigger:die` — 454 hits

Examples:

- **A-Blood Artist** (ability[0] triggered) — `event="die" phase=""`
- **A-Elderfang Ritualist** (ability[0] triggered) — `event="die" phase=""`
- **A-Guildsworn Prowler** (ability[1] triggered) — `event="die" phase=""`
- **A-Haywire Mite** (ability[0] triggered) — `event="die" phase=""`
- **A-Heartfire Hero** (ability[1] triggered) — `event="die" phase=""`

### `trigger:combat_damage_player` — 382 hits

Examples:

- **A-Alrund, God of the Cosmos // A-Hakka, Whispering Raven** (ability[4] triggered) — `event="combat_damage_player" phase=""`
- **A-Dokuchi Silencer** (ability[1] triggered) — `event="combat_damage_player" phase=""`
- **A-Goggles of Night** (ability[0] triggered) — `event="combat_damage_player" phase=""`
- **A-Krydle of Baldur's Gate** (ability[0] triggered) — `event="combat_damage_player" phase=""`
- **A-Mischievous Catgeist // A-Catlike Curiosity** (ability[0] triggered) — `event="combat_damage_player" phase=""`

### `trigger:phase` — 257 hits

Examples:

- **A-Dreamshackle Geist** (ability[1] triggered) — `event="phase" phase="combat_start_yours"`
- **A-Sigil of Myrkul** (ability[0] triggered) — `event="phase" phase="combat_start_yours"`
- **Abhorrent Oculus** (ability[2] triggered) — `event="phase" phase="opponent"`
- **Academy Loremaster** (ability[0] triggered) — `event="phase" phase="player"`
- **Accursed Witch // Infectious Curse** (ability[4] triggered) — `event="phase" phase="enchanted_player"`

### `trigger:when_you_do` — 177 hits

Examples:

- **A-Ancestral Katana** (ability[1] triggered) — `event="when_you_do" phase=""`
- **A-Dokuchi Silencer** (ability[2] triggered) — `event="when_you_do" phase=""`
- **A-Guide of Souls** (ability[2] triggered) — `event="when_you_do" phase=""`
- **A-Minsc & Boo, Timeless Heroes** (ability[4] triggered) — `event="when_you_do" phase=""`
- **A-Sigil of Myrkul** (ability[1] triggered) — `event="when_you_do" phase=""`

### `trigger:etb_as` — 131 hits

Examples:

- **A-Thran Portal** (ability[0] triggered) — `event="etb_as" phase=""`
- **Adaptive Automaton** (ability[0] triggered) — `event="etb_as" phase=""`
- **Alhammarret, High Arbiter** (ability[1] triggered) — `event="etb_as" phase=""`
- **Alpine Moon** (ability[0] triggered) — `event="etb_as" phase=""`
- **Ancient Amphitheater** (ability[0] triggered) — `event="etb_as" phase=""`

### `cond:other` — 98 hits

Examples:

- **A-Armory Veteran** (ability[1] static.condition) — `as long as armory veteran is equipped, it has menace`
- **A-Dwarfhold Champion** (ability[1] static.condition) — `as long as dwarfhold champion is equipped, it gets +0/+2`
- **A-Nadu, Winged Wisdom** (ability[2] static.condition) — `if it's a land card, put it onto the battlefield. otherwise, put it into your hand`
- **Adanto Vanguard** (ability[0] static.condition) — `as long as this creature is attacking, it gets +2/+0`
- **Aerial Engineer** (ability[0] static.condition) — `as long as you control an artifact, this creature gets +2/+0 and has flying`

### `trigger:turned_face_up` — 73 hits

Examples:

- **Acid-Spewer Dragon** (ability[3] triggered) — `event="turned_face_up" phase=""`
- **Ainok Survivalist** (ability[1] triggered) — `event="turned_face_up" phase=""`
- **Alley Assailant** (ability[2] triggered) — `event="turned_face_up" phase=""`
- **Ashcloud Phoenix** (ability[3] triggered) — `event="turned_face_up" phase=""`
- **Atarka Efreet** (ability[1] triggered) — `event="turned_face_up" phase=""`

### `trigger:beginning_of_ordinal_step` — 69 hits

Examples:

- **Alora, Cheerful Assassin** (ability[1] triggered) — `event="beginning_of_ordinal_step" phase=""`
- **Alora, Cheerful Mastermind** (ability[1] triggered) — `event="beginning_of_ordinal_step" phase=""`
- **Alora, Cheerful Scout** (ability[1] triggered) — `event="beginning_of_ordinal_step" phase=""`
- **Alora, Cheerful Swashbuckler** (ability[1] triggered) — `event="beginning_of_ordinal_step" phase=""`
- **Alora, Cheerful Thief** (ability[1] triggered) — `event="beginning_of_ordinal_step" phase=""`

### `trigger:you_whenever` — 67 hits

Examples:

- **A-Satoru Umezawa** (ability[0] triggered) — `event="you_whenever" phase=""`
- **Alandra, Sky Dreamer** (ability[0] triggered) — `event="you_whenever" phase=""`
- **Ashnod the Uncaring** (ability[1] triggered) — `event="you_whenever" phase=""`
- **Bloodhaze Wolverine** (ability[0] triggered) — `event="you_whenever" phase=""`
- **Case File Auditor** (ability[1] triggered) — `event="you_whenever" phase=""`

### `trigger:self_and_another` — 63 hits

Examples:

- **Aetherstorm Roc** (ability[1] triggered) — `event="self_and_another" phase=""`
- **Ambuscade Shaman** (ability[0] triggered) — `event="self_and_another" phase=""`
- **Angelic Sell-Sword** (ability[2] triggered) — `event="self_and_another" phase=""`
- **Archon of Redemption** (ability[1] triggered) — `event="self_and_another" phase=""`
- **Basri's Lieutenant** (ability[3] triggered) — `event="self_and_another" phase=""`

### `trigger:self_and` — 46 hits

Examples:

- **Anax, Hardened in the Forge** (ability[1] triggered) — `event="self_and" phase=""`
- **Arahbo, the First Fang** (ability[1] triggered) — `event="self_and" phase=""`
- **Arbaaz Mir** (ability[0] triggered) — `event="self_and" phase=""`
- **Ardoz, Cobbler of War** (ability[1] triggered) — `event="self_and" phase=""`
- **Arthur, Marigold Knight** (ability[1] triggered) — `event="self_and" phase=""`

### `trigger:tribe_you_control_etb` — 45 hits

Examples:

- **Ajani's Welcome** (ability[0] triggered) — `event="tribe_you_control_etb" phase=""`
- **Angelic Chorus** (ability[0] triggered) — `event="tribe_you_control_etb" phase=""`
- **Answered Prayers** (ability[0] triggered) — `event="tribe_you_control_etb" phase=""`
- **Bat Colony** (ability[1] triggered) — `event="tribe_you_control_etb" phase=""`
- **Bishop of Wings** (ability[0] triggered) — `event="tribe_you_control_etb" phase=""`

### `trigger:group_combat_damage_player` — 43 hits

Examples:

- **A-Prosperous Thief** (ability[1] triggered) — `event="group_combat_damage_player" phase=""`
- **Alela, Cunning Conqueror** (ability[2] triggered) — `event="group_combat_damage_player" phase=""`
- **Anowon, the Ruin Thief** (ability[1] triggered) — `event="group_combat_damage_player" phase=""`
- **Automated Assembly Line** (ability[0] triggered) — `event="group_combat_damage_player" phase=""`
- **Bottomless Pool // Locker Room** (ability[1] triggered) — `event="group_combat_damage_player" phase=""`

### `trigger:deals_damage` — 43 hits

Examples:

- **Archfiend of Spite** (ability[1] triggered) — `event="deals_damage" phase=""`
- **Armadillo Cloak** (ability[2] triggered) — `event="deals_damage" phase=""`
- **Awaken the Sky Tyrant** (ability[0] triggered) — `event="deals_damage" phase=""`
- **Belltower Sphinx** (ability[1] triggered) — `event="deals_damage" phase=""`
- **Bloodfeather Phoenix** (ability[2] triggered) — `event="deals_damage" phase=""`

### `trigger:becomes_target` — 39 hits

Examples:

- **A-Brine Comber // A-Brinebound Gift** (ability[0] triggered) — `event="becomes_target" phase=""`
- **A-Nadu, Winged Wisdom** (ability[1] triggered) — `event="becomes_target" phase=""`
- **Agrus Kos, Eternal Soldier** (ability[1] triggered) — `event="becomes_target" phase=""`
- **Amulet of Safekeeping** (ability[0] triggered) — `event="becomes_target" phase=""`
- **Angelic Cub** (ability[0] triggered) — `event="becomes_target" phase=""`

### `trigger:nontoken_ally_event` — 37 hits

Examples:

- **Abzan Ascendancy** (ability[1] triggered) — `event="nontoken_ally_event" phase=""`
- **Anafenza, Kin-Tree Spirit** (ability[0] triggered) — `event="nontoken_ally_event" phase=""`
- **Bane, Lord of Darkness** (ability[1] triggered) — `event="nontoken_ally_event" phase=""`
- **Bhaal, Lord of Murder** (ability[1] triggered) — `event="nontoken_ally_event" phase=""`
- **Blessed Sanctuary** (ability[1] triggered) — `event="nontoken_ally_event" phase=""`

### `trigger:one_or_more_typed_event` — 36 hits

Examples:

- **Amzu, Swarm's Hunger** (ability[3] triggered) — `event="one_or_more_typed_event" phase=""`
- **Baron Bertram Graywater** (ability[0] triggered) — `event="one_or_more_typed_event" phase=""`
- **Blood Spatter Analysis** (ability[1] triggered) — `event="one_or_more_typed_event" phase=""`
- **Caretaker's Talent** (ability[0] triggered) — `event="one_or_more_typed_event" phase=""`
- **Defiled Crypt // Cadaver Lab** (ability[0] triggered) — `event="one_or_more_typed_event" phase=""`

### `trigger:dealt_damage` — 35 hits

Examples:

- **Arcbond** (ability[1] triggered) — `event="dealt_damage" phase=""`
- **Barbed Servitor** (ability[3] triggered) — `event="dealt_damage" phase=""`
- **Blazing Sunsteel** (ability[1] triggered) — `event="dealt_damage" phase=""`
- **Body of Knowledge** (ability[2] triggered) — `event="dealt_damage" phase=""`
- **Boros Reckoner** (ability[0] triggered) — `event="dealt_damage" phase=""`

### `trigger:specialize_creature` — 31 hits

Examples:

- **Gut, Bestial Fanatic** (ability[0] triggered) — `event="specialize_creature" phase=""`
- **Gut, Brutal Fanatic** (ability[0] triggered) — `event="specialize_creature" phase=""`
- **Gut, Devious Fanatic** (ability[0] triggered) — `event="specialize_creature" phase=""`
- **Gut, Furious Fanatic** (ability[0] triggered) — `event="specialize_creature" phase=""`
- **Gut, Zealous Fanatic** (ability[0] triggered) — `event="specialize_creature" phase=""`

### `trigger:cycle` — 30 hits

Examples:

- **Avian Oddity** (ability[2] triggered) — `event="cycle" phase=""`
- **Choking Tethers** (ability[2] triggered) — `event="cycle" phase=""`
- **Decree of Justice** (ability[2] triggered) — `event="cycle" phase=""`
- **Decree of Pain** (ability[4] triggered) — `event="cycle" phase=""`
- **Decree of Savagery** (ability[2] triggered) — `event="cycle" phase=""`

### `trigger:mutates` — 30 hits

Examples:

- **Archipelagore** (ability[1] triggered) — `event="mutates" phase=""`
- **Boneyard Lurker** (ability[1] triggered) — `event="mutates" phase=""`
- **Cavern Whisperer** (ability[2] triggered) — `event="mutates" phase=""`
- **Chittering Harvester** (ability[1] triggered) — `event="mutates" phase=""`
- **Cloudpiercer** (ability[2] triggered) — `event="mutates" phase=""`

### `trigger:unlock_door` — 29 hits

Examples:

- **Bottomless Pool // Locker Room** (ability[0] triggered) — `event="unlock_door" phase=""`
- **Cramped Vents // Access Maze** (ability[0] triggered) — `event="unlock_door" phase=""`
- **Crude Abattoir // Unsavory Kitchen** (ability[0] triggered) — `event="unlock_door" phase=""`
- **Defiled Crypt // Cadaver Lab** (ability[2] triggered) — `event="unlock_door" phase=""`
- **Derelict Attic // Widow's Walk** (ability[0] triggered) — `event="unlock_door" phase=""`

### `trigger:becomes_blocked` — 28 hits

Examples:

- **Anzrag, the Quake-Mole** (ability[0] triggered) — `event="becomes_blocked" phase=""`
- **Azusa's Many Journeys // Likeness of the Seeker** (ability[3] triggered) — `event="becomes_blocked" phase=""`
- **Baneblade Scoundrel // Baneclaw Marauder** (ability[0] triggered) — `event="becomes_blocked" phase=""`
- **Battle-Scarred Goblin** (ability[0] triggered) — `event="becomes_blocked" phase=""`
- **Bill Ferny, Bree Swindler** (ability[0] triggered) — `event="becomes_blocked" phase=""`

### `trigger:becomes_tapped` — 26 hits

Examples:

- **Annie Flash, the Veteran** (ability[2] triggered) — `event="becomes_tapped" phase=""`
- **Armored Scrapgorger** (ability[2] triggered) — `event="becomes_tapped" phase=""`
- **Attentive Sunscribe** (ability[0] triggered) — `event="becomes_tapped" phase=""`
- **Emmara, Soul of the Accord** (ability[0] triggered) — `event="becomes_tapped" phase=""`
- **Fallowsage** (ability[0] triggered) — `event="becomes_tapped" phase=""`

### `trigger:misc_when` — 25 hits

Examples:

- **A-Stitched Assistant** (ability[1] triggered) — `event="misc_when" phase=""`
- **Blessed Defiance** (ability[1] triggered) — `event="misc_when" phase=""`
- **Burn Away** (ability[1] triggered) — `event="misc_when" phase=""`
- **Cryptek** (ability[1] triggered) — `event="misc_when" phase=""`
- **Felonious Rage** (ability[1] triggered) — `event="misc_when" phase=""`

### `cond:has-counters-on` — 25 hits

Examples:

- **A-Sigardian Paladin** (ability[0] static.condition) — `as long as you've put one or more +1/+1 counters on a creature this turn, sigardian paladin has trample and lifelink`
- **Angelic Cub** (ability[1] static.condition) — `as long as this creature has three or more +1/+1 counters on it, it has flying`
- **Beastmaster Ascension** (ability[1] static.condition) — `as long as this enchantment has seven or more quest counters on it, creatures you control get +5/+5`
- **Bounty of the Luxa** (ability[1] static.condition) — `if no counters were removed this way, put a flood counter on this enchantment and draw a card. otherwise, add {c}{g}{u}`
- **Faithbound Judge // Sinner's Judgment** (ability[4] static.condition) — `as long as this creature has three or more judgment counters on it, it can attack as though it didn't have defender`

### `trigger:you_action` — 21 hits

Examples:

- **Alluring Suitor // Deadly Dancer** (ability[0] triggered) — `event="you_action" phase=""`
- **Argentum Masticore** (ability[3] triggered) — `event="you_action" phase=""`
- **Bartered Cow** (ability[1] triggered) — `event="you_action" phase=""`
- **Borborygmos and Fblthp** (ability[1] triggered) — `event="you_action" phase=""`
- **Calibrated Blast** (ability[2] triggered) — `event="you_action" phase=""`

### `trigger:to_graveyard` — 20 hits

Examples:

- **Aetherworks Marvel** (ability[0] triggered) — `event="to_graveyard" phase=""`
- **Audacity** (ability[2] triggered) — `event="to_graveyard" phase=""`
- **Bitter Chill** (ability[3] triggered) — `event="to_graveyard" phase=""`
- **Creeping Chill** (ability[1] triggered) — `event="to_graveyard" phase=""`
- **Demonic Ruckus** (ability[2] triggered) — `event="to_graveyard" phase=""`

### `trigger:exploits_creature` — 19 hits

Examples:

- **Diver Skaab** (ability[1] triggered) — `event="exploits_creature" phase=""`
- **Fell Stinger** (ability[2] triggered) — `event="exploits_creature" phase=""`
- **Graf Reaver** (ability[1] triggered) — `event="exploits_creature" phase=""`
- **Gurmag Drowner** (ability[1] triggered) — `event="exploits_creature" phase=""`
- **Infernal Captor** (ability[1] triggered) — `event="exploits_creature" phase=""`

### `trigger:opp_creature_event` — 19 hits

Examples:

- **Archfiend of the Dross** (ability[3] triggered) — `event="opp_creature_event" phase=""`
- **Authority of the Consuls** (ability[1] triggered) — `event="opp_creature_event" phase=""`
- **Blood Seeker** (ability[0] triggered) — `event="opp_creature_event" phase=""`
- **Fall of Cair Andros** (ability[0] triggered) — `event="opp_creature_event" phase=""`
- **Gimli, Counter of Kills** (ability[1] triggered) — `event="opp_creature_event" phase=""`

### `trigger:commit_crime` — 18 hits

Examples:

- **At Knifepoint** (ability[1] triggered) — `event="commit_crime" phase=""`
- **Bandit's Haul** (ability[0] triggered) — `event="commit_crime" phase=""`
- **Blood Hustler** (ability[0] triggered) — `event="commit_crime" phase=""`
- **Deepmuck Desperado** (ability[0] triggered) — `event="commit_crime" phase=""`
- **Duelist of the Mind** (ability[3] triggered) — `event="commit_crime" phase=""`

### `trigger:self_put_into_graveyard_from_bf` — 17 hits

Examples:

- **Chromatic Star** (ability[1] triggered) — `event="self_put_into_graveyard_from_bf" phase=""`
- **Dunes of the Dead** (ability[1] triggered) — `event="self_put_into_graveyard_from_bf" phase=""`
- **Flagstones of Trokair** (ability[1] triggered) — `event="self_put_into_graveyard_from_bf" phase=""`
- **Goblin Boom Keg** (ability[1] triggered) — `event="self_put_into_graveyard_from_bf" phase=""`
- **Implement of Combustion** (ability[1] triggered) — `event="self_put_into_graveyard_from_bf" phase=""`

### `cond:tap-state` — 16 hits

Examples:

- **Archangel of Tithes** (ability[1] static.condition) — `as long as this creature is untapped, creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures`
- **Archelos, Lagoon Mystic** (ability[0] static.condition) — `as long as ~ is tapped, other permanents enter tapped`
- **Castle Raptors** (ability[1] static.condition) — `as long as this creature is untapped, it gets +0/+2`
- **Dire-Strain Rampage** (ability[1] static.condition) — `if a land was destroyed this way, its controller may search their library for up to two basic land cards, put them onto the battlefield tapped, then shuffle. otherwise, its controller may search their library for a basic…`
- **Hans Eriksson** (ability[1] static.condition) — `if it's a creature card, put it onto the battlefield tapped and attacking defending player or a planeswalker they control. otherwise, put that card into your hand`

### `trigger:transforms` — 16 hits

Examples:

- **Alluring Suitor // Deadly Dancer** (ability[2] triggered) — `event="transforms" phase=""`
- **Avacynian Missionaries // Lunarch Inquisitors** (ability[1] triggered) — `event="transforms" phase=""`
- **Blightreaper Thallid // Blightsower Thallid** (ability[2] triggered) — `event="transforms" phase=""`
- **Brutal Cathar // Moonrage Brute** (ability[0] triggered) — `event="transforms" phase=""`
- **Captive Weird // Compleated Conjurer** (ability[3] triggered) — `event="transforms" phase=""`

### `trigger:ally_targeted_by_opp` — 16 hits

Examples:

- **Altanak, the Thrice-Called** (ability[1] triggered) — `event="ally_targeted_by_opp" phase=""`
- **Ashenmoor Liege** (ability[2] triggered) — `event="ally_targeted_by_opp" phase=""`
- **Battle Mammoth** (ability[1] triggered) — `event="ally_targeted_by_opp" phase=""`
- **Cactarantula** (ability[2] triggered) — `event="ally_targeted_by_opp" phase=""`
- **Diffusion Sliver** (ability[0] triggered) — `event="ally_targeted_by_opp" phase=""`

### `trigger:fully_unlock_room` — 16 hits

Examples:

- **Balemurk Leech** (ability[1] triggered) — `event="fully_unlock_room" phase=""`
- **Cult Healer** (ability[1] triggered) — `event="fully_unlock_room" phase=""`
- **Dashing Bloodsucker** (ability[1] triggered) — `event="fully_unlock_room" phase=""`
- **Entity Tracker** (ability[2] triggered) — `event="fully_unlock_room" phase=""`
- **Erratic Apparition** (ability[3] triggered) — `event="fully_unlock_room" phase=""`

### `trigger:nontoken_creature_event` — 15 hits

Examples:

- **Alharu, Solemn Ritualist** (ability[1] triggered) — `event="nontoken_creature_event" phase=""`
- **Bridge from Below** (ability[0] triggered) — `event="nontoken_creature_event" phase=""`
- **Curse of Clinging Webs** (ability[1] triggered) — `event="nontoken_creature_event" phase=""`
- **Faerie Artisans** (ability[1] triggered) — `event="nontoken_creature_event" phase=""`
- **Field of Souls** (ability[0] triggered) — `event="nontoken_creature_event" phase=""`

### `trigger:counters_put_on_self` — 15 hits

Examples:

- **Axgard Artisan** (ability[0] triggered) — `event="counters_put_on_self" phase=""`
- **Basking Broodscale** (ability[2] triggered) — `event="counters_put_on_self" phase=""`
- **Benthic Biomancer** (ability[1] triggered) — `event="counters_put_on_self" phase=""`
- **Constable of the Realm** (ability[1] triggered) — `event="counters_put_on_self" phase=""`
- **Dreamdrinker Vampire** (ability[2] triggered) — `event="counters_put_on_self" phase=""`

### `trigger:self_deals_damage_player` — 13 hits

Examples:

- **Abyssal Specter** (ability[1] triggered) — `event="self_deals_damage_player" phase=""`
- **Atemsis, All-Seeing** (ability[2] triggered) — `event="self_deals_damage_player" phase=""`
- **Charnelhoard Wurm** (ability[1] triggered) — `event="self_deals_damage_player" phase=""`
- **Hidetsugu Consumes All // Vessel of the All-Consuming** (ability[5] triggered) — `event="self_deals_damage_player" phase=""`
- **Looter il-Kor** (ability[1] triggered) — `event="self_deals_damage_player" phase=""`

### `trigger:you_expend_n` — 13 hits

Examples:

- **Bakersbane Duo** (ability[1] triggered) — `event="you_expend_n" phase=""`
- **Bark-Knuckle Boxer** (ability[0] triggered) — `event="you_expend_n" phase=""`
- **Brambleguard Veteran** (ability[0] triggered) — `event="you_expend_n" phase=""`
- **Byway Barterer** (ability[1] triggered) — `event="you_expend_n" phase=""`
- **Hoarder's Overflow** (ability[1] triggered) — `event="you_expend_n" phase=""`

### `trigger:combat_damage` — 13 hits

Examples:

- **Aisha of Sparks and Smoke** (ability[2] triggered) — `event="combat_damage" phase=""`
- **Arashin War Beast** (ability[0] triggered) — `event="combat_damage" phase=""`
- **Edric, Spymaster of Trest** (ability[0] triggered) — `event="combat_damage" phase=""`
- **Firkraag, Cunning Instigator** (ability[3] triggered) — `event="combat_damage" phase=""`
- **Great Train Heist** (ability[5] triggered) — `event="combat_damage" phase=""`

### `trigger:you_scry` — 12 hits

Examples:

- **Arwen Undómiel** (ability[0] triggered) — `event="you_scry" phase=""`
- **Celeborn the Wise** (ability[1] triggered) — `event="you_scry" phase=""`
- **Chance-Met Elves** (ability[0] triggered) — `event="you_scry" phase=""`
- **Council's Deliberation** (ability[1] triggered) — `event="you_scry" phase=""`
- **Elminster** (ability[0] triggered) — `event="you_scry" phase=""`

### `cond:blocked-or-blocking` — 12 hits

Examples:

- **A-Futurist Operative** (ability[0] static.condition) — `as long as futurist operative is tapped, it's a human citizen with base power and toughness 1/1 and can't be blocked`
- **Ace's Baseball Bat** (ability[1] static.condition) — `as long as equipped creature is attacking, it has first strike and must be blocked by a dalek if able`
- **Detective of the Month** (ability[1] static.condition) — `as long as you have the city's blessing, detectives you control can't be blocked`
- **Enkira, Hostile Scavenger** (ability[1] static.condition) — `as long as ~ is equipped, it must be blocked if able`
- **Frodo Baggins** (ability[1] static.condition) — `as long as ~ is your ring-bearer, it must be blocked if able`

### `trigger:another_typed_etb` — 12 hits

Examples:

- **Garruk's Packleader** (ability[0] triggered) — `event="another_typed_etb" phase=""`
- **Inspiring Commander** (ability[0] triggered) — `event="another_typed_etb" phase=""`
- **Irreverent Gremlin** (ability[1] triggered) — `event="another_typed_etb" phase=""`
- **Marketwatch Phantom** (ability[0] triggered) — `event="another_typed_etb" phase=""`
- **Neighborhood Guardian** (ability[0] triggered) — `event="another_typed_etb" phase=""`

### `trigger:cycle_card` — 12 hits

Examples:

- **Benalish Partisan** (ability[1] triggered) — `event="cycle_card" phase=""`
- **Crystalline Resonance** (ability[0] triggered) — `event="cycle_card" phase=""`
- **Drannith Healer** (ability[0] triggered) — `event="cycle_card" phase=""`
- **Drannith Stinger** (ability[0] triggered) — `event="cycle_card" phase=""`
- **Escape Protocol** (ability[0] triggered) — `event="cycle_card" phase=""`

### `trigger:combat_damage_player_or_pw` — 11 hits

Examples:

- **Bladewing, Deathless Tyrant** (ability[2] triggered) — `event="combat_damage_player_or_pw" phase=""`
- **Dreadhorde Butcher** (ability[1] triggered) — `event="combat_damage_player_or_pw" phase=""`
- **Garruk's Harbinger** (ability[1] triggered) — `event="combat_damage_player_or_pw" phase=""`
- **Grateful Apparition** (ability[1] triggered) — `event="combat_damage_player_or_pw" phase=""`
- **Guildpact Informant** (ability[1] triggered) — `event="combat_damage_player_or_pw" phase=""`

### `trigger:tap_for_mana` — 11 hits

Examples:

- **Crypt Ghast** (ability[1] triggered) — `event="tap_for_mana" phase=""`
- **Forbidden Orchard** (ability[1] triggered) — `event="tap_for_mana" phase=""`
- **Leyline of Abundance** (ability[1] triggered) — `event="tap_for_mana" phase=""`
- **Mirari's Wake** (ability[1] triggered) — `event="tap_for_mana" phase=""`
- **Nikya of the Old Ways** (ability[1] triggered) — `event="tap_for_mana" phase=""`

### `trigger:tapped_for_mana` — 11 hits

Examples:

- **Blighted Burgeoning** (ability[2] triggered) — `event="tapped_for_mana" phase=""`
- **Buried in the Garden** (ability[2] triggered) — `event="tapped_for_mana" phase=""`
- **C.A.M.P.** (ability[0] triggered) — `event="tapped_for_mana" phase=""`
- **Dawn's Reflection** (ability[1] triggered) — `event="tapped_for_mana" phase=""`
- **Extraplanar Lens** (ability[1] triggered) — `event="tapped_for_mana" phase=""`

### `trigger:block` — 11 hits

Examples:

- **A-Dorothea, Vengeful Victim // A-Dorothea's Retribution** (ability[1] triggered) — `event="block" phase=""`
- **Carnage Gladiator** (ability[0] triggered) — `event="block" phase=""`
- **Dwindle** (ability[2] triggered) — `event="block" phase=""`
- **Guardian of the Gateless** (ability[2] triggered) — `event="block" phase=""`
- **Jareth, Leonine Titan** (ability[0] triggered) — `event="block" phase=""`

### `trigger:becomes_monstrous` — 11 hits

Examples:

- **Arbor Colossus** (ability[2] triggered) — `event="becomes_monstrous" phase=""`
- **Death Kiss** (ability[2] triggered) — `event="becomes_monstrous" phase=""`
- **Grim Giganotosaurus** (ability[2] triggered) — `event="becomes_monstrous" phase=""`
- **Hydra Broodmaster** (ability[1] triggered) — `event="becomes_monstrous" phase=""`
- **Kalemne's Captain** (ability[2] triggered) — `event="becomes_monstrous" phase=""`

### `trigger:class_becomes_level` — 11 hits

Examples:

- **A-Druid Class** (ability[4] triggered) — `event="class_becomes_level" phase="3"`
- **A-Wizard Class** (ability[2] triggered) — `event="class_becomes_level" phase="2"`
- **Artificer Class** (ability[2] triggered) — `event="class_becomes_level" phase="2"`
- **Builder's Talent** (ability[4] triggered) — `event="class_becomes_level" phase="3"`
- **Caretaker's Talent** (ability[3] triggered) — `event="class_becomes_level" phase="2"`

### `trigger:day_night_flip` — 10 hits

Examples:

- **Brimstone Vandal** (ability[2] triggered) — `event="day_night_flip" phase=""`
- **Celestus Sanctifier** (ability[1] triggered) — `event="day_night_flip" phase=""`
- **Component Collector** (ability[1] triggered) — `event="day_night_flip" phase=""`
- **Firmament Sage** (ability[1] triggered) — `event="day_night_flip" phase=""`
- **Gavony Dawnguard** (ability[2] triggered) — `event="day_night_flip" phase=""`

### `trigger:equipped_trigger` — 10 hits

Examples:

- **Aegis of the Legion** (ability[1] triggered) — `event="equipped_trigger" phase=""`
- **Bladehold Cleaver** (ability[2] triggered) — `event="equipped_trigger" phase=""`
- **Eater of Virtue** (ability[0] triggered) — `event="equipped_trigger" phase=""`
- **Forebear's Blade** (ability[1] triggered) — `event="equipped_trigger" phase=""`
- **Halvar, God of Battle // Sword of the Realms** (ability[3] triggered) — `event="equipped_trigger" phase=""`

### `trigger:block_or_becomes_blocked` — 10 hits

Examples:

- **Ashmouth Hound** (ability[0] triggered) — `event="block_or_becomes_blocked" phase=""`
- **Assembled Alphas** (ability[0] triggered) — `event="block_or_becomes_blocked" phase=""`
- **Barrow-Blade** (ability[1] triggered) — `event="block_or_becomes_blocked" phase=""`
- **Corrosive Ooze** (ability[0] triggered) — `event="block_or_becomes_blocked" phase=""`
- **Gorgon Recluse** (ability[0] triggered) — `event="block_or_becomes_blocked" phase=""`

### `trigger:specialize_from_zone` — 10 hits

Examples:

- **Karlach, Tiefling Berserker** (ability[2] triggered) — `event="specialize_from_zone" phase=""`
- **Karlach, Tiefling Guardian** (ability[2] triggered) — `event="specialize_from_zone" phase=""`
- **Karlach, Tiefling Punisher** (ability[2] triggered) — `event="specialize_from_zone" phase=""`
- **Karlach, Tiefling Spellrager** (ability[2] triggered) — `event="specialize_from_zone" phase=""`
- **Karlach, Tiefling Zealot** (ability[2] triggered) — `event="specialize_from_zone" phase=""`

### `trigger:misc_whenever_a` — 10 hits

Examples:

- **Angelic Renewal** (ability[0] triggered) — `event="misc_whenever_a" phase=""`
- **Erebos's Titan** (ability[1] triggered) — `event="misc_whenever_a" phase=""`
- **Lazav, Dimir Mastermind** (ability[1] triggered) — `event="misc_whenever_a" phase=""`
- **Lorcan, Warlock Collector** (ability[1] triggered) — `event="misc_whenever_a" phase=""`
- **Oglor, Devoted Assistant** (ability[1] triggered) — `event="misc_whenever_a" phase=""`

### `trigger:ally_typed_etb` — 10 hits

Examples:

- **Anointer Priest** (ability[0] triggered) — `event="ally_typed_etb" phase=""`
- **Cloudstone Curio** (ability[0] triggered) — `event="ally_typed_etb" phase=""`
- **Cryptid Inspector** (ability[1] triggered) — `event="ally_typed_etb" phase=""`
- **Nesting Dovehawk** (ability[2] triggered) — `event="ally_typed_etb" phase=""`
- **Replication Specialist** (ability[1] triggered) — `event="ally_typed_etb" phase=""`

### `trigger:combat_damage_opponent` — 9 hits

Examples:

- **Coastal Piracy** (ability[0] triggered) — `event="combat_damage_opponent" phase=""`
- **Etrata, Deadly Fugitive** (ability[2] triggered) — `event="combat_damage_opponent" phase=""`
- **Grenzo's Ruffians** (ability[1] triggered) — `event="combat_damage_opponent" phase=""`
- **Hydra Omnivore** (ability[0] triggered) — `event="combat_damage_opponent" phase=""`
- **Kediss, Emberclaw Familiar** (ability[0] triggered) — `event="combat_damage_opponent" phase=""`

### `trigger:counters_put_on_actor` — 9 hits

Examples:

- **A-Moss-Pit Skeleton** (ability[2] triggered) — `event="counters_put_on_actor" phase=""`
- **Bioessence Hydra** (ability[2] triggered) — `event="counters_put_on_actor" phase=""`
- **Botanical Brawler** (ability[2] triggered) — `event="counters_put_on_actor" phase=""`
- **Cloaked Cadet** (ability[1] triggered) — `event="counters_put_on_actor" phase=""`
- **Enduring Scalelord** (ability[1] triggered) — `event="counters_put_on_actor" phase=""`

### `trigger:block_creature` — 9 hits

Examples:

- **A-High-Rise Sawjack** (ability[1] triggered) — `event="block_creature" phase=""`
- **Aetherplasm** (ability[0] triggered) — `event="block_creature" phase=""`
- **Cleric of Chill Depths** (ability[0] triggered) — `event="block_creature" phase=""`
- **Ezuri's Archers** (ability[1] triggered) — `event="block_creature" phase=""`
- **High-Rise Sawjack** (ability[1] triggered) — `event="block_creature" phase=""`

### `trigger:combat_damage_creature` — 9 hits

Examples:

- **Frostwalk Bastion** (ability[3] triggered) — `event="combat_damage_creature" phase=""`
- **Giant's Skewer** (ability[1] triggered) — `event="combat_damage_creature" phase=""`
- **Mirri the Cursed** (ability[3] triggered) — `event="combat_damage_creature" phase=""`
- **Obelisk Spider** (ability[1] triggered) — `event="combat_damage_creature" phase=""`
- **Ohran Viper** (ability[0] triggered) — `event="combat_damage_creature" phase=""`

### `trigger:self_combat_damage` — 8 hits

Examples:

- **Cartographer's Hawk** (ability[1] triggered) — `event="self_combat_damage" phase=""`
- **Jeskai Infiltrator** (ability[1] triggered) — `event="self_combat_damage" phase=""`
- **Kathari Bomber** (ability[1] triggered) — `event="self_combat_damage" phase=""`
- **RMS Titanic** (ability[2] triggered) — `event="self_combat_damage" phase=""`
- **Rahilda, Wanted Cutthroat // Rahilda, Feral Outlaw** (ability[1] triggered) — `event="self_combat_damage" phase=""`

### `trigger:combat_damage_player_or_battle` — 8 hits

Examples:

- **Aetherblade Agent // Gitaxian Mindstinger** (ability[4] triggered) — `event="combat_damage_player_or_battle" phase=""`
- **Archpriest of Shadows** (ability[2] triggered) — `event="combat_damage_player_or_battle" phase=""`
- **Attentive Skywarden** (ability[1] triggered) — `event="combat_damage_player_or_battle" phase=""`
- **Beamtown Beatstick** (ability[1] triggered) — `event="combat_damage_player_or_battle" phase=""`
- **Deeproot Wayfinder** (ability[0] triggered) — `event="combat_damage_player_or_battle" phase=""`

### `trigger:end_step` — 8 hits

Examples:

- **Agitator Ant** (ability[0] triggered) — `event="end_step" phase=""`
- **Arvinox, the Mind Flail** (ability[1] triggered) — `event="end_step" phase=""`
- **Dress Down** (ability[3] triggered) — `event="end_step" phase=""`
- **Gluntch, the Bestower** (ability[1] triggered) — `event="end_step" phase=""`
- **Sandwurm Convergence** (ability[1] triggered) — `event="end_step" phase=""`

### `cond:top-of-library` — 8 hits

Examples:

- **Bucolic Ranch** (ability[4] static.condition) — `if it's a mount card, you may reveal it and put it into your hand. if you don't put it into your hand, you may put it on the bottom of your library`
- **Cabaretti Ascendancy** (ability[1] static.condition) — `if it's a creature or planeswalker card, you may reveal it and put it into your hand. if you don't put the card into your hand, you may put it on the bottom of your library`
- **Chittering Illuminator** (ability[0] static.condition) — `as long as ~ is at the top of your library, you may look at it any time and you may cast it`
- **Expert-Level Safe** (ability[2] static.condition) — `if they match, sacrifice this artifact and put all cards exiled with it into their owners' hands. otherwise, exile the top card of your library face down`
- **Kenessos, Priest of Thassa** (ability[2] static.condition) — `if it's a kraken, leviathan, octopus, or serpent creature card, you may put it onto the battlefield. if you don't put the card onto the battlefield, you may put it on the bottom of your library`

### `trigger:to_gy_from_anywhere` — 8 hits

Examples:

- **Emrakul, the Aeons Torn** (ability[5] triggered) — `event="to_gy_from_anywhere" phase=""`
- **Guile** (ability[2] triggered) — `event="to_gy_from_anywhere" phase=""`
- **Hostility** (ability[3] triggered) — `event="to_gy_from_anywhere" phase=""`
- **Kozilek, Butcher of Truth** (ability[2] triggered) — `event="to_gy_from_anywhere" phase=""`
- **Serra Avatar** (ability[1] triggered) — `event="to_gy_from_anywhere" phase=""`

### `trigger:spend_this_mana` — 8 hits

Examples:

- **A-Jade Orb of Dragonkind** (ability[1] triggered) — `event="spend_this_mana" phase=""`
- **A-Lapis Orb of Dragonkind** (ability[1] triggered) — `event="spend_this_mana" phase=""`
- **Forger's Foundry** (ability[1] triggered) — `event="spend_this_mana" phase=""`
- **Gilanra, Caller of Wirewood** (ability[1] triggered) — `event="spend_this_mana" phase=""`
- **Jade Orb of Dragonkind** (ability[1] triggered) — `event="spend_this_mana" phase=""`

### `trigger:self_or_typed_event` — 7 hits

Examples:

- **Cacophony Unleashed** (ability[1] triggered) — `event="self_or_typed_event" phase=""`
- **Dowsing Device // Geode Grotto** (ability[0] triggered) — `event="self_or_typed_event" phase=""`
- **Dragonspark Reactor** (ability[0] triggered) — `event="self_or_typed_event" phase=""`
- **Field of the Dead** (ability[2] triggered) — `event="self_or_typed_event" phase=""`
- **Historian's Boon** (ability[0] triggered) — `event="self_or_typed_event" phase=""`

### `trigger:you_roll_dice` — 7 hits

Examples:

- **Barbarian Class** (ability[2] triggered) — `event="you_roll_dice" phase=""`
- **Brazen Dwarf** (ability[0] triggered) — `event="you_roll_dice" phase=""`
- **Feywild Trickster** (ability[0] triggered) — `event="you_roll_dice" phase=""`
- **Mr. House, President and CEO** (ability[0] triggered) — `event="you_roll_dice" phase=""`
- **Vexing Puzzlebox** (ability[0] triggered) — `event="you_roll_dice" phase=""`

### `trigger:ally_source_damage` — 7 hits

Examples:

- **Auntie Blyte, Bad Influence** (ability[1] triggered) — `event="ally_source_damage" phase=""`
- **Chandra's Incinerator** (ability[2] triggered) — `event="ally_source_damage" phase=""`
- **Chandra's Pyreling** (ability[0] triggered) — `event="ally_source_damage" phase=""`
- **Crude Abattoir // Unsavory Kitchen** (ability[1] triggered) — `event="ally_source_damage" phase=""`
- **Niv-Mizzet, Visionary** (ability[2] triggered) — `event="ally_source_damage" phase=""`

### `cond:strike-keyword` — 7 hits

Examples:

- **Faithful Pikemaster** (ability[1] static.condition) — `as long as it's your turn, this creature has first strike`
- **Gnoll Hunting Party** (ability[2] static.condition) — `as long as it's your turn, ~ has first strike`
- **Hanweir Lancer** (ability[1] static.condition) — `as long as this creature is paired with another creature, both creatures have first strike`
- **Kor Duelist** (ability[0] static.condition) — `as long as this creature is equipped, it has double strike`
- **Spinehorn Minotaur** (ability[0] static.condition) — `as long as you've drawn two or more cards this turn, this creature has double strike`

### `trigger:ally_type_to_gy_from_bf` — 7 hits

Examples:

- **Ashiok's Reaper** (ability[0] triggered) — `event="ally_type_to_gy_from_bf" phase=""`
- **Knight of Doves** (ability[0] triggered) — `event="ally_type_to_gy_from_bf" phase=""`
- **Marionette Master** (ability[1] triggered) — `event="ally_type_to_gy_from_bf" phase=""`
- **Neva, Stalked by Nightmares** (ability[2] triggered) — `event="ally_type_to_gy_from_bf" phase=""`
- **Savior of the Sleeping** (ability[1] triggered) — `event="ally_type_to_gy_from_bf" phase=""`

### `trigger:ring_tempts_you` — 7 hits

Examples:

- **Aragorn, Company Leader** (ability[0] triggered) — `event="ring_tempts_you" phase=""`
- **Faramir, Field Commander** (ability[1] triggered) — `event="ring_tempts_you" phase=""`
- **Galadriel of Lothlórien** (ability[0] triggered) — `event="ring_tempts_you" phase=""`
- **Gandalf, Friend of the Shire** (ability[2] triggered) — `event="ring_tempts_you" phase=""`
- **Nazgûl** (ability[2] triggered) — `event="ring_tempts_you" phase=""`

### `trigger:becomes_untapped` — 6 hits

Examples:

- **Fishing Pole** (ability[1] triggered) — `event="becomes_untapped" phase=""`
- **Ghostly Pilferer** (ability[0] triggered) — `event="becomes_untapped" phase=""`
- **Immovable Rod** (ability[1] triggered) — `event="becomes_untapped" phase=""`
- **Key to the City** (ability[1] triggered) — `event="becomes_untapped" phase=""`
- **Mesmeric Orb** (ability[0] triggered) — `event="becomes_untapped" phase=""`

### `trigger:you_put_counters_on` — 6 hits

Examples:

- **Defiant Greatmaw** (ability[1] triggered) — `event="you_put_counters_on" phase=""`
- **Kate Stewart** (ability[0] triggered) — `event="you_put_counters_on" phase=""`
- **Nest of Scarabs** (ability[0] triggered) — `event="you_put_counters_on" phase=""`
- **Obelisk Spider** (ability[2] triggered) — `event="you_put_counters_on" phase=""`
- **Omarthis, Ghostfire Initiate** (ability[1] triggered) — `event="you_put_counters_on" phase=""`

### `trigger:creature_cards_leave_gy` — 6 hits

Examples:

- **Chalk Outline** (ability[0] triggered) — `event="creature_cards_leave_gy" phase=""`
- **Desecrated Tomb** (ability[0] triggered) — `event="creature_cards_leave_gy" phase=""`
- **Insidious Roots** (ability[1] triggered) — `event="creature_cards_leave_gy" phase=""`
- **Rot Farm Mortipede** (ability[0] triggered) — `event="creature_cards_leave_gy" phase=""`
- **Skeleton Crew** (ability[1] triggered) — `event="creature_cards_leave_gy" phase=""`

### `trigger:upkeep` — 6 hits

Examples:

- **Bell Borca, Spectral Sergeant** (ability[2] triggered) — `event="upkeep" phase=""`
- **Chitterspitter** (ability[0] triggered) — `event="upkeep" phase=""`
- **Herald's Horn** (ability[2] triggered) — `event="upkeep" phase=""`
- **Journey to the Lost City** (ability[0] triggered) — `event="upkeep" phase=""`
- **Master of the Wild Hunt** (ability[0] triggered) — `event="upkeep" phase=""`

### `trigger:pay_cost_multiple` — 6 hits

Examples:

- **Bloodthirsty Adversary** (ability[2] triggered) — `event="pay_cost_multiple" phase=""`
- **Intrepid Adversary** (ability[2] triggered) — `event="pay_cost_multiple" phase=""`
- **Primal Adversary** (ability[2] triggered) — `event="pay_cost_multiple" phase=""`
- **Spectral Adversary** (ability[3] triggered) — `event="pay_cost_multiple" phase=""`
- **Tainted Adversary** (ability[2] triggered) — `event="pay_cost_multiple" phase=""`

### `trigger:until_eot_trigger` — 6 hits

Examples:

- **Apple of Eden, Isu Relic** (ability[2] triggered) — `event="until_eot_trigger" phase=""`
- **Benefactor's Draught** (ability[1] triggered) — `event="until_eot_trigger" phase=""`
- **Bonus Round** (ability[0] triggered) — `event="until_eot_trigger" phase=""`
- **High Tide** (ability[0] triggered) — `event="until_eot_trigger" phase=""`
- **Leori, Sparktouched Hunter** (ability[3] triggered) — `event="until_eot_trigger" phase=""`

### `cond:exile-state` — 6 hits

Examples:

- **Agrus Kos, Spirit of Justice** (ability[3] static.condition) — `if it's suspected, exile it. otherwise, suspect it`
- **Check for Traps** (ability[3] static.condition) — `if an instant card or a card with flash is exiled this way, they lose 1 life. otherwise, you lose 1 life`
- **Eater of Virtue** (ability[2] static.condition) — `as long as a card exiled with ~ has flying, equipped creature has flying`
- **Mirage Phalanx** (ability[1] static.condition) — `as long as ~ is paired with another creature, each of those creatures has "at the beginning of combat on your turn, create a token that's a copy of this creature, except it has haste and loses soulbond. exile it at end o…`
- **Protection Racket** (ability[3] static.condition) — `if they do, exile that card. otherwise, put it into your hand`

### `trigger:becomes_state` — 6 hits

Examples:

- **Assimilation Aegis** (ability[1] triggered) — `event="becomes_state" phase=""`
- **Blade of Shared Souls** (ability[1] triggered) — `event="becomes_state" phase=""`
- **Bramble Elemental** (ability[0] triggered) — `event="becomes_state" phase=""`
- **Deeproot Pilgrimage** (ability[0] triggered) — `event="becomes_state" phase=""`
- **Enormous Energy Blade** (ability[1] triggered) — `event="becomes_state" phase=""`

### `trigger:becomes_blocked_by` — 6 hits

Examples:

- **Acolyte of the Inferno** (ability[1] triggered) — `event="becomes_blocked_by" phase=""`
- **Gloom Sower** (ability[0] triggered) — `event="becomes_blocked_by" phase=""`
- **Grasping Giant** (ability[1] triggered) — `event="becomes_blocked_by" phase=""`
- **Kolaghan Aspirant** (ability[0] triggered) — `event="becomes_blocked_by" phase=""`
- **Nessian Boar** (ability[1] triggered) — `event="becomes_blocked_by" phase=""`

### `trigger:you_surveil` — 6 hits

Examples:

- **Blood Operative** (ability[2] triggered) — `event="you_surveil" phase=""`
- **Copy Catchers** (ability[1] triggered) — `event="you_surveil" phase=""`
- **Dimir Spybug** (ability[2] triggered) — `event="you_surveil" phase=""`
- **Disinformation Campaign** (ability[1] triggered) — `event="you_surveil" phase=""`
- **Mirko, Obsessive Theorist** (ability[2] triggered) — `event="you_surveil" phase=""`

### `trigger:ally_etb` — 6 hits

Examples:

- **Ezuri, Claw of Progress** (ability[0] triggered) — `event="ally_etb" phase=""`
- **Kiora, Behemoth Beckoner** (ability[0] triggered) — `event="ally_etb" phase=""`
- **Kronch Wrangler** (ability[1] triggered) — `event="ally_etb" phase=""`
- **Symmetry Matrix** (ability[0] triggered) — `event="ally_etb" phase=""`
- **Territorial Boar** (ability[0] triggered) — `event="ally_etb" phase=""`

### `cond:toughness-matters` — 6 hits

Examples:

- **Awakened Awareness** (ability[2] static.condition) — `as long as enchanted permanent is a creature, it has base power and toughness 1/1`
- **Duplicant** (ability[1] static.condition) — `as long as a card exiled with this creature is a creature card, this creature has the power, toughness, and creature types of the last creature card exiled with it`
- **Imperious Mindbreaker** (ability[1] static.condition) — `as long as ~ is paired with another creature, each of those creatures has "whenever this creature attacks, each opponent mills cards equal to its toughness."`
- **Living Conundrum** (ability[2] static.condition) — `as long as there are no cards in your library, this creature has base power and toughness 10/10 and has flying and vigilance`
- **Timber Paladin** (ability[0] static.condition) — `as long as this creature is enchanted by exactly one aura, it has base power and toughness 3/3`

### `trigger:ally_explore` — 6 hits

Examples:

- **Lurking Chupacabra** (ability[0] triggered) — `event="ally_explore" phase=""`
- **Merfolk Cave-Diver** (ability[0] triggered) — `event="ally_explore" phase=""`
- **Nicanzil, Current Conductor** (ability[0] triggered) — `event="ally_explore" phase=""`
- **Shadowed Caravel** (ability[0] triggered) — `event="ally_explore" phase=""`
- **Wildgrowth Walker** (ability[0] triggered) — `event="ally_explore" phase=""`

### `trigger:you_exert_creature` — 5 hits

Examples:

- **Battlefield Scavenger** (ability[1] triggered) — `event="you_exert_creature" phase=""`
- **Resolute Survivors** (ability[1] triggered) — `event="you_exert_creature" phase=""`
- **Rohirrim Chargers** (ability[1] triggered) — `event="you_exert_creature" phase=""`
- **Trueheart Twins** (ability[1] triggered) — `event="you_exert_creature" phase=""`
- **Vizier of the True** (ability[1] triggered) — `event="you_exert_creature" phase=""`

### `cond:hand-size` — 5 hits

Examples:

- **Bounty of the Deep** (ability[0] static.condition) — `if you have no land cards in your hand, seek a land card and a nonland card. otherwise, seek two nonland cards`
- **Carnage Interpreter** (ability[1] static.condition) — `as long as you have one or fewer cards in hand, this creature gets +2/+2 and has menace`
- **Djeru and Hazoret** (ability[0] static.condition) — `as long as you have one or fewer cards in hand, ~ has vigilance and haste`
- **Neheb, the Worthy** (ability[2] static.condition) — `as long as you have one or fewer cards in hand, minotaurs you control get +2/+0`
- **New Perspectives** (ability[1] static.condition) — `as long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs`

### `trigger:creature_etb_any` — 5 hits

Examples:

- **Dredging Claw** (ability[1] triggered) — `event="creature_etb_any" phase=""`
- **Mana Echoes** (ability[0] triggered) — `event="creature_etb_any" phase=""`
- **Mirror of Life Trapping** (ability[0] triggered) — `event="creature_etb_any" phase=""`
- **Pandemonium** (ability[0] triggered) — `event="creature_etb_any" phase=""`
- **Wild Pair** (ability[0] triggered) — `event="creature_etb_any" phase=""`

### `trigger:permanent_to_gy` — 5 hits

Examples:

- **A-Shipwreck Sifters** (ability[1] triggered) — `event="permanent_to_gy" phase=""`
- **Disa the Restless** (ability[0] triggered) — `event="permanent_to_gy" phase=""`
- **Mazzy, Truesword Paladin** (ability[1] triggered) — `event="permanent_to_gy" phase=""`
- **Pia's Revolution** (ability[0] triggered) — `event="permanent_to_gy" phase=""`
- **Yuma, Proud Protector** (ability[2] triggered) — `event="permanent_to_gy" phase=""`

### `trigger:legend_ally_event` — 5 hits

Examples:

- **Hero's Blade** (ability[1] triggered) — `event="legend_ally_event" phase=""`
- **Kellan Joins Up** (ability[1] triggered) — `event="legend_ally_event" phase=""`
- **Rakdos Joins Up** (ability[1] triggered) — `event="legend_ally_event" phase=""`
- **The Irencrag** (ability[1] triggered) — `event="legend_ally_event" phase=""`
- **Tinybones Joins Up** (ability[1] triggered) — `event="legend_ally_event" phase=""`

### `trigger:you_proliferate` — 5 hits

Examples:

- **Contagion Dispenser** (ability[1] triggered) — `event="you_proliferate" phase=""`
- **Ezuri, Stalker of Spheres** (ability[1] triggered) — `event="you_proliferate" phase=""`
- **Ichor Aberration** (ability[3] triggered) — `event="you_proliferate" phase=""`
- **Scheming Aspirant** (ability[0] triggered) — `event="you_proliferate" phase=""`
- **Venser, Corpse Puppet** (ability[2] triggered) — `event="you_proliferate" phase=""`

### `cond:by-name` — 5 hits

Examples:

- **Approach of the Second Sun** (ability[0] static.condition) — `if this spell was cast from your hand and you've cast another spell named ~ this game, you win the game. otherwise, put ~ into its owner's library seventh from the top and you gain 7 life`
- **Awestruck Cygnet** (ability[2] static.condition) — `as long as this card's intensity is 3 or more, it has base power and toughness 4/4, has flying and vigilance, and is named radiant swan`
- **Mothers Yamazaki** (ability[1] static.condition) — `as long as you control exactly two permanents named ~, the "legend rule" doesn't apply to them, and samurai you control get +2/+2 and have vigilance and haste`
- **Nazahn, Revered Bladesmith** (ability[1] static.condition) — `if you reveal a card named hammer of ~ this way, put it onto the battlefield. otherwise, put that card into your hand. then shuffle`
- **Predict** (ability[1] static.condition) — `if a card with the chosen name was milled this way, you draw two cards. otherwise, you draw a card`

### `cond:dungeon` — 5 hits

Examples:

- **A-Cloister Gargoyle** (ability[1] static.condition) — `as long as you've completed a dungeon, cloister gargoyle gets +3/+0 and has flying`
- **A-Triumphant Adventurer** (ability[1] static.condition) — `as long as it's your turn, triumphant adventurer has first strike`
- **Caves of Chaos Adventurer** (ability[3] static.condition) — `if you've completed a dungeon, you may play that card this turn without paying its mana cost. otherwise, you may play that card this turn`
- **Cloister Gargoyle** (ability[1] static.condition) — `as long as you've completed a dungeon, this creature gets +3/+0 and has flying`
- **Gloom Stalker** (ability[0] static.condition) — `as long as you've completed a dungeon, this creature has double strike`

### `cond:ring-tempted` — 5 hits

Examples:

- **Archdemon of Paliano** (ability[1] static.condition) — `as long as this card is face up during the draft, you can't look at booster packs and must draft cards at random`
- **Conqueror's Flail** (ability[1] static.condition) — `as long as this equipment is attached to a creature, your opponents can't cast spells during your turn`
- **Melt Through** (ability[1] static.condition) — `if it's a creature, it perpetually gains "as long as this creature is on the battlefield, damage isn't removed from it during cleanup steps."`
- **Switchgrass Grazer** (ability[2] static.condition) — `if ~ is saddled and a creature was dealt damage this way, that creature perpetually gains "this creature can't block" and "damage isn't removed from this creature during cleanup steps."`
- **Tail Swipe** (ability[1] static.condition) — `if you cast this spell during your main phase, the creature you control gets +1/+1 until end of turn. then those creatures fight each other`

### `trigger:token_event` — 5 hits

Examples:

- **Boomer Scrapper** (ability[1] triggered) — `event="token_event" phase=""`
- **Junk Winder** (ability[1] triggered) — `event="token_event" phase=""`
- **Nadier's Nightblade** (ability[0] triggered) — `event="token_event" phase=""`
- **Nadier, Agent of the Duskenel** (ability[0] triggered) — `event="token_event" phase=""`
- **Wildwood Mentor** (ability[0] triggered) — `event="token_event" phase=""`

### `cond:permanent-count` — 5 hits

Examples:

- **Fecund Greenshell** (ability[1] static.condition) — `as long as you control ten or more lands, creatures you control get +2/+2`
- **Path of Bravery** (ability[0] static.condition) — `as long as your life total is greater than or equal to your starting life total, creatures you control get +1/+1`
- **Raksha Golden Cub** (ability[1] static.condition) — `as long as ~ is equipped, cat creatures you control get +2/+2 and have double strike`
- **Sentinel Sarah Lyons** (ability[1] static.condition) — `as long as an artifact entered the battlefield under your control this turn, creatures you control get +2/+2`
- **Sylvan Advocate** (ability[1] static.condition) — `as long as you control six or more lands, this creature and land creatures you control get +2/+2`

### `trigger:another_creature_or_artifact_event` — 5 hits

Examples:

- **Charforger** (ability[1] triggered) — `event="another_creature_or_artifact_event" phase=""`
- **Exuberant Fuseling** (ability[3] triggered) — `event="another_creature_or_artifact_event" phase=""`
- **Forgehammer Centurion** (ability[0] triggered) — `event="another_creature_or_artifact_event" phase=""`
- **Marionette Apprentice** (ability[1] triggered) — `event="another_creature_or_artifact_event" phase=""`
- **Necrosquito** (ability[3] triggered) — `event="another_creature_or_artifact_event" phase=""`

### `trigger:card_to_gy_anywhere` — 4 hits

Examples:

- **Bloodchief Ascension** (ability[1] triggered) — `event="card_to_gy_anywhere" phase=""`
- **Nihilith** (ability[2] triggered) — `event="card_to_gy_anywhere" phase=""`
- **The Haunt of Hightower** (ability[3] triggered) — `event="card_to_gy_anywhere" phase=""`
- **Vulturous Zombie** (ability[1] triggered) — `event="card_to_gy_anywhere" phase=""`

### `cond:mana-value` — 4 hits

Examples:

- **Containment Breach** (ability[1] static.condition) — `if its mana value is 2 or less, create a 1/1 black and green pest creature token with "when this token dies, you gain 1 life."`
- **Cosmic Rebirth** (ability[1] static.condition) — `if it has mana value 3 or less, you may put it onto the battlefield. if you don't put it onto the battlefield, put it into your hand`
- **Sheoldred's Restoration** (ability[2] static.condition) — `if this spell was kicked, you gain life equal to that card's mana value. otherwise, you lose that much life`
- **Sin Prodder** (ability[3] static.condition) — `if a player does, this creature deals damage to that player equal to that card's mana value. otherwise, put that card into your hand`

### `trigger:compound_opp_tribe_event` — 4 hits

Examples:

- **Magmatic Galleon** (ability[1] triggered) — `event="compound_opp_tribe_event" phase=""`
- **Nelly Borca, Impulsive Accuser** (ability[2] triggered) — `event="compound_opp_tribe_event" phase=""`
- **Norn's Decree** (ability[0] triggered) — `event="compound_opp_tribe_event" phase=""`
- **Spiteful Banditry** (ability[1] triggered) — `event="compound_opp_tribe_event" phase=""`

### `cond:power-comparison` — 4 hits

Examples:

- **Ichor Aberration** (ability[2] static.condition) — `as long as ~'s power is 7 or greater, it can attack as though it didn't have defender`
- **Karsus Depthguard** (ability[1] static.condition) — `as long as this creature's power is 5 or greater, it can attack as though it didn't have defender`
- **Starfield of Nyx** (ability[1] static.condition) — `as long as you control five or more enchantments, each other non-aura enchantment you control is a creature in addition to its other types and has base power and base toughness each equal to its mana value`
- **Timber Paladin** (ability[2] static.condition) — `as long as this creature is enchanted by three or more auras, it has base power and toughness 10/10, vigilance, and trample`

### `trigger:one_or_more_lands` — 4 hits

Examples:

- **Hedge Shredder** (ability[1] triggered) — `event="one_or_more_lands" phase=""`
- **Sand Scout** (ability[1] triggered) — `event="one_or_more_lands" phase=""`
- **Titania, Voice of Gaea** (ability[1] triggered) — `event="one_or_more_lands" phase=""`
- **Turntimber Sower** (ability[0] triggered) — `event="one_or_more_lands" phase=""`

### `trigger:land_etb_any` — 4 hits

Examples:

- **A-Tatyova, Steward of Tides** (ability[1] triggered) — `event="land_etb_any" phase=""`
- **Gitrog, Horror of Zhava** (ability[3] triggered) — `event="land_etb_any" phase=""`
- **Sarinth Greatwurm** (ability[1] triggered) — `event="land_etb_any" phase=""`
- **Zo-Zu the Punisher** (ability[0] triggered) — `event="land_etb_any" phase=""`

### `trigger:you_put_counters_on_any` — 4 hits

Examples:

- **All Will Be One** (ability[0] triggered) — `event="you_put_counters_on_any" phase=""`
- **Aragorn, Company Leader** (ability[1] triggered) — `event="you_put_counters_on_any" phase=""`
- **Generous Patron** (ability[1] triggered) — `event="you_put_counters_on_any" phase=""`
- **Kros, Defense Contractor** (ability[1] triggered) — `event="you_put_counters_on_any" phase=""`

### `trigger:ally_exploits` — 4 hits

Examples:

- **A-Skull Skaab** (ability[1] triggered) — `event="ally_exploits" phase=""`
- **Colonel Autumn** (ability[3] triggered) — `event="ally_exploits" phase=""`
- **Henry Wu, InGen Geneticist** (ability[1] triggered) — `event="ally_exploits" phase=""`
- **Skull Skaab** (ability[1] triggered) — `event="ally_exploits" phase=""`

### `trigger:opp_type_to_gy_from_bf` — 4 hits

Examples:

- **Kibo, Uktabi Prince** (ability[1] triggered) — `event="opp_type_to_gy_from_bf" phase=""`
- **Pain Distributor** (ability[2] triggered) — `event="opp_type_to_gy_from_bf" phase=""`
- **Sardian Avenger** (ability[3] triggered) — `event="opp_type_to_gy_from_bf" phase=""`
- **Sarulf, Realm Eater** (ability[0] triggered) — `event="opp_type_to_gy_from_bf" phase=""`

### `trigger:creature_cards_to_zone` — 4 hits

Examples:

- **Crawling Infestation** (ability[1] triggered) — `event="creature_cards_to_zone" phase=""`
- **Sefris of the Hidden Ways** (ability[0] triggered) — `event="creature_cards_to_zone" phase=""`
- **Sidisi, Brood Tyrant** (ability[1] triggered) — `event="creature_cards_to_zone" phase=""`
- **Unshakable Tail** (ability[2] triggered) — `event="creature_cards_to_zone" phase=""`

### `trigger:player_land_play` — 4 hits

Examples:

- **Cemetery Gatekeeper** (ability[2] triggered) — `event="player_land_play" phase=""`
- **Horn of Greed** (ability[0] triggered) — `event="player_land_play" phase=""`
- **Rocco, Street Chef** (ability[2] triggered) — `event="player_land_play" phase=""`
- **Sandcloud Harbinger** (ability[1] triggered) — `event="player_land_play" phase=""`

### `trigger:you_get_energy` — 4 hits

Examples:

- **Aether Revolt** (ability[1] triggered) — `event="you_get_energy" phase=""`
- **Brotherhood Scribe** (ability[2] triggered) — `event="you_get_energy" phase=""`
- **Fabrication Module** (ability[0] triggered) — `event="you_get_energy" phase=""`
- **Territorial Gorger** (ability[1] triggered) — `event="you_get_energy" phase=""`

### `trigger:type_to_gy_from_bf` — 4 hits

Examples:

- **Disciple of the Vault** (ability[0] triggered) — `event="type_to_gy_from_bf" phase=""`
- **Ich-Tekik, Salvage Splicer** (ability[1] triggered) — `event="type_to_gy_from_bf" phase=""`
- **Krenko, Baron of Tin Street** (ability[2] triggered) — `event="type_to_gy_from_bf" phase=""`
- **Molder Beast** (ability[1] triggered) — `event="type_to_gy_from_bf" phase=""`

### `trigger:opp_activate` — 3 hits

Examples:

- **Harsh Mentor** (ability[0] triggered) — `event="opp_activate" phase=""`
- **Immolation Shaman** (ability[0] triggered) — `event="opp_activate" phase=""`
- **Runic Armasaur** (ability[0] triggered) — `event="opp_activate" phase=""`

### `trigger:combat_damage_to_player` — 3 hits

Examples:

- **Ancient Gold Dragon** (ability[1] triggered) — `event="combat_damage_to_player" phase=""`
- **Gishath, Sun's Avatar** (ability[3] triggered) — `event="combat_damage_to_player" phase=""`
- **Sword of Hearth and Home** (ability[1] triggered) — `event="combat_damage_to_player" phase=""`

### `trigger:to_gy_from_bf` — 3 hits

Examples:

- **Hopeful Vigil** (ability[1] triggered) — `event="to_gy_from_bf" phase=""`
- **Hopeless Nightmare** (ability[1] triggered) — `event="to_gy_from_bf" phase=""`
- **Induced Amnesia** (ability[1] triggered) — `event="to_gy_from_bf" phase=""`

### `trigger:one_or_more_creatures_combat_damage` — 3 hits

Examples:

- **Contaminant Grafter** (ability[2] triggered) — `event="one_or_more_creatures_combat_damage" phase=""`
- **Forth Eorlingas!** (ability[1] triggered) — `event="one_or_more_creatures_combat_damage" phase=""`
- **Witch-king of Angmar** (ability[1] triggered) — `event="one_or_more_creatures_combat_damage" phase=""`

### `cond:x-equals` — 3 hits

Examples:

- **Elturel Survivors** (ability[2] static.condition) — `as long as this creature is attacking, it gets +x/+0, where x is the number of lands defending player controls`
- **Mephit's Enthusiasm** (ability[1] static.condition) — `if excess damage was dealt this way, note that excess damage, then you get a one-time boon with "when you cast a creature spell, it perpetually gets +x/+0, where x is the noted number."`
- **Sardian Cliffstomper** (ability[0] static.condition) — `as long as it's your turn and you control four or more mountains, this creature gets +x/+0, where x is the number of mountains you control`

### `trigger:creature_combat_damage_you` — 3 hits

Examples:

- **Hixus, Prison Warden** (ability[1] triggered) — `event="creature_combat_damage_you" phase=""`
- **Strixhaven Stadium** (ability[2] triggered) — `event="creature_combat_damage_you" phase=""`
- **Teysa, Envoy of Ghosts** (ability[2] triggered) — `event="creature_combat_damage_you" phase=""`

### `trigger:any_player_sacs` — 3 hits

Examples:

- **Carmen, Cruel Skymarcher** (ability[1] triggered) — `event="any_player_sacs" phase=""`
- **Mortician Beetle** (ability[0] triggered) — `event="any_player_sacs" phase=""`
- **Thraximundar** (ability[2] triggered) — `event="any_player_sacs" phase=""`

### `trigger:ally_typed_to_gy` — 3 hits

Examples:

- **Farid, Enterprising Salvager** (ability[0] triggered) — `event="ally_typed_to_gy" phase=""`
- **Sly Requisitioner** (ability[1] triggered) — `event="ally_typed_to_gy" phase=""`
- **Teysa, Opulent Oligarch** (ability[2] triggered) — `event="ally_typed_to_gy" phase=""`

### `trigger:saga_final_chapter` — 3 hits

Examples:

- **Historian's Boon** (ability[1] triggered) — `event="saga_final_chapter" phase=""`
- **Narci, Fable Singer** (ability[2] triggered) — `event="saga_final_chapter" phase=""`
- **Tom Bombadil** (ability[1] triggered) — `event="saga_final_chapter" phase=""`

### `trigger:counter_removed_from_self` — 3 hits

Examples:

- **Aeon Chronicler** (ability[3] triggered) — `event="counter_removed_from_self" phase=""`
- **Benalish Commander** (ability[3] triggered) — `event="counter_removed_from_self" phase=""`
- **Dinosaurs on a Spaceship** (ability[4] triggered) — `event="counter_removed_from_self" phase=""`

### `trigger:attached_as` — 3 hits

Examples:

- **Paleontologist's Pick-Axe // Dinosaur Headdress** (ability[4] triggered) — `event="attached_as" phase=""`
- **Psychic Paper** (ability[0] triggered) — `event="attached_as" phase=""`
- **Sanctuary Blade** (ability[0] triggered) — `event="attached_as" phase=""`

### `trigger:exiled_event` — 3 hits

Examples:

- **Agatha's Soul Cauldron** (ability[3] triggered) — `event="exiled_event" phase=""`
- **Market Gnome** (ability[1] triggered) — `event="exiled_event" phase=""`
- **Soulherder** (ability[0] triggered) — `event="exiled_event" phase=""`

### `trigger:self_card_zone_to_zone` — 3 hits

Examples:

- **Gaea's Blessing** (ability[2] triggered) — `event="self_card_zone_to_zone" phase=""`
- **Golgari Brownscale** (ability[0] triggered) — `event="self_card_zone_to_zone" phase=""`
- **Narcomoeba** (ability[1] triggered) — `event="self_card_zone_to_zone" phase=""`

### `cond:card-count-compare` — 3 hits

Examples:

- **Evangel of Synthesis** (ability[1] static.condition) — `as long as you've drawn two or more cards this turn, this creature gets +1/+0 and has menace`
- **Gnarled Sage** (ability[1] static.condition) — `as long as you've drawn two or more cards this turn, this creature gets +0/+2 and has vigilance`
- **Trench Stalker** (ability[0] static.condition) — `as long as you've drawn two or more cards this turn, this creature has deathtouch and lifelink`

### `trigger:you_dealt_damage` — 3 hits

Examples:

- **Contested Game Ball** (ability[0] triggered) — `event="you_dealt_damage" phase=""`
- **Darien, King of Kjeldor** (ability[0] triggered) — `event="you_dealt_damage" phase=""`
- **Sun Droplet** (ability[0] triggered) — `event="you_dealt_damage" phase=""`

### `trigger:you_mechanic` — 3 hits

Examples:

- **Lurker in the Deep** (ability[2] triggered) — `event="you_mechanic" phase=""`
- **Paranormal Analyst** (ability[0] triggered) — `event="you_mechanic" phase=""`
- **Vexyr, Ich-Tekik's Heir** (ability[0] triggered) — `event="you_mechanic" phase=""`

### `trigger:ally_subtype_deal_damage` — 3 hits

Examples:

- **Breeches, Brazen Plunderer** (ability[1] triggered) — `event="ally_subtype_deal_damage" phase=""`
- **Francisco, Fowl Marauder** (ability[2] triggered) — `event="ally_subtype_deal_damage" phase=""`
- **Malcolm, Keen-Eyed Navigator** (ability[1] triggered) — `event="ally_subtype_deal_damage" phase=""`

### `trigger:self_crews_vehicle` — 3 hits

Examples:

- **Gearshift Ace** (ability[1] triggered) — `event="self_crews_vehicle" phase=""`
- **Speedway Fanatic** (ability[1] triggered) — `event="self_crews_vehicle" phase=""`
- **Veteran Motorist** (ability[1] triggered) — `event="self_crews_vehicle" phase=""`

### `trigger:as_you_draft_a_card` — 3 hits

Examples:

- **Animus of Predation** (ability[1] triggered) — `event="as_you_draft_a_card" phase=""`
- **Leovold's Operative** (ability[1] triggered) — `event="as_you_draft_a_card" phase=""`
- **Smuggler Captain** (ability[1] triggered) — `event="as_you_draft_a_card" phase=""`

### `trigger:land_tapped_for_mana` — 3 hits

Examples:

- **Treasure Nabber** (ability[0] triggered) — `event="land_tapped_for_mana" phase=""`
- **Vorinclex, Voice of Hunger** (ability[2] triggered) — `event="land_tapped_for_mana" phase=""`
- **War's Toll** (ability[0] triggered) — `event="land_tapped_for_mana" phase=""`

### `trigger:tap_opp_creature` — 3 hits

Examples:

- **Hylda of the Icy Crown** (ability[0] triggered) — `event="tap_opp_creature" phase=""`
- **Icewrought Sentry** (ability[3] triggered) — `event="tap_opp_creature" phase=""`
- **Solitary Sanctuary** (ability[1] triggered) — `event="tap_opp_creature" phase=""`

### `trigger:one_or_more_other_ally_event` — 3 hits

Examples:

- **Dour Port-Mage** (ability[0] triggered) — `event="one_or_more_other_ally_event" phase=""`
- **Frantic Scapegoat** (ability[2] triggered) — `event="one_or_more_other_ally_event" phase=""`
- **Sally Sparrow** (ability[1] triggered) — `event="one_or_more_other_ally_event" phase=""`

### `trigger:counters_removed_from_self` — 2 hits

Examples:

- **Chandra, Fire Artisan** (ability[0] triggered) — `event="counters_removed_from_self" phase=""`
- **Regenerations Restored** (ability[1] triggered) — `event="counters_removed_from_self" phase=""`

### `trigger:you_conjure_one_or_more` — 2 hits

Examples:

- **Thayan Evokers** (ability[2] triggered) — `event="you_conjure_one_or_more" phase=""`
- **Third Little Pig** (ability[0] triggered) — `event="you_conjure_one_or_more" phase=""`

### `trigger:equipment_attach_state_change` — 2 hits

Examples:

- **Captain's Hook** (ability[1] triggered) — `event="equipment_attach_state_change" phase=""`
- **Grafted Wargear** (ability[1] triggered) — `event="equipment_attach_state_change" phase=""`

### `trigger:becomes_renowned` — 2 hits

Examples:

- **Relic Seeker** (ability[1] triggered) — `event="becomes_renowned" phase=""`
- **Valeron Wardens** (ability[1] triggered) — `event="becomes_renowned" phase=""`

### `trigger:becomes_crewed` — 2 hits

Examples:

- **Mobilizer Mech** (ability[1] triggered) — `event="becomes_crewed" phase=""`
- **Protean War Engine** (ability[1] triggered) — `event="becomes_crewed" phase=""`

### `trigger:ally_typed_etb_a` — 2 hits

Examples:

- **Chishiro, the Shattered Blade** (ability[0] triggered) — `event="ally_typed_etb_a" phase=""`
- **Era of Innovation** (ability[0] triggered) — `event="ally_typed_etb_a" phase=""`

### `trigger:self_blocks` — 2 hits

Examples:

- **Loyal Sentry** (ability[0] triggered) — `event="self_blocks" phase=""`
- **Wall of Junk** (ability[1] triggered) — `event="self_blocks" phase=""`

### `trigger:become_monarch` — 2 hits

Examples:

- **Custodi Lich** (ability[1] triggered) — `event="become_monarch" phase=""`
- **Knights of the Black Rose** (ability[1] triggered) — `event="become_monarch" phase=""`

### `trigger:compound_card_zone_event` — 2 hits

Examples:

- **Dennick, Pious Apprentice // Dennick, Pious Apparition** (ability[4] triggered) — `event="compound_card_zone_event" phase=""`
- **Thran Vigil** (ability[0] triggered) — `event="compound_card_zone_event" phase=""`

### `trigger:opp_searches_library` — 2 hits

Examples:

- **Archivist of Oghma** (ability[1] triggered) — `event="opp_searches_library" phase=""`
- **Ob Nixilis, Unshackled** (ability[2] triggered) — `event="opp_searches_library" phase=""`

### `trigger:you_misc_event` — 2 hits

Examples:

- **Jolly Gerbils** (ability[0] triggered) — `event="you_misc_event" phase=""`
- **Model of Unity** (ability[0] triggered) — `event="you_misc_event" phase=""`

### `trigger:opp_creature_to_gy` — 2 hits

Examples:

- **Bridge from Below** (ability[1] triggered) — `event="opp_creature_to_gy" phase=""`
- **Magus of the Bridge** (ability[1] triggered) — `event="opp_creature_to_gy" phase=""`

### `trigger:any_cycle` — 2 hits

Examples:

- **Invigorating Boon** (ability[0] triggered) — `event="any_cycle" phase=""`
- **Lightning Rift** (ability[0] triggered) — `event="any_cycle" phase=""`

### `trigger:one_or_more_ally_with_x_enter` — 2 hits

Examples:

- **Bess, Soul Nourisher** (ability[0] triggered) — `event="one_or_more_ally_with_X_enter" phase=""`
- **Enduring Innocence** (ability[1] triggered) — `event="one_or_more_ally_with_X_enter" phase=""`

### `trigger:becomes_crewed_first` — 2 hits

Examples:

- **Mighty Servant of Leuk-o** (ability[2] triggered) — `event="becomes_crewed_first" phase=""`
- **Mindlink Mech** (ability[1] triggered) — `event="becomes_crewed_first" phase=""`

### `trigger:any_player_tap_land` — 2 hits

Examples:

- **Barbflare Gremlin** (ability[2] triggered) — `event="any_player_tap_land" phase=""`
- **Heartbeat of Spring** (ability[0] triggered) — `event="any_player_tap_land" phase=""`

### `trigger:face_down_creature_event` — 2 hits

Examples:

- **Cryptic Pursuit** (ability[1] triggered) — `event="face_down_creature_event" phase=""`
- **Yarus, Roar of the Old Gods** (ability[2] triggered) — `event="face_down_creature_event" phase=""`

### `trigger:any_player_loses_game` — 2 hits

Examples:

- **Blood Tyrant** (ability[4] triggered) — `event="any_player_loses_game" phase=""`
- **Ramses, Assassin Lord** (ability[2] triggered) — `event="any_player_loses_game" phase=""`

### `trigger:enchanted_perm_to_gy` — 2 hits

Examples:

- **Gremlin Infestation** (ability[2] triggered) — `event="enchanted_perm_to_gy" phase=""`
- **Tezzeret's Touch** (ability[2] triggered) — `event="enchanted_perm_to_gy" phase=""`

### `trigger:coin_flip_result` — 2 hits

Examples:

- **Chance Encounter** (ability[0] triggered) — `event="coin_flip_result" phase=""`
- **Tavern Scoundrel** (ability[0] triggered) — `event="coin_flip_result" phase=""`

### `trigger:mill_event` — 2 hits

Examples:

- **Glowing One** (ability[2] triggered) — `event="mill_event" phase=""`
- **Infesting Radroach** (ability[3] triggered) — `event="mill_event" phase=""`

### `cond:would-effect` — 2 hits

Examples:

- **Multiclass Baldric** (ability[1] static.condition) — `as long as you have a full party, prevent all damage that would be dealt to equipped creature`
- **Pokey, the Scallywagg** (ability[1] static.condition) — `if you would flip a coin, you may instead roll a d20. 1−10 is tails and 11−20 is heads`

### `trigger:remove_counter` — 2 hits

Examples:

- **Immard, the Stormcleaver** (ability[1] triggered) — `event="remove_counter" phase=""`
- **Watcher of Hours** (ability[2] triggered) — `event="remove_counter" phase=""`

### `trigger:this_card_event` — 2 hits

Examples:

- **Aloe Alchemist** (ability[1] triggered) — `event="this_card_event" phase=""`
- **Longhorn Sharpshooter** (ability[1] triggered) — `event="this_card_event" phase=""`

### `trigger:lose_game` — 2 hits

Examples:

- **Curse of Vengeance** (ability[2] triggered) — `event="lose_game" phase=""`
- **Sengir, the Dark Baron** (ability[2] triggered) — `event="lose_game" phase=""`

### `trigger:conditional_state` — 2 hits

Examples:

- **Floodgate** (ability[1] triggered) — `event="conditional_state" phase=""`
- **Impetuous Devils** (ability[2] triggered) — `event="conditional_state" phase=""`

### `trigger:player_wins_coin_flip` — 2 hits

Examples:

- **Okaun, Eye of Chaos** (ability[2] triggered) — `event="player_wins_coin_flip" phase=""`
- **Zndrsplt, Eye of Wisdom** (ability[2] triggered) — `event="player_wins_coin_flip" phase=""`

### `trigger:aura_attached_event` — 2 hits

Examples:

- **Eriette, the Beguiler** (ability[1] triggered) — `event="aura_attached_event" phase=""`
- **Siona, Captain of the Pyleas** (ability[3] triggered) — `event="aura_attached_event" phase=""`

### `trigger:opp_dealt_damage` — 2 hits

Examples:

- **Chandra's Spitfire** (ability[1] triggered) — `event="opp_dealt_damage" phase=""`
- **Wildfire Elemental** (ability[0] triggered) — `event="opp_dealt_damage" phase=""`

### `trigger:it_state_change` — 2 hits

Examples:

- **Flavor Disaster** (ability[2] triggered) — `event="it_state_change" phase=""`
- **Melira, the Living Cure** (ability[2] triggered) — `event="it_state_change" phase=""`

### `cond:dealt-damage` — 2 hits

Examples:

- **Molten Impact** (ability[1] static.condition) — `if excess damage was dealt this way, note that excess damage, then you get a one-time boon with "when you cast an instant or sorcery spell, this boon deals damage equal to the noted number to target creature or planeswal…`
- **Tandem Lookout** (ability[1] static.condition) — `as long as ~ is paired with another creature, each of those creatures has "whenever this creature deals damage to an opponent, draw a card."`

### `trigger:one_or_more_other_creatures` — 2 hits

Examples:

- **Sengir Connoisseur** (ability[1] triggered) — `event="one_or_more_other_creatures" phase=""`
- **Soul Shredder** (ability[1] triggered) — `event="one_or_more_other_creatures" phase=""`

### `trigger:compound_tribe_enter` — 2 hits

Examples:

- **Builder's Talent** (ability[2] triggered) — `event="compound_tribe_enter" phase=""`
- **Losheel, Clockwork Scholar** (ability[1] triggered) — `event="compound_tribe_enter" phase=""`

### `trigger:one_or_more_milled` — 2 hits

Examples:

- **Mirelurk Queen** (ability[2] triggered) — `event="one_or_more_milled" phase=""`
- **Screeching Scorchbeast** (ability[3] triggered) — `event="one_or_more_milled" phase=""`

### `trigger:flip` — 2 hits

Examples:

- **Breeches, the Blastmaker** (ability[2] triggered) — `event="flip" phase=""`

### `trigger:creature_etb` — 2 hits

Examples:

- **Champion of Lambholt** (ability[1] triggered) — `event="creature_etb" phase=""`
- **Terror of the Peaks** (ability[2] triggered) — `event="creature_etb" phase=""`

### `trigger:dealt_combat_damage` — 2 hits

Examples:

- **Wall of Essence** (ability[1] triggered) — `event="dealt_combat_damage" phase=""`
- **Wall of Souls** (ability[1] triggered) — `event="dealt_combat_damage" phase=""`

### `trigger:damage_prevented_this_way` — 2 hits

Examples:

- **Outfitted Jouster** (ability[2] triggered) — `event="damage_prevented_this_way" phase=""`
- **Phyrexian Vindicator** (ability[2] triggered) — `event="damage_prevented_this_way" phase=""`

### `trigger:one_or_more_ally_creatures` — 2 hits

Examples:

- **Big Spender** (ability[1] triggered) — `event="one_or_more_ally_creatures" phase=""`
- **Hezrou // Demonic Stench** (ability[0] triggered) — `event="one_or_more_ally_creatures" phase=""`

### `trigger:self_ability_activated` — 2 hits

Examples:

- **Battlemage's Bracers** (ability[1] triggered) — `event="self_ability_activated" phase=""`
- **Illusionist's Bracers** (ability[0] triggered) — `event="self_ability_activated" phase=""`

### `trigger:face_up_as` — 2 hits

Examples:

- **Bubble Smuggler** (ability[1] triggered) — `event="face_up_as" phase=""`
- **Hooded Hydra** (ability[3] triggered) — `event="face_up_as" phase=""`

### `trigger:opponent_pays_tax` — 1 hits

Examples:

- **Tax Taker** (ability[0] triggered) — `event="opponent_pays_tax" phase=""`

### `trigger:phaseout_or_exile` — 1 hits

Examples:

- **The War Doctor** (ability[0] triggered) — `event="phaseout_or_exile" phase=""`

### `trigger:surveil_first_time` — 1 hits

Examples:

- **Whispering Snitch** (ability[0] triggered) — `event="surveil_first_time" phase=""`

### `trigger:next_time_one_or_more_enter` — 1 hits

Examples:

- **Mystic Reflection** (ability[1] triggered) — `event="next_time_one_or_more_enter" phase=""`

### `trigger:compound_opponents_event` — 1 hits

Examples:

- **Ob Nixilis, Captive Kingpin** (ability[2] triggered) — `event="compound_opponents_event" phase=""`

### `trigger:compound_tribe_die_or_leave` — 1 hits

Examples:

- **Thopter Shop** (ability[0] triggered) — `event="compound_tribe_die_or_leave" phase=""`

### `trigger:self_or_enchantment_etb_or_room_unlock` — 1 hits

Examples:

- **Fear of Sleep Paralysis** (ability[1] triggered) — `event="self_or_enchantment_etb_or_room_unlock" phase=""`

### `trigger:you_sac_one_or_more` — 1 hits

Examples:

- **Camellia, the Seedmiser** (ability[2] triggered) — `event="you_sac_one_or_more" phase=""`

### `trigger:elf_etb` — 1 hits

Examples:

- **Elvish Warmaster** (ability[0] triggered) — `event="elf_etb" phase=""`

### `trigger:damage_to_x_prevented` — 1 hits

Examples:

- **Selfless Squire** (ability[2] triggered) — `event="damage_to_x_prevented" phase=""`

### `trigger:chosen_color_mana_added` — 1 hits

Examples:

- **Caged Sun** (ability[2] triggered) — `event="chosen_color_mana_added" phase=""`

### `trigger:opp_landfall` — 1 hits

Examples:

- **Burgeoning** (ability[0] triggered) — `event="opp_landfall" phase=""`

### `trigger:place_counter` — 1 hits

Examples:

- **Bold Plagiarist** (ability[1] triggered) — `event="place_counter" phase=""`

### `trigger:exiled` — 1 hits

Examples:

- **Kaya, Spirits' Justice** (ability[0] triggered) — `event="exiled" phase=""`

### `trigger:you_control_7_thrulls` — 1 hits

Examples:

- **Endrek Sahr, Master Breeder** (ability[1] triggered) — `event="you_control_7_thrulls" phase=""`

### `trigger:create_token` — 1 hits

Examples:

- **Rosie Cotton of South Lane** (ability[1] triggered) — `event="create_token" phase=""`

### `cond:transform` — 1 hits

Examples:

- **Enduring Angel // Angelic Enforcer** (ability[3] static.condition) — `if your life total would be reduced to 0 or less, instead transform this creature and your life total becomes 3. then if this creature didn't transform this way, you lose the game`

### `trigger:each_player_upkeep` — 1 hits

Examples:

- **Rite of the Raging Storm** (ability[1] triggered) — `event="each_player_upkeep" phase=""`

### `trigger:you_commit_crime` — 1 hits

Examples:

- **Gisa, the Hellraiser** (ability[2] triggered) — `event="you_commit_crime" phase=""`

### `trigger:self_die_or_ally_gy` — 1 hits

Examples:

- **Scrap Trawler** (ability[0] triggered) — `event="self_die_or_ally_gy" phase=""`

### `trigger:nonland_tapped_for_mana` — 1 hits

Examples:

- **Kinnan, Bonder Prodigy** (ability[0] triggered) — `event="nonland_tapped_for_mana" phase=""`

### `trigger:landfall` — 1 hits

Examples:

- **The Necrobloom** (ability[0] triggered) — `event="landfall" phase=""`

### `trigger:becomes_target_by_opp` — 1 hits

Examples:

- **Parnesse, the Subtle Brush** (ability[0] triggered) — `event="becomes_target_by_opp" phase=""`

### `cond:mana-spent` — 1 hits

Examples:

- **Mythos of Illuna** (ability[1] static.condition) — `if {r}{g} was spent to cast this spell, instead create a token that's a copy of that permanent, except the token has "when this token enters, if it's a creature, it fights up to one target creature you don't control."`

### `trigger:upkeep_life_leader` — 1 hits

Examples:

- **Wild Dogs** (ability[0] triggered) — `event="upkeep_life_leader" phase=""`

### `trigger:desert_etb` — 1 hits

Examples:

- **Hazezon, Shaper of Sand** (ability[2] triggered) — `event="desert_etb" phase=""`

### `trigger:compound_tribe_combat_damage` — 1 hits

Examples:

- **Hordewing Skaab** (ability[2] triggered) — `event="compound_tribe_combat_damage" phase=""`

### `trigger:nontoken_type_to_gy` — 1 hits

Examples:

- **Prowess of the Fair** (ability[0] triggered) — `event="nontoken_type_to_gy" phase=""`

### `cond:citys-blessing` — 1 hits

Examples:

- **Radiant Destiny** (ability[3] static.condition) — `as long as you have the city's blessing, they also have vigilance`

### `trigger:creature_modified_event` — 1 hits

Examples:

- **Essence Symbiote** (ability[0] triggered) — `event="creature_modified_event" phase=""`

### `trigger:counter_put_on_self` — 1 hits

Examples:

- **Fathom Mage** (ability[1] triggered) — `event="counter_put_on_self" phase=""`

### `trigger:forest_etb` — 1 hits

Examples:

- **Titania, Nature's Force** (ability[1] triggered) — `event="forest_etb" phase=""`

### `trigger:proliferate` — 1 hits

Examples:

- **Voidwing Hybrid** (ability[2] triggered) — `event="proliferate" phase=""`

### `trigger:untap_step` — 1 hits

Examples:

- **The Millennium Calendar** (ability[0] triggered) — `event="untap_step" phase=""`

### `trigger:self_enter_or_die` — 1 hits

Examples:

- **Undercellar Myconid** (ability[0] triggered) — `event="self_enter_or_die" phase=""`

### `trigger:opp_tokens_event` — 1 hits

Examples:

- **Kambal, Profiteering Mayor** (ability[0] triggered) — `event="opp_tokens_event" phase=""`

### `trigger:first_main` — 1 hits

Examples:

- **Party Thrasher** (ability[1] triggered) — `event="first_main" phase=""`

### `trigger:evolve_event` — 1 hits

Examples:

- **Watchful Radstag** (ability[1] triggered) — `event="evolve_event" phase=""`

### `trigger:self_squad_action` — 1 hits

Examples:

- **Guardian of New Benalia** (ability[1] triggered) — `event="self_squad_action" phase=""`

### `trigger:self_dealt_damage` — 1 hits

Examples:

- **Innocent Bystander** (ability[0] triggered) — `event="self_dealt_damage" phase=""`

### `trigger:investigate` — 1 hits

Examples:

- **Erdwal Illuminator** (ability[1] triggered) — `event="investigate" phase=""`

### `trigger:transform_into_phyrexian` — 1 hits

Examples:

- **Norn's Inquisitor** (ability[1] triggered) — `event="transform_into_phyrexian" phase=""`

### `trigger:artifact_etb_yours` — 1 hits

Examples:

- **Map to Lorthos's Temple** (ability[0] triggered) — `event="artifact_etb_yours" phase=""`

### `trigger:card_milled_via` — 1 hits

Examples:

- **Saruman of Many Colors** (ability[2] triggered) — `event="card_milled_via" phase=""`

### `trigger:play_land` — 1 hits

Examples:

- **City of Traitors** (ability[0] triggered) — `event="play_land" phase=""`

### `trigger:excess_noncombat_damage` — 1 hits

Examples:

- **Toralf, God of Fury // Toralf's Hammer** (ability[1] triggered) — `event="excess_noncombat_damage" phase=""`

### `trigger:transform_as` — 1 hits

Examples:

- **Ludevic, Necrogenius // Olag, Ludevic's Hubris** (ability[4] triggered) — `event="transform_as" phase=""`

### `trigger:becomes_untapped_once` — 1 hits

Examples:

- **Coffin Queen** (ability[2] triggered) — `event="becomes_untapped_once" phase=""`

### `trigger:modified_creature_event` — 1 hits

Examples:

- **Guardian of the Forgotten** (ability[1] triggered) — `event="modified_creature_event" phase=""`

### `trigger:typed_combat_dmg` — 1 hits

Examples:

- **Spawning Kraken** (ability[0] triggered) — `event="typed_combat_dmg" phase=""`

### `trigger:compound_bounce_shuffle_event` — 1 hits

Examples:

- **Tameshi, Reality Architect** (ability[0] triggered) — `event="compound_bounce_shuffle_event" phase=""`

### `cond:color-matters` — 1 hits

Examples:

- **Ria Ivor, Bane of Bladehold** (ability[2] static.condition) — `if damage is prevented this way, create that many 1/1 colorless phyrexian mite artifact creature tokens with toxic 1 and "this token can't block."`

### `trigger:lose_control` — 1 hits

Examples:

- **Khârn the Betrayer** (ability[1] triggered) — `event="lose_control" phase=""`

### `trigger:enchanted_end_step` — 1 hits

Examples:

- **Tenuous Truce** (ability[1] triggered) — `event="enchanted_end_step" phase=""`

### `trigger:put_onto_bf` — 1 hits

Examples:

- **Vivien's Invocation** (ability[3] triggered) — `event="put_onto_bf" phase=""`

### `trigger:damage_to_chosen_player` — 1 hits

Examples:

- **Sower of Discord** (ability[2] triggered) — `event="damage_to_chosen_player" phase=""`

### `trigger:as_transform` — 1 hits

Examples:

- **Curse of Leeches // Leeching Lurker** (ability[1] triggered) — `event="as_transform" phase=""`

### `trigger:tap_for_c` — 1 hits

Examples:

- **Forsaken Monument** (ability[1] triggered) — `event="tap_for_C" phase=""`

### `trigger:three_or_more` — 1 hits

Examples:

- **Inniaz, the Gale Force** (ability[2] triggered) — `event="three_or_more" phase=""`

### `trigger:combat_damage_planeswalker` — 1 hits

Examples:

- **Zagras, Thief of Heartbeats** (ability[5] triggered) — `event="combat_damage_planeswalker" phase=""`

### `trigger:self_to_gy` — 1 hits

Examples:

- **Enigma Sphinx** (ability[1] triggered) — `event="self_to_gy" phase=""`

### `trigger:becomes_saddled_first` — 1 hits

Examples:

- **Stubborn Burrowfiend** (ability[0] triggered) — `event="becomes_saddled_first" phase=""`

### `trigger:on_card_advantage` — 1 hits

Examples:

- **Toofer, Keeper of the Full Grip** (ability[0] triggered) — `event="on_card_advantage" phase=""`

### `trigger:ninja_combat_damage` — 1 hits

Examples:

- **Yuriko, the Tiger's Shadow** (ability[1] triggered) — `event="ninja_combat_damage" phase=""`

### `trigger:tempting_offer` — 1 hits

Examples:

- **Tempt with Glory** (ability[1] triggered) — `event="tempting_offer" phase=""`

### `trigger:train` — 1 hits

Examples:

- **Savior of Ollenbock** (ability[1] triggered) — `event="train" phase=""`

### `cond:numeric-threshold-other` — 1 hits

Examples:

- **Jugan Defends the Temple // Remnant of the Rising Star** (ability[6] static.condition) — `as long as you control five or more modified creatures, this creature gets +5/+5 and has trample`

### `trigger:you_put_counter_on` — 1 hits

Examples:

- **Sigurd, Jarl of Ravensthorpe** (ability[4] triggered) — `event="you_put_counter_on" phase=""`

### `trigger:you_create_one_or_more_tokens` — 1 hits

Examples:

- **Akim, the Soaring Wind** (ability[1] triggered) — `event="you_create_one_or_more_tokens" phase=""`

### `trigger:self_and_or_others_event` — 1 hits

Examples:

- **Satoru, the Infiltrator** (ability[1] triggered) — `event="self_and_or_others_event" phase=""`

### `trigger:chosen_color_mana_tapped` — 1 hits

Examples:

- **Gauntlet of Power** (ability[2] triggered) — `event="chosen_color_mana_tapped" phase=""`

### `trigger:foretell_card` — 1 hits

Examples:

- **Dream Devourer** (ability[2] triggered) — `event="foretell_card" phase=""`

### `trigger:merfolk_etb_any` — 1 hits

Examples:

- **Map to Lorthos's Temple** (ability[1] triggered) — `event="merfolk_etb_any" phase=""`

### `trigger:counter_threshold_reached` — 1 hits

Examples:

- **Midnight Clock** (ability[3] triggered) — `event="counter_threshold_reached" phase=""`

### `trigger:activation_non_mana` — 1 hits

Examples:

- **Flamescroll Celebrant // Revel in Silence** (ability[0] triggered) — `event="activation_non_mana" phase=""`

### `trigger:self_or_another_when` — 1 hits

Examples:

- **Tiana, Angelic Mechanic** (ability[1] triggered) — `event="self_or_another_when" phase=""`

### `trigger:combat_damage_to_you` — 1 hits

Examples:

- **Risona, Asari Commander** (ability[2] triggered) — `event="combat_damage_to_you" phase=""`

## Highest-leverage clusters — proposed scaffold kinds

Sorted by hits, the top clusters that would cover the most cards if
added as new `conditionScaffoldKind` entries:

| Rank | Cluster | Hits | Suggested scaffold name |
|---:|---|---:|---|
| 1 | `trigger:die` | 454 | `trigger_die` |
| 2 | `trigger:combat_damage_player` | 382 | `trigger_combat_damage_player` |
| 3 | `trigger:phase` | 257 | `trigger_phase` |
| 4 | `trigger:when_you_do` | 177 | `trigger_when_you_do` |
| 5 | `trigger:etb_as` | 131 | `trigger_etb_as` |
| 6 | `cond:other` | 98 | `condScaffoldOther` |
| 7 | `trigger:turned_face_up` | 73 | `trigger_turned_face_up` |
| 8 | `trigger:beginning_of_ordinal_step` | 69 | `trigger_beginning_of_ordinal_step` |
| 9 | `trigger:you_whenever` | 67 | `trigger_you_whenever` |
| 10 | `trigger:self_and_another` | 63 | `trigger_self_and_another` |
| 11 | `trigger:self_and` | 46 | `trigger_self_and` |
| 12 | `trigger:tribe_you_control_etb` | 45 | `trigger_tribe_you_control_etb` |

## Next steps

The top three clusters (`trigger:die`, `trigger:combat_damage_player`, `trigger:phase`) account for 1093 unmatched abilities (33.3% of all flagged nodes). Adding scaffolds for these three patterns would deliver the most coverage per scaffold added. The biggest single jump
comes from the `STRUCTURED:` condition kinds (you_control, life_threshold,
tribal, etc.) — `detectConditionScaffold` deliberately ignores these because
they have non-raw AST forms; bridging them into the priming registry would
convert `STRUCTURED:*` rows into real scaffold coverage and immediately fix
the largest bucket of "no priming applied" Goldilocks failures in Era 1.
