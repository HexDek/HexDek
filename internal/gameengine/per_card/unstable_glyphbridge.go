package per_card

import (
	"fmt"
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerUnstableGlyphbridge wires Unstable Glyphbridge //
// Sandswirl Wanderglyph (modal DFC / craft transform).
//
// Oracle text (Scryfall, verified 2026-05-16):
//
//   Front face — Unstable Glyphbridge  {3}{W}{W}  Artifact
//     When this artifact enters, if you cast it, for each player,
//     choose a creature with power 2 or less that player controls.
//     Then destroy all creatures except creatures chosen this way.
//     Craft with artifact {3}{W}{W} ({3}{W}{W}, Exile this artifact,
//     Exile another artifact you control or an artifact card from your
//     graveyard: Return this card transformed under its owner's
//     control. Craft only as a sorcery.)
//
//   Back face — Sandswirl Wanderglyph  Artifact Creature — Golem
//     Flying
//     Whenever an opponent casts a spell during their turn, they can't
//     attack you or planeswalkers you control this turn.
//     Each opponent who attacked you or a planeswalker you control
//     this turn can't cast spells.
//
// Implementation (Muninn gap #37 — 25K hits):
//   - OnETB (front face): gated on perm.Flags["was_cast"]. For each
//     player, pick the highest-toughness creature of power ≤ 2 they
//     control (their best survivor — pessimises our wipe least for
//     our own board while still nuking opponents' big threats). Then
//     destroy every creature not in the survivor set via
//     gameengine.DestroyPermanent.
//   - Back-face static "can't attack you" / "can't cast spells"
//     effects require post-transform replacement-effect registration
//     and turn-scoped predicates that we don't expose to per-card
//     hooks yet. emitPartial. Craft transform cost itself is owned by
//     the AST cost pipeline.
func registerUnstableGlyphbridge(r *Registry) {
	r.OnETB("Unstable Glyphbridge", unstableGlyphbridgeETB)
	r.OnETB("Unstable Glyphbridge // Sandswirl Wanderglyph", unstableGlyphbridgeETB)
}

func unstableGlyphbridgeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "unstable_glyphbridge_etb_wipe"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if perm.Flags == nil || perm.Flags["was_cast"] != 1 {
		emitFail(gs, slug, perm.Card.DisplayName(), "not_cast", nil)
		emitPartial(gs, slug, perm.Card.DisplayName(),
			"back_face_static_cant_attack_cant_cast_unmodeled")
		return
	}
	survivors := map[*gameengine.Permanent]bool{}
	picks := map[int]string{}
	for seatIdx, s := range gs.Seats {
		if s == nil || s.Lost {
			continue
		}
		var pick *gameengine.Permanent
		bestT := -1
		for _, p := range s.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			if p.Power() > 2 {
				continue
			}
			t := p.Toughness()
			if t > bestT {
				bestT = t
				pick = p
			}
		}
		if pick != nil {
			survivors[pick] = true
			picks[seatIdx] = pick.Card.DisplayName()
		}
	}
	var victims []*gameengine.Permanent
	for _, s := range gs.Seats {
		if s == nil {
			continue
		}
		for _, p := range s.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			if survivors[p] {
				continue
			}
			victims = append(victims, p)
		}
	}
	destroyed := 0
	for _, v := range victims {
		if gameengine.DestroyPermanent(gs, v, nil) {
			destroyed++
		}
	}
	survivorNames := make([]string, 0, len(picks))
	for s, name := range picks {
		survivorNames = append(survivorNames, fmt.Sprintf("seat%d:%s", s, name))
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"destroyed": destroyed,
		"survivors": strings.Join(survivorNames, ","),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"back_face_sandswirl_static_effects_unmodeled_craft_transform_in_ast_cost_pipeline")
}
