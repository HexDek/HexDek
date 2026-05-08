package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// TraceEntry tests
// ---------------------------------------------------------------------------

func TestTraceEntryString(t *testing.T) {
	e := TraceEntry{
		Step:   1,
		Phase:  PhaseSetup,
		Card:   "Sol Ring",
		Detail: "2-seat game, Sol Ring on seat-0 battlefield",
	}
	s := e.String()
	if !strings.Contains(s, "[1]") {
		t.Errorf("expected step [1], got: %s", s)
	}
	if !strings.Contains(s, "SETUP:") {
		t.Errorf("expected SETUP phase label, got: %s", s)
	}
	if !strings.Contains(s, "Sol Ring on seat-0") {
		t.Errorf("expected detail in output, got: %s", s)
	}
}

func TestTraceEntryStringWithHash(t *testing.T) {
	e := TraceEntry{
		Step:      3,
		Phase:     PhaseSnapshotDiff,
		Card:      "Lightning Bolt",
		Detail:    "seat1.life 40→37",
		StateHash: "abc123",
	}
	s := e.String()
	if !strings.Contains(s, "[abc123]") {
		t.Errorf("expected state hash in brackets, got: %s", s)
	}
}

func TestTraceEntryStringWithTimestamp(t *testing.T) {
	now := time.Date(2026, 5, 7, 14, 30, 0, 0, time.UTC)
	e := TraceEntry{
		Step:      1,
		Phase:     PhaseSetup,
		Detail:    "test",
		Timestamp: now,
	}
	s := e.String()
	if !strings.Contains(s, "@14:30:00.000") {
		t.Errorf("expected timestamp, got: %s", s)
	}
}

func TestTraceEntryStringNoTimestamp(t *testing.T) {
	e := TraceEntry{
		Step:   1,
		Phase:  PhaseSetup,
		Detail: "test",
	}
	s := e.String()
	if strings.Contains(s, "@") {
		t.Errorf("expected no timestamp when zero, got: %s", s)
	}
}

// ---------------------------------------------------------------------------
// CardTrace tests
// ---------------------------------------------------------------------------

func TestCardTraceAddEntry(t *testing.T) {
	ct := &CardTrace{
		CardName:    "Sol Ring",
		Interaction: "destroy",
		Entries:     make([]TraceEntry, 0),
	}
	ct.AddEntry(PhaseSetup, "placed on battlefield")
	ct.AddEntry(PhaseHandlerEntry, "DestroyPermanent")
	ct.AddEntry(PhaseResolution, "moved to graveyard")

	if len(ct.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(ct.Entries))
	}
	if ct.Entries[0].Step != 1 {
		t.Errorf("first entry step should be 1, got %d", ct.Entries[0].Step)
	}
	if ct.Entries[1].Step != 2 {
		t.Errorf("second entry step should be 2, got %d", ct.Entries[1].Step)
	}
	if ct.Entries[2].Step != 3 {
		t.Errorf("third entry step should be 3, got %d", ct.Entries[2].Step)
	}
	if ct.Entries[0].Card != "Sol Ring" {
		t.Errorf("entry card should be Sol Ring, got %s", ct.Entries[0].Card)
	}
}

func TestCardTraceAddEntryWithHash(t *testing.T) {
	ct := &CardTrace{
		CardName:    "Mana Crypt",
		Interaction: "exile",
		Entries:     make([]TraceEntry, 0),
	}
	ct.AddEntryWithHash(PhaseSnapshotDiff, "bf changed", "deadbeef")

	if len(ct.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(ct.Entries))
	}
	if ct.Entries[0].StateHash != "deadbeef" {
		t.Errorf("expected hash deadbeef, got %s", ct.Entries[0].StateHash)
	}
}

func TestCardTraceFailed(t *testing.T) {
	cases := []struct {
		outcome TraceOutcome
		want    bool
	}{
		{OutcomePass, false},
		{OutcomeFailNoChange, true},
		{OutcomeFailPanic, true},
		{OutcomeFailInvariant, true},
	}
	for _, tc := range cases {
		ct := &CardTrace{Outcome: tc.outcome}
		if ct.Failed() != tc.want {
			t.Errorf("outcome %s: Failed()=%v, want %v", tc.outcome, ct.Failed(), tc.want)
		}
	}
}

func TestCardTraceRender(t *testing.T) {
	ct := &CardTrace{
		CardName:    "Sol Ring",
		Interaction: "destroy",
		Entries:     make([]TraceEntry, 0),
		Outcome:     OutcomePass,
		Duration:    42 * time.Microsecond,
	}
	ct.AddEntry(PhaseSetup, "2-seat game, Sol Ring on seat-0 battlefield")
	ct.AddEntry(PhaseSetup, "snapshot taken (life=40/40, bf=1/0)")
	ct.AddEntry(PhaseEffectDispatch, "stock AST dispatch: Destroy effect")
	ct.AddEntry(PhaseResolution, "Sol Ring moved battlefield→graveyard")

	out := ct.Render()

	// Check header.
	if !strings.Contains(out, "=== Sol Ring × destroy ===") {
		t.Errorf("missing header in render:\n%s", out)
	}
	// Check entries are numbered.
	if !strings.Contains(out, "[1] SETUP:") {
		t.Errorf("missing step [1] in render:\n%s", out)
	}
	if !strings.Contains(out, "[4] RESOLUTION:") {
		t.Errorf("missing step [4] in render:\n%s", out)
	}
	// Check outcome.
	if !strings.Contains(out, "OUTCOME: PASS") {
		t.Errorf("missing PASS outcome in render:\n%s", out)
	}
	// Check duration.
	if !strings.Contains(out, "42µs") {
		t.Errorf("missing duration in render:\n%s", out)
	}
}

func TestCardTraceRenderFailure(t *testing.T) {
	ct := &CardTrace{
		CardName:     "Mana Crypt",
		Interaction:  "goldilocks",
		Entries:      make([]TraceEntry, 0),
		Outcome:      OutcomeFailNoChange,
		FailurePoint: "effect resolved but state unchanged",
	}
	ct.AddEntry(PhaseSetup, "placed Mana Crypt")
	ct.AddEntry(PhaseEffectDispatch, "tap for mana")
	ct.AddEntry(PhaseSnapshotDiff, "no change")

	out := ct.Render()
	if !strings.Contains(out, "OUTCOME: FAIL_NO_CHANGE") {
		t.Errorf("missing failure outcome:\n%s", out)
	}
	if !strings.Contains(out, "effect resolved but state unchanged") {
		t.Errorf("missing failure point:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// CardTraceRecorder tests
// ---------------------------------------------------------------------------

func TestCardTraceRecorderNilSafe(t *testing.T) {
	// All methods should be no-ops on nil receiver.
	var r *CardTraceRecorder
	r.Setup("test")
	r.Snapshot("test")
	r.SnapshotWithHash("test", "hash")
	r.ConditionCheck("test")
	r.HandlerEntry("test")
	r.EffectDispatch("test")
	r.Resolution("test")
	r.SnapshotDiff("test")
	r.Assert("test")
	r.Panic("test")
	r.Complete(OutcomePass, "")

	if r.Trace() != nil {
		t.Errorf("nil recorder should return nil trace")
	}
}

func TestCardTraceRecorderFlow(t *testing.T) {
	rec := newCardTraceRecorder("Sol Ring", "destroy")
	rec.Setup("2-seat game, Sol Ring on seat-0 battlefield")
	rec.Snapshot("life=40/40, bf=1/0")
	rec.ConditionCheck("indestructible → false")
	rec.HandlerEntry("DestroyPermanent")
	rec.EffectDispatch("move battlefield→graveyard")
	rec.Resolution("Sol Ring moved to graveyard")
	rec.SnapshotDiff("bf=0/0, gy=1/0")
	rec.Assert("no invariant violations")
	rec.Complete(OutcomePass, "")

	tr := rec.Trace()
	if tr == nil {
		t.Fatal("trace should not be nil")
	}
	if tr.CardName != "Sol Ring" {
		t.Errorf("card name: want Sol Ring, got %s", tr.CardName)
	}
	if tr.Interaction != "destroy" {
		t.Errorf("interaction: want destroy, got %s", tr.Interaction)
	}
	if len(tr.Entries) != 8 {
		t.Errorf("expected 8 entries, got %d", len(tr.Entries))
	}
	if tr.Outcome != OutcomePass {
		t.Errorf("outcome: want pass, got %s", tr.Outcome)
	}
	if tr.Duration <= 0 {
		t.Errorf("duration should be positive, got %s", tr.Duration)
	}
}

func TestCardTraceRecorderPanic(t *testing.T) {
	rec := newCardTraceRecorder("Chaos Orb", "flicker")
	rec.Setup("placed Chaos Orb")
	rec.Panic("nil pointer dereference")
	rec.Complete(OutcomeFailPanic, "handler panicked in FlickerPermanent")

	tr := rec.Trace()
	if tr.Outcome != OutcomeFailPanic {
		t.Errorf("expected fail_panic, got %s", tr.Outcome)
	}
	if len(tr.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(tr.Entries))
	}
	if tr.Entries[1].Phase != PhasePanic {
		t.Errorf("expected panic phase, got %s", tr.Entries[1].Phase)
	}
}

// ---------------------------------------------------------------------------
// TraceCollector tests
// ---------------------------------------------------------------------------

func TestTraceCollectorNilSafe(t *testing.T) {
	var tc *TraceCollector
	rec := tc.Begin("Sol Ring", "destroy")
	if rec != nil {
		t.Errorf("nil collector should return nil recorder")
	}
	tc.Collect(nil)
	c, w, s := tc.Stats()
	if c != 0 || w != 0 || s != 0 {
		t.Errorf("nil collector stats should be 0/0/0")
	}
	if err := tc.WriteSummary(); err != nil {
		t.Errorf("nil collector WriteSummary should return nil")
	}
}

func TestTraceCollectorEmptyDir(t *testing.T) {
	tc := NewTraceCollector("", false)
	if tc != nil {
		t.Errorf("empty dir should return nil collector")
	}
}

func TestTraceCollectorCollect(t *testing.T) {
	dir := t.TempDir()
	tc := NewTraceCollector(dir, false)

	rec := tc.Begin("Sol Ring", "destroy")
	rec.Setup("test setup")
	rec.Complete(OutcomePass, "")
	tc.Collect(rec.Trace())

	c, w, _ := tc.Stats()
	if c != 1 {
		t.Errorf("expected 1 collected, got %d", c)
	}
	if w != 1 {
		t.Errorf("expected 1 written, got %d", w)
	}

	// Verify file was written.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	traceFiles := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".trace") {
			traceFiles++
		}
	}
	if traceFiles != 1 {
		t.Errorf("expected 1 trace file, got %d", traceFiles)
	}
}

func TestTraceCollectorFailuresOnly(t *testing.T) {
	dir := t.TempDir()
	tc := NewTraceCollector(dir, true) // failures only

	// Passing test — should be skipped.
	rec1 := tc.Begin("Sol Ring", "destroy")
	rec1.Complete(OutcomePass, "")
	tc.Collect(rec1.Trace())

	// Failing test — should be written.
	rec2 := tc.Begin("Black Lotus", "exile")
	rec2.Complete(OutcomeFailInvariant, "zone count mismatch")
	tc.Collect(rec2.Trace())

	c, w, s := tc.Stats()
	if c != 1 {
		t.Errorf("collected: want 1 (only failure), got %d", c)
	}
	if w != 1 {
		t.Errorf("written: want 1, got %d", w)
	}
	if s != 1 {
		t.Errorf("skipped: want 1, got %d", s)
	}
}

func TestTraceCollectorFileContent(t *testing.T) {
	dir := t.TempDir()
	tc := NewTraceCollector(dir, false)

	rec := tc.Begin("Lightning Bolt", "spell_resolve")
	rec.Setup("instant on stack")
	rec.HandlerEntry("ResolveSpell")
	rec.EffectDispatch("deal 3 damage to target creature")
	rec.Resolution("opponent bear 2/2 → marked damage 3")
	rec.SnapshotDiff("seat1.bf 1→0, seat1.gy 0→1")
	rec.Complete(OutcomePass, "")
	tc.Collect(rec.Trace())

	// Find and read the trace file.
	entries, _ := os.ReadDir(dir)
	var content []byte
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".trace") {
			content, _ = os.ReadFile(filepath.Join(dir, e.Name()))
			break
		}
	}
	if len(content) == 0 {
		t.Fatal("no trace file found or empty")
	}

	body := string(content)
	if !strings.Contains(body, "=== Lightning Bolt × spell_resolve ===") {
		t.Errorf("missing header in file")
	}
	if !strings.Contains(body, "OUTCOME: PASS") {
		t.Errorf("missing outcome in file")
	}
	if !strings.Contains(body, "HANDLER_ENTRY:") {
		t.Errorf("missing handler entry phase")
	}
}

func TestTraceCollectorWriteSummary(t *testing.T) {
	dir := t.TempDir()
	tc := NewTraceCollector(dir, false)

	// Add a pass and a fail.
	rec1 := tc.Begin("Sol Ring", "destroy")
	rec1.Complete(OutcomePass, "")
	tc.Collect(rec1.Trace())

	rec2 := tc.Begin("Mox Pearl", "sacrifice")
	rec2.Complete(OutcomeFailNoChange, "effect was dead code")
	tc.Collect(rec2.Trace())

	if err := tc.WriteSummary(); err != nil {
		t.Fatal(err)
	}

	summaryPath := filepath.Join(dir, "SUMMARY.trace")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("failed to read summary: %v", err)
	}

	body := string(data)
	if !strings.Contains(body, "THOR TRACE SUMMARY") {
		t.Errorf("missing summary header")
	}
	if !strings.Contains(body, "Total traces: 2") {
		t.Errorf("missing total count")
	}
	if !strings.Contains(body, "Sol Ring") {
		t.Errorf("missing Sol Ring in summary")
	}
	if !strings.Contains(body, "Mox Pearl") {
		t.Errorf("missing Mox Pearl in summary")
	}
}

func TestTraceCollectorConcurrent(t *testing.T) {
	dir := t.TempDir()
	tc := NewTraceCollector(dir, false)

	const N = 50
	done := make(chan struct{})
	for i := 0; i < N; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			rec := tc.Begin(
				strings.Repeat("x", 5), // short card name
				"destroy",
			)
			rec.Setup("test")
			rec.Complete(OutcomePass, "")
			tc.Collect(rec.Trace())
		}(i)
	}
	for i := 0; i < N; i++ {
		<-done
	}

	c, w, _ := tc.Stats()
	if c != N {
		t.Errorf("collected: want %d, got %d", N, c)
	}
	if w != N {
		t.Errorf("written: want %d, got %d", N, w)
	}
}

// ---------------------------------------------------------------------------
// FormatSnapshotSummary / FormatSnapshotDiffSummary tests
// ---------------------------------------------------------------------------

func TestFormatSnapshotSummary(t *testing.T) {
	snap := goldilocksSnapshot{}
	snap.life = [4]int{40, 40, 0, 0}
	snap.battlefieldCnt = [4]int{3, 1, 0, 0}
	snap.handSize = [4]int{7, 7, 0, 0}
	snap.graveyardSize = [4]int{0, 0, 0, 0}

	s := FormatSnapshotSummary(snap, 2)
	if !strings.Contains(s, "life=40/40") {
		t.Errorf("expected life=40/40 in summary, got: %s", s)
	}
	if !strings.Contains(s, "bf=3/1") {
		t.Errorf("expected bf=3/1 in summary, got: %s", s)
	}
	if !strings.Contains(s, "hand=7/7") {
		t.Errorf("expected hand=7/7 in summary, got: %s", s)
	}
}

func TestFormatSnapshotSummaryClampSeats(t *testing.T) {
	snap := goldilocksSnapshot{}
	// Invalid seat counts should be clamped to 2.
	s := FormatSnapshotSummary(snap, 0)
	if !strings.Contains(s, "life=0/0") {
		t.Errorf("expected 2-seat default, got: %s", s)
	}
	s = FormatSnapshotSummary(snap, 99)
	if !strings.Contains(s, "life=0/0") {
		t.Errorf("expected 2-seat default for overflow, got: %s", s)
	}
}

func TestFormatSnapshotDiffSummary(t *testing.T) {
	before := goldilocksSnapshot{}
	before.life = [4]int{40, 40, 0, 0}
	before.battlefieldCnt = [4]int{1, 1, 0, 0}

	after := goldilocksSnapshot{}
	after.life = [4]int{40, 37, 0, 0}
	after.battlefieldCnt = [4]int{1, 0, 0, 0}

	s := FormatSnapshotDiffSummary(before, after, 2)
	if !strings.Contains(s, "seat1.life 40→37") {
		t.Errorf("expected life diff, got: %s", s)
	}
	if !strings.Contains(s, "seat1.bf 1→0") {
		t.Errorf("expected bf diff, got: %s", s)
	}
	// seat0 should NOT appear since it didn't change.
	if strings.Contains(s, "seat0") {
		t.Errorf("seat0 should not appear (unchanged), got: %s", s)
	}
}

func TestFormatSnapshotDiffSummaryNoChange(t *testing.T) {
	snap := goldilocksSnapshot{}
	snap.life = [4]int{40, 40, 0, 0}
	s := FormatSnapshotDiffSummary(snap, snap, 2)
	if s != "no change" {
		t.Errorf("expected 'no change', got: %s", s)
	}
}

func TestFormatSnapshotDiffSummaryAllFields(t *testing.T) {
	before := goldilocksSnapshot{}
	after := goldilocksSnapshot{}

	// Set up changes in all tracked fields for seat 0.
	before.life[0] = 40
	after.life[0] = 35
	before.battlefieldCnt[0] = 2
	after.battlefieldCnt[0] = 1
	before.handSize[0] = 7
	after.handSize[0] = 8
	before.graveyardSize[0] = 0
	after.graveyardSize[0] = 1
	before.exileSize[0] = 0
	after.exileSize[0] = 1
	before.libSize[0] = 50
	after.libSize[0] = 49
	before.manaPool[0] = 0
	after.manaPool[0] = 3
	before.poisonCounters[0] = 0
	after.poisonCounters[0] = 2
	before.energy[0] = 0
	after.energy[0] = 4
	before.stackSize = 0
	after.stackSize = 1

	s := FormatSnapshotDiffSummary(before, after, 1)
	for _, want := range []string{
		"seat0.life 40→35",
		"seat0.bf 2→1",
		"seat0.hand 7→8",
		"seat0.gy 0→1",
		"seat0.exile 0→1",
		"seat0.lib 50→49",
		"seat0.mana 0→3",
		"seat0.poison 0→2",
		"seat0.energy 0→4",
		"stack 0→1",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in diff: %s", want, s)
		}
	}
}

// ---------------------------------------------------------------------------
// Phase constant coverage
// ---------------------------------------------------------------------------

func TestTracePhaseValues(t *testing.T) {
	phases := []TracePhase{
		PhaseSetup, PhaseConditionCheck, PhaseHandlerEntry,
		PhaseEffectDispatch, PhaseResolution, PhaseSnapshotDiff,
		PhaseAssert, PhasePanic,
	}
	seen := map[TracePhase]bool{}
	for _, p := range phases {
		if seen[p] {
			t.Errorf("duplicate phase: %s", p)
		}
		seen[p] = true
		if string(p) == "" {
			t.Errorf("phase should not be empty string")
		}
	}
}

func TestTraceOutcomeValues(t *testing.T) {
	outcomes := []TraceOutcome{
		OutcomePass, OutcomeFailNoChange, OutcomeFailPanic, OutcomeFailInvariant,
	}
	seen := map[TraceOutcome]bool{}
	for _, o := range outcomes {
		if seen[o] {
			t.Errorf("duplicate outcome: %s", o)
		}
		seen[o] = true
	}
}
