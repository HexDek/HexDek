# Parser Coverage Report

Pool: **31,963 real cards**.

## Headline

- 🟢 GREEN (every ability parsed cleanly): **31,949** (99.96%)
- 🟡 PARTIAL (some abilities parsed, others left as raw text): **12** (0.04%)
- 🔴 UNPARSED (parser couldn't recognize any abilities): **2** (0.01%)

**Goal: 100% GREEN.** Every PARTIAL/UNPARSED entry corresponds to a specific
unhandled grammar production — fixable by adding an effect rule, a trigger pattern,
or a keyword. No heuristic catch-alls.

## Top unparsed fragments — the work queue

Each row is an unparsed clause prefix. The count is how many cards' parse failed
at this prefix. Tackling the highest-count entries first shrinks the queue fastest.

| Count | Fragment prefix |
|---:|---|
| 1 | `whenever a kithkin card is put into your` |
| 1 | `other goblin creatures you control attack each combat` |
| 1 | `when hythonia becomes monstrous, destroy all non-gorgon creatures` |
| 1 | `teysa intensifies by 1` |
| 1 | `zurgo attacks each combat if able` |
| 1 | `braulios's power and toughness are each equal to` |
| 1 | `ruhan attacks that player this combat if able` |
| 1 | `whenever beregond or another human you control enters,` |
| 1 | `haktos attacks each combat if able` |
| 1 | `as haktos enters, choose 2, 3, or 4` |
| 1 | `as shimatsu enters, sacrifice any number of permanents` |
| 1 | `maraxus's power and toughness are each equal to` |
| 1 | `namor's power is equal to the number of` |
| 1 | `as saskia enters, choose a player` |
| 1 | `whenever you search your library, scry 1` |