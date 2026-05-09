package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAcererakTheArchlichCustom wires the dungeon ETB and the
// combat-damage reanimation activation. The auto-generated stub
// registerAcererakTheArchlich in gen_acererak_the_archlich.go remains
// in place — both handlers fire (its body only emits a partial).
//
// Oracle text (Adventures in the Forgotten Realms Commander, {3}{B}{B}):
//
//	Menace
//	When Acererak the Archlich enters the battlefield, venture into
//	the dungeon. Then if you've completed Tomb of Annihilation, each
//	opponent loses life equal to the number of creature cards in your
//	graveyard.
//	Whenever Acererak the Archlich deals combat damage to a player,
//	you may pay {2}{B}. If you do, return target creature card from
//	your graveyard to the battlefield with two +1/+1 counters on it.
//
// Implementation:
//   - OnETB: VentureIntoDungeon. The full dungeon model is simplified
//     to a 4-room linear track in keywords_misc.go; it doesn't have a
//     "Tomb of Annihilation completion" gate. We use the cumulative
//     "dungeon_completed" counter as a stand-in: if the player has
//     completed at least one dungeon by this ETB (room 4 was hit during
//     this venture or a prior one), drain each opponent for the
//     creature-card count in the controller's graveyard. emitPartial
//     flags the simplification.
//   - OnTrigger("combat_damage_player"): gate on attacker == Acererak
//     herself dealing damage to a player. Greedy: always pay {2}{B}
//     (modeled by emitPartial; mana enforcement isn't on the per-card
//     hook path). Pull the highest-MV creature card from the
//     controller's graveyard back to the battlefield with two +1/+1
//     counters. The "this creature is also a 2/2 black Zombie" rider
//     in the printed wording was errata'd off — current Scryfall text
//     reads "with two +1/+1 counters on it" only.
//   - Menace is granted by the AST keyword pipeline.
func registerAcererakTheArchlichCustom(r *Registry) {
	r.OnETB("Acererak the Archlich", acererakArchlichETB)
	r.OnTrigger("Acererak the Archlich", "combat_damage_player", acererakArchlichDamageReanimate)
}

func acererakArchlichETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "acererak_archlich_etb_venture"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}

	roomReached := gameengine.VentureIntoDungeon(gs, seatIdx)

	// Tomb of Annihilation completion gate — approximated by the cumulative
	// dungeon-completed counter being positive AFTER this venture call.
	dungeonCompleted := 0
	if seat.Flags != nil {
		dungeonCompleted = seat.Flags["dungeon_completed"]
	}
	drain := 0
	if dungeonCompleted > 0 {
		// Count creature cards in the controller's graveyard.
		for _, c := range seat.Graveyard {
			if c == nil {
				continue
			}
			if cardHasType(c, "creature") {
				drain++
			}
		}
		if drain > 0 {
			for i, opp := range gs.Seats {
				if opp == nil || i == seatIdx || opp.Lost {
					continue
				}
				gameengine.LoseLife(gs, i, drain, perm.Card.DisplayName())
			}
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":              seatIdx,
		"room_reached":      roomReached,
		"dungeon_completed": dungeonCompleted,
		"drain":             drain,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"tomb_of_annihilation_specific_dungeon_track_not_modeled_using_completion_count")
	_ = gs.CheckEnd()
}

func acererakArchlichDamageReanimate(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "acererak_archlich_damage_reanimate"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	src, _ := ctx["source"].(*gameengine.Permanent)
	if src == nil {
		// Some emitters use "attacker_perm" / "source_perm".
		if v, ok := ctx["source_perm"].(*gameengine.Permanent); ok {
			src = v
		}
		if src == nil {
			if v, ok := ctx["attacker_perm"].(*gameengine.Permanent); ok {
				src = v
			}
		}
	}
	if src != perm {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}

	// Pick highest-CMC creature card from graveyard.
	bestIdx := -1
	bestCMC := -1
	for i, c := range seat.Graveyard {
		if c == nil {
			continue
		}
		if !cardHasType(c, "creature") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_creature_in_graveyard", nil)
		emitPartial(gs, slug, perm.Card.DisplayName(), "may_pay_2b_cost_not_enforced")
		return
	}

	card := seat.Graveyard[bestIdx]
	gameengine.MoveCard(gs, card, seatIdx, "graveyard", "battlefield", "acererak_reanimate")
	newPerm := createPermanent(gs, seatIdx, card, false)
	if newPerm != nil {
		newPerm.AddCounter("+1/+1", 2)
		gameengine.RegisterReplacementsForPermanent(gs, newPerm)
		gameengine.FirePermanentETBTriggers(gs, newPerm)
	}
	gs.InvalidateCharacteristicsCache()

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         seatIdx,
		"reanimated":   card.DisplayName(),
		"reanimated_cmc": bestCMC,
		"counters":     2,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(), "may_pay_2b_cost_not_enforced")
}
