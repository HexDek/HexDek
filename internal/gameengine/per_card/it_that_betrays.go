package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerItThatBetrays wires It That Betrays (Muninn parser-gap rank
// ~150, Eldrazi reanimator finisher).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{12}
//	Creature — Eldrazi
//	Annihilator 2 (Whenever this creature attacks, defending player
//	sacrifices two permanents of their choice.)
//	Whenever an opponent sacrifices a nontoken permanent, put that card
//	onto the battlefield under your control.
//
// Implementation:
//   - Annihilator 2 is handled by the engine combat keyword pipeline
//     (CR §702.86, GetAnnihilatorN). No per-card wiring required.
//   - OnTrigger("permanent_sacrificed"): gate on the sacrificing seat
//     being an opponent of perm.Controller AND the sacrificed card NOT
//     being a token (per oracle "nontoken permanent"). Pull the card
//     out of the graveyard it just landed in and route it through the
//     full ETB cascade under perm.Controller's seat — including
//     replacements / triggers so the cheated entry is observed by the
//     rest of the table.
func registerItThatBetrays(r *Registry) {
	r.OnTrigger("It That Betrays", "permanent_sacrificed", itThatBetraysOppSac)
}

func itThatBetraysOppSac(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "it_that_betrays_steal_sacrificed"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	controllerSeat, _ := ctx["controller_seat"].(int)
	if controllerSeat == perm.Controller {
		return
	}
	if controllerSeat < 0 || controllerSeat >= len(gs.Seats) {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	// Skip tokens — they cease to exist on sac and oracle excludes them.
	if cardHasType(card, "token") {
		return
	}
	// The card just moved from battlefield to graveyard via the sacrifice
	// path. Pull it back out of its owner's graveyard before reanimating
	// under our control so MoveCard's source-removal path sees it cleanly.
	owner := card.Owner
	if owner < 0 || owner >= len(gs.Seats) {
		owner = controllerSeat
	}
	gameengine.RemoveCardFromAllPrivateZones(gs, owner, card)
	enterBattlefieldWithETB(gs, perm.Controller, card, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"from_seat":     controllerSeat,
		"stolen":        card.DisplayName(),
	})
}
