package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBelakorDarkMaster wires Be'lakor, the Dark Master.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	Prince of Chaos — When Be'lakor enters, you draw X cards and you
//	  lose X life, where X is the number of Demons you control.
//	Lord of Torment — Whenever another Demon you control enters, it
//	  deals damage equal to its power to any target.
//
// Implementation:
//   - OnETB: count Demons controlled (including Be'lakor itself, since
//     he resolves before this trigger). Draw X and lose X life.
//   - "permanent_etb": filter to Demon creatures controlled by us, not
//     Be'lakor itself. Use the entering creature's power as damage. Pick
//     the opponent with lowest life (best chance of being lethal).
func registerBelakorDarkMaster(r *Registry) {
	r.OnETB("Be'lakor, the Dark Master", belakorETB)
	r.OnTrigger("Be'lakor, the Dark Master", "permanent_etb", belakorOtherDemonETB)
}

func belakorETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "belakor_etb_draw_lose"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	x := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if p.IsCreature() && cardHasType(p.Card, "demon") {
			x++
		}
	}
	drawn := 0
	for i := 0; i < x; i++ {
		if drawOne(gs, perm.Controller, perm.Card.DisplayName()) != nil {
			drawn++
		}
	}
	if x > 0 {
		gameengine.LoseLife(gs, perm.Controller, x, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"x":         x,
		"drawn":     drawn,
		"life_loss": x,
	})
	_ = gs.CheckEnd()
}

func belakorOtherDemonETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "belakor_lord_of_torment"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enteringPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if enteringPerm == nil || enteringPerm == perm {
		return
	}
	if enteringPerm.Controller != perm.Controller {
		return
	}
	if enteringPerm.Card == nil || !enteringPerm.IsCreature() || !cardHasType(enteringPerm.Card, "demon") {
		return
	}
	dmg := enteringPerm.Power()
	if dmg <= 0 {
		return
	}
	// Pick the opponent with the lowest life total.
	target := -1
	bestLife := 1 << 30
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		if s.Life < bestLife {
			bestLife = s.Life
			target = opp
		}
	}
	if target < 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_opponent", nil)
		return
	}
	gameengine.DealDamage(gs, target, dmg, enteringPerm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"demon":       enteringPerm.Card.DisplayName(),
		"target_seat": target,
		"damage":      dmg,
	})
	_ = gs.CheckEnd()
}
