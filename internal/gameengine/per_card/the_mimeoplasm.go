package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheMimeoplasm wires The Mimeoplasm.
//
// Oracle text:
//
//	{2}{B}{G}{U}
//	Legendary Creature — Ooze
//	As The Mimeoplasm enters, you may exile two creature cards from
//	  graveyards. If you do, it enters as a copy of one of those cards
//	  with a number of additional +1/+1 counters on it equal to the
//	  power of the other card.
//
// Implementation:
//   - "As enters" replacement: handled at ETB time. Pick the two
//     highest-power creature cards across all graveyards. Exile both,
//     turn the Mimeoplasm into a token-style copy of the higher-power
//     one (we mutate perm.Card in place, copying name + power/toughness
//   - types + colors), and add (other.power) +1/+1 counters.
//   - This is a simplification — true "enters as a copy" copies the AST
//     and abilities; emitPartial covers the gap.
func registerTheMimeoplasm(r *Registry) {
	r.OnETB("The Mimeoplasm", theMimeoplasmETB)
}

func theMimeoplasmETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "the_mimeoplasm_etb_copy"
	if gs == nil || perm == nil {
		return
	}

	// Collect all creature cards in graveyards.
	type cand struct {
		card *gameengine.Card
		seat int
	}
	cands := []cand{}
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, c := range s.Graveyard {
			if c == nil {
				continue
			}
			if cardHasType(c, "creature") {
				cands = append(cands, cand{card: c, seat: i})
			}
		}
	}
	if len(cands) < 2 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"chosen": 0,
			"reason": "fewer_than_two_creature_cards_in_graveyards",
		})
		return
	}

	// Pick two highest-power.
	bestA, bestB := -1, -1
	for i := range cands {
		if bestA < 0 || cands[i].card.BasePower > cands[bestA].card.BasePower {
			bestB = bestA
			bestA = i
		} else if bestB < 0 || cands[i].card.BasePower > cands[bestB].card.BasePower {
			bestB = i
		}
	}
	copyOf := cands[bestA]
	powerOf := cands[bestB]

	// Exile both.
	moveCardBetweenZones(gs, copyOf.seat, copyOf.card, "graveyard", "exile", "the_mimeoplasm_exile")
	moveCardBetweenZones(gs, powerOf.seat, powerOf.card, "graveyard", "exile", "the_mimeoplasm_exile")

	// Mutate Mimeoplasm into a copy. Copy core fields but keep Owner.
	originalName := perm.Card.DisplayName()
	perm.Card = &gameengine.Card{
		Name:          copyOf.card.DisplayName(),
		Owner:         perm.Card.Owner,
		BasePower:     copyOf.card.BasePower,
		BaseToughness: copyOf.card.BaseToughness,
		Types:         append([]string(nil), copyOf.card.Types...),
		Colors:        append([]string(nil), copyOf.card.Colors...),
		TypeLine:      copyOf.card.TypeLine,
	}

	// Add power-of +1/+1 counters.
	addedCounters := powerOf.card.BasePower
	if addedCounters < 0 {
		addedCounters = 0
	}
	perm.AddCounter("+1/+1", addedCounters)

	emit(gs, slug, originalName, map[string]interface{}{
		"seat":     perm.Controller,
		"copy_of":  copyOf.card.DisplayName(),
		"power_of": powerOf.card.DisplayName(),
		"counters": addedCounters,
	})
	emitPartial(gs, slug, originalName,
		"enters_as_copy_of_only_copies_pt_types_colors_no_abilities_partial")
}
