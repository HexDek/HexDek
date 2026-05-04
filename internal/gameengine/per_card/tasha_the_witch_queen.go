package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTashaTheWitchQueen wires Tasha, the Witch Queen.
//
// Oracle text:
//
//	{3}{U}{B}
//	Legendary Planeswalker — Tasha
//	Whenever you cast a spell you don't own, create a 3/3 black Demon
//	  creature token.
//	+1: Draw a card. For each opponent, exile up to one target instant or
//	  sorcery card from that player's graveyard and put a page counter on it.
//	-3: You may cast a spell from among cards in exile with page counters
//	  on them without paying its mana cost.
//	Tasha, the Witch Queen can be your commander.
//
// Implementation:
//   - "spell_cast" trigger: when Tasha's controller casts a spell whose
//     Card.Owner != controller, create a 3/3 black Demon token. The
//     gameengine sets card.Owner at deck-load; cards taken via Tasha,
//     Bolas's Citadel, Gonti, etc. retain their original owner.
//   - +1 / -3 loyalty abilities: emitPartial — engine planeswalker
//     activation pipeline does not yet route through per_card.
func registerTashaTheWitchQueen(r *Registry) {
	r.OnTrigger("Tasha, the Witch Queen", "spell_cast", tashaOnSpellCast)
	r.OnETB("Tasha, the Witch Queen", tashaETB)
}

func tashaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	emitPartial(gs, "tasha_witch_queen_etb", perm.Card.DisplayName(),
		"loyalty_abilities_plus1_exile_minus3_cast_from_exile_partial")
}

func tashaOnSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tasha_witch_queen_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if card.Owner == perm.Controller {
		return
	}
	token := &gameengine.Card{
		Name:          "Demon Token",
		Owner:         perm.Controller,
		BasePower:     3,
		BaseToughness: 3,
		Types:         []string{"token", "creature", "demon"},
		Colors:        []string{"B"},
		TypeLine:      "Token Creature — Demon",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"spell": card.DisplayName(),
	})
}
