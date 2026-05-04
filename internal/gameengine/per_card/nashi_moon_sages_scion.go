package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNashiMoonSagesScion wires Nashi, Moon Sage's Scion.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Ninjutsu {3}{B}
//	Whenever Nashi deals combat damage to a player, exile the top card
//	  of each player's library. Until end of turn, you may play one of
//	  those cards. If you cast a spell this way, pay life equal to its
//	  mana value rather than paying its mana cost.
//
// Implementation:
//   - "combat_damage_player": gate on source_perm == perm and damage_seat
//     == perm.Controller. For each player, mill-to-exile the top card.
//     Free-play within the turn isn't routed through the per-card
//     pipeline; we emitPartial.
//   - Ninjutsu cost activation handled by the AST keyword pipeline.
func registerNashiMoonSagesScion(r *Registry) {
	r.OnTrigger("Nashi, Moon Sage's Scion", "combat_damage_player", nashiCombatDamage)
}

func nashiCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "nashi_combat_damage_exile"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	src, _ := ctx["source_perm"].(*gameengine.Permanent)
	if src != perm {
		return
	}
	dmgSeat, _ := ctx["damage_seat"].(int)
	if dmgSeat != perm.Controller {
		return
	}
	exiled := 0
	for i := range gs.Seats {
		s := gs.Seats[i]
		if s == nil || s.Lost {
			continue
		}
		if len(s.Library) == 0 {
			continue
		}
		c := s.Library[0]
		gameengine.MoveCard(gs, c, i, "library", "exile", "nashi_combat_exile")
		exiled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"exiled": exiled,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"free_play_from_exile_with_life_payment_not_modeled")
}
