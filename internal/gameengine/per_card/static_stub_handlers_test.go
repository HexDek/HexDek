package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Tests for the static-stub-handlers batch. Focused per-handler
// behaviors — registry smoke tests + the headline mechanic each card
// brings.

// ---------------------------------------------------------------------
// Avacyn, Angel of Hope — indestructible anthem
// ---------------------------------------------------------------------

func TestAvacyn_GrantsIndestructibleToOtherPermanents(t *testing.T) {
	gs := newGame(t, 2)
	avacyn := addPerm(gs, 0, "Avacyn, Angel of Hope", "creature", "angel")
	bear := addPerm(gs, 0, "Grizzly Bears", "creature")
	enemy := addPerm(gs, 1, "Goblin", "creature")

	avacynGrantIndestructible(gs, avacyn)

	if !bear.IsIndestructible() {
		t.Fatalf("Avacyn should grant indestructible to other permanents we control")
	}
	if enemy.IsIndestructible() {
		t.Fatalf("Avacyn should NOT grant indestructible to opponents' permanents")
	}
}

func TestAvacyn_DoesNotRegrantToSelf(t *testing.T) {
	gs := newGame(t, 2)
	avacyn := addPerm(gs, 0, "Avacyn, Angel of Hope", "creature", "angel")
	avacynGrantIndestructible(gs, avacyn)
	// The grant predicate excludes the source — verifying via the
	// applied flag that we don't double-stamp.
	chars := gameengine.GetEffectiveCharacteristics(gs, avacyn)
	count := 0
	if chars != nil {
		for _, k := range chars.Keywords {
			if k == "indestructible" {
				count++
			}
		}
	}
	if count > 1 {
		t.Fatalf("Avacyn shouldn't double-grant indestructible to herself; got %d", count)
	}
}

// ---------------------------------------------------------------------
// Maelstrom Wanderer — second cascade + haste anthem
// ---------------------------------------------------------------------

func TestMaelstromWanderer_ETBFiresSecondCascadeAndGrantsHaste(t *testing.T) {
	gs := newGame(t, 2)
	// Empty library — cascade whiffs, but the haste anthem grant
	// should still fire. (A non-empty library would resolve a
	// cascade-cast spell back through the engine, which has
	// downstream effects we don't want this focused test to depend
	// on.)
	mw := addPerm(gs, 0, "Maelstrom Wanderer", "creature", "elemental")
	mw.Card.CMC = 8

	bear := addPerm(gs, 0, "Grizzly Bears", "creature")
	bear.SummoningSick = true

	maelstromWandererETB(gs, mw)

	// Cascade event should be in the log (whiff is fine — we just
	// want to verify the second cascade fired).
	if hasEvent(gs, "cascade_trigger")+hasEvent(gs, "cascade_hit")+hasEvent(gs, "cascade_whiff") == 0 {
		t.Fatalf("Maelstrom Wanderer ETB should fire a cascade event")
	}

	// Haste grant — bear should have kw:haste flag and not be sick.
	if bear.Flags["kw:haste"] != 1 {
		t.Fatalf("Maelstrom should grant haste to creatures we control; flags=%v", bear.Flags)
	}
	if bear.SummoningSick {
		t.Fatalf("Maelstrom haste should clear summoning sickness on creatures we control")
	}
}

func TestMaelstromWanderer_DoesNotGrantHasteToOpponents(t *testing.T) {
	gs := newGame(t, 2)
	addLibrary(gs, 0, "Forest")
	mw := addPerm(gs, 0, "Maelstrom Wanderer", "creature")
	mw.Card.CMC = 8
	enemy := addPerm(gs, 1, "Goblin", "creature")
	enemy.SummoningSick = true

	maelstromWandererETB(gs, mw)

	if enemy.Flags["kw:haste"] == 1 {
		t.Fatalf("Maelstrom should NOT grant haste to opponents' creatures")
	}
}

// ---------------------------------------------------------------------
// Feather, the Redeemed — exile-on-resolve + return at end step
// ---------------------------------------------------------------------

func TestFeather_StampsExileOnResolveWhenSpellTargetsOurCreature(t *testing.T) {
	gs := newGame(t, 2)
	feather := addPerm(gs, 0, "Feather, the Redeemed", "creature", "angel")
	target := addPerm(gs, 0, "Bear", "creature")

	bolt := &gameengine.Card{Name: "Lightning Helix", Owner: 0, Types: []string{"instant"}}
	item := &gameengine.StackItem{
		Controller: 0,
		Card:       bolt,
		Kind:       "spell",
		Targets:    []gameengine.Target{{Kind: gameengine.TargetKindPermanent, Permanent: target, Seat: 0}},
	}
	gs.Stack = append(gs.Stack, item)

	featherExileAndReturn(gs, feather, map[string]interface{}{
		"caster_seat": 0,
		"card":        bolt,
	})

	v, ok := item.CostMeta["exile_on_resolve"].(bool)
	if !ok || !v {
		t.Fatalf("Feather should stamp exile_on_resolve=true on the spell; got %v", item.CostMeta)
	}
	if len(gs.DelayedTriggers) == 0 {
		t.Fatalf("Feather should register a delayed trigger to return the card at next end step")
	}
}

func TestFeather_DoesNotStampWhenSpellTargetsOpponentCreature(t *testing.T) {
	gs := newGame(t, 2)
	feather := addPerm(gs, 0, "Feather, the Redeemed", "creature")
	enemy := addPerm(gs, 1, "Goblin", "creature")

	bolt := &gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}}
	item := &gameengine.StackItem{
		Controller: 0,
		Card:       bolt,
		Kind:       "spell",
		Targets:    []gameengine.Target{{Kind: gameengine.TargetKindPermanent, Permanent: enemy, Seat: 1}},
	}
	gs.Stack = append(gs.Stack, item)

	featherExileAndReturn(gs, feather, map[string]interface{}{
		"caster_seat": 0,
		"card":        bolt,
	})

	if v, _ := item.CostMeta["exile_on_resolve"].(bool); v {
		t.Fatalf("Feather should NOT stamp exile_on_resolve when the target is an opponent's creature")
	}
}

// ---------------------------------------------------------------------
// Kess, Dissident Mage — once-per-turn cast-from-graveyard
// ---------------------------------------------------------------------

func TestKess_CastsHighestCMCFromGraveyard(t *testing.T) {
	gs := newGame(t, 2)
	kess := addPerm(gs, 0, "Kess, Dissident Mage", "creature")
	kess.Flags = map[string]int{}

	bolt := &gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant", "cmc:1"}}
	wrath := &gameengine.Card{Name: "Wrath of God", Owner: 0, Types: []string{"sorcery", "cmc:4"}}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, bolt, wrath)
	stackBefore := len(gs.Stack)

	kessCastFromGY(gs, kess, 0, nil)

	if len(gs.Stack) != stackBefore+1 {
		t.Fatalf("Kess should push 1 spell onto the stack; got delta %d", len(gs.Stack)-stackBefore)
	}
	top := gs.Stack[len(gs.Stack)-1]
	if top.Card != wrath {
		t.Fatalf("Kess should pick highest-CMC spell (Wrath); got %s", top.Card.DisplayName())
	}
	if v, _ := top.CostMeta["exile_on_resolve"].(bool); !v {
		t.Fatalf("Kess-cast spell should be flagged exile_on_resolve")
	}
	if kess.Flags["kess_used_this_turn"] != 1 {
		t.Fatalf("Kess should mark used_this_turn=1 after casting")
	}
}

func TestKess_RefusesSecondCastSameTurn(t *testing.T) {
	gs := newGame(t, 2)
	kess := addPerm(gs, 0, "Kess, Dissident Mage", "creature")
	kess.Flags = map[string]int{"kess_used_this_turn": 1}
	bolt := &gameengine.Card{Name: "Lightning Bolt", Owner: 0, Types: []string{"instant"}}
	gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, bolt)
	stackBefore := len(gs.Stack)

	kessCastFromGY(gs, kess, 0, nil)

	if len(gs.Stack) != stackBefore {
		t.Fatalf("Kess should refuse a second cast same turn")
	}
}

// ---------------------------------------------------------------------
// Hogaak, Arisen Necropolis — cast-restriction flags
// ---------------------------------------------------------------------

func TestHogaak_ETBSetsCastRestrictionFlags(t *testing.T) {
	gs := newGame(t, 2)
	hog := addPerm(gs, 0, "Hogaak, Arisen Necropolis", "creature", "avatar")

	hogaakRegisterCastFlags(gs, hog)

	if gs.Flags["hogaak_graveyard_castable_seat"] != 1 {
		t.Fatalf("Hogaak ETB should set graveyard_castable_seat=controller+1; got %d",
			gs.Flags["hogaak_graveyard_castable_seat"])
	}
	if gs.Flags["hogaak_no_mana_cast_seat"] != 1 {
		t.Fatalf("Hogaak ETB should set no_mana_cast_seat=controller+1; got %d",
			gs.Flags["hogaak_no_mana_cast_seat"])
	}
}
