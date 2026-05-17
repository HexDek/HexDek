package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTitaniaVoiceOfGaea wires Titania, Voice of Gaea (Muninn
// parser-gap #45, ~18K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{G}{G}
//	Legendary Creature — Elemental
//	Reach
//	Whenever one or more land cards are put into your graveyard from
//	anywhere, you gain 2 life.
//	At the beginning of your upkeep, if there are four or more land
//	cards in your graveyard and you both own and control Titania, Voice
//	of Gaea and a land named Argoth, Sanctum of Nature, exile them,
//	then meld them into Titania, Gaea Incarnate.
//
// Implementation:
//   - Reach is AST-side.
//   - Lifegain trigger: "land_to_graveyard" (engine fires once per land
//     entering any graveyard from any zone — zone_change.go:506). Gate on
//     ctx["owner_seat"] == perm.Controller and gain 2 life via GainLife.
//     The printed text says "one or more land cards" — the trigger fires
//     once per batch, but the engine fires it per-land. We clamp to
//     one trigger per game-state tick via a per-permanent flag stamped
//     with the current event-log length, so a batch of N lands hitting
//     in one resolve only grants 2 life total.
//   - Meld upkeep clause: not implemented — Titania, Gaea Incarnate is a
//     meld card with no parser support yet. emitPartial.
func registerTitaniaVoiceOfGaea(r *Registry) {
	r.OnTrigger("Titania, Voice of Gaea", "land_to_graveyard", titaniaVoiceLandToGraveyard)
}

func titaniaVoiceLandToGraveyard(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "titania_voice_of_gaea_lifegain"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	ownerSeat, _ := ctx["owner_seat"].(int)
	if ownerSeat != perm.Controller {
		return
	}
	// Per-tick dedup so batched lands (e.g., multi-land sac, Armageddon
	// followed by graveyard rules) only grant 2 life once per stack item.
	stackDepth := len(gs.Stack)
	tickKey := "titania_voice_tick"
	depthKey := "titania_voice_tick_depth"
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags[tickKey] == gs.Turn+1 && perm.Flags[depthKey] == stackDepth {
		return
	}
	perm.Flags[tickKey] = gs.Turn + 1
	perm.Flags[depthKey] = stackDepth
	gameengine.GainLife(gs, perm.Controller, 2, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"life": 2,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"upkeep_meld_clause_into_titania_gaea_incarnate_unimplemented")
}
