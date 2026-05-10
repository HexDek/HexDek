package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMisterNegativeCustom implements Mister Negative's life-swap
// ETB. The auto-generated stub is a no-op.
//
// Oracle text:
//
//	Vigilance, lifelink
//	Darkforce Inversion — When Mister Negative enters, you may exchange
//	life totals with target opponent. If you lost life this way, draw
//	that many cards.
//
// Implementation notes:
//   - Vigilance / lifelink are static keywords handled by the AST.
//   - Target choice: exchange with the OPPONENT WITH HIGHEST LIFE if
//     they're above us (gain the upside). If our life >= every
//     opponent's, the swap is strictly negative — we skip the "may"
//     and just emit the no-swap event.
//   - Draw equals (newLife - oldLife) when we GAINED life from the
//     swap. Per oracle, draw triggers when we LOST life — so for
//     ours_before > theirs_before, we'd draw |delta|. We compute
//     this AFTER the swap.
//   - SetLife is not exported; we use Gain/LoseLife to walk to the
//     target value.
func registerMisterNegativeCustom(r *Registry) {
	r.OnETB("Mister Negative", misterNegativeETB)
}

func misterNegativeETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "mister_negative_life_swap"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	me := gs.Seats[seatIdx]
	if me == nil {
		return
	}

	// Pick the opponent with the most life — swapping with them gives
	// us upside. If no opponent has more life than us, the "may" is
	// declined.
	target := -1
	bestLife := me.Life
	for i, s := range gs.Seats {
		if i == seatIdx || s == nil || s.Lost || s.LeftGame {
			continue
		}
		if s.Life > bestLife {
			bestLife = s.Life
			target = i
		}
	}
	if target < 0 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat": seatIdx,
			"note": "no_swap_no_upside",
		})
		return
	}

	myOld := me.Life
	their := gs.Seats[target]
	theirOld := their.Life
	source := perm.Card.DisplayName()

	// Walk life to swapped values.
	if myOld < theirOld {
		gameengine.GainLife(gs, seatIdx, theirOld-myOld, source)
	} else if myOld > theirOld {
		gameengine.LoseLife(gs, seatIdx, myOld-theirOld, source)
	}
	if theirOld < myOld {
		gameengine.GainLife(gs, target, myOld-theirOld, source)
	} else if theirOld > myOld {
		gameengine.LoseLife(gs, target, theirOld-myOld, source)
	}

	// Draw cards equal to life lost (if we lost any).
	drew := 0
	if myOld > me.Life {
		drew = myOld - me.Life
		for i := 0; i < drew; i++ {
			if len(me.Library) == 0 {
				break
			}
			card := me.Library[0]
			gameengine.MoveCard(gs, card, seatIdx, "library", "hand", "mister_negative_draw")
		}
	}

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":        seatIdx,
		"target_seat": target,
		"my_old":      myOld,
		"my_new":      me.Life,
		"their_old":   theirOld,
		"their_new":   their.Life,
		"drew":        drew,
	})
	_ = gs.CheckEnd()
}
