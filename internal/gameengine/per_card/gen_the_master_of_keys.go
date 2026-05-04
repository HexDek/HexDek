package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheMasterOfKeys wires The Master of Keys.
//
// Oracle text:
//
//   Flying
//   When The Master of Keys enters, put X +1/+1 counters on it and mill twice X cards.
//   Each enchantment card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from your graveyard. (You may cast cards from your graveyard for their escape cost.)
//
// Cost: {X}{W}{U}{B}. X is read from the printed CMC; X-cost tracking
// at resolution time is not surfaced cleanly here, so we use CMC minus
// the three fixed colored pips. The static "enchantments have escape"
// rider is a continuous granted-keyword effect not modeled here.
func registerTheMasterOfKeys(r *Registry) {
	r.OnETB("The Master of Keys", theMasterOfKeysETB)
}

func theMasterOfKeysETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "the_master_of_keys_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	x := 0
	if perm.Card != nil {
		x = perm.Card.CMC - 3
	}
	if x < 0 {
		x = 0
	}
	if x > 0 {
		perm.AddCounter("+1/+1", x)
		gs.InvalidateCharacteristicsCache()
	}
	milled := 0
	for i := 0; i < 2*x && len(s.Library) > 0; i++ {
		c := s.Library[0]
		gameengine.MoveCard(gs, c, seat, "library", "graveyard", "the_master_of_keys_mill")
		milled++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"x":        x,
		"counters": x,
		"milled":   milled,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"escape_grant_to_graveyard_enchantments_unimplemented")
}
