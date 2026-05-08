package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTamObservantSequencerDeepSight wires Tam, Observant Sequencer // Deep Sight.
//
// Oracle text:
//
//	Front face — Tam, Observant Sequencer ({2}{G}{U}, Legendary Creature —
//	Human Druid, 2/4):
//
//	  Landfall — Whenever a land you control enters, Tam becomes prepared.
//	  (While it's prepared, you may cast a copy of its spell. Doing so
//	  unprepares it.)
//
//	Back face — Deep Sight ({G/U}, Sorcery):
//
//	  You draw a card and gain 1 life.
//
// Implementation:
//   - "permanent_etb" trigger gated on controller_seat == perm.Controller
//     AND the entering permanent being a land (standard landfall pattern).
//   - When triggered: set perm.Prepared = true, then immediately resolve
//     the copy of Deep Sight (draw 1 card + gain 1 life for controller),
//     then call gameengine.Unprepare(perm). This bypasses the literal
//     "may cast a copy" stack interaction; the AI always wants the draw +
//     life. emitPartial flags the simplification.
func registerTamObservantSequencerDeepSight(r *Registry) {
	r.OnTrigger("Tam, Observant Sequencer // Deep Sight", "permanent_etb", tamLandfallPreparedTrigger)
	r.OnTrigger("Tam, Observant Sequencer", "permanent_etb", tamLandfallPreparedTrigger)
}

func tamLandfallPreparedTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tam_observant_sequencer_deep_sight_prepared_copy"
	if gs == nil || perm == nil || ctx == nil {
		return
	}

	// Gate on controller: only lands entering under Tam's controller.
	ctrl, _ := ctx["controller_seat"].(int)
	if ctrl != perm.Controller {
		return
	}

	// Gate on land: the entering permanent must be a land.
	enterPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if enterPerm == nil || !enterPerm.IsLand() {
		return
	}

	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}

	// Mark Tam as prepared.
	perm.Prepared = true
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["prepared"] = 1

	// Resolve Deep Sight copy: draw a card and gain 1 life.
	if len(seat.Library) > 0 {
		card := seat.Library[0]
		gameengine.MoveCard(gs, card, perm.Controller, "library", "hand", "draw")
	}
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())

	// Unprepare after resolving the copy.
	gameengine.Unprepare(perm)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"copied_back": "Deep Sight",
		"drew":        1,
		"life_gained": 1,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"prepare_keyword_resolves_back_face_directly_skipping_stack_and_mana_cost")
}
