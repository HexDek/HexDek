package hexapi

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

func TestFormatEvent_NewKinds(t *testing.T) {
	commanders := []string{"Tinybones, Trinket Thief", "Korvold, Fae-Cursed King", "Atraxa, Praetors' Voice", "Urza, Lord High Artificer"}

	tests := []struct {
		name   string
		event  gameengine.Event
		wantOK bool
		kind   string
		source string
	}{
		{"untap_done", gameengine.Event{Kind: "untap_done", Seat: 0, Source: "Swamp"}, true, "untap", "Swamp"},
		{"tap", gameengine.Event{Kind: "tap", Seat: 1, Source: "Sol Ring"}, true, "tap", "Sol Ring"},
		{"discard_card", gameengine.Event{Kind: "discard", Seat: 2, Source: "Lightning Bolt", Amount: 1}, true, "discard", "Lightning Bolt"},
		{"scry", gameengine.Event{Kind: "scry", Seat: 0, Amount: 3}, true, "scry", ""},
		{"surveil", gameengine.Event{Kind: "surveil", Seat: 1, Amount: 2}, true, "surveil", ""},
		{"shuffle", gameengine.Event{Kind: "shuffle", Seat: 3}, true, "shuffle", ""},
		{"bounce", gameengine.Event{Kind: "bounce", Seat: 0, Source: "Rhystic Study"}, true, "removal", "Rhystic Study"},
		{"flicker", gameengine.Event{Kind: "flicker", Seat: 1, Source: "Deadeye Navigator"}, true, "removal", "Deadeye Navigator"},
		{"equip", gameengine.Event{Kind: "equip", Seat: 2, Source: "Lightning Greaves"}, true, "equip", "Lightning Greaves"},
		{"pay_life", gameengine.Event{Kind: "pay_life", Seat: 0, Amount: 3, Source: "Necropotence"}, true, "life", "Necropotence"},
		{"cascade_hit", gameengine.Event{Kind: "cascade_hit", Seat: 1, Source: "Maelstrom Wanderer"}, true, "cast", "Maelstrom Wanderer"},
		{"become_monarch", gameengine.Event{Kind: "become_monarch", Seat: 2}, true, "monarch", ""},
		{"draw_single", gameengine.Event{Kind: "draw", Seat: 0, Amount: 1}, true, "draw", ""},
		{"counter_fizzle", gameengine.Event{Kind: "counter_spell_fizzle", Seat: 1, Source: "Counterspell"}, true, "counter", "Counterspell"},
		// Skipped events
		{"empty_tap", gameengine.Event{Kind: "tap", Seat: 0, Source: ""}, false, "", ""},
		{"zero_scry", gameengine.Event{Kind: "scry", Seat: 0, Amount: 0}, false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, ok := formatEvent(tt.event, commanders, 5)
			if ok != tt.wantOK {
				t.Fatalf("formatEvent(%s) ok=%v, want %v", tt.name, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if entry.Kind != tt.kind {
				t.Errorf("kind=%q, want %q", entry.Kind, tt.kind)
			}
			if entry.Source != tt.source {
				t.Errorf("source=%q, want %q", entry.Source, tt.source)
			}
			if entry.Seat != tt.event.Seat {
				t.Errorf("seat=%d, want %d", entry.Seat, tt.event.Seat)
			}
		})
	}
}

func TestCoalesceEntries_UntapGroup(t *testing.T) {
	entries := []LogEntry{
		{Turn: 3, Seat: 0, Action: "TINYBONES UNTAPS SWAMP", Kind: "untap", Source: "Swamp", Amount: 1},
		{Turn: 3, Seat: 0, Action: "TINYBONES UNTAPS SWAMP", Kind: "untap", Source: "Swamp", Amount: 1},
		{Turn: 3, Seat: 0, Action: "TINYBONES UNTAPS SOL RING", Kind: "untap", Source: "Sol Ring", Amount: 1},
		{Turn: 3, Seat: 0, Action: "TINYBONES UNTAPS DARK RITUAL", Kind: "untap", Source: "Dark Ritual", Amount: 1},
		{Turn: 3, Seat: 0, Action: "TINYBONES CASTS SOMETHING", Kind: "cast", Source: "Something", Amount: 1},
	}

	result := coalesceEntries(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(result), result)
	}

	// First should be coalesced untap group
	if result[0].Count != 4 {
		t.Errorf("coalesced count=%d, want 4", result[0].Count)
	}
	if result[0].Kind != "untap" {
		t.Errorf("kind=%q, want untap", result[0].Kind)
	}
	// Should have unique targets
	if len(result[0].Targets) != 3 { // Swamp (deduped), Sol Ring, Dark Ritual
		t.Errorf("targets=%v, want 3 unique", result[0].Targets)
	}

	// Second should be the cast (not coalesced)
	if result[1].Kind != "cast" {
		t.Errorf("second entry kind=%q, want cast", result[1].Kind)
	}
}

func TestCoalesceEntries_DifferentSeatsNotMerged(t *testing.T) {
	entries := []LogEntry{
		{Turn: 3, Seat: 0, Action: "A UNTAPS X", Kind: "untap", Source: "X", Amount: 1},
		{Turn: 3, Seat: 1, Action: "B UNTAPS Y", Kind: "untap", Source: "Y", Amount: 1},
	}

	result := coalesceEntries(entries)

	if len(result) != 2 {
		t.Fatalf("different seats should not merge, got %d entries", len(result))
	}
}

func TestDedupEntries_CastThenETB(t *testing.T) {
	entries := []LogEntry{
		{Turn: 5, Seat: 0, Kind: "cast", Source: "Rhystic Study", Action: "TINYBONES CASTS RHYSTIC STUDY"},
		{Turn: 5, Seat: 0, Kind: "etb", Source: "Rhystic Study", Action: "TINYBONES → ETB: RHYSTIC STUDY"},
		{Turn: 5, Seat: 0, Kind: "cast", Source: "Sol Ring", Action: "TINYBONES CASTS SOL RING"},
	}

	result := dedupEntries(entries)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries (ETB deduped), got %d", len(result))
	}
	if result[0].Source != "Rhystic Study" || result[0].Kind != "cast" {
		t.Errorf("first should be cast Rhystic Study, got %+v", result[0])
	}
	if result[1].Source != "Sol Ring" || result[1].Kind != "cast" {
		t.Errorf("second should be cast Sol Ring, got %+v", result[1])
	}
}

func TestDedupEntries_ETBWithoutCastKept(t *testing.T) {
	entries := []LogEntry{
		{Turn: 5, Seat: 0, Kind: "etb", Source: "Dryad Arbor", Action: "TINYBONES → ETB: DRYAD ARBOR"},
		{Turn: 5, Seat: 0, Kind: "cast", Source: "Sol Ring", Action: "TINYBONES CASTS SOL RING"},
	}

	result := dedupEntries(entries)

	if len(result) != 2 {
		t.Fatalf("ETB without preceding cast should be kept, got %d", len(result))
	}
}

func TestCoalesceEntries_SingleEntryPassthrough(t *testing.T) {
	entries := []LogEntry{
		{Turn: 1, Seat: 0, Action: "TINYBONES CASTS SOL RING", Kind: "cast"},
	}
	result := coalesceEntries(entries)
	if len(result) != 1 {
		t.Fatalf("single entry should pass through, got %d", len(result))
	}
}
