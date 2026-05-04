package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerValkiGodOfLies wires Valki, God of Lies // Tibalt, Cosmic Impostor.
//
// Oracle text (front face, Valki):
//
//	When Valki enters, each opponent reveals their hand. For each
//	opponent, exile a creature card they revealed this way until Valki
//	leaves the battlefield.
//	{X}: Choose a creature card exiled with Valki with mana value X.
//	Valki becomes a copy of that card.
//
// Implementation:
//   - ETB: walk each opponent's hand, exile the highest-CMC creature card
//     found (greedy "best disruption"). Tag each exiled card with a Flag
//     so an LTB (not implemented here) could in principle return them.
//   - Activated copy ability: emitPartial — copy-of-creature requires AST
//     deep-clone semantics that are out of scope for a per-card handler.
//   - Tibalt back face: not implemented (no transform path tracked).
func registerValkiGodOfLies(r *Registry) {
	r.OnETB("Valki, God of Lies // Tibalt, Cosmic Impostor", valkiGodOfLiesETB)
	r.OnETB("Valki, God of Lies", valkiGodOfLiesETB)
	r.OnActivated("Valki, God of Lies // Tibalt, Cosmic Impostor", valkiGodOfLiesActivate)
	r.OnActivated("Valki, God of Lies", valkiGodOfLiesActivate)
}

func valkiGodOfLiesETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "valki_god_of_lies_etb_exile"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	exiled := 0
	for i, opp := range gs.Seats {
		if opp == nil || i == seat || opp.Lost {
			continue
		}
		var pick *gameengine.Card
		bestCMC := -1
		for _, c := range opp.Hand {
			if c == nil || !cardHasType(c, "creature") {
				continue
			}
			if cmc := cardCMC(c); cmc > bestCMC {
				bestCMC = cmc
				pick = c
			}
		}
		if pick == nil {
			continue
		}
		moveCardBetweenZones(gs, i, pick, "hand", "exile", "valki_etb")
		exiled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   seat,
		"exiled": exiled,
	})
}

func valkiGodOfLiesActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "valki_god_of_lies_copy"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(), "copy_creature_card_not_implemented")
}
