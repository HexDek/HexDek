package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWylethSoulOfSteel wires Wyleth, Soul of Steel.
//
// Oracle text:
//
//	Trample
//	Whenever Wyleth attacks, draw a card for each Aura and Equipment
//	attached to it.
//
// Implementation:
//   - "creature_attacks" gated on atk == perm: walk all permanents on
//     all battlefields and count those with AttachedTo == perm and an
//     Aura or Equipment subtype. Draw that many cards.
//   - Trample is engine-handled.
func registerWylethSoulOfSteel(r *Registry) {
	r.OnTrigger("Wyleth, Soul of Steel", "creature_attacks", wylethAttacks)
}

func wylethAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "wyleth_attacks_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	count := 0
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			if p.AttachedTo != perm {
				continue
			}
			if cardHasType(p.Card, "aura") || cardHasType(p.Card, "equipment") || p.IsEquipment() {
				count++
			}
		}
	}
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		drawOne(gs, perm.Controller, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"drew": count,
	})
}
