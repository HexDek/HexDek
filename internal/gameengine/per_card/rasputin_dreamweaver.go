package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRasputinDreamweaver wires Rasputin Dreamweaver.
//
// Oracle text:
//
//	Rasputin enters with seven dream counters on it.
//	Remove a dream counter from Rasputin: Add {C}.
//	Remove a dream counter from Rasputin: Prevent the next 1 damage
//	that would be dealt to Rasputin this turn.
//	At the beginning of your upkeep, if Rasputin started the turn
//	untapped, put a dream counter on it.
//	Rasputin can't have more than seven dream counters on it.
//
// Implementation:
//   - OnETB stamps seven dream counters.
//   - OnTrigger("untap_step") snapshots the tapped state in
//     perm.Flags["rasputin_started_untapped"]; "started the turn untapped"
//     is interpreted as "was untapped at the moment of the untap step
//     trigger". The engine performs the untap-step trigger before the
//     auto-untap loop in phases.go.
//   - OnTrigger("upkeep_controller"): if the snapshot flag is set and the
//     dream counter count is below 7, add one dream counter (capped).
//   - OnActivated handles both activated abilities by abilityIdx:
//       0 → remove a dream counter, add {C}.
//       1 → remove a dream counter, install a 1-damage prevention shield
//           on Rasputin until end of turn.
//     Both abilities pay their own cost (decrement the dream counter)
//     and emitFail when no counters are available.
func registerRasputinDreamweaver(r *Registry) {
	r.OnETB("Rasputin Dreamweaver", rasputinETB)
	r.OnTrigger("Rasputin Dreamweaver", "untap_step", rasputinUntapSnapshot)
	r.OnTrigger("Rasputin Dreamweaver", "upkeep_controller", rasputinUpkeep)
	r.OnActivated("Rasputin Dreamweaver", rasputinActivate)
}

func rasputinETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "rasputin_dream_counters"
	if gs == nil || perm == nil {
		return
	}
	perm.AddCounter("dream", 7)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": 7,
	})
}

// rasputinUntapSnapshot records whether Rasputin was untapped at the
// untap step. The engine fires the untap-step trigger before the
// auto-untap loop, so this captures the canonical "started the turn"
// state per oracle text.
func rasputinUntapSnapshot(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	stepSeat, _ := ctx["seat"].(int)
	if stepSeat != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	if perm.Tapped {
		perm.Flags["rasputin_started_untapped"] = 0
	} else {
		perm.Flags["rasputin_started_untapped"] = 1
	}
}

func rasputinUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rasputin_dream_growth"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	upkeepSeat, _ := ctx["seat"].(int)
	if upkeepSeat != perm.Controller {
		return
	}
	if perm.Flags == nil || perm.Flags["rasputin_started_untapped"] != 1 {
		return
	}
	if perm.Counters["dream"] >= 7 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"capped": true,
		})
		return
	}
	perm.AddCounter("dream", 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": perm.Counters["dream"],
	})
}

func rasputinActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil {
		return
	}
	switch abilityIdx {
	case 0:
		const slug = "rasputin_mana_ability"
		if src.Counters["dream"] <= 0 {
			emitFail(gs, slug, src.Card.DisplayName(), "no_dream_counters", nil)
			return
		}
		src.Counters["dream"]--
		if src.Counters["dream"] <= 0 {
			delete(src.Counters, "dream")
		}
		gameengine.AddManaFromPermanent(gs, s, src, "C", 1)
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":      seat,
			"added":     1,
			"color":     "C",
			"remaining": src.Counters["dream"],
		})
	case 1:
		const slug = "rasputin_prevention_ability"
		if src.Counters["dream"] <= 0 {
			emitFail(gs, slug, src.Card.DisplayName(), "no_dream_counters", nil)
			return
		}
		src.Counters["dream"]--
		if src.Counters["dream"] <= 0 {
			delete(src.Counters, "dream")
		}
		gameengine.AddPreventionShield(gs, gameengine.PreventionShield{
			TargetSeat: -1,
			TargetPerm: src,
			Amount:     1,
			SourceCard: src.Card.DisplayName(),
			OneShot:    true,
		})
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":      seat,
			"prevent":   1,
			"remaining": src.Counters["dream"],
		})
	default:
		emitFail(gs, "rasputin_unknown_ability", src.Card.DisplayName(), "unknown_ability_idx", map[string]interface{}{
			"idx": abilityIdx,
		})
	}
}
