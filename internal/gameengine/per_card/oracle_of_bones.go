package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOracleOfBones wires Oracle of Bones (Muninn parser-gap #94, 4.7K
// hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{2}{R}{R}
//	Creature — Minotaur Shaman
//	Haste
//	Tribute 2 (As this creature enters, an opponent of your choice may
//	put two +1/+1 counters on it.)
//	When this creature enters, if tribute wasn't paid, you may cast an
//	instant or sorcery spell from your hand without paying its mana
//	cost.
//
// Implementation:
//   - Haste handled by AST keyword pipeline.
//   - Tribute is a replacement-style "as it enters" choice belonging to
//     an opponent (CR §702.105). The engine doesn't yet model tribute
//     choice; we model an opponent-makes-the-rational-call shortcut:
//     the chosen opponent pays tribute iff the controller's hand
//     contains a juicy free-cast target (instant/sorcery whose CMC
//     is >= 4). Otherwise tribute is declined and we resolve the free
//     cast. This matches what most pilots actually do — giving 2 +1/+1
//     counters is usually preferable to letting a 4+ CMC bomb resolve
//     for free.
//   - When tribute isn't paid, drop the best free-cast spell from hand
//     onto the stack via gameengine.CastFromHandFree if available, else
//     fall back to a graceful partial emit.
func registerOracleOfBones(r *Registry) {
	r.OnETB("Oracle of Bones", oracleOfBonesETB)
}

func oracleOfBonesETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "oracle_of_bones_tribute_free_cast"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}
	// Scan hand for the highest-CMC instant or sorcery — the candidate
	// the controller would want to free-cast.
	var bestSpell *gameengine.Card
	bestSpellIdx := -1
	bestCMC := -1
	for i, c := range s.Hand {
		if c == nil {
			continue
		}
		if !cardHasType(c, "instant") && !cardHasType(c, "sorcery") {
			continue
		}
		cmc := cardCMC(c)
		if cmc > bestCMC {
			bestCMC = cmc
			bestSpell = c
			bestSpellIdx = i
		}
	}
	// Opponent's rational decision: pay tribute iff free-cast target is
	// CMC >= 4 (worth the +2 +1/+1 trade to deny the free bomb).
	tributePaid := bestSpell != nil && bestCMC >= 4
	if tributePaid {
		perm.AddCounter("+1/+1", 2)
		gs.InvalidateCharacteristicsCache()
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":           seat,
			"tribute_paid":   true,
			"deferred_spell": bestSpell.DisplayName(),
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"tribute_choice_modelled_as_rational_opponent_no_engine_pipeline")
		return
	}
	if bestSpell == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":         seat,
			"tribute_paid": false,
			"reason":       "no_instant_or_sorcery_in_hand",
		})
		return
	}
	// Tribute declined — free-cast the spell. The engine has no public
	// cast-from-hand-free helper; we shortcut to graveyard-resolution
	// for spells (their effects will be approximated by their own ETB
	// or resolve hooks if registered) and emitPartial.
	_ = bestSpellIdx
	gameengine.MoveCard(gs, bestSpell, seat, "hand", "graveyard", "oracle_of_bones_free_cast")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         seat,
		"tribute_paid": false,
		"free_cast":    bestSpell.DisplayName(),
		"cmc":          bestCMC,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"free_cast_routes_to_graveyard_resolution_pending_cast_from_hand_free_helper")
}
