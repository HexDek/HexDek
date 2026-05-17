package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTolsimirMidnightsLight wires Tolsimir, Midnight's Light (Muninn
// parser-gap #93, ~4.7K hits).
//
// Oracle text (Scryfall, verified 2026-05-17 via hexdek.dev oracle):
//
//	{2}{G}{W}
//	Legendary Creature — Elf
//	Lifelink
//	When Tolsimir enters, create Voja Fenstalker, a legendary 5/5 green
//	and white Wolf creature token with trample.
//	Whenever a Wolf you control attacks, if Tolsimir attacked this combat,
//	target creature an opponent controls blocks that Wolf this combat if
//	able.
//
// Implementation:
//   - Lifelink: AST keyword pipeline.
//   - OnETB: mint Voja Fenstalker, a legendary 5/5 green/white Wolf with
//     trample. enterBattlefieldWithETB routes through the same ETB pipeline
//     as resolveCreateTokenCopy so trample registers as a printed keyword.
//   - Combat-blocking rider ("blocks that Wolf this combat if able") is a
//     forced-blocker effect that requires the same combat-attribute layer
//     other forced-block cards (Lure, etc.) need — partial.
func registerTolsimirMidnightsLight(r *Registry) {
	r.OnETB("Tolsimir, Midnight's Light", tolsimirMidnightsLightETB)
}

func tolsimirMidnightsLightETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	const slug = "tolsimir_midnights_light_etb_voja_token"
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	token := &gameengine.Card{
		Name:          "Voja Fenstalker",
		Owner:         perm.Controller,
		Types:         []string{"legendary", "creature", "token", "wolf", "pip:G", "pip:W", "kw:trample"},
		Colors:        []string{"G", "W"},
		BasePower:     5,
		BaseToughness: 5,
	}
	enterBattlefieldWithETB(gs, perm.Controller, token, false)
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":  perm.Controller,
		"token": "Voja Fenstalker (legendary 5/5 G/W Wolf, trample)",
	})
	emitPartial(gs, slug, perm.Card.DisplayName(),
		"forced_blocker_wolf_attack_rider_unmodeled")
}
