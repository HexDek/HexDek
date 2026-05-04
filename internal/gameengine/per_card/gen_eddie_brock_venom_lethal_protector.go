package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEddieBrockVenomLethalProtector wires Eddie Brock // Venom, Lethal Protector.
//
// Front face — Eddie Brock:
//
//	When Eddie Brock enters, return target creature card with mana value
//	1 or less from your graveyard to the battlefield.
//	{3}{B}{R}{G}: Transform Eddie Brock. Activate only as a sorcery.
//
// Back face — Venom, Lethal Protector:
//
//	Menace, trample, haste
//	Whenever Venom attacks, you may sacrifice another creature. If you
//	do, draw X cards, then you may put a permanent card with mana value
//	X or less from your hand onto the battlefield, where X is the
//	sacrificed creature's mana value.
func registerEddieBrockVenomLethalProtector(r *Registry) {
	r.OnETB("Eddie Brock // Venom, Lethal Protector", eddieBrockVenomLethalProtectorETB)
	r.OnActivated("Eddie Brock // Venom, Lethal Protector", eddieBrockActivated)
	r.OnTrigger("Venom, Lethal Protector", "attacks", venomAttacks)
}

func eddieBrockVenomLethalProtectorETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "eddie_brock_etb_recur_mv1"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Pick the highest-power creature with mv <= 1 in graveyard.
	var best *gameengine.Card
	bestScore := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		if cardCMC(c) > 1 {
			continue
		}
		score := int(c.BasePower) + int(c.BaseToughness)
		if score > bestScore {
			bestScore = score
			best = c
		}
	}
	if best == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"target": "none",
		})
		return
	}
	gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "battlefield", "eddie_brock_recur")
	enterBattlefieldWithETB(gs, perm.Controller, best, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": best.DisplayName(),
	})
}

func eddieBrockActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "eddie_brock_activated_transform"
	if gs == nil || src == nil {
		return
	}
	if !gameengine.TransformPermanent(gs, src, "eddie_brock_to_venom") {
		emitPartial(gs, slug, src.Card.DisplayName(),
			"transform_failed_face_data_missing")
		return
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":    src.Controller,
		"ability": "transform",
	})
}

func venomAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "venom_lethal_protector_attacks"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	attackerSeat, _ := ctx["seat"].(int)
	if attackerSeat != perm.Controller {
		return
	}
	// Only fires when Venom himself attacks — gate on identity.
	if attackingPerm, ok := ctx["perm"].(*gameengine.Permanent); ok && attackingPerm != nil {
		if attackingPerm != perm {
			return
		}
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	// Pick a sacrificial creature: avoid Venom himself; prefer the
	// highest-mv creature so X is biggest (more cards drawn + bigger
	// permanent into play). Only consider creatures we control.
	var sac *gameengine.Permanent
	bestMV := -1
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || !p.IsCreature() || p.Card == nil {
			continue
		}
		mv := cardCMC(p.Card)
		if mv > bestMV {
			bestMV = mv
			sac = p
		}
	}
	if sac == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":       perm.Controller,
			"sacrificed": "none",
		})
		return
	}
	x := cardCMC(sac.Card)
	sacName := sac.Card.DisplayName()
	gameengine.SacrificePermanent(gs, sac, "venom_attacks_sac")

	drawn := 0
	for i := 0; i < x && len(seat.Library) > 0; i++ {
		top := seat.Library[0]
		gameengine.MoveCard(gs, top, perm.Controller, "library", "hand", "venom_attacks_draw")
		drawn++
	}

	// Optionally cheat in a permanent of mv <= X from hand. Pick the
	// highest-mv permanent that fits (bigger = bigger swing).
	var dropped *gameengine.Card
	dropMV := -1
	for _, c := range seat.Hand {
		if c == nil {
			continue
		}
		if !cardHasType(c, "creature") && !cardHasType(c, "artifact") &&
			!cardHasType(c, "enchantment") && !cardHasType(c, "planeswalker") &&
			!cardHasType(c, "battle") && !cardHasType(c, "land") {
			continue
		}
		mv := cardCMC(c)
		if mv > x {
			continue
		}
		if mv > dropMV {
			dropMV = mv
			dropped = c
		}
	}
	if dropped != nil {
		gameengine.MoveCard(gs, dropped, perm.Controller, "hand", "battlefield", "venom_attacks_cheat_in")
		enterBattlefieldWithETB(gs, perm.Controller, dropped, false)
	}

	details := map[string]interface{}{
		"seat":       perm.Controller,
		"sacrificed": sacName,
		"x":          x,
		"drawn":      drawn,
	}
	if dropped != nil {
		details["dropped"] = dropped.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), details)
}
