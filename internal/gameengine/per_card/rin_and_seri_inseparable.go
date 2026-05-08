package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerRinAndSeri wires Rin and Seri, Inseparable.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	Whenever you cast a Dog spell, create a 1/1 green Cat creature
//	token.
//	Whenever you cast a Cat spell, create a 1/1 white Dog creature
//	token.
//	{R}{G}{W}, {T}: Rin and Seri deals damage to any target equal to
//	the number of Dogs you control. You gain life equal to the number
//	of Cats you control.
func registerRinAndSeri(r *Registry) {
	r.OnTrigger("Rin and Seri, Inseparable", "spell_cast", rinAndSeriSpellCast)
	r.OnActivated("Rin and Seri, Inseparable", rinAndSeriActivate)
}

func rinAndSeriSpellCast(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "rin_and_seri_dog_cat_token"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	casterSeat, _ := ctx["caster_seat"].(int)
	if casterSeat != perm.Controller {
		return
	}
	card, _ := ctx["card"].(*gameengine.Card)
	if card == nil {
		return
	}
	isDog := false
	isCat := false
	for _, t := range card.Types {
		switch strings.ToLower(t) {
		case "dog":
			isDog = true
		case "cat":
			isCat = true
		}
	}
	if !isDog && !isCat {
		return
	}
	if isDog {
		gameengine.CreateCreatureToken(gs, perm.Controller, "Cat",
			[]string{"creature", "cat", "pip:G"}, 1, 1)
	}
	if isCat {
		gameengine.CreateCreatureToken(gs, perm.Controller, "Dog",
			[]string{"creature", "dog", "pip:W"}, 1, 1)
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":     perm.Controller,
		"is_dog":   isDog,
		"is_cat":   isCat,
		"spell":    card.DisplayName(),
	})
}

func rinAndSeriActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "rin_and_seri_activated_dog_damage"
	if gs == nil || src == nil {
		return
	}
	seat := gs.Seats[src.Controller]
	if seat == nil {
		return
	}
	dogs := 0
	cats := 0
	for _, p := range seat.Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		for _, t := range p.Card.Types {
			switch strings.ToLower(t) {
			case "dog":
				dogs++
			case "cat":
				cats++
			}
		}
	}
	src.Tapped = true
	// Damage to lowest-life opponent (best heuristic target).
	tgt := -1
	bestLife := 1 << 30
	for _, oppIdx := range gs.Opponents(src.Controller) {
		s := gs.Seats[oppIdx]
		if s == nil || s.Lost {
			continue
		}
		if s.Life < bestLife {
			bestLife = s.Life
			tgt = oppIdx
		}
	}
	if tgt >= 0 && dogs > 0 {
		gameengine.DealDamage(gs, tgt, dogs, src.Card.DisplayName())
		_ = gs.CheckEnd()
	}
	if cats > 0 {
		gameengine.GainLife(gs, src.Controller, cats, src.Card.DisplayName())
	}
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":   src.Controller,
		"dogs":   dogs,
		"cats":   cats,
		"target": tgt,
	})
}
