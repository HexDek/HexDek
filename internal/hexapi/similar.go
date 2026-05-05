package hexapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// handleSimilarDecks answers GET /api/decks/{owner}/{id}/similar with the
// top-N decks ranked by composite similarity to the target deck.
//
// Score = shared_cards (raw count, after stripping the commander)
//       + 30  if same commander_card
//       + 10  if same archetype  (Freya strategy.json)
//       +  5  if same bracket    (Freya strategy.json or filename slug)
//
// Decks with shared_cards <= 10 AND no commander/archetype/bracket bonus
// are dropped — the frontend's "no similar decks found" empty state
// kicks in when the response is empty.
//
// Doing this server-side avoids the N+1 fan-out the frontend would need
// otherwise (the deck-list endpoint returns metadata only). The walk is
// O(n_decks * avg_deck_size); at the current scale (a few hundred decks,
// ~100 cards each) it's a single-digit-ms scan.
func (h *Handler) handleSimilarDecks(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	id := r.PathValue("id")
	if !validatePathComponent(owner) || !validatePathComponent(id) {
		http.Error(w, "invalid owner or id", http.StatusBadRequest)
		return
	}
	ownPath := findDeckFile(h.DecksDir, owner, id)
	if ownPath == "" {
		http.Error(w, "deck not found", http.StatusNotFound)
		return
	}

	target := loadDeckSignature(h.DecksDir, owner, id, ownPath)
	if len(target.cards) == 0 {
		// No card list parsed — nothing to compare against.
		writeJSON(w, []map[string]any{})
		return
	}

	limit := 5
	if n := parseInt(r.URL.Query().Get("limit")); n > 0 && n <= 50 {
		limit = n
	}

	type scored struct {
		row   map[string]any
		score int
	}
	var ranked []scored

	owners, err := os.ReadDir(h.DecksDir)
	if err != nil {
		http.Error(w, "cannot read decks dir", http.StatusInternalServerError)
		return
	}
	for _, oent := range owners {
		if !oent.IsDir() {
			continue
		}
		ow := oent.Name()
		if ow == "freya" || ow == "benched" || ow == "test" || ow == "moxfield_300" || ow == ".versions" {
			continue
		}
		ownerDir := filepath.Join(h.DecksDir, ow)
		files, ferr := os.ReadDir(ownerDir)
		if ferr != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			n := f.Name()
			if !strings.HasSuffix(n, ".txt") && !strings.HasSuffix(n, ".json") {
				continue
			}
			otherID := strings.TrimSuffix(n, filepath.Ext(n))
			if ow == owner && otherID == id {
				continue // skip self
			}
			otherPath := filepath.Join(ownerDir, n)
			sig := loadDeckSignature(h.DecksDir, ow, otherID, otherPath)
			if len(sig.cards) == 0 {
				continue
			}
			shared := 0
			for c := range target.cards {
				if sig.cards[c] {
					shared++
				}
			}
			score := shared
			if target.commanderCard != "" && sig.commanderCard != "" &&
				strings.EqualFold(target.commanderCard, sig.commanderCard) {
				score += 30
			}
			if target.archetype != "" && sig.archetype != "" &&
				strings.EqualFold(target.archetype, sig.archetype) {
				score += 10
			}
			if target.bracket != "" && sig.bracket != "" && target.bracket == sig.bracket {
				score += 5
			}
			// Drop floor: trivially-similar decks (3 cards in common, no
			// shared commander/archetype/bracket) shouldn't pollute the
			// list. Anything with >10 shared cards OR any bonus passes.
			if shared <= 10 && score == shared {
				continue
			}

			ranked = append(ranked, scored{
				row: map[string]any{
					"owner":          ow,
					"id":             otherID,
					"name":           sig.commanderName,
					"commander":      sig.commanderName,
					"commander_card": sig.commanderCard,
					"bracket":        sig.bracket,
					"archetype":      sig.archetype,
					"shared_cards":   shared,
					"same_commander": strings.EqualFold(target.commanderCard, sig.commanderCard) && target.commanderCard != "",
					"same_archetype": strings.EqualFold(target.archetype, sig.archetype) && target.archetype != "",
					"same_bracket":   target.bracket == sig.bracket && target.bracket != "",
					"score":          score,
				},
				score: score,
			})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score != ranked[j].score {
			return ranked[i].score > ranked[j].score
		}
		// Tie-break: more shared cards wins, then alphabetical for stability.
		ai := ranked[i].row["shared_cards"].(int)
		bj := ranked[j].row["shared_cards"].(int)
		if ai != bj {
			return ai > bj
		}
		return ranked[i].row["id"].(string) < ranked[j].row["id"].(string)
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	out := make([]map[string]any, 0, len(ranked))
	for _, r := range ranked {
		out = append(out, r.row)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=120")
	_ = json.NewEncoder(w).Encode(out)
}

// deckSig is the minimal projection of a deck used for similarity scoring.
type deckSig struct {
	cards          map[string]bool // normalized lowercased card names (no commander)
	commanderName  string
	commanderCard  string
	bracket        string
	archetype      string
}

// loadDeckSignature reads a deck file and pulls just the data we need
// for similarity scoring. Reads the Freya strategy.json sidecar when
// present for archetype + canonical bracket.
func loadDeckSignature(decksDir, owner, id, deckPath string) deckSig {
	sig := deckSig{cards: map[string]bool{}}

	commander, bracket, _ := parseDeckFilename(id)
	sig.commanderName = commander
	sig.bracket = bracket
	sig.commanderCard = extractCommander(deckPath)

	data, err := os.ReadFile(deckPath)
	if err != nil {
		return sig
	}
	var raw []map[string]any
	if strings.HasSuffix(deckPath, ".json") {
		raw = parseDeckJSON(data)
	} else {
		raw = parseDeckList(string(data))
	}
	for _, c := range raw {
		name, _ := c["name"].(string)
		if name == "" {
			continue
		}
		// Drop the commander tag prefix and any set-code parens.
		name = strings.TrimPrefix(name, "COMMANDER:")
		name = strings.TrimSpace(name)
		if idx := strings.Index(name, "("); idx > 0 {
			name = strings.TrimSpace(name[:idx])
		}
		// Skip the commander card itself — every deck shares its
		// commander with itself; same-commander gets its own bonus.
		if sig.commanderCard != "" && strings.EqualFold(name, sig.commanderCard) {
			continue
		}
		sig.cards[strings.ToLower(name)] = true
	}

	// Freya sidecar — for archetype and an authoritative bracket.
	strategyFile := filepath.Join(decksDir, owner, "freya", id+".strategy.json")
	if sd, err := os.ReadFile(strategyFile); err == nil {
		var strat struct {
			Archetype string `json:"archetype"`
			Bracket   int    `json:"bracket"`
		}
		if json.Unmarshal(sd, &strat) == nil {
			if strat.Archetype != "" {
				sig.archetype = strat.Archetype
			}
			if strat.Bracket > 0 {
				sig.bracket = itoa(strat.Bracket)
			}
		}
	}
	return sig
}

func itoa(n int) string {
	// Tiny single-digit fast path so we don't pull in strconv just for bracket.
	if n >= 0 && n < 10 {
		return string(rune('0' + n))
	}
	// Fallback for >9 brackets (won't happen, but be safe).
	if n < 0 {
		return "-" + itoa(-n)
	}
	out := ""
	for n > 0 {
		out = string(rune('0'+(n%10))) + out
		n /= 10
	}
	return out
}
