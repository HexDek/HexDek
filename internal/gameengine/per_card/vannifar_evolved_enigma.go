package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVannifarEvolvedEnigma wires Vannifar, Evolved Enigma.
//
// Oracle text:
//
//	At the beginning of combat on your turn, choose one —
//	• Cloak a card from your hand. (Put it onto the battlefield face down
//	  as a 2/2 creature with ward {2}. Turn it face up any time for its
//	  mana cost if it's a creature card.)
//	• Put a +1/+1 counter on each colorless creature you control.
//
// Implementation:
//   - "combat_begin" (only on Vannifar's controller's turn): if there's at
//     least one colorless creature controlled, prefer the +1/+1 counter
//     mode (board-wide upside). Else cloak the top card of hand as a face-
//     down 2/2 token. Cloak's "turn face up" mechanic is not modeled — the
//     token stays a 2/2 face-down creature.
func registerVannifarEvolvedEnigma(r *Registry) {
	r.OnTrigger("Vannifar, Evolved Enigma", "combat_begin", vannifarCombatBegin)
}

func vannifarCombatBegin(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "vannifar_evolved_enigma_combat"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	// Count colorless creatures we control.
	colorless := 0
	var bumped []string
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if len(p.Card.Colors) == 0 {
			colorless++
			bumped = append(bumped, p.Card.DisplayName())
		}
	}

	if colorless > 0 {
		for _, p := range seat.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			if len(p.Card.Colors) == 0 {
				p.AddCounter("+1/+1", 1)
			}
		}
		gs.InvalidateCharacteristicsCache()
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"mode":     "counters",
			"affected": colorless,
		})
		return
	}

	// Cloak fallback — needs a card in hand.
	if len(seat.Hand) == 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_targets_or_hand", nil)
		return
	}
	pick := seat.Hand[0]
	moveCardBetweenZones(gs, perm.Controller, pick, "hand", "exile", "vannifar_cloak_source")
	token := &gameengine.Card{
		Name:          "Cloaked Creature",
		Owner:         perm.Controller,
		BasePower:     2,
		BaseToughness: 2,
		Types:         []string{"token", "creature"},
		Colors:        nil,
		TypeLine:      "Token Creature",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emitPartial(gs, slug, perm.Card.DisplayName(), "cloak_turn_face_up_not_modeled")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"mode": "cloak",
	})
	_ = bumped
}
