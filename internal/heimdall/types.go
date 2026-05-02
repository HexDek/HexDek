package heimdall

// GameSeed stores the minimum data needed to deterministically replay a game.
// 28 bytes per game. At 50M games/day = 1.4 GB/day.
type GameSeed struct {
	RNGSeed    int64     `json:"rng_seed"`
	DeckKeys   [4]string `json:"deck_keys"`
	Winner     int       `json:"winner"`
	Turns      int       `json:"turns"`
	KillMethod string    `json:"kill_method"` // combat, commander, combo, mill, poison, timeout
}

// Observation is the lightweight per-game extraction. ~200 bytes.
// Only populated during batch replay or live observation mode.
type Observation struct {
	Seed           GameSeed
	ParserGaps     []string        // card names that hit unhandled abilities
	DeadTriggers   []DeadTrigger   // registered but never fired
	ComboAttempted bool            // did the deck attempt its Freya combo?
	ComboSucceeded bool            // did the combo resolve to a win?
	ComboMissed    bool            // were combo pieces available but not used?
	CoTriggers     []CoTriggerPair // Huginn food: cards that synergized
	CausalPivot    *PivotEvent     // Tesla: the turn/action that decided the game
}

// DeadTrigger records a trigger that was registered but never fired during the game.
type DeadTrigger struct {
	CardName    string
	TriggerType string // etb, ltb, dies, cast, etc.
}

// CoTriggerPair records two cards that synergized within a turn window.
type CoTriggerPair struct {
	CardA       string
	CardB       string
	ImpactScore float64
	TurnWindow  int
}

// PivotEvent records the turn/action that decided the game outcome.
type PivotEvent struct {
	Turn   int
	Action string
	Seat   int
}

// HealthPulse is the periodic telemetry snapshot sent to GA4.
type HealthPulse struct {
	GamesPlayed   int
	ParserGaps    int
	Crashes       int
	DeadTriggers  int
	TopGapCards   []string
	EngineVersion string
}
