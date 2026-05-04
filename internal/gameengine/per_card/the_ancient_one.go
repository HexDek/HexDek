package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheAncientOne wires The Ancient One.
//
// Oracle text:
//
//	{U}{B}
//	Legendary Creature — Spirit God
//	Descend 8 — The Ancient One can't attack or block unless there are
//	  eight or more permanent cards in your graveyard.
//	{2}{U}{B}: Draw a card, then discard a card. When you discard a card
//	  this way, target player mills cards equal to its mana value.
//
// Implementation:
//   - Descend 8 attack/block restriction: stash the requirement on
//     perm.Flags so combat code can read it. Engine's combat-restriction
//     pipeline does not currently consult this flag — emitPartial covers
//     the gap.
//   - Activated ability: emitPartial — the {2}{U}{B} loot + mill rider is
//     a chained-trigger effect that needs the engine's discard-event
//     pipeline to surface mana value of the discarded card.
func registerTheAncientOne(r *Registry) {
	r.OnETB("The Ancient One", theAncientOneETB)
	r.OnActivated("The Ancient One", theAncientOneActivated)
}

func theAncientOneETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["descend_8_attack_block_restriction"] = 1
	emit(gs, "the_ancient_one_etb", perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, "the_ancient_one_etb", perm.Card.DisplayName(),
		"descend_8_attack_block_restriction_not_consumed_by_combat_partial")
}

func theAncientOneActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "the_ancient_one_loot_mill"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	// Draw a card.
	drew := drawOne(gs, src.Controller, src.Card.DisplayName())
	// Discard a card. We don't have a clean per_card discard helper;
	// do it inline by grabbing the first hand card and moving to GY.
	discarded := (*gameengine.Card)(nil)
	if len(seat.Hand) > 0 {
		discarded = seat.Hand[0]
		moveCardBetweenZones(gs, src.Controller, discarded, "hand", "graveyard", "the_ancient_one_discard")
	}
	// Mill = MV of discarded.
	if discarded != nil {
		mv := cardCMC(discarded)
		if mv > 0 {
			oppSeat := -1
			for _, opp := range gs.Opponents(src.Controller) {
				oppSeat = opp
				break
			}
			if oppSeat >= 0 {
				target := gs.Seats[oppSeat]
				milled := 0
				for milled < mv && len(target.Library) > 0 {
					top := target.Library[0]
					moveCardBetweenZones(gs, oppSeat, top, "library", "graveyard", "the_ancient_one_mill")
					milled++
				}
				emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
					"seat":         src.Controller,
					"target_seat":  oppSeat,
					"discarded_mv": mv,
					"milled":       milled,
					"discarded":    discarded.DisplayName(),
					"drew":         theAncientOneCardName(drew),
				})
				return
			}
		}
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":      src.Controller,
		"drew":      theAncientOneCardName(drew),
		"discarded": theAncientOneCardName(discarded),
	})
}

func theAncientOneCardName(c *gameengine.Card) string {
	if c == nil {
		return ""
	}
	return c.DisplayName()
}
