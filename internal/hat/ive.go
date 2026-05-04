package hat

import (
	"fmt"
	"strings"
)

// Ive Three-Act Spectator — transforms a game's causal graph into a
// narrative arc for spectator rendering.
//
// Every Commander game has a story: setup (ramp, early plays), conflict
// (the pivot — a boardwipe, combo attempt, or political betrayal), and
// resolution (the winner emerges). Ive extracts this structure from
// Tesla's causal pivots and the game's event log.

// GameNarrative is the structured three-act summary of a completed game.
type GameNarrative struct {
	Acts       [3]Act         `json:"acts"`
	Highlights []Highlight    `json:"highlights"`
	Winner     string         `json:"winner"`     // commander name
	WinnerSeat int            `json:"winner_seat"`
	TotalTurns int            `json:"total_turns"`
	Pivot      CausalPivot    `json:"pivot"`
	Synopsis   string         `json:"synopsis"` // one-sentence summary
}

// Act is one segment of the three-act structure.
type Act struct {
	Name      string `json:"name"`       // "Setup", "Conflict", "Resolution"
	StartTurn int    `json:"start_turn"`
	EndTurn   int    `json:"end_turn"`
	Summary   string `json:"summary"`
}

// Highlight is a notable moment in the game worth calling out.
type Highlight struct {
	Turn        int    `json:"turn"`
	Seat        int    `json:"seat"`
	Kind        string `json:"kind"` // "first_blood", "pivot", "elimination", "combo", "boardwipe"
	Description string `json:"description"`
}

// GameEvent is a simplified event from the engine's event log,
// used to extract highlights.
type GameEvent struct {
	Turn   int
	Seat   int
	Kind   string
	Source string
	Amount int
}

// ComposeNarrative builds a GameNarrative from game data.
func ComposeNarrative(
	pivot CausalPivot,
	events []GameEvent,
	seatNames []string, // commander name per seat
	winnerSeat int,
	totalTurns int,
) GameNarrative {
	n := GameNarrative{
		WinnerSeat: winnerSeat,
		TotalTurns: totalTurns,
		Pivot:      pivot,
	}
	if winnerSeat >= 0 && winnerSeat < len(seatNames) {
		n.Winner = seatNames[winnerSeat]
	}

	// Divide game into three acts based on pivot location.
	// Act 1: turns 1 to pivot-2 (or min 3 turns)
	// Act 2: pivot-2 to pivot+2 (the conflict window)
	// Act 3: pivot+2 to end
	pivotTurn := pivot.Turn
	if pivotTurn < 3 {
		pivotTurn = 3
	}
	if pivotTurn > totalTurns-2 {
		pivotTurn = totalTurns - 2
	}

	act1End := pivotTurn - 2
	if act1End < 1 {
		act1End = 1
	}
	act2End := pivotTurn + 2
	if act2End > totalTurns {
		act2End = totalTurns
	}

	n.Acts[0] = Act{
		Name:      "Setup",
		StartTurn: 1,
		EndTurn:   act1End,
		Summary:   composeSetupSummary(events, act1End, seatNames),
	}
	n.Acts[1] = Act{
		Name:      "Conflict",
		StartTurn: act1End + 1,
		EndTurn:   act2End,
		Summary:   composeConflictSummary(events, act1End+1, act2End, pivot, seatNames),
	}
	n.Acts[2] = Act{
		Name:      "Resolution",
		StartTurn: act2End + 1,
		EndTurn:   totalTurns,
		Summary:   composeResolutionSummary(n.Winner, totalTurns-act2End),
	}

	// Extract highlights.
	n.Highlights = extractHighlights(events, pivot, seatNames)

	// Synopsis.
	n.Synopsis = composeSynopsis(n, seatNames)

	return n
}

func composeSetupSummary(events []GameEvent, endTurn int, names []string) string {
	landDrops := make(map[int]int) // seat → count
	firstCast := make(map[int]string)
	for _, e := range events {
		if e.Turn > endTurn {
			break
		}
		if e.Kind == "enter_battlefield" && strings.Contains(strings.ToLower(e.Source), "land") {
			landDrops[e.Seat]++
		}
		if e.Kind == "cast_spell" {
			if _, ok := firstCast[e.Seat]; !ok {
				firstCast[e.Seat] = e.Source
			}
		}
	}
	// Find who ramped fastest.
	maxLands := 0
	fastSeat := 0
	for s, c := range landDrops {
		if c > maxLands {
			maxLands = c
			fastSeat = s
		}
	}
	name := seatName(names, fastSeat)
	return fmt.Sprintf("%s established the early board with %d lands by turn %d", name, maxLands, endTurn)
}

func composeConflictSummary(events []GameEvent, start, end int, pivot CausalPivot, names []string) string {
	// Look for boardwipes, eliminations, or combo events near the pivot.
	for _, e := range events {
		if e.Turn < start || e.Turn > end {
			continue
		}
		switch e.Kind {
		case "boardwipe", "destroy_all":
			return fmt.Sprintf("Turn %d: %s wiped the board with %s — the turning point",
				e.Turn, seatName(names, e.Seat), e.Source)
		case "player_lost":
			return fmt.Sprintf("Turn %d: %s fell, shifting the balance of power",
				e.Turn, seatName(names, e.Seat))
		case "combo_assembled":
			return fmt.Sprintf("Turn %d: %s assembled a combo with %s",
				e.Turn, seatName(names, e.Seat), e.Source)
		}
	}
	return fmt.Sprintf("Turn %d: the decisive swing occurred", pivot.Turn)
}

func composeResolutionSummary(winner string, turnsAfterPivot int) string {
	if turnsAfterPivot <= 2 {
		return fmt.Sprintf("%s closed out immediately after seizing the advantage", winner)
	}
	return fmt.Sprintf("%s converted the advantage over %d turns to take the game", winner, turnsAfterPivot)
}

func extractHighlights(events []GameEvent, pivot CausalPivot, names []string) []Highlight {
	var highlights []Highlight
	firstBloodSeen := false
	eliminationSeen := make(map[int]bool)

	for _, e := range events {
		switch e.Kind {
		case "damage":
			if !firstBloodSeen && e.Amount > 0 {
				firstBloodSeen = true
				highlights = append(highlights, Highlight{
					Turn: e.Turn, Seat: e.Seat, Kind: "first_blood",
					Description: fmt.Sprintf("%s drew first blood with %s",
						seatName(names, e.Seat), e.Source),
				})
			}
		case "player_lost":
			if !eliminationSeen[e.Seat] {
				eliminationSeen[e.Seat] = true
				highlights = append(highlights, Highlight{
					Turn: e.Turn, Seat: e.Seat, Kind: "elimination",
					Description: fmt.Sprintf("%s was eliminated", seatName(names, e.Seat)),
				})
			}
		case "boardwipe", "destroy_all":
			highlights = append(highlights, Highlight{
				Turn: e.Turn, Seat: e.Seat, Kind: "boardwipe",
				Description: fmt.Sprintf("%s cast %s", seatName(names, e.Seat), e.Source),
			})
		case "combo_assembled":
			highlights = append(highlights, Highlight{
				Turn: e.Turn, Seat: e.Seat, Kind: "combo",
				Description: fmt.Sprintf("%s assembled a combo", seatName(names, e.Seat)),
			})
		}
	}

	// Always include the pivot.
	highlights = append(highlights, Highlight{
		Turn: pivot.Turn, Seat: pivot.WinnerSeat, Kind: "pivot",
		Description: fmt.Sprintf("The decisive moment (swing=%.2f)", pivot.DeltaScore),
	})

	return highlights
}

func composeSynopsis(n GameNarrative, names []string) string {
	nPlayers := len(names)
	if nPlayers == 0 {
		return ""
	}
	return fmt.Sprintf("%d-player game over %d turns. %s won after a pivotal turn %d.",
		nPlayers, n.TotalTurns, n.Winner, n.Pivot.Turn)
}

func seatName(names []string, seat int) string {
	if seat >= 0 && seat < len(names) && names[seat] != "" {
		return names[seat]
	}
	return fmt.Sprintf("Seat %d", seat)
}
