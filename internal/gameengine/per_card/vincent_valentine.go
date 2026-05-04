package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVincentValentine wires Vincent Valentine // Galian Beast.
//
// Oracle text (front face):
//
//	Whenever a creature an opponent controls dies, put a number of +1/+1
//	counters on Vincent Valentine equal to that creature's power.
//	Whenever Vincent Valentine attacks, you may transform it.
//
// Back face (Galian Beast):
//
//	Trample, lifelink
//	When Galian Beast dies, return it to the battlefield tapped (front
//	face up).
//
// Implementation:
//   - "creature_dies" trigger: if the dying creature was controlled by an
//     opponent of Vincent's controller, add power-many +1/+1 counters to
//     Vincent (clamped to >=0). The dying creature's power is taken from
//     ctx["dying_power"] when present; falls back to BasePower from the
//     card. Filter on perm == Vincent so we don't re-trigger on Vincent
//     himself dying.
//   - The transform-on-attack and Galian Beast die-return are not modeled
//     (DFC transform pipeline doesn't have a per-card hook here).
//     emitPartial.
func registerVincentValentine(r *Registry) {
	r.OnTrigger("Vincent Valentine // Galian Beast", "creature_dies", vincentValentineCreatureDies)
	r.OnTrigger("Vincent Valentine", "creature_dies", vincentValentineCreatureDies)
	r.OnTrigger("Vincent Valentine // Galian Beast", "creature_attacks", vincentValentineAttacks)
	r.OnTrigger("Vincent Valentine", "creature_attacks", vincentValentineAttacks)
}

func vincentValentineCreatureDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "vincent_valentine_dies_counters"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dyingCtrl, _ := ctx["controller_seat"].(int)
	if dyingCtrl == perm.Controller {
		return
	}
	dyingCard, _ := ctx["card"].(*gameengine.Card)
	if dyingCard == nil {
		return
	}
	if !cardHasType(dyingCard, "creature") {
		return
	}
	power := 0
	if v, ok := ctx["dying_power"].(int); ok {
		power = v
	} else {
		power = dyingCard.BasePower
	}
	if power <= 0 {
		return
	}
	perm.AddCounter("+1/+1", power)
	gs.InvalidateCharacteristicsCache()
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"counters": power,
		"dying":    dyingCard.DisplayName(),
	})
}

func vincentValentineAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "vincent_valentine_attack_transform"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk != perm {
		return
	}
	emitPartial(gs, slug, perm.Card.DisplayName(), "transform_to_galian_beast_not_implemented")
}
