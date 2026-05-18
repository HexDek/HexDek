package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMahaItsFeathersNight wires Maha, Its Feathers Night.
//
// Oracle text:
//
//   Flying, trample
//   Ward—Discard a card.
//   Creatures your opponents control have base toughness 1.
//
// R37 port:
//
//   - Flying / trample / Ward: AST keyword pipeline.
//   - "Creatures your opponents control have base toughness 1" —
//     layer 7c base-PT replacement. The HexDek layer system doesn't
//     yet expose a "base toughness override" entry point (Layer 7a-d
//     have continuous +N/+N support but base-PT *replacement* like
//     Humility/Mirror Mockery/Maha is a distinct hook). We set the
//     same kind of seat-keyed runtime flag that Torbran uses for its
//     damage-replacement breadcrumb so future layer-system work has a
//     consumer it can wire against.
//
//     Flag scheme: gs.Flags["maha_base_tough_one_active"] = (seat+1).
//     +1 offset so default-zero is distinguishable from "seat 0
//     controls a Maha". Cleared by an LTB hook (not yet wired) once
//     the layer-7c consumer lands; until then a manual clear via
//     ClearMahaBaseToughness is the supported reset for tests +
//     bounce-and-recast scenarios.
//
//     Stacks aren't tracked because "base toughness 1" doesn't stack —
//     §613.3 makes multiple Mahas redundant. Two Maha-controllers would
//     each set their own opponent-bucket; the flag holds the latest
//     setter only as a "is this active anywhere" breadcrumb.
func registerMahaItsFeathersNight(r *Registry) {
	r.OnETB("Maha, Its Feathers Night", mahaETBSetBaseToughnessFlag)
	// permanent_ltb fires for every leave-the-battlefield event;
	// mahaLTB filters to only the Maha-self LTB case before clearing.
	r.OnTrigger("Maha, Its Feathers Night", "permanent_ltb", mahaLTBClearBaseToughnessFlag)
}

func mahaETBSetBaseToughnessFlag(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "maha_base_toughness_one"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	// +1 offset so seat 0 is distinguishable from "unset".
	gs.Flags["maha_base_tough_one_active"] = seat + 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     seat,
		"flag":     "maha_base_tough_one_active",
		"value":    seat + 1,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"layer-7c base-toughness override consumer not yet wired; seat-keyed runtime flag set as breadcrumb")
}

// mahaLTBClearBaseToughnessFlag fires on any permanent LTB; filters to
// Maha-self by checking the trigger ctx's leaving perm pointer.
func mahaLTBClearBaseToughnessFlag(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || gs.Flags == nil {
		return
	}
	leaving, _ := ctx["perm"].(*gameengine.Permanent)
	if leaving != perm {
		return
	}
	delete(gs.Flags, "maha_base_tough_one_active")
}

// ClearMahaBaseToughness is the public reset (mostly for tests and
// for hypothetical board-wipe-then-recast scenarios) that drops the
// Maha base-toughness-1 flag.
func ClearMahaBaseToughness(gs *gameengine.GameState) {
	if gs == nil || gs.Flags == nil {
		return
	}
	delete(gs.Flags, "maha_base_tough_one_active")
}
