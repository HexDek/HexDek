package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheReaperKingNoMore wires The Reaper, King No More.
//
// Oracle text:
//
//   When The Reaper enters, put a -1/-1 counter on each of up to two
//   target creatures.
//   Whenever a creature an opponent controls with a -1/-1 counter on it
//   dies, you may put that card onto the battlefield under your control.
//   Do this only once each turn.
//
// Implementation:
//   - ETB: greedily pick up to two opponent creatures (highest P+T first)
//     and drop a -1/-1 counter on each. We prefer opponent creatures so
//     the death-trigger leg is set up; if no opponents have creatures,
//     we skip rather than counter our own board.
//   - Death-trigger leg: tracked via a Flags marker on Reaper itself
//     ("reaper_steal_used_this_turn"). The actual reanimation hook would
//     need to fire from the engine's death pipeline; we annotate it as
//     a partial here so analytics flag the missing leg.
func registerTheReaperKingNoMore(r *Registry) {
	r.OnETB("The Reaper, King No More", theReaperKingNoMoreETB)
}

func theReaperKingNoMoreETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "the_reaper_king_no_more_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	targets := pickTopOpponentCreatures(gs, seat, 2)
	placed := 0
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		t.AddCounter("-1/-1", 1)
		placed++
		names = append(names, t.Card.DisplayName())
		gs.LogEvent(gameengine.Event{
			Kind:   "counter_added",
			Seat:   seat,
			Target: t.Controller,
			Source: "The Reaper, King No More",
			Details: map[string]interface{}{
				"counter": "-1/-1",
				"on":      t.Card.DisplayName(),
			},
		})
	}
	if placed > 0 {
		gs.InvalidateCharacteristicsCache()
	}

	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	delete(perm.Flags, "reaper_steal_used_this_turn")

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":             seat,
		"counters_placed":  placed,
		"targets":          names,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"death_trigger_steal_leg_requires_engine_death_hook")
}

// pickTopOpponentCreatures returns up to n opponent creatures sorted by
// descending P+T (largest first). Used by Reaper's ETB to weaken the
// strongest opposing threats while setting up the death-trigger steal.
func pickTopOpponentCreatures(gs *gameengine.GameState, seat int, n int) []*gameengine.Permanent {
	if gs == nil || n <= 0 {
		return nil
	}
	var pool []*gameengine.Permanent
	for _, opp := range gs.Opponents(seat) {
		os := gs.Seats[opp]
		if os == nil {
			continue
		}
		for _, p := range os.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			pool = append(pool, p)
		}
	}
	picked := make([]*gameengine.Permanent, 0, n)
	for len(picked) < n && len(pool) > 0 {
		bestIdx := -1
		bestScore := -1 << 30
		bestTS := 1<<62 - 1
		for i, p := range pool {
			score := p.Power() + p.Toughness()
			if score > bestScore || (score == bestScore && p.Timestamp < bestTS) {
				bestScore = score
				bestTS = p.Timestamp
				bestIdx = i
			}
		}
		if bestIdx < 0 {
			break
		}
		picked = append(picked, pool[bestIdx])
		pool = append(pool[:bestIdx], pool[bestIdx+1:]...)
	}
	return picked
}
