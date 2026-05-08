package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerElasIlKor wires Elas il-Kor, Sadistic Pilgrim.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Deathtouch
//	Whenever another creature you control enters, you gain 1 life.
//	Whenever another creature you control dies, each opponent loses 1
//	  life.
//
// Implementation:
//   - "permanent_etb": creature you control entering; not Elas itself.
//     GainLife(1).
//   - "creature_dies": creature you controlled dies; not Elas himself.
//     Each living opponent loses 1 life.
//   - Deathtouch handled by AST keyword pipeline.
func registerElasIlKor(r *Registry) {
	r.OnTrigger("Elas il-Kor, Sadistic Pilgrim", "permanent_etb", elasIlKorOtherETB)
	r.OnTrigger("Elas il-Kor, Sadistic Pilgrim", "creature_dies", elasIlKorOtherDies)
}

func elasIlKorOtherETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "elas_il_kor_etb_lifegain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	enter, _ := ctx["perm"].(*gameengine.Permanent)
	if enter == nil || enter == perm {
		return
	}
	if enter.Controller != perm.Controller || !enter.IsCreature() {
		return
	}
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": enter.Card.DisplayName(),
	})
}

func elasIlKorOtherDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "elas_il_kor_dies_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dyingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	dyingCard, _ := ctx["card"].(*gameengine.Card)
	if dyingCard == nil {
		return
	}
	if dyingPerm == perm {
		return
	}
	// "another creature you control" — controlled by Elas's controller.
	if dyingPerm != nil {
		if dyingPerm.Controller != perm.Controller {
			return
		}
	} else if dyingCard.Owner != perm.Controller {
		return
	}
	if !cardHasType(dyingCard, "creature") {
		return
	}
	for _, opp := range gs.Opponents(perm.Controller) {
		gameengine.LoseLife(gs, opp, 1, perm.Card.DisplayName())
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"creature": dyingCard.DisplayName(),
	})
	_ = gs.CheckEnd()
}
