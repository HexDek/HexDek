package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSyrVondamSunstarExemplar wires Syr Vondam, Sunstar Exemplar
// (distinct from Syr Vondam, the Lucent — different printing).
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{W}{B}
//	Legendary Creature — Human Knight
//	2/2
//	Vigilance, menace
//	Whenever another creature you control dies or is put into exile,
//	put a +1/+1 counter on Syr Vondam and you gain 1 life.
//	When Syr Vondam dies or is put into exile while its power is 4 or
//	greater, destroy up to one target nonland permanent.
//
// Implementation:
//   - "creature_dies" gated on controller_seat == perm.Controller and
//     dying card != Syr Vondam himself: +1/+1 counter + 1 life.
//   - "card_exiled" same gate (per_card uses card_exiled for permanent
//     leaves to exile from the battlefield).
//   - Self dies/exile: if Syr Vondam's power was 4+ at death, destroy
//     the highest-CMC opponent nonland permanent.
func registerSyrVondamSunstarExemplar(r *Registry) {
	r.OnTrigger("Syr Vondam, Sunstar Exemplar", "creature_dies", syrVondamSunstarOtherDies)
	r.OnTrigger("Syr Vondam, Sunstar Exemplar", "card_exiled", syrVondamSunstarOtherExiled)
	r.OnTrigger("Syr Vondam, Sunstar Exemplar", "dies", syrVondamSunstarSelfDies)
}

func syrVondamSunstarOtherDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	syrVondamSunstarOtherCreatureLeft(gs, perm, ctx, "syr_vondam_sunstar_other_died")
}

func syrVondamSunstarOtherExiled(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	syrVondamSunstarOtherCreatureLeft(gs, perm, ctx, "syr_vondam_sunstar_other_exiled")
}

func syrVondamSunstarOtherCreatureLeft(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}, slug string) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil || card == perm.Card {
		return
	}
	if !cardHasType(card, "creature") {
		return
	}
	perm.AddCounter("+1/+1", 1)
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"other": card.DisplayName(),
	})
}

func syrVondamSunstarSelfDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "syr_vondam_sunstar_self_dies_destroy"
	if gs == nil || perm == nil {
		return
	}
	if perm.Power() < 4 {
		return
	}
	// Find the most expensive opponent nonland permanent.
	var target *gameengine.Permanent
	bestCMC := -1
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || p.IsLand() {
				continue
			}
			cmc := gameengine.ManaCostOf(p.Card)
			if cmc > bestCMC {
				bestCMC = cmc
				target = p
			}
		}
	}
	if target == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_destroy_target", nil)
		return
	}
	targetName := target.Card.DisplayName()
	gameengine.SacrificePermanent(gs, target, "syr_vondam_destroy")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"destroyed": targetName,
	})
}
