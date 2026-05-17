package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSenuKeenEyedProtector wires Senu, Keen-Eyed Protector
// (Muninn parser-gap #125, ~28 hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	Senu, Keen-Eyed Protector — {1}{W} Legendary Creature — Bird Scout 2/2
//	Flying, vigilance
//	{T}, Exile Senu: You gain 2 life and scry 2.
//	When a legendary creature you control attacks and isn't blocked,
//	if this card is exiled, put it onto the battlefield attacking.
//
// Implementation:
//   - OnActivated index 0: tap, exile self, gain 2 life. The scry 2
//     half emits partial — engine has no programmatic scry chooser hook
//     that per_card can drive without the AST decision interface.
//   - The return-from-exile rider is a phase trigger on a card sitting
//     in exile. Same dispatch gap as Ichorid's graveyard upkeep, plus
//     "isn't blocked" requires combat-result inspection — partial.
func registerSenuKeenEyedProtector(r *Registry) {
	r.OnActivated("Senu, Keen-Eyed Protector", senuActivate)
	r.OnETB("Senu, Keen-Eyed Protector", senuETB)
}

func senuETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emit(gs, "senu_etb", "Senu, Keen-Eyed Protector", map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, "senu", "Senu, Keen-Eyed Protector",
		"return_from_exile_when_legendary_attacks_unblocked_requires_exile_phase_trigger_dispatch")
}

func senuActivate(gs *gameengine.GameState, src *gameengine.Permanent, idx int, ctx map[string]interface{}) {
	const slug = "senu_tap_exile_self"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, "Senu, Keen-Eyed Protector", "already_tapped", nil)
		return
	}
	src.Tapped = true
	gameengine.ExilePermanent(gs, src, src)
	gameengine.GainLife(gs, seat, 2, "Senu, Keen-Eyed Protector")
	emit(gs, slug, "Senu, Keen-Eyed Protector", map[string]interface{}{
		"seat":        seat,
		"life_gained": 2,
	})
	emitPartial(gs, slug, "Senu, Keen-Eyed Protector",
		"scry_2_requires_engine_scry_decision_hook")
}
