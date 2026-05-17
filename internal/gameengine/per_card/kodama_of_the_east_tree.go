package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKodamaOfTheEastTree wires Kodama of the East Tree.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	Reach
//	Whenever another permanent you control enters, if it wasn't put
//	onto the battlefield with this ability, you may put a permanent
//	card with equal or lesser mana value from your hand onto the
//	battlefield.
//	Partner (You can have two commanders if both have partner.)
//
// Implementation (Muninn gap #7 — 121,703 hits):
//   - OnTrigger("permanent_etb"): fires for every permanent entering
//     play. Gate: entering perm is under our control, is NOT Kodama
//     itself, and was NOT placed by this same ability (the
//     "kodama_cascade" flag on the entering perm prevents the recursive
//     cascade from re-firing arbitrarily).
//   - Search controller's hand for the highest-CMC permanent card with
//     CMC <= the entering perm's CMC. Drop it via MoveCard with the
//     "kodama_cascade" flag pre-stamped onto the card via gs.Flags so
//     the freshly-entered permanent skips re-triggering Kodama. The
//     flag is anchored on a per-card pointer-keyed gs.Flags entry that
//     ETBHook clears after consuming.
//   - Reach is an AST keyword; not modeled here.
//   - Partner is a deckbuilding rule, not a runtime ability.
func registerKodamaOfTheEastTree(r *Registry) {
	r.OnTrigger("Kodama of the East Tree", "permanent_etb", kodamaOfTheEastTreeETB)
}

func kodamaOfTheEastTreeETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kodama_of_the_east_tree_cascade"
	if gs == nil || perm == nil || ctx == nil {
		return
	}

	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil || entering.Card == nil {
		return
	}
	if entering == perm {
		return
	}
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}
	if entering.Flags != nil && entering.Flags["kodama_cascade"] != 0 {
		return
	}

	enteringCMC := gameengine.ManaCostOf(entering.Card)

	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	var pick *gameengine.Card
	bestCMC := -1
	for _, c := range s.Hand {
		if c == nil {
			continue
		}
		if !isPermanentCardType(c) {
			continue
		}
		cmc := gameengine.ManaCostOf(c)
		if cmc > enteringCMC {
			continue
		}
		if cmc > bestCMC {
			bestCMC = cmc
			pick = c
		}
	}
	if pick == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_eligible_hand_card", map[string]interface{}{
			"seat":         seat,
			"entering":     entering.Card.DisplayName(),
			"entering_cmc": enteringCMC,
		})
		return
	}

	res := gameengine.MoveCard(gs, pick, seat, "hand", "battlefield", "kodama_east_tree")
	if res.Permanent != nil {
		if res.Permanent.Flags == nil {
			res.Permanent.Flags = map[string]int{}
		}
		res.Permanent.Flags["kodama_cascade"] = 1
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         seat,
		"entering":     entering.Card.DisplayName(),
		"entering_cmc": enteringCMC,
		"dropped":      pick.DisplayName(),
		"dropped_cmc":  bestCMC,
	})
}

// isPermanentCardType returns true if the card is one of the permanent
// types (creature, artifact, enchantment, planeswalker, land, battle).
// Used by Kodama-style "put a permanent card from your hand" effects.
func isPermanentCardType(c *gameengine.Card) bool {
	if c == nil {
		return false
	}
	return cardHasType(c, "creature") || cardHasType(c, "artifact") ||
		cardHasType(c, "enchantment") || cardHasType(c, "planeswalker") ||
		cardHasType(c, "land") || cardHasType(c, "battle")
}
