package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWitchOfTheMoors wires Witch of the Moors (Muninn parser-gap #38, 24,749 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{3}{B}{B}
//	Creature — Human Warlock
//	Deathtouch
//	At the beginning of your end step, if you gained life this turn,
//	each opponent sacrifices a creature of their choice and you return
//	up to one target creature card from your graveyard to your hand.
//
// Pattern mirrors lasting_tarfire.go: a per-turn life_gained flag plus an
// end_step listener that consumes the flag. Each opponent's "creature of
// their choice" is approximated as the lowest-CMC creature (the choice
// the controlling AI would prefer to give up).
func registerWitchOfTheMoors(r *Registry) {
	r.OnTrigger("Witch of the Moors", "life_gained", witchOfTheMoorsLifeGained)
	r.OnTrigger("Witch of the Moors", "end_step", witchOfTheMoorsEndStep)
}

func witchOfTheMoorsLifeGained(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat, _ := ctx["seat"].(int)
	if seat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	s := gs.Seats[perm.Controller]
	if s == nil {
		return
	}
	if s.Flags == nil {
		s.Flags = map[string]int{}
	}
	s.Flags[witchOfTheMoorsKey(gs.Turn)] = 1
}

func witchOfTheMoorsEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "witch_of_the_moors_end_step"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	s := gs.Seats[perm.Controller]
	if s == nil || s.Lost {
		return
	}
	key := witchOfTheMoorsKey(gs.Turn)
	if s.Flags == nil || s.Flags[key] == 0 {
		return
	}
	delete(s.Flags, key)
	witchOfTheMoorsPruneKeys(s, gs.Turn)

	sacrificed := 0
	for _, opp := range gs.Opponents(perm.Controller) {
		op := gs.Seats[opp]
		if op == nil || op.Lost {
			continue
		}
		var victim *gameengine.Permanent
		victimCMC := 1 << 30
		for _, p := range op.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			cmc := cardCMC(p.Card)
			if cmc < victimCMC {
				victimCMC = cmc
				victim = p
			}
		}
		if victim != nil {
			gameengine.SacrificePermanent(gs, victim, "witch_of_the_moors_drain")
			sacrificed++
		}
	}

	var best *gameengine.Card
	bestCMC := -1
	for _, c := range s.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			best = c
		}
	}
	returned := ""
	if best != nil {
		returned = best.DisplayName()
		gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "hand", "witch_of_the_moors_recur")
	}
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"opps_sac":   sacrificed,
		"returned":   returned,
	})
}

func witchOfTheMoorsKey(turn int) string {
	return fmt.Sprintf("witch_moors_life_t%d", turn+1)
}

func witchOfTheMoorsPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "witch_moors_life_t"
	cutoff := currentTurn + 1
	for k := range seat.Flags {
		if len(k) <= len(prefix) || k[:len(prefix)] != prefix {
			continue
		}
		n := 0
		_, err := fmt.Sscanf(k[len(prefix):], "%d", &n)
		if err != nil {
			continue
		}
		if n < cutoff {
			delete(seat.Flags, k)
		}
	}
}
