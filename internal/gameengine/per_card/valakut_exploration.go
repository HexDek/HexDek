package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerValakutExploration wires Valakut Exploration.
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//	Landfall — Whenever a land you control enters, exile the top card
//	of your library. You may play that card for as long as it remains
//	exiled.
//	At the beginning of your end step, if there are cards exiled with
//	this enchantment, put them into their owner's graveyard, then this
//	enchantment deals that much damage to each opponent.
//
// Implementation:
//   - permanent_etb gated on entering land controlled by us: exile the
//     top of our library; tag with "valakut_exiled" so the end-step
//     sweeper can find it. The "may play that card" half is a
//     play-from-exile hook gap (same as Ixhel / Bre / Urabrask) —
//     flagged once on first ETB via emitPartial.
//   - end_step (controller's): scan controller's exile for tagged
//     cards, move them to graveyard, then deal that many damage to
//     each opponent via LoseLife (engine treats non-combat damage to
//     players as life-loss for resolution purposes).
const valakutExplorationTag = "valakut_exiled"

func registerValakutExploration(r *Registry) {
	r.OnETB("Valakut Exploration", valakutExplorationETB)
	r.OnTrigger("Valakut Exploration", "permanent_etb", valakutExplorationLandfall)
	r.OnTrigger("Valakut Exploration", "end_step", valakutExplorationEndStep)
}

func valakutExplorationETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "valakut_exploration_play_from_exile", perm.Card.DisplayName(),
		"\"may play that card for as long as it remains exiled\" requires play-from-exile hook; tag set on each exiled card")
}

func valakutExplorationLandfall(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "valakut_exploration_landfall_exile"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil || entering == perm {
		return
	}
	if !entering.IsLand() {
		return
	}
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || len(seat.Library) == 0 {
		return
	}
	top := seat.Library[0]
	gameengine.MoveCard(gs, top, perm.Controller, "library", "exile", "valakut_exploration_exile")
	if top != nil {
		top.Types = append(top.Types, valakutExplorationTag)
	}
	exiledName := ""
	if top != nil {
		exiledName = top.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"land":   entering.Card.DisplayName(),
		"exiled": exiledName,
	})
}

func valakutExplorationEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "valakut_exploration_end_step_burn"
	if gs == nil || perm == nil || ctx == nil {
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

	tagged := []*gameengine.Card{}
	for _, c := range seat.Exile {
		if c == nil {
			continue
		}
		if cardHasType(c, valakutExplorationTag) {
			tagged = append(tagged, c)
		}
	}
	if len(tagged) == 0 {
		return
	}
	for _, c := range tagged {
		gameengine.MoveCard(gs, c, perm.Controller, "exile", "graveyard", "valakut_exploration_eos")
		valakutStripTag(c)
	}
	dmg := len(tagged)
	hit := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		gameengine.LoseLife(gs, i, dmg, perm.Card.DisplayName())
		hit++
	}
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"yarded":   dmg,
		"opps_hit": hit,
		"damage":   dmg,
	})
}

func valakutStripTag(c *gameengine.Card) {
	if c == nil {
		return
	}
	out := c.Types[:0]
	for _, t := range c.Types {
		if t == valakutExplorationTag {
			continue
		}
		out = append(out, t)
	}
	c.Types = out
}
