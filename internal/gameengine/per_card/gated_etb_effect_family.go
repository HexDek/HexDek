package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// gated_etb_effect_family.go — generic handler for the
// "When ~ enters, if <self-gate>, <effect>" family.
//
// Many gap cards share this shape but vary in two orthogonal axes:
//   - the self-gate ("if you cast it", "if it isn't a token", "if you cast
//     it from your hand", or no gate at all), and
//   - the body (shuffle/draw, create self-copies, free-cast another spell,
//     destroy stuff, etc.).
//
// The dispatch + slug-emission + gate evaluation are identical across
// every member. Each entry plugs in its own gate enum + body closure.
//
// Adding a new family member is one entry in gatedEtbEffectEntries.

type etbSelfGate int

const (
	// No intervening-if — fire whenever the card enters.
	etbGateNone etbSelfGate = iota
	// Fire only when the permanent was cast (perm.Flags["was_cast"] != 0).
	// Stack.go stamps this when the card resolves through the normal
	// cast/resolve pipeline. Reanimate/blink/copy paths leave it unset.
	etbGateWasCast
	// Fire only when the permanent was cast from its owner's hand
	// (perm.Flags["cast_from_hand"] != 0). Strict subset of WasCast.
	etbGateCastFromHand
	// Fire only when the permanent is not a token. Token copies of a
	// non-token card carry "token" in card.Types thanks to
	// resolveCreateTokenCopy.
	etbGateNotToken
	// Fire only when the permanent entered via the Sneak alternate
	// cost. ninja_sneak.go stamps perm.Flags["sneak_entry"] = 1 on
	// the entering permanent when sneak resolves.
	etbGateSneakEntry
)

type gatedEtbEffectEntry struct {
	cardName string
	gate     etbSelfGate
	effect   func(gs *gameengine.GameState, perm *gameengine.Permanent)
	partial  string // optional emitPartial reason if the body is approximate
}

// Hand-rolled siblings that ALSO match one of the gate shapes above
// (Skitterbeam Battalion: was_cast → create N self-copies; Gruff Triplets:
// not_token → create N self-copies; Crackling Spellslinger / Eon Frolicker
// / Bringer of the Last Gift / Breaching Leviathan / Tiamat: was_cast →
// bespoke body) keep their bespoke handlers — this file owns only the gap
// cards whose ETB body wasn't already covered by a sibling per-card file
// by the time the family scaffold landed. Each entry plugs in its own
// body closure; the scaffold dispatches gate-check + emit/emitPartial.
var gatedEtbEffectEntries = []gatedEtbEffectEntry{
	{
		// Weftwalking — {4}{U}{U}, enchantment.
		//   When this enchantment enters, if you cast it, shuffle your
		//   hand and graveyard into your library, then draw seven cards.
		//   The first spell each player casts during each of their turns
		//   may be cast without paying its mana cost.
		// The static "first spell free" half is a global cast-cost
		// replacement we don't yet model — flagged via partial.
		cardName: "Weftwalking",
		gate:     etbGateWasCast,
		effect:   weftwalkingShuffleAndDraw,
		partial:  "static_first_spell_each_turn_cost_free_replacement_unmodeled",
	},
	{
		// Leonardo, Leader in Blue — {W}, 1/1 Legendary Mutant Ninja
		// Turtle with sneak.
		//   Sneak {3}{W}{W} (You may cast this spell for {3}{W}{W} if
		//   you also return an unblocked attacker you control to hand
		//   during the declare blockers step. He enters tapped and
		//   attacking.)
		//   When Leonardo enters, if his sneak cost was paid, creatures
		//   you control get +2/+0 until end of turn.
		//   {1}{W}: Leonardo gains first strike until end of turn.
		// The {1}{W} first-strike activation is an activated ability we
		// don't synthesize here. Sneak's alternate-cost machinery lives
		// in ninja_sneak.go and stamps perm.Flags["sneak_entry"]; we
		// gate on that.
		cardName: "Leonardo, Leader in Blue",
		gate:     etbGateSneakEntry,
		effect:   leonardoLeaderInBlueBuffTeam,
		partial:  "first_strike_activated_ability_one_white_unmodeled",
	},
}

func registerGatedEtbEffectFamily(r *Registry) {
	for _, e := range gatedEtbEffectEntries {
		e := e
		r.OnETB(e.cardName, func(gs *gameengine.GameState, perm *gameengine.Permanent) {
			runGatedEtbEffect(gs, perm, e)
		})
	}
}

func runGatedEtbEffect(gs *gameengine.GameState, perm *gameengine.Permanent, e gatedEtbEffectEntry) {
	slug := "gated_etb_effect_family:" + landFetchSlug(e.cardName)
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if !etbGateOpen(perm, e.gate) {
		emitFail(gs, slug, perm.Card.DisplayName(), etbGateFailReason(e.gate), map[string]interface{}{
			"seat": perm.Controller,
		})
		return
	}
	if e.effect != nil {
		e.effect(gs, perm)
	}
	_ = gs.CheckEnd()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
		"gate": etbGateName(e.gate),
	})
	if e.partial != "" {
		emitPartial(gs, slug, perm.Card.DisplayName(), e.partial)
	}
}

func etbGateOpen(perm *gameengine.Permanent, gate etbSelfGate) bool {
	switch gate {
	case etbGateNone:
		return true
	case etbGateWasCast:
		return perm.Flags != nil && perm.Flags["was_cast"] != 0
	case etbGateCastFromHand:
		return perm.Flags != nil && perm.Flags["cast_from_hand"] != 0
	case etbGateNotToken:
		if perm.Card == nil {
			return false
		}
		for _, t := range perm.Card.Types {
			if strings.EqualFold(t, "token") {
				return false
			}
		}
		return true
	case etbGateSneakEntry:
		return perm.Flags != nil && perm.Flags["sneak_entry"] != 0
	}
	return false
}

func etbGateFailReason(gate etbSelfGate) string {
	switch gate {
	case etbGateWasCast:
		return "not_cast"
	case etbGateCastFromHand:
		return "not_cast_from_hand"
	case etbGateNotToken:
		return "is_token_copy"
	case etbGateSneakEntry:
		return "sneak_cost_not_paid"
	}
	return "gate_closed"
}

func etbGateName(gate etbSelfGate) string {
	switch gate {
	case etbGateNone:
		return "none"
	case etbGateWasCast:
		return "was_cast"
	case etbGateCastFromHand:
		return "cast_from_hand"
	case etbGateNotToken:
		return "not_token"
	case etbGateSneakEntry:
		return "sneak_entry"
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// Card bodies.
// ---------------------------------------------------------------------------

func leonardoLeaderInBlueBuffTeam(gs *gameengine.GameState, perm *gameengine.Permanent) {
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}
	ts := gs.NextTimestamp()
	buffed := 0
	for _, p := range s.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		p.Modifications = append(p.Modifications, gameengine.Modification{
			Power:     2,
			Duration:  "until_end_of_turn",
			Timestamp: ts,
		})
		buffed++
	}
	if buffed > 0 {
		gs.InvalidateCharacteristicsCache()
	}
	gs.LogEvent(gameengine.Event{
		Kind:   "buff",
		Seat:   seat,
		Source: perm.Card.DisplayName(),
		Details: map[string]interface{}{
			"slug":    "gated_etb_effect_family:leonardo_leader_in_blue",
			"buffed":  buffed,
			"power":   2,
		},
	})
}

func weftwalkingShuffleAndDraw(gs *gameengine.GameState, perm *gameengine.Permanent) {
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}
	src := perm.Card.DisplayName()

	// 1. Move hand → library.
	moved := 0
	for _, c := range s.Hand {
		if c == nil {
			continue
		}
		s.Library = append(s.Library, c)
		moved++
	}
	s.Hand = nil

	// 2. Move graveyard → library (skip the entering Weftwalking — it's a
	// permanent on the battlefield at this point, but defensive: a
	// concurrent grave entry would still be swept up).
	for _, c := range s.Graveyard {
		if c == nil {
			continue
		}
		s.Library = append(s.Library, c)
		moved++
	}
	s.Graveyard = nil

	// 3. Shuffle.
	shuffleLibraryPerCard(gs, seat)

	// 4. Draw 7.
	drawn := 0
	for i := 0; i < 7; i++ {
		if c := drawOne(gs, seat, src); c != nil {
			drawn++
		}
	}
	gs.LogEvent(gameengine.Event{
		Kind:   "shuffle_into_library",
		Seat:   seat,
		Source: src,
		Details: map[string]interface{}{
			"moved": moved,
			"drawn": drawn,
		},
	})
}

