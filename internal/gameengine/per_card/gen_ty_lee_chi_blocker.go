package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTyLeeChiBlocker wires Ty Lee, Chi Blocker.
//
// Oracle text:
//
//	Flash
//	Prowess (Whenever you cast a noncreature spell, this creature gets
//	+1/+1 until end of turn.)
//	When Ty Lee enters, tap up to one target creature. It doesn't
//	untap during its controller's untap step for as long as you
//	control Ty Lee.
//
// Implementation:
//   - Flash + prowess: handled by AST keyword pipeline (no per-card
//     wiring needed).
//   - ETB: pick the strongest opponent creature, tap it, and set
//     Flags["doesnt_untap"]=1 so the engine's untap step skips it
//     while Ty Lee remains in play.
//
// emitPartial: full "as long as you control Ty Lee" cleanup (clearing
// the doesnt_untap flag when Ty Lee leaves the battlefield) needs a
// permanent_ltb hook keyed to the locked target. We currently rely on
// the static-once-tagged behavior; engine-side TODO for full revert.
func registerTyLeeChiBlocker(r *Registry) {
	r.OnETB("Ty Lee, Chi Blocker", tyLeeETB)
	r.OnTrigger("Ty Lee, Chi Blocker", "permanent_ltb", tyLeeReleaseLock)
}

func tyLeeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "ty_lee_etb_tap_lock"
	if gs == nil || perm == nil {
		return
	}
	var best *gameengine.Permanent
	bestPow := -1
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == perm.Controller {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			pow := gs.PowerOf(p)
			if pow > bestPow {
				bestPow = pow
				best = p
			}
		}
	}
	if best == nil {
		return
	}
	best.Tapped = true
	if best.Flags == nil {
		best.Flags = map[string]int{}
	}
	best.Flags["doesnt_untap"] = 1
	best.Flags["ty_lee_locked_by"] = perm.Controller + 1 // +1 so 0-seat is non-zero
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"target":      best.Card.DisplayName(),
		"target_seat": best.Controller,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"target_choice_resolved_heuristically_strongest_opponent_creature")
}

func tyLeeReleaseLock(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ty_lee_release_lock"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	leavingPerm, _ := ctx["perm"].(*gameengine.Permanent)
	if leavingPerm != perm {
		return
	}
	released := 0
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Flags == nil {
				continue
			}
			if p.Flags["ty_lee_locked_by"] == perm.Controller+1 {
				delete(p.Flags, "doesnt_untap")
				delete(p.Flags, "ty_lee_locked_by")
				released++
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"released": released,
	})
}
