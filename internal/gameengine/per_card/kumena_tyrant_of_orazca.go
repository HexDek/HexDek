package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerKumenaTyrantOfOrazca wires Kumena, Tyrant of Orazca.
//
// Oracle text:
//
//	Tap another untapped Merfolk you control: Kumena can't be blocked
//	this turn.
//	Tap three untapped Merfolk you control: Draw a card.
//	Tap five untapped Merfolk you control: Put a +1/+1 counter on each
//	Merfolk you control.
//
// Implementation: each ability requires tapping N untapped non-Kumena
// Merfolk as cost. We pick from controller's untapped merfolk, tap them,
// and apply the effect. Ability dispatch: 0 = unblockable, 1 = draw,
// 2 = +1/+1 counters.
func registerKumenaTyrantOfOrazca(r *Registry) {
	r.OnActivated("Kumena, Tyrant of Orazca", kumenaActivated)
}

func kumenaTapMerfolk(gs *gameengine.GameState, seat int, src *gameengine.Permanent, n int) []*gameengine.Permanent {
	if seat < 0 || seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return nil
	}
	tapped := []*gameengine.Permanent{}
	for _, p := range gs.Seats[seat].Battlefield {
		if len(tapped) >= n {
			break
		}
		if p == nil || p == src || p.Tapped || !p.IsCreature() || p.Card == nil {
			continue
		}
		if !cardHasType(p.Card, "merfolk") {
			continue
		}
		p.Tapped = true
		tapped = append(tapped, p)
	}
	if len(tapped) < n {
		// Roll back if we couldn't tap enough.
		for _, p := range tapped {
			p.Tapped = false
		}
		return nil
	}
	return tapped
}

func kumenaActivated(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "kumena_merfolk_activation"
	if gs == nil || src == nil {
		return
	}
	seat := src.Controller
	switch abilityIdx {
	case 0:
		t := kumenaTapMerfolk(gs, seat, src, 1)
		if t == nil {
			emitFail(gs, slug, src.Card.DisplayName(), "not_enough_merfolk", map[string]interface{}{"need": 1})
			return
		}
		if src.Flags == nil {
			src.Flags = map[string]int{}
		}
		src.Flags["unblockable_until_eot"] = 1
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"effect": "unblockable",
			"tapped": len(t),
		})
	case 1:
		t := kumenaTapMerfolk(gs, seat, src, 3)
		if t == nil {
			emitFail(gs, slug, src.Card.DisplayName(), "not_enough_merfolk", map[string]interface{}{"need": 3})
			return
		}
		drawOne(gs, seat, src.Card.DisplayName())
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"effect": "draw",
			"tapped": len(t),
		})
	case 2:
		t := kumenaTapMerfolk(gs, seat, src, 5)
		if t == nil {
			emitFail(gs, slug, src.Card.DisplayName(), "not_enough_merfolk", map[string]interface{}{"need": 5})
			return
		}
		count := 0
		for _, p := range gs.Seats[seat].Battlefield {
			if p == nil || !p.IsCreature() || p.Card == nil {
				continue
			}
			if !cardHasType(p.Card, "merfolk") {
				continue
			}
			p.AddCounter("+1/+1", 1)
			count++
		}
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":         seat,
			"effect":       "counters",
			"merfolk_hit":  count,
			"tapped":       len(t),
		})
	}
}
