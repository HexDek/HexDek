package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDecorumDissertation wires up Decorum Dissertation.
//
// Oracle text:
//
//	Target player draws two cards and loses 2 life. Paradigm
//
// {2}{B} Sorcery — Secrets of Strixhaven.
//
// Paradigm means: after this spell resolves, exile it instead of
// putting it into the graveyard. Then at the beginning of each of
// your first main phases, you may cast a free copy from exile.
func registerDecorumDissertation(r *Registry) {
	r.OnResolve("Decorum Dissertation", decorumDissertationResolve)
}

func decorumDissertationResolve(gs *gameengine.GameState, item *gameengine.StackItem) {
	const slug = "decorum_dissertation"
	const cardName = "Decorum Dissertation"
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

	// --- Effect: target player draws two cards and loses 2 life. ---
	// MVP: target is always the controller.
	drawn := 0
	for i := 0; i < 2; i++ {
		if len(s.Library) > 0 {
			card := s.Library[0]
			gameengine.MoveCard(gs, card, seat, "library", "hand", "draw")
			drawn++
		}
	}
	gs.LogEvent(gameengine.Event{
		Kind:   "draw",
		Seat:   seat,
		Source: cardName,
		Amount: drawn,
	})
	gameengine.LoseLife(gs, seat, 2, cardName)

	// --- Paradigm: exile instead of graveyard, register for auto-copy. ---
	paradigmExileItem(gs, item, seat, slug, cardName)

	emit(gs, slug, cardName, map[string]interface{}{
		"seat":      seat,
		"drawn":     drawn,
		"life_lost": 2,
	})
	_ = gs.CheckEnd()
}
