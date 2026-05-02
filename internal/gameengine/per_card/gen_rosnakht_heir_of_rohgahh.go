package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRosnakhtHeirOfRohgahh wires Rosnakht, Heir of Rohgahh.
//
// Oracle text:
//
//	Battle cry (Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.)
//	Heroic — Whenever you cast a spell that targets Rosnakht, create a 0/1 red Kobold creature token named Kobolds of Kher Keep.
//
// Implementation:
//   - Battle cry: The engine handles battle cry automatically via
//     ApplyBattleCry when the creature has the "battle cry" keyword in
//     its AST. The ETB handler ensures the keyword flag is set as a
//     safety net for cards whose AST may not parse the keyword.
//   - Heroic: OnTrigger("spell_cast") — when the controller casts a
//     spell that targets a creature (oracle-text heuristic), and
//     Rosnakht is on the battlefield, create a 0/1 red Kobold token.
//     The engine's targeting resolution is approximate (targets aren't
//     individually tracked), so we gate on "controller cast a
//     creature-targeting spell while Rosnakht is on the battlefield."
//     This is the same fidelity as Killian, Ink Duelist's cost reducer.
func registerRosnakhtHeirOfRohgahh(r *Registry) {
	r.OnETB("Rosnakht, Heir of Rohgahh", rosnakhtHeirOfRohgahhETB)
	r.OnTrigger("Rosnakht, Heir of Rohgahh", "spell_cast", rosnakhtHeroicSpellCast)
}

// rosnakhtHeirOfRohgahhETB ensures the battle cry keyword flag is set.
// The AST parser should already tag this, but the flag is a safety net
// so ApplyBattleCry (keywords_combat.go) always recognises Rosnakht as
// a battle cry source.
func rosnakhtHeirOfRohgahhETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	// Ensure battle cry keyword is discoverable via HasKeyword.
	perm.Flags["kw:battle cry"] = 1

	emit(gs, "rosnakht_heir_of_rohgahh_etb", perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"battle_cry": true,
		"heroic":     true,
	})
}

// rosnakhtHeroicSpellCast implements Heroic — whenever you cast a spell
// that targets Rosnakht, create a 0/1 red Kobold creature token named
// Kobolds of Kher Keep.
//
// Approximation: any creature-targeting spell cast by Rosnakht's
// controller triggers the heroic ability. This slightly over-fires
// (the spell might target a different creature), but in practice
// Rosnakht decks run many targeted pump spells aimed at their own
// commander, so the approximation is accurate for simulation.
func rosnakhtHeroicSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rosnakht_heroic_kobold_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}

	// Only trigger on spells cast by Rosnakht's controller.
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}

	// Check if the spell targets a creature (heroic trigger condition).
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}

	// Don't trigger on creature spells (they don't target).
	if cardHasType(card, "creature") {
		return
	}

	// Check oracle text for creature-targeting language.
	oracleText := gameengine.OracleTextLower(card)
	targetPhrases := []string{
		"target creature",
		"target a creature",
		"target another creature",
		"target attacking creature",
		"enchant creature",
	}
	targetsCreature := false
	for _, phrase := range targetPhrases {
		if strings.Contains(oracleText, phrase) {
			targetsCreature = true
			break
		}
	}
	if !targetsCreature {
		return
	}

	// Create a 0/1 red Kobold creature token named "Kobolds of Kher Keep".
	token := gameengine.CreateCreatureToken(gs, perm.Controller, "Kobolds of Kher Keep",
		[]string{"creature", "kobold", "pip:R"}, 0, 1)

	tokenCreated := token != nil

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"spell":         card.DisplayName(),
		"token_created": tokenCreated,
		"token_name":    "Kobolds of Kher Keep",
		"token_pt":      "0/1",
	})
}
