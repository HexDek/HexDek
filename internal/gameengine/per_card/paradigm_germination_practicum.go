package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGerminationPracticum wires up Germination Practicum.
//
// Oracle text:
//
//	Put two +1/+1 counters on each creature you control. Paradigm
//
// {3}{G} Sorcery — Secrets of Strixhaven.
func registerGerminationPracticum(r *Registry) {
	r.OnResolve("Germination Practicum", germinationPracticumResolve)
}

func germinationPracticumResolve(gs *gameengine.GameState, item *gameengine.StackItem) {
	const slug = "germination_practicum"
	const cardName = "Germination Practicum"
	if gs == nil || item == nil {
		return
	}
	seat := item.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}

	// --- Effect: +2/+2 counters on each creature you control. ---
	buffed := 0
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "creature") {
			continue
		}
		p.AddCounter("+1/+1", 2)
		buffed++
		gs.LogEvent(gameengine.Event{
			Kind:   "counter_added",
			Seat:   seat,
			Source: cardName,
			Details: map[string]interface{}{
				"target":  p.Card.DisplayName(),
				"counter": "+1/+1",
				"amount":  2,
			},
		})
	}

	// --- Paradigm: exile instead of graveyard, register for auto-copy. ---
	paradigmExileItem(gs, item, seat, slug, cardName)

	emit(gs, slug, cardName, map[string]interface{}{
		"seat":           seat,
		"creatures_buffed": buffed,
	})
}
