package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerVraskaTheSilencer wires Vraska, the Silencer.
//
// Oracle text:
//
//	Deathtouch
//	Whenever a nontoken creature an opponent controls dies, you may pay
//	{1}. If you do, return that card to the battlefield tapped under your
//	control. It's a Treasure artifact with "{T}, Sacrifice this artifact:
//	Add one mana of any color," and it loses all other card types.
//
// Implementation:
//   - "creature_dies": when a nontoken opp creature dies, mint a Treasure
//     under our control as a stand-in for "the dying card returns as a
//     Treasure". The actual reanimation-as-Treasure transformation
//     requires moving the original card from the graveyard and stripping
//     all card types — emitPartial flags that gap; we approximate by
//     creating a fresh Treasure to capture the value gain.
//   - AI policy: always pay {1} (pure value).
func registerVraskaTheSilencer(r *Registry) {
	r.OnTrigger("Vraska, the Silencer", "creature_dies", vraskaTheSilencerDies)
}

func vraskaTheSilencerDies(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "vraska_the_silencer_treasure_steal"
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
	if cardHasType(dyingCard, "token") {
		return
	}
	gameengine.CreateTreasureToken(gs, perm.Controller)
	emitPartial(gs, slug, perm.Card.DisplayName(), "reanimate_dying_card_as_treasure_approximated_with_fresh_treasure")
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"dying": dyingCard.DisplayName(),
	})
}
