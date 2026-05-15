package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBumiUnleashed wires Bumi, Unleashed.
//
// Oracle text (Scryfall, verified 2026-05-14):
//
//	{3}{R}{G}
//	Legendary Creature — Human Noble Ally
//	Trample
//	When Bumi enters, earthbend 4.
//	Whenever Bumi deals combat damage to a player, untap all lands
//	you control. After this phase, there is an additional combat
//	phase. Only land creatures can attack during that combat phase.
//
// Engine handling:
//
//   - ETB earthbend 4: handled by the AST engine via the Earthbend
//     keyword/effect grammar. No per_card hook needed for that part.
//   - Combat damage trigger: this handler.
//     1. Untap all lands controlled by Bumi's controller.
//     2. Queue an extra combat phase with the "land_creatures_only"
//        restriction, so the new combat's DeclareAttackers step
//        filters its legal-attacker pool to permanents that are
//        BOTH a Land AND a Creature (Toph's artifact-as-land
//        permanents while earthbent, manlands, Dryad Arbor, etc.).
//     3. An OnBegin hook on that extra combat re-untaps lands the
//        same way (covers the case where lands get tapped between
//        the trigger's resolution and the extra combat starting —
//        most often by activated mana abilities the AI spent in the
//        main phase between damage step and the extra combat).
//
// This handler uses the typed extra-combat queue introduced in
// "engine: typed extra-combat queue with restrictions + OnBegin hooks"
// (PendingExtraCombat struct + Restriction tag + OnBegin closure).
func registerBumiUnleashed(r *Registry) {
	r.OnTrigger("Bumi, Unleashed", "combat_damage_player", bumiUnleashedCombatDamage)
}

func bumiUnleashedCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "bumi_unleashed_extra_combat"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	// Gate to Bumi's own combat damage to a player (not ally-attack
	// triggers, and not damage from other sources owned by Bumi's
	// controller). The combat_damage_player trigger context carries
	// source_card / source_seat — both must match Bumi specifically.
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != "" && sourceName != perm.Card.DisplayName() {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	controller := perm.Controller
	if controller < 0 || controller >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[controller]
	if seat == nil {
		return
	}

	// Untap all lands you control (immediate effect of the trigger).
	untapped := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if !p.IsLand() {
			continue
		}
		if p.Tapped {
			p.Tapped = false
			untapped++
		}
	}

	// Queue the extra combat phase with the land-only restriction.
	// OnBegin hook re-untaps lands at the start of the new combat to
	// handle the gap-between-trigger-and-extra-combat case (lands
	// tapped for activated mana abilities during the intervening
	// stack drain or main-phase land/spell activity).
	owner := controller
	gs.AddExtraCombat(gameengine.PendingExtraCombat{
		Restriction: "land_creatures_only",
		SourceCard:  perm.Card.DisplayName(),
		OnBegin: func(g *gameengine.GameState) {
			if g == nil || owner < 0 || owner >= len(g.Seats) {
				return
			}
			s := g.Seats[owner]
			if s == nil {
				return
			}
			for _, p := range s.Battlefield {
				if p == nil || p.Card == nil {
					continue
				}
				if p.IsLand() && p.Tapped {
					p.Tapped = false
				}
			}
		},
	})

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":             controller,
		"damage":           amount,
		"lands_untapped":   untapped,
		"extra_combats":    len(gs.PendingExtraCombats),
		"restriction":      "land_creatures_only",
	})
}
