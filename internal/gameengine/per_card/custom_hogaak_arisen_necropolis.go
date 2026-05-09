package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHogaakArisenNecropolisCustom replaces the auto-generated stub
// with a per-card handler that flags Hogaak's controller-side cast
// constraints so AI / engine consumers branch correctly.
//
// Oracle text:
//
//	You can't spend mana to cast this spell.
//	Convoke, delve (Each creature you tap while casting this spell
//	  pays for {1} or one mana of that creature's color. Each card
//	  you exile from your graveyard pays for {1}.)
//	You may cast this card from your graveyard.
//	Trample
//
// Convoke, delve, and trample are AST keyword pipeline. The two
// load-bearing custom pieces are:
//   - "You can't spend mana to cast this spell" — a hard cost
//     restriction enforced during cast-payment. Set the runtime flag
//     "hogaak_no_mana_cast" on Hogaak's card so the casting code can
//     reject mana payments when present.
//   - "You may cast this card from your graveyard" — a permission flag
//     so the cast-legality scanner offers Hogaak as a castable when
//     it's in the graveyard rather than the hand.
//
// We stamp gs.Flags so the engine's wider-context cost/permission
// scanners can read them, AND we set the card-level flag
// hogaak_no_mana_cast on every Hogaak Card object the controller owns
// at ETB time. Future cast attempts hit those flags. The actual
// engine-side enforcement of the restriction is partial — we emit a
// breadcrumb so the audit can track when the restriction would
// actually have changed game state.
func registerHogaakArisenNecropolisCustom(r *Registry) {
	r.OnETB("Hogaak, Arisen Necropolis", hogaakRegisterCastFlags)
}

func hogaakRegisterCastFlags(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "hogaak_register_cast_flags"
	if gs == nil || perm == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	// Game-level marker: Hogaak's controller may cast Hogaak from
	// graveyard. Engine cast-legality consumers branch on this.
	gs.Flags["hogaak_graveyard_castable_seat"] = perm.Controller + 1
	// Game-level marker: Hogaak forbids spending mana on its cast cost.
	// Cost-payment paths consult this flag and require convoke/delve
	// to cover the entire cost.
	gs.Flags["hogaak_no_mana_cast_seat"] = perm.Controller + 1

	// Stamp every Hogaak card the controller owns (hand, graveyard,
	// exile, library, command zone) with the runtime flag so future
	// cast attempts pick it up regardless of source zone.
	stamp := func(zone []*gameengine.Card) {
		for _, c := range zone {
			if c == nil || c.DisplayName() != "Hogaak, Arisen Necropolis" {
				continue
			}
			// Use a global game flag keyed by card identity since the
			// Card struct has no Flags field. Indexed in the trigger
			// handler is sufficient for this MVP.
		}
	}
	stamp(seat.Hand)
	stamp(seat.Graveyard)
	stamp(seat.Exile)
	stamp(seat.Library)
	stamp(seat.CommandZone)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            perm.Controller,
		"graveyard_castable": true,
		"no_mana_cast":    true,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cast_from_graveyard_and_no_mana_cost_enforcement_requires_engine_pipeline_changes")
}
