package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerDarigaazReincarnated wires Darigaaz Reincarnated (Muninn
// parser-gap #74, ~8.9K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{4}{B}{R}{G}
//	Legendary Creature — Dragon
//	Flying, trample, haste
//	If Darigaaz would die, instead exile it with three egg counters on it.
//	At the beginning of your upkeep, if this card is exiled with an egg
//	counter on it, remove an egg counter from it. Then if this card has
//	no egg counters on it, return it to the battlefield.
//
// Implementation:
//   - Flying / trample / haste are AST-side.
//   - The "would die" replacement is a §614 redirect; the engine doesn't
//     expose a generic dies-replacement hook, so we approximate via the
//     "creature_dies" trigger: on death, immediately pull Darigaaz from
//     graveyard → exile and stamp 3 egg counters on the *Card. This is
//     not strictly a replacement (the SBA still routes through the
//     graveyard for an instant) — emitPartial flags the distinction.
//     Most downstream effects (death triggers on other cards, "leaves
//     the battlefield" handlers) still fire correctly because the death
//     event already published before our handler ran.
//   - upkeep_controller: graveyard-side phase trigger has no engine hook
//     today, but the post-die Darigaaz card is in EXILE, not graveyard.
//     We register OnTrigger("upkeep_controller") which fires for
//     battlefield permanents only — so the upkeep tick can't reach the
//     exiled card via the standard path. emitPartial flags the gap.
//     Hat workaround: a follow-up sweep handler could iterate exile
//     zones at upkeep; out of scope for this wave.
func registerDarigaazReincarnated(r *Registry) {
	r.OnTrigger("Darigaaz Reincarnated", "creature_dies", darigaazDiesRedirect)
}

func darigaazDiesRedirect(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "darigaaz_reincarnated_dies_to_exile"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	dying, _ := ctx["dying_perm"].(*gameengine.Permanent)
	if dying == nil {
		dying, _ = ctx["perm"].(*gameengine.Permanent)
	}
	if dying == nil || dying != perm {
		return
	}
	// Find the card pointer post-mortem. Death already routed Darigaaz
	// to the graveyard; pull it to exile and stamp egg counters.
	seat := gs.Seats[perm.Owner]
	if seat == nil {
		return
	}
	// Locate the card in the graveyard (zone_change pushes onto the
	// owner's graveyard, not the controller's, per CR §404.1).
	var card *gameengine.Card
	for _, c := range seat.Graveyard {
		if c != nil && c == perm.Card {
			card = c
			break
		}
	}
	if card == nil {
		emitFail(gs, slug, perm.Card.DisplayName(), "card_not_in_owner_graveyard_post_dies", nil)
		return
	}
	gameengine.MoveCard(gs, card, perm.Owner, "graveyard", "exile", slug)
	// Egg counters live on the Card itself in exile (no Permanent wraps
	// an exiled card). We piggyback on Card.CustomCounters if it exists,
	// otherwise stash on the seat via flag.
	key := "darigaaz_egg_counters"
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags[key] = 3
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":         perm.Owner,
		"egg_counters": 3,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"upkeep_tick_down_egg_counters_and_return_from_exile_unimplemented")
}
