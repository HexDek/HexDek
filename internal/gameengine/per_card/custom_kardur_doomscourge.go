package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKardurDoomscourgeCustom adds Kardur's "attacking creature
// dies → drain each opponent" trigger that the auto-gen ETB stub
// gained-life-on-ETB-only handler omits.
//
// Oracle text:
//
//	When Kardur enters, until your next turn, creatures your opponents
//	control attack each combat if able and attack a player other than
//	you if able.
//	Whenever an attacking creature dies, each opponent loses 1 life
//	and you gain 1 life.
//
// The "goad-everything" ETB rider is engine-side (combat-mandatory
// flag); we surface it with emitPartial. The death drain is wired
// fully — fires on `creature_dies` when the dying creature was
// attacking (we check perm.Flags["attacking"] which the engine sets
// during DealCombatDamageStep).
func registerKardurDoomscourgeCustom(r *Registry) {
	r.OnETB("Kardur, Doomscourge", kardurETBGoadFlag)
	r.OnTrigger("Kardur, Doomscourge", "creature_dies", kardurAttackerDeathDrain)
}

func kardurETBGoadFlag(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "kardur_etb_goad_all"
	if gs == nil || perm == nil {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	// Set a flag the engine combat layer can branch on. +1 so default
	// 0 means "off"; controller seat encoded.
	gs.Flags["kardur_goad_seat"] = perm.Controller + 1
	// "Until your next turn": queue a delayed trigger to clear the
	// goad-seat flag at the start of Kardur's controller's next upkeep.
	// Without this expiry the flag would persist all game.
	controller := perm.Controller
	gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
		TriggerAt:      "next_upkeep",
		ControllerSeat: controller,
		SourceCardName: perm.Card.DisplayName(),
		OneShot:        true,
		EffectFn: func(gs *gameengine.GameState) {
			if gs == nil || gs.Flags == nil {
				return
			}
			if gs.Flags["kardur_goad_seat"] == controller+1 {
				delete(gs.Flags, "kardur_goad_seat")
				gs.LogEvent(gameengine.Event{
					Kind:   "per_card_handler",
					Source: "Kardur, Doomscourge",
					Details: map[string]interface{}{
						"slug":   "kardur_goad_expires",
						"reason": "until_your_next_turn_elapsed",
						"seat":   controller,
					},
				})
			}
		},
	})
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": perm.Controller,
	})
	emitPartial(gs, "kardur_goad_until_next_turn", perm.Card.DisplayName(),
		"opponent attack-mandatory + redirect-to-other-player not yet enforced by combat layer; flag set with expiry")
}

func kardurAttackerDeathDrain(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "kardur_attacker_death_drain"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	dying, _ := ctx["perm"].(*gameengine.Permanent)
	if dying == nil {
		return
	}
	// Only attacking creatures count.
	if dying.Flags == nil || dying.Flags["attacking"] != 1 {
		return
	}
	gameengine.GainLife(gs, perm.Controller, 1, perm.Card.DisplayName())
	drained := 0
	for _, oppSeat := range gs.Opponents(perm.Controller) {
		if oppSeat < 0 || oppSeat >= len(gs.Seats) || gs.Seats[oppSeat] == nil || gs.Seats[oppSeat].Lost {
			continue
		}
		gameengine.LoseLife(gs, oppSeat, 1, perm.Card.DisplayName())
		drained++
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          perm.Controller,
		"opps_drained":  drained,
		"dying_creature": dying.Card.DisplayName(),
	})
}
