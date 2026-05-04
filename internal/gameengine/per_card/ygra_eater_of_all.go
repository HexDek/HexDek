package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYgraEaterOfAll wires Ygra, Eater of All.
//
// Oracle text:
//
//	Ward—Sacrifice a Food.
//	Other creatures are Food artifacts in addition to their other types
//	and have "{2}, {T}, Sacrifice this permanent: You gain 3 life."
//	Whenever a Food is put into a graveyard from the battlefield, put two
//	+1/+1 counters on Ygra.
//
// Implementation:
//   - "permanent_ltb" trigger: if the leaving permanent's destination is
//     graveyard and it had Food OR was a creature (per Ygra's "other
//     creatures are Food" type-changing static), add 2 +1/+1 counters
//     to Ygra. The static "other creatures are Food" effect itself is
//     not implemented; emitPartial flags that.
//   - The activated 3-life sacrifice ability granted to other creatures
//     is not implemented.
func registerYgraEaterOfAll(r *Registry) {
	r.OnTrigger("Ygra, Eater of All", "permanent_ltb", ygraPermLeftBattlefield)
	r.OnTrigger("Ygra, Eater of All", "creature_dies", ygraCreatureDies)
}

func ygraPermLeftBattlefield(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ygra_food_ltb_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dest, _ := ctx["to_zone"].(string)
	if dest != "graveyard" {
		return
	}
	leaver, _ := ctx["card"].(*gameengine.Card)
	if leaver == nil {
		return
	}
	isFood := false
	for _, t := range leaver.Types {
		if t == "food" || t == "Food" {
			isFood = true
			break
		}
	}
	if !isFood {
		return
	}
	perm.AddCounter("+1/+1", 2)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": 2,
	})
}

func ygraCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "ygra_creature_dies_as_food"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(), "other_creatures_are_food_static_not_implemented")
}
