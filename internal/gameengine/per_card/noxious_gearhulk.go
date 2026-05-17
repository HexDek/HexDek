package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNoxiousGearhulk wires Noxious Gearhulk (Muninn snowflake first
// seen 2026-05-15).
//
// Oracle text (Scryfall, verified 2026-05-17):
//
//	Noxious Gearhulk — {5}{B} Artifact Creature — Construct 5/4
//	Menace
//	When this creature enters, you may destroy another target creature.
//	If a creature is destroyed this way, you gain life equal to its
//	toughness.
//
// Implementation:
//   - OnETB: pick the highest-toughness opponent creature as default
//     target (ctx-provided target_perm wins). Destroy it via
//     SacrificePermanent style → DestroyPermanent, then gain life equal
//     to its printed toughness. Skip silently if no legal target.
func registerNoxiousGearhulk(r *Registry) {
	r.OnETB("Noxious Gearhulk", noxiousGearhulkETB)
}

func noxiousGearhulkETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "noxious_gearhulk_etb"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	// Pick a target: highest-toughness opponent creature.
	var target *gameengine.Permanent
	highest := -1
	for i, s := range gs.Seats {
		if s == nil || i == seat {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			t := p.Card.BaseToughness + p.Counters["+1/+1"]
			if t > highest {
				highest = t
				target = p
			}
		}
	}
	if target == nil {
		emit(gs, slug, "Noxious Gearhulk", map[string]interface{}{
			"seat":   seat,
			"target": nil,
		})
		return
	}
	toughness := target.Card.BaseToughness + target.Counters["+1/+1"]
	gameengine.DestroyPermanent(gs, target, perm)
	gameengine.GainLife(gs, seat, toughness, "Noxious Gearhulk")
	emit(gs, slug, "Noxious Gearhulk", map[string]interface{}{
		"seat":        seat,
		"target":      target.Card.DisplayName(),
		"target_seat": target.Controller,
		"life_gained": toughness,
	})
}
