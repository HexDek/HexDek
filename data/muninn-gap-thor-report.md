# Thor — Per-Card Interaction Stress Test Report

**Date:** 2026-05-05 23:23:15
**Cards tested:** 104
**Total tests:** 2496
**Failures:** 43
**Time:** 0s
**Rate:** 12287 tests/s

## Invariant Violations (43)

| Card | Interaction | Invariant | Message |
|------|-------------|-----------|--------|
| Tolsimir, Midnight's Light | goldilocks_dead_effect |  | effect=parsed_effect_residual abilityKind=triggered filterBase="": board was set up but nothing changed |
| Quicksilver Fountain | goldilocks_dead_effect |  | effect=modification_effect abilityKind=triggered filterBase="": board was set up but nothing changed |
| Chainer, Nightmare Adept | goldilocks_dead_effect |  | effect=parsed_effect_residual abilityKind=activated filterBase="": board was set up but nothing changed |
| Goblin Goliath | goldilocks_dead_effect |  | effect=parsed_effect_residual abilityKind=triggered filterBase="": board was set up but nothing changed |
| Taii Wakeen, Perfect Shot | goldilocks_dead_effect |  | effect=untyped_effect abilityKind=triggered filterBase="": board was set up but nothing changed |
| Titania, Voice of Gaea | goldilocks_dead_effect |  | effect=untyped_effect abilityKind=triggered filterBase="": board was set up but nothing changed |
| River Song's Diary | goldilocks_dead_effect |  | effect=ability_word abilityKind=static filterBase="": board was set up but nothing changed |
| Fearless Swashbuckler | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Gau, Feral Youth | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Lighthouse Chronologist | goldilocks_dead_effect |  | effect=modification_effect abilityKind=static filterBase="": board was set up but nothing changed |
| Sunderflock | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Darigaaz Reincarnated | goldilocks_dead_effect |  | effect=if_intervening_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Valakut Exploration | goldilocks_dead_effect |  | effect=ability_word abilityKind=static filterBase="": board was set up but nothing changed |
| Sproutback Trudge | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Minthara, Merciless Soul | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Quicksilver Fountain | goldilocks_dead_effect |  | effect=modification_effect abilityKind=triggered filterBase="": board was set up but nothing changed |
| Taii Wakeen, Perfect Shot | goldilocks_dead_effect |  | effect=untyped_effect abilityKind=triggered filterBase="": board was set up but nothing changed |
| Tolsimir, Midnight's Light | goldilocks_dead_effect |  | effect=parsed_effect_residual abilityKind=triggered filterBase="": board was set up but nothing changed |
| Chainer, Nightmare Adept | goldilocks_dead_effect |  | effect=parsed_effect_residual abilityKind=activated filterBase="": board was set up but nothing changed |
| Titania, Voice of Gaea | goldilocks_dead_effect |  | effect=untyped_effect abilityKind=triggered filterBase="": board was set up but nothing changed |
| River Song's Diary | goldilocks_dead_effect |  | effect=ability_word abilityKind=static filterBase="": board was set up but nothing changed |
| Fearless Swashbuckler | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Gau, Feral Youth | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Lighthouse Chronologist | goldilocks_dead_effect |  | effect=modification_effect abilityKind=static filterBase="": board was set up but nothing changed |
| Sunderflock | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Darigaaz Reincarnated | goldilocks_dead_effect |  | effect=if_intervening_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Valakut Exploration | goldilocks_dead_effect |  | effect=ability_word abilityKind=static filterBase="": board was set up but nothing changed |
| Goblin Goliath | goldilocks_dead_effect |  | effect=parsed_effect_residual abilityKind=triggered filterBase="": board was set up but nothing changed |
| Sproutback Trudge | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Minthara, Merciless Soul | goldilocks_dead_effect |  | effect=parsed_tail abilityKind=static filterBase="": board was set up but nothing changed |
| Ingenious Prodigy | corpus_audit_draw |  | event_log: no events logged for draw effect |
| Frodo, Adventurous Hobbit | corpus_audit_draw |  | event_log: no events logged for draw effect |
| Vibrance | corpus_audit_gain_life |  | event_log: no events logged for gain_life effect |
| Kaito Shizuki | corpus_audit_draw |  | event_log: no events logged for draw effect |
| Arguel's Blood Fast // Temple of Aclazotz | corpus_audit_draw |  | event_log: no events logged for draw effect |
| Arguel's Blood Fast // Temple of Aclazotz | corpus_audit_gain_life |  | event_log: no events logged for gain_life effect |
| Markov Purifier | corpus_audit_draw |  | event_log: no events logged for draw effect |
| Senu, Keen-Eyed Protector | corpus_audit_gain_life |  | event_log: no events logged for gain_life effect |
| Crown of Gondor | corpus_audit_buff |  | buff: expected P/T change (+1/+1), no modification observed |
| Bloodchief Ascension | corpus_audit_gain_life |  | event_log: no events logged for gain_life effect |
| Grave Venerations | corpus_audit_lose_life |  | event_log: no events logged for lose_life effect |
| Grave Venerations | corpus_audit_gain_life |  | event_log: no events logged for gain_life effect |
| Deceit | corpus_audit_discard |  | discard: expected hand-1 (seat 1), got delta=0 |

## Failures by Interaction

| Interaction | Failures |
|-------------|----------|
| goldilocks_dead_effect | 30 |
| corpus_audit_draw | 5 |
| corpus_audit_gain_life | 5 |
| corpus_audit_discard | 1 |
| corpus_audit_buff | 1 |
| corpus_audit_lose_life | 1 |
