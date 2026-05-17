package per_card

import (
	"strconv"
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerHurkylMasterWizard wires Hurkyl, Master Wizard (Muninn parser-gap
// #96, 4.0K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{U}{U}
//	Legendary Creature — Human Wizard Advisor
//	At the beginning of your end step, if you've cast a noncreature
//	spell this turn, reveal the top five cards of your library. For each
//	card type among noncreature spells you've cast this turn, you may
//	put a card of that type from among the revealed cards into your
//	hand. Put the rest on the bottom of your library in a random order.
//
// Implementation:
//   - spell_cast listener gated on caster_seat == controller and noncreature
//     spell. Tally the SPELL's card types into a per-perm flag map so we
//     can recall the distinct type set at end step.
//   - end_step gated on controller == active seat. If we cast at least
//     one noncreature spell this turn, reveal top 5; for each tracked
//     type, grab the first revealed card matching that type and route
//     it to hand. Remaining cards go to the bottom of the library
//     (order randomization is approximated via reverse-deterministic
//     append — the engine has no per-turn priority RNG hook here).
func registerHurkylMasterWizard(r *Registry) {
	r.OnTrigger("Hurkyl, Master Wizard", "spell_cast", hurkylTrackNoncreatureCast)
	r.OnTrigger("Hurkyl, Master Wizard", "end_step", hurkylEndStep)
}

func hurkylTypeKey(turn int, t string) string {
	return "hurkyl_t" + strconv.Itoa(turn+1) + "_type_" + t
}

func hurkylCastFlagKey(turn int) string {
	return "hurkyl_t" + strconv.Itoa(turn+1) + "_cast"
}

// hurkylTrackedTypes returns the canonical list of card types we'll fan
// out across (matching the MTG "card type" definition, excluding
// supertypes/subtypes). Mirrors gameengine card-type taxonomy.
var hurkylTrackedTypes = []string{
	"artifact", "enchantment", "instant", "sorcery", "planeswalker", "land", "battle",
}

func hurkylTrackNoncreatureCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if cardHasType(card, "creature") {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags[hurkylCastFlagKey(gs.Turn)] = 1
	for _, t := range hurkylTrackedTypes {
		if cardHasType(card, t) {
			perm.Flags[hurkylTypeKey(gs.Turn, t)] = 1
		}
	}
}

func hurkylEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "hurkyl_end_step_reveal_grab"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	castFlag := hurkylCastFlagKey(gs.Turn)
	if perm.Flags[castFlag] == 0 {
		hurkylPruneKeys(perm, gs.Turn)
		return
	}
	// Collect tracked types for this turn.
	wantTypes := []string{}
	for _, t := range hurkylTrackedTypes {
		k := hurkylTypeKey(gs.Turn, t)
		if perm.Flags[k] == 1 {
			wantTypes = append(wantTypes, t)
		}
	}
	delete(perm.Flags, castFlag)
	for _, t := range hurkylTrackedTypes {
		delete(perm.Flags, hurkylTypeKey(gs.Turn, t))
	}
	hurkylPruneKeys(perm, gs.Turn)

	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	// Reveal top 5.
	reveal := make([]*gameengine.Card, 0, 5)
	for i := 0; i < 5 && i < len(seat.Library); i++ {
		reveal = append(reveal, seat.Library[i])
	}
	if len(reveal) == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":     perm.Controller,
			"revealed": 0,
		})
		return
	}
	// For each wanted type, claim the first matching revealed card.
	claimed := map[*gameengine.Card]bool{}
	claimedNames := []string{}
	for _, t := range wantTypes {
		for _, c := range reveal {
			if c == nil || claimed[c] {
				continue
			}
			if cardHasType(c, t) {
				claimed[c] = true
				claimedNames = append(claimedNames, c.DisplayName())
				gameengine.MoveCard(gs, c, perm.Controller, "library", "hand", "hurkyl_reveal_grab")
				break
			}
		}
	}
	// Remaining revealed cards go to the bottom (in revealed order; the
	// engine has no scoped RNG here — we keep it deterministic).
	for _, c := range reveal {
		if c == nil || claimed[c] {
			continue
		}
		// MoveCard library→library would no-op; manually relocate within
		// the library slice instead.
		for i, lc := range seat.Library {
			if lc == c {
				seat.Library = append(seat.Library[:i], seat.Library[i+1:]...)
				seat.Library = append(seat.Library, c)
				break
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"revealed": len(reveal),
		"types":    strings.Join(wantTypes, ","),
		"claimed":  strings.Join(claimedNames, ","),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"library_bottom_random_order_approximated_with_deterministic_reveal_order")
}

func hurkylPruneKeys(perm *gameengine.Permanent, currentTurn int) {
	if perm == nil || perm.Flags == nil {
		return
	}
	prefix := "hurkyl_t"
	cutoff := currentTurn + 1
	for k := range perm.Flags {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		rest := k[len(prefix):]
		under := strings.IndexByte(rest, '_')
		if under <= 0 {
			continue
		}
		n, err := strconv.Atoi(rest[:under])
		if err != nil {
			continue
		}
		if n < cutoff {
			delete(perm.Flags, k)
		}
	}
}
