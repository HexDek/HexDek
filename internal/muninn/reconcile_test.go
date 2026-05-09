package muninn

import (
	"path/filepath"
	"testing"
)

func TestArchiveFixedCards_FiltersAndPersists(t *testing.T) {
	dir := t.TempDir()

	// Seed dead_triggers.json with three entries — two fixed, one not.
	seed := []DeadTrigger{
		{TriggerName: "triggered_ability", CardName: "Giada, Font of Hope", Count: 7, GamesSeen: 3, LastSeen: "2026-05-08T12:00:00Z"},
		{TriggerName: "triggered_ability", CardName: "Sliver Gravemother", Count: 4, GamesSeen: 2, LastSeen: "2026-05-08T12:00:00Z"},
		{TriggerName: "triggered_ability", CardName: "Avenger of Zendikar", Count: 11, GamesSeen: 5, LastSeen: "2026-05-08T12:00:00Z"},
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersFile), seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := ArchiveFixedCards(dir, []string{"Giada, Font of Hope", "Sliver Gravemother"}, "test cause")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}

	if res.DeadTriggersBefore != 3 {
		t.Errorf("before=%d, want 3", res.DeadTriggersBefore)
	}
	if res.DeadTriggersAfter != 1 {
		t.Errorf("after=%d, want 1", res.DeadTriggersAfter)
	}
	if res.DeadTriggersArchived != 2 {
		t.Errorf("archived=%d, want 2", res.DeadTriggersArchived)
	}
	if len(res.UnmatchedCards) != 0 {
		t.Errorf("unmatched=%v, want []", res.UnmatchedCards)
	}

	live, err := ReadDeadTriggers(dir)
	if err != nil {
		t.Fatalf("read live: %v", err)
	}
	if len(live) != 1 || live[0].CardName != "Avenger of Zendikar" {
		t.Errorf("live=%v, want only Avenger", live)
	}

	arch, err := ReadDeadTriggersArchive(dir)
	if err != nil {
		t.Fatalf("read arch: %v", err)
	}
	if len(arch) != 2 {
		t.Fatalf("arch len=%d, want 2", len(arch))
	}
	for _, a := range arch {
		if a.ArchiveCause != "test cause" {
			t.Errorf("cause=%q, want test cause", a.ArchiveCause)
		}
		if a.ArchivedAt == "" {
			t.Errorf("archived_at empty for %s", a.CardName)
		}
	}
}

func TestArchiveFixedCards_AppendsToExistingArchive(t *testing.T) {
	dir := t.TempDir()

	// Pre-existing archive entry from a prior run.
	prior := []ArchivedDeadTrigger{
		{
			DeadTrigger:  DeadTrigger{TriggerName: "triggered_ability", CardName: "Old Card", Count: 1, GamesSeen: 1, LastSeen: "2026-04-01T00:00:00Z"},
			ArchivedAt:   "2026-04-01T00:00:00Z",
			ArchiveCause: "earlier pass",
		},
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersArchiveFile), prior); err != nil {
		t.Fatalf("seed archive: %v", err)
	}

	live := []DeadTrigger{
		{TriggerName: "triggered_ability", CardName: "Tiamat", Count: 2, GamesSeen: 1, LastSeen: "2026-05-08T12:00:00Z"},
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersFile), live); err != nil {
		t.Fatalf("seed live: %v", err)
	}

	if _, err := ArchiveFixedCards(dir, []string{"Tiamat"}, "era4"); err != nil {
		t.Fatalf("archive: %v", err)
	}

	arch, err := ReadDeadTriggersArchive(dir)
	if err != nil {
		t.Fatalf("read arch: %v", err)
	}
	if len(arch) != 2 {
		t.Fatalf("arch len=%d, want 2 (prior + new)", len(arch))
	}
	if arch[0].CardName != "Old Card" || arch[1].CardName != "Tiamat" {
		t.Errorf("arch order=%v, want [Old Card, Tiamat]", []string{arch[0].CardName, arch[1].CardName})
	}
}

func TestArchiveFixedCards_DFCFaceMatching(t *testing.T) {
	dir := t.TempDir()

	// Live entry uses just the front face; fixed list has the joined name.
	seed := []DeadTrigger{
		{TriggerName: "triggered_ability", CardName: "Eddie Brock", Count: 3, GamesSeen: 1, LastSeen: "2026-05-08T12:00:00Z"},
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersFile), seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	res, err := ArchiveFixedCards(dir, []string{"Eddie Brock // Venom, Lethal Protector"}, "dfc test")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if res.DeadTriggersArchived != 1 {
		t.Fatalf("archived=%d, want 1 (front face match)", res.DeadTriggersArchived)
	}
}

func TestArchiveFixedCards_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	seed := []DeadTrigger{
		{TriggerName: "triggered_ability", CardName: "  GIADA, Font of Hope ", Count: 1, GamesSeen: 1, LastSeen: "2026-05-08T12:00:00Z"},
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersFile), seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := ArchiveFixedCards(dir, []string{"giada, font of hope"}, "case test")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if res.DeadTriggersArchived != 1 {
		t.Errorf("archived=%d, want 1", res.DeadTriggersArchived)
	}
}

func TestArchiveFixedCards_NoMatchesReturnsUnmatched(t *testing.T) {
	dir := t.TempDir()
	seed := []DeadTrigger{
		{TriggerName: "triggered_ability", CardName: "Avenger of Zendikar", Count: 11, GamesSeen: 5, LastSeen: "2026-05-08T12:00:00Z"},
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersFile), seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := ArchiveFixedCards(dir, []string{"Tiamat", "Mayael the Anima"}, "no-op")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if res.DeadTriggersArchived != 0 {
		t.Errorf("archived=%d, want 0", res.DeadTriggersArchived)
	}
	if len(res.UnmatchedCards) != 2 {
		t.Errorf("unmatched=%v, want both", res.UnmatchedCards)
	}

	// Live file should be untouched.
	live, _ := ReadDeadTriggers(dir)
	if len(live) != 1 {
		t.Errorf("live len=%d, want 1 (untouched)", len(live))
	}
}

func TestArchiveFixedCards_EmptyDirEmptyList(t *testing.T) {
	dir := t.TempDir()
	res, err := ArchiveFixedCards(dir, nil, "")
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	if res.DeadTriggersArchived != 0 || res.DeadTriggersBefore != 0 {
		t.Errorf("expected zero result, got %+v", res)
	}
}

func TestEraPassFixedCardsManifest(t *testing.T) {
	if len(EraPassFixedCards) < 40 {
		t.Errorf("EraPassFixedCards has %d entries; expected >=40 from 2026-05-09 era passes",
			len(EraPassFixedCards))
	}
	seen := make(map[string]bool)
	for _, name := range EraPassFixedCards {
		if name == "" {
			t.Errorf("empty entry in EraPassFixedCards")
			continue
		}
		if seen[name] {
			t.Errorf("duplicate entry: %q", name)
		}
		seen[name] = true
	}
}
