package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJawsRelentlessPredator wires Jaws, Relentless Predator.
//
// Oracle text:
//
//	Trample, haste
//	Whenever Jaws deals combat damage to a player, create that many
//	Blood tokens.
//	Whenever a noncreature artifact is sacrificed or destroyed, Jaws
//	deals 1 damage to each opponent.
//
// Implementation:
//   - combat_damage_player by Jaws → create N Blood tokens.
//   - permanent_ltb (to graveyard) for noncreature artifact → ping each
//     opponent for 1. Covers both sacrifice and destruction; excludes
//     bounce/exile because we gate on to_zone="graveyard".
func registerJawsRelentlessPredator(r *Registry) {
	r.OnTrigger("Jaws, Relentless Predator", "combat_damage_player", jawsCombatDamage)
	r.OnTrigger("Jaws, Relentless Predator", "permanent_ltb", jawsArtifactGone)
}

func jawsCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "jaws_blood_tokens"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	if sourceSeat != perm.Controller {
		return
	}
	sourceName, _ := ctx["source_card"].(string)
	if sourceName != "" && sourceName != perm.Card.DisplayName() {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	for i := 0; i < amount; i++ {
		token := &gameengine.Card{
			Name:     "Blood Token",
			Owner:    perm.Controller,
			Types:    []string{"token", "artifact", "blood"},
			TypeLine: "Token Artifact — Blood",
		}
		enterBattlefieldWithETB(gs, perm.Controller, token, false)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"tokens": amount,
	})
}

func jawsArtifactGone(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "jaws_artifact_ping"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	gone, _ := ctx["perm"].(*gameengine.Permanent)
	if gone == nil || gone.Card == nil {
		return
	}
	toZone, _ := ctx["to_zone"].(string)
	if toZone != "graveyard" {
		return
	}
	if cardHasType(gone.Card, "creature") {
		return
	}
	if !cardHasType(gone.Card, "artifact") {
		return
	}
	pinged := 0
	for i := range gs.Seats {
		if i == perm.Controller {
			continue
		}
		s := gs.Seats[i]
		if s == nil || s.Lost {
			continue
		}
		gameengine.DealDamage(gs, i, 1, perm.Card.DisplayName())
		pinged++
	}
	_ = gs.CheckEnd()
	if pinged > 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"opponents": pinged,
		})
	}
}
