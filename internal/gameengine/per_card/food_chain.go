package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerFoodChain wires up Food Chain.
//
// Oracle text:
//
//	Exile a creature you control: Add an amount of mana of any one
//	color equal to the exiled creature's mana value plus one. Spend
//	this mana only to cast creature spells.
//
// Infinite-mana enabler. Combined with a creature that returns itself
// from exile to hand (Misthollow Griffin / Eternal Scourge / Squee,
// the Immortal), this loops: exile Griffin (mana_value=4) -> get 5 mana
// -> cast Griffin for 4 -> exile again -> net +1 mana per loop, unbounded
// creature-only mana.
//
// Implementation:
//   - OnActivated: pick a target creature by ctx or highest-CMC; exile
//     it; add CMC+1 mana to the controller's pool.
//   - "Spend only on creatures" restriction: tracked via a
//     gs.Flags["food_chain_mana_seat_N"] counter. The engine's cost-
//     payment path can consult this flag to restrict the mana to creature
//     spells. The counter represents how much food-chain-restricted mana
//     is in the pool.
//   - OnETB: grant "cast from exile" zone-cast permissions for any
//     creature in exile that has a static "cast from exile" ability
//     (e.g. Misthollow Griffin, Eternal Scourge). This enables the
//     Food Chain infinite loop.
//
// Activation contract:
//
//	ctx["creature_perm"] *gameengine.Permanent  -- which creature to exile.
//	                                             When absent, pick the
//	                                             highest-CMC creature.
func registerFoodChain(r *Registry) {
	r.OnETB("Food Chain", foodChainETB)
	r.OnActivated("Food Chain", foodChainActivate)
}

func foodChainETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "food_chain_etb"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	// Check exile for creatures with static "cast from exile" ability
	// (Misthollow Griffin, Eternal Scourge, Squee the Immortal) and
	// grant them zone-cast permissions.
	s := gs.Seats[seat]
	granted := 0
	for _, c := range s.Exile {
		if c == nil || !cardHasType(c, "creature") {
			continue
		}
		if cardHasExileCastStatic(c) {
			perm := gameengine.NewExileCastPermission()
			perm.RequireController = seat
			perm.SourceName = "Food Chain"
			gameengine.RegisterZoneCastGrant(gs, c, perm)
			granted++
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":          seat,
		"exile_grants":  granted,
	})
}

func foodChainActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "food_chain_exile_for_mana"
	if gs == nil || src == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]

	// Pick the victim creature.
	var victim *gameengine.Permanent
	if v, ok := ctx["creature_perm"].(*gameengine.Permanent); ok && v != nil {
		victim = v
	} else {
		// Highest-CMC creature under controller.
		best := -1
		for _, p := range s.Battlefield {
			if p == nil || !p.IsCreature() {
				continue
			}
			cmc := cardCMC(p.Card)
			if cmc > best {
				best = cmc
				victim = p
			}
		}
	}
	if victim == nil {
		emitFail(gs, slug, src.Card.DisplayName(), "no_creature_to_exile", nil)
		return
	}

	cmc := cardCMC(victim.Card)
	manaGained := cmc + 1

	// Exile the creature via ExilePermanent for proper zone-change handling:
	// replacement effects, LTB triggers, commander redirect.
	exiledCard := victim.Card
	gameengine.ExilePermanent(gs, victim, src)

	// If the exiled creature has a static "cast from exile" ability,
	// grant it a zone-cast permission so it can be re-cast (enabling
	// the Food Chain loop).
	if cardHasExileCastStatic(exiledCard) {
		exilePerm := gameengine.NewExileCastPermission()
		exilePerm.RequireController = seat
		exilePerm.SourceName = "Food Chain"
		gameengine.RegisterZoneCastGrant(gs, exiledCard, exilePerm)
	}

	// Add mana. Track the creature-only restriction via a flag.
	s.ManaPool += manaGained
	gameengine.SyncManaAfterAdd(s, manaGained)

	// Track food-chain-restricted mana for downstream creature-only
	// enforcement.
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["food_chain_mana_seat_"+intToStr(seat)] += manaGained

	gs.LogEvent(gameengine.Event{
		Kind:   "add_mana",
		Seat:   seat,
		Target: seat,
		Source: src.Card.DisplayName(),
		Amount: manaGained,
		Details: map[string]interface{}{
			"reason":      "food_chain_exile",
			"exiled_card": exiledCard.DisplayName(),
			"exiled_cmc":  cmc,
			"new_pool":    s.ManaPool,
			"creature_only": true,
		},
	})
	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"exiled_card":   exiledCard.DisplayName(),
		"mana_gained":   manaGained,
		"new_pool":      s.ManaPool,
		"creature_only": true,
	})
}

// cardHasExileCastStatic checks whether a card has a static ability that
// lets it be cast from exile (Misthollow Griffin, Eternal Scourge, Squee
// the Immortal, Torrent Elemental). These cards have "You may cast this
// card from exile" in their oracle text. We check for:
//   - A "cast_from_exile" type tag (test-friendly)
//   - The card name matching known exile-casters
func cardHasExileCastStatic(c *gameengine.Card) bool {
	if c == nil {
		return false
	}
	for _, t := range c.Types {
		if t == "cast_from_exile" {
			return true
		}
	}
	// Known cards with "You may cast this card from exile":
	name := normalizeName(c.DisplayName())
	switch name {
	case "misthollow griffin",
		"eternal scourge",
		"squee the immortal",
		"torrent elemental":
		return true
	}
	return false
}
