package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKomaCosmosSerpent wires Koma, Cosmos Serpent.
//
// Oracle text:
//
//	This spell can't be countered.
//	At the beginning of each upkeep, create a 3/3 blue Serpent creature
//	token named Koma's Coil.
//	Sacrifice another Serpent: Choose one —
//	  • Tap target permanent. Its activated abilities can't be activated
//	    this turn.
//	  • Koma gains indestructible until end of turn.
//
// Implementation: upkeep token creation (every player's upkeep). The
// activated sacrifice ability is non-trivial dispatch — emitPartial.
func registerKomaCosmosSerpent(r *Registry) {
	r.OnTrigger("Koma, Cosmos Serpent", "upkeep_controller", komaUpkeepToken)
	r.OnTrigger("Koma, Cosmos Serpent", "upkeep_start", komaUpkeepToken)
	r.OnActivated("Koma, Cosmos Serpent", komaActivated)
}

func komaUpkeepToken(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "koma_upkeep_serpent_token"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	token := &gameengine.Card{
		Name:          "Koma's Coil",
		Owner:         seat,
		BasePower:     3,
		BaseToughness: 3,
		Types:         []string{"token", "creature", "serpent"},
		Colors:        []string{"U"},
		TypeLine:      "Token Creature — Serpent",
	}
	enterBattlefieldWithETB(gs, seat, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": seat,
	})
}

func komaActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "koma_sacrifice_serpent_modal"
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, slug, src.Card.DisplayName(),
		"sac_serpent_modal_tap_or_indestructible_unimplemented")
}
