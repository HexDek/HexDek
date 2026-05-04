package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSidisiBroodTyrant wires Sidisi, Brood Tyrant.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever Sidisi enters or attacks, mill three cards.
//	Whenever one or more creature cards are put into your graveyard
//	from your library, create a 2/2 black Zombie creature token.
//
// Implementation:
//   - ETB and "creature_attacks" both call sidisiMillThree.
//   - "land_to_graveyard" / mill events fire generic "card_went_to_grave"
//     paths, but the engine emits specific zone_change events. We use
//     a custom pass: after milling, scan for creature cards just put in
//     and mint a Zombie token if at least one was a creature.
//   - The second ability is also expected to fire from any other source
//     of self-mill (e.g. Mesmeric Orb). We listen on "zone_change" with
//     to=graveyard, from=library, and the changed card is a creature.
func registerSidisiBroodTyrant(r *Registry) {
	r.OnETB("Sidisi, Brood Tyrant", sidisiETB)
	r.OnTrigger("Sidisi, Brood Tyrant", "creature_attacks", sidisiAttacks)
	r.OnTrigger("Sidisi, Brood Tyrant", "zone_change", sidisiZoneChange)
}

func sidisiETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	sidisiMillThree(gs, perm, "etb")
}

func sidisiAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	sidisiMillThree(gs, perm, "attack")
}

func sidisiMillThree(gs *gameengine.GameState, perm *gameengine.Permanent, source string) {
	const slug = "sidisi_brood_tyrant_mill_three"
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	creatures := 0
	for i := 0; i < 3 && len(seat.Library) > 0; i++ {
		top := seat.Library[0]
		if top != nil && cardHasType(top, "creature") {
			creatures++
		}
		gameengine.MoveCard(gs, top, perm.Controller, "library", "graveyard", "sidisi_mill")
	}
	if creatures > 0 {
		gameengine.CreateCreatureToken(gs, perm.Controller, "Zombie",
			[]string{"creature", "zombie", "pip:B"}, 2, 2)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"trigger":   source,
		"creatures": creatures,
	})
}

func sidisiZoneChange(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "sidisi_creature_milled_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	from, _ := ctx["from"].(string)
	to, _ := ctx["to"].(string)
	if from != "library" || to != "graveyard" {
		return
	}
	ownerSeat, _ := ctx["owner_seat"].(int)
	if ownerSeat != perm.Controller {
		return
	}
	// Avoid double-firing for Sidisi's own ETB/attack mill — those mints
	// happen inline. Use a per-event flag.
	reason, _ := ctx["reason"].(string)
	if reason == "sidisi_mill" {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil || !cardHasType(card, "creature") {
		return
	}
	gameengine.CreateCreatureToken(gs, perm.Controller, "Zombie",
		[]string{"creature", "zombie", "pip:B"}, 2, 2)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"milled": card.DisplayName(),
	})
}
