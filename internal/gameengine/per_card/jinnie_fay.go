package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJinnieFay wires Jinnie Fay, Jetmir's Second.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	If you would create one or more tokens, you may instead create that
//	  many 2/2 green Cat creature tokens with haste or that many 3/1
//	  green Dog creature tokens with vigilance.
//
// This is a token-creation replacement effect (CR §614). The engine's
// token creation pipeline doesn't yet expose a per-card replacement
// hook, so this handler emits a partial flag and is registered on ETB
// only as a marker.
func registerJinnieFay(r *Registry) {
	r.OnETB("Jinnie Fay, Jetmir's Second", jinnieFayETB)
}

func jinnieFayETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "jinnie_fay_token_replacement", perm.Card.DisplayName(),
		"token_creation_replacement_to_cat_or_dog_not_modeled")
}
