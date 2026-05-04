package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNineFingersKeene wires Nine-Fingers Keene.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Menace
//	Ward—Pay 9 life.
//	Whenever Nine-Fingers Keene deals combat damage to a player, look
//	  at the top nine cards of your library. You may put a Gate card
//	  from among them onto the battlefield. Then if you control nine
//	  or more Gates, put the rest into your hand. Otherwise, put the
//	  rest on the bottom of your library in a random order.
//
// Implementation:
//   - "combat_damage_player": gate on source_perm == perm. Look at top 9
//     of our library; if any Gate, put the highest-CMC Gate onto the
//     battlefield. Count Gates we control after; if >= 9, remaining 8
//     cards go to hand, else they go to bottom of library.
//   - Menace and Ward handled by AST keyword pipeline.
func registerNineFingersKeene(r *Registry) {
	r.OnTrigger("Nine-Fingers Keene", "combat_damage_player", nineFingersCombat)
}

func nineFingersCombat(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "nine_fingers_keene_gate_dig"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	src, _ := ctx["source_perm"].(*gameengine.Permanent)
	if src != perm {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || len(seat.Library) == 0 {
		return
	}
	n := 9
	if len(seat.Library) < n {
		n = len(seat.Library)
	}
	top := seat.Library[:n]

	var gateIdx = -1
	bestCMC := -1
	for i, c := range top {
		if c == nil {
			continue
		}
		if cardHasType(c, "gate") {
			cm := gameengine.ManaCostOf(c)
			if cm > bestCMC {
				bestCMC = cm
				gateIdx = i
			}
		}
	}

	gateName := ""
	if gateIdx >= 0 {
		gate := top[gateIdx]
		gateName = gate.DisplayName()
		gameengine.MoveCard(gs, gate, perm.Controller, "library", "battlefield", "nine_fingers_keene")
		createPermanent(gs, perm.Controller, gate, false)
	}

	// Count gates we now control.
	gates := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "gate") {
			gates++
		}
	}

	// Remaining cards in original "top n" — those still in library top.
	// After the gate move, library may have shrunk by 1.
	remaining := []*gameengine.Card{}
	keepN := n
	if gateIdx >= 0 {
		keepN = n - 1
	}
	if keepN > len(seat.Library) {
		keepN = len(seat.Library)
	}
	for i := 0; i < keepN; i++ {
		remaining = append(remaining, seat.Library[i])
	}

	if gates >= 9 {
		for _, c := range remaining {
			gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", "nine_fingers_keene_remainder")
		}
	} else {
		for _, c := range remaining {
			gameengine.MoveCard(gs, c, perm.Controller, "library", "library", "nine_fingers_keene_to_bottom")
			// Move to bottom: simplest — put at end.
			// (MoveCard re-inserts at end of destination zone for "library".)
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"gate":      gateName,
		"gates_now": gates,
		"top_n":     n,
		"to_hand":   gates >= 9,
	})
}
