package muninn

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBatcher_FlushOnClose(t *testing.T) {
	dir := t.TempDir()
	b := NewBatcher(BatcherConfig{Dir: dir, BatchSize: 1000, FlushInterval: time.Hour})

	b.AddParserGaps(map[string]int{"snippet-a": 3, "snippet-b": 2})
	b.AddCrash("panic: boom", []string{"Tinybones"}, 100, 4)
	b.AddDeadTrigger("triggered_ability", "Rhystic Study", 5, 1)
	b.AddConcessions([]ConcessionRecord{{Commander: "Atraxa", Turn: 7, Life: 12}})

	// Nothing should be on disk yet (interval is 1h, batch size 1000).
	if _, err := os.Stat(filepath.Join(dir, parserGapsFile)); !os.IsNotExist(err) {
		t.Errorf("expected no parser_gaps.json before close, got err=%v", err)
	}

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	gaps, err := ReadParserGaps(dir)
	if err != nil {
		t.Fatalf("ReadParserGaps: %v", err)
	}
	if len(gaps) != 2 {
		t.Errorf("expected 2 parser gaps, got %d", len(gaps))
	}

	crashes, err := ReadCrashLogs(dir)
	if err != nil {
		t.Fatalf("ReadCrashLogs: %v", err)
	}
	if len(crashes) != 1 || crashes[0].StackTrace != "panic: boom" {
		t.Errorf("crashes: got %+v", crashes)
	}

	dead, err := ReadDeadTriggers(dir)
	if err != nil {
		t.Fatalf("ReadDeadTriggers: %v", err)
	}
	if len(dead) != 1 || dead[0].Count != 5 || dead[0].GamesSeen != 1 {
		t.Errorf("dead triggers: got %+v", dead)
	}

	cons, err := ReadConcessions(dir)
	if err != nil {
		t.Fatalf("ReadConcessions: %v", err)
	}
	if len(cons) != 1 || cons[0].Commander != "Atraxa" {
		t.Errorf("concessions: got %+v", cons)
	}
}

func TestBatcher_FlushOnBatchSize(t *testing.T) {
	dir := t.TempDir()
	b := NewBatcher(BatcherConfig{Dir: dir, BatchSize: 3, FlushInterval: time.Hour})
	t.Cleanup(func() { _ = b.Close() })

	// Add 3 parser-gap entries; the 3rd should trigger a synchronous flush.
	b.AddParserGaps(map[string]int{"a": 1})
	b.AddParserGaps(map[string]int{"b": 1})

	// Before threshold, file should not exist.
	if _, err := os.Stat(filepath.Join(dir, parserGapsFile)); !os.IsNotExist(err) {
		t.Errorf("expected no parser_gaps.json before threshold, got err=%v", err)
	}

	b.AddParserGaps(map[string]int{"c": 1})

	// Synchronous flush from inside Add — file should exist now.
	gaps, err := ReadParserGaps(dir)
	if err != nil {
		t.Fatalf("ReadParserGaps: %v", err)
	}
	if len(gaps) != 3 {
		t.Errorf("expected 3 parser gaps after batch flush, got %d (%+v)", len(gaps), gaps)
	}
}

func TestBatcher_FlushOnInterval(t *testing.T) {
	dir := t.TempDir()
	b := NewBatcher(BatcherConfig{Dir: dir, BatchSize: 1000, FlushInterval: 50 * time.Millisecond})
	t.Cleanup(func() { _ = b.Close() })

	b.AddParserGaps(map[string]int{"interval-snippet": 1})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		gaps, err := ReadParserGaps(dir)
		if err != nil {
			t.Fatalf("ReadParserGaps: %v", err)
		}
		if len(gaps) == 1 && gaps[0].Snippet == "interval-snippet" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("interval-driven flush did not occur within 2s")
}

func TestBatcher_DeadTriggerMerge(t *testing.T) {
	dir := t.TempDir()
	b := NewBatcher(BatcherConfig{Dir: dir, BatchSize: 1000, FlushInterval: time.Hour})

	// Two pre-existing entries — verify merge into existing file.
	if err := PersistDeadTriggersRaw(dir, []DeadTrigger{
		{TriggerName: "triggered_ability", CardName: "Rhystic Study", Count: 10, GamesSeen: 5},
	}); err != nil {
		t.Fatalf("seed dead_triggers: %v", err)
	}

	// Same key + new key, repeated calls accumulate.
	b.AddDeadTrigger("triggered_ability", "Rhystic Study", 1, 1)
	b.AddDeadTrigger("triggered_ability", "Rhystic Study", 1, 1)
	b.AddDeadTrigger("triggered_ability", "Smothering Tithe", 3, 1)

	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got, err := ReadDeadTriggers(dir)
	if err != nil {
		t.Fatalf("ReadDeadTriggers: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d (%+v)", len(got), got)
	}
	byCard := map[string]DeadTrigger{}
	for _, dt := range got {
		byCard[dt.CardName] = dt
	}
	if r := byCard["Rhystic Study"]; r.Count != 12 || r.GamesSeen != 7 {
		t.Errorf("Rhystic Study: got Count=%d GamesSeen=%d, want 12/7", r.Count, r.GamesSeen)
	}
	if s := byCard["Smothering Tithe"]; s.Count != 3 || s.GamesSeen != 1 {
		t.Errorf("Smothering Tithe: got Count=%d GamesSeen=%d, want 3/1", s.Count, s.GamesSeen)
	}
}

func TestBatcher_CloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	b := NewBatcher(BatcherConfig{Dir: dir, BatchSize: 1000, FlushInterval: time.Hour})
	if err := b.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestBatcher_EmptyFlushNoOp(t *testing.T) {
	dir := t.TempDir()
	b := NewBatcher(BatcherConfig{Dir: dir, BatchSize: 1000, FlushInterval: time.Hour})
	if err := b.Flush(); err != nil {
		t.Fatalf("Flush on empty: %v", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// No files should have been created.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir, got %d entries", len(entries))
	}
}
