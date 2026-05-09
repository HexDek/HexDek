package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// makeCLITestCard returns a creature oracleCard wired with whatever
// oracle text the caller supplies. Designed to drive testCard /
// testInteraction without touching the real corpus loader.
func makeCLITestCard(name, oracle string) *oracleCard {
	return &oracleCard{
		Name:       name,
		TypeLine:   "Creature — Bear",
		Types:      []string{"creature"},
		Colors:     []string{"G"},
		CMC:        2,
		Power:      2,
		Toughness:  2,
		OracleText: oracle,
	}
}

// TestThorFeaturesZeroValueIsBackwardCompatible documents the contract
// for the zero-value features struct: traces off, scaffold off, and —
// the one trap to watch out for — opponent-detect off. main() flips
// OpponentDetect back on via the CLI default, which is what preserves
// pre-2.0 behaviour for invocations with no flags.
func TestThorFeaturesZeroValueIsBackwardCompatible(t *testing.T) {
	var f thorFeatures
	if f.Trace || f.TraceFailuresOnly || f.OpponentDetect || f.Scaffold {
		t.Errorf("zero-value features should be all-false, got %+v", f)
	}
}

// TestTraceCollectorRespectsFailuresOnlyToggle verifies the new
// --trace-failures-only flag's user-facing semantics: with it true,
// passing traces are skipped; with it false, all traces are kept.
// The wiring lives in main() but the behaviour is on TraceCollector,
// so we test it there.
func TestTraceCollectorRespectsFailuresOnlyToggle(t *testing.T) {
	for _, tc := range []struct {
		name         string
		failuresOnly bool
		outcome      TraceOutcome
		wantWritten  int
		wantSkipped  int
	}{
		{"failuresOnly skips passes", true, OutcomePass, 0, 1},
		{"failuresOnly keeps failures", true, OutcomeFailNoChange, 1, 0},
		{"!failuresOnly keeps passes", false, OutcomePass, 1, 0},
		{"!failuresOnly keeps failures", false, OutcomeFailInvariant, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			coll := NewTraceCollector(dir, tc.failuresOnly)
			rec := coll.Begin("Sol Ring", "destroy")
			rec.Setup("placed")
			rec.Complete(tc.outcome, "")
			coll.Collect(rec.Trace())

			_, written, skipped := coll.Stats()
			if int(written) != tc.wantWritten {
				t.Errorf("written = %d, want %d", written, tc.wantWritten)
			}
			if int(skipped) != tc.wantSkipped {
				t.Errorf("skipped = %d, want %d", skipped, tc.wantSkipped)
			}
		})
	}
}

// TestTraceCollectorWritesFilesToDir is the file-level proof that the
// --trace-dir flag actually drops files where the user asked. A
// regression here would mean the user ran with --trace and got
// silence on disk despite "X collected, Y written" in the log.
func TestTraceCollectorWritesFilesToDir(t *testing.T) {
	dir := t.TempDir()
	coll := NewTraceCollector(dir, false)
	rec := coll.Begin("Lightning Bolt", "damage_3")
	rec.EffectDispatch("damage 3")
	rec.Complete(OutcomePass, "")
	coll.Collect(rec.Trace())

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	var traceFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".trace") {
			traceFiles = append(traceFiles, e.Name())
		}
	}
	if len(traceFiles) != 1 {
		t.Fatalf("expected 1 .trace file in %s, got %v", dir, traceFiles)
	}

	body, err := os.ReadFile(filepath.Join(dir, traceFiles[0]))
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	if !strings.Contains(string(body), "Lightning Bolt") {
		t.Errorf("trace file missing card name; body:\n%s", string(body))
	}
}

// TestOpponentDetectFlagGatesEnrichment is the behavioural test for
// the new --opponent-detect flag. With it on, EnrichOpponentSeat
// populates seat 1 with a creature target when the oracle text
// requests one. With it off, that path is skipped and the seat is
// left with only the baseline vanilla Opponent Bear that
// makeGameState seeds. The test asserts the on-path adds creatures
// past the off-path baseline — the flag is "add things", not
// "remove the seed".
func TestOpponentDetectFlagGatesEnrichment(t *testing.T) {
	oc := makeCLITestCard("Targeter",
		"Destroy target creature an opponent controls.")

	// ON path mirrors what testInteraction does inside the
	// feat.OpponentDetect branch.
	gsOn := makeGameState(oc, nil)
	feat := thorFeatures{OpponentDetect: true}
	if feat.OpponentDetect {
		req := DetectOpponentRequirements(oc.OracleText)
		if !req.HasAny() {
			t.Fatal("DetectOpponentRequirements should have flagged the oracle text")
		}
		EnrichOpponentSeat(gsOn, 1, req)
	}
	creaturesOn := countCreaturesOnSeat(gsOn, 1)

	// OFF path — no enrichment call.
	gsOff := makeGameState(oc, nil)
	creaturesOff := countCreaturesOnSeat(gsOff, 1)

	if creaturesOn <= creaturesOff {
		t.Errorf("opponent-detect ON should add creatures past baseline (on=%d off=%d)",
			creaturesOn, creaturesOff)
	}
}

// TestTestInteractionBackwardCompatNoFlags is the regression guard
// for the headline backwards-compat claim: "existing test runs still
// work without flags." We construct the same feat the CLI builds
// with no flags supplied (i.e. opponent-detect on, the rest off),
// drive a single interaction, and assert it doesn't panic. We do
// not assert the failure result either way — the point is the
// pipeline executes without aborting on the new code paths.
func TestTestInteractionBackwardCompatNoFlags(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("backward-compat path panic'd: %v", r)
		}
	}()
	oc := makeCLITestCard("Test Bear", "")
	inter := interaction{
		Name: "noop",
		Fn:   func(_ *gameengine.GameState, _ *gameengine.Permanent) {},
	}
	feat := thorFeatures{OpponentDetect: true} // CLI default
	_ = testInteraction(oc, nil, inter, nil /* nil collector */, feat)
}

// TestTestInteractionAllFlagsOn drives the same path with every flag
// enabled to make sure the new branches don't blow up either.
func TestTestInteractionAllFlagsOn(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("all-flags-on path panic'd: %v", r)
		}
	}()
	oc := makeCLITestCard("Test Bear", "Destroy target creature an opponent controls.")
	inter := interaction{
		Name: "noop",
		Fn:   func(_ *gameengine.GameState, _ *gameengine.Permanent) {},
	}

	dir := t.TempDir()
	coll := NewTraceCollector(dir, false)
	feat := thorFeatures{
		Trace:             true,
		TraceFailuresOnly: false,
		OpponentDetect:    true,
		Scaffold:          true,
	}
	_ = testInteraction(oc, nil, inter, coll, feat)
}

// countCreaturesOnSeat reports how many battlefield permanents on
// the seat are creatures. Used to gauge whether opponent-detect has
// enriched past the bare baseline.
func countCreaturesOnSeat(gs *gameengine.GameState, seat int) int {
	if gs == nil || seat < 0 || seat >= len(gs.Seats) {
		return 0
	}
	n := 0
	for _, p := range gs.Seats[seat].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		for _, t := range p.Card.Types {
			if t == "creature" {
				n++
				break
			}
		}
	}
	return n
}
