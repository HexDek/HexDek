package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSmeagolHelpfulGuide wires Sméagol, Helpful Guide.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{B}{G}
//	Legendary Creature — Halfling Horror
//	At the beginning of your end step, if a creature died under your
//	control this turn, the Ring tempts you.
//	Whenever the Ring tempts you, target opponent reveals cards from the
//	top of their library until they reveal a land card. Put that card
//	onto the battlefield tapped under your control and the rest into
//	their graveyard.
//
// Implementation:
//   - "creature_dies" listener: when a creature dies whose Controller (at
//     death time, ctx["controller_seat"]) is Sméagol's controller, stamp
//     a per-turn flag on the seat.
//   - "end_step" listener gated on active_seat == controller AND stamp set
//     for current turn: call TheRingTemptsYou. Tempt is normally followed
//     by the ring-tempt trigger below firing as a downstream effect.
//   - "ring_tempt" listener (canonicalized event): only fire for Sméagol's
//     controller's tempt. Pick a target opponent (highest-life living
//     opponent — they have the deepest deck, biggest exile of cards).
//     Walk the opponent's library from the top, milling each card to their
//     graveyard until we find a land. The found land goes onto Sméagol's
//     controller's battlefield tapped under our control.
func registerSmeagolHelpfulGuide(r *Registry) {
	r.OnTrigger("Sméagol, Helpful Guide", "creature_dies", smeagolHelpfulCreatureDies)
	r.OnTrigger("Sméagol, Helpful Guide", "end_step", smeagolHelpfulEndStep)
	r.OnTrigger("Sméagol, Helpful Guide", "ring_tempt", smeagolHelpfulRingTempt)
}

func smeagolHelpfulCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	controller, _ := ctx["controller_seat"].(int)
	if controller != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[smeagolHelpfulDiedKey(gs.Turn)] = 1
}

func smeagolHelpfulEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "smeagol_helpful_end_step_ring_tempt"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	key := smeagolHelpfulDiedKey(gs.Turn)
	if seat.Flags == nil || seat.Flags[key] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_creature_died_under_your_control_this_turn",
		})
		return
	}
	delete(seat.Flags, key)
	smeagolHelpfulPruneKeys(seat, gs.Turn)
	gameengine.TheRingTemptsYou(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"triggered": true,
		"ring_level": gameengine.GetRingLevel(gs, perm.Controller),
	})
}

func smeagolHelpfulRingTempt(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "smeagol_helpful_ring_tempt_land_grab"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	temptedSeat, _ := ctx["seat"].(int)
	if temptedSeat != perm.Controller {
		return
	}
	// Pick the highest-life living opponent.
	target := -1
	bestLife := -1
	for _, opp := range gs.Opponents(perm.Controller) {
		s := gs.Seats[opp]
		if s == nil || s.Lost {
			continue
		}
		if s.Life > bestLife {
			bestLife = s.Life
			target = opp
		}
	}
	if target < 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_opponent", nil)
		return
	}
	oppSeat := gs.Seats[target]
	if oppSeat == nil || len(oppSeat.Library) == 0 {
		emitFail(gs, slug, perm.Card.DisplayName(), "empty_library", map[string]interface{}{
			"target_seat": target,
		})
		return
	}
	revealed := []string{}
	var foundLand *gameengine.Card
	for len(oppSeat.Library) > 0 {
		top := oppSeat.Library[0]
		if top == nil {
			oppSeat.Library = oppSeat.Library[1:]
			continue
		}
		revealed = append(revealed, top.DisplayName())
		if cardHasType(top, "land") {
			foundLand = top
			break
		}
		gameengine.MoveCard(gs, top, target, "library", "graveyard", slug)
	}
	landName := ""
	if foundLand != nil {
		// Move into the controller's battlefield tapped, under our
		// control. The card's Owner stays as target opponent (the land
		// card itself belongs to them per CR §108.3 — owner does not
		// change), but Controller becomes Sméagol's controller.
		landName = foundLand.DisplayName()
		// Strip from opponent's library first (MoveCard would default to
		// owner's battlefield via the card.Owner path; the manual sweep +
		// createPermanent under our control is the canonical pattern here).
		for i, c := range oppSeat.Library {
			if c == foundLand {
				oppSeat.Library = append(oppSeat.Library[:i], oppSeat.Library[i+1:]...)
				break
			}
		}
		enterBattlefieldWithETB(gs, perm.Controller, foundLand, true)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"target_seat":   target,
		"revealed":      revealed,
		"land_to_play":  landName,
	})
}

func smeagolHelpfulDiedKey(turn int) string {
	return fmt.Sprintf("smeagol_died_t%d", turn+1)
}

func smeagolHelpfulPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "smeagol_died_t"
	cutoff := currentTurn + 1
	for k := range seat.Flags {
		if len(k) <= len(prefix) || k[:len(prefix)] != prefix {
			continue
		}
		n := 0
		_, err := fmt.Sscanf(k[len(prefix):], "%d", &n)
		if err != nil {
			continue
		}
		if n < cutoff {
			delete(seat.Flags, k)
		}
	}
}
