package db

import (
	"context"
	"testing"
)

func TestImportLogRoundtrip(t *testing.T) {
	d, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer d.Close()
	ctx := context.Background()

	for _, e := range []ImportLogEntry{
		{Owner: "josh", DeckKey: "josh/brudiclad", DeckName: "Brudiclad", Commander: "Brudiclad, Telchor Engineer", Source: "paste", CardCount: 99, ImportedAt: 100},
		{Owner: "josh", DeckKey: "josh/sisay", DeckName: "Sisay", Commander: "Sisay, Weatherlight Captain", Source: "moxfield", SourceURL: "https://moxfield.com/decks/abc", CardCount: 100, ImportedAt: 200},
		{Owner: "alice", DeckKey: "alice/jarad", DeckName: "Jarad", Source: "paste", CardCount: 100, ImportedAt: 150},
	} {
		if _, err := InsertImportLog(ctx, d, e); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	got, err := ListImportLogs(ctx, d, "josh", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries for josh, got %d", len(got))
	}
	// Newest first.
	if got[0].DeckKey != "josh/sisay" {
		t.Fatalf("expected sisay first (newer), got %q", got[0].DeckKey)
	}
	if got[0].Source != "moxfield" || got[0].SourceURL == "" {
		t.Fatalf("moxfield source/url not preserved: %+v", got[0])
	}
	if got[1].DeckKey != "josh/brudiclad" {
		t.Fatalf("expected brudiclad second, got %q", got[1].DeckKey)
	}

	// Owner isolation.
	aliceLogs, _ := ListImportLogs(ctx, d, "alice", 10)
	if len(aliceLogs) != 1 || aliceLogs[0].DeckKey != "alice/jarad" {
		t.Fatalf("owner isolation broken: %+v", aliceLogs)
	}
}
