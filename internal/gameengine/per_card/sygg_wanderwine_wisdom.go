package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSyggWanderwineWisdom wires Sygg, Wanderwine Wisdom (DFC front).
//
// Oracle text (front: Sygg, Wanderwine Wisdom):
//
//	{1}{U}
//	Legendary Creature — Merfolk Wizard
//	Sygg can't be blocked.
//	Whenever this creature enters or transforms into Sygg, Wanderwine
//	  Wisdom, target creature gains "Whenever this creature deals combat
//	  damage to a player or planeswalker, draw a card" until end of turn.
//	At the beginning of your first main phase, you may pay {W}. If you
//	  do, transform Sygg.
//
// Back face (Sygg, Wanderbrine Shield):
//
//	Sygg can't be blocked.
//	Whenever this creature transforms into Sygg, Wanderbrine Shield,
//	  target creature you control gains protection from each color
//	  until your next turn.
//	At the beginning of your first main phase, you may pay {U}. If you
//	  do, transform Sygg.
//
// Implementation:
//   - "Can't be blocked" — set Permanent.Flags["unblockable"] = 1 on ETB.
//   - ETB: pick own best attacking creature, mark it with
//     Flags["temp_combat_damage_draw"] = 1 UEOT. The combat-damage
//     pipeline reads this flag to fire a draw on combat damage.
//   - First-main transform clauses: emitPartial.
func registerSyggWanderwineWisdom(r *Registry) {
	r.OnETB("Sygg, Wanderwine Wisdom", syggWWWisdomETB)
	r.OnETB("Sygg, Wanderwine Wisdom // Sygg, Wanderbrine Shield", syggWWWisdomETB)
}

func syggWWWisdomETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "sygg_wanderwine_wisdom_etb"
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["unblockable"] = 1

	target := syggWWPickAttacker(gs, perm)
	targetName := ""
	if target != nil {
		if target.Flags == nil {
			target.Flags = map[string]int{}
		}
		target.Flags["temp_combat_damage_draw"] = 1
		targetName = target.Card.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"target": targetName,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"first_main_phase_optional_transform_partial_and_temp_grant_uses_flag")
}

func syggWWPickAttacker(gs *gameengine.GameState, perm *gameengine.Permanent) *gameengine.Permanent {
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return nil
	}
	var best *gameengine.Permanent
	bestPow := -1
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Card == nil {
			continue
		}
		if !p.IsCreature() {
			continue
		}
		if p.Card.BasePower > bestPow {
			bestPow = p.Card.BasePower
			best = p
		}
	}
	if best != nil {
		return best
	}
	return perm
}
