package per_card

import (
	"fmt"
	"strconv"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTivashGloomSummoner wires Tivash, Gloom Summoner (Muninn parser-gap
// #102, 1.7K hits).
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	{4}{B}
//	Legendary Creature — Human Warlock
//	Lifelink
//	At the beginning of your end step, if you gained life this turn,
//	you may pay X life, where X is the amount of life you gained this
//	turn. If you do, create an X/X black Demon creature token with flying.
//
// Implementation (mirrors Markov Purifier / Witch of the Moors life-gain
// tracking):
//   - Lifelink handled by AST keyword pipeline.
//   - life_gained: tally per-(turn) gain into a per-perm flag.
//   - end_step gated on controller == active seat. AI policy: pay X if
//     life > X + 5 cushion AND X >= 2. An X/X flier for X life is a
//     near-break-even body trade; we only pay when leaving comfortably
//     above the lethal floor so we don't gift opponents lethal swings.
func registerTivashGloomSummoner(r *Registry) {
	r.OnTrigger("Tivash, Gloom Summoner", "life_gained", tivashGloomTrackLifeGain)
	r.OnTrigger("Tivash, Gloom Summoner", "end_step", tivashGloomEndStep)
}

func tivashGloomGainKey(turn int) string {
	return "tivash_gloom_gain_t" + strconv.Itoa(turn+1)
}

func tivashGloomTrackLifeGain(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
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
	perm.Flags[tivashGloomGainKey(gs.Turn)] += amount
}

func tivashGloomEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "tivash_gloom_end_step_demon"
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
	key := tivashGloomGainKey(gs.Turn)
	gained := perm.Flags[key]
	delete(perm.Flags, key)
	tivashGloomPruneKeys(perm, gs.Turn)
	if gained <= 0 {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil || seat.Lost {
		return
	}
	// Conservative AI: only pay when we'd remain above 6 life AND X >= 2.
	if gained < 2 || seat.Life-gained < 6 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"life_gain": gained,
			"life":      seat.Life,
			"paid":      false,
			"reason":    "ai_safety_threshold",
		})
		return
	}
	gameengine.DealDamage(gs, perm.Controller, gained, perm.Card.DisplayName())

	token := &gameengine.Card{
		Name:          fmt.Sprintf("Demon Token (%d/%d)", gained, gained),
		Owner:         perm.Controller,
		BasePower:     gained,
		BaseToughness: gained,
		Types:         []string{"token", "creature", "demon"},
		Colors:        []string{"B"},
		TypeLine:      "Token Creature — Demon",
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":      perm.Controller,
		"life_gain": gained,
		"paid":      true,
		"token":     token.Name,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"demon_token_flying_keyword_attach_pending_keyword_grant_pipeline")
	_ = gs.CheckEnd()
}

func tivashGloomPruneKeys(perm *gameengine.Permanent, currentTurn int) {
	if perm == nil || perm.Flags == nil {
		return
	}
	prefix := "tivash_gloom_gain_t"
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
