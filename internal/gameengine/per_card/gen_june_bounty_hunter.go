package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJuneBountyHunter wires June, Bounty Hunter.
//
// Oracle text:
//
//   June can't be blocked as long as you've drawn two or more cards this turn.
//   {1}, Sacrifice another creature: Create a Clue token. Activate only during your turn. (It's an artifact with "{2}, Sacrifice this token: Draw a card.")
//
// Implementation:
//   - Activated: create a Clue token. The {1} + sacrifice cost and the
//     "your turn" timing restriction are enforced by the activation
//     pipeline.
//   - Static unblockability is engine-side state (cards-drawn-this-turn
//     gating).
func registerJuneBountyHunter(r *Registry) {
	r.OnActivated("June, Bounty Hunter", juneBountyHunterActivate)
}

func juneBountyHunterActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "june_bounty_hunter_clue"
	if gs == nil || src == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	clue := &gameengine.Card{
		Name:     "Clue Token",
		Owner:    src.Controller,
		Types:    []string{"token", "artifact", "clue"},
		Colors:   []string{},
		TypeLine: "Token Artifact — Clue",
	}
	enterBattlefieldWithETB(gs, src.Controller, clue, false)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat": seat,
	})
}
