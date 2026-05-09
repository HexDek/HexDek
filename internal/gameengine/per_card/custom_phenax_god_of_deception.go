package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPhenaxGodOfDeceptionCustom implements the *granted* mill
// activation that Phenax confers on every creature its controller
// owns. The auto-generated stub (gen_phenax_god_of_deception.go) is a
// no-op.
//
// Oracle text:
//
//	Indestructible
//	As long as your devotion to blue and black is less than seven, Phenax
//	isn't a creature.
//	Creatures you control have "{T}: Target player mills X cards, where
//	X is this creature's toughness."
//
// Implementation notes:
//   - Indestructible / devotion-creature toggle are static — handled by
//     the AST keyword pipeline / countDevotion machinery elsewhere.
//   - The granted ability is on every creature, not on Phenax itself,
//     but the engine surfaces a single "Phenax-flavored" activation
//     hook because that's what the audited stub registered. We fold the
//     full grant into one resolver: pick our biggest untapped
//     non-summoning-sick creature, tap it, and mill X = its toughness
//     from the most-vulnerable opponent (smallest library).
//   - Devotion gate is enforced defensively: if devotion < 7 Phenax
//     isn't a creature and shouldn't be tapping anything, but the
//     granted ability is still active because Phenax merely needs to
//     be on the battlefield (CR §400.2). Devotion does NOT gate the
//     grant — it gates whether Phenax himself is a creature.
func registerPhenaxGodOfDeceptionCustom(r *Registry) {
	r.OnActivated("Phenax, God of Deception", phenaxGrantedMill)
}

func phenaxGrantedMill(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "phenax_granted_mill"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	seatIdx := src.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return
	}

	// Pick our highest-toughness untapped, non-summoning-sick creature
	// (Phenax himself is eligible if his devotion gate is satisfied).
	var miller *gameengine.Permanent
	bestT := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if p.Tapped || p.SummoningSick {
			continue
		}
		t := gs.ToughnessOf(p)
		if t <= 0 {
			continue
		}
		if miller == nil || t > bestT {
			miller = p
			bestT = t
		}
	}
	if miller == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_untapped_creature", nil)
		return
	}

	// Pick the opponent with the smallest library (closest to decking).
	target := -1
	smallestLib := 1<<31 - 1
	for _, oppIdx := range gs.Opponents(seatIdx) {
		opp := gs.Seats[oppIdx]
		if opp == nil || opp.Lost || opp.LeftGame {
			continue
		}
		if len(opp.Library) < smallestLib {
			smallestLib = len(opp.Library)
			target = oppIdx
		}
	}
	if target < 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "no_opponent", nil)
		return
	}

	// Tap and mill.
	miller.Tapped = true
	milled := 0
	for i := 0; i < bestT; i++ {
		ts := gs.Seats[target]
		if ts == nil || len(ts.Library) == 0 {
			break
		}
		card := ts.Library[0]
		gameengine.MoveCard(gs, card, target, "library", "graveyard", "phenax_mill")
		milled++
	}

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":      seatIdx,
		"miller":    miller.Card.DisplayName(),
		"toughness": bestT,
		"target":    target,
		"milled":    milled,
	})
}
