package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGoroGoroAndSatoru wires Goro-Goro and Satoru.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever one or more creatures you control that entered this turn
//	  deal combat damage to a player, create a 5/5 red Dragon Spirit
//	  creature token with flying.
//	{1}{R}: Creatures you control gain haste until end of turn.
//
// Implementation:
//   - "combat_damage_player": gate on damage_seat == perm.Controller and
//     check that the source perm entered this turn (perm.Flags["entered_turn"] == gs.Turn,
//     or fall back to ETB summoning-sick proxy). Once-per-combat via
//     turn-keyed flag on Goro-Goro himself.
//   - Activated haste-grant uses the layers pipeline — emitPartial.
func registerGoroGoroAndSatoru(r *Registry) {
	r.OnTrigger("Goro-Goro and Satoru", "combat_damage_player", goroGoroCombatDamage)
	r.OnActivated("Goro-Goro and Satoru", goroGoroHasteGrant)
}

func goroGoroCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "goro_goro_dragon_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dmgSeat, _ := ctx["damage_seat"].(int)
	if dmgSeat != perm.Controller {
		return
	}
	src, _ := ctx["source_perm"].(*gameengine.Permanent)
	if src == nil || src.Controller != perm.Controller {
		return
	}
	// Did this attacker enter this turn? We treat summoning sickness as a
	// proxy for "entered this turn" since the engine doesn't expose a
	// separate entered_turn field on every perm.
	if !src.SummoningSick {
		// Fall back to checking a Flags entry.
		if src.Flags == nil || src.Flags["entered_turn"] != gs.Turn {
			return
		}
	}
	// Once per combat phase per turn.
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["goro_token_turn"] == gs.Turn {
		return
	}
	perm.Flags["goro_token_turn"] = gs.Turn

	token := &gameengine.Card{
		Name:          "Dragon Spirit Token",
		Owner:         perm.Controller,
		BasePower:     5,
		BaseToughness: 5,
		Types:         []string{"token", "creature", "dragon", "spirit", "kw:flying"},
		Colors:        []string{"R"},
		TypeLine:      "Token Creature — Dragon Spirit",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": src.Card.DisplayName(),
	})
}

func goroGoroHasteGrant(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	emitPartial(gs, "goro_goro_haste_grant", src.Card.DisplayName(),
		"ueot_haste_grant_not_modeled_by_layers_pipeline")
}
