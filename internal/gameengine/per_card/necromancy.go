package per_card

import (
	"sync"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerNecromancy wires Necromancy.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	You may cast this spell as though it had flash. If you cast it any
//	time a sorcery couldn't have been cast, the controller of the
//	permanent it becomes sacrifices it at the beginning of the next
//	cleanup step.
//	When this enchantment enters, if it's on the battlefield, it becomes
//	an Aura with "enchant creature put onto the battlefield with
//	Necromancy."
//	Put target creature card from a graveyard onto the battlefield under
//	your control and attach this enchantment to it.
//	When this enchantment leaves the battlefield, that creature's
//	controller sacrifices it.
//
// Implementation (Muninn gap #3 — 237,901 hits):
//   - OnETB: scan every seat's graveyard for the highest-CMC creature
//     card (prefer our own graveyard first), MoveCard it to our
//     battlefield, attach this Necromancy to the new permanent, and
//     record the pair in necromancyTargets for the LTB-sacrifice half.
//   - Flash and "sac at next cleanup if cast off-curve" are upstream
//     concerns (cast-time gates, cleanup-step delayed triggers) that the
//     general engine doesn't model granularly enough yet — emitPartial.
//   - LTB-sacrifice: matches the Animate Dead status quo (also
//     unimplemented for self-LTB because fireTrigger iterates the
//     current battlefield, not the leaving permanent). emitPartial so
//     Muninn tracks the residual gap.
func registerNecromancy(r *Registry) {
	r.OnETB("Necromancy", necromancyETB)
}

// necromancyTargets maps a Necromancy *Permanent to the creature *Permanent
// it animated. Lives alongside animateDeadTargets so any future LTB-sweep
// pass can consume both ledgers uniformly.
var necromancyTargets sync.Map

func necromancyETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "necromancy_reanimate"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	// Pick the highest-CMC creature card across all graveyards. Prefer
	// the controller's own graveyard on ties so cEDH self-reanimation
	// (Razaketh / Griselbrand / Worldspine Wurm) keeps the body local.
	var bestCard *gameengine.Card
	bestCMC := -1
	bestSeat := -1
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, c := range s.Graveyard {
			if c == nil || !cardHasType(c, "creature") {
				continue
			}
			cmc := gameengine.ManaCostOf(c)
			if cmc > bestCMC {
				bestCMC = cmc
				bestCard = c
				bestSeat = i
				continue
			}
			if cmc == bestCMC && bestSeat != seat && i == seat {
				bestCard = c
				bestSeat = i
			}
		}
	}
	if bestCard == nil {
		emitFail(gs, slug, "Necromancy", "no_creature_in_any_graveyard", map[string]interface{}{
			"seat": seat,
		})
		return
	}

	gameengine.MoveCard(gs, bestCard, bestSeat, "graveyard", "battlefield", "necromancy")

	// Locate the new permanent. MoveCard places it on the OWNER's
	// battlefield first; Necromancy's "under your control" clause then
	// moves control to the Necromancy controller.
	var creature *gameengine.Permanent
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p != nil && p.Card == bestCard {
				creature = p
				break
			}
		}
		if creature != nil {
			break
		}
	}
	if creature == nil {
		emitFail(gs, slug, "Necromancy", "reanimated_card_not_on_battlefield", map[string]interface{}{
			"seat":       seat,
			"reanimated": bestCard.DisplayName(),
		})
		return
	}

	// Move control to Necromancy's controller per "under your control".
	if creature.Controller != seat {
		oldSeat := creature.Controller
		if oldSeat >= 0 && oldSeat < len(gs.Seats) && gs.Seats[oldSeat] != nil {
			bf := gs.Seats[oldSeat].Battlefield
			for i, p := range bf {
				if p == creature {
					gs.Seats[oldSeat].Battlefield = append(bf[:i], bf[i+1:]...)
					break
				}
			}
		}
		creature.Controller = seat
		gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, creature)
	}

	// Attach Necromancy to the reanimated creature.
	perm.AttachedTo = creature
	necromancyTargets.Store(perm, creature)

	emit(gs, slug, "Necromancy", map[string]interface{}{
		"seat":       seat,
		"reanimated": bestCard.DisplayName(),
		"from_seat":  bestSeat,
		"cmc":        bestCMC,
	})

	emitPartial(gs, slug, "Necromancy",
		"flash_off_curve_cleanup_sacrifice_and_aura_ltb_creature_sacrifice_clauses_not_modeled")
}
