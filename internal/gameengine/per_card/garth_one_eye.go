package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGarthOneEye wires Garth One-Eye.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{T}: Choose a card name that hasn't been chosen from among
//	  Disenchant, Braingeyser, Terror, Shivan Dragon, Regrowth, and
//	  Black Lotus. Create a copy of the card with the chosen name. You
//	  may cast the copy. (You still pay its costs.)
//
// Implementation:
//   - The activated ability creates a temporary copy of one of six cards
//     and lets the controller cast it. Building copies of arbitrary
//     legacy cards inside the per-card handler requires the corpus
//     loader to find each name; rather than constructing partial card
//     stubs, we stamp a flag tracking which names have been chosen and
//     emitPartial for the actual spell creation step.
//   - Used-name tracking lives on perm.Flags as bits keyed by name.
func registerGarthOneEye(r *Registry) {
	r.OnActivated("Garth One-Eye", garthOneEyeActivate)
}

var garthChoices = []string{
	"Disenchant",
	"Braingeyser",
	"Terror",
	"Shivan Dragon",
	"Regrowth",
	"Black Lotus",
}

func garthOneEyeActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "garth_one_eye_choose_card"
	if gs == nil || src == nil {
		return
	}
	if src.Flags == nil {
		src.Flags = map[string]int{}
	}
	var choice string
	for _, name := range garthChoices {
		key := "garth_used_" + name
		if src.Flags[key] == 0 {
			choice = name
			src.Flags[key] = 1
			break
		}
	}
	if choice == "" {
		emitFail(gs, slug, src.Card.DisplayName(), "all_names_chosen", nil)
		return
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   src.Controller,
		"choice": choice,
	})
	emitPartial(gs, slug, src.Card.DisplayName(),
		"copy_of_named_card_creation_and_cast_not_modeled")
}
