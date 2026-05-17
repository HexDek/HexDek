package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBreachingLeviathan wires Breaching Leviathan (Muninn parser-gap
// #65, ~10K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{7}{U}{U}
//	Creature — Leviathan
//	When this creature enters, if you cast it from your hand, tap all
//	nonblue creatures. Those creatures don't untap during their
//	controllers' next untap steps.
//
// Implementation:
//   - Gate on was_cast (cast clause). The engine doesn't yet distinguish
//     cast-from-hand from cast-from-elsewhere; the was_cast flag plus
//     emitPartial covers the typical case (Cyclone Summoner does the same).
//   - Iterate every battlefield creature; if it has no blue color/pip,
//     tap it and set DoesNotUntap=true so the next untap step skips it.
//   - We don't filter out controller's own creatures — the printed text
//     hits "all nonblue creatures" regardless of controller.
func registerBreachingLeviathan(r *Registry) {
	r.OnETB("Breaching Leviathan", breachingLeviathanETB)
}

func breachingLeviathanETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "breaching_leviathan_tap_nonblue"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		return
	}
	tapped := 0
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			if cardIsBlue(p.Card) {
				continue
			}
			p.Tapped = true
			p.DoesNotUntap = true
			tapped++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"tapped": tapped,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cast_from_hand_specifically_unmodeled_was_cast_flag_only")
}

// cardIsBlue returns true if the card has the blue color (U) — either an
// explicit Colors entry, or a "pip:U" / "color:blue" / "color:u" type tag.
func cardIsBlue(c *gameengine.Card) bool {
	if c == nil {
		return false
	}
	for _, col := range c.Colors {
		if strings.EqualFold(col, "U") || strings.EqualFold(col, "blue") {
			return true
		}
	}
	for _, t := range c.Types {
		lt := strings.ToLower(t)
		if lt == "pip:u" || lt == "color:blue" || lt == "color:u" {
			return true
		}
	}
	return false
}
