package tournament

import (
	"fmt"
	"os"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
	"github.com/hexdek/hexdek/internal/hat"
)

// TestConvictionHarvest_R38 runs a focused tournament with the
// conviction telemetry ring active and reports per-trigger fire
// counts. This is the empirical-evidence side of the
// conviction-readiness evaluation described in
// docs/conviction-r38-readiness.md.
//
// Procedure:
//   1. Clear the global conviction ring.
//   2. Run a small tournament (4 seats × N games) using the production
//      YggdrasilHat factory. ShouldConcede stays non-acting (returns
//      false); recordConvictionSample writes one diagnostic event to
//      the ring every turn for every seat.
//   3. Snapshot the ring.
//   4. Report fire counts per candidate trigger, broken down by:
//      - total samples
//      - samples on which any trigger fired
//      - samples per individual trigger (score-window, win-line-extinct)
//      - unique (seed, seat) pairs that fired each trigger
//
// What this CAN'T yet measure: per-game false-positive rate (did the
// triggered seat actually win this game?). TournamentResult.Analyses
// would carry per-game winners, but the trigger-vs-winner correlation
// needs a join key (seed → winner-seat). The runner currently
// aggregates wins by commander name across all games rather than
// keeping a per-seed map, so this test reports raw fire-rate only and
// surfaces the structural gap that blocks the FP calculation.
//
// Gated behind HEXDEK_CONVICTION_HARVEST=1 (heavy: loads the full AST
// corpus + runs 30+ games). To run:
//
//   HEXDEK_CONVICTION_HARVEST=1 go test ./internal/tournament/ \
//       -run TestConvictionHarvest_R38 -v -count=1 -timeout 600s
func TestConvictionHarvest_R38(t *testing.T) {
	if os.Getenv("HEXDEK_CONVICTION_HARVEST") != "1" {
		t.Skip("set HEXDEK_CONVICTION_HARVEST=1 to enable the harvest run")
	}

	corpus, meta := loadCorpus(t)
	const nSeats = 4
	const nGames = 60
	const baseSeed = int64(38_000)

	paths := findDecks(t, nSeats)
	if len(paths) < nSeats {
		t.Skipf("need at least %d decks, got %d", nSeats, len(paths))
	}
	decks := loadNDecks(t, paths, corpus, meta)
	strategies := make([]*hat.StrategyProfile, nSeats)
	for i := 0; i < nSeats; i++ {
		strategies[i] = hat.LoadStrategyFromFreya(paths[i])
	}

	factories := make([]HatFactory, nSeats)
	for i := 0; i < nSeats; i++ {
		idx := i
		factories[i] = func() gameengine.Hat {
			return hat.NewYggdrasilHat(strategies[idx], 50)
		}
	}

	hat.ResetConvictionTelemetry()

	cfg := TournamentConfig{
		Decks:           decks,
		NSeats:          nSeats,
		NGames:          nGames,
		Seed:            baseSeed,
		HatFactories:    factories,
		Workers:         1,
		CommanderMode:   true,
		MaxTurnsPerGame: 60,
	}

	result, err := Run(cfg)
	if err != nil {
		t.Fatalf("tournament Run: %v", err)
	}
	if result == nil {
		t.Fatal("nil result")
	}

	events, totalSeen := hat.SnapshotConvictionEvents(0, 0)

	type seatGame struct {
		seed int64
		seat int
	}
	scoreSeats := map[seatGame]bool{}
	winlineSeats := map[seatGame]bool{}
	anySeats := map[seatGame]bool{}

	scoreFires := 0
	winlineFires := 0
	anyFires := 0

	for _, e := range events {
		key := seatGame{seed: e.GameSeed, seat: e.Seat}
		if e.ScoreTriggered {
			scoreFires++
			scoreSeats[key] = true
		}
		if e.WinLineExtinct {
			winlineFires++
			winlineSeats[key] = true
		}
		if e.AnyTriggered {
			anyFires++
			anySeats[key] = true
		}
	}

	// Sample → seat density (samples / unique seat-game) tells us the
	// trigger "stickiness" — when score fires, does it fire for many
	// turns of one game (sticky) or one-shot (noisy)?
	scoreStickiness := 0.0
	if len(scoreSeats) > 0 {
		scoreStickiness = float64(scoreFires) / float64(len(scoreSeats))
	}
	winlineStickiness := 0.0
	if len(winlineSeats) > 0 {
		winlineStickiness = float64(winlineFires) / float64(len(winlineSeats))
	}

	totalSeatGames := nGames * nSeats
	scoreSeatRate := float64(len(scoreSeats)) / float64(totalSeatGames)
	winlineSeatRate := float64(len(winlineSeats)) / float64(totalSeatGames)

	t.Logf("=== Conviction Harvest R38 ===")
	t.Logf("games:                 %d", nGames)
	t.Logf("seats:                 %d", nSeats)
	t.Logf("base seed:             %d", baseSeed)
	t.Logf("seat-games (max FP):   %d", totalSeatGames)
	t.Logf("samples in ring:       %d", len(events))
	t.Logf("samples seen ever:     %d", totalSeen)
	t.Logf("avg samples/seat-game: %.1f", float64(len(events))/float64(totalSeatGames))
	t.Logf("")
	t.Logf("Score-Window Trigger (4-turn relpos avg < -0.35, turn ≥ 10):")
	t.Logf("  total fire-samples:    %d", scoreFires)
	t.Logf("  unique seat-games:     %d  (%.1f%% of all seat-games)",
		len(scoreSeats), 100*scoreSeatRate)
	t.Logf("  stickiness (s/sg):     %.1f", scoreStickiness)
	t.Logf("")
	t.Logf("Win-Line-Extinct Trigger (all combo pieces / finishers in exile):")
	t.Logf("  total fire-samples:    %d", winlineFires)
	t.Logf("  unique seat-games:     %d  (%.1f%% of all seat-games)",
		len(winlineSeats), 100*winlineSeatRate)
	t.Logf("  stickiness (s/sg):     %.1f", winlineStickiness)
	t.Logf("")
	t.Logf("ANY trigger fired:")
	t.Logf("  total fire-samples:    %d", anyFires)
	t.Logf("  unique seat-games:     %d  (%.1f%% of all seat-games)",
		len(anySeats), 100*float64(len(anySeats))/float64(totalSeatGames))
	t.Logf("")
	t.Logf("Result.WinsByCommander:")
	for name, wins := range result.WinsByCommander {
		t.Logf("  %s: %d wins", name, wins)
	}

	out := os.Getenv("HEXDEK_CONVICTION_OUT")
	if out == "" {
		out = "/tmp/conviction-r38-harvest.txt"
	}
	f, err := os.Create(out)
	if err != nil {
		t.Logf("warning: could not write harvest file: %v", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "harvest_games=%d\n", nGames)
	fmt.Fprintf(f, "harvest_seats=%d\n", nSeats)
	fmt.Fprintf(f, "harvest_seed=%d\n", baseSeed)
	fmt.Fprintf(f, "harvest_seat_games=%d\n", totalSeatGames)
	fmt.Fprintf(f, "harvest_samples=%d\n", len(events))
	fmt.Fprintf(f, "score_fire_samples=%d\n", scoreFires)
	fmt.Fprintf(f, "score_unique_seat_games=%d\n", len(scoreSeats))
	fmt.Fprintf(f, "score_seat_game_rate=%.4f\n", scoreSeatRate)
	fmt.Fprintf(f, "score_stickiness=%.2f\n", scoreStickiness)
	fmt.Fprintf(f, "winline_fire_samples=%d\n", winlineFires)
	fmt.Fprintf(f, "winline_unique_seat_games=%d\n", len(winlineSeats))
	fmt.Fprintf(f, "winline_seat_game_rate=%.4f\n", winlineSeatRate)
	fmt.Fprintf(f, "winline_stickiness=%.2f\n", winlineStickiness)
	fmt.Fprintf(f, "any_fire_samples=%d\n", anyFires)
	fmt.Fprintf(f, "any_unique_seat_games=%d\n", len(anySeats))
}
