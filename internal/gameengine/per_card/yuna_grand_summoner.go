package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYunaGrandSummoner wires Yuna, Grand Summoner.
//
// Oracle text:
//
//	Grand Summon — {T}: Add one mana of any color. When you next cast a
//	creature spell this turn, that creature enters with two additional
//	+1/+1 counters on it.
//	Whenever another permanent you control is put into a graveyard from
//	the battlefield, if it had one or more counters on it, you may put
//	that number of +1/+1 counters on target creature.
//
// Implementation:
//   - "permanent_ltb" gated on dying perm having any counters and same
//     controller: pick highest-power other creature we control as the
//     target and copy that count of +1/+1 counters.
//   - Activated mana ability + "next creature spell gets +1/+1 +1/+1"
//     trigger requires deferred-cast tracking; emitPartial.
func registerYunaGrandSummoner(r *Registry) {
	r.OnTrigger("Yuna, Grand Summoner", "permanent_ltb", yunaPermLTB)
	r.OnActivated("Yuna, Grand Summoner", yunaActivate)
}

func yunaActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "yuna_grand_summon_mana"
	if gs == nil || src == nil {
		return
	}
	if src.Tapped {
		return
	}
	src.Tapped = true
	emitPartial(gs, slug, src.Card.DisplayName(), "deferred_creature_spell_counter_bonus_not_implemented")
}

func yunaPermLTB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "yuna_perm_with_counters_dies"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dest, _ := ctx["to_zone"].(string)
	if dest != "graveyard" {
		return
	}
	dyingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if dyingPerm == nil || dyingPerm == perm {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	totalCounters := 0
	if dyingPerm.Counters != nil {
		for _, n := range dyingPerm.Counters {
			if n > 0 {
				totalCounters += n
			}
		}
	}
	if totalCounters <= 0 {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	var best *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if best == nil || p.Power() > best.Power() {
			best = p
		}
	}
	if best == nil {
		return
	}
	best.AddCounter("+1/+1", totalCounters)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"target":   best.Card.DisplayName(),
		"counters": totalCounters,
	})
}
