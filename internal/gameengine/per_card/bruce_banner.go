package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBruceBanner wires Bruce Banner // The Incredible Hulk.
//
// Front face — Bruce Banner:
//
//	{X}{X}, {T}: Draw X cards. Activate only as a sorcery.
//	{2}{R}{R}{G}{G}: Transform Bruce Banner. Activate only as a sorcery.
//
// Back face — The Incredible Hulk:
//
//	Reach, trample
//	Enrage — Whenever The Incredible Hulk is dealt damage, put a +1/+1
//	counter on him. If he's attacking, untap him and there is an
//	additional combat phase after this phase.
//
// Both activation paths and the additional combat phase are non-trivial.
func registerBruceBanner(r *Registry) {
	r.OnActivated("Bruce Banner", bruceBannerActivated)
	r.OnActivated("Bruce Banner // The Incredible Hulk", bruceBannerActivated)
	r.OnTrigger("The Incredible Hulk", "damaged", bruceBannerEnrage)
	r.OnTrigger("Bruce Banner // The Incredible Hulk", "damaged", bruceBannerEnrage)
}

func bruceBannerActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "bruce_banner_activated"
	if gs == nil || src == nil {
		return
	}
	switch abilityIdx {
	case 0:
		// {X}{X}, {T}: Draw X cards.
		x := 0
		if ctx != nil {
			if v, ok := ctx["x"].(int); ok {
				x = v
			} else if v, ok := ctx["x_paid"].(int); ok {
				x = v
			}
		}
		seat := gs.Seats[src.Controller]
		if seat == nil {
			return
		}
		drawn := 0
		for i := 0; i < x && len(seat.Library) > 0; i++ {
			top := seat.Library[0]
			gameengine.MoveCard(gs, top, src.Controller, "library", "hand", "bruce_banner_draw")
			drawn++
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":    src.Controller,
			"ability": "draw_x",
			"x":       x,
			"drawn":   drawn,
		})
	case 1:
		// {2}{R}{R}{G}{G}: Transform Bruce Banner.
		if !gameengine.TransformPermanent(gs, src, "bruce_banner_to_hulk") {
			emitPartial(gs, slug, src.Card.DisplayName(),
				"transform_failed_face_data_missing")
			return
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":    src.Controller,
			"ability": "transform",
		})
	default:
		emitPartial(gs, slug, src.Card.DisplayName(),
			"unknown_ability_index")
	}
}

func bruceBannerEnrage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "incredible_hulk_enrage"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	targetPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if targetPerm != perm {
		return
	}
	perm.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()

	untapped := false
	extraCombat := false
	if perm.IsAttacking() {
		if perm.Tapped {
			perm.Tapped = false
			untapped = true
		}
		gs.AddExtraCombat(gameengine.PendingExtraCombat{
			SourceCard: perm.Card.DisplayName(),
		})
		extraCombat = true
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            perm.Controller,
		"untapped":        untapped,
		"extra_combat":    extraCombat,
		"pending_combats": len(gs.PendingExtraCombats),
	})
}
