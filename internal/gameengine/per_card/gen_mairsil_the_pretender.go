package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMairsilThePretender wires Mairsil, the Pretender.
//
// Oracle text:
//
//   When Mairsil enters, you may exile an artifact or creature card from
//   your hand or graveyard and put a cage counter on it.
//   Mairsil has all activated abilities of all cards you own in exile with
//   cage counters on them. You may activate each of those abilities only
//   once each turn.
//
// The static "has all activated abilities" clause requires the activated-
// ability granting / per-turn activation tracking subsystem. Implemented:
// the ETB exile-and-cage clause. The cage_counters flag on Mairsil tracks
// how many caged cards exist for downstream tooling.
func registerMairsilThePretender(r *Registry) {
	r.OnETB("Mairsil, the Pretender", mairsilThePretenderETB)
}

func mairsilThePretenderETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "mairsil_the_pretender_etb"
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

	// Greedy: prefer the highest-CMC artifact or creature available in
	// hand or graveyard (best activated abilities tend to ride bigger
	// cards). Search hand first, then graveyard.
	var best *gameengine.Card
	bestZone := ""
	bestIdx := -1
	bestCMC := -1
	for i, c := range s.Hand {
		if c == nil {
			continue
		}
		if !cardHasType(c, "artifact") && !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			best = c
			bestCMC = cmc
			bestZone = "hand"
			bestIdx = i
		}
	}
	for i, c := range s.Graveyard {
		if c == nil {
			continue
		}
		if !cardHasType(c, "artifact") && !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			best = c
			bestCMC = cmc
			bestZone = "graveyard"
			bestIdx = i
		}
	}

	if best == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"caged":  "",
			"reason": "no_eligible_artifact_or_creature",
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"static_grant_activated_abilities_unimplemented")
		return
	}

	_ = bestIdx
	gameengine.MoveCard(gs, best, seat, bestZone, "exile", "mairsil_cage")
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["cage_counters"]++
	gs.LogEvent(gameengine.Event{
		Kind:   "counter_added",
		Seat:   seat,
		Source: perm.Card.DisplayName(),
		Amount: 1,
		Details: map[string]interface{}{
			"counter_kind": "cage",
			"target_card":  best.DisplayName(),
			"from_zone":    bestZone,
		},
	})

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      seat,
		"caged":     best.DisplayName(),
		"from_zone": bestZone,
		"cmc":       bestCMC,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_grant_activated_abilities_unimplemented")
}
