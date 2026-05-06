# Thor ↔ Muninn Cross-Reference — 2026-05-06

**Thor failures parsed:** 43 (across 26 unique cards)  
**Muninn gap cards:** 105  
**Both fail (confirmed bugs):** 26  
**Thor-only (likely harness issues):** 0  
**Muninn-only (Thor blind spots):** 79

## Both Fail (confirmed bugs) — 26

| Card | Thor Failure | Muninn Context |
|------|--------------|----------------|
| Arguel's Blood Fast // Temple of Aclazotz | corpus_audit_draw; corpus_audit_gain_life — event_log: no events logged for draw effect | parser_gaps=2 |
| Bloodchief Ascension | corpus_audit_gain_life — event_log: no events logged for gain_life effect | parser_gaps=15 |
| Chainer, Nightmare Adept | goldilocks_dead_effect (x2) — effect=parsed_effect_residual abilityKind=activated filterBase="": board was se… | parser_gaps=1 |
| Crown of Gondor | corpus_audit_buff — buff: expected P/T change (+1/+1), no modification observed | parser_gaps=3 |
| Darigaaz Reincarnated | goldilocks_dead_effect (x2) — effect=if_intervening_tail abilityKind=static filterBase="": board was set up b… | parser_gaps=1 |
| Deceit | corpus_audit_discard — discard: expected hand-1 (seat 1), got delta=0 | parser_gaps=2 |
| Fearless Swashbuckler | goldilocks_dead_effect (x2) — effect=parsed_tail abilityKind=static filterBase="": board was set up but nothi… | parser_gaps=2 |
| Frodo, Adventurous Hobbit | corpus_audit_draw — event_log: no events logged for draw effect | (listed gap, no JSON detail) |
| Gau, Feral Youth | goldilocks_dead_effect (x2) — effect=parsed_tail abilityKind=static filterBase="": board was set up but nothi… | (listed gap, no JSON detail) |
| Goblin Goliath | goldilocks_dead_effect (x2) — effect=parsed_effect_residual abilityKind=triggered filterBase="": board was se… | (listed gap, no JSON detail) |
| Grave Venerations | corpus_audit_lose_life; corpus_audit_gain_life — event_log: no events logged for lose_life effect | parser_gaps=6 |
| Ingenious Prodigy | corpus_audit_draw — event_log: no events logged for draw effect | (listed gap, no JSON detail) |
| Kaito Shizuki | corpus_audit_draw — event_log: no events logged for draw effect | parser_gaps=3 |
| Lighthouse Chronologist | goldilocks_dead_effect (x2) — effect=modification_effect abilityKind=static filterBase="": board was set up b… | (listed gap, no JSON detail) |
| Markov Purifier | corpus_audit_draw — event_log: no events logged for draw effect | (listed gap, no JSON detail) |
| Minthara, Merciless Soul | goldilocks_dead_effect (x2) — effect=parsed_tail abilityKind=static filterBase="": board was set up but nothi… | (listed gap, no JSON detail) |
| Quicksilver Fountain | goldilocks_dead_effect (x2) — effect=modification_effect abilityKind=triggered filterBase="": board was set u… | (listed gap, no JSON detail) |
| River Song's Diary | goldilocks_dead_effect (x2) — effect=ability_word abilityKind=static filterBase="": board was set up but noth… | parser_gaps=1 |
| Senu, Keen-Eyed Protector | corpus_audit_gain_life — event_log: no events logged for gain_life effect | (listed gap, no JSON detail) |
| Sproutback Trudge | goldilocks_dead_effect (x2) — effect=parsed_tail abilityKind=static filterBase="": board was set up but nothi… | (listed gap, no JSON detail) |
| Sunderflock | goldilocks_dead_effect (x2) — effect=parsed_tail abilityKind=static filterBase="": board was set up but nothi… | parser_gaps=2 |
| Taii Wakeen, Perfect Shot | goldilocks_dead_effect (x2) — effect=untyped_effect abilityKind=triggered filterBase="": board was set up but… | parser_gaps=1 |
| Titania, Voice of Gaea | goldilocks_dead_effect (x2) — effect=untyped_effect abilityKind=triggered filterBase="": board was set up but… | parser_gaps=1 |
| Tolsimir, Midnight's Light | goldilocks_dead_effect (x2) — effect=parsed_effect_residual abilityKind=triggered filterBase="": board was se… | (listed gap, no JSON detail) |
| Valakut Exploration | goldilocks_dead_effect (x2) — effect=ability_word abilityKind=static filterBase="": board was set up but noth… | parser_gaps=2 |
| Vibrance | corpus_audit_gain_life — event_log: no events logged for gain_life effect | parser_gaps=5 |

## Thor-Only (likely harness issues) — 0

_None._

## Muninn-Only (Thor blind spots) — 79

| Card | Muninn Gap | Why Thor Misses It |
|------|------------|--------------------|
| Acclaimed Contender | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Acererak the Archlich | parser_gaps=3 | passes Thor — interaction not exercised by current battery |
| Aerial Surveyor | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Angel of Destiny | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Aradesh, the Founder | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Archmage Ascension | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Birthing Ritual | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Breaching Leviathan | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Bringer of the Last Gift | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Claim Jumper | parser_gaps=3 | passes Thor — interaction not exercised by current battery |
| Compy Swarm | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Courier Bat | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Crackling Spellslinger | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Curious Homunculus // Voracious Reader | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Cyclone Summoner | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Dreamcaller Siren | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Eccentric Pestfinder // Turn Stones | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Elanor Gardner | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Elderscale Wurm | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Emeritus of Woe // Demonic Tutor | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Eon Frolicker | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Evercoat Ursine | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Feast of the Victorious Dead | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Frodo, Sauron's Bane | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Genesis Chamber | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Geological Appraiser | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Ghitu Journeymage | parser_gaps=4 | passes Thor — interaction not exercised by current battery |
| Gisela, the Broken Blade | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Grave Scrabbler | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Great Hall of the Biblioplex | parser_gaps=5 | passes Thor — interaction not exercised by current battery |
| Gruff Triplets | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Hurkyl, Master Wizard | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Ichorid | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Kami of Transience | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Knight of the White Orchid | parser_gaps=12 | passes Thor — interaction not exercised by current battery |
| Kodama of the East Tree | parser_gaps=4 | passes Thor — interaction not exercised by current battery |
| Land Tax | parser_gaps=26 | passes Thor — interaction not exercised by current battery |
| Lasting Tarfire | parser_gaps=7 | passes Thor — interaction not exercised by current battery |
| Lathiel, the Bounteous Dawn | parser_gaps=4 | passes Thor — interaction not exercised by current battery |
| Leonardo, Leader in Blue | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Life of the Party | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Light-Paws, Emperor's Voice | parser_gaps=12 | passes Thor — interaction not exercised by current battery |
| Lord Jyscal Guado | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Lorehold Archivist // Restore Relic | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Loyal Warhound | parser_gaps=5 | passes Thor — interaction not exercised by current battery |
| Lux Artillery | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Necromancy | parser_gaps=10 | passes Thor — interaction not exercised by current battery |
| Nessian Wilds Ravager | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Oracle of Bones | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Oversold Cemetery | parser_gaps=3 | passes Thor — interaction not exercised by current battery |
| Phoenix Fleet Airship | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Preston, the Vanisher | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Rankle and Torbran | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Ravenloft Adventurer | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Rocco, Cabaretti Caterer | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Sand Scout | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Sarcomancy | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Scorpion, Seething Striker | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Septic Rats | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Skitterbeam Battalion | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Smirking Spelljacker | parser_gaps=4 | passes Thor — interaction not exercised by current battery |
| Sméagol, Helpful Guide | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Star Charter | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Starlit Soothsayer | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| The One Ring | parser_gaps=86, dead_triggers=58 | passes Thor — interaction not exercised by current battery |
| Tiamat | parser_gaps=13 | passes Thor — interaction not exercised by current battery |
| Tiamat (Miirym Token) | parser_gaps=3 | passes Thor — interaction not exercised by current battery |
| Tivash, Gloom Summoner | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Tombstone Stairwell | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Transcendent Dragon | parser_gaps=2 | passes Thor — interaction not exercised by current battery |
| Twilight Prophet | parser_gaps=8 | passes Thor — interaction not exercised by current battery |
| Unstable Glyphbridge // Sandswirl Wanderglyph | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Viconia, Drow Apostate | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Wary Farmer | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Weftwalking | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Wild Pair | parser_gaps=1 | passes Thor — interaction not exercised by current battery |
| Wistfulness | parser_gaps=6 | passes Thor — interaction not exercised by current battery |
| Witch of the Moors | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |
| Yathan Roadwatcher | (listed gap, no JSON detail) | passes Thor — interaction not exercised by current battery |

