package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEowynShieldmaiden wires Éowyn, Shieldmaiden.
// (Slug owyn_shieldmaiden in the batch — actual card name is Éowyn.)
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	First strike
//	At the beginning of combat on your turn, if another Human entered
//	  the battlefield under your control this turn, create two 2/2 red
//	  Human Knight creature tokens with trample and haste. Then if you
//	  control six or more Humans, draw a card.
//
// Implementation:
//   - "permanent_etb": for our Human creatures entering, mark a turn-keyed
//     flag on Éowyn (eowyn_human_etb_turn = gs.Turn).
//   - "combat_begin": at our combat start, check the flag. Mint two 2/2
//     red Human Knight tokens with trample and haste (kw:* flags). Then
//     count Humans we control; if >= 6, draw a card.
//   - First strike handled by AST keyword pipeline.
func registerEowynShieldmaiden(r *Registry) {
	r.OnTrigger("Éowyn, Shieldmaiden", "permanent_etb", eowynHumanETBMark)
	r.OnTrigger("Éowyn, Shieldmaiden", "combat_begin", eowynCombatBegin)
}

func eowynHumanETBMark(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enter, _ := ctx["perm"].(*gameengine.Permanent)
	if enter == nil || enter == perm || enter.Card == nil {
		return
	}
	if enter.Controller != perm.Controller {
		return
	}
	if !cardHasType(enter.Card, "creature") || !cardHasType(enter.Card, "human") {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["eowyn_human_etb_turn"] = gs.Turn
}

func eowynCombatBegin(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "eowyn_shieldmaiden_human_knight_swarm"
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["eowyn_human_etb_turn"] != gs.Turn {
		return
	}
	if gs.Active != perm.Controller {
		return
	}
	for i := 0; i < 2; i++ {
		token := &gameengine.Card{
			Name:          "Human Knight Token",
			Owner:         perm.Controller,
			BasePower:     2,
			BaseToughness: 2,
			Types:         []string{"token", "creature", "human", "knight", "kw:trample", "kw:haste"},
			Colors:        []string{"R"},
			TypeLine:      "Token Creature — Human Knight",
		}
		enterBattlefieldWithETB(gs, perm.Controller, token, false)
	}
	humans := 0
	seat := gs.Seats[perm.Controller]
	if seat != nil {
		for _, p := range seat.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			if cardHasType(p.Card, "human") {
				humans++
			}
		}
	}
	drewCard := false
	if humans >= 6 {
		if drawOne(gs, perm.Controller, perm.Card.DisplayName()) != nil {
			drewCard = true
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"humans":   humans,
		"tokens":   2,
		"drew":     drewCard,
	})
}
