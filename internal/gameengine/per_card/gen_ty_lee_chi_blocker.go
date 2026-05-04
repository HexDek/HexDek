package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTyLeeChiBlocker wires Ty Lee, Chi Blocker.
//
// Oracle text:
//
//   Flash
//   Prowess (Whenever you cast a noncreature spell, this creature gets
//     +1/+1 until end of turn.)
//   When Ty Lee enters, tap up to one target creature. It doesn't untap
//   during its controller's untap step for as long as you control Ty Lee.
//
// Implementation:
//   - Flash and Prowess are AST keywords.
//   - ETB: pick the strongest opponent creature (highest P+T, prefer
//     untapped) and tap it. Mark DoesNotUntap + Flags["skip_untap"] so
//     phases.go skips the target on its controller's untap step.
//   - The "as long as you control Ty Lee" cleanup (clearing the lock if
//     Ty Lee leaves the battlefield) requires an LTB hook the engine
//     does not yet expose for arbitrary tracked targets — we record the
//     source permanent's timestamp on the target so a future LTB sweep
//     can locate and release these locks. Emitted as partial.
func registerTyLeeChiBlocker(r *Registry) {
	r.OnETB("Ty Lee, Chi Blocker", tyLeeChiBlockerETB)
}

func tyLeeChiBlockerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ty_lee_chi_blocker_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	target := pickTyLeeLockTarget(gs, seat)
	if target == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"locked": false,
			"reason": "no_opponent_creature_target",
		})
		return
	}

	target.Tapped = true
	target.DoesNotUntap = true
	if target.Flags == nil {
		target.Flags = map[string]int{}
	}
	target.Flags["skip_untap"] = 1
	target.Flags["ty_lee_lock_source_ts"] = int(perm.Timestamp)

	gs.LogEvent(gameengine.Event{
		Kind:   "tap_permanent",
		Seat:   seat,
		Target: target.Controller,
		Source: "Ty Lee, Chi Blocker",
		Details: map[string]interface{}{
			"locked": target.Card.DisplayName(),
			"reason": "ty_lee_etb",
		},
	})

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          seat,
		"locked":        true,
		"target":        target.Card.DisplayName(),
		"target_seat":   target.Controller,
		"source_ts":     perm.Timestamp,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"lock_release_on_ty_lee_ltb_not_wired_yet")
}

// pickTyLeeLockTarget picks the most disruptive opponent creature to lock
// down. Priority: untapped creatures first (locking already-tapped is
// less impactful), then highest P+T, tiebreak by oldest Timestamp.
func pickTyLeeLockTarget(gs *gameengine.GameState, seat int) *gameengine.Permanent {
	if gs == nil {
		return nil
	}
	var best *gameengine.Permanent
	bestKey := struct {
		untapped bool
		score    int
		tsNeg    int
	}{}
	for _, opp := range gs.Opponents(seat) {
		os := gs.Seats[opp]
		if os == nil {
			continue
		}
		for _, p := range os.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			key := struct {
				untapped bool
				score    int
				tsNeg    int
			}{
				untapped: !p.Tapped,
				score:    p.Power() + p.Toughness(),
				tsNeg:    -p.Timestamp,
			}
			if best == nil ||
				(key.untapped && !bestKey.untapped) ||
				(key.untapped == bestKey.untapped && key.score > bestKey.score) ||
				(key.untapped == bestKey.untapped && key.score == bestKey.score && key.tsNeg > bestKey.tsNeg) {
				best = p
				bestKey = key
			}
		}
	}
	return best
}
