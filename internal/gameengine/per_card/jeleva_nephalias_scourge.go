package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerJelevaNephaliasScourge wires Jeleva, Nephalia's Scourge.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Flying
//	When Jeleva enters, each player exiles the top X cards of their
//	  library, where X is the amount of mana spent to cast Jeleva.
//	Whenever Jeleva attacks, you may cast an instant or sorcery spell
//	  from among cards exiled with Jeleva without paying its mana cost.
//
// Implementation:
//   - OnETB: X equals Jeleva's CMC by default (mana spent to cast is not
//     tracked granularly). For each player, move top-X cards from their
//     library to exile. Tag each with a flag in Card.Types so the attack
//     trigger can find them.
//   - "creature_attacks": find an instant or sorcery in our exile that
//     was exiled by Jeleva, cast it for free. Engine doesn't expose a
//     full free-cast pipeline for arbitrary cards from exile, so we move
//     the chosen card from exile to graveyard and emitPartial.
func registerJelevaNephaliasScourge(r *Registry) {
	r.OnETB("Jeleva, Nephalia's Scourge", jelevaETB)
	r.OnTrigger("Jeleva, Nephalia's Scourge", "creature_attacks", jelevaAttack)
}

const jelevaExileTag = "jeleva_exiled"

func jelevaETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "jeleva_etb_exile"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	x := gameengine.ManaCostOf(perm.Card)
	if x <= 0 {
		x = 4 // commander tax + base CMC fallback
	}
	exiledTotal := 0
	for i := range gs.Seats {
		s := gs.Seats[i]
		if s == nil || s.Lost {
			continue
		}
		for j := 0; j < x && len(s.Library) > 0; j++ {
			c := s.Library[0]
			c.Types = append(c.Types, jelevaExileTag)
			gameengine.MoveCard(gs, c, i, "library", "exile", "jeleva_etb")
			exiledTotal++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Controller,
		"x":            x,
		"exiled_total": exiledTotal,
	})
}

func jelevaAttack(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "jeleva_attack_freecast"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	for _, c := range seat.Exile {
		if c == nil {
			continue
		}
		hasTag := false
		for _, t := range c.Types {
			if t == jelevaExileTag {
				hasTag = true
				break
			}
		}
		if !hasTag {
			continue
		}
		if !cardHasType(c, "instant") && !cardHasType(c, "sorcery") {
			continue
		}
		gameengine.MoveCard(gs, c, c.Owner, "exile", "graveyard", "jeleva_freecast")
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":   perm.Controller,
			"chosen": c.DisplayName(),
		})
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"freecast_resolution_not_routed_through_stack")
		return
	}
}
