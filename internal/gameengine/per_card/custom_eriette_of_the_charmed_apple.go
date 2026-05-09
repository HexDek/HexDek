package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerErietteCustom adds Eriette's end-step drain trigger that the
// auto-generated static stub omits.
//
// Oracle text:
//
//	Each creature that's enchanted by an Aura you control can't attack
//	you or planeswalkers you control.
//	At the beginning of your end step, each opponent loses X life and
//	you gain X life, where X is the number of Auras you control.
//
// The combat-ban static is engine-side. The end-step drain re-registers
// itself as a delayed trigger so it fires every end step Eriette is on
// the battlefield.
func registerErietteCustom(r *Registry) {
	r.OnETB("Eriette of the Charmed Apple", erietteScheduleEndStep)
}

func erietteScheduleEndStep(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	scheduleErietteEndStep(gs, perm.Controller)
}

func scheduleErietteEndStep(gs *gameengine.GameState, seat int) {
	if gs == nil || seat < 0 || seat >= len(gs.Seats) {
		return
	}
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "your_next_end_step",
		ControllerSeat: seat,
		SourceCardName: "Eriette of the Charmed Apple",
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			erietteEndStepDrain(gs, seat)
		},
	})
}

func erietteEndStepDrain(gs *gameengine.GameState, seat int) {
	const slug = "eriette_end_step_drain"
	if gs == nil || seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	// Eriette must still be on the battlefield to fire.
	stillOn := false
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if normalizeName(p.Card.DisplayName()) == normalizeName("Eriette of the Charmed Apple") {
			stillOn = true
			break
		}
	}
	if !stillOn {
		return
	}
	x := 0
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "aura") {
			x++
		}
	}
	if x == 0 {
		// Still re-schedule so future auras are picked up.
		scheduleErietteEndStep(gs, seat)
		emit(gs, slug, "Eriette of the Charmed Apple", map[string]interface{}{
			"seat": seat,
			"x":    0,
			"note": "no_auras_in_play",
		})
		return
	}
	for i := range gs.Seats {
		if i == seat || gs.Seats[i] == nil || gs.Seats[i].Lost {
			continue
		}
		gameengine.LoseLife(gs, i, x, "Eriette of the Charmed Apple")
	}
	gameengine.GainLife(gs, seat, x, "Eriette of the Charmed Apple")
	emit(gs, slug, "Eriette of the Charmed Apple", map[string]interface{}{
		"seat":      seat,
		"x":         x,
		"life_gain": x,
	})
	scheduleErietteEndStep(gs, seat)
}
