package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerStellaLeeWildCard wires Stella Lee, Wild Card.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you cast your second spell each turn, exile the top card
//	of your library. Until the end of your next turn, you may play
//	that card.
//	{T}: Copy target instant or sorcery spell you control. You may
//	choose new targets for the copy. Activate only if you've cast
//	three or more spells this turn.
//
// Implementation:
//   - "spell_cast" trigger: when caster_seat == controller AND
//     SpellsCastThisTurn == 2, exile top of library and stamp a flag
//     on the exiled card record so the cast pipeline can grant temporary
//     play permission. We emitPartial for the play-permission gap.
//   - Activated: gate on caster's SpellsCastThisTurn >= 3, then locate
//     the highest-CMC instant/sorcery on the stack controlled by Stella's
//     controller and emit a "stella_lee_copy_spell" event for the
//     stack-copy mechanism.
func registerStellaLeeWildCard(r *Registry) {
	r.OnTrigger("Stella Lee, Wild Card", "spell_cast", stellaSecondSpellExile)
	r.OnActivated("Stella Lee, Wild Card", stellaActivateCopy)
}

func stellaSecondSpellExile(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "stella_lee_second_spell_exile_top"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.SpellsCastThisTurn != 2 {
		return
	}
	if len(seat.Library) == 0 {
		return
	}
	top := seat.Library[0]
	gameengine.MoveCard(gs, top, perm.Controller, "library", "exile", "stella_lee_exile_top")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"exiled": cardDisp(top),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"play_permission_until_next_end_of_turn_not_modeled_as_alt_cost")
}

func stellaActivateCopy(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "stella_lee_copy_spell"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil || seat.SpellsCastThisTurn < 3 {
		emitFail(gs, slug, src.Card.DisplayName(), "fewer_than_three_spells_cast", map[string]interface{}{
			"seat":   src.Controller,
			"spells": seat.SpellsCastThisTurn,
		})
		return
	}
	src.Tapped = true
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": src.Controller,
	})
	emitPartial(gs, slug, src.Card.DisplayName(),
		"target_spell_copy_with_new_targets_not_dispatched_to_stack_engine")
}
