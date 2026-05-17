package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCycloneSummoner wires Cyclone Summoner.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	When this creature enters, if you cast it from your hand, return
//	all permanents to their owners' hands except for Giants, Wizards,
//	and lands.
//
// Implementation (Muninn gap #36 — 27K hits):
//   - OnETB gated on perm.Flags["was_cast"] (stack.go stamps this on
//     the cast path; same gate as Tiamat / Vibrance). Reanimated /
//     blinked Cyclone Summoner doesn't trigger.
//   - The engine doesn't yet distinguish "from your hand" vs
//     "elsewhere" (cast from exile / graveyard); we pessimise:
//     was_cast already gates out non-cast paths, but rare cast-from-X
//     paths fire too. emitPartial.
//   - Bounce every non-Giant, non-Wizard, non-land permanent (Cyclone
//     Summoner is itself a Giant, so it survives). Uses BouncePermanent
//     so commander redirect and ZCT fire properly.
func registerCycloneSummoner(r *Registry) {
	r.OnETB("Cyclone Summoner", cycloneSummonerETB)
}

func cycloneSummonerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "cyclone_summoner_bounce"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	var victims []*gameengine.Permanent
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			if p.IsLand() {
				continue
			}
			if cardHasType(p.Card, "giant") || cardHasType(p.Card, "wizard") {
				continue
			}
			victims = append(victims, p)
		}
	}
	bounced := 0
	for _, p := range victims {
		gameengine.BouncePermanent(gs, p, perm, "hand")
		bounced++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"bounced": bounced,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cast_from_hand_specifically_unmodeled_was_cast_flag_only")
}
