package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSmirkingSpelljacker wires Smirking Spelljacker
// (Muninn parser-gap #28, 38,908 hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{4}{U}
//	Creature — Djinn Wizard Rogue
//	Flash
//	Flying
//	When this creature enters, exile target spell an opponent controls.
//	Whenever this creature attacks, if a card is exiled with it, you
//	may cast the exiled card without paying its mana cost.
//
// Implementation:
//   - OnETB: find the topmost opponent-controlled spell on the stack via
//     counterspells.go::findCounterableSpell, mark Countered AND set
//     CostMeta["exile_on_resolve"] = true so the engine's resolver
//     routes the countered card to exile rather than graveyard (Force-
//     of-Negation precedent). Tag the card's ExiledByTimestamp so the
//     attack trigger can find it.
//   - OnTrigger("creature_attacks"): search every exile pile for a card
//     whose ExiledByTimestamp matches our Timestamp. The cast-for-free
//     resolution path for arbitrary spells doesn't exist in the engine;
//     emitPartial documents the gap, but the linked card is logged so
//     tooling can see the trigger fired and identify the candidate.
func registerSmirkingSpelljacker(r *Registry) {
	r.OnETB("Smirking Spelljacker", smirkingSpelljackerETB)
	r.OnTrigger("Smirking Spelljacker", "creature_attacks", smirkingSpelljackerAttack)
}

func smirkingSpelljackerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "smirking_spelljacker_etb_exile"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	target := findCounterableSpell(gs, perm.Controller, nil)
	if target == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_opponent_spell_on_stack", nil)
		return
	}
	target.Countered = true
	if target.Card != nil {
		if target.CostMeta == nil {
			target.CostMeta = map[string]interface{}{}
		}
		target.CostMeta["exile_on_resolve"] = true
		target.CostMeta["exiled_by_timestamp"] = perm.Timestamp
		// Eagerly tag the linked card so attack-time lookup works even
		// if the engine's resolver doesn't propagate CostMeta into the
		// exiled card's ExiledByTimestamp.
		target.Card.ExiledByTimestamp = perm.Timestamp
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"exiled":  target.Card.DisplayName(),
		"opp":     target.Controller,
		"link_ts": perm.Timestamp,
	})
}

func smirkingSpelljackerAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "smirking_spelljacker_attack_cast_free"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	var linked *gameengine.Card
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, c := range s.Exile {
			if c == nil {
				continue
			}
			if c.ExiledByTimestamp == perm.Timestamp {
				linked = c
				break
			}
		}
		if linked != nil {
			break
		}
	}
	if linked == nil {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": perm.Controller,
			"cast": false,
		})
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"candidate": linked.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"cast_exiled_card_without_paying_mana_no_engine_free_cast_path_for_arbitrary_spell")
}
