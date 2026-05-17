package per_card

import (
	"strconv"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerMarkovPurifier wires Markov Purifier (Muninn parser-gap #88, 5.7K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{1}{W}{B}
//	Creature — Vampire Cleric
//	Lifelink
//	At the beginning of your end step, if you gained life this turn,
//	you may pay {2}. If you do, draw a card.
//
// Implementation (mirrors Bre of Clan Stoutarm life-gain tracking):
//   - Lifelink handled by AST keyword pipeline.
//   - life_gained: tally per-turn life into a per-perm flag (each
//     Markov tracks its own counter so multiple copies are independent).
//   - end_step gated on controller == active seat. AI policy: always
//     pay {2} when affordable — draw-a-card for {2} at instant-speed-ish
//     pacing is essentially always worth it. Mana spend is bookkeeping
//     only (engine has no end-step mana priority window today).
func registerMarkovPurifier(r *Registry) {
	r.OnTrigger("Markov Purifier", "life_gained", markovPurifierTrackLifeGain)
	r.OnTrigger("Markov Purifier", "end_step", markovPurifierEndStep)
}

func markovPurifierGainKey(turn int) string {
	return "markov_purifier_gain_t" + strconv.Itoa(turn+1)
}

func markovPurifierTrackLifeGain(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	seat, _ := ctx["seat"].(int)
	if seat != perm.Controller {
		return
	}
	amount, _ := ctx["amount"].(int)
	if amount <= 0 {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags[markovPurifierGainKey(gs.Turn)] += amount
}

func markovPurifierEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "markov_purifier_end_step_draw"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	key := markovPurifierGainKey(gs.Turn)
	gained := perm.Flags[key]
	delete(perm.Flags, key)
	markovPurifierPruneKeys(perm, gs.Turn)
	if gained <= 0 {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	if seat.ManaPool < 2 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"life_gain": gained,
			"paid":      false,
			"reason":    "insufficient_mana",
		})
		return
	}
	seat.ManaPool -= 2
	gameengine.SyncManaAfterSpend(seat)
	drawOne(gs, perm.Controller, perm.Card.DisplayName())
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"life_gain": gained,
		"paid":      true,
	})
}

func markovPurifierPruneKeys(perm *gameengine.Permanent, currentTurn int) {
	if perm == nil || perm.Flags == nil {
		return
	}
	prefix := "markov_purifier_gain_t"
	cutoff := currentTurn + 1
	for k := range perm.Flags {
		if len(k) <= len(prefix) || k[:len(prefix)] != prefix {
			continue
		}
		n, err := strconv.Atoi(k[len(prefix):])
		if err != nil {
			continue
		}
		if n < cutoff {
			delete(perm.Flags, k)
		}
	}
}
