package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// isSorcerySpeed reports whether `seatIdx` could legally take a
// sorcery-speed action right now: they're the active player, the stack
// is empty, and the phase is one of their two main phases. Mirrors the
// minimal subset CR §307.1 / §608 expect; per-card handlers use it to
// gate "Activate only as a sorcery" abilities defensively when callers
// reach them through paths that skip the engine's stack-time check
// (test fixtures, AI evaluator probes, replay rebuilds).
func isSorcerySpeed(gs *gameengine.GameState, seatIdx int) bool {
	if gs == nil {
		return false
	}
	if gs.Active != seatIdx {
		return false
	}
	if len(gs.Stack) > 0 {
		return false
	}
	switch gs.Phase {
	case "precombat_main", "postcombat_main", "main", "main1", "main2":
		return true
	}
	return false
}

// payManaFromPool deducts `amount` from the seat's mana pool if it can
// cover, returning true. Returns false (and does NOT decrement) when
// the pool is short. Use as a per-handler defensive cost gate when the
// engine's activated-ability dispatcher might be skipped (test paths,
// AI evaluator, replay rebuild).
func payManaFromPool(seat *gameengine.Seat, amount int) bool {
	if seat == nil {
		return false
	}
	if amount <= 0 {
		return true
	}
	if seat.ManaPool < amount {
		return false
	}
	seat.ManaPool -= amount
	return true
}
