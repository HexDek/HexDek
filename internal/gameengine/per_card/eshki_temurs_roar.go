package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEshkiTemursRoar wires Eshki, Temur's Roar.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you cast a creature spell, put a +1/+1 counter on Eshki.
//	  If that spell's power is 4 or greater, draw a card. If that spell's
//	  power is 6 or greater, Eshki deals damage equal to Eshki's power
//	  to each opponent.
//
// Implementation:
//   - "creature_spell_cast": gate on caster_seat == perm.Controller.
//     Always +1/+1 counter on Eshki. If the cast card's BasePower >= 4,
//     draw one. If >= 6, deal Eshki.Power() damage to each opponent.
func registerEshkiTemursRoar(r *Registry) {
	r.OnTrigger("Eshki, Temur's Roar", "creature_spell_cast", eshkiTemurCreatureSpellCast)
}

func eshkiTemurCreatureSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "eshki_temurs_roar_creature_cast"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	caster, _ := ctx["caster_seat"].(int)
	if caster != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}

	perm.AddCounter("+1/+1", 1)
	gs.InvalidateCharacteristicsCache()

	pw := card.BasePower
	drewCard := false
	dmgDealt := 0

	if pw >= 4 {
		if drawOne(gs, perm.Controller, perm.Card.DisplayName()) != nil {
			drewCard = true
		}
	}
	if pw >= 6 {
		dmg := perm.Power()
		if dmg > 0 {
			dmgDealt = dmg
			for _, opp := range gs.Opponents(perm.Controller) {
				s := gs.Seats[opp]
				if s == nil || s.Lost {
					continue
				}
				s.Life -= dmg
				gs.LogEvent(gameengine.Event{
					Kind:   "damage",
					Seat:   perm.Controller,
					Target: opp,
					Source: perm.Card.DisplayName(),
					Amount: dmg,
					Details: map[string]interface{}{
						"slug":   slug,
						"reason": "eshki_temurs_roar_six_power",
					},
				})
			}
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        perm.Controller,
		"spell_power": pw,
		"spell_card":  card.DisplayName(),
		"counter_add": 1,
		"drew_card":   drewCard,
		"damage_each": dmgDealt,
	})
	_ = gs.CheckEnd()
}
