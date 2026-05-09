package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKalamaxTheStormsireCustom wires the first-instant-each-turn
// copy trigger. The auto-generated stub registerKalamaxTheStormsire in
// the matching gen_*.go remains in place — both handlers fire (its body
// only emits a partial).
//
// Oracle text (Ikoria Commander 2020 / C20, {1}{U}{R}{G}):
//
//	Kalamax has double strike as long as it's tapped.
//	Whenever you cast your first instant spell each turn, if Kalamax
//	is tapped, copy that spell. You may choose new targets for the
//	copy.
//
// Implementation:
//   - "instant_or_sorcery_cast" trigger gated on caster == controller +
//     card has type "instant" + Kalamax tapped + first-instant-this-turn
//     flag not yet set.
//   - Copy mirrors alania.go: locate the cast spell's StackItem, deep-
//     copy, mark IsCopy, push above the original. New-target choice is
//     not modeled (defaults to original targets).
//   - "double strike while tapped" is a static — handled by the AST /
//     conditional-keyword pipeline. emitPartial flags the gap at trigger
//     time so audits see Kalamax is partially modeled.
func registerKalamaxTheStormsireCustom(r *Registry) {
	r.OnTrigger("Kalamax, the Stormsire", "instant_or_sorcery_cast", kalamaxFirstInstantCopy)
}

func kalamaxFirstInstantCopy(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kalamax_the_stormsire_first_instant_copy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	// Only instants — sorceries don't trigger Kalamax.
	if !cardHasType(card, "instant") {
		return
	}
	// Kalamax must be tapped.
	if !perm.Tapped {
		emitFail(gs, slug, perm.Card.DisplayName(), "kalamax_not_tapped", nil)
		return
	}
	// First-instant-this-turn gate.
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["kalamax_first_instant_turn"] == gs.Turn {
		return
	}
	perm.Flags["kalamax_first_instant_turn"] = gs.Turn

	// Locate the spell's StackItem.
	var stackItem *gameengine.StackItem
	for i := len(gs.Stack) - 1; i >= 0; i-- {
		si := gs.Stack[i]
		if si == nil || si.Card != card {
			continue
		}
		stackItem = si
		break
	}
	if stackItem == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "spell_not_on_stack", map[string]interface{}{
			"spell": card.DisplayName(),
		})
		return
	}

	copyCard := card.DeepCopy()
	copyCard.IsCopy = true
	copyItem := &gameengine.StackItem{
		Controller: perm.Controller,
		Card:       copyCard,
		Effect:     stackItem.Effect,
		Kind:       stackItem.Kind,
		IsCopy:     true,
	}
	if len(stackItem.Targets) > 0 {
		copyItem.Targets = append([]gameengine.Target(nil), stackItem.Targets...)
	}
	gameengine.PushStackItem(gs, copyItem)

	gs.LogEvent(gameengine.Event{
		Kind:   "copy_spell",
		Seat:   perm.Controller,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"slug":   slug,
			"copied": card.DisplayName(),
			"rule":   "707.2",
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"spell": card.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"double_strike_while_tapped_static_handled_by_ast_pipeline_only")
}
