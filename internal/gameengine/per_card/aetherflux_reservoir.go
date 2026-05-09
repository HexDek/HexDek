package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAetherfluxReservoir wires up Aetherflux Reservoir.
//
// Oracle text:
//
//	Whenever you cast a spell, you gain 1 life for each spell you've
//	cast this turn.
//	Pay 50 life: Aetherflux Reservoir deals 50 damage to any target.
//	Activate only as a sorcery.
//
// Wincon for Oloro-style lifegain decks. The "cast this turn" counter is
// already tracked by the storm agent's cast_counts.go (shared state). We
// re-compute it conservatively by scanning the event log for this turn's
// "cast" events from the controller — if cast_counts.go is not yet
// landed, we fall back to 1 (the triggering spell itself counts).
//
// Handlers:
//   - OnTrigger("spell_cast", ...) — gain N life where N = cast count.
//   - OnActivated(0, ...) — pay 50 life, deal 50 damage to ctx["target_seat"].
func registerAetherfluxReservoir(r *Registry) {
	r.OnTrigger("Aetherflux Reservoir", "spell_cast", aetherfluxOnSpellCast)
	r.OnActivated("Aetherflux Reservoir", aetherfluxActivate)
}

func aetherfluxOnSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "aetherflux_reservoir_on_cast"
	if gs == nil || perm == nil {
		return
	}
	// Only trigger when the caster is the Reservoir's controller. Oracle
	// text is "Whenever YOU cast a spell".
	casterSeatAny := ctx["caster_seat"]
	casterSeat, _ := casterSeatAny.(int)
	if casterSeat != perm.Controller {
		return
	}
	// Read centralized cast counter — includes the triggering spell.
	casts := gs.Seats[perm.Controller].Turn.SpellsCast
	if casts < 1 {
		casts = 1
	}
	// Gain `casts` life.
	gameengine.GainLife(gs, perm.Controller, casts, perm.Card.DisplayName())
	gs.LogEvent(gameengine.Event{
		Kind:   "gain_life",
		Seat:   perm.Controller,
		Target: perm.Controller,
		Source: perm.Card.DisplayName(),
		Amount: casts,
		Details: map[string]interface{}{
			"reason":       "aetherflux_reservoir",
			"casts_this_turn": casts,
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":            perm.Controller,
		"casts_this_turn": casts,
		"life_gained":     casts,
	})
}

func aetherfluxActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "aetherflux_reservoir_activate"
	if gs == nil || src == nil {
		return
	}
	seat := src.Controller
	s := gs.Seats[seat]
	if s.Life < 50 {
		emitFail(gs, slug, src.Card.DisplayName(), "insufficient_life", map[string]interface{}{
			"life": s.Life,
		})
		return
	}
	// Pay 50 life.
	gameengine.LoseLife(gs, seat, 50, src.Card.DisplayName())
	// Deal 50 damage to target.
	targetSeat := -1
	if v, ok := ctx["target_seat"].(int); ok {
		targetSeat = v
	}
	if targetSeat < 0 || targetSeat >= len(gs.Seats) {
		// Default to first opponent.
		opps := gs.Opponents(seat)
		if len(opps) > 0 {
			targetSeat = opps[0]
		}
	}
	if targetSeat >= 0 && targetSeat < len(gs.Seats) {
		gameengine.DealDamage(gs, targetSeat, 50, src.Card.DisplayName())
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":        seat,
			"target_seat": targetSeat,
			"damage":      50,
		})
		_ = gs.CheckEnd()
	}
}

