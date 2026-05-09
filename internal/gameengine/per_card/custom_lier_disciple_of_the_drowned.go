package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerLierDiscipleOfTheDrownedCustom marks instants and sorceries
// the controller casts as uncounterable. The auto-generated stub
// registerLierDiscipleOfTheDrowned in gen_lier_disciple_of_the_drowned.go
// remains an inert breadcrumb.
//
// Oracle text (Innistrad: Midnight Hunt, {3}{U}{U}):
//
//	Flash
//	Instant and sorcery spells you cast can't be countered.
//	You may cast instant and sorcery cards from your graveyard.
//	If an instant or sorcery card would be put into your graveyard
//	from anywhere, exile it instead.
//
// Implementation:
//   - OnTrigger("instant_or_sorcery_cast"): the engine fires this when
//     an instant/sorcery hits the stack. We retrieve the stack item via
//     ctx["stack_item"] and stamp CostMeta["cannot_be_countered"] when
//     the caster is Lier's controller. Counterspell handlers honor this
//     flag (see Dovin's Veto / Force of Will / etc.).
//   - The graveyard-exile replacement and graveyard-cast permission are
//     replacement / continuous effects that require engine pipeline
//     work; per_card hooks aren't on those code paths. emitPartial.
//   - Flash is granted by the AST keyword pipeline.
func registerLierDiscipleOfTheDrownedCustom(r *Registry) {
	r.OnTrigger("Lier, Disciple of the Drowned", "instant_or_sorcery_cast", lierUncounterableMark)
}

func lierUncounterableMark(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "lier_disciple_of_the_drowned_uncounterable_mark"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	item, _ := ctx["stack_item"].(*gameengine.StackItem)
	if item == nil {
		// Try alternate ctx keys some emitters use.
		item, _ = ctx["item"].(*gameengine.StackItem)
	}
	if item != nil {
		if item.CostMeta == nil {
			item.CostMeta = map[string]interface{}{}
		}
		item.CostMeta["cannot_be_countered"] = true
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"marked": true,
		})
	} else {
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"stack_item_not_in_trigger_ctx_uncounterable_flag_skipped")
	}
	// The graveyard-cast permission and graveyard-exile replacement are
	// not implementable through the per-card trigger path.
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"flashback_grant_and_graveyard_exile_replacement_not_modeled")
}
