package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerProgenitus wires Progenitus.
//
// Oracle text:
//
//	Protection from everything
//	If Progenitus would be put into a graveyard from anywhere, reveal
//	Progenitus and shuffle it into its owner's library instead.
//
// Implementation:
//   - "Protection from everything" — set the prot:* flag (engine
//     attackerHasProtectionFrom recognizes "*" as the universal
//     protection sentinel for both blocking and damage-prevention).
//   - "Shuffle into library instead of graveyard" — register a
//     zone-change replacement on creature_dies and a generic graveyard
//     entry. The full replacement layer is engine-side; we handle the
//     death case by listening on creature_dies for Progenitus and
//     moving the card from graveyard back to library + shuffling.
func registerProgenitus(r *Registry) {
	r.OnETB("Progenitus", progenitusETBSetProtectionFlag)
	r.OnTrigger("Progenitus", "creature_dies", progenitusShuffleBackOnDeath)
}

func progenitusETBSetProtectionFlag(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "progenitus_protection_from_everything"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["prot:*"] = 1
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
}

func progenitusShuffleBackOnDeath(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "progenitus_shuffle_back_on_death"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	dying, _ := ctx["perm"].(*gameengine.Permanent)
	if dying == nil {
		// Fall back to identity-by-card-name in case the perm pointer
		// has already been moved off the battlefield.
		card, _ := ctx["card"].(*gameengine.Card)
		if card == nil || card.DisplayName() != "Progenitus" {
			return
		}
		dying = perm
	}
	if dying != perm && dying.Card != nil && dying.Card.DisplayName() != "Progenitus" {
		return
	}
	owner := perm.Owner
	if owner < 0 || owner >= len(gs.Seats) {
		owner = perm.Controller
	}
	ownerSeat := gs.Seats[owner]
	if ownerSeat == nil {
		return
	}
	// Find and remove from graveyard.
	moved := false
	for i, c := range ownerSeat.Graveyard {
		if c == perm.Card {
			ownerSeat.Graveyard = append(ownerSeat.Graveyard[:i], ownerSeat.Graveyard[i+1:]...)
			ownerSeat.Library = append(ownerSeat.Library, c)
			moved = true
			break
		}
	}
	if moved && len(ownerSeat.Library) > 1 && gs.Rng != nil {
		gs.Rng.Shuffle(len(ownerSeat.Library), func(i, j int) {
			ownerSeat.Library[i], ownerSeat.Library[j] = ownerSeat.Library[j], ownerSeat.Library[i]
		})
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  owner,
		"moved": moved,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"true graveyard-entry replacement needs ZoneChange replacement hook; this fires post-graveyard and rescues")
}
