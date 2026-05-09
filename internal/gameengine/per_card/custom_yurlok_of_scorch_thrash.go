package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerYurlokOfScorchThrashCustom implements Yurlok's group-Mana
// activation. The unspent-mana → life-loss static is a replacement
// effect that's engine territory; we record a flag and emit a partial.
//
// Oracle text:
//
//	Vigilance
//	A player losing unspent mana causes that player to lose that much
//	life.
//	{1}, {T}: Each player adds {B}{R}{G}.
//
// {1} + {T} are engine cost dispatch. The "each player" effect is wired
// here. Yurlok's pain-mana static is a replacement on the empty-mana-
// pool transition; recorded as a flag for the cost system.
func registerYurlokOfScorchThrashCustom(r *Registry) {
	r.OnActivated("Yurlok of Scorch Thrash", yurlokGroupMana)
	r.OnETB("Yurlok of Scorch Thrash", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["yurlok_pain_mana_active"] = 1
		emitPartial(gs, "yurlok_pain_mana_static", perm.Card.DisplayName(),
			"unspent-mana-causes-life-loss replacement needs cost-system mana-pool drain hook")
	})
}

func yurlokGroupMana(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "yurlok_group_mana"
	if gs == nil || src == nil {
		return
	}
	added := 0
	for i, s := range gs.Seats {
		if s == nil || s.Lost {
			continue
		}
		s.ManaPool += 3
		gs.LogEvent(gameengine.Event{
			Kind:   "add_mana",
			Seat:   i,
			Source: src.Card.DisplayName(),
			Amount: 3,
			Details: map[string]interface{}{
				"colors": "BRG",
				"reason": "yurlok_group_ritual",
			},
		})
		added++
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":          src.Controller,
		"players_added": added,
	})
}
