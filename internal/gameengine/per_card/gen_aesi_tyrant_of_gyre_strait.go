package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAesiTyrantOfGyreStrait wires Aesi, Tyrant of Gyre Strait.
//
// Oracle text (Duskmourn Commander reprint, {4}{G}{U}, 5/5):
//
//	You may play an additional land on each of your turns.
//	Landfall — Whenever a land you control enters, you may draw a card.
//
// Implementation:
//   - ETB grants the controller an extra land drop via the
//     `extra_land_drops` seat flag (matches Flubs / Hearthhull pattern;
//     consumed best-effort by the land-play action).
//   - "permanent_etb" trigger gated on the entering permanent being a
//     land controlled by Aesi's controller fires the landfall draw.
//     The "may" choice always opts in (drawing a card is strictly
//     positive in nearly all positions).
func registerAesiTyrantOfGyreStrait(r *Registry) {
	r.OnETB("Aesi, Tyrant of Gyre Strait", aesiETB)
	r.OnTrigger("Aesi, Tyrant of Gyre Strait", "permanent_etb", aesiLandfall)
}

func aesiETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "aesi_etb_extra_land_drop"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["extra_land_drops"]++
	gs.LogEvent(gameengine.Event{
		Kind:   "extra_land_drop",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"slug":   slug,
			"reason": "aesi_static_additional_land",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func aesiLandfall(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "aesi_landfall_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil || entering == perm {
		return
	}
	if !entering.IsLand() {
		return
	}
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}
	if c := drawOne(gs, perm.Controller, perm.Card.DisplayName()); c != nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":  perm.Controller,
			"land":  entering.Card.DisplayName(),
			"drawn": c.DisplayName(),
		})
	}
}
