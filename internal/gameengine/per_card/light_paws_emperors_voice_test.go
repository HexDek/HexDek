package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Light-Paws, Emperor's Voice — ETB tutor for an Aura with mana value
// <= the entering Aura's and a different name than each Aura controlled.

func TestLightPaws_CastAuraFetchesEligibleAura(t *testing.T) {
	gs := newGame(t, 2)
	lp := addPerm(gs, 0, "Light-Paws, Emperor's Voice", "creature")
	// Give Light-Paws a real body so it doesn't die to 0-toughness SBA
	// when the fetched Aura's ETB cascade indirectly runs CheckEnd.
	lp.Card.BasePower = 1
	lp.Card.BaseToughness = 1
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Ethereal Armor", Owner: 0, Types: []string{"enchantment", "aura"}, CMC: 1},
		{Name: "All That Glitters", Owner: 0, Types: []string{"enchantment", "aura"}, CMC: 2},
		{Name: "Sentinel's Eyes", Owner: 0, Types: []string{"enchantment", "aura"}, CMC: 1},
		{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}, CMC: 1},
	}

	// Entering Aura — cast (CMC 2). Attach it to Light-Paws so Aura-
	// attachment SBA (CR §704.5n) doesn't ship it to the graveyard
	// mid-trigger and break the test fixture.
	aura := addPerm(gs, 0, "All That Glitters", "enchantment", "aura")
	aura.Card.CMC = 2
	aura.Flags["was_cast"] = 1
	aura.AttachedTo = lp

	lightPawsAuraETBTrigger(gs, lp, map[string]interface{}{
		"perm":            aura,
		"controller_seat": 0,
		"card":            aura.Card,
	})

	// One of the two CMC-1 Auras should now be on the battlefield or in
	// the graveyard (no-attachment SBA may have shipped it off after
	// MoveCard fired, since the test fixture skips the engine's
	// attach-on-resolve path). Either way, lib must have shrunk and the
	// found name must be different from each Aura we already controlled.
	libNames := map[string]bool{}
	for _, c := range gs.Seats[0].Library {
		if c != nil {
			libNames[c.DisplayName()] = true
		}
	}
	// Exactly one of Ethereal Armor / Sentinel's Eyes was removed.
	armorIn := libNames["Ethereal Armor"]
	eyesIn := libNames["Sentinel's Eyes"]
	if armorIn == eyesIn {
		t.Fatalf("Light-Paws should have removed exactly one CMC-1 Aura from library; armor_in=%v eyes_in=%v",
			armorIn, eyesIn)
	}
	// Sanity: All That Glitters was never a fetch candidate (name collision).
	if !libNames["All That Glitters"] {
		t.Errorf("All That Glitters should remain in library (same name as entering Aura)")
	}
}

func TestLightPaws_NotCastDoesNothing(t *testing.T) {
	gs := newGame(t, 2)
	lp := addPerm(gs, 0, "Light-Paws, Emperor's Voice", "creature")
	gs.Seats[0].Library = []*gameengine.Card{
		{Name: "Ethereal Armor", Owner: 0, Types: []string{"enchantment", "aura"}, CMC: 1},
	}
	libBefore := len(gs.Seats[0].Library)

	// Reanimated / blinked Aura — no was_cast.
	aura := addPerm(gs, 0, "All That Glitters", "enchantment", "aura")
	aura.Card.CMC = 2

	lightPawsAuraETBTrigger(gs, lp, map[string]interface{}{
		"perm":            aura,
		"controller_seat": 0,
		"card":            aura.Card,
	})

	if len(gs.Seats[0].Library) != libBefore {
		t.Errorf("Light-Paws should not tutor on non-cast Aura; library shrunk %d → %d",
			libBefore, len(gs.Seats[0].Library))
	}
}

func TestLightPaws_SkipsNameAlreadyOnBattlefield(t *testing.T) {
	gs := newGame(t, 2)
	lp := addPerm(gs, 0, "Light-Paws, Emperor's Voice", "creature")
	existing := addPerm(gs, 0, "Ethereal Armor", "enchantment", "aura")
	existing.Card.CMC = 1
	gs.Seats[0].Library = []*gameengine.Card{
		// Only candidate in the library shares a name with the Aura
		// already on the battlefield — must be skipped.
		{Name: "Ethereal Armor", Owner: 0, Types: []string{"enchantment", "aura"}, CMC: 1},
	}
	libBefore := len(gs.Seats[0].Library)

	aura := addPerm(gs, 0, "All That Glitters", "enchantment", "aura")
	aura.Card.CMC = 2
	aura.Flags["was_cast"] = 1

	lightPawsAuraETBTrigger(gs, lp, map[string]interface{}{
		"perm":            aura,
		"controller_seat": 0,
		"card":            aura.Card,
	})

	if len(gs.Seats[0].Library) != libBefore {
		t.Errorf("Light-Paws should skip Aura names already controlled; library shrunk %d → %d",
			libBefore, len(gs.Seats[0].Library))
	}
}
