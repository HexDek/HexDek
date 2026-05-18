package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerPrestonTheVanisher wires Preston, the Vanisher (Muninn parser-gap
// #99, ~3.5K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{2}{W}
//	Legendary Creature — Human Wizard
//	Whenever another nontoken creature you control enters, if it wasn't
//	cast, create a token that's a copy of that creature, except it's a
//	0/1 white Illusion.
//	{1}{W}, Sacrifice five Illusions: Exile target nonland permanent.
//
// Implementation:
//   - "permanent_etb" listener: for each ETB whose entering perm is a
//     nontoken creature under our control, with was_cast == 0 (so blink/
//     reanimate/cheat-into-play paths qualify) and not Preston itself,
//     mint a token copy with 0/1 white Illusion characteristics overlaid.
//   - The activated "sacrifice five Illusions" ability is an explicit
//     activated cost requiring stack/cost machinery and is best left to
//     the hat's activation pass when it materializes — partial.
func registerPrestonTheVanisher(r *Registry) {
	r.OnTrigger("Preston, the Vanisher", "permanent_etb", prestonVanisherETBTrigger)
}

func prestonVanisherETBTrigger(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "preston_vanisher_illusion_copy"
	if gs == nil || perm == nil || ctx == nil || perm.Card == nil {
		return
	}
	entering, _ := ctx["perm"].(*gameengine.Permanent)
	if entering == nil {
		entering, _ = ctx["permanent"].(*gameengine.Permanent)
	}
	if entering == nil || entering.Card == nil || entering == perm {
		return
	}
	if entering.Controller != perm.Controller {
		return
	}
	if !entering.IsCreature() {
		return
	}
	if cardHasType(entering.Card, "token") {
		return
	}
	// "if it wasn't cast" — was_cast flag set by stack.go on the cast path.
	if entering.Flags != nil && entering.Flags["was_cast"] == 1 {
		emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
			"seat":      perm.Controller,
			"triggered": false,
			"reason":    "creature_was_cast",
			"source":    entering.Card.DisplayName(),
		})
		return
	}
	// Token-copy overlay: 0/1 white Illusion, otherwise copyable values
	// (name, types, keywords) carried from the entering card. CR §707.10f:
	// the copy effect changes the new token's P/T, color, and creature type
	// to 0/1 white Illusion, but other copyable characteristics stay.
	src := entering.Card
	types := append([]string{}, src.Types...)
	// Drop original creature-subtype-ish characteristics tagged with pip:
	// (color) since we're overriding to white-only, and any non-Illusion
	// subtypes still passthrough — the copy effect only swaps in Illusion
	// for the creature subtype slot. We add "illusion" and ensure "token".
	filtered := types[:0]
	for _, t := range types {
		switch t {
		case "pip:R", "pip:G", "pip:B", "pip:U":
			continue
		default:
			filtered = append(filtered, t)
		}
	}
	hasToken := false
	hasIllusion := false
	hasPipW := false
	for _, t := range filtered {
		if t == "token" {
			hasToken = true
		}
		if t == "illusion" {
			hasIllusion = true
		}
		if t == "pip:W" {
			hasPipW = true
		}
	}
	if !hasToken {
		filtered = append(filtered, "token")
	}
	if !hasIllusion {
		filtered = append(filtered, "illusion")
	}
	if !hasPipW {
		filtered = append(filtered, "pip:W")
	}
	token := &gameengine.Card{
		Name:          src.DisplayName() + " (Illusion token)",
		Owner:         perm.Controller,
		BasePower:     0,
		BaseToughness: 1,
		Types:         filtered,
		Colors:        []string{"W"},
		TypeLine:      src.TypeLine,
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"source": src.DisplayName(),
		"token":  token.DisplayName(),
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"activated_sacrifice_five_illusions_exile_unmodeled")
}
