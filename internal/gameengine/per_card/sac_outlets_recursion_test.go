package per_card

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// TestSacVictim_PrefersUnearthCreature: given two creatures (one with
// unearth, one vanilla), chooseSacVictim should pick the unearth one.
// Unearth is a graveyard zone-cast grant — sacrificing the creature is
// half-free because it can be returned later.
func TestSacVictim_PrefersUnearthCreature(t *testing.T) {
	gs := newGame(t, 2)
	outlet := addPerm(gs, 0, "Ashnod's Altar", "artifact")

	vanilla := addPerm(gs, 0, "Grizzly Bears", "creature")
	vanilla.Card.BasePower = 2
	vanilla.Card.BaseToughness = 2

	unearther := addPerm(gs, 0, "Stitched Drake", "creature")
	unearther.Card.BasePower = 3
	unearther.Card.BaseToughness = 3
	unearther.Flags["kw:unearth"] = 1

	pick := chooseSacVictim(gs, 0, outlet, nil)
	if pick == nil {
		t.Fatal("expected a sacrifice victim, got nil")
	}
	if pick != unearther {
		t.Fatalf("expected unearth creature to be preferred, got %s",
			pick.Card.DisplayName())
	}
}

// TestSacVictim_PrefersPersistOverVanilla: persist is automatic recursion;
// the creature comes right back with a -1/-1 counter. Almost free fodder.
func TestSacVictim_PrefersPersistOverVanilla(t *testing.T) {
	gs := newGame(t, 2)
	outlet := addPerm(gs, 0, "Viscera Seer", "creature")

	vanilla := addPerm(gs, 0, "Grizzly Bears", "creature")
	vanilla.Card.BasePower = 2
	vanilla.Card.BaseToughness = 2

	persister := addPerm(gs, 0, "Murderous Redcap", "creature")
	persister.Card.BasePower = 2
	persister.Card.BaseToughness = 2
	persister.Flags["kw:persist"] = 1

	pick := chooseSacVictimNotSelf(gs, 0, outlet, nil)
	if pick == nil {
		t.Fatal("expected a sacrifice victim, got nil")
	}
	if pick != persister {
		t.Fatalf("expected persist creature to be preferred, got %s",
			pick.Card.DisplayName())
	}
}

// TestSacVictim_RecursionEngineBoostsScore: a Phyrexian Reclamation on
// the battlefield should raise every creature's sacrifice score (since
// any creature in the graveyard becomes recurrable). All-vanilla plus
// engine vs all-vanilla without engine — the engine seat's score for a
// given creature should be higher.
func TestSacVictim_RecursionEngineBoostsScore(t *testing.T) {
	gs := newGame(t, 2)
	outlet := addPerm(gs, 0, "Ashnod's Altar", "artifact")

	target := addPerm(gs, 0, "Grizzly Bears", "creature")
	target.Card.BasePower = 2
	target.Card.BaseToughness = 2

	scoreBefore := sacVictimScore(gs, 0, target, outlet)

	addPerm(gs, 0, "Phyrexian Reclamation", "enchantment")
	scoreAfter := sacVictimScore(gs, 0, target, outlet)

	if scoreAfter <= scoreBefore {
		t.Fatalf("recursion engine should raise sacrifice score; before=%.2f after=%.2f",
			scoreBefore, scoreAfter)
	}
}

// TestSacVictim_EscapeCreaturePreferred: escape grants graveyard
// zone-cast just like unearth — should be preferred over a vanilla.
func TestSacVictim_EscapeCreaturePreferred(t *testing.T) {
	gs := newGame(t, 2)
	outlet := addPerm(gs, 0, "Ashnod's Altar", "artifact")

	vanilla := addPerm(gs, 0, "Grizzly Bears", "creature")
	vanilla.Card.BasePower = 2
	vanilla.Card.BaseToughness = 2

	escaper := addPerm(gs, 0, "Uro, Titan of Nature's Wrath", "creature")
	escaper.Card.BasePower = 6
	escaper.Card.BaseToughness = 6
	escaper.Flags["kw:escape"] = 1

	pick := chooseSacVictim(gs, 0, outlet, nil)
	if pick == nil {
		t.Fatal("expected a sacrifice victim, got nil")
	}
	if pick != escaper {
		t.Fatalf("expected escape creature to be preferred over vanilla, got %s",
			pick.Card.DisplayName())
	}
}

// TestSacVictim_ZoneCastGrantBoostsScore: a card-level grant in
// gs.ZoneCastGrants (e.g. Underworld Breach grants escape to instants/
// sorceries; we approximate via a creature whose card has a graveyard
// permission entry) should boost the score.
func TestSacVictim_ZoneCastGrantBoostsScore(t *testing.T) {
	gs := newGame(t, 2)
	gs.ZoneCastGrants = map[*gameengine.Card]*gameengine.ZoneCastPermission{}
	outlet := addPerm(gs, 0, "Ashnod's Altar", "artifact")

	target := addPerm(gs, 0, "Plain Goblin", "creature")
	target.Card.BasePower = 2
	target.Card.BaseToughness = 2

	scoreBefore := sacVictimScore(gs, 0, target, outlet)

	gs.ZoneCastGrants[target.Card] = &gameengine.ZoneCastPermission{
		Zone:    gameengine.ZoneGraveyard,
		Keyword: "escape",
	}
	scoreAfter := sacVictimScore(gs, 0, target, outlet)

	if scoreAfter <= scoreBefore {
		t.Fatalf("graveyard zone-cast grant should raise sacrifice score; before=%.2f after=%.2f",
			scoreBefore, scoreAfter)
	}
}
