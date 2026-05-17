package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSandScout wires Sand Scout.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	When this creature enters, if an opponent controls more lands than
//	you, search your library for a Desert card, put it onto the
//	battlefield tapped, then shuffle.
//
//	Whenever one or more land cards are put into your graveyard from
//	anywhere, create a 1/1 red, green, and white Sand Warrior creature
//	token. This ability triggers only once each turn.
//
// Implementation (Muninn gap #40 — 23K hits):
//   - ETB Desert fetch is already covered by registerLandTaxFamily in
//     land_tax_family.go (Sand Scout is listed in landTaxFamilyEntries
//     and that helper is wired into registry.go at line 1731). We
//     deliberately do NOT re-register an ETB here to avoid double-fetch.
//   - Token trigger: OnTrigger("land_to_graveyard"). The engine fires
//     this generic event whenever a land card enters any graveyard from
//     any zone (zone_change.go:499). We filter to lands owned by Sand
//     Scout's controller, and clamp to once per turn using a permanent
//     flag stamped with gs.Turn.
func registerSandScout(r *Registry) {
	r.OnTrigger("Sand Scout", "land_to_graveyard", sandScoutLandToGraveyard)
}

func sandScoutLandToGraveyard(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sand_scout_sand_warrior_token"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	ownerSeat, _ := ctx["owner_seat"].(int)
	if ownerSeat != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Flags["sand_scout_token_turn"] == gs.Turn+1 {
		// Already fired this turn (+1 offset avoids zero-collision on
		// turn 0).
		return
	}
	perm.Flags["sand_scout_token_turn"] = gs.Turn + 1
	token := &gameengine.Card{
		Name:          "Sand Warrior Token",
		Owner:         perm.Controller,
		Types:         []string{"creature", "token", "sand", "warrior", "pip:R", "pip:G", "pip:W"},
		Colors:        []string{"R", "G", "W"},
		BasePower:     1,
		BaseToughness: 1,
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"token": "Sand Warrior",
	})
}

