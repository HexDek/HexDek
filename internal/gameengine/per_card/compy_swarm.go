package per_card

import (
	"fmt"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerCompySwarm wires Compy Swarm (Muninn parser-gap #69, ~9.5K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{B}{G}
//	Creature — Dinosaur
//	At the beginning of your end step, if a creature died this turn,
//	create a tapped token that's a copy of this creature.
//
// Implementation:
//   - "creature_dies" listener stamps a per-turn flag on the controller's
//     seat. "A creature died this turn" — any creature, any controller.
//   - "end_step" listener gated on active_seat == controller: if the flag
//     is set, mint a tapped token copy of Compy Swarm. We deep-copy the
//     printed stats and types, tag with "token" so IsToken() reports true,
//     and route through enterBattlefieldWithETB so ETB/ZCT triggers fire.
func registerCompySwarm(r *Registry) {
	r.OnTrigger("Compy Swarm", "creature_dies", compySwarmCreatureDies)
	r.OnTrigger("Compy Swarm", "end_step", compySwarmEndStep)
}

func compySwarmCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[compySwarmDiedKey(gs.Turn)] = 1
}

func compySwarmEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "compy_swarm_end_step_token_copy"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok || activeSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	key := compySwarmDiedKey(gs.Turn)
	if seat.Flags == nil || seat.Flags[key] == 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "no_creature_died_this_turn",
		})
		return
	}
	delete(seat.Flags, key)
	compySwarmPruneKeys(seat, gs.Turn)

	src := perm.Card
	types := append([]string{}, src.Types...)
	hasToken := false
	for _, t := range types {
		if t == "token" {
			hasToken = true
			break
		}
	}
	if !hasToken {
		types = append(types, "token")
	}
	token := &gameengine.Card{
		Name:          src.DisplayName() + " (Compy token)",
		Owner:         perm.Controller,
		BasePower:     src.BasePower,
		BaseToughness: src.BaseToughness,
		Types:         types,
		Colors:        append([]string{}, src.Colors...),
		TypeLine:      src.TypeLine,
	}
	tok := enterBattlefieldWithETB(gs, perm.Controller, token, true)
	tokenName := ""
	if tok != nil {
		tokenName = token.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"token": tokenName,
	})
}

func compySwarmDiedKey(turn int) string {
	return fmt.Sprintf("compy_died_t%d", turn+1)
}

func compySwarmPruneKeys(seat *gameengine.Seat, currentTurn int) {
	if seat == nil || seat.Flags == nil {
		return
	}
	prefix := "compy_died_t"
	cutoff := currentTurn + 1
	for k := range seat.Flags {
		if len(k) <= len(prefix) || k[:len(prefix)] != prefix {
			continue
		}
		n := 0
		_, err := fmt.Sscanf(k[len(prefix):], "%d", &n)
		if err != nil {
			continue
		}
		if n < cutoff {
			delete(seat.Flags, k)
		}
	}
}
