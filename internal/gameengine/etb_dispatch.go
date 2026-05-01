package gameengine

import "github.com/hexdek/hexdek/internal/gameast"

// FirePermanentETBTriggers fires the complete ETB trigger cascade for a
// permanent that has already been added to the battlefield. Handles:
//
//  1. Self-AST ETB triggers (the entering permanent's own ETB abilities)
//  2. Per-card self-ETB hook (snowflake handlers)
//  3. Ascend check (city's blessing at 10+ permanents)
//  4. Per-card hook events (permanent_etb, nonland_permanent_etb)
//  5. AST observer ETB triggers (other permanents watching for ETBs)
//
// The permanent must already be on the battlefield and replacement effects
// registered before calling this function.
func FirePermanentETBTriggers(gs *GameState, perm *Permanent) {
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}

	// CR §708.4: face-down permanents have no abilities — skip self-triggers.
	faceDown := perm.Flags != nil && perm.Flags["face_down"] != 0

	if !faceDown && perm.Card.AST != nil {
		for _, ab := range perm.Card.AST.Abilities {
			trig, ok := ab.(*gameast.Triggered)
			if !ok || trig.Effect == nil {
				continue
			}
			if !EventEquals(trig.Trigger.Event, "etb") {
				continue
			}
			PushTriggeredAbility(gs, perm, trig.Effect)
			if gs.CheckEnd() {
				return
			}
		}
	}

	if !faceDown {
		InvokeETBHook(gs, perm)
	}

	if !faceDown && perm.IsSaga() {
		initSagaLoreCounters(gs, perm)
	}

	CheckAscend(gs, perm.Controller)

	if !perm.IsLand() {
		FireCardTrigger(gs, "nonland_permanent_etb", map[string]interface{}{
			"perm":            perm,
			"controller_seat": perm.Controller,
			"card":            perm.Card,
		})
	}
	FireCardTrigger(gs, "permanent_etb", map[string]interface{}{
		"perm":            perm,
		"controller_seat": perm.Controller,
		"card":            perm.Card,
	})

	fireObserverETBTriggers(gs, perm)
}

var romanToInt = map[string]int{
	"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5,
	"VI": 6, "VII": 7, "VIII": 8, "IX": 9, "X": 10,
}

// initSagaLoreCounters scans the saga's AST for chapter abilities and sets
// saga_final_chapter to the highest chapter number found. Then adds 1 lore
// counter per CR §714.3a (saga gets first lore counter on ETB).
func initSagaLoreCounters(gs *GameState, perm *Permanent) {
	if perm.Card.AST == nil {
		return
	}
	maxChapter := 0
	for _, ab := range perm.Card.AST.Abilities {
		st, ok := ab.(*gameast.Static)
		if !ok || st.Modification == nil {
			continue
		}
		if st.Modification.ModKind != "saga_chapter" && st.Modification.ModKind != "parsed_tail" {
			continue
		}
		if st.Modification.ModKind == "saga_chapter" && len(st.Modification.Args) >= 1 {
			switch v := st.Modification.Args[0].(type) {
			case string:
				if n, ok := romanToInt[v]; ok && n > maxChapter {
					maxChapter = n
				}
			case []interface{}:
				for _, item := range v {
					if s, ok := item.(string); ok {
						if n, ok := romanToInt[s]; ok && n > maxChapter {
							maxChapter = n
						}
					}
				}
			}
		}
	}
	if maxChapter == 0 {
		maxChapter = 3
	}
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	perm.Counters["saga_final_chapter"] = maxChapter
	perm.AddCounter("lore", 1)
	gs.LogEvent(Event{
		Kind:   "saga_etb",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"final_chapter": maxChapter,
			"lore":          perm.Counters["lore"],
		},
	})
}
