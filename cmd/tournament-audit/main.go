// tournament-audit is a one-shot deep-audit runner for the
// 2026-05-17 audit. Runs a pool tournament against a curated deck
// list and emits:
//
//   - winrate distribution histogram (per-commander + bucketed)
//   - crash records (Loki-style) with stack traces
//   - concession blunder records (life > threshold, early concession)
//   - per-game MissedCombos / MissedFinishers (AI gap proxy)
//
// All structured output is written under --out-dir as JSONL/JSON so
// the audit report can be assembled offline.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"math/rand"

	"github.com/hexdek/hexdek/internal/astload"
	"github.com/hexdek/hexdek/internal/deckparser"
	"github.com/hexdek/hexdek/internal/gameengine"
	"github.com/hexdek/hexdek/internal/hat"
	"github.com/hexdek/hexdek/internal/muninn"
	"github.com/hexdek/hexdek/internal/tournament"
)

type cliFlags struct {
	deckListPath string
	games        int
	sampleGames  int
	workers      int
	seed         int64
	sampleSeed   int64
	outDir       string
	astPath      string
	oraclePath   string
	maxTurns     int
	hatBudget    int
}

type perGameRecord struct {
	GameIdx      int      `json:"game_idx"`
	Turns        int      `json:"turns"`
	EndReason    string   `json:"end_reason"`
	Winner       int      `json:"winner_seat"`
	WinnerCmdr   string   `json:"winner_commander,omitempty"`
	Participants []string `json:"participants"`
	Crashed      bool     `json:"crashed,omitempty"`
	CrashErr     string   `json:"crash_err,omitempty"`
	Concessions  int      `json:"concessions,omitempty"`
}

type sampleGameRecord struct {
	GameIdx          int                       `json:"game_idx"`
	Turns            int                       `json:"turns"`
	EndReason        string                    `json:"end_reason"`
	Participants     []string                  `json:"participants"`
	WinnerCmdr       string                    `json:"winner_commander,omitempty"`
	WinCondition     string                    `json:"win_condition,omitempty"`
	WinningCard      string                    `json:"winning_card,omitempty"`
	MissedCombos     []missedComboRec          `json:"missed_combos,omitempty"`
	MissedFinishers  []missedFinisherRec       `json:"missed_finishers,omitempty"`
	Concessions      []muninn.ConcessionRecord `json:"concessions,omitempty"`
	FirstBlood       int                       `json:"first_blood_turn"`
	StallHitTurnCap  bool                      `json:"stall_hit_turn_cap"`
	StallSurvivors   int                       `json:"stall_survivors"`
	StallCause       string                    `json:"stall_cause,omitempty"`
}

type missedComboRec struct {
	Seat      int      `json:"seat"`
	Commander string   `json:"commander"`
	Turn      int      `json:"turn"`
	ComboName string   `json:"combo_name"`
	Pieces    []string `json:"pieces"`
	WinType   string   `json:"win_type"`
	ManaAvail int      `json:"mana_available"`
}

type missedFinisherRec struct {
	Seat         int    `json:"seat"`
	Commander    string `json:"commander"`
	Turn         int    `json:"turn"`
	FinisherName string `json:"finisher"`
	BoardPower   int    `json:"board_power"`
	OppLifeMin   int    `json:"opp_life_min"`
}

func main() {
	var f cliFlags
	flag.StringVar(&f.deckListPath, "deck-list", "", "file listing one deck path per line")
	flag.IntVar(&f.games, "games", 2000, "number of pool games (main run)")
	flag.IntVar(&f.sampleGames, "sample-games", 50, "number of audited games for blunder spot-check")
	flag.IntVar(&f.workers, "workers", 0, "worker count (0=NumCPU)")
	flag.Int64Var(&f.seed, "seed", 51717, "master RNG seed for main run")
	flag.Int64Var(&f.sampleSeed, "sample-seed", 91317, "master RNG seed for sample run")
	flag.StringVar(&f.outDir, "out-dir", "hexdek/tournament-audit-2026-05-17", "output directory for JSON artifacts")
	flag.StringVar(&f.astPath, "ast", "data/rules/ast_dataset.jsonl", "AST dataset path")
	flag.StringVar(&f.oraclePath, "oracle", "data/rules/oracle-cards.json", "Scryfall oracle path")
	flag.IntVar(&f.maxTurns, "max-turns", 0, "per-game turn cap (0=default)")
	flag.IntVar(&f.hatBudget, "hat-budget", 50, "Yggdrasil hat budget")
	flag.Parse()

	if f.deckListPath == "" {
		log.Fatal("--deck-list required")
	}
	if f.workers == 0 {
		f.workers = runtime.NumCPU()
	}

	if err := os.MkdirAll(f.outDir, 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", f.outDir, err)
	}

	deckPaths, err := readDeckList(f.deckListPath)
	if err != nil {
		log.Fatalf("read deck list: %v", err)
	}
	log.Printf("tournament-audit: %d decks, %d main games, %d sample games, %d workers",
		len(deckPaths), f.games, f.sampleGames, f.workers)

	log.Printf("loading AST corpus ...")
	t0 := time.Now()
	corpus, err := astload.Load(f.astPath)
	if err != nil {
		log.Fatalf("astload: %v", err)
	}
	log.Printf("  %d cards in %s", corpus.Count(), time.Since(t0))
	meta, err := deckparser.LoadMetaFromJSONL(f.astPath)
	if err != nil {
		log.Fatalf("meta: %v", err)
	}
	if f.oraclePath != "" {
		if err := meta.SupplementWithOracleJSON(f.oraclePath); err != nil {
			log.Printf("  oracle supplement: %v (continuing)", err)
		}
	}

	log.Printf("parsing %d decks ...", len(deckPaths))
	t0 = time.Now()
	decks := make([]*deckparser.TournamentDeck, 0, len(deckPaths))
	for _, p := range deckPaths {
		d, err := deckparser.ParseDeckFile(p, corpus, meta)
		if err != nil {
			log.Printf("  skip %s: %v", p, err)
			continue
		}
		decks = append(decks, d)
	}
	log.Printf("  parsed %d decks in %s", len(decks), time.Since(t0))

	if len(decks) < 4 {
		log.Fatalf("need at least 4 decks (got %d)", len(decks))
	}

	hatFactories := buildHatFactories(decks, deckPaths, f.hatBudget)

	// PASS A: main 2000-game pool run.
	log.Printf("=== PASS A: %d-game pool run ===", f.games)
	mainResult := runPool(decks, deckPaths, hatFactories, f.games, f.workers, f.seed, f.maxTurns, false /* analytics */, f.outDir, "main")

	// PASS B: sample run with analytics for blunder spot-check.
	log.Printf("=== PASS B: %d-game sample run (analytics on) ===", f.sampleGames)
	sampleResult := runPool(decks, deckPaths, hatFactories, f.sampleGames, f.workers, f.sampleSeed, f.maxTurns, true /* analytics */, f.outDir, "sample")

	// Emit summary JSON.
	summary := map[string]any{
		"date":         "2026-05-17",
		"deck_count":   len(decks),
		"main_run":     mainResult.summary(),
		"sample_run":   sampleResult.summary(),
		"top_decks":    deckPathSnippets(deckPaths, 100),
	}
	writeJSON(filepath.Join(f.outDir, "summary.json"), summary)
	log.Printf("summary written to %s/summary.json", f.outDir)
}

func readDeckList(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		paths = append(paths, line)
	}
	return paths, nil
}

func buildHatFactories(decks []*deckparser.TournamentDeck, paths []string, budget int) []tournament.HatFactory {
	factories := make([]tournament.HatFactory, len(decks))
	for i, p := range paths {
		if i >= len(decks) {
			break
		}
		prof := hat.LoadStrategyFromFreya(p)
		b := budget
		if prof != nil && prof.PowerPercentile > 0 {
			b = hat.BudgetForPower(b, prof.PowerPercentile)
		}
		profCap := prof
		bCap := b
		factories[i] = func() gameengine.Hat {
			return hat.NewYggdrasilHatWithNoise(profCap, bCap, 0.2)
		}
	}
	return factories
}

// passResult collects per-game data for one pool run.
type passResult struct {
	games           int
	crashes         int
	draws           int
	totalTurns      int
	totalConcessions int
	duration        time.Duration

	winsByCmdr   map[string]int
	gamesByCmdr  map[string]int
	turnsByCmdr  map[string]int // total turn count of won games for avg turn to win
	turnDist     [6]int         // 1-5,6-10,11-20,21-40,41-60,61+
	endReasons   map[string]int
	crashRecords []crashRecord
	concessions  []muninn.ConcessionRecord
	perGame      []perGameRecord
	samples      []sampleGameRecord // populated only for sample pass
}

type crashRecord struct {
	GameIdx      int      `json:"game_idx"`
	Participants []string `json:"participants"`
	Err          string   `json:"error"`
}

func (p *passResult) summary() map[string]any {
	gps := 0.0
	if p.duration.Seconds() > 0 {
		gps = float64(p.games) / p.duration.Seconds()
	}
	winrates := make([]map[string]any, 0, len(p.winsByCmdr))
	for name, played := range p.gamesByCmdr {
		w := p.winsByCmdr[name]
		var rate float64
		if played > 0 {
			rate = float64(w) / float64(played)
		}
		winrates = append(winrates, map[string]any{
			"commander": name, "games": played, "wins": w, "winrate": rate,
		})
	}
	sort.Slice(winrates, func(i, j int) bool {
		return winrates[i]["winrate"].(float64) > winrates[j]["winrate"].(float64)
	})
	return map[string]any{
		"games":               p.games,
		"crashes":             p.crashes,
		"draws":               p.draws,
		"avg_turns":           safeDiv(float64(p.totalTurns), float64(p.games)),
		"total_concessions":   p.totalConcessions,
		"duration_sec":        p.duration.Seconds(),
		"games_per_sec":       gps,
		"end_reasons":         p.endReasons,
		"turn_distribution":   map[string]int{"1-5": p.turnDist[0], "6-10": p.turnDist[1], "11-20": p.turnDist[2], "21-40": p.turnDist[3], "41-60": p.turnDist[4], "61+": p.turnDist[5]},
		"winrate_buckets":     bucketWinrates(winrates),
		"top_10_winrate":      topN(winrates, 10),
		"bottom_10_winrate":   bottomN(winrates, 10),
		"crash_records":       p.crashRecords,
		"concession_blunders": classifyConcessions(p.concessions),
	}
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func bucketWinrates(rates []map[string]any) map[string]int {
	buckets := map[string]int{"0%": 0, "0-10%": 0, "10-20%": 0, "20-30%": 0, "30-40%": 0, "40-50%": 0, "50%+": 0}
	for _, r := range rates {
		wr := r["winrate"].(float64) * 100
		switch {
		case wr == 0:
			buckets["0%"]++
		case wr < 10:
			buckets["0-10%"]++
		case wr < 20:
			buckets["10-20%"]++
		case wr < 30:
			buckets["20-30%"]++
		case wr < 40:
			buckets["30-40%"]++
		case wr < 50:
			buckets["40-50%"]++
		default:
			buckets["50%+"]++
		}
	}
	return buckets
}

func topN(rates []map[string]any, n int) []map[string]any {
	if n > len(rates) {
		n = len(rates)
	}
	return rates[:n]
}

func bottomN(rates []map[string]any, n int) []map[string]any {
	if n > len(rates) {
		n = len(rates)
	}
	out := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		out[i] = rates[len(rates)-1-i]
	}
	return out
}

type concessionClassification struct {
	HighLife    []muninn.ConcessionRecord `json:"high_life"`     // life >= 25
	Early       []muninn.ConcessionRecord `json:"early"`         // turn <= 5
	StrongBoard []muninn.ConcessionRecord `json:"strong_board"`  // board_power >= 8
	Total       int                       `json:"total"`
}

func classifyConcessions(records []muninn.ConcessionRecord) concessionClassification {
	c := concessionClassification{Total: len(records)}
	for _, r := range records {
		if r.Life >= 25 {
			c.HighLife = append(c.HighLife, r)
		}
		if r.Turn > 0 && r.Turn <= 5 {
			c.Early = append(c.Early, r)
		}
		if r.BoardPower >= 8 {
			c.StrongBoard = append(c.StrongBoard, r)
		}
	}
	return c
}

func deckPathSnippets(paths []string, n int) []string {
	if n > len(paths) {
		n = len(paths)
	}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = filepath.Base(paths[i])
	}
	return out
}

func writeJSON(path string, v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("marshal %s: %v", path, err)
		return
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		log.Printf("write %s: %v", path, err)
	}
}

// runPool drives a pool-mode tournament directly so we can capture
// per-game data (the canonical hexdek-tournament binary aggregates
// into Muninn but drops per-game records before returning).
func runPool(decks []*deckparser.TournamentDeck, paths []string, factories []tournament.HatFactory, nGames, workers int, seed int64, maxTurns int, analytics bool, outDir, tag string) *passResult {
	nSeats := 4
	res := &passResult{
		winsByCmdr:  map[string]int{},
		gamesByCmdr: map[string]int{},
		turnsByCmdr: map[string]int{},
		endReasons:  map[string]int{},
	}

	// One uniform hat factory for the whole pool — yggdrasil with neutral
	// budget. (Per-deck Freya weights still load via the factory because
	// each factory closure carries its own profile.)
	uniformHat := factories[0]

	type job struct {
		gameIdx  int
		deckIdxs []int
	}
	type outcomeWrap struct {
		out      tournament.GameOutcome
		deckIdxs []int
	}

	jobs := make(chan job, workers*2)
	outcomes := make(chan outcomeWrap, workers*4)

	var completed int64
	start := time.Now()
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				podDecks := make([]*deckparser.TournamentDeck, nSeats)
				podHats := make([]tournament.HatFactory, nSeats)
				for i, idx := range j.deckIdxs {
					podDecks[i] = decks[idx]
					// Use the deck's own factory (with its Freya strategy) — fall
					// back to the uniform hat if out of range.
					if idx < len(factories) && factories[idx] != nil {
						podHats[i] = factories[idx]
					} else {
						podHats[i] = uniformHat
					}
				}
				o := tournament.RunOneGameForAudit(j.gameIdx, podDecks, podHats, nSeats, seed, maxTurns, 0, true, true, analytics)
				outcomes <- outcomeWrap{o, j.deckIdxs}
				done := atomic.AddInt64(&completed, 1)
				if done%100 == 0 {
					gps := float64(done) / time.Since(start).Seconds()
					fmt.Fprintf(os.Stderr, "  [%s] %d/%d games (%.1f g/s)\n", tag, done, nGames, gps)
				}
			}
		}()
	}

	go func() {
		rng := rand.New(rand.NewSource(seed))
		nDecks := len(decks)
		for i := 0; i < nGames; i++ {
			perm := rng.Perm(nDecks)
			idxs := make([]int, nSeats)
			copy(idxs, perm[:nSeats])
			jobs <- job{gameIdx: i, deckIdxs: idxs}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(outcomes)
	}()

	// Open per-game JSONL files.
	pgFile, err := os.Create(filepath.Join(outDir, tag+"_per_game.jsonl"))
	if err != nil {
		log.Fatalf("create per-game jsonl: %v", err)
	}
	defer pgFile.Close()
	pgEnc := json.NewEncoder(pgFile)

	var sampleFile *os.File
	var sampleEnc *json.Encoder
	if analytics {
		sampleFile, err = os.Create(filepath.Join(outDir, tag+"_samples.jsonl"))
		if err != nil {
			log.Fatalf("create samples jsonl: %v", err)
		}
		defer sampleFile.Close()
		sampleEnc = json.NewEncoder(sampleFile)
	}

	for ow := range outcomes {
		o := ow.out
		participants := make([]string, 0, len(ow.deckIdxs))
		for _, idx := range ow.deckIdxs {
			if idx < len(decks) {
				participants = append(participants, decks[idx].CommanderName)
			}
		}
		winnerName := ""
		if o.WinnerCommanderIdx >= 0 && o.WinnerCommanderIdx < len(participants) {
			// WinnerCommanderIdx is a pod-local index (0..nSeats-1); look up
			// the global commander name via the pod's deckIdxs.
			winnerName = participants[o.WinnerCommanderIdx]
		}

		if o.CrashErr != "" {
			res.crashes++
			res.crashRecords = append(res.crashRecords, crashRecord{
				GameIdx: o.GameIdx, Participants: participants, Err: truncate(o.CrashErr, 2000),
			})
			rec := perGameRecord{
				GameIdx: o.GameIdx, Turns: o.Turns, EndReason: o.EndReason,
				Winner: o.Winner, WinnerCmdr: winnerName, Participants: participants,
				Crashed: true, CrashErr: truncate(o.CrashErr, 400),
				Concessions: o.Concessions,
			}
			_ = pgEnc.Encode(&rec)
			continue
		}

		res.games++
		res.totalTurns += o.Turns
		res.totalConcessions += o.Concessions
		res.endReasons[o.EndReason]++
		switch {
		case o.Turns <= 5:
			res.turnDist[0]++
		case o.Turns <= 10:
			res.turnDist[1]++
		case o.Turns <= 20:
			res.turnDist[2]++
		case o.Turns <= 40:
			res.turnDist[3]++
		case o.Turns <= 60:
			res.turnDist[4]++
		default:
			res.turnDist[5]++
		}

		for _, idx := range ow.deckIdxs {
			if idx < len(decks) {
				res.gamesByCmdr[decks[idx].CommanderName]++
			}
		}
		if winnerName != "" {
			res.winsByCmdr[winnerName]++
			res.turnsByCmdr[winnerName] += o.Turns
		} else {
			res.draws++
		}
		res.concessions = append(res.concessions, o.ConcessionRecords...)

		_ = pgEnc.Encode(&perGameRecord{
			GameIdx: o.GameIdx, Turns: o.Turns, EndReason: o.EndReason,
			Winner: o.Winner, WinnerCmdr: winnerName, Participants: participants,
			Concessions: o.Concessions,
		})

		// Sample-pass deep record.
		if analytics && o.Analysis != nil {
			sr := sampleGameRecord{
				GameIdx: o.GameIdx, Turns: o.Turns, EndReason: o.EndReason,
				Participants: participants, WinnerCmdr: winnerName,
				WinCondition: o.Analysis.WinCondition, WinningCard: o.Analysis.WinningCard,
				FirstBlood: o.Analysis.FirstBlood,
			}
			for _, mc := range o.Analysis.MissedCombos {
				cmd := ""
				if mc.Seat >= 0 && mc.Seat < len(participants) {
					cmd = participants[mc.Seat]
				}
				sr.MissedCombos = append(sr.MissedCombos, missedComboRec{
					Seat: mc.Seat, Commander: cmd, Turn: mc.Turn,
					ComboName: mc.ComboName, Pieces: mc.Pieces, WinType: mc.WinType, ManaAvail: mc.ManaAvail,
				})
			}
			for _, mf := range o.Analysis.MissedFinishers {
				cmd := ""
				if mf.Seat >= 0 && mf.Seat < len(participants) {
					cmd = participants[mf.Seat]
				}
				sr.MissedFinishers = append(sr.MissedFinishers, missedFinisherRec{
					Seat: mf.Seat, Commander: cmd, Turn: mf.Turn,
					FinisherName: mf.FinisherName, BoardPower: mf.BoardPower, OppLifeMin: mf.OppLifeMin,
				})
			}
			if o.Analysis.StallIndicators != nil {
				sr.StallHitTurnCap = o.Analysis.StallIndicators.HitTurnCap
				sr.StallSurvivors = o.Analysis.StallIndicators.SurvivorsAtEnd
				sr.StallCause = o.Analysis.StallIndicators.Cause
			}
			sr.Concessions = o.ConcessionRecords
			_ = sampleEnc.Encode(&sr)
			res.samples = append(res.samples, sr)
		}
	}
	res.duration = time.Since(start)
	log.Printf("[%s] complete: %d games, %d crashes, %d draws in %s (%.1f g/s)",
		tag, res.games, res.crashes, res.draws, res.duration.Round(time.Millisecond),
		float64(res.games+res.crashes)/res.duration.Seconds())
	return res
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
