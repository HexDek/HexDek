package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKardurDoomscourge wires Kardur, Doomscourge.
//
// Oracle text (Duskmourn Commander reprint, {2}{B}{R}, 4/3):
//
//	When Kardur enters, until your next turn, creatures your opponents
//	control attack each combat if able and attack a player other than
//	you if able.
//	Whenever an attacking creature dies, each opponent loses 1 life and
//	you gain 1 life.
//
// Implementation:
//   - ETB goads every opponent creature on the battlefield (the
//     "attack each combat / attack a player other than you" effect is
//     equivalent to applying goad). Duration "until your next turn" is
//     not modeled — goad persists; this overshoots, but represents the
//     intended pressure correctly until a delayed-trigger cleanup is
//     wired.
//   - "creature_dies" trigger: if the dead creature was attacking
//     when it died, each opponent of Kardur's controller loses 1 life
//     and Kardur's controller gains 1.
func registerKardurDoomscourge(r *Registry) {
	r.OnETB("Kardur, Doomscourge", kardurDoomscourgeETB)
	r.OnTrigger("Kardur, Doomscourge", "creature_dies", kardurDoomscourgeDies)
}

func kardurDoomscourgeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "kardur_etb_goad_opponents"
	if gs == nil || perm == nil {
		return
	}
	goaded := 0
	for i, s := range gs.Seats {
		if s == nil || i == perm.Controller {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			if p.Flags == nil {
				p.Flags = map[string]int{}
			}
			p.Flags["goaded"] = 1
			goaded++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"goaded": goaded,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"goad_duration_until_your_next_turn_not_cleared_no_delayed_trigger")
}

func kardurDoomscourgeDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kardur_attacker_dies_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	died, _ := ctx["perm"].(*gameengine.Permanent)
	if died == nil || died.Card == nil {
		return
	}
	if died.Flags == nil || died.Flags["was_attacking"] == 0 {
		// Fallback: check still-set "attacking" flag (some death paths
		// fire dies before clearing the attack flag).
		if died.Flags == nil || died.Flags["attacking"] == 0 {
			return
		}
	}
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		gameengine.LoseLife(gs, i, 1, perm.Card.DisplayName())
	}
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"died":  died.Card.DisplayName(),
	})
}
