package per_card

import (
	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerZurgoThundersDecree wires Zurgo, Thunder's Decree.
//
// Oracle text (Scryfall, verified 2026-05-04):
//
//	{R}{W}{B}
//	Legendary Creature — Orc Warrior
//	2/4
//	Mobilize 2 (Whenever this creature attacks, create two tapped and
//	attacking 1/1 red Warrior creature tokens. Sacrifice them at the
//	beginning of the next end step.)
//	During your end step, Warrior tokens you control have "This token
//	can't be sacrificed."
//
// Implementation:
//   - creature_attacks gated on Zurgo himself: mint 2 tapped 1/1 red
//     Warrior tokens. They enter attacking the same defender (engine
//     handles the attacking-token addition via enterBattlefieldWithETB
//     plus a manual Tapped flag). Set Flags["mobilize_token"] = 1 so the
//     end-step sweep knows which tokens are mobilize-spawned.
//   - Zurgo's static "during your end step the tokens can't be
//     sacrificed" effectively means the mobilize tokens persist past the
//     turn they were made. emitPartial flags that the engine doesn't
//     wire the mobilize cleanup at all — these tokens stay until normal
//     SBAs / wipes remove them. That's actually correct behavior for
//     Zurgo (the static prevents the cleanup) — so the simplification is
//     accidentally faithful for Zurgo specifically.
func registerZurgoThundersDecree(r *Registry) {
	r.OnTrigger("Zurgo, Thunder's Decree", "creature_attacks", zurgoThundersDecreeAttacks)
}

func zurgoThundersDecreeAttacks(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
	const slug = "zurgo_thunders_decree_mobilize_2"
	if gs == nil || perm == nil || ctx == nil {
		return
	}
	atk, _ := ctx["attacker_perm"].(*gameengine.Permanent)
	if atk == nil || atk != perm {
		return
	}
	for i := 0; i < 2; i++ {
		token := &gameengine.Card{
			Name:          "Warrior Token",
			Owner:         perm.Controller,
			BasePower:     1,
			BaseToughness: 1,
			Types:         []string{"token", "creature", "warrior"},
			Colors:        []string{"R"},
			TypeLine:      "Token Creature — Warrior",
		}
		tokenPerm := enterBattlefieldWithETB(gs, perm.Controller, token, true)
		if tokenPerm != nil {
			if tokenPerm.Flags == nil {
				tokenPerm.Flags = map[string]int{}
			}
			tokenPerm.Flags["mobilize_token"] = 1
			tokenPerm.Flags["zurgo_no_sac_at_eot"] = 1
		}
	}
	emit(gs, slug, perm.Card.DisplayName(), map[string]interface{}{
		"seat":   perm.Controller,
		"tokens": 2,
	})
}
