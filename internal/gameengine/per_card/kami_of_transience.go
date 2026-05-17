package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKamiOfTransience wires Kami of Transience (Muninn parser-gap
// #76, ~8.3K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{G}
//	Creature — Spirit
//	Trample
//	Whenever you cast an enchantment spell, put a +1/+1 counter on this
//	creature.
//	At the beginning of each end step, if an enchantment was put into
//	your graveyard from the battlefield this turn, you may return this
//	card from your graveyard to your hand.
//
// Implementation:
//   - Trample is AST-side.
//   - "spell_cast": gate on caster_seat == controller AND the cast card
//     being an enchantment (Kami is itself a creature, not an enchantment,
//     so it doesn't double-up on self-cast). Stack a +1/+1 counter and
//     invalidate the characteristics cache for SBA P/T recompute.
//   - End-step graveyard-return clause: parser doesn't yet have a clean
//     "enchantment_LTB_to_graveyard_this_turn" tracker on a graveyard
//     card. emitPartial; out-of-scope for the parser-gap handler.
func registerKamiOfTransience(r *Registry) {
	r.OnTrigger("Kami of Transience", "spell_cast", kamiOfTransienceSpellCast)
}

func kamiOfTransienceSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kami_of_transience_enchantment_cast_counter"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil || !cardHasType(card, "enchantment") {
		return
	}
	perm.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"spell":    card.DisplayName(),
		"counters": perm.Counters["+1/+1"],
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"end_step_self_recursion_on_enchantment_ltb_unimplemented")
}
