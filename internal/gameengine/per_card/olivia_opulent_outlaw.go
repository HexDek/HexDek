package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerOliviaOpulentOutlaw wires Olivia, Opulent Outlaw.
//
// Oracle text:
//
//	Flying, lifelink
//	Whenever one or more outlaws you control deal combat damage to a
//	player, create a Treasure token. (Assassins, Mercenaries, Pirates,
//	Rogues, and Warlocks are outlaws.)
//	{3}, Sacrifice two Treasures: Put two +1/+1 counters on each
//	creature you control. Activate only as a sorcery.
//
// We wire the outlaw-combat treasure trigger; the activated sac-two-
// treasures-for-anthem is left as a parser gap (multi-resource activation).
var oliviaOutlawTypes = []string{"assassin", "mercenary", "pirate", "rogue", "warlock"}

func registerOliviaOpulentOutlaw(r *Registry) {
	r.OnTrigger("Olivia, Opulent Outlaw", "combat_damage_player", oliviaOutlawDamage)
}

func oliviaOutlawDamage(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "olivia_outlaw_treasure"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	sourceSeat, _ := ctx["source_seat"].(int)
	sourceName, _ := ctx["source_card"].(string)
	if sourceSeat != perm.Controller {
		return
	}
	if !oliviaSourceIsOutlaw(gs, sourceSeat, sourceName) {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	turnKey := "olivia_treasure_combat"
	if perm.Flags[turnKey] == gs.Turn {
		return
	}
	perm.Flags[turnKey] = gs.Turn

	tok := gameengine.CreateCreatureToken(gs, perm.Controller, "Treasure",
		[]string{"artifact", "treasure"}, 0, 0)
	if tok != nil && tok.Card != nil {
		tok.Card.Types = []string{"token", "artifact", "treasure"}
		tok.Card.BasePower = 0
		tok.Card.BaseToughness = 0
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": sourceName,
	})
}

func oliviaSourceIsOutlaw(gs *gameengine.GameState, seatIdx int, name string) bool {
	if gs == nil || seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	s := gs.Seats[seatIdx]
	if s == nil {
		return false
	}
	for _, p := range s.Battlefield {
		if p == nil || p.Card == nil || p.Card.DisplayName() != name {
			continue
		}
		typeLine := strings.ToLower(p.Card.TypeLine)
		for _, t := range p.Card.Types {
			lc := strings.ToLower(t)
			for _, ot := range oliviaOutlawTypes {
				if lc == ot {
					return true
				}
			}
		}
		for _, ot := range oliviaOutlawTypes {
			if strings.Contains(typeLine, ot) {
				return true
			}
		}
		return false
	}
	return false
}
