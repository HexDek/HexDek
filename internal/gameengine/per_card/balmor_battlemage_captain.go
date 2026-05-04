package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBalmorBattlemageCaptain wires Balmor, Battlemage Captain.
//
// Oracle text:
//
//	Flying
//	Whenever you cast an instant or sorcery spell, creatures you
//	control get +1/+0 and gain trample until end of turn.
func registerBalmorBattlemageCaptain(r *Registry) {
	r.OnTrigger("Balmor, Battlemage Captain", "instant_or_sorcery_cast", balmorTeamPump)
}

func balmorTeamPump(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "balmor_team_pump"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	ts := gs.NextTimestamp()
	count := 0
	var granted []*gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || !p.IsCreature() {
			continue
		}
		p.Modifications = append(p.Modifications, gameengine.Modification{
			Power:     1,
			Duration:  "until_end_of_turn",
			Timestamp: ts,
		})
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		// Only grant trample to creatures that don't already have it, so
		// we don't strip printed trample at end of turn.
		if p.Flags["kw:trample"] == 0 {
			p.Flags["kw:trample"] = 1
			granted = append(granted, p)
		}
		count++
	}
	if count > 0 {
		gs.InvalidateCharacteristicsCache()
	}
	if len(granted) > 0 {
		captured := granted
		gs.RegisterDelayedTrigger(&gameengine.DelayedTrigger{
			TriggerAt:      "next_end_step",
			ControllerSeat: perm.Controller,
			SourceCardName: perm.Card.DisplayName(),
			EffectFn: func(gs *gameengine.GameState) {
				for _, p := range captured {
					if p != nil && p.Flags != nil {
						delete(p.Flags, "kw:trample")
					}
				}
			},
		})
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"creatures": count,
	})
}
