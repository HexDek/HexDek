package main

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

// hasCastConditionalETB scans the AST for "When ~ enters, if you cast it, ..."
// triggered abilities (CR §603.6c). The 19 cards in this category — The One
// Ring, Tiamat, Eon Frolicker, Skitterbeam Battalion, Geological Appraiser,
// Cyclone Summoner, Breaching Leviathan, Transcendent Dragon, Light-Paws
// Emperor's Voice, Wild Pair, Acererak the Archlich, Wistfulness, Gruff
// Triplets, Kodama of the East Tree, Frodo Sauron's Bane, Acclaimed
// Contender, Oracle of Bones, Rankle and Torbran — never fire their ETB
// effect when Thor places them directly on the battlefield because they
// require the cast pipeline.
//
// Returns:
//   - cast: true iff the card has a cast-conditional ETB trigger.
//   - fromHand: true iff the conditional explicitly requires casting from hand
//     (Cyclone Summoner, Breaching Leviathan, Wild Pair).
//   - inverse: true iff the conditional fires when the card was NOT cast
//     (Preston the Vanisher) — caller should keep direct placement.
func hasCastConditionalETB(ast *gameast.CardAST) (cast, fromHand, inverse bool) {
	if ast == nil {
		return
	}
	for _, ab := range ast.Abilities {
		trig, ok := ab.(*gameast.Triggered)
		if !ok {
			continue
		}
		if !strings.EqualFold(trig.Trigger.Event, "etb") {
			continue
		}
		raw := castConditionalRawText(trig.Effect)
		if raw == "" {
			continue
		}
		rawL := strings.ToLower(raw)
		switch {
		case strings.Contains(rawL, "if it wasn't cast"),
			strings.Contains(rawL, "if it was not cast"),
			strings.Contains(rawL, "if you didn't cast"),
			strings.Contains(rawL, "if you did not cast"):
			inverse = true
		case strings.Contains(rawL, "if you cast it from your hand"),
			strings.Contains(rawL, "if you cast this from your hand"),
			strings.Contains(rawL, "if it was cast from your hand"):
			cast = true
			fromHand = true
		case strings.Contains(rawL, "if you cast it"),
			strings.Contains(rawL, "if you cast this"),
			strings.Contains(rawL, "if it was cast"):
			cast = true
		}
	}
	return
}

// removeCardFromHand removes the card from the hand slice (by pointer
// identity). Returns the slice with the first matching entry removed.
func removeCardFromHand(hand []*gameengine.Card, card *gameengine.Card) []*gameengine.Card {
	for i, c := range hand {
		if c == card {
			return append(hand[:i], hand[i+1:]...)
		}
	}
	return hand
}

// castConditionalRawText extracts the raw conditional text from a Triggered
// effect node when it's a ModificationEffect with kind "conditional_effect".
// Returns "" otherwise.
func castConditionalRawText(eff gameast.Effect) string {
	if eff == nil {
		return ""
	}
	if me, ok := eff.(*gameast.ModificationEffect); ok &&
		me.ModKind == "conditional_effect" && len(me.Args) > 0 {
		if s, ok := me.Args[0].(string); ok {
			return s
		}
	}
	return ""
}

// configureForCastPath rewires a goldilocks setup that placed the source on
// the battlefield directly into one where the source is in hand and ready
// to be cast. Returns the card pointer to cast plus true when the cast path
// should be used; otherwise returns nil, false (caller falls back to the
// normal direct-place flow).
//
// After this returns true, the caller MUST replace the ETB resolution step
// with gameengine.CastSpell(gs, 0, card, nil) — that path drains the stack,
// fires ETB triggers naturally, and stamps perm.Flags["was_cast"]=1 on the
// resolved permanent so "if you cast it" intervening-ifs evaluate true.
func configureForCastPath(gs *gameengine.GameState, oc *oracleCard,
	srcPerm *gameengine.Permanent, info *effectInfo, tr *Tracer) (*gameengine.Card, bool) {
	if info == nil || info.abilityKind != "triggered" || info.trigger == nil {
		return nil, false
	}
	if !strings.EqualFold(info.trigger.Event, "etb") {
		return nil, false
	}
	if oc == nil || oc.ast == nil {
		return nil, false
	}
	cast, fromHand, inverse := hasCastConditionalETB(oc.ast)
	if !cast || inverse {
		return nil, false
	}
	if srcPerm == nil || srcPerm.Card == nil || gs == nil || len(gs.Seats) == 0 {
		return nil, false
	}
	card := srcPerm.Card
	seat := gs.Seats[0]
	// Pull the source off the battlefield — the cast pipeline will put it
	// back via resolvePermanentSpellETB, which is the path that stamps
	// was_cast on the new permanent.
	for i, p := range seat.Battlefield {
		if p == srcPerm {
			seat.Battlefield = append(seat.Battlefield[:i], seat.Battlefield[i+1:]...)
			break
		}
	}
	seat.Hand = append(seat.Hand, card)
	// Provide enough mana of every color so any printed cost can be paid.
	for _, color := range []string{"W", "U", "B", "R", "G", "C", "any"} {
		gameengine.AddMana(gs, seat, color, 10, "thor_cast_setup")
	}
	if tr != nil {
		tr.Record("CAST_SETUP", "casting %q from hand (wasCast=true, fromHand=%v)",
			card.Name, fromHand)
	}
	return card, true
}
