package per_card

import "github.com/hexdek/hexdek/internal/gameengine"

// Aristocrat death-trigger cards: whenever a creature dies, drain opponents.

func registerBloodArtist(r *Registry) {
	r.OnTrigger("Blood Artist", "creature_dies", bloodArtistTrigger)
}

func bloodArtistTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Target opponent loses 1, you gain 1.
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == seat {
			continue
		}
		gameengine.LoseLife(gs, i, 1, "Blood Artist")
		break // "target opponent" — pick first alive opponent
	}
	gameengine.GainLife(gs, seat, 1, "Blood Artist")
}

func registerZulaportCutthroat(r *Registry) {
	r.OnTrigger("Zulaport Cutthroat", "creature_dies", zulaportTrigger)
}

func zulaportTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != seat {
		return
	}
	// Each opponent loses 1 life, you gain 1 life.
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == seat {
			continue
		}
		gameengine.LoseLife(gs, i, 1, "Zulaport Cutthroat")
	}
	gameengine.GainLife(gs, seat, 1, "Zulaport Cutthroat")
}

func registerBastionOfRemembrance(r *Registry) {
	r.OnTrigger("Bastion of Remembrance", "creature_dies", bastionTrigger)
}

func bastionTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Only triggers on creatures YOU control dying.
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != seat {
		return
	}
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == seat {
			continue
		}
		gameengine.LoseLife(gs, i, 1, "Bastion of Remembrance")
	}
	gameengine.GainLife(gs, seat, 1, "Bastion of Remembrance")
}

func registerCruelCelebrant(r *Registry) {
	r.OnTrigger("Cruel Celebrant", "creature_dies", cruelCelebrantTrigger)
}

func cruelCelebrantTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Only on creatures/planeswalkers YOU control dying.
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != seat {
		return
	}
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == seat {
			continue
		}
		gameengine.LoseLife(gs, i, 1, "Cruel Celebrant")
	}
	gameengine.GainLife(gs, seat, 1, "Cruel Celebrant")
}

func registerVindictiveVampire(r *Registry) {
	r.OnTrigger("Vindictive Vampire", "creature_dies", vindictiveTrigger)
}

func vindictiveTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Only another creature YOU control dying.
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat != seat {
		return
	}
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == seat {
			continue
		}
		gameengine.LoseLife(gs, i, 1, "Vindictive Vampire")
	}
	gameengine.GainLife(gs, seat, 1, "Vindictive Vampire")
}

func registerSyrKonrad(r *Registry) {
	r.OnTrigger("Syr Konrad, the Grim", "creature_dies", syrKonradTrigger)
}

func syrKonradTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Whenever ANY creature dies → each opponent loses 1.
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == seat {
			continue
		}
		gameengine.LoseLife(gs, i, 1, "Syr Konrad, the Grim")
	}
}
