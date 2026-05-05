// Package achievements awards badges to deck owners based on per-game
// outcomes. State is persisted as one JSON file per owner under a
// configured directory; the Tracker is safe for concurrent use.
package achievements

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Rarity tiers. Order matters for UI sorting (see rarityOrder).
type Rarity string

const (
	Common   Rarity = "common"
	Uncommon Rarity = "uncommon"
	Rare     Rarity = "rare"
	Mythic   Rarity = "mythic"
	Secret   Rarity = "secret"
)

// Badge is a single earnable achievement definition.
type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rarity      Rarity `json:"rarity"`
	Icon        string `json:"icon"`
}

// Catalog is the full set of defined achievements, keyed by ID.
var Catalog = map[string]Badge{
	"first_blood": {
		ID: "first_blood", Name: "FIRST BLOOD",
		Description: "Win your first game.",
		Rarity:      Common, Icon: "🩸",
	},
	"comeback_sub5": {
		ID: "comeback_sub5", Name: "COMEBACK KID",
		Description: "Win a game with 5 or fewer life remaining.",
		Rarity:      Uncommon, Icon: "💫",
	},
	"perfect_sweep": {
		ID: "perfect_sweep", Name: "PERFECT SWEEP",
		Description: "Win a game without taking a single point of damage (final life ≥ 40).",
		Rarity:      Rare, Icon: "🛡️",
	},
	"ten_users": {
		ID: "ten_users", Name: "REGULAR",
		Description: "Sit across from 10 unique opponents.",
		Rarity:      Common, Icon: "🤝",
	},
	"hundred_users": {
		ID: "hundred_users", Name: "POD VETERAN",
		Description: "Sit across from 100 unique opponents.",
		Rarity:      Uncommon, Icon: "🏛️",
	},
	"thousand_users": {
		ID: "thousand_users", Name: "TABLE LEGEND",
		Description: "Sit across from 1,000 unique opponents.",
		Rarity:      Rare, Icon: "👑",
	},
	"ten_games": {
		ID: "ten_games", Name: "WARMED UP",
		Description: "Play 10 games.",
		Rarity:      Common, Icon: "🎲",
	},
	"hundred_games": {
		ID: "hundred_games", Name: "SEASONED",
		Description: "Play 100 games.",
		Rarity:      Uncommon, Icon: "📚",
	},
	"thousand_games": {
		ID: "thousand_games", Name: "GRINDER",
		Description: "Play 1,000 games.",
		Rarity:      Rare, Icon: "⚙️",
	},
	"early_win": {
		ID: "early_win", Name: "BLITZ",
		Description: "Win a game by turn 5.",
		Rarity:      Uncommon, Icon: "⚡",
	},
	"long_haul": {
		ID: "long_haul", Name: "MARATHON",
		Description: "Win a game lasting 20 or more turns.",
		Rarity:      Uncommon, Icon: "⏳",
	},
	"iron_grip": {
		ID: "iron_grip", Name: "IRON GRIP",
		Description: "Win 5 games in a row.",
		Rarity:      Mythic, Icon: "✊",
	},
	"mythic_run": {
		ID: "mythic_run", Name: "MYTHIC RUN",
		Description: "Win 25 games in a row.",
		Rarity:      Secret, Icon: "🌟",
	},
}

// EarnedBadge is one badge earned by an owner at a specific moment.
type EarnedBadge struct {
	BadgeID   string    `json:"badge_id"`
	AwardedAt time.Time `json:"awarded_at"`
	DeckKey   string    `json:"deck_key,omitempty"`
}

// OwnerState is the persisted per-owner achievement record.
type OwnerState struct {
	Owner            string        `json:"owner"`
	TotalGames       int           `json:"total_games"`
	TotalWins        int           `json:"total_wins"`
	CurrentWinStreak int           `json:"current_win_streak"`
	MaxWinStreak     int           `json:"max_win_streak"`
	OpponentsFaced   []string      `json:"opponents_faced"`
	EarnedBadges     []EarnedBadge `json:"earned_badges"`
}

// SeatOutcome describes one seat in a completed game from the
// achievement-checker's point of view.
type SeatOutcome struct {
	Owner     string
	DeckKey   string
	Won       bool
	FinalLife int
}

// GameOutcome is the input to OnGameComplete. Constructed by callers
// from whatever game-result type they have.
type GameOutcome struct {
	Turns      int
	Seats      []SeatOutcome
	FinishedAt time.Time
}

// EarnedDetail pairs an EarnedBadge with its catalog metadata, suitable
// for direct JSON serialization to the API.
type EarnedDetail struct {
	Badge
	AwardedAt time.Time `json:"awarded_at"`
	DeckKey   string    `json:"deck_key,omitempty"`
}

// Snapshot is the API response shape combining owner state with badge
// metadata and the full catalog (for showing unearned tiles).
type Snapshot struct {
	Owner            string         `json:"owner"`
	TotalGames       int            `json:"total_games"`
	TotalWins        int            `json:"total_wins"`
	CurrentWinStreak int            `json:"current_win_streak"`
	MaxWinStreak     int            `json:"max_win_streak"`
	OpponentsFaced   int            `json:"opponents_faced"`
	Badges           []EarnedDetail `json:"badges"`
	Catalog          []Badge        `json:"catalog"`
}

// Tracker holds per-owner achievement state, persisted under
// dir/{owner}.json files. Safe for concurrent use.
type Tracker struct {
	dir    string
	mu     sync.Mutex
	owners map[string]*OwnerState
}

// NewTracker creates a Tracker and loads any existing per-owner state
// from dir. An empty dir disables persistence (in-memory only).
func NewTracker(dir string) (*Tracker, error) {
	t := &Tracker{dir: dir, owners: map[string]*OwnerState{}}
	if dir == "" {
		return t, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("achievements: mkdir %s: %w", dir, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("achievements: readdir %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var state OwnerState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}
		if state.Owner != "" {
			t.owners[state.Owner] = &state
		}
	}
	return t, nil
}

// GetOwner returns a deep copy of an owner's state, or a zero-value
// state if the owner has no record.
func (t *Tracker) GetOwner(owner string) OwnerState {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.owners[owner]
	if s == nil {
		return OwnerState{Owner: owner, OpponentsFaced: []string{}, EarnedBadges: []EarnedBadge{}}
	}
	out := *s
	out.OpponentsFaced = append([]string(nil), s.OpponentsFaced...)
	out.EarnedBadges = append([]EarnedBadge(nil), s.EarnedBadges...)
	return out
}

// Snapshot resolves an owner's earned badges against the catalog and
// returns the API-ready response. Always includes the full catalog so
// the UI can render unearned tiles.
func (t *Tracker) Snapshot(owner string) Snapshot {
	state := t.GetOwner(owner)
	badges := make([]EarnedDetail, 0, len(state.EarnedBadges))
	for _, b := range state.EarnedBadges {
		meta, ok := Catalog[b.BadgeID]
		if !ok {
			continue
		}
		badges = append(badges, EarnedDetail{Badge: meta, AwardedAt: b.AwardedAt, DeckKey: b.DeckKey})
	}
	catalog := make([]Badge, 0, len(Catalog))
	for _, b := range Catalog {
		catalog = append(catalog, b)
	}
	sort.Slice(catalog, func(i, j int) bool {
		ri, rj := rarityOrder(catalog[i].Rarity), rarityOrder(catalog[j].Rarity)
		if ri != rj {
			return ri < rj
		}
		return catalog[i].ID < catalog[j].ID
	})
	return Snapshot{
		Owner:            owner,
		TotalGames:       state.TotalGames,
		TotalWins:        state.TotalWins,
		CurrentWinStreak: state.CurrentWinStreak,
		MaxWinStreak:     state.MaxWinStreak,
		OpponentsFaced:   len(state.OpponentsFaced),
		Badges:           badges,
		Catalog:          catalog,
	}
}

// OnGameComplete updates state for every owner in the game and awards
// any newly earned badges. Each affected owner's file is rewritten.
func (t *Tracker) OnGameComplete(g GameOutcome) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Collapse seats to one record per owner. An owner with multiple
	// seats in the same pod (rare) wins the game if any of their seats
	// wins; the winning seat's deck key is preserved for award context.
	type ownerGame struct {
		won         bool
		winningSeat SeatOutcome
	}
	perOwner := map[string]*ownerGame{}
	var ownerOrder []string
	for _, seat := range g.Seats {
		if seat.Owner == "" {
			continue
		}
		og := perOwner[seat.Owner]
		if og == nil {
			og = &ownerGame{}
			perOwner[seat.Owner] = og
			ownerOrder = append(ownerOrder, seat.Owner)
		}
		if seat.Won {
			og.won = true
			og.winningSeat = seat
		}
	}

	var firstErr error
	for _, owner := range ownerOrder {
		og := perOwner[owner]
		opponents := make(map[string]bool, len(perOwner)-1)
		for o := range perOwner {
			if o != owner {
				opponents[o] = true
			}
		}
		state := t.owners[owner]
		if state == nil {
			state = &OwnerState{Owner: owner}
			t.owners[owner] = state
		}
		applyGameToOwner(state, g, og.won, og.winningSeat, opponents)
		if err := t.persistOwner(state); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func applyGameToOwner(state *OwnerState, g GameOutcome, won bool, winSeat SeatOutcome, opponents map[string]bool) {
	state.TotalGames++
	if won {
		state.TotalWins++
		state.CurrentWinStreak++
		if state.CurrentWinStreak > state.MaxWinStreak {
			state.MaxWinStreak = state.CurrentWinStreak
		}
	} else {
		state.CurrentWinStreak = 0
	}

	// Merge opponent set.
	seen := make(map[string]bool, len(state.OpponentsFaced))
	for _, o := range state.OpponentsFaced {
		seen[o] = true
	}
	for o := range opponents {
		if !seen[o] {
			seen[o] = true
			state.OpponentsFaced = append(state.OpponentsFaced, o)
		}
	}
	sort.Strings(state.OpponentsFaced)

	// Run badge checks.
	already := make(map[string]bool, len(state.EarnedBadges))
	for _, b := range state.EarnedBadges {
		already[b.BadgeID] = true
	}
	award := func(id string) {
		if already[id] {
			return
		}
		if _, ok := Catalog[id]; !ok {
			return
		}
		deckKey := ""
		if won {
			deckKey = winSeat.DeckKey
		}
		state.EarnedBadges = append(state.EarnedBadges, EarnedBadge{
			BadgeID: id, AwardedAt: g.FinishedAt, DeckKey: deckKey,
		})
		already[id] = true
	}

	if won {
		if state.TotalWins == 1 {
			award("first_blood")
		}
		if winSeat.FinalLife > 0 && winSeat.FinalLife <= 5 {
			award("comeback_sub5")
		}
		if winSeat.FinalLife >= 40 {
			award("perfect_sweep")
		}
		if g.Turns > 0 && g.Turns <= 5 {
			award("early_win")
		}
		if g.Turns >= 20 {
			award("long_haul")
		}
		if state.CurrentWinStreak >= 5 {
			award("iron_grip")
		}
		if state.CurrentWinStreak >= 25 {
			award("mythic_run")
		}
	}

	n := len(state.OpponentsFaced)
	if n >= 10 {
		award("ten_users")
	}
	if n >= 100 {
		award("hundred_users")
	}
	if n >= 1000 {
		award("thousand_users")
	}

	switch {
	case state.TotalGames >= 1000:
		award("thousand_games")
		fallthrough
	case state.TotalGames >= 100:
		award("hundred_games")
		fallthrough
	case state.TotalGames >= 10:
		award("ten_games")
	}
}

func (t *Tracker) persistOwner(s *OwnerState) error {
	if t.dir == "" {
		return nil
	}
	path := filepath.Join(t.dir, sanitizeOwner(s.Owner)+".json")
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("achievements: marshal %s: %w", s.Owner, err)
	}
	out = append(out, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return fmt.Errorf("achievements: write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("achievements: rename %s: %w", path, err)
	}
	return nil
}

// sanitizeOwner strips path separators and other filesystem-hostile
// characters so an owner name can safely become a filename.
func sanitizeOwner(o string) string {
	if o == "" {
		return "_anon"
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		}
		return r
	}, o)
}

func rarityOrder(r Rarity) int {
	switch r {
	case Common:
		return 0
	case Uncommon:
		return 1
	case Rare:
		return 2
	case Mythic:
		return 3
	case Secret:
		return 4
	}
	return 99
}
