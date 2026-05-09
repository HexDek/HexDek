package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRendmawCreakingNest wires Rendmaw, Creaking Nest.
//
// Oracle text (Duskmourn Commander, {3}{B}{G}, 5/5):
//
//	Menace, reach
//	When Rendmaw enters and whenever you play a card with two or more
//	card types, each player creates a tapped 2/2 black Bird creature
//	token with flying. The tokens are goaded for the rest of the game.
//
// Implementation:
//   - ETB fires the each-player-bird-spawn payload.
//   - "spell_cast" trigger gated on caster_seat == perm.Controller and
//     the spell having two or more card types (creature, instant,
//     sorcery, artifact, enchantment, land, planeswalker, tribal,
//     battle, kindred). Lands aren't cast — the engine doesn't fire
//     spell_cast for played lands — so multi-type lands miss this
//     trigger; flagged via emitPartial.
//   - Each token is created tapped with the goaded flag set so the
//     engine's combat-phase scanner picks them up correctly.
func registerRendmawCreakingNest(r *Registry) {
	r.OnETB("Rendmaw, Creaking Nest", rendmawETB)
	r.OnTrigger("Rendmaw, Creaking Nest", "spell_cast", rendmawSpellCast)
}

func rendmawETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "rendmaw_etb_birds"
	if gs == nil || perm == nil {
		return
	}
	rendmawSpawnBirdsForEachPlayer(gs, perm, "etb")
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"played_land_with_2_plus_types_not_detected_no_land_play_event_for_per_card")
}

func rendmawSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
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
	if countDistinctCardTypes(card) < 2 {
		return
	}
	rendmawSpawnBirdsForEachPlayer(gs, perm, "multi_type_cast")
}

func rendmawSpawnBirdsForEachPlayer(gs *gameengine.GameState, perm *gameengine.Permanent, source string) {
	const slug = "rendmaw_each_player_bird_token"
	created := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost {
			continue
		}
		token := gameengine.CreateCreatureToken(gs, i, "Bird Token",
			[]string{"creature", "bird", "pip:B", "flying"}, 2, 2)
		if token != nil {
			token.Tapped = true
			if token.Flags == nil {
				token.Flags = map[string]int{}
			}
			token.Flags["goaded"] = 1
			created++
		}
		_ = i
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    perm.Controller,
		"source":  source,
		"created": created,
	})
}

func countDistinctCardTypes(card *gameengine.Card) int {
	if card == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, t := range card.Types {
		switch strings.ToLower(t) {
		case "creature", "instant", "sorcery", "artifact", "enchantment",
			"land", "planeswalker", "tribal", "battle", "kindred":
			seen[strings.ToLower(t)] = true
		}
	}
	return len(seen)
}
