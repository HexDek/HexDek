package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerUreniTheSongUnendingCustom implements Ureni's land-count damage
// ETB. The auto-generated stub is a no-op.
//
// Oracle text:
//
//	Flying, protection from white and from black
//	When Ureni enters, it deals X damage divided as you choose among
//	any number of target creatures and/or planeswalkers your opponents
//	control, where X is the number of lands you control.
//
// Implementation notes:
//   - Flying / protection are static keywords handled by the AST pipeline.
//   - X = lands controlled at resolve time per CR §608.2b.
//   - Damage division: greedy assignment, biggest-toughness creatures
//     first up to lethal. Excess damage spills to the next-largest
//     opponent threat. Planeswalkers are eligible targets but de-
//     prioritized vs creatures (creature threats kill us faster).
func registerUreniTheSongUnendingCustom(r *Registry) {
	r.OnETB("Ureni, the Song Unending", ureniETB)
}

func ureniETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ureni_song_unending_damage"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return
	}

	// Count lands we control.
	x := 0
	for _, p := range seat.Battlefield {
		if p != nil && p.IsLand() {
			x++
		}
	}
	if x <= 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": seatIdx,
			"x":    0,
		})
		return
	}

	// Build a target list: opponent creatures sorted by toughness desc
	// (biggest threats first), then opponent planeswalkers.
	type target struct {
		perm *gameengine.Permanent
		t    int
		isPW bool
	}
	var targets []target
	for i, s := range gs.Seats {
		if i == seatIdx || s == nil || s.Lost || s.LeftGame {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil {
				continue
			}
			if p.IsCreature() {
				targets = append(targets, target{p, gs.ToughnessOf(p) - p.MarkedDamage, false})
			} else if p.IsPlaneswalker() {
				loyalty := 0
				if p.Counters != nil {
					loyalty = p.Counters["loyalty"]
				}
				targets = append(targets, target{p, loyalty, true})
			}
		}
	}
	// Sort: creatures before PWs, larger toughness first.
	for i := 0; i < len(targets); i++ {
		for j := i + 1; j < len(targets); j++ {
			ai, bj := targets[i], targets[j]
			swap := false
			if ai.isPW && !bj.isPW {
				swap = true
			} else if ai.isPW == bj.isPW && bj.t > ai.t {
				swap = true
			}
			if swap {
				targets[i], targets[j] = targets[j], targets[i]
			}
		}
	}

	// Greedy: kill from the top; leftover spills to the next.
	remaining := x
	hits := 0
	for _, tgt := range targets {
		if remaining <= 0 {
			break
		}
		dealt := tgt.t
		if dealt <= 0 {
			dealt = 1
		}
		if dealt > remaining {
			dealt = remaining
		}
		tgt.perm.MarkedDamage += dealt
		remaining -= dealt
		hits++
	}
	if remaining > 0 && len(targets) > 0 {
		// All targets killed; remainder goes to the first target as
		// overkill (engine just records MarkedDamage).
		targets[0].perm.MarkedDamage += remaining
	}
	gs.InvalidateCharacteristicsCache()
	_ = gs.CheckEnd()

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   seatIdx,
		"x":      x,
		"hits":   hits,
		"spill":  remaining,
	})
}
