package per_card

import (
	"sort"
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerBabaLysagaNightWitch wires Baba Lysaga, Night Witch.
//
// Oracle text:
//
//	{T}, Sacrifice up to three permanents: If there were three or more
//	card types among the sacrificed permanents, each opponent loses 3
//	life, you gain 3 life, and you draw three cards.
//
// Implementation:
//   - The {T} cost is the activation pipeline's responsibility.
//   - ctx["sacrifice_perms"] (a []*Permanent) supplies up to three
//     permanents Baba's controller chose to sacrifice. We sacrifice
//     them via SacrificePermanent (truncating to 3), then count the
//     distinct card types across them. The payoff fires only when the
//     count is >= 3 — the prior auto-generated stub always fired,
//     which gave Baba a free always-on payoff.
//   - When ctx is nil or supplies no permanents, the activation is
//     legal under the "up to three" wording but the type-count is
//     necessarily 0, so no payoff fires (we still tap & log).
func registerBabaLysagaNightWitch(r *Registry) {
	r.OnActivated("Baba Lysaga, Night Witch", babaLysagaNightWitchActivate)
}

func babaLysagaNightWitchActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "baba_lysaga_night_witch_activate"
	if gs == nil || src == nil {
		return
	}
	if abilityIdx != 0 {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}

	var picked []*gameengine.Permanent
	if ctx != nil {
		if sacs, ok := ctx["sacrifice_perms"].([]*gameengine.Permanent); ok {
			picked = sacs
		}
	}
	if len(picked) > 3 {
		picked = picked[:3]
	}
	typeSet := map[string]bool{}
	sacNames := make([]string, 0, len(picked))
	for _, p := range picked {
		if p == nil || p.Card == nil {
			continue
		}
		if p.Controller != seat {
			continue
		}
		for _, t := range p.Card.Types {
			lt := strings.ToLower(strings.TrimSpace(t))
			switch lt {
			case "creature", "artifact", "enchantment", "land",
				"planeswalker", "battle", "tribal":
				typeSet[lt] = true
			}
		}
		sacNames = append(sacNames, p.Card.DisplayName())
		gameengine.SacrificePermanent(gs, p, "baba_lysaga_cost")
	}

	distinctTypes := len(typeSet)
	gotPayoff := distinctTypes >= 3
	if gotPayoff {
		for i := 0; i < 3; i++ {
			drawOne(gs, seat, src.Card.DisplayName())
		}
		gameengine.GainLife(gs, seat, 3, src.Card.DisplayName())
		for _, opp := range gs.Opponents(seat) {
			if gs.Seats[opp] != nil && !gs.Seats[opp].Lost {
				gameengine.LoseLife(gs, opp, 3, src.Card.DisplayName())
			}
		}
		_ = gs.CheckEnd()
	}
	// Sorted for stable test snapshots.
	typeList := make([]string, 0, len(typeSet))
	for t := range typeSet {
		typeList = append(typeList, t)
	}
	sort.Strings(typeList)
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           seat,
		"sacrificed":     sacNames,
		"distinct_types": distinctTypes,
		"types":          typeList,
		"payoff":         gotPayoff,
	})
}
