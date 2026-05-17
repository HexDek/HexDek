package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBygoneBishop wires Bygone Bishop (Muninn parser-gap, ~3.7K hits
// across investigate-pile decks).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{W}
//	Creature — Spirit Cleric
//	Flying
//	Whenever you cast a creature spell with mana value 3 or less,
//	investigate. (Create a Clue token. It's an artifact with "{2},
//	Sacrifice this token: Draw a card.")
//
// Implementation:
//   - Flying handled by AST keyword pipeline.
//   - spell_cast gated on caster_seat == controller, card is a creature,
//     and ManaCostOf(card) <= 3. Engine has first-class CreateClueToken
//     support (internal/gameengine/tokens.go) — no need to roll our own
//     token shell, so we reuse it to stay consistent with The Rani's
//     investigate handler.
func registerBygoneBishop(r *Registry) {
	r.OnTrigger("Bygone Bishop", "spell_cast", bygoneBishopSpellCast)
}

func bygoneBishopSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "bygone_bishop_investigate"
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
	if !cardHasType(card, "creature") {
		return
	}
	if gameengine.ManaCostOf(card) > 3 {
		return
	}
	gameengine.CreateClueToken(gs, perm.Controller)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"spell": card.DisplayName(),
		"cmc":   gameengine.ManaCostOf(card),
	})
}
