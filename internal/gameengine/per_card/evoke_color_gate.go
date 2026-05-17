package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// evoke_color_gate.go — generic ETB handler for the "hybrid-evoke,
// two color-gated ETBs" family.
//
// Shape (Vibrance, Wistfulness, Deceit, ...):
//
//	Hybrid mana cost {C1/C2}{C1/C2} (+ filler).
//	When ~ enters, if {C1}{C1} was spent to cast it, <effect A>.
//	When ~ enters, if {C2}{C2} was spent to cast it, <effect B>.
//	Evoke {C1/C2}{C1/C2}
//
// The engine doesn't yet track per-cast hybrid-pip payment (resolve.go's
// "mana_spent" condition defaults true — see internal/gameengine/resolve.go).
// Vibrance's hand-rolled handler set the convention: when the spell was
// cast (was_cast flag set), fire BOTH color modes; when it entered via a
// non-cast path (reanimate, blink), fire NEITHER. We replicate that here
// for the rest of the family and emitPartial on the residual gap so
// Muninn keeps tracking it.
//
// Adding a new evoke-hybrid card is two pieces:
//   1. Two ETB-effect closures (modeA, modeB), one per color half.
//   2. A one-line entry in evokeColorGateEntries below.

type evokeColorGateEntry struct {
	cardName string
	colorA   string // documentation only
	colorB   string
	modeA    func(gs *gameengine.GameState, perm *gameengine.Permanent)
	modeB    func(gs *gameengine.GameState, perm *gameengine.Permanent)
}

func registerEvokeColorGateFamily(r *Registry) {
	for _, e := range evokeColorGateEntries {
		e := e
		r.OnETB(e.cardName, func(gs *gameengine.GameState, perm *gameengine.Permanent) {
			runEvokeColorGate(gs, perm, e)
		})
	}
}

func runEvokeColorGate(gs *gameengine.GameState, perm *gameengine.Permanent, e evokeColorGateEntry) {
	slug := "evoke_color_gate:" + landFetchSlug(e.cardName)
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}
	wasCast := perm.Flags != nil && perm.Flags["was_cast"] != 0
	modes := []string{}
	if wasCast {
		if e.modeA != nil {
			e.modeA(gs, perm)
			modes = append(modes, e.colorA)
		}
		if e.modeB != nil {
			e.modeB(gs, perm)
			modes = append(modes, e.colorB)
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"was_cast": wasCast,
		"modes":    modes,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"hybrid_pip_payment_tracking_unmodeled_both_modes_fire_when_cast")
}

// ---------------------------------------------------------------------------
// Card bodies.
// ---------------------------------------------------------------------------

var evokeColorGateEntries = []evokeColorGateEntry{
	{
		// Wistfulness — evoke {G/U}{G/U}.
		//   {G}{G}: exile target artifact or enchantment an opponent controls.
		//   {U}{U}: draw two cards, then discard a card.
		cardName: "Wistfulness",
		colorA:   "green_exile_artifact_or_enchantment",
		colorB:   "blue_draw_two_discard_one",
		modeA:    wistfulnessExileArtifactOrEnchantment,
		modeB:    wistfulnessDrawTwoDiscardOne,
	},
	{
		// Deceit — evoke {U/B}{U/B}.
		//   {U}{U}: return up to one other target nonland permanent to its
		//           owner's hand.
		//   {B}{B}: target opponent reveals their hand. You choose a nonland
		//           card from it. That player discards that card.
		cardName: "Deceit",
		colorA:   "blue_bounce_nonland_permanent",
		colorB:   "black_opponent_discards_nonland_choice",
		modeA:    deceitBounceNonland,
		modeB:    deceitOpponentDiscardsChoice,
	},
}

// pickThreateningOpponentPermanent returns the highest-priority artifact or
// enchantment among opponents (highest CMC, deterministic ordering).
func pickThreateningOpponentPermanent(gs *gameengine.GameState, mySeat int, types ...string) *gameengine.Permanent {
	var best *gameengine.Permanent
	bestCMC := -1
	for i, s := range gs.Seats {
		if i == mySeat || s == nil || s.Lost {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			match := false
			for _, t := range types {
				if cardHasType(p.Card, t) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
			cmc := cardCMC(p.Card)
			if cmc > bestCMC {
				best = p
				bestCMC = cmc
			}
		}
	}
	return best
}

// pickHighestCMCNonlandPermanent returns the highest-CMC nonland permanent
// among opponents (deterministic, used for "return target nonland permanent").
func pickHighestCMCNonlandPermanent(gs *gameengine.GameState, mySeat int, excludeSelf *gameengine.Permanent) *gameengine.Permanent {
	var best *gameengine.Permanent
	bestCMC := -1
	for i, s := range gs.Seats {
		if i == mySeat || s == nil || s.Lost {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p == excludeSelf || p.Card == nil {
				continue
			}
			if cardHasType(p.Card, "land") {
				continue
			}
			cmc := cardCMC(p.Card)
			if cmc > bestCMC {
				best = p
				bestCMC = cmc
			}
		}
	}
	return best
}

func wistfulnessExileArtifactOrEnchantment(gs *gameengine.GameState, perm *gameengine.Permanent) {
	target := pickThreateningOpponentPermanent(gs, perm.Controller, "artifact", "enchantment")
	if target != nil {
		gameengine.ExilePermanent(gs, target, perm)
	}
}

func wistfulnessDrawTwoDiscardOne(gs *gameengine.GameState, perm *gameengine.Permanent) {
	seat := perm.Controller
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	for i := 0; i < 2 && len(s.Library) > 0; i++ {
		card := s.Library[0]
		gameengine.MoveCard(gs, card, seat, "library", "hand", "wistfulness_draw")
	}
	if len(s.Hand) > 0 {
		pick := pickLowestValueDiscard(s.Hand)
		if pick != nil {
			gameengine.DiscardCard(gs, pick, seat)
		}
	}
}

func deceitBounceNonland(gs *gameengine.GameState, perm *gameengine.Permanent) {
	target := pickHighestCMCNonlandPermanent(gs, perm.Controller, perm)
	if target != nil {
		gameengine.BouncePermanent(gs, target, perm, "hand")
	}
}

func deceitOpponentDiscardsChoice(gs *gameengine.GameState, perm *gameengine.Permanent) {
	// Pick the opponent with the largest hand (most disruption).
	bestSeat := -1
	bestN := -1
	for i, s := range gs.Seats {
		if i == perm.Controller || s == nil || s.Lost {
			continue
		}
		if len(s.Hand) > bestN {
			bestSeat = i
			bestN = len(s.Hand)
		}
	}
	if bestSeat < 0 || bestN <= 0 {
		return
	}
	s := gs.Seats[bestSeat]
	// "You choose a nonland card": pick the highest-CMC nonland in hand.
	var pick *gameengine.Card
	pickCMC := -1
	for _, c := range s.Hand {
		if c == nil || cardHasType(c, "land") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > pickCMC {
			pick = c
			pickCMC = cmc
		}
	}
	if pick == nil {
		return
	}
	gameengine.DiscardCard(gs, pick, bestSeat)
}

// pickLowestValueDiscard returns the lowest-value card in hand to discard.
// "Lowest value" = lowest CMC, preferring lands. Deterministic.
func pickLowestValueDiscard(hand []*gameengine.Card) *gameengine.Card {
	var landPick *gameengine.Card
	var nonlandPick *gameengine.Card
	nonlandCMC := 99
	for _, c := range hand {
		if c == nil {
			continue
		}
		if cardHasType(c, "land") {
			if landPick == nil {
				landPick = c
			}
			continue
		}
		cmc := cardCMC(c)
		if cmc < nonlandCMC {
			nonlandPick = c
			nonlandCMC = cmc
		}
	}
	// If we already control 4+ lands, prefer to discard a land first;
	// otherwise discard the lowest-value nonland.
	// Heuristic: hand-size dependent — simpler default: discard the
	// lowest-CMC nonland, fallback to a land.
	if nonlandPick != nil {
		return nonlandPick
	}
	return landPick
}
