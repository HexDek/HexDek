package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAraumiOfTheDeadTideCustom replaces the auto-generated activated
// stub (which spawns a 1/1) with the actual encore-grant effect.
//
// Oracle text:
//
//	{T}, Exile cards from your graveyard equal to the number of
//	opponents you have: Target creature card in your graveyard gains
//	encore until end of turn. The encore cost is equal to its mana
//	cost. (Exile the creature card and pay its mana cost: For each
//	opponent, create a token copy that attacks that opponent this turn
//	if able. They gain haste. Sacrifice them at the beginning of the
//	next end step.)
//
// {T} is engine cost dispatch. The "exile N graveyard cards" part is
// NOT in the AST cost grammar — it's a custom additional cost that the
// gen_*.go template doesn't enforce. We enforce it here.
//
// Heuristic: pick the highest-power creature in our graveyard for the
// encore target; exile the lowest-CMC noncreature graveyard cards to
// pay the cost. Then immediately resolve the encore: spawn one token
// copy per opponent on the battlefield with haste + end-step sacrifice.
func registerAraumiOfTheDeadTideCustom(r *Registry) {
	r.OnActivated("Araumi of the Dead Tide", araumiEncore)
}

func araumiEncore(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "araumi_encore"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	// Count opponents (not lost).
	opps := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == src.Controller {
			continue
		}
		opps++
	}
	if opps <= 0 {
		emitFail(gs, slug, src.Card.DisplayName(), "no_opponents", nil)
		return
	}
	// Pick best creature in own graveyard.
	var bestCreature *gameengine.Card
	bestCreatureIdx := -1
	bestPower := -1
	for i, c := range seat.Graveyard {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		if c.BasePower > bestPower {
			bestCreature = c
			bestCreatureIdx = i
			bestPower = c.BasePower
		}
	}
	if bestCreature == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_creature_in_graveyard", nil)
		return
	}
	// Exile N additional cards from our graveyard (avoid double-exiling
	// the encore target).
	exiledCount := 0
	exileCandidates := []int{}
	for i, c := range seat.Graveyard {
		if i == bestCreatureIdx || c == nil {
			continue
		}
		exileCandidates = append(exileCandidates, i)
		if len(exileCandidates) >= opps {
			break
		}
	}
	if len(exileCandidates) < opps {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_graveyard_for_cost", map[string]interface{}{
			"need":      opps,
			"available": len(exileCandidates),
		})
		return
	}
	// Build new graveyard slice excluding the encore target and the
	// exiled cost cards. Move exiled to seat.Exile.
	skip := map[int]bool{bestCreatureIdx: true}
	for _, idx := range exileCandidates {
		skip[idx] = true
	}
	newGY := seat.Graveyard[:0]
	for i, c := range seat.Graveyard {
		if !skip[i] {
			newGY = append(newGY, c)
			continue
		}
		if i == bestCreatureIdx {
			// Encore exiles the target itself.
			seat.Exile = append(seat.Exile, c)
			continue
		}
		seat.Exile = append(seat.Exile, c)
		exiledCount++
	}
	seat.Graveyard = newGY
	// Spawn a token copy per opponent with haste; schedule sacrifice.
	tokens := []*gameengine.Permanent{}
	for i, s := range gs.Seats {
		if s == nil || s.Lost || i == src.Controller {
			continue
		}
		tokenCard := &gameengine.Card{
			Name:          bestCreature.DisplayName() + " (Araumi token)",
			Owner:         src.Controller,
			Types:         append([]string{"token"}, bestCreature.Types...),
			BasePower:     bestCreature.BasePower,
			BaseToughness: bestCreature.BaseToughness,
		}
		tokenPerm := &gameengine.Permanent{
			Card:       tokenCard,
			Controller: src.Controller,
			Owner:      src.Controller,
			Timestamp:  gs.NextTimestamp(),
			Counters:   map[string]int{},
			Flags:      map[string]int{"kw:haste": 1, "araumi_attacks_seat": i + 1},
		}
		seat.Battlefield = append(seat.Battlefield, tokenPerm)
		tokens = append(tokens, tokenPerm)
		_ = tokenPerm
	}
	// End-step sacrifice.
	for _, tk := range tokens {
		tk := tk
		gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
			TriggerAt:      "next_end_step",
			ControllerSeat: src.Controller,
			SourceCardName: src.Card.DisplayName(),
			OneShot:        true,
			EffectFn: func(gs *gameengine.GameState) {
				// SacrificePermanent will fire LTB triggers; the token
				// then ceases to exist per CR §111.10.
				gameengine.SacrificePermanent(gs, tk, "araumi_encore_eos")
			},
		})
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":          src.Controller,
		"target":        bestCreature.DisplayName(),
		"opps":          opps,
		"cost_exiled":   exiledCount,
		"tokens_spawned": len(tokens),
	})
}
