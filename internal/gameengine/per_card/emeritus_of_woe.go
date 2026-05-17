package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEmeritusOfWoe wires Emeritus of Woe // Demonic Tutor
// (Muninn parser-gap #82, ~7.2K hits).
//
// Oracle text (Scryfall, verified 2026-05-17):
//
//	Emeritus of Woe — {3}{B} Creature — Vampire Warlock 4/4
//	This creature enters prepared. (While it's prepared, you may cast
//	a copy of its spell. Doing so unprepares it.)
//	At the beginning of your end step, if two or more creatures died
//	this turn, this creature becomes prepared.
//
//	Demonic Tutor (back face) — {1}{B} Sorcery
//	Search your library for a card, put that card into your hand, then
//	shuffle.
//
// Implementation:
//   - ETB: stamp prepared flag and emit partial — the "cast a copy of
//     its spell" mechanic is a cast-pipeline concern (sneak-cast of the
//     back-face sorcery), not modeled per-card.
//   - end_step: if active_seat == controller and Turn.CreaturesDied >= 2,
//     re-stamp prepared flag.
func registerEmeritusOfWoe(r *Registry) {
	r.OnETB("Emeritus of Woe", emeritusOfWoeETB)
	r.OnTrigger("Emeritus of Woe", "end_step", emeritusOfWoeEndStep)
}

func emeritusOfWoeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["prepared"] = 1
	emit(gs, "emeritus_of_woe_etb_prepared", "Emeritus of Woe", map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, "emeritus_of_woe", "Emeritus of Woe",
		"prepared_cast_copy_of_back_face_demonic_tutor_requires_cast_pipeline_hook")
}

func emeritusOfWoeEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "emeritus_of_woe_end_step"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	died := gs.Seats[perm.Controller].Turn.CreaturesDied
	if died < 2 {
		emit(gs, slug, "Emeritus of Woe", map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"died":      died,
		})
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["prepared"] = 1
	emit(gs, slug, "Emeritus of Woe", map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"died":      died,
	})
}
