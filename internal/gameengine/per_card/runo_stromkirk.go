package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRunoStromkirk wires Runo Stromkirk // Krothuss, Lord of the Deep.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
// Front — Runo Stromkirk ({1}{U}{B}, Vampire Cleric, 1/4):
//
//	Flying
//	When Runo enters, put up to one target creature card from your
//	graveyard on top of your library.
//	At the beginning of your upkeep, look at the top card of your
//	library. You may reveal that card. If a creature card with mana
//	value 6 or greater is revealed this way, transform Runo.
//
// Back — Krothuss, Lord of the Deep (Kraken Horror, 3/5):
//
//	Flying
//	Whenever Krothuss attacks, create a tapped and attacking token
//	that's a copy of another target attacking creature. If that
//	creature is a Kraken, Leviathan, Octopus, or Serpent, create two
//	of those tokens instead.
//
// Implementation:
//   - ETB front: pick highest-CMC creature in graveyard, top-of-library
//     it. Heuristic mirrors "what would I most want to draw next?".
//   - "upkeep_controller": peek top of library; if it's a creature with
//     CMC >= 6, transform Runo via TransformPermanent.
//   - Krothuss attack trigger: clone the highest-power other attacker;
//     double if it's a sea-creature subtype.
func registerRunoStromkirk(r *Registry) {
	r.OnETB("Runo Stromkirk // Krothuss, Lord of the Deep", runoETB)
	r.OnETB("Runo Stromkirk", runoETB)
	r.OnTrigger("Runo Stromkirk // Krothuss, Lord of the Deep", "upkeep_controller", runoUpkeepPeek)
	r.OnTrigger("Runo Stromkirk", "upkeep_controller", runoUpkeepPeek)
	r.OnTrigger("Runo Stromkirk // Krothuss, Lord of the Deep", "creature_attacks", krothussAttack)
	r.OnTrigger("Krothuss, Lord of the Deep", "creature_attacks", krothussAttack)
}

func runoETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "runo_etb_top_of_library"
	if gs == nil || perm == nil {
		return
	}
	if perm.Transformed {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var best *gameengine.Card
	bestCMC := -1
	for _, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		if cardCMC(c) > bestCMC {
			bestCMC = cardCMC(c)
			best = c
		}
	}
	if best == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_graveyard_creature", map[string]interface{}{
			"seat": perm.Controller,
		})
		return
	}
	gameengine.MoveCard(gs, best, perm.Controller, "graveyard", "library", "runo_top_of_library")
	// Move to top: MoveCard appends; physically swap to position 0.
	if len(seat.Library) > 1 {
		idx := -1
		for i, c := range seat.Library {
			if c == best {
				idx = i
				break
			}
		}
		if idx > 0 {
			seat.Library = append([]*gameengine.Card{best},
				append(seat.Library[:idx], seat.Library[idx+1:]...)...)
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"card": best.DisplayName(),
	})
}

func runoUpkeepPeek(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "runo_upkeep_peek_transform"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	if perm.Transformed {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || len(seat.Library) == 0 {
		return
	}
	top := seat.Library[0]
	if top == nil || !cardHasType(top, "creature") || cardCMC(top) < 6 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"peeked": cardDisp(top),
			"flip":   false,
		})
		return
	}
	if !gameengine.TransformPermanent(gs, perm, "runo_revealed_six_plus") {
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"transform_failed_face_data_missing")
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"peeked": top.DisplayName(),
		"flip":   true,
	})
}

func krothussAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "krothuss_clone_attacker"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	// Pick highest-power other attacker.
	var best *gameengine.Permanent
	bestPow := -1
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p == perm || p.Card == nil || !p.IsAttacking() {
				continue
			}
			if p.Power() > bestPow {
				bestPow = p.Power()
				best = p
			}
		}
	}
	if best == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_other_attacker", nil)
		return
	}
	double := false
	for _, t := range best.Card.Types {
		switch t {
		case "kraken", "leviathan", "octopus", "serpent":
			double = true
		}
	}
	count := 1
	if double {
		count = 2
	}
	for i := 0; i < count; i++ {
		copyTok := *best.Card
		copyTok.Name = best.Card.DisplayName() + " Token"
		copyTok.Owner = perm.Controller
		copyTok.Types = append([]string{"token"}, best.Card.Types...)
		tokenPerm := createPermanent(gs, perm.Controller, &copyTok, true)
		if tokenPerm != nil {
			tokenPerm.SummoningSick = false
			gameengine.RegisterReplacementsForPermanent(gs, tokenPerm)
			gameengine.FirePermanentETBTriggers(gs, tokenPerm)
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"copied": best.Card.DisplayName(),
		"tokens": count,
	})
}

func cardDisp(c *gameengine.Card) string {
	if c == nil {
		return ""
	}
	return c.DisplayName()
}
