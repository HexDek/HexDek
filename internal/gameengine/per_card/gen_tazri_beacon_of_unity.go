package per_card

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// registerTazriBeaconOfUnity wires Tazri, Beacon of Unity.
//
// Oracle text:
//
//	This spell costs {1} less to cast for each creature in your party.
//	{2/U}{2/B}{2/R}{2/G}: Look at the top six cards of your library. You may reveal up to two Cleric, Rogue, Warrior, Wizard, and/or Ally cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.
//
// Implementation:
//   - Cost reduction: Handled in ScanCostModifiers (cost_modifiers.go)
//     via the "Tazri, Beacon of Unity" case in the switch statement.
//     Reduces cost by CountParty(gs, seatIdx) (0-4).
//   - Activated ability: Look at the top 6 cards of the library. Pick
//     up to 2 that are Cleric, Rogue, Warrior, Wizard, or Ally type.
//     Put them into hand. Put the rest on the bottom in random order.
//     The hat picks greedily by mana value (highest first) to maximize
//     card quality.
func registerTazriBeaconOfUnity(r *Registry) {
	r.OnActivated("Tazri, Beacon of Unity", tazriBeaconOfUnityActivate)
}

// tazriPartyTypes are the creature types Tazri's activated ability can
// pick from the top of the library. Ally is included per oracle text.
var tazriPartyTypes = []string{"cleric", "rogue", "warrior", "wizard", "ally"}

func tazriBeaconOfUnityActivate(gs *gameengine.GameState, src *gameengine.Permanent, abilityIdx int, ctx map[string]interface{}) {
	const slug = "tazri_beacon_of_unity_activate"
	if gs == nil || src == nil || src.Card == nil {
		return
	}
	seat := src.Controller
	if seat < 0 || seat >= len(gs.Seats) {
		return
	}
	s := gs.Seats[seat]
	if s == nil || s.Lost {
		return
	}

	// Look at the top 6 cards.
	look := 6
	if look > len(s.Library) {
		look = len(s.Library)
	}
	if look == 0 {
		emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
			"seat":   seat,
			"reason": "library_empty",
		})
		return
	}

	// Snapshot the top cards.
	top := make([]*gameengine.Card, look)
	copy(top, s.Library[:look])

	// Find eligible cards (Cleric, Rogue, Warrior, Wizard, or Ally).
	type candidate struct {
		card *gameengine.Card
		idx  int
		mv   int // mana value for greedy selection
	}
	var eligible []candidate
	for i, c := range top {
		if c == nil {
			continue
		}
		if tazriCardIsPartyOrAlly(c) {
			mv := 0
			if c.CMC > 0 {
				mv = c.CMC
			}
			eligible = append(eligible, candidate{card: c, idx: i, mv: mv})
		}
	}

	// Pick up to 2, preferring highest mana value (greedy for card quality).
	// Sort by MV descending — simple selection sort for at most 6 items.
	for i := 0; i < len(eligible)-1; i++ {
		best := i
		for j := i + 1; j < len(eligible); j++ {
			if eligible[j].mv > eligible[best].mv {
				best = j
			}
		}
		if best != i {
			eligible[i], eligible[best] = eligible[best], eligible[i]
		}
	}

	maxPick := 2
	if len(eligible) < maxPick {
		maxPick = len(eligible)
	}
	picked := make(map[int]bool)
	var pulledNames []string
	for i := 0; i < maxPick; i++ {
		c := eligible[i]
		picked[c.idx] = true
		// Remove from the library by pointer match.
		for j, lc := range s.Library {
			if lc == c.card {
				s.Library = append(s.Library[:j], s.Library[j+1:]...)
				break
			}
		}
		// Put into hand.
		s.Hand = append(s.Hand, c.card)
		gs.LogEvent(gameengine.Event{
			Kind:   "draw",
			Seat:   seat,
			Source: src.Card.DisplayName(),
			Details: map[string]interface{}{
				"slug":   slug,
				"reason": "tazri_party_pick",
				"card":   c.card.DisplayName(),
			},
		})
		pulledNames = append(pulledNames, c.card.DisplayName())
	}

	// Remaining looked-at cards go to the bottom in random order.
	// Adjust for cards already removed: we need to move the first
	// `look - len(picked)` cards from the top of the library to the bottom.
	remaining := look - len(picked)
	if remaining > len(s.Library) {
		remaining = len(s.Library)
	}
	if remaining > 0 {
		rest := make([]*gameengine.Card, remaining)
		copy(rest, s.Library[:remaining])
		s.Library = s.Library[remaining:]
		if gs.Rng != nil && len(rest) > 1 {
			gs.Rng.Shuffle(len(rest), func(i, j int) {
				rest[i], rest[j] = rest[j], rest[i]
			})
		}
		s.Library = append(s.Library, rest...)
	}

	emit(gs, slug, src.Card.DisplayName(), map[string]interface{}{
		"seat":           seat,
		"looked":         look,
		"eligible":       len(eligible),
		"taken_count":    len(pulledNames),
		"taken_cards":    pulledNames,
		"bottomed_count": remaining,
	})
}

// tazriCardIsPartyOrAlly returns true if the card has a party creature
// type (Cleric, Rogue, Warrior, Wizard) or the Ally type.
func tazriCardIsPartyOrAlly(c *gameengine.Card) bool {
	if c == nil {
		return false
	}
	tl := strings.ToLower(strings.Join(c.Types, " "))
	for _, role := range tazriPartyTypes {
		if strings.Contains(tl, role) {
			return true
		}
	}
	// Also check TypeLine for broader matching (some cards have subtypes
	// only in TypeLine, not in the Types slice).
	if c.TypeLine != "" {
		tl2 := strings.ToLower(c.TypeLine)
		for _, role := range tazriPartyTypes {
			if strings.Contains(tl2, role) {
				return true
			}
		}
	}
	return false
}
