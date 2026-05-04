package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPlarggAndNassari wires Plargg and Nassari.
//
// Oracle text:
//
//   At the beginning of your upkeep, each player exiles cards from the top
//   of their library until they exile a nonland card. An opponent chooses a
//   nonland card exiled this way. You may cast up to two spells from among
//   the other cards exiled this way without paying their mana costs.
//
// Implementation policy:
//   - Each player exiles top cards until a nonland is exiled.
//   - The opposing-choice clause is approximated: the opponent (chosen as
//     the next seat after Plargg's controller) is treated as picking the
//     card whose CMC is highest (denying the controller the most upside).
//     Real player choice would require a UI/agent hook.
//   - Up to two of the remaining nonland exiled cards are "free-cast" by
//     dropping them onto the battlefield directly if they're permanents,
//     or moved to hand otherwise (instant/sorcery free-cast resolution
//     shortcut not implemented — same workaround Bre uses).
func registerPlarggAndNassari(r *Registry) {
	r.OnTrigger("Plargg and Nassari", "upkeep_controller", plarggAndNassariTrigger)
}

func plarggAndNassariTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "plargg_and_nassari_trigger"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}

	type exiled struct {
		card     *gameengine.Card
		fromSeat int
		cmc      int
	}
	var nonlands []exiled
	totalLands := 0
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for len(s.Library) > 0 {
			top := s.Library[0]
			if top == nil {
				s.Library = s.Library[1:]
				continue
			}
			gameengine.MoveCard(gs, top, i, "library", "exile", "plargg_nassari_dig")
			if cardHasType(top, "land") {
				totalLands++
				continue
			}
			nonlands = append(nonlands, exiled{card: top, fromSeat: i, cmc: cardCMC(top)})
			break
		}
	}

	if len(nonlands) == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":         perm.Controller,
			"lands_exiled": totalLands,
			"nonlands":     0,
		})
		return
	}

	// Opponent picks one card to give to Plargg's controller (i.e., the
	// "given" card cannot be free-cast). Approximate the opponent's choice
	// as the highest-CMC card (denies controller the biggest payoff).
	givenIdx := 0
	for i := 1; i < len(nonlands); i++ {
		if nonlands[i].cmc > nonlands[givenIdx].cmc {
			givenIdx = i
		}
	}
	given := nonlands[givenIdx]
	// "Chooses a nonland card exiled this way" — common ruling: the chosen
	// card stays in exile (the controller doesn't get to do anything with
	// it). We leave it in exile.
	_ = given

	// Free-cast up to two of the others. Greedy: highest-CMC first (best
	// value).
	type cand struct {
		card     *gameengine.Card
		fromSeat int
		cmc      int
	}
	var pool []cand
	for i, e := range nonlands {
		if i == givenIdx {
			continue
		}
		pool = append(pool, cand{card: e.card, fromSeat: e.fromSeat, cmc: e.cmc})
	}
	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			if pool[j].cmc > pool[i].cmc {
				pool[i], pool[j] = pool[j], pool[i]
			}
		}
	}
	if len(pool) > 2 {
		pool = pool[:2]
	}

	freeCast := 0
	partial := 0
	for _, c := range pool {
		// Pull from the original owner's exile pile, hand control to
		// Plargg's controller's battlefield (permanent) or hand (spell).
		owner := gs.Seats[c.fromSeat]
		if owner == nil {
			continue
		}
		removed := false
		for i, ec := range owner.Exile {
			if ec == c.card {
				owner.Exile = append(owner.Exile[:i], owner.Exile[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			continue
		}
		if cardHasType(c.card, "creature") || cardHasType(c.card, "artifact") ||
			cardHasType(c.card, "enchantment") || cardHasType(c.card, "planeswalker") ||
			cardHasType(c.card, "battle") {
			enterBattlefieldWithETB(gs, perm.Controller, c.card, false)
			freeCast++
		} else {
			// Instant/sorcery free-cast — engine doesn't support shortcut
			// casting. Send to controller's hand as a graceful fallback.
			gs.Seats[perm.Controller].Hand = append(gs.Seats[perm.Controller].Hand, c.card)
			partial++
		}
	}
	if partial > 0 {
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"instant_or_sorcery_free_cast_resolution_shortcut_unimplemented")
	}

	givenName := ""
	if given.card != nil {
		givenName = given.card.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":               perm.Controller,
		"lands_exiled":       totalLands,
		"nonlands":           len(nonlands),
		"given_to_opponent":  givenName,
		"free_cast_to_field": freeCast,
		"sent_to_hand":       partial,
	})
}
