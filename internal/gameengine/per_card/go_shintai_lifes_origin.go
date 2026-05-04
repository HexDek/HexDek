package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerGoShintaiLifesOrigin wires Go-Shintai of Life's Origin.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{W}{U}{B}{R}{G}, {T}: Return target enchantment card from your
//	  graveyard to the battlefield.
//	Whenever Go-Shintai of Life's Origin or another nontoken Shrine you
//	  control enters, create a 1/1 colorless Shrine enchantment creature
//	  token.
//
// Implementation:
//   - "permanent_etb": creature/enchantment Shrine entering, controlled
//     by us, non-token. Mint a 1/1 colorless Shrine enchantment creature
//     token via standard ETB cascade.
//   - Activated ability (5-color reanimate) is documented via
//     emitPartial — full reanimation is supported elsewhere via the AST
//     pipeline; this handler covers only the unique ETB swarm.
func registerGoShintaiLifesOrigin(r *Registry) {
	r.OnETB("Go-Shintai of Life's Origin", goShintaiSelfETB)
	r.OnTrigger("Go-Shintai of Life's Origin", "permanent_etb", goShintaiOtherShrineETB)
}

func goShintaiSelfETB(gs *gameengine.GameState, perm *gameengine.Permanent) {
	goShintaiSpawnToken(gs, perm, perm)
}

func goShintaiOtherShrineETB(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	if ctx == nil {
		return
	}
	enter, _ := ctx["perm"].(*gameengine.Permanent)
	if enter == nil || enter == perm {
		return
	}
	if enter.Controller != perm.Controller {
		return
	}
	if enter.Card == nil || !cardHasType(enter.Card, "shrine") {
		return
	}
	if enter.IsToken() {
		return
	}
	goShintaiSpawnToken(gs, perm, enter)
}

func goShintaiSpawnToken(gs *gameengine.GameState, perm, source *gameengine.Permanent) {
	const slug = "go_shintai_spawn_shrine_token"
	if gs == nil || perm == nil {
		return
	}
	seat := perm.Controller
	token := &gameengine.Card{
		Name:          "Shrine Token",
		Owner:         seat,
		BasePower:     1,
		BaseToughness: 1,
		Types:         []string{"token", "creature", "enchantment", "shrine"},
		Colors:        []string{},
		TypeLine:      "Token Enchantment Creature — Shrine",
	}
	enterBattlefieldWithETB(gs, seat, token, false)
	srcName := ""
	if source != nil && source.Card != nil {
		srcName = source.Card.DisplayName()
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":    seat,
		"trigger": srcName,
	})
}
