package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOhabiCaleria wires Ohabi Caleria.
//
// Oracle text:
//
//	Reach
//	Untap all Archers you control during each other player's untap step.
//	Whenever an Archer you control deals damage to a creature, you may
//	pay {2}. If you do, draw a card.
//
// The bonus-untap-step clause is left as a parser gap (untap-step
// observer hooks aren't wired). The damage-on-archer draw is wired via
// combat_damage_creature, with a {2} pay heuristic.
func registerOhabiCaleria(r *Registry) {
	r.OnTrigger("Ohabi Caleria", "combat_damage_creature", ohabiArcherDamage)
}

func ohabiArcherDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ohabi_archer_damage_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	sourceName, _ := ctx["source_card"].(string)
	if sourceSeat != perm.Controller {
		return
	}
	if !ohabiSourceIsArcher(gs, sourceSeat, sourceName) {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if !ohabiTryPay2(seat) {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": perm.Controller,
			"paid": false,
		})
		return
	}
	if len(seat.Library) > 0 {
		c := seat.Library[0]
		gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", "draw")
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"paid": true,
	})
}

func ohabiSourceIsArcher(gs *gameengine.GameState, seatIdx int, name string) bool {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	s := gs.Seats[seatIdx]
	if s == nil {
		return false
	}
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil || p.Card.DisplayName() != name {
			continue
		}
		for _, t := range p.Card.Types {
			if strings.EqualFold(t, "archer") {
				return true
			}
		}
		return strings.Contains(strings.ToLower(p.Card.TypeLine), "archer")
	}
	return false
}

func ohabiTryPay2(seat *gameengine.Seat) bool {
	if seat == nil {
		return false
	}
	if seat.Mana != nil {
		total := seat.Mana.W + seat.Mana.U + seat.Mana.B + seat.Mana.R + seat.Mana.G + seat.Mana.C + seat.Mana.Any
		if total < 2 {
			return false
		}
		paid := 0
		for paid < 2 {
			switch {
			case seat.Mana.Any > 0:
				seat.Mana.Any--
			case seat.Mana.C > 0:
				seat.Mana.C--
			case seat.Mana.W > 0:
				seat.Mana.W--
			case seat.Mana.U > 0:
				seat.Mana.U--
			case seat.Mana.B > 0:
				seat.Mana.B--
			case seat.Mana.R > 0:
				seat.Mana.R--
			case seat.Mana.G > 0:
				seat.Mana.G--
			default:
				return false
			}
			paid++
		}
		gameengine.SyncManaAfterSpend(seat)
		return true
	}
	if seat.ManaPool >= 2 {
		seat.ManaPool -= 2
		gameengine.SyncManaAfterSpend(seat)
		return true
	}
	return false
}
