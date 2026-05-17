package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerAbzanFalconer wires Abzan Falconer (Muninn parser-gap rank ~134,
// counters-matter +1/+1 anthem-grant).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{2}{W}
//	Creature — Human Soldier
//	Outlast {W} ({W}, {T}: Put a +1/+1 counter on this creature. Outlast
//	only as a sorcery.)
//	Each creature you control with a +1/+1 counter on it has flying.
//
// Implementation:
//   - Static "creatures with +1/+1 counters have flying" is a CR §613
//     continuous effect — the engine doesn't yet evaluate per-card
//     static grants from arbitrary battlefield permanents. Closest
//     existing pattern is Modular's keyword grant (post-death) which
//     stamps Flags["kw:flying"]. We mirror that by sweeping on every
//     counter_placed event and stamping kw:flying on each creature the
//     controller owns that currently has a +1/+1 counter. We also sweep
//     on permanent_etb so freshly-entered creatures pick up the grant.
//   - Outlast activation (cost {W}, tap, sorcery-speed) is the same
//     activated-ability cost-gating gap noted across the outlast family
//     — emitPartial.
//   - Stripping the grant when Abzan Falconer leaves the battlefield
//     uses permanent_ltb on Abzan itself: clear kw:flying flags we set.
//     We can't reliably distinguish flags we set from a native AST
//     flying; safe approach is to leave already-flying creatures alone
//     (skip the grant when they already have flying via AST/HasKeyword).
func registerAbzanFalconer(r *Registry) {
	r.OnETB("Abzan Falconer", abzanFalconerETB)
	r.OnTrigger("Abzan Falconer", "counter_placed", abzanFalconerCounterPlaced)
	r.OnTrigger("Abzan Falconer", "permanent_etb", abzanFalconerPermETB)
	r.OnTrigger("Abzan Falconer", "permanent_ltb", abzanFalconerOwnLTB)
}

func abzanFalconerETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	abzanFalconerGrantSweep(gs, perm)
}

func abzanFalconerCounterPlaced(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	abzanFalconerGrantSweep(gs, perm)
}

func abzanFalconerPermETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil {
		return
	}
	abzanFalconerGrantSweep(gs, perm)
}

func abzanFalconerOwnLTB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	leaving, _ := ctx["perm"].(*gameengine.Permanent)
	if leaving != perm {
		return
	}
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	stripped := 0
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Flags == nil {
			continue
		}
		if p.Flags["abzan_falconer_grant"] == 1 {
			delete(p.Flags, "kw:flying")
			delete(p.Flags, "abzan_falconer_grant")
			stripped++
		}
	}
	gs.InvalidateCharacteristicsCache()
	emit(gs, "abzan_falconer_grant_strip", perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"stripped": stripped,
	})
}

func abzanFalconerGrantSweep(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "abzan_falconer_flying_grant"
	seat := gs.Seats[perm.Controller]
	if seat == nil {
		return
	}
	granted := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil || !p.IsCreature() {
			continue
		}
		if p.Counters == nil || p.Counters["+1/+1"] <= 0 {
			continue
		}
		if p.HasKeyword("flying") {
			continue
		}
		if p.Flags == nil {
			p.Flags = map[string]int{}
		}
		if p.Flags["kw:flying"] == 1 {
			continue
		}
		p.Flags["kw:flying"] = 1
		p.Flags["abzan_falconer_grant"] = 1
		granted++
	}
	if granted > 0 {
		gs.InvalidateCharacteristicsCache()
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":    perm.Controller,
			"granted": granted,
		})
	}
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"outlast_activation_cost_w_tap_sorcery_speed_unwired_pending_activated_cost_gate")
}
