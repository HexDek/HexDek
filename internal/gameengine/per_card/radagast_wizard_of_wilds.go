package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRadagastWizardOfWilds wires Radagast, Wizard of Wilds.
//
// Oracle text:
//
//	Ward {1}
//	Beasts and Birds you control have ward {1}.
//	Whenever you cast a spell with mana value 5 or greater, choose
//	one —
//	  • Create a 3/3 green Beast creature token.
//	  • Create a 2/2 blue Bird creature token with flying.
//
// AI policy: pick the Bird (2/2 flying, blue) when any opponent
// controls a creature with toughness >= 3 and we have no flyers yet —
// flying provides evasion that ground bodies don't. Otherwise pick the
// Beast (3/3 green) since the larger body is generally stronger and
// triggers more "creature ETB" effects.
func registerRadagastWizardOfWilds(r *Registry) {
	r.OnTrigger("Radagast, Wizard of Wilds", "spell_cast", radagastSpellCast)
}

func radagastSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "radagast_big_spell_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	if cardCMC(card) < 5 {
		return
	}

	pickBird := radagastShouldPickBird(gs, perm.Controller)
	if pickBird {
		// 2/2 blue Bird with flying. Token keywords are encoded as "kw:*"
		// entries in Types so the keyword pipeline picks them up.
		token := &gameengine.Card{
			Name:          "Bird Token",
			Owner:         perm.Controller,
			BasePower:     2,
			BaseToughness: 2,
			Types:         []string{"token", "creature", "bird", "kw:flying"},
			Colors:        []string{"U"},
			TypeLine:      "Token Creature — Bird",
		}
		enterBattlefieldWithETB(gs, perm.Controller, token, false)
	} else {
		tok := gameengine.CreateCreatureToken(gs, perm.Controller, "Beast",
			[]string{"creature", "beast"}, 3, 3)
		if tok != nil && tok.Card != nil {
			tok.Card.Colors = []string{"G"}
		}
	}

	mode := "beast_3_3"
	if pickBird {
		mode = "bird_2_2_flying"
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"spell": card.DisplayName(),
		"mode":  mode,
	})
}

// radagastShouldPickBird returns true when the controller would benefit
// more from an evasive flyer than a beefier ground body. Heuristic: any
// opponent controls a creature with toughness >= 3 AND the controller
// currently has no flying creatures of their own.
func radagastShouldPickBird(gs *gameengine.GameState, ctrl int) bool {
	if gs == nil {
		return false
	}
	hasOwnFlyer := false
	if ctrl >= 0 && ctrl < len(gs.Seats) && gs.Seats[ctrl] != nil {
		for _, p := range gs.Seats[ctrl].Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			if cardHasKeyword(p.Card, "flying") {
				hasOwnFlyer = true
				break
			}
		}
	}
	if hasOwnFlyer {
		return false
	}
	for i, s := range gs.Seats {
		if i == ctrl || s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || p.Card == nil || !p.IsCreature() {
				continue
			}
			if p.Toughness() >= 3 {
				return true
			}
		}
	}
	return false
}
