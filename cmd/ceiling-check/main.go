package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type winLine struct {
	Pieces []string `json:"pieces"`
	Type   string   `json:"type"`
}

type strategy struct {
	Bracket         int       `json:"bracket"`
	Archetype       string    `json:"archetype"`
	GameplanSummary string    `json:"gameplan_summary"`
	WinLines        []winLine `json:"win_lines"`
	TutorTargets    []string  `json:"tutor_targets"`
	PowerPercentile int       `json:"power_percentile"`
}

type candidate struct {
	Key             string
	Commander       string
	Archetype       string
	ComboCount      int
	TutorCount      int
	PowerPercentile int
	Score           float64
}

func main() {
	root := flag.String("decks", "data/decks", "deck directory to scan")
	top := flag.Int("top", 10, "number of candidates to print")
	flag.Parse()

	var cands []candidate
	err := filepath.Walk(*root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".strategy.json") {
			return nil
		}
		c, ok := loadCandidate(path)
		if !ok {
			return nil
		}
		cands = append(cands, c)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk %s: %v\n", *root, err)
		os.Exit(1)
	}

	sort.Slice(cands, func(i, j int) bool {
		if cands[i].Score != cands[j].Score {
			return cands[i].Score > cands[j].Score
		}
		return cands[i].ComboCount > cands[j].ComboCount
	})

	fmt.Printf("Ceiling calibration candidates -- bracket 5 decks ranked by combo density + power percentile\n")
	fmt.Printf("Scanned %s, found %d B5 deck(s)\n\n", *root, len(cands))

	if len(cands) == 0 {
		return
	}

	limit := *top
	if limit > len(cands) {
		limit = len(cands)
	}

	fmt.Printf("%-4s  %-50s  %-32s  %-14s  %5s  %5s  %6s\n",
		"#", "Deck Key", "Commander", "Archetype", "Combo", "Tutor", "Power")
	fmt.Println(strings.Repeat("-", 130))
	for i, c := range cands[:limit] {
		fmt.Printf("%-4d  %-50s  %-32s  %-14s  %5d  %5d  %5s\n",
			i+1, truncate(c.Key, 50), truncate(c.Commander, 32), truncate(c.Archetype, 14),
			c.ComboCount, c.TutorCount, formatPower(c.PowerPercentile))
	}
}

func loadCandidate(stratPath string) (candidate, bool) {
	data, err := os.ReadFile(stratPath)
	if err != nil {
		return candidate{}, false
	}
	var s strategy
	if err := json.Unmarshal(data, &s); err != nil {
		return candidate{}, false
	}
	if s.Bracket != 5 {
		return candidate{}, false
	}

	base := strings.TrimSuffix(filepath.Base(stratPath), ".strategy.json")
	deckDir := filepath.Dir(filepath.Dir(stratPath))
	commander := readCommander(filepath.Join(deckDir, base+".txt"))

	c := candidate{
		Key:             base,
		Commander:       commander,
		Archetype:       s.Archetype,
		ComboCount:      len(s.WinLines),
		TutorCount:      len(s.TutorTargets),
		PowerPercentile: s.PowerPercentile,
	}
	c.Score = float64(c.ComboCount)*10 + float64(c.PowerPercentile) + float64(c.TutorCount)*0.1
	return c, true
}

func readCommander(deckPath string) string {
	f, err := os.Open(deckPath)
	if err != nil {
		return "(unknown)"
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToUpper(line), "COMMANDER:") {
			return strings.TrimSpace(line[len("COMMANDER:"):])
		}
	}
	return "(unknown)"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func formatPower(p int) string {
	if p <= 0 {
		return "  n/a"
	}
	return fmt.Sprintf("%4d", p)
}
