package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// etb_tribe_gate_family.go — generic handler for the
// "When this creature enters, if you control another <tribe>, <effect>"
// family.
//
// Shape (Ghitu Journeymage, Dreamcaller Siren, Acclaimed Contender, ...):
//
//	Creature — <Tribe> ...
//	When this creature enters, if you control another <Tribe>, <effect>.
//
// The differences are confined to the tribe and the effect. The gate is
// always "controller has another battlefield permanent with the named
// subtype (excluding self)". Self exclusion matters: the entering card
// itself shares the subtype, so a naive count-of-tribe ≥ 1 would always
// fire.
//
// Adding a new family member is one entry in etbTribeGateEntries.

type etbTribeGateEntry struct {
	cardName string
	tribe    string // subtype, e.g. "wizard", "pirate", "knight"
	effect   func(gs *gameengine.GameState, perm *gameengine.Permanent)
	partial  string // optional emitPartial reason
}

var etbTribeGateEntries = []etbTribeGateEntry{
	{
		// Ghitu Journeymage — {2}{R}, 2/2 Human Wizard.
		//   When this creature enters, if you control another Wizard,
		//   this creature deals 2 damage to each opponent.
		cardName: "Ghitu Journeymage",
		tribe:    "wizard",
		effect:   ghituJourneymageBurnEachOpponent,
	},
	{
		// Dreamcaller Siren — {3}{U}, flash, flying. 3/3 Siren Pirate.
		//   When this creature enters, if you control another Pirate, tap
		//   up to two target nonland permanents.
		cardName: "Dreamcaller Siren",
		tribe:    "pirate",
		effect:   dreamcallerSirenTapTwo,
	},
	{
		// Acclaimed Contender — {3}{W}, 3/3 Human Knight.
		//   When this creature enters, if you control another Knight, look
		//   at the top five cards of your library. You may reveal a Knight,
		//   Aura, Equipment, or legendary artifact card from among them and
		//   put it into your hand. Put the rest on the bottom of your
		//   library in a random order.
		cardName: "Acclaimed Contender",
		tribe:    "knight",
		effect:   acclaimedContenderTutorTopFive,
	},
}

func registerEtbTribeGateFamily(r *Registry) {
	for _, e := range etbTribeGateEntries {
		e := e
		r.OnETB(e.cardName, func(gs *gameengine.GameState, perm *gameengine.Permanent) {
			runEtbTribeGate(gs, perm, e)
		})
	}
}

func runEtbTribeGate(gs *gameengine.GameState, perm *gameengine.Permanent, e etbTribeGateEntry) {
	slug := "etb_tribe_gate_family:" + landFetchSlug(e.cardName)
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	if !controlsAnotherWithSubtype(seat, perm, e.tribe) {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_other_tribe_member", map[string]interface{}{
			"seat":  seatIdx,
			"tribe": e.tribe,
		})
		return
	}
	if e.effect != nil {
		e.effect(gs, perm)
	}
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  seatIdx,
		"tribe": e.tribe,
	})
	if e.partial != "" {
		emitPartial(gs, slug, perm.Card.DisplayName(), e.partial)
	}
}

// controlsAnotherWithSubtype reports whether seat controls at least one
// permanent with the given creature subtype that isn't `self`. Used to
// resolve "if you control another <tribe>" gates.
func controlsAnotherWithSubtype(seat *gameengine.Seat, self *gameengine.Permanent, sub string) bool {
	if seat == nil {
		return false
	}
	for _, p := range seat.Battlefield {
		if p == nil || p == self || p.Card == nil {
			continue
		}
		if cardHasSubtype(p.Card, sub) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Card bodies.
// ---------------------------------------------------------------------------

func ghituJourneymageBurnEachOpponent(gs *gameengine.GameState, perm *gameengine.Permanent) {
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		gameengine.LoseLife(gs, opp, 2, perm.Card.DisplayName())
	}
}

func dreamcallerSirenTapTwo(gs *gameengine.GameState, perm *gameengine.Permanent) {
	// "Up to two target nonland permanents". Priority: tap untapped
	// permanents an opponent controls, biggest impact first
	// (highest-CMC creatures > highest-CMC noncreatures). Then fall back
	// to our own permanents only if there are no opponent targets — but
	// since this is an aggressive tempo card we prefer to do nothing if
	// no opponent permanent is open.
	picks := dreamcallerSirenPickTargets(gs, perm.Controller, 2)
	for _, p := range picks {
		p.Tapped = true
	}
}

func dreamcallerSirenPickTargets(gs *gameengine.GameState, mySeat int, limit int) []*gameengine.Permanent {
	type cand struct {
		p        *gameengine.Permanent
		creature bool
		cmc      int
	}
	var pool []cand
	for i, s := range gs.Seats {
		if i == mySeat || s == nil || s.Lost {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || p.Tapped {
				continue
			}
			if cardHasType(p.Card, "land") {
				continue
			}
			pool = append(pool, cand{p: p, creature: p.IsCreature(), cmc: cardCMC(p.Card)})
		}
	}
	// Sort: creatures first, higher CMC first. Stable bubble — pool is small.
	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			swap := false
			if pool[j].creature && !pool[i].creature {
				swap = true
			} else if pool[j].creature == pool[i].creature && pool[j].cmc > pool[i].cmc {
				swap = true
			}
			if swap {
				pool[i], pool[j] = pool[j], pool[i]
			}
		}
	}
	if limit > len(pool) {
		limit = len(pool)
	}
	out := make([]*gameengine.Permanent, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, pool[i].p)
	}
	return out
}

func acclaimedContenderTutorTopFive(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "etb_tribe_gate_family:acclaimed_contender"
	seat := gs.Seats[perm.Controller]
	if seat == nil || len(seat.Library) == 0 {
		return
	}
	n := 5
	if n > len(seat.Library) {
		n = len(seat.Library)
	}
	top := append([]*gameengine.Card(nil), seat.Library[:n]...)

	// "Knight, Aura, Equipment, or legendary artifact card."
	// Priority among qualifying cards:
	//   1. Highest-CMC Knight (biggest body / impact)
	//   2. Equipment
	//   3. Aura
	//   4. Legendary artifact
	pickIdx := -1
	pickRank := 99
	pickCMC := -1
	for i, c := range top {
		if c == nil {
			continue
		}
		rank := -1
		switch {
		case cardHasSubtype(c, "knight") && cardHasType(c, "creature"):
			rank = 0
		case cardHasSubtype(c, "equipment"):
			rank = 1
		case cardHasSubtype(c, "aura"):
			rank = 2
		case cardHasType(c, "legendary") && cardHasType(c, "artifact"):
			rank = 3
		}
		if rank < 0 {
			continue
		}
		cmc := cardCMC(c)
		if rank < pickRank || (rank == pickRank && cmc > pickCMC) {
			pickRank = rank
			pickCMC = cmc
			pickIdx = i
		}
	}

	seat.Library = seat.Library[n:]
	var revealed string
	if pickIdx >= 0 {
		picked := top[pickIdx]
		revealed = picked.DisplayName()
		gameengine.MoveCard(gs, picked, perm.Controller, "library", "hand", "acclaimed_contender_reveal")
	}
	for i, c := range top {
		if i == pickIdx {
			continue
		}
		seat.Library = append(seat.Library, c)
	}
	// "In a random order" — the unseen-order shim used in birthing_ritual.
	// Bottom order is deterministic here; from the player's POV the cards
	// are no longer trackable, so the order is observationally random.
	if revealed != "" {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"revealed": revealed,
			"rank":     pickRank,
		})
	} else {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_qualifying_card_in_top_five", map[string]interface{}{
			"seat":     perm.Controller,
			"looked":   n,
		})
	}
}
