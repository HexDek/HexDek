package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTheOneRing wires up The One Ring (LTR, cEDH staple).
//
// Oracle text:
//
//	Indestructible
//	When The One Ring enters, if you cast it, you gain protection from
//	everything until your next turn.
//	At the beginning of your upkeep, you lose 1 life for each burden
//	counter on The One Ring.
//	{T}: Put a burden counter on The One Ring, then draw a card for
//	each burden counter on The One Ring.
//
// Engine coverage:
//   - Indestructible: stamp perm.Flags["indestructible"] (the keyword
//     pipeline already handles bare AST keywords, but the parser-gap
//     audit shows The One Ring is the #1 unparsed card so we stamp it
//     belt-and-suspenders here).
//   - ETB protection: gated on perm.Flags["was_cast"] (set by stack.go
//     when a permanent resolves through the cast path — blink/copy
//     paths leave it unset, matching the rules text "if you cast it").
//     We use the seat-level protection_from_everything flag (same
//     mechanism Teferi's Protection uses) which both prevention.go
//     and combat.go consult to bypass damage, and we register a
//     delayed trigger to clear the flag at the controller's next turn.
//   - Upkeep life loss: standard LoseLife per burden counter on the
//     controller's upkeep. Fires via FireCardTrigger("upkeep") emitted
//     from phases.FirePhaseTriggers.
//   - Activated tap ability: tap, add a burden counter, draw N cards
//     (N = burden counters AFTER incrementing, per the "then" clause).
func registerTheOneRing(r *Registry) {
	r.OnETB("The One Ring", theOneRingETB)
	r.OnActivated("The One Ring", theOneRingActivated)
	r.OnTrigger("The One Ring", "upkeep_controller", theOneRingUpkeep)
}

func theOneRingETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "the_one_ring_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller

	// Indestructible (CR §702.12). Belt-and-suspenders alongside the
	// keyword pipeline since the AST parser currently fails on this card.
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["indestructible"] = 1

	// "If you cast it" intervening-if (CR §603.6c). When the permanent
	// reached the battlefield via the cast path, stack.go sets was_cast.
	// Blink, reanimation, Sneak Attack, token copy — none of those set
	// was_cast, so the protection clause silently no-ops, matching the
	// rules text.
	if perm.Flags["was_cast"] != 1 {
		emit(gs, slug, "The One Ring", map[string]interface{}{
			"seat":              seat,
			"protection":        "skipped",
			"reason":            "not_cast",
			"indestructible":    true,
		})
		return
	}

	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s.Flags == nil {
		s.Flags = map[string]int{}
	}
	// Protection from everything (CR §702.16). The seat flag is consulted
	// by prevention.go and combat.go to prevent ALL damage to this player;
	// "from everything" also implies can't-be-targeted / can't-be-attached,
	// which the engine models elsewhere via prot:* on permanents — the
	// seat-level coverage matches the modeling Teferi's Protection uses.
	s.Flags["protection_from_everything"]++

	// Keep a per-card prevention shield too so the existing prevention
	// pipeline emits a damage_prevented event with The One Ring as the
	// source (downstream analytics / Heimdall replay use the shield's
	// SourceCard field).
	gameengine.AddPreventionShield(gs, gameengine.PreventionShield{
		TargetSeat: seat,
		Amount:     -1, // infinite while shield is live
		SourceCard: "The One Ring",
	})

	// "Until your next turn" — delayed trigger at the controller's next
	// turn (CR §603.7c). Same idiom as Teferi's Protection.
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "your_next_turn",
		ControllerSeat: seat,
		SourceCardName: "The One Ring",
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			if seat < 0 || seat >= len(gs.Seats) {
				return
			}
			s := gs.Seats[seat]
			if s != nil && s.Flags != nil && s.Flags["protection_from_everything"] > 0 {
				s.Flags["protection_from_everything"]--
				if s.Flags["protection_from_everything"] <= 0 {
					delete(s.Flags, "protection_from_everything")
				}
			}
			// Consume any unused One Ring prevention shields belonging to
			// this seat — the protection expires whether or not damage
			// was dealt.
			for i := range gs.PreventionShields {
				sh := &gs.PreventionShields[i]
				if sh.SourceCard == "The One Ring" && sh.TargetSeat == seat && !sh.Consumed {
					sh.Consumed = true
				}
			}
			gs.LogEvent(gameengine.Event{
				Kind:   "per_card_handler",
				Seat:   seat,
				Source: "The One Ring",
				Details: map[string]interface{}{
					"slug":   "the_one_ring_protection_expired",
					"effect": "protection_expired",
				},
			})
		},
	})

	emit(gs, slug, "The One Ring", map[string]interface{}{
		"seat":           seat,
		"protection":     "until_your_next_turn",
		"indestructible": true,
	})
}

func theOneRingActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "the_one_ring_draw"
	if gs == nil || src == nil {
		return
	}
	if src.Tapped {
		emitFail(gs, slug, "The One Ring", "already_tapped", nil)
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	// Pay the tap cost.
	src.Tapped = true

	// Put a burden counter on The One Ring.
	src.AddCounter("burden", 1)

	// "Then draw a card for each burden counter" — read AFTER the
	// increment, so the first activation draws 1.
	burdens := src.Counters["burden"]
	drawn := 0
	for i := 0; i < burdens; i++ {
		c := drawOne(gs, seat, "The One Ring")
		if c != nil {
			drawn++
		}
	}

	emit(gs, slug, "The One Ring", map[string]interface{}{
		"seat":    seat,
		"burdens": burdens,
		"drawn":   drawn,
	})
}

func theOneRingUpkeep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "the_one_ring_upkeep"
	if gs == nil || perm == nil {
		return
	}
	// "At the beginning of your upkeep" — controller gating. The phases
	// dispatch threads active_seat through ctx; cross-check against
	// gs.Active as a fallback for callers that omit it.
	activeSeat, ok := ctx["active_seat"].(int)
	if !ok {
		activeSeat = gs.Active
	}
	if activeSeat != perm.Controller {
		return
	}
	seat := perm.Controller
	burdens := 0
	if perm.Counters != nil {
		burdens = perm.Counters["burden"]
	}
	if burdens <= 0 {
		return
	}
	gameengine.LoseLife(gs, seat, burdens, "The One Ring")
	emit(gs, slug, "The One Ring", map[string]interface{}{
		"seat":      seat,
		"burdens":   burdens,
		"life_lost": burdens,
		"life_now":  gs.Seats[seat].Life,
	})
	_ = gs.CheckEnd()
}
