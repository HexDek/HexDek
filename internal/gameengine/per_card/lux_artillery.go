package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLuxArtillery wires Lux Artillery (Muninn parser-gap #26, 43,952 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{4}
//	Artifact
//	Whenever you cast an artifact creature spell, it gains sunburst.
//	(It enters with a +1/+1 counter on it for each color of mana spent
//	to cast it.)
//	At the beginning of your end step, if there are thirty or more
//	counters among artifacts and creatures you control, this artifact
//	deals 10 damage to each opponent.
//
// Implementation:
//   - "spell_cast" gated on caster_seat == controller AND artifact
//     creature. The grant-sunburst clause can't be applied at ETB
//     without an engine hook; we stamp a seat-level flag so analysis
//     tooling sees the active grant, and log via emitPartial.
//   - "end_step" gated on active_seat == controller. Sum all counters
//     across the controller's artifacts and creatures; if >=30, deal
//     10 to each opponent.
func registerLuxArtillery(r *Registry) {
	r.OnTrigger("Lux Artillery", "spell_cast", luxArtillerySpellCast)
	r.OnTrigger("Lux Artillery", "end_step", luxArtilleryEndStep)
}

func luxArtillerySpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lux_artillery_grant_sunburst"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if !cardHasType(card, "artifact") || !cardHasType(card, "creature") {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat != nil {
		if seat.Flags == nil {
			seat.Flags = map[string]int{}
		}
		seat.Flags["lux_grants_sunburst_active"]++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"spell": card.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"sunburst_grant_requires_etb_pipeline_hook_not_yet_wired")
}

func luxArtilleryEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lux_artillery_end_step_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	total := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "artifact") && !p.IsCreature() {
			continue
		}
		if p.Counters == nil {
			continue
		}
		for _, n := range p.Counters {
			if n > 0 {
				total += n
			}
		}
	}
	if total < 30 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"counters":  total,
			"triggered": false,
		})
		return
	}

	hit := 0
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		gameengine.LoseLife(gs, opp, 10, perm.Card.DisplayName())
		hit++
	}
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"counters":  total,
		"opps_hit":  hit,
		"damage":    10,
		"triggered": true,
	})
}
