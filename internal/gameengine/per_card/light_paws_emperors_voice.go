package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// Light-Paws, Emperor's Voice
//
// Oracle text (Scryfall, verified 2026-05-16):
//   Whenever an Aura you control enters, if you cast it, you may search
//   your library for an Aura card with mana value less than or equal to
//   that Aura and with a different name than each Aura you control, put
//   that card onto the battlefield attached to Light-Paws, then shuffle.
//
// Implementation:
//   - OnTrigger("permanent_etb"): fires whenever any permanent enters the
//     battlefield. Gates:
//       (a) entering permanent is an Aura (cardHasType "aura"),
//       (b) entering Aura is controlled by Light-Paws' controller, and
//       (c) entering.Flags["was_cast"] == 1 — the "if you cast it"
//           intervening-if (CR §603.6c). stack.go stamps was_cast on the
//           cast path only; blink, reanimate, and Light-Paws' own fetch
//           leave it unset, which naturally prevents the fetched Aura's
//           ETB from re-triggering this ability.
//   - Search library for the highest-CMC Aura with CMC <= entering Aura's
//     CMC whose name does NOT match any Aura the controller already
//     controls (including the entering Aura itself).
//   - Put found Aura onto battlefield attached to Light-Paws, shuffle.

func registerLightPawsEmperorsVoiceETB(r *Registry) {
	r.OnTrigger("Light-Paws, Emperor's Voice", "permanent_etb", lightPawsAuraETBTrigger)
}

func lightPawsAuraETBTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "light_paws_emperors_voice_aura_etb"
	if gs == nil || perm == nil || perm.Card == nil || ctx == nil {
		return
	}

	// perm is the Light-Paws permanent observing the trigger.
	if perm.Card.DisplayName() != "Light-Paws, Emperor's Voice" {
		return
	}

	// Extract the entering permanent from the trigger context.
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil || entering.Card == nil {
		return
	}

	// Gate: entering permanent must be an Aura.
	if !cardHasType(entering.Card, "aura") {
		return
	}

	// Gate: entering Aura must be under Light-Paws' controller.
	enteringSeat, _ := ctx["controller_seat"].(int)
	if enteringSeat != perm.Controller {
		return
	}

	// "If you cast it" — CR §603.6c. stack.go sets was_cast on the cast
	// path only. Blink / reanimate / Light-Paws' own fetch never set it,
	// so the trigger correctly no-ops on those (and never recurses on
	// the fetched Aura).
	if entering.Flags == nil || entering.Flags["was_cast"] != 1 {
		return
	}

	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	enteringCMC := entering.Card.CMC
	enteringName := entering.Card.DisplayName()

	// "Different name than each Aura you control" — collect the names of
	// every Aura currently on the controller's battlefield. The entering
	// Aura is already on the battlefield by the time permanent_etb fires
	// (see etb_dispatch.go), so this set already contains enteringName.
	controlledAuraNames := map[string]bool{}
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		if cardHasType(p.Card, "aura") {
			controlledAuraNames[p.Card.DisplayName()] = true
		}
	}

	// Search library for the best Aura: highest CMC that is still
	// <= the entering Aura's CMC, whose name doesn't collide with any
	// controlled Aura.
	bestIdx := -1
	bestCMC := -1
	for i, c := range s.Library {
		if c == nil || !cardHasType(c, "aura") {
			continue
		}
		if c.CMC > enteringCMC {
			continue
		}
		if controlledAuraNames[c.DisplayName()] {
			continue
		}
		if c.CMC > bestCMC {
			bestCMC = c.CMC
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		shuffleLibraryPerCard(gs, seat)
		emitFail(gs, slug, perm.Card.DisplayName(), "no_eligible_aura", map[string]interface{}{
			"entering_aura": enteringName,
			"max_cmc":       enteringCMC,
		})
		return
	}

	card := s.Library[bestIdx]
	gameengine.MoveCard(gs, card, seat, "library", "battlefield", "light_paws_emperors_voice")

	// Find the permanent on the battlefield and attach to Light-Paws.
	var fetched *gameengine.Permanent
	for _, p := range gs.Seats[seat].Battlefield {
		if p != nil && p.Card == card {
			fetched = p
			break
		}
	}
	if fetched != nil {
		fetched.AttachedTo = perm
	}

	shuffleLibraryPerCard(gs, seat)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          seat,
		"entering_aura": enteringName,
		"entering_cmc":  enteringCMC,
		"found":         card.DisplayName(),
		"found_cmc":     card.CMC,
	})
}
