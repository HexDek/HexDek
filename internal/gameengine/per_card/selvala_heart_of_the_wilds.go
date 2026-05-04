package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSelvalaHeartOfTheWilds wires Selvala, Heart of the Wilds.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever another creature enters, its controller may draw a card
//	if its power is greater than each other creature's power.
//	{G}, {T}: Add X mana in any combination of colors, where X is the
//	greatest power among creatures you control.
//
// Implementation:
//   - "creature_enters_battlefield" trigger: when another creature
//     enters anywhere, compare its current power to every other
//     creature's power on every battlefield. If strictly greater than
//     all others, that creature's controller may draw a card. Selvala's
//     trigger fires for every controller, not just hers — that's why
//     we route the draw to ctx-derived controller seat.
//   - Activated mana ability: handled by AST mana pipeline; we don't
//     write a hook here (variable-color mana abilities are emitted by
//     the engine's per-color resolver).
func registerSelvalaHeartOfTheWilds(r *Registry) {
	r.OnTrigger("Selvala, Heart of the Wilds", "creature_enters_battlefield", selvalaHeartCreatureETB)
}

func selvalaHeartCreatureETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "selvala_heart_etb_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enteringPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if enteringPerm == nil || enteringPerm == perm || enteringPerm.Card == nil {
		return
	}
	if !enteringPerm.IsCreature() {
		return
	}
	enteringPow := enteringPerm.Power()
	// "Greater than each other creature's power": strictly more than every
	// other creature on the battlefield, including Selvala.
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p == enteringPerm || p.Card == nil || !p.IsCreature() {
				continue
			}
			if p.Power() >= enteringPow {
				return
			}
		}
	}
	// May-draw — always opt in (free card).
	owner := enteringPerm.Controller
	if owner < 0 || owner >= len(gs.Seats) || gs.Seats[owner] == nil || gs.Seats[owner].Lost {
		return
	}
	drawOne(gs, owner, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":           perm.Controller,
		"draw_seat":      owner,
		"entering":       enteringPerm.Card.DisplayName(),
		"entering_power": enteringPow,
	})
}
