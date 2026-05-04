package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerThaliaAndTheGitrogMonster wires Thalia and The Gitrog Monster.
//
// Oracle text:
//
//	{1}{W}{B}{G}
//	Legendary Creature — Human Frog Horror
//	First strike, deathtouch
//	You may play an additional land on each of your turns.
//	Creatures and nonbasic lands your opponents control enter tapped.
//	Whenever Thalia and The Gitrog Monster attacks, sacrifice a creature
//	  or land, then draw a card.
//
// Implementation:
//   - First strike + deathtouch via AST keywords.
//   - ETB: grant +1 extra land drop (seat.Flags["extra_land_drops"]) and
//     log a parser_gap noting the static "opponent creatures and nonbasic
//     lands ETB tapped" effect (engine layers system handles ETB-tapped
//     replacements; we don't have a clean per_card seam for that).
//   - "creature_attacks" trigger gated to attacker == perm: pick the
//     least-valuable creature or land we control to sacrifice (preferring
//     a tapped basic land over an untapped creature), sac it, then draw.
func registerThaliaAndTheGitrogMonster(r *Registry) {
	r.OnETB("Thalia and The Gitrog Monster", thaliaAndGitrogETB)
	r.OnTrigger("Thalia and The Gitrog Monster", "creature_attacks", thaliaAndGitrogAttack)
}

func thaliaAndGitrogETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "thalia_and_gitrog_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["extra_land_drops"]++
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"static_opponent_creatures_and_nonbasic_lands_enter_tapped_partial")
}

func thaliaAndGitrogAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "thalia_and_gitrog_attack_sac_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atkPerm, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atkPerm == nil || atkPerm != perm {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}

	target := pickWorstSacForThaliaGitrog(seat, perm)
	saccedName := ""
	if target != nil {
		saccedName = target.Card.DisplayName()
		gameengine.SacrificePermanent(gs, target, "thalia_and_gitrog_attack_cost")
	}
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"sacrificed": saccedName,
	})
}

func pickWorstSacForThaliaGitrog(seat *gameengine.Seat, src *gameengine.Permanent) *gameengine.Permanent {
	if seat == nil {
		return nil
	}
	// Prefer a tapped basic land (cheapest), then any tapped land,
	// then a tapped non-token creature, then any creature/land.
	for _, p := range seat.Battlefield {
		if p == nil || p == src || p.Card == nil {
			continue
		}
		if p.IsLand() && cardHasType(p.Card, "basic") && p.Tapped {
			return p
		}
	}
	for _, p := range seat.Battlefield {
		if p == nil || p == src || p.Card == nil {
			continue
		}
		if p.IsLand() && p.Tapped {
			return p
		}
	}
	for _, p := range seat.Battlefield {
		if p == nil || p == src || p.Card == nil {
			continue
		}
		if p.IsCreature() && p.Tapped {
			return p
		}
	}
	for _, p := range seat.Battlefield {
		if p == nil || p == src || p.Card == nil {
			continue
		}
		if p.IsLand() || p.IsCreature() {
			return p
		}
	}
	return nil
}
