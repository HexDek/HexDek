package per_card

import (
	"strconv"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerSkullbriarTheWalkingGrave wires Skullbriar, the Walking Grave.
//
// Oracle text:
//
//	Haste
//	Whenever Skullbriar deals combat damage to a player, put a +1/+1
//	counter on it.
//	Counters remain on Skullbriar as it moves to any zone other than a
//	player's hand or library.
//
// Implementation:
//   - Haste is AST keyword territory.
//   - Combat damage trigger: on "combat_damage_player" with source_card
//     == this Skullbriar's display name and source_seat == this
//     controller, add a +1/+1 counter.
//   - Counter persistence: on "permanent_ltb" for Skullbriar to a zone
//     other than hand or library, stash the +1/+1 (and -1/-1) counter
//     count on a per-game flag keyed by Skullbriar's owner. On ETB,
//     restore those counts onto the new Permanent. Skullbriar is a
//     legendary commander, so one-per-owner is a safe key. The stash
//     lives on gs.Flags because the Card struct has no Flags field.
func registerSkullbriarTheWalkingGrave(r *Registry) {
	r.OnETB("Skullbriar, the Walking Grave", skullbriarOnETB)
	r.OnTrigger("Skullbriar, the Walking Grave", "combat_damage_player", skullbriarOnCombatDamage)
	r.OnTrigger("Skullbriar, the Walking Grave", "permanent_ltb", skullbriarOnLeaveBattlefield)
}

func skullbriarPlusKey(owner int) string {
	return "skullbriar_plus1_seat_" + strconv.Itoa(owner)
}

func skullbriarMinusKey(owner int) string {
	return "skullbriar_minus1_seat_" + strconv.Itoa(owner)
}

func skullbriarOnETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "skullbriar_etb_restore"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	if gs.Flags == nil {
		return
	}
	owner := perm.Card.Owner
	plus := gs.Flags[skullbriarPlusKey(owner)]
	minus := gs.Flags[skullbriarMinusKey(owner)]
	if plus > 0 {
		perm.AddCounter("+1/+1", plus)
	}
	if minus > 0 {
		perm.AddCounter("-1/-1", minus)
	}
	if plus > 0 || minus > 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"plus1plus": plus,
			"minus1":    minus,
		})
	}
	// The stash is consumed on restore.
	delete(gs.Flags, skullbriarPlusKey(owner))
	delete(gs.Flags, skullbriarMinusKey(owner))
}

func skullbriarOnCombatDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "skullbriar_combat_damage_counter"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	srcSeat, _ := ctx["source_seat"].(int)
	if srcSeat != perm.Controller {
		return
	}
	srcName, _ := ctx["source_card"].(string)
	if srcName != perm.Card.DisplayName() {
		return
	}
	perm.AddCounter("+1/+1", 1)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"new_count": perm.Counters["+1/+1"],
	})
}

func skullbriarOnLeaveBattlefield(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "skullbriar_persist_counters"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	leaving, _ := ctx["perm"].(*gameengine.Permanent)
	if leaving != perm {
		return
	}
	toZone, _ := ctx["to_zone"].(string)
	// Counters DO NOT persist when moving to hand or library.
	if toZone == "hand" || toZone == "library" {
		return
	}
	if perm.Counters == nil {
		return
	}
	plus := perm.Counters["+1/+1"]
	minus := perm.Counters["-1/-1"]
	if plus == 0 && minus == 0 {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	owner := perm.Card.Owner
	gs.Flags[skullbriarPlusKey(owner)] = plus
	gs.Flags[skullbriarMinusKey(owner)] = minus
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"to_zone":   toZone,
		"plus1plus": plus,
		"minus1":    minus,
	})
}
