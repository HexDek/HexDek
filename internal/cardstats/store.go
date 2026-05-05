// Package cardstats accumulates per-card win/loss/play telemetry across
// games run by the grinder. Lightweight, in-memory only — exists to
// power the /api/analytics/cards endpoint and any future card-page
// analytics that need empirical performance data.
//
// Persistence is intentionally absent: the grinder runs continuously
// from boot, so a fresh process simply rebuilds the stats from the
// next N games. Add disk persistence behind a flag if/when needed.
package cardstats

import (
	"sort"
	"strings"
	"sync"
)

// CardStats is the accumulator per unique card name (case-folded key).
type CardStats struct {
	// Wins/Losses/Games count any deck containing the card, regardless
	// of whether the card was actually played in that game. This
	// matches the user-facing "this card is in winning decks" intuition.
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Games  int `json:"games"`

	// TotalTurnPlayed and TimesPlayed track only games where the card
	// hit the battlefield. avgTurn = TotalTurnPlayed / TimesPlayed.
	TotalTurnPlayed int `json:"total_turn_played"`
	TimesPlayed     int `json:"times_played"`
}

// WinRate returns wins / (wins+losses). Returns 0 when no games.
func (c CardStats) WinRate() float64 {
	denom := c.Wins + c.Losses
	if denom == 0 {
		return 0
	}
	return float64(c.Wins) / float64(denom)
}

// AvgTurnPlayed returns TotalTurnPlayed / TimesPlayed, or 0 if never played.
func (c CardStats) AvgTurnPlayed() float64 {
	if c.TimesPlayed == 0 {
		return 0
	}
	return float64(c.TotalTurnPlayed) / float64(c.TimesPlayed)
}

// Store is the thread-safe in-memory aggregate. A single Default
// instance lives at package scope; the grinder writes, the API reads.
type Store struct {
	mu    sync.RWMutex
	cards map[string]*CardStats
}

func NewStore() *Store {
	return &Store{cards: make(map[string]*CardStats, 1024)}
}

// Default is the process-wide store the grinder writes into and the
// API reads from. Safe for concurrent use.
var Default = NewStore()

// Record ingests one seat's deck for one finished game.
//
//	deckCards    : every unique card name on the seat (commanders + library).
//	               Duplicates are de-duped internally so quantity doesn't
//	               skew the per-card tally.
//	won          : true if this seat was the winner; false otherwise (loss
//	               or draw). Draws are counted as losses for every seat,
//	               which keeps the data simple and is rare enough not to
//	               distort win rates over thousands of games.
//	firstTurn    : map[card name (case-insensitive)] -> turn the card was
//	               first observed entering the battlefield in this game.
//	               Cards absent from the map are treated as "never played"
//	               and only contribute to wins/losses/games, not to the
//	               turn-played averages.
//
// Card names are case-folded on insert so callers don't have to.
func (s *Store) Record(deckCards []string, won bool, firstTurn map[string]int) {
	if s == nil || len(deckCards) == 0 {
		return
	}

	// Normalize firstTurn keys + the deck-card list to lowercased,
	// trimmed names, and dedupe so a deck running 7 Mountains doesn't
	// give Mountain 7× the win share of a singleton.
	normFirst := make(map[string]int, len(firstTurn))
	for k, v := range firstTurn {
		clean := strings.ToLower(strings.TrimSpace(k))
		if clean == "" {
			continue
		}
		// If multiple entries collapse to the same key, keep the
		// earliest turn — matches the "first played" semantic.
		if existing, ok := normFirst[clean]; !ok || v < existing {
			normFirst[clean] = v
		}
	}
	seen := make(map[string]struct{}, len(deckCards))

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, raw := range deckCards {
		key := strings.ToLower(strings.TrimSpace(raw))
		if key == "" {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}

		c, ok := s.cards[key]
		if !ok {
			c = &CardStats{}
			s.cards[key] = c
		}
		c.Games++
		if won {
			c.Wins++
		} else {
			c.Losses++
		}
		if turn, played := normFirst[key]; played && turn > 0 {
			c.TotalTurnPlayed += turn
			c.TimesPlayed++
		}
	}
}

// CardRow is the wire format for a card in the analytics dump.
type CardRow struct {
	Name           string  `json:"name"`
	Games          int     `json:"games"`
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	WinRate        float64 `json:"win_rate"`
	TimesPlayed    int     `json:"times_played"`
	AvgTurnPlayed  float64 `json:"avg_turn_played"`
}

// Snapshot returns a sorted list of every card with at least minGames.
// Sort order: win rate descending, then games descending, then name.
func (s *Store) Snapshot(minGames int) []CardRow {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	out := make([]CardRow, 0, len(s.cards))
	for name, c := range s.cards {
		if c.Games < minGames {
			continue
		}
		out = append(out, CardRow{
			Name:          titleCaseName(name),
			Games:         c.Games,
			Wins:          c.Wins,
			Losses:        c.Losses,
			WinRate:       c.WinRate(),
			TimesPlayed:   c.TimesPlayed,
			AvgTurnPlayed: c.AvgTurnPlayed(),
		})
	}
	s.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool {
		if out[i].WinRate != out[j].WinRate {
			return out[i].WinRate > out[j].WinRate
		}
		if out[i].Games != out[j].Games {
			return out[i].Games > out[j].Games
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// TopBottom returns the top-n highest win rate cards and the bottom-n
// lowest, both gated by minGames. The two slices are disjoint when the
// total card count exceeds 2n.
func (s *Store) TopBottom(n, minGames int) (top, bottom []CardRow) {
	all := s.Snapshot(minGames)
	if n <= 0 || len(all) == 0 {
		return nil, nil
	}
	if n > len(all) {
		n = len(all)
	}
	top = append(top, all[:n]...)

	// Bottom by win rate ascending. Re-sort a copy.
	rev := make([]CardRow, len(all))
	copy(rev, all)
	sort.Slice(rev, func(i, j int) bool {
		if rev[i].WinRate != rev[j].WinRate {
			return rev[i].WinRate < rev[j].WinRate
		}
		if rev[i].Games != rev[j].Games {
			return rev[i].Games > rev[j].Games
		}
		return rev[i].Name < rev[j].Name
	})
	if n > len(rev) {
		n = len(rev)
	}
	bottom = append(bottom, rev[:n]...)
	return top, bottom
}

// Size returns how many distinct cards have been recorded.
func (s *Store) Size() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cards)
}

// titleCaseName converts a lowercased card key back to a reasonable
// display form. Mirrors the helper in hexapi/search.go but lives here
// to keep cardstats independent of the API package.
func titleCaseName(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	upNext := true
	for _, r := range s {
		if r == ' ' || r == '-' || r == '/' || r == ',' || r == '(' {
			b.WriteRune(r)
			upNext = true
			continue
		}
		if upNext && r >= 'a' && r <= 'z' {
			b.WriteRune(r - 32)
			upNext = false
			continue
		}
		b.WriteRune(r)
		upNext = false
	}
	return b.String()
}
