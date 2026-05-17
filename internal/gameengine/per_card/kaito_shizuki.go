package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKaitoShizuki wires Kaito Shizuki (Muninn parser-gap #23, 57,388 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{U}{B}
//	Legendary Planeswalker — Kaito  (Loyalty 3)
//	At the beginning of your end step, if Kaito entered this turn, he
//	phases out.
//	+1: Draw a card. Then discard a card unless you attacked this turn.
//	-2: Create a 1/1 blue Ninja creature token with "This token can't be
//	    blocked."
//	-7: You get an emblem with "Whenever a creature you control deals
//	    combat damage to a player, search your library for a blue or
//	    black creature card, put it onto the battlefield, then shuffle."
//
// Implementation:
//   - OnETB: pin starting loyalty to 3 and mark perm.Flags["kaito_etb_turn"]
//     = gs.Turn+1 (CR §603.4 intervening-if check at end of turn).
//   - OnTrigger("end_step"): if active seat is controller AND
//     perm.Flags["kaito_etb_turn"] == gs.Turn+1, phase Kaito out.
//   - OnActivated:
//       idx 0 (+1): draw 1; if seat.Flags["attacked_this_turn"] == 0,
//         discard the lowest-CMC card.
//       idx 1 (-2): create a 1/1 blue Ninja token that can't be blocked.
//       idx 2 (-7): emblem is logged via emitPartial — engine has no
//         emblem registry runtime that fires the combat-damage tutor.
func registerKaitoShizuki(r *Registry) {
	r.OnETB("Kaito Shizuki", kaitoShizukiETB)
	r.OnTrigger("Kaito Shizuki", "end_step", kaitoShizukiEndStep)
	r.OnActivated("Kaito Shizuki", kaitoShizukiActivate)
}

func kaitoShizukiETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	if gs == nil || perm == nil {
		return
	}
	if perm.Counters == nil {
		perm.Counters = map[string]int{}
	}
	if perm.Counters["loyalty"] == 0 {
		perm.Counters["loyalty"] = 3
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	// Store turn+1 so a Turn 0 ETB is still distinguishable from "no ETB
	// recorded" (map zero-value).
	perm.Flags["kaito_etb_turn"] = gs.Turn + 1
}

func kaitoShizukiEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kaito_phase_out"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	if perm.Flags == nil || perm.Flags["kaito_etb_turn"] != gs.Turn+1 {
		return
	}
	if perm.PhasedOut {
		return
	}
	perm.PhasedOut = true
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"rule": "702.26",
		"turn": gs.Turn,
	})
}

func kaitoShizukiActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	switch abilityIdx {
	case 0:
		kaitoShizukiPlusOne(gs, src)
	case 1:
		kaitoShizukiMinusTwo(gs, src)
	case 2:
		kaitoShizukiMinusSeven(gs, src)
	}
}

func kaitoShizukiPlusOne(gs *gameengine.GameState, src *gameengine.Permanent) {
	const slug = "kaito_plus1_draw_maybe_discard"
	src.AddCounter("loyalty", 1)
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	drawOne(gs, src.Controller, src.Card.DisplayName())

	attacked := 0
	if seat.Flags != nil {
		attacked = seat.Flags["attacked_this_turn"]
	}
	discarded := ""
	if attacked == 0 && len(seat.Hand) > 0 {
		var worst *gameengine.Card
		worstCMC := 1 << 30
		for _, c := range seat.Hand {
			if c == nil {
				continue
			}
			cmc := cardCMC(c)
			if cmc < worstCMC {
				worstCMC = cmc
				worst = c
			}
		}
		if worst != nil {
			discarded = worst.DisplayName()
			gameengine.DiscardCard(gs, worst, src.Controller)
		}
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":      src.Controller,
		"loyalty":   src.Counters["loyalty"],
		"attacked":  attacked,
		"discarded": discarded,
	})
}

func kaitoShizukiMinusTwo(gs *gameengine.GameState, src *gameengine.Permanent) {
	const slug = "kaito_minus2_ninja_token"
	src.AddCounter("loyalty", -2)
	token := &gameengine.Card{
		Name:          "Ninja Token",
		Owner:         src.Controller,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "creature", "ninja"},
		Colors:        []string{"U"},
		TypeLine:      "Token Creature — Ninja",
	}
	perm := enterBattlefieldWithETB(gs, src.Controller, token, false)
	if perm != nil {
		if perm.Flags == nil {
			perm.Flags = map[string]int{}
		}
		perm.Flags["unblockable"] = 1
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":    src.Controller,
		"loyalty": src.Counters["loyalty"],
	})
}

func kaitoShizukiMinusSeven(gs *gameengine.GameState, src *gameengine.Permanent) {
	const slug = "kaito_minus7_emblem"
	src.AddCounter("loyalty", -7)
	seat := gs.Seats[src.Controller]
	if seat != nil {
		if seat.Flags == nil {
			seat.Flags = map[string]int{}
		}
		seat.Flags["kaito_emblem"] = 1
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":    src.Controller,
		"loyalty": src.Counters["loyalty"],
	})
	emitPartial(gs, slug, src.Card.DisplayName(),
		"emblem_combat_damage_recursive_creature_tutor_partial_no_emblem_trigger_runtime")
}
