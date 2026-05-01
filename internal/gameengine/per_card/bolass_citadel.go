package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBolassCitadel wires up Bolas's Citadel.
//
// Oracle text:
//
//	You may look at the top card of your library any time.
//	You may play lands and cast spells from the top of your library.
//	If you cast a spell this way, pay life equal to its mana value
//	rather than pay its mana cost.
//	{T}, Sacrifice ten nontoken permanents: Each opponent loses 10
//	life.
//
// THE Aetherflux Reservoir combo: Citadel lets you cast from the top
// paying life instead of mana, Aetherflux lifegains on every cast, so
// a sufficiently low-curve top-of-library produces a loop that gains
// 1-per-cast life and eventually activates Aetherflux for 50 damage.
//
// Implementation:
//   - OnETB: register a ZoneCastPermission for the controller's top-
//     of-library cards (library_cast keyword with life cost). This
//     integrates with the engine's zone_cast.go CastFromZone primitive
//     so the Hat/AI can actually cast spells from the top of library
//     paying life instead of mana. Also set gs.Flags for the cast-
//     from-top presence check.
//   - OnActivated(0, ...): the "play top of library paying life" mode.
//     Moves the top card to hand and pays life = CMC as an MVP fallback
//     for when the zone-cast path isn't exercised directly. In normal
//     engine operation the CastFromZone path is preferred.
//   - OnActivated(1, ...): the sac-10-nontoken mode. Each opponent
//     loses 10 life. Sacrifice cost is assumed paid by the caller.
//
// The "you may look at the top card" clause is a no-op in the current
// observation model (no hidden information yet).
func registerBolassCitadel(r *Registry) {
	r.OnETB("Bolas's Citadel", bolassCitadelETB)
	r.OnActivated("Bolas's Citadel", bolassCitadelActivate)
}

func bolassCitadelETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "bolass_citadel_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["bolas_citadel_active_seat_"+intToStr(seat)] = perm.Timestamp

	// Register zone-cast permission for the top of library. The engine's
	// zone_cast.go CastFromZone path will consult this when the Hat/AI
	// decides to cast. The LifeCostInsteadOfMana field tells the engine
	// to pay life = CMC instead of mana.
	//
	// We grant the permission to the current top card of the library. As
	// the Hat exercises CastFromZone, subsequent top cards become eligible
	// via the citadel_active flag check in the zone-cast scanner.
	if seat >= 0 && seat < len(gs.Seats) {
		s := gs.Seats[seat]
		if len(s.Library) > 0 {
			topCard := s.Library[0]
			cmc := cardCMC(topCard)
			perm := gameengine.NewLibraryCastPermission(cmc)
			perm.RequireController = seat
			perm.SourceName = "Bolas's Citadel"
			gameengine.RegisterZoneCastGrant(gs, topCard, perm)
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          seat,
		"zone_cast":     "library_cast_with_life_cost",
		"integrated":    true,
	})
}

func bolassCitadelActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	if gs == nil || src == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	switch abilityIdx {
	case 0:
		// "Play top of library for life" mode. Move top card of
		// library into hand and pay life = its CMC. This is the
		// fallback path when the zone-cast primitive isn't exercised
		// directly by the Hat. Downstream zone-cast integration
		// (CastFromZone) handles the real cast-from-top pipeline.
		const slug = "bolass_citadel_play_top"
		if len(s.Library) == 0 {
			emitFail(gs, slug, src.Card.DisplayName(), "library_empty", nil)
			return
		}
		c := s.Library[0]
		cmc := cardCMC(c)
		gameengine.MoveCard(gs, c, seat, "library", "hand", "effect")
		s.Life -= cmc
		gs.LogEvent(gameengine.Event{
			Kind:   "lose_life",
			Seat:   seat,
			Target: seat,
			Source: src.Card.DisplayName(),
			Amount: cmc,
			Details: map[string]interface{}{
				"reason":      "bolass_citadel_cast_top_paying_life",
				"card_played": c.DisplayName(),
				"cmc":         cmc,
			},
		})
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":        seat,
			"card_played": c.DisplayName(),
			"life_paid":   cmc,
			"life_after":  s.Life,
		})

		// Re-register zone-cast grant for the new top card so the
		// Hat can continue the Citadel chain.
		if len(s.Library) > 0 {
			nextTop := s.Library[0]
			nextCMC := cardCMC(nextTop)
			perm := gameengine.NewLibraryCastPermission(nextCMC)
			perm.RequireController = seat
			perm.SourceName = "Bolas's Citadel"
			gameengine.RegisterZoneCastGrant(gs, nextTop, perm)
		}
		_ = gs.CheckEnd()
	case 1:
		// Sacrifice-10 → each opponent loses 10 life.
		const slug = "bolass_citadel_sac_ten"
		// We don't enforce the sac cost here (caller pays). Effect only.
		for _, opp := range gs.Opponents(seat) {
			os := gs.Seats[opp]
			os.Life -= 10
			gs.LogEvent(gameengine.Event{
				Kind:   "lose_life",
				Seat:   seat,
				Target: opp,
				Source: src.Card.DisplayName(),
				Amount: 10,
				Details: map[string]interface{}{
					"reason": "bolass_citadel_activated",
				},
			})
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":           seat,
			"opponents_hit":  len(gs.Opponents(seat)),
			"damage_per_opp": 10,
		})
		_ = gs.CheckEnd()
	}
}
