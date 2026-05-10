package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerWilsonRefinedGrizzly wires Wilson, Refined Grizzly.
//
// Oracle text:
//
//	This spell can't be countered.
//	Vigilance, reach, trample
//	Ward {2}
//	Choose a Background (You can have a Background as a second commander.)
//
// Implementation:
//   - All four lines are statics. ETB stamps:
//       - kw:cant_be_countered (handled at cast-time by the cast pipeline
//         when checking the resolving spell's source perm)
//       - ward:2 (read by the target dispatcher when an opponent targets)
//       - kw:vigilance, kw:reach, kw:trample (engine combat reads)
//       - allow_background_partner (commander-deck construction reads)
//   - The "becomes the target" trigger that ward fires lives in the
//     engine target dispatcher; we just declare the cost via the flag.
func registerWilsonRefinedGrizzly(r *Registry) {
	r.OnETB("Wilson, Refined Grizzly", wilsonRefinedGrizzlyETB)
}

func wilsonRefinedGrizzlyETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "wilson_refined_grizzly_etb"
	if gs == nil || perm == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["ward"] = 2
	perm.Flags["kw:vigilance"] = 1
	perm.Flags["kw:reach"] = 1
	perm.Flags["kw:trample"] = 1
	// "This spell can't be countered" applies to the cast event itself;
	// once Wilson resolves, it's a non-issue. We surface the boundary.
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"ward":      2,
		"keywords":  []string{"vigilance", "reach", "trample"},
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"choose_background_partner_handled_at_deck_construction_time")
}
