package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGiselaTheBrokenBlade wires Gisela, the Broken Blade.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	Flying, first strike, lifelink
//	At the beginning of your end step, if you both own and control
//	Gisela and a creature named Bruna, the Fading Light, exile them,
//	then meld them into Brisela, Voice of Nightmares.
//
// Implementation (Muninn gap #32 — 31K hits):
//   - Keywords handled by AST keyword pipeline.
//   - OnTrigger("end_step"): gated on the controller being active seat,
//     plus owner-AND-control invariants for both Gisela and Bruna.
//     Wires gameengine.Meld from keywords_batch5.go (mishra_claimed_by_gix
//     punted on this).
func registerGiselaTheBrokenBlade(r *Registry) {
	r.OnTrigger("Gisela, the Broken Blade", "end_step", giselaBrokenBladeEndStep)
}

func giselaBrokenBladeEndStep(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "gisela_broken_blade_meld_brisela"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	activeSeat, _ := ctx["active_seat"].(int)
	if activeSeat != perm.Controller {
		return
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return
	}
	seat := gs.Seats[seatIdx]
	if seat == nil || seat.Lost {
		return
	}
	if perm.Owner != perm.Controller {
		emitFail(gs, slug, perm.Card.DisplayName(), "gisela_owner_controller_mismatch", nil)
		return
	}
	var bruna *gameengine.Permanent
	for _, p := range seat.Battlefield {
		if p == nil || p == perm || p.Card == nil {
			continue
		}
		if !strings.EqualFold(p.Card.DisplayName(), "Bruna, the Fading Light") {
			continue
		}
		if p.Owner != seatIdx {
			continue
		}
		bruna = p
		break
	}
	if bruna == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_bruna_on_battlefield", map[string]interface{}{
			"seat": seatIdx,
		})
		return
	}
	melded := gameengine.Meld(gs, perm, bruna)
	if melded == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "meld_failed", nil)
		return
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   seatIdx,
		"with":   bruna.Card.DisplayName(),
		"melded": "Brisela, Voice of Nightmares",
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"brisela_combined_stats_use_summed_power_toughness_not_canonical_9_10")
}
