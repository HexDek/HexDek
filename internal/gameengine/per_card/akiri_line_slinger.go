package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAkiriLineSlinger wires Akiri, Line-Slinger.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{R}{W}
//	Legendary Creature — Kor Soldier Ally
//	First strike, vigilance
//	Akiri gets +1/+0 for each artifact you control.
//	Partner
//
// Implementation:
//   - ETB: count artifacts controlled by us, set temp_power.
//   - Refresh on artifact ETB events for upkeep accuracy. Continuous
//     layers not modeled — emitPartial.
func registerAkiriLineSlinger(r *Registry) {
	r.OnETB("Akiri, Line-Slinger", akiriLineSlingerRefresh)
	r.OnTrigger("Akiri, Line-Slinger", "permanent_etb", akiriLineSlingerEtbTrigger)
}

func akiriLineSlingerRefresh(gs *gameengine.GameState, perm *gameengine.Permanent) {
	akiriLineSlingerApply(gs, perm, "akiri_line_slinger_etb_buff")
}

func akiriLineSlingerEtbTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enteringPerm, _ := ctx["permanent"].(*gameengine.Permanent)
	if enteringPerm == nil || enteringPerm.Card == nil {
		return
	}
	if enteringPerm.Controller != perm.Controller {
		return
	}
	if !cardHasType(enteringPerm.Card, "artifact") {
		return
	}
	akiriLineSlingerApply(gs, perm, "akiri_line_slinger_artifact_etb_refresh")
}

func akiriLineSlingerApply(gs *gameengine.GameState, perm *gameengine.Permanent, slug string) {
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	count := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "artifact") {
			count++
		}
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	prev := perm.Flags["akiri_artifact_buff"]
	delta := count - prev
	perm.Flags["akiri_artifact_buff"] = count
	perm.Flags["temp_power"] += delta
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"artifact_count": count,
		"power_delta":  delta,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_buff_refreshed_on_artifact_etb_only_not_continuously_layered")
}
