package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// shuffle_self_from_grave_family.go — generic handler for the Planar Chaos
// eternal cycle's signature line:
//
//	When ~ is put into a graveyard from anywhere, shuffle it into its
//	owner's library.
//
// Members (Time Spiral block / Planar Chaos eternal cycle):
//
//   - Dread        (Whenever a creature deals damage to you, destroy it.)
//   - Purity       (Prevent noncombat damage; you gain life equal.)
//   - Guile        (Counter→exile-and-cast replacement.)
//   - Vigor        (Prevent damage to your other creatures, +1/+1 instead.)
//
// Worldspine Wurm shares the exact line but already has a dedicated
// handler (worldspine_wurm.go) that bundles the die→3 Wurm tokens half;
// the reshuffle half there is implemented inline. The family entry list
// flags Worldspine in a comment so a future migration can fold it back
// in cleanly.
//
// CR / engine notes:
//
//   - "from anywhere" canonically covers battlefield→graveyard,
//     library→graveyard (mill / Liliana / Hermit Druid), hand→graveyard
//     (discard), exile→graveyard (rare, e.g. Decree of Pain rebound),
//     and stack→graveyard (countered).
//   - The per_card trigger dispatcher (registry.go fireTrigger) iterates
//     gs.Seats[*].Battlefield only — so a registered handler on a card
//     name fires only when a permanent of that name is on the
//     battlefield. We hook creature_dies, which fires AFTER the
//     battlefield→graveyard zone move but BEFORE the perm is removed
//     from the battlefield iteration target (the dying permanent is
//     passed in as ctx["perm"]).
//   - That covers the battlefield path — by far the dominant one for
//     these {6}-{8} mana fatties: cast → die in combat / wrath. The
//     other paths (mill / discard / exile→graveyard) DO need a trigger
//     dispatcher that scans the graveyard at zone-change time. Same
//     gap noted on Ichorid / Master of Death / Sproutback Trudge for
//     the inverse direction (graveyard-resident upkeep triggers). We
//     emitPartial on every body so Heimdall/Muninn track the gap.

type shuffleSelfFromGraveEntry struct {
	cardName string
}

var shuffleSelfFromGraveEntries = []shuffleSelfFromGraveEntry{
	// Dread — {2}{B}{B}{B}, 6/6 Horror with Fear + damage-destroy.
	{cardName: "Dread"},
	// Purity — {2}{W}{W}{W}, 6/6 Incarnation with flying + noncombat
	// damage prevention/lifegain.
	{cardName: "Purity"},
	// Guile — {2}{U}{U}{U}, 6/6 Incarnation that exile-replaces your
	// own counterspells into "exile + cast for free".
	{cardName: "Guile"},
	// Vigor — {2}{G}{G}{G}, 6/6 Incarnation with trample that
	// prevent-replaces damage to your other creatures into +1/+1
	// counters.
	{cardName: "Vigor"},
	// Sibling intentionally NOT registered here: Worldspine Wurm
	// already has a dedicated handler (worldspine_wurm.go) that bundles
	// the die→three-Wurm-tokens half with the same reshuffle.
}

func registerShuffleSelfFromGraveFamily(r *Registry) {
	for _, e := range shuffleSelfFromGraveEntries {
		e := e
		r.OnTrigger(e.cardName, "creature_dies", func(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
			runShuffleSelfFromGrave(gs, perm, e, ctx)
		})
	}
}

func runShuffleSelfFromGrave(gs *gameengine.GameState, perm *gameengine.Permanent, e shuffleSelfFromGraveEntry, ctx map[string]interface{}) {
	slug := "shuffle_self_from_grave_family:" + landFetchSlug(e.cardName)
	if gs == nil || perm == nil || perm.Card == nil || ctx == nil {
		return
	}
	dying, _ := ctx["perm"].(*gameengine.Permanent)
	if dying != perm {
		return
	}

	// creature_dies fires after the battlefield→graveyard zone move,
	// so the card is in its owner's graveyard now. Defend against a
	// missing/oob owner by falling back to perm.Controller.
	owner := perm.Card.Owner
	if owner < 0 || owner >= len(gs.Seats) {
		owner = perm.Controller
	}
	if owner < 0 || owner >= len(gs.Seats) {
		emitFail(gs, slug, perm.Card.DisplayName(), "no_owner", nil)
		return
	}

	gameengine.MoveCard(gs, perm.Card, owner, "graveyard", "library", slug)
	shuffleLibraryPerCard(gs, owner)

	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"owner": owner,
		"path":  "battlefield_to_graveyard",
	})

	// "from anywhere" also covers mill / discard / exile→graveyard /
	// stack→graveyard. Those routes never produce a battlefield
	// permanent, so this handler can't fire from them; document the
	// gap so Heimdall/Muninn track it.
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"reshuffle_only_modelled_for_battlefield_death_path_other_zones_unhooked")
}
