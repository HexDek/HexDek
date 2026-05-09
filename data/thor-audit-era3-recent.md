# Thor 2.1 Corpus Audit — Era 3 (Recent 2025+)

- Cutoff: cards with `released_at >= 2025-01-01`
- Unique cards in window: **4922**
- Cards with conditional clauses parsed: **684**
- Cards with triggers but no `if`/`as long as` clause: 2142
- Total condition clauses examined: **753**
  - Bucketed: **179** (23.8%)
  - Flagged (no scaffold match): **574** (76.2%)

## Proposed new scaffold kinds (synthesis)

Distilled from the multi-card flagged clusters below. Each row maps a
working name to the cluster keys carrying its clauses. Use this as the
starting point for new `condScaffold*` constants in `conditional_setup.go`.

| Proposed kind | Description | Source clusters | Total clauses |
|---|---|---|---:|
| **SpellKicked** | trigger/condition: kicker was paid | `spell kicked`, `kicked`, `spell's additional cost paid` | 21 |
| **ManaSpentThreshold** | N+ mana spent to cast / mana-spent comparison | `five more mana spent`, `at least four mana`, `amount mana spent greater`, `amount mana spent cast`, `spell's power greater`, `r r spent cast`, `g g spent cast` | 32 |
| **PriorTurnSpellCount** | 0 / 2+ spells cast last turn (Werewolf transform) | `no spells cast last`, `player cast two more` | 20 |
| **CounterReplacementBoost** | +1/+1 counter doubler replacement | `one more counters put`, `one more counters` | 11 |
| **TokenReplacementBoost** | token doubler replacement | `one more tokens created` | 5 |
| **DiscardedNonland** | discarded a nonland card (Spider-Verse mechanic) | `discarded nonland card` | 9 |
| **ControlNTappedCreatures** | tapped-creature count threshold (EoE Pilots) | `control two more tapped` | 8 |
| **PersistUndyingCheck** | no -1/-1 / no +1/+1 counters at death | `no counters` | 6 |
| **WouldDieExileReplacement** | creature would die → exile replacement | `creature die turn` | 6 |
| **ControlNCreatures** | creatures-you-control count (with-different-powers variant) | `control three more creatures`, `two more those creatures` | 7 |
| **CountersOnAtMove** | permanent had counters on it (counter relocation) | `counters`, `counter`, `creature counter` | 11 |
| **ControlNLands** | land-count threshold (Domain / Land Matters) | `control eight more lands`, `control seven more lands` | 7 |
| **DragonBeheld** | Tarkir Dragonstorm 'a Dragon was beheld' check | `dragon beheld` | 4 |
| **CardTypeReveal** | revealed/exiled card type check (land/permanent/creature) | `it's land card`, `it's permanent card`, `it's creature card`, `it's vampire`, `it's not creature`, `isn't creature` | 15 |
| **FirstCombatPhase** | phase: first combat phase | `it's first combat phase` | 4 |
| **MainPhase** | phase: main phase | `it's your main phase` | 2 |
| **EquipmentAttached** | equipment attached / equipped check | `equipment attached creature`, `cloud equipped`, `equipped creature human` | 6 |
| **ControlArtifact** | you control an artifact (subtype-style for type) | `control artifact` | 4 |
| **DrawnNCardsThisTurn** | drawn N+ cards this turn (variant of DrawnCardThisTurn with count) | `you've drawn two more` | 3 |
| **CastNSpellsThisTurn** | cast N+ spells this turn (variant of CastSpellThisTurn with count) | `you've cast two more` | 3 |
| **TotalToughness** | Formidable variant: total toughness N+ | `creatures control total toughness` | 3 |
| **QuestCounters** | Ascend-style 4+ quest counters / lore counters | `four more quest counters` | 3 |
| **FullParty** | Zendikar Rising 'full party' (4 distinct classes) | `full party` | 4 |
| **StartingPlayer** | you-were/weren't-the-starting-player check | `starting player` | 2 |
| **PutCounterThisTurn** | you put a counter on a creature this turn | `put counter creature turn` | 3 |
| **CreatureCardExiled** | a creature card was exiled this way (Surveil/Shadow templates) | `creature card exiled way` | 3 |

## Bucketed breakdown — per scaffold kind

| Scaffold kind | Clauses | Sample cards |
|---|---:|---|
| `CardInGraveyard` | 52 | Aang, A Lot to Learn, Anger, Celes, Rune Knight, Cephalid Coliseum, Covetous Castaway // Ghostly Castigator |
| `YouControlSubtype` | 27 | Apothecary Geist, Bleachbone Verge, Brave Falconhawk, Brood Astronomer, Doc Ock, Sinister Scientist |
| `Revolt` | 20 | Alpharael, Stonechosen, Axavar, Fate Thief, Decode Transmissions, Elegy Acolyte, Foot Mystic |
| `GainedLifeThisTurn` | 15 | Aerith, Last Ancient, Bre of Clan Stoutarm, Eccentric Pestfinder // Turn Stones, Efflorescence, Follow the Lumarets |
| `CastSpellThisTurn` | 11 | Conduit of Worlds, Hall of Oracles, Human Torch, Invisible Woman, Mister Fantastic |
| `AttackedThisTurn` | 8 | Baron Helmut Zemo, Deepway Navigator, Fire Nation Engineer, Fire Nation Raider, Michelangelo, the Heart |
| `OpponentMoreLands` | 7 | Archaeomancer's Map, Claim Jumper, Emeritus of Truce // Swords to Plowshares, Land Tax, Sunstar Expansionist |
| `CreatureDiedThisTurn` | 5 | Emeritus of Woe // Demonic Tutor, Morkrut Banshee, Perennial Gravewarden, Scorpion, Seething Striker, Skirsdag High Priest |
| `OpponentLostLife` | 4 | Bloodtithe Collector, Gomif, Fast Racer, Lion Vulture, Voldaren Ambusher |
| `CombatDamageDealt` | 4 | Blitzball, Falko, Showoff Pilot, Lost Monarch of Ifnir, Sidequest: Play Blitzball // World Champion, Celestial Weapon |
| `CastFromExile` | 3 | Lifestream's Blessing, Ultimate Magic: Holy, Ultimate Magic: Meteor |
| `Ferocious` | 3 | Garruk's Uprising, Master's Guidance, Raucous Audience |
| `CreatureETBThisTurn` | 3 | Bristlebane Outrider, Thoughtweft Charge, Wary Farmer |
| `CreatureCardsInGraveyard` | 3 | Grave Researcher // Reanimate, Grizzled Angler // Grisly Anglerfish, Lorehold Archivist // Restore Relic |
| `Monarch` | 2 | Garland, Royal Kidnapper, Grave Venerations |
| `LandfallThisTurn` | 2 | Earth Rumble Wrestlers, Zimone, All-Questioning |
| `Hellbent` | 2 | Asylum Visitor, Bloodhall Priest |
| `SacrificedThisTurn` | 2 | Evendo Brushrazer, Phoenix Fleet Airship |
| `Metalcraft` | 1 | Dispatch |
| `SpellMastery` | 1 | Animist's Awakening |
| `UpkeepPhase` | 1 | Karmic Guide |
| `Delirium` | 1 | Traverse the Ulvenwald |
| `Formidable` | 1 | Surrak, the Hunt Caller |
| `LifeBelowThreshold` | 1 | Gather the Townsfolk |

## Flagged clusters — proposed new scaffold kinds

### Multi-card clusters (73 clusters, 276 clauses)

#### `spell kicked`  — 11 clauses
- **Stomped by the Foot** (tmt, 2026-03-06): _this spell was kicked_
- **Jet's Brainwashing** (tla, 2025-11-21): _this spell was kicked_
- **Tear Asunder** (eoc, 2025-08-01): _this spell was kicked_
- **Whoosh!** (spm, 2025-09-26): _this spell was kicked_
- **Vayne's Treachery** (fin, 2025-06-13): _this spell was kicked_
- **Firebending Lesson** (tla, 2025-11-21): _this spell was kicked_
- **Chocobo Kick** (fin, 2025-06-13): _this spell was kicked_
- **Aang's Journey** (tla, 2025-11-21): _this spell was kicked_
- _…and 3 more_

#### `no spells cast last`  — 10 clauses
- **Hinterland Logger // Timber Shredder** (inr, 2025-01-24): _no spells were cast last turn_
- **Scorned Villager // Moonscarred Werewolf** (inr, 2025-01-24): _no spells were cast last turn_
- **Huntmaster of the Fells // Ravager of the Fells** (inr, 2025-01-24): _no spells were cast last turn_
- **Geier Reach Bandit // Vildin-Pack Alpha** (inr, 2025-01-24): _no spells were cast last turn_
- **Mayor of Avabruck // Howlpack Alpha** (inr, 2025-01-24): _no spells were cast last turn_
- **Duskwatch Recruiter // Krallenhorde Howler** (inr, 2025-01-24): _no spells were cast last turn_
- **Hanweir Watchkeep // Bane of Hanweir** (inr, 2025-01-24): _no spells were cast last turn_
- **Kruin Outlaw // Terror of Kruin Pass** (inr, 2025-01-24): _no spells were cast last turn_
- _…and 2 more_

#### `player cast two more`  — 10 clauses
- **Hinterland Logger // Timber Shredder** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Scorned Villager // Moonscarred Werewolf** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Huntmaster of the Fells // Ravager of the Fells** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Geier Reach Bandit // Vildin-Pack Alpha** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Mayor of Avabruck // Howlpack Alpha** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Duskwatch Recruiter // Krallenhorde Howler** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Hanweir Watchkeep // Bane of Hanweir** (inr, 2025-01-24): _a player cast two or more spells last turn_
- **Kruin Outlaw // Terror of Kruin Pass** (inr, 2025-01-24): _a player cast two or more spells last turn_
- _…and 2 more_

#### `amount mana spent greater`  — 9 clauses
- **Cuboid Colony** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Pensive Professor** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Topiary Lecturer** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Berta, Wise Extrapolator** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Hungry Graffalon** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Tester of the Tangential** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Fractal Tender** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- **Ambitious Augmenter** (sos, 2026-04-24): _the amount of mana you spent is greater than this creature's power or toughness_
- _…and 1 more_

#### `five more mana spent`  — 9 clauses
- **Tackle Artist** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Deluge Virtuoso** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Thunderdrum Soloist** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Colorstorm Stallion** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Molten-Core Maestro** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Spectacular Skywhale** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Elemental Mascot** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- **Exhibition Tidecaller** (sos, 2026-04-24): _five or more mana was spent to cast that spell_
- _…and 1 more_

#### `discarded nonland card`  — 9 clauses
- **Mob Lookout** (spm, 2025-09-26): _you discarded a nonland card_
- **Unstable Experiment** (spm, 2025-09-26): _you discarded a nonland card_
- **Prowler, Clawed Thief** (spm, 2025-09-26): _you discarded a nonland card_
- **Doctor Doom, King of Latveria** (msc, 2026-06-26): _you discarded a nonland card_
- **Norman Osborn // Green Goblin** (spm, 2025-09-26): _you discarded a nonland card_
- **Doc Ock's Henchmen** (spm, 2025-09-26): _you discarded a nonland card_
- **Mechanical Mobster** (spm, 2025-09-26): _you discarded a nonland card_
- **Scorpion, Seething Striker** (spm, 2025-09-26): _you discarded a nonland card_
- _…and 1 more_

#### `cast`  — 8 clauses
- **Weftwalking** (eoe, 2025-08-01): _you cast it_
- **Yathan Roadwatcher** (tdm, 2025-04-11): _you cast it_
- **North Wind Avatar** (tmt, 2026-03-06): _you cast it_
- **The Sibsig Ceremony** (tdm, 2025-04-11): _you cast it_
- **Sunderflock** (ecl, 2026-01-23): _you cast it_
- **Iridescent Tiger** (tdm, 2025-04-11): _you cast it_
- **Transcendent Dragon** (tdc, 2025-04-11): _you cast it_
- **Prototype X-8** (yeoe, 2025-08-19): _you cast it_

#### `one more counters put`  — 8 clauses
- **Loading Zone** (eoe, 2025-08-01): _one or more counters would be put on a creature_
- **Michelangelo, Weirdness to 11** (tmt, 2026-03-06): _one or more +1/+1 counters would be put on a creature you control_
- **Ozolith, the Shattered Spire** (soc, 2026-04-24): _one or more +1/+1 counters would be put on an artifact or creature you control_
- **Solid Ground** (tle, 2025-11-21): _one or more +1/+1 counters would be put on a permanent you control_
- **Hardened Scales** (soc, 2026-04-24): _one or more +1/+1 counters would be put on a creature you control_
- **Caradora, Heart of Alacria** (dft, 2025-02-14): _one or more +1/+1 counters would be put on a creature or Vehicle you control_
- **High Score** (tmc, 2026-03-06): _one or more +1/+1 counters would be put on a creature you control_
- **Kami of Whispered Hopes** (soc, 2026-04-24): _one or more +1/+1 counters would be put on a permanent you control_

#### `control two more tapped`  — 8 clauses
- **Current Curriculum** (yecl, 2026-02-03): _you control two or more tapped creatures_
- **Thendar, the Overminer** (yeoe, 2025-08-19): _you control two or more tapped creatures_
- **Sami, Ship's Engineer** (eoe, 2025-08-01): _you control two or more tapped creatures_
- **Flight-Deck Coordinator** (eoe, 2025-08-01): _you control two or more tapped creatures_
- **Frontline War-Rager** (eoe, 2025-08-01): _you control two or more tapped creatures_
- **Dawnstrike Vanguard** (eoe, 2025-08-01): _you control two or more tapped creatures_
- **Sunstar Chaplain** (eoe, 2025-08-01): _you control two or more tapped creatures_
- **Vaultguard Trooper** (eoe, 2025-08-01): _you control two or more tapped creatures_

#### `spell's additional cost paid`  — 7 clauses
- **Pyrrhic Strike** (ecl, 2026-01-23): _this spell's additional cost was paid_
- **Spirit Water Revival** (tla, 2025-11-21): _this spell's additional cost was paid_
- **Secret of Bloodbending** (tla, 2025-11-21): _this spell's additional cost was paid_
- **Ruinous Waterbending** (tla, 2025-11-21): _this spell's additional cost was paid_
- **Burning Curiosity** (ecl, 2026-01-23): _this spell's additional cost was paid_
- **Celestial Reunion** (ecl, 2026-01-23): _this spell's additional cost was paid and the revealed card is the chosen type_
- **Requiting Hex** (ecl, 2026-01-23): _this spell's additional cost was paid_

#### `no counters`  — 6 clauses
- **Puppeteer Clique** (ecc, 2026-01-23): _it had no -1/-1 counters on it_
- **Butcher Ghoul** (inr, 2025-01-24): _it had no +1/+1 counters on it_
- **Eirdu, Carrier of Dawn // Isilu, Carrier of Twilight** (ecl, 2026-01-23): _it had no -1/-1 counters on it_
- **Young Wolf** (inr, 2025-01-24): _it had no +1/+1 counters on it_
- **Rhys, the Evermore** (ecl, 2026-01-23): _it had no -1/-1 counters on it_
- **River Kelpie** (tdc, 2025-04-11): _it had no -1/-1 counters on it_

#### `at least four mana`  — 6 clauses
- **The Emperor of Palamecia // The Lord Master of Hell** (fin, 2025-06-13): _at least four mana was spent to cast it_
- **Blazing Bomb** (fin, 2025-06-13): _at least four mana was spent to cast it_
- **Sahagin** (fin, 2025-06-13): _at least four mana was spent to cast it_
- **The Prima Vista** (fin, 2025-06-13): _at least four mana was spent to cast it_
- **Ultros, Obnoxious Octopus** (fin, 2025-06-13): _at least four mana was spent to cast it_
- **Prompto Argentum** (fin, 2025-06-13): _at least four mana was spent to cast it_

#### `creature die turn`  — 6 clauses
- **Bot Bashing Time** (tmt, 2026-03-06): _that creature would die this turn_
- **Narset's Rebuke** (tdm, 2025-04-11): _that creature would die this turn_
- **Suplex** (fin, 2025-06-13): _that creature would die this turn_
- **Wilt in the Heat** (sos, 2026-04-24): _that creature would die this turn_
- **Combustion Technique** (tla, 2025-11-21): _that creature would die this turn_
- **Feed the Flames** (ecl, 2026-01-23): _that creature would die this turn_

#### `search your library way`  — 6 clauses
- **Claim Jumper** (soc, 2026-04-24): _you search your library this way_
- **Fang-Druid Summoner** (dft, 2025-02-14): _you search your library this way_
- **Delivery Moogle** (fin, 2025-06-13): _you search your library this way_
- **Tale of Momo** (tle, 2025-11-21): _you search your library this way_
- **Unlucky Cabbage Merchant** (tla, 2025-11-21): _you search your library this way_
- **Guidelight Pathmaker** (dft, 2025-02-14): _you search your library this way_

#### `one more tokens created`  — 5 clauses
- **Donatello, the Brains** (tmc, 2026-03-06): _one or more tokens would be created under your control_
- **Quina, Qu Gourmet** (fin, 2025-06-13): _one or more tokens would be created under your control_
- **Exalted Sunborn** (eoe, 2025-08-01): _one or more tokens would be created under your control_
- **Kaya, Geist Hunter** (tdc, 2025-04-11): _one or more tokens would be created under your control_
- **Elspeth, Storm Slayer** (tdm, 2025-04-11): _one or more tokens would be created under your control_

#### `control three more creatures`  — 5 clauses
- **Dundoolin Weaver** (ecl, 2026-01-23): _you control three or more creatures_
- **Duel for Dominance** (inr, 2025-01-24): _you control three or more creatures with different powers_
- **You're Not Alone** (fin, 2025-06-13): _you control three or more creatures_
- **Augur of Autumn** (eoc, 2025-08-01): _you control three or more creatures with different powers_
- **Ambitious Farmhand // Seasoned Cathar** (inr, 2025-01-24): _you control three or more creatures with different powers_

#### `counters`  — 5 clauses
- **Buzzard-Wasp Colony** (tla, 2025-11-21): _it had counters on it_
- **Host of the Hereafter** (tdm, 2025-04-11): _it had counters on it_
- **Scolding Administrator** (sos, 2026-04-24): _it had counters on it_
- **Resourceful Defense** (eoc, 2025-08-01): _it had counters on it_
- **Donatello, Mutant Mechanic** (tmt, 2026-03-06): _it had counters on it_

#### `x more`  — 5 clauses
- **Shantotto, Tactician Magician** (fin, 2025-06-13): _X is 4 or more_
- **Zero Point Ballad** (eoe, 2025-08-01): _X is 6 or more_
- **Kinetic Ooze** (soc, 2026-04-24): _X is 5 or more_
- **Kinetic Ooze** (soc, 2026-04-24): _X is 10 or more_
- **Roku's Mastery** (tle, 2025-11-21): _X is 4 or more_

#### `remains exiled`  — 4 clauses
- **The Windy City** (punk, 2025-02-21): _it remains exiled_
- **The Pro Tour** (punk, 2025-02-21): _it remains exiled_
- **Lightstall Inquisitor** (eoe, 2025-08-01): _it remains exiled_
- **Gonti, Night Minister** (dft, 2025-02-14): _it remains exiled_

#### `it's first combat phase`  — 4 clauses
- **Tifa, Martial Artist** (fic, 2025-06-13): _it's the first combat phase of your turn_
- **Balthier and Fran** (fin, 2025-06-13): _it's the first combat phase of the turn_
- **Raph & Leo, Sibling Rivals** (tmt, 2026-03-06): _it's the first combat phase of the turn_
- **Genji Glove** (fin, 2025-06-13): _it's the first combat phase of the turn_

#### `counter`  — 4 clauses
- **Retched Wretch** (ecl, 2026-01-23): _it had a -1/-1 counter on it_
- **Leader's Talent** (tmt, 2026-03-06): _it had a counter on it_
- **Blowfly Infestation** (ecc, 2026-01-23): _it had a -1/-1 counter on it_
- **Dual-Sun Technique** (eoe, 2025-08-01): _it has a +1/+1 counter on it_

#### `cast spell way`  — 4 clauses
- **Noctis, Prince of Lucis** (fin, 2025-06-13): _you cast a spell this way_
- **Leonardo, Sewer Samurai** (tmt, 2026-03-06): _you cast a spell this way_
- **Edgar, Master Machinist** (fic, 2025-06-13): _you cast a spell this way_
- **Gwenom, Remorseless** (spm, 2025-09-26): _you cast a spell this way_

#### `cast creature spell way`  — 4 clauses
- **Strago and Relm** (fic, 2025-06-13): _you cast a creature spell this way_
- **Mikey & Don, Party Planners** (tmt, 2026-03-06): _you cast a creature spell this way_
- **Thundermane Dragon** (tdc, 2025-04-11): _you cast a creature spell this way_
- **The Mysterious Sphere** (unk, 2025-06-20): _you cast a creature spell this way_

#### `full party`  — 4 clauses
- **The Destined Warrior** (fic, 2025-12-05): _you have a full party_
- **The Destined Thief** (fic, 2025-12-05): _you have a full party_
- **The Destined White Mage** (fic, 2025-12-05): _you have a full party_
- **The Destined Black Mage** (fic, 2025-12-05): _you have a full party_

#### `control eight more lands`  — 4 clauses
- **Kyoshi Warrior Exemplars** (tle, 2025-11-21): _you control eight or more lands_
- **Emeritus of Abundance // Regrowth** (sos, 2026-04-24): _you control eight or more lands_
- **Zimone, Quandrix Prodigy** (soc, 2026-04-24): _you control eight or more lands_
- **Omnath, Locus of the Roil** (ecc, 2026-01-23): _you control eight or more lands_

#### `control artifact`  — 4 clauses
- **Spire of Industry** (eoc, 2025-08-01): _you control an artifact_
- **Gravblade Heavy** (eoe, 2025-08-01): _you control an artifact_
- **Cloudsculpt Technician** (eoe, 2025-08-01): _you control an artifact_
- **Donatello, Turtle Techie** (tmt, 2026-03-06): _you control an artifact_

#### `dragon beheld`  — 4 clauses
- **Osseous Exhale** (tdm, 2025-04-11): _a Dragon was beheld_
- **Draconic Fealty** (ytdm, 2025-04-29): _a Dragon was beheld_
- **Dispelling Exhale** (tdm, 2025-04-11): _a Dragon was beheld_
- **Piercing Exhale** (tdm, 2025-04-11): _a Dragon was beheld_

#### `control creature`  — 3 clauses
- **Taster of Wares** (ecl, 2026-01-23): _you control this creature_
- **Would You Have Done the Same?** (unk, 2025-02-21): _you control a creature_
- **Gwen Stacy // Ghost-Spider** (spm, 2025-09-26): _you control this creature_

#### `it's permanent card`  — 3 clauses
- **Chaos Warp** (soc, 2026-04-24): _it's a permanent card_
- **Fishing Gear** (fic, 2025-12-05): _it's a permanent card_
- **Esper Origins // Summon: Esper Maduin** (fin, 2025-06-13): _it's a permanent card_

#### `it's land card`  — 3 clauses
- **Traveling Botanist** (tdm, 2025-04-11): _it's a land card_
- **Risen Reef** (ecc, 2026-01-23): _it's a land card_
- **Currency Converter** (soc, 2026-04-24): _it's a land card_

#### `you've drawn two more`  — 3 clauses
- **June, Bounty Hunter** (tla, 2025-11-21): _you've drawn two or more cards this turn_
- **Foggy Swamp Hunters** (tla, 2025-11-21): _you've drawn two or more cards this turn_
- **Messenger Hawk** (tla, 2025-11-21): _you've drawn two or more cards this turn_

#### `gained more life turn`  — 3 clauses
- **Indulging Patrician** (tdc, 2025-04-11): _you gained 3 or more life this turn_
- **Scheming Silvertongue // Sign in Blood** (sos, 2026-04-24): _you gained 2 or more life this turn_
- **Aerith, Last Ancient** (fic, 2025-06-13): _you gained 7 or more life this turn_

#### `toughness less`  — 3 clauses
- **Iron-Shield Elf** (ecl, 2026-01-23): _its toughness is 0 or less_
- **Gilt-Leaf's Embrace** (ecl, 2026-01-23): _its toughness is 0 or less_
- **Burdened Stoneback** (ecl, 2026-01-23): _its toughness is 0 or less_

#### `you've cast two more`  — 3 clauses
- **Reverberating Summons** (tdm, 2025-04-11): _you've cast two or more spells this turn_
- **Brightspear Zealot** (eoe, 2025-08-01): _you've cast two or more spells this turn_
- **Lyse Hext** (fic, 2025-06-13): _you've cast two or more noncreature spells this turn_

#### `creature card exiled way`  — 3 clauses
- **Ignis Scientia** (fin, 2025-06-13): _a creature card was exiled this way_
- **Feral Appetite** (soc, 2026-04-24): _a creature card is exiled this way_
- **Raven Eagle** (tla, 2025-11-21): _a creature card is exiled this way_

#### `put counter creature turn`  — 3 clauses
- **Lord Jyscal Guado** (fic, 2025-06-13): _you put a counter on a creature this turn_
- **Fractal Tender** (sos, 2026-04-24): _you put a counter on this creature this turn_
- **Lasting Tarfire** (ecl, 2026-01-23): _you put a counter on a creature this turn_

#### `control seven more lands`  — 3 clauses
- **Scorpion Sentinel** (fin, 2025-06-13): _you control seven or more lands_
- **Tend the Sprigs** (ecl, 2026-01-23): _you control seven or more lands and/or Treefolk_
- **Gigantoad** (fin, 2025-06-13): _you control seven or more lands_

#### `one more counters`  — 3 clauses
- **Oft-Nabbed Goat** (ecc, 2026-01-23): _it had one or more -1/-1 counters on it_
- **Yuna, Grand Summoner** (pw25, 2025-12-05): _it had one or more counters on it_
- **Ambitious Augmenter** (sos, 2026-04-24): _it had one or more counters on it_

#### `kicked`  — 3 clauses
- **Gatekeeper of Malakir** (fdn, 2026-04-24): _it was kicked_
- **Sprouting Goblin** (eoc, 2025-08-01): _it was kicked_
- **Verix Bladewing** (tdc, 2025-04-11): _it was kicked_

#### `creatures control total toughness`  — 3 clauses
- **Betor, Kin to All** (tdm, 2025-04-11): _creatures you control have total toughness 10 or greater_
- **Betor, Kin to All** (tdm, 2025-04-11): _creatures you control have total toughness 20 or greater_
- **Betor, Kin to All** (tdm, 2025-04-11): _creatures you control have total toughness 40 or greater_

#### `four more quest counters`  — 3 clauses
- **Firebender Ascension** (tla, 2025-11-21): _it has four or more quest counters on it_
- **Waterbender Ascension** (tla, 2025-11-21): _it has four or more quest counters on it_
- **Earthbender Ascension** (tla, 2025-11-21): _it has four or more quest counters on it_

#### `isn't creature`  — 3 clauses
- **Does Machines** (tmt, 2026-03-06): _it isn't a creature_
- **Pizza Face, Gastromancer** (tmt, 2026-03-06): _it isn't a creature_
- **Donatello, Mutant Mechanic** (tmt, 2026-03-06): _it isn't a creature_

#### `creature counter`  — 2 clauses
- **Demon Wall** (fin, 2025-06-13): _this creature has a counter on it_
- **Skarrgan Hellkite** (tdc, 2025-04-11): _this creature has a +1/+1 counter on it_

#### `it's not creature`  — 2 clauses
- **Tezzeret, Cruel Captain** (eoe, 2025-08-01): _it's not a creature_
- **Tezzeret, Cruel Captain Emblem** (teoe, 2025-08-01): _it's not a creature_

#### `it's vampire`  — 2 clauses
- **A-Sorin, Imperious Bloodlord** (m20, 2025-08-18): _it's a Vampire_
- **Sorin, Imperious Bloodlord** (inr, 2025-01-24): _it's a Vampire_

#### `amount mana spent cast`  — 2 clauses
- **Unravel** (eoe, 2025-08-01): _the amount of mana spent to cast that spell was less than its mana value_
- **Tokka & Rahzar, Terrible Twos** (tmt, 2026-03-06): _the amount of mana spent to cast it was less than its mana value_

#### `spell's power greater`  — 2 clauses
- **Eshki, Temur's Roar** (tdc, 2025-04-11): _that spell's power is 4 or greater_
- **Eshki, Temur's Roar** (tdc, 2025-04-11): _that spell's power is 6 or greater_

#### `it's attacking`  — 2 clauses
- **Blood Petal Celebrant** (inr, 2025-01-24): _it's attacking_
- **Kitesail Corsair** (fdn, 2026-04-24): _it's attacking_

#### `creature died under your`  — 2 clauses
- **Barrensteppe Siege** (tdm, 2025-04-11): _a creature died under your control this turn_
- **Essenceknit Scholar** (sos, 2026-04-24): _a creature died under your control this turn_

#### `starting player`  — 2 clauses
- **Rising Chicane** (ydft, 2025-03-04): _you were the starting player_
- **Desert Cenote** (ytdm, 2025-04-29): _you were the starting player_

#### `equipment attached creature`  — 2 clauses
- **Summoning Materia** (fic, 2025-06-13): _this Equipment is attached to a creature_
- **Mirrormind Crown** (ecl, 2026-01-23): _this Equipment is attached to a creature_

#### `two more those creatures`  — 2 clauses
- **Tomik, Wielder of Law** (soc, 2026-04-24): _two or more of those creatures are attacking you and/or planeswalkers you control_
- **Mangara, the Diplomat** (soc, 2026-04-24): _two or more of those creatures are attacking you and/or planeswalkers you control_

#### `spell countered way`  — 2 clauses
- **Syncopate** (fin, 2025-06-13): _that spell is countered this way_
- **Transcendent Dragon** (tdc, 2025-04-11): _that spell is countered this way_

#### `cloud equipped`  — 2 clauses
- **Cloud, Midgar Mercenary** (fin, 2025-06-13): _Cloud is equipped_
- **Cloud, Planet's Champion** (fin, 2025-06-13): _Cloud is equipped_

#### `it's creature card`  — 2 clauses
- **Bison Whistle** (tle, 2025-11-21): _it's a creature card_
- **Reality Shift** (soc, 2026-04-24): _it's a creature card_

#### `r r spent cast`  — 2 clauses
- **Vibrance** (ecl, 2026-01-23): _{R}{R} was spent to cast it_
- **Catharsis** (ecl, 2026-01-23): _{R}{R} was spent to cast it_

#### `g g spent cast`  — 2 clauses
- **Vibrance** (ecl, 2026-01-23): _{G}{G} was spent to cast it_
- **Wistfulness** (ecl, 2026-01-23): _{G}{G} was spent to cast it_

#### `equipped creature human`  — 2 clauses
- **Butcher's Cleaver** (inr, 2025-01-24): _equipped creature is a Human_
- **Harvest Hand // Scrounged Scythe** (inr, 2025-01-24): _equipped creature is a Human_

#### `mana value less`  — 2 clauses
- **Tainted Treats** (tmt, 2026-03-06): _its mana value was 4 or less_
- **Seedship Impact** (eoe, 2025-08-01): _its mana value was 2 or less_

#### `it's your main phase`  — 2 clauses
- **Moraug, Fury of Akoum** (eoc, 2025-08-01): _it's your main phase_
- **All-Out Assault** (tdm, 2025-04-11): _it's your main phase_

### Singleton flags (298 clauses) — possible noise / one-off mechanics

- **Voice of the Blessed** (inr, 2025-01-24): _this creature has four or more +1/+1 counters on it_
- **Voice of the Blessed** (inr, 2025-01-24): _this creature has ten or more +1/+1 counters on it_
- **Taster of Wares** (ecl, 2026-01-23): _an instant or sorcery card is exiled this way_
- **War Balloon** (tla, 2025-11-21): _this Vehicle has three or more fire counters on it_
- **Mindblade Render** (tdc, 2025-04-11): _any of that damage was dealt by a Warrior_
- **Meren of Clan Nel Toth** (tdc, 2025-04-11): _that card's mana value is less than or equal to the number of experience counters you have_
- **The Emperor of Palamecia // The Lord Master of Hell** (fin, 2025-06-13): _it has three or more +1/+1 counters on it_
- **Efteekay, Flame of the Kav** (unk, 2025-08-02): _Efteekay is in the command zone or on the battlefield_
- **Vraska, Betrayal's Sting** (ecc, 2026-01-23): _life was paid_
- **Goatnap** (ecl, 2026-01-23): _that creature is a Goat_
- **Scourge of the Throne** (tdc, 2025-04-11): _it's attacking the player with the most life or tied for most life_
- **Sab-Sunen, Luxa Embodied** (dft, 2025-02-14): _it has an odd number of counters on it_
- **Ty Lee, Chi Blocker** (tla, 2025-11-21): _you control Ty Lee_
- **Princess Yue** (tle, 2025-11-21): _she was a nonland creature_
- **Katara, the Fearless** (tla, 2025-11-21): _a triggered ability of an Ally you control triggers_
- **Need for Speed (Not the Odyssey One)** (unk, 2025-02-21): _you haven't paid its mana cost_
- **Need for Speed (Not the Odyssey One)** (unk, 2025-02-21): _your speed is 1 or higher_
- **Shadow the Hedgehog** (sld, 2025-07-14): _it's on the stack_
- **Fabled Passage** (soc, 2026-04-24): _you control four or more lands_
- **Unstickerify** (unk, 2025-02-21): _enchanted permanent isn't a stickered playtest card_
- **Unstickerify** (unk, 2025-02-21): _enchanted permanent is a stickered playtest card_
- **Unstickerify** (unk, 2025-02-21): _it's unclear without peeling the sticker_
- **Tezzeret, Cruel Captain** (eoe, 2025-08-01): _it's an artifact creature_
- **Turncoat Kunoichi** (tmt, 2026-03-06): _this creature's sneak cost was paid_
- **Zealous Display** (eoe, 2025-08-01): _it's not your turn_
- **Scarlet Spider, Ben Reilly** (spm, 2025-09-26): _Scarlet Spider was cast using web-slinging_
- **Shen, Wish Granter** (unk, 2025-06-20): _you haven't scattered the Dragonstorm Globes this game_
- **Shen, Wish Granter** (unk, 2025-06-20): _you control seven permanents named Dragonstorm Globe_
- **The Vast Scrier** (unk, 2025-09-26): _it has any "Whenever this creature attacks" triggers_
- **Stolen Uniform** (fin, 2025-06-13): _it's attached to a creature you control_
- **Tellah, Great Sage** (fin, 2025-06-13): _four or more mana was spent to cast that spell_
- **Tellah, Great Sage** (fin, 2025-06-13): _eight or more mana was spent to cast that spell_
- **Primordial Hydra** (soc, 2026-04-24): _it has ten or more +1/+1 counters on it_
- **Gavel of the Righteous** (eoc, 2025-08-01): _this Equipment has four or more counters on it_
- **Ray Fillet, Wave Warrior** (tmc, 2026-03-06): _that creature has greater power or toughness than this creature_
- **Catch-Up Mechanic** (unk, 2025-02-21): _an opponent has at least 5 more life than you_
- **Spikeshell Harrier** (dft, 2025-02-14): _that opponent's speed is greater than each other player's speed_
- **Chic // Ago** (unk, 2025-02-21): _it's sophisticated_
- **Amber-Plate Ainok** (ytdm, 2025-04-29): _this creature is tapped_
- **Wedding Announcement // Wedding Festivity** (inr, 2025-01-24): _this enchantment has three or more invitation counters on it_
- **Chalice of Life // Chalice of Death** (inr, 2025-01-24): _you have at least 10 life more than your starting life total_
- **Monet, Sensei of the Sewers** (unk, 2025-08-02): _Monet was ninjutsu'd_
- **The Fact Checker** (unk, 2025-02-21): _they got anything wrong_
- **Retto, Family Racer** (unk, 2025-02-21): _this is your commander or in your deck_
- **The Policy Maker** (unk, 2025-09-26): _each player in the game agrees to the rule_
- **The Policy Maker** (unk, 2025-09-26): _an opponent doesn't agree to the rule_
- **Desculpting Blast** (eoe, 2025-08-01): _it was attacking_
- **Harmonic Prodigy** (soc, 2026-04-24): _a triggered ability of a Shaman or another Wizard you control triggers_
- **Venat, Heart of Hydaelyn // Hydaelyn, the Mothercrystal** (fin, 2025-06-13): _that creature is legendary_
- **Doctor Octopus, Master Planner** (spm, 2025-09-26): _you have fewer than eight cards in hand_
- **Tend to the Kiln** (yecl, 2026-02-03): _it has three or more flame counters on it_
- **Y'shtola, Night's Blessed** (fic, 2025-06-13): _a player lost 4 or more life this turn_
- **Haunting Voyage** (ecc, 2026-01-23): _this spell was foretold_
- **Cloud, Midgar Mercenary** (fin, 2025-06-13): _a triggered ability of Cloud or an Equipment attached to it triggers_
- **Leonardo, Leader in Blue** (tmt, 2026-03-06): _his sneak cost was paid_
- **Nyla, Shirshu Sleuth** (tle, 2025-11-21): _you control no Clues_
- **Gaelicat** (fin, 2025-06-13): _you control two or more artifacts_
- **Kuja, Genome Sorcerer // Trance Kuja, Fate Defied** (fin, 2025-06-13): _you control four or more Wizards_
- **Kuja, Genome Sorcerer // Trance Kuja, Fate Defied** (fin, 2025-06-13): _a Wizard you control would deal damage to a permanent or player_
- **Unforgiving Overtake** (ydft, 2025-03-04): _you weren't the starting player_
- _…and 238 more singletons_