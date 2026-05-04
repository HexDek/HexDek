package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRexCyberHound wires Rex, Cyber-Hound.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever Rex deals combat damage to a player, they mill two cards
//	  and you get {E}{E} (two energy counters).
//	Pay {E}{E}: Choose target creature card in a graveyard. Exile it
//	  with a brain counter on it. Activate only as a sorcery.
//	Rex has all activated abilities of all cards in exile with brain
//	  counters on them.
//
// Implementation:
//   - "combat_damage_player": gate on source_perm == perm. Defender
//     mills 2; controller gains 2 energy counters via seat.EnergyCounters
//     if available, else stash on perm.Flags.
//   - Activated brain-counter exile and the dynamic ability adoption are
//     stack-aware features outside the per-card layer; emitPartial.
func registerRexCyberHound(r *Registry) {
	r.OnTrigger("Rex, Cyber-Hound", "combat_damage_player", rexCyberHoundDamage)
	r.OnActivated("Rex, Cyber-Hound", rexCyberHoundActivate)
}

func rexCyberHoundDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rex_cyber_hound_combat_mill_energy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	src, _ := ctx["source_perm"].(*gameengine.Permanent)
	if src != perm {
		return
	}
	tgt, _ := ctx["target_seat"].(int)
	if tgt < 0 || tgt >= len(gs.Seats) {
		return
	}
	defender := gs.Seats[tgt]
	if defender == nil {
		return
	}
	milled := 0
	for i := 0; i < 2 && len(defender.Library) > 0; i++ {
		c := defender.Library[0]
		gameengine.MoveCard(gs, c, tgt, "library", "graveyard", "rex_mill")
		milled++
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["rex_energy"] += 2
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"target_seat":  tgt,
		"milled":       milled,
		"energy_total": perm.Flags["rex_energy"],
	})
}

func rexCyberHoundActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, "rex_cyber_hound_brain_counter", src.Card.DisplayName(),
		"brain_counter_exile_and_ability_adoption_not_modeled")
}
