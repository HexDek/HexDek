# Funny Incidents Log

Memorable bugs and accidental features from HexDek development.

---

## The Depressed Hat (2026-05-04)

**What happened:** Added a "conviction concession" system to YggdrasilHat so the AI would concede when losing badly. Used a 4-turn rolling window of relative position scores with a -0.35 threshold after turn 10.

**The bug:** The hat started conceding at 38 life with a full board because it *felt sad* about its position relative to opponents. Not dying. Not facing lethal. Just... bummed out. Showing "GG CONCESSION" on screen while the player was objectively fine.

**Josh's reaction:** "actually genuinely fucking hilarious we accidentally added depression to the hats" and "no crying in the casino, everyone fights to the death"

**Root cause:** The concession was meant for infinite loop / resolver lockup bailouts, but the score-based threshold turned it into an emotional surrender system. A player at 95% health got a life dimension score of only +0.225, which was easily overwhelmed by board-state metrics.

**Fix:** Removed score-based conviction entirely. The engine already handles infinite loops via SBA cap and stack loop detection. Everyone fights to the death now.

**Commit:** `hat: remove depression-based conviction concession`
