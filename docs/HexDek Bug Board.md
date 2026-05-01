---

kanban-plugin: board

---

## Known Unknowns



## Confirmed Bugs



## Under Investigation



## Fixed

- [x] Tergrid recursive trigger crash (depth guard + total trigger cap) #engine
- [x] Obeka wrong ability resolution #engine
- [x] DFC commander name mismatch #engine
- [x] Compound type filter for cast triggers #engine
- [x] 8 dead per_card triggers — 7 fixed, alias normalization added #engine
- [x] Freya false positives (~20/28) — self-exile, hand vs battlefield, attack-trigger dependency, randomness #freya
- [x] Copy token mana value — BaseCharacteristics() missing `c.CMC = p.Card.CMC`. Copy tokens (Satya, Kiki-Jiki, Clone, Sakashima, Rite of Replication) now inherit mana cost per CR §707.2 #engine #copy



%% kanban:settings
```
{"kanban-plugin":"board","list-collapse":[false,false,false,false]}
```
%%
