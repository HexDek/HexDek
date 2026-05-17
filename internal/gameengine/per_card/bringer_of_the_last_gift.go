package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBringerOfTheLastGift wires Bringer of the Last Gift (Muninn
// parser-gap #68, ~9.6K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{6}{B}{B}
//	Creature — Vampire Demon
//	Flying
//	When this creature enters, if you cast it, each player sacrifices
//	all other creatures they control. Then each player returns all
//	creature cards from their graveyard that weren't put there this way
//	to the battlefield.
//
// Implementation:
//   - Flying is AST-side.
//   - Gate on was_cast.
//   - Snapshot every graveyard's creature-card pointer set BEFORE the
//     sacrifice wave; that snapshot is the "weren't put there this way"
//     set. Each opponent's set is independent (a player only returns
//     their OWN graveyard's pre-existing creatures per CR §612 grammar).
//   - Walk all seats, sacrifice every non-Bringer creature they control,
//     via SacrificePermanent so dies-triggers and SBAs run.
//   - Then for each seat, return only those creature cards present in
//     their pre-snapshot set, via MoveCard("graveyard"→"battlefield")
//     followed by enterBattlefieldWithETB so per-card ETB handlers fire.
//   - Hat policy: not a choice — printed "all creature cards", so we
//     return every eligible pointer regardless of board impact.
func registerBringerOfTheLastGift(r *Registry) {
	r.OnETB("Bringer of the Last Gift", bringerOfTheLastGiftETB)
}

func bringerOfTheLastGiftETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "bringer_of_the_last_gift_mass_sac_return"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	// Snapshot pre-existing creature cards in each graveyard.
	preSnapshot := make([][]*gameengine.Card, len(gs.Seats))
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, c := range s.Graveyard {
			if c == nil || !cardHasType(c, "creature") {
				continue
			}
			preSnapshot[i] = append(preSnapshot[i], c)
		}
	}
	// Sacrifice all other creatures on the battlefield.
	sacced := 0
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		victims := []*gameengine.Permanent{}
		for _, p := range s.Battlefield {
			if p == nil || p == perm || !p.IsCreature() {
				continue
			}
			victims = append(victims, p)
		}
		for _, v := range victims {
			gameengine.SacrificePermanent(gs, v, "bringer_of_the_last_gift_mass_sac")
			sacced++
		}
	}
	// Return each seat's pre-snapshot creatures still in their graveyard.
	returned := 0
	for i, s := range gs.Seats {
		if s == nil {
			continue
		}
		wanted := map[*gameengine.Card]bool{}
		for _, c := range preSnapshot[i] {
			wanted[c] = true
		}
		// Re-scan current graveyard for surviving pre-snapshot pointers
		// (mass-sac may have caused replacement-effect redirects).
		picks := []*gameengine.Card{}
		for _, c := range s.Graveyard {
			if wanted[c] {
				picks = append(picks, c)
			}
		}
		for _, c := range picks {
			gameengine.MoveCard(gs, c, i, "graveyard", "battlefield", slug)
			enterBattlefieldWithETB(gs, i, c, false)
			returned++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":       perm.Controller,
		"sacrificed": sacced,
		"returned":   returned,
	})
}
