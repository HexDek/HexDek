package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerEvercoatUrsine wires Evercoat Ursine.
//
// Oracle text (Scryfall, verified 2026-05-16 via hexdek.dev oracle):
//
//	Trample
//	Hideaway 3, hideaway 3 (When this creature enters, look at the top
//	three cards of your library, exile one face down, then put the rest
//	on the bottom in a random order. Then do it again.)
//	Whenever this creature deals combat damage to a player, if there
//	are cards exiled with it, you may play one of them without paying
//	its mana cost.
//
// Implementation (Muninn gap #33 — 29K hits):
//   - Trample handled by AST keyword pipeline.
//   - ETB fires ApplyHideaway twice. The keywords_batch4.go helper
//     exiles one of the top three face down and shuffles the rest to
//     the bottom in random order.
//   - The combat-damage-trigger free-cast is left as emitPartial:
//     ApplyHideaway only stamps perm.Flags["hideaway"]=1 and doesn't
//     maintain a per-permanent list of exiled-with-it cards, and there
//     is no engine API yet for "cast from exile without paying" tied
//     to a specific source.
func registerEvercoatUrsine(r *Registry) {
	r.OnETB("Evercoat Ursine", evercoatUrsineETB)
}

func evercoatUrsineETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "evercoat_ursine_hideaway_twice"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	seat := perm.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	gameengine.ApplyHideaway(gs, perm, 3)
	gameengine.ApplyHideaway(gs, perm, 3)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat": seat,
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"combat_damage_play_exiled_card_for_free_unmodeled_no_per_perm_exile_list")
}
