package hexapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CardSearchHit is a row in the /api/cards/search payload.
type CardSearchHit struct {
	Name     string  `json:"name"`
	ManaCost string  `json:"mana_cost"`
	TypeLine string  `json:"type_line"`
	CMC      float64 `json:"cmc"`
	Set      string  `json:"set,omitempty"`
	ImageURI string  `json:"image_uri"`
	Score    int     `json:"score,omitempty"`
}

// CardDetail is the response for /api/cards/{name}.
type CardDetail struct {
	Name       string         `json:"name"`
	ManaCost   string         `json:"mana_cost"`
	TypeLine   string         `json:"type_line"`
	OracleText string         `json:"oracle_text"`
	CMC        float64        `json:"cmc"`
	Set        string         `json:"set,omitempty"`
	ImageURI   string         `json:"image_uri"`
	// Printings is empty when the loaded corpus is the oracle-cards bulk
	// (one entry per name). Switch the loader to default-cards.json to
	// populate this with per-set printings.
	Printings  []CardPrinting `json:"printings"`
	DecksUsing []DeckRef      `json:"decks_using"`
}

type CardPrinting struct {
	Set      string `json:"set"`
	SetName  string `json:"set_name,omitempty"`
	Released string `json:"released_at,omitempty"`
	ImageURI string `json:"image_uri,omitempty"`
}

type DeckRef struct {
	Owner     string `json:"owner"`
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Commander string `json:"commander,omitempty"`
	// Role is the Freya card-role tag (ramp/draw/removal/combo/...) if a
	// strategy.json with card_roles exists for this deck. Empty string
	// means the deck has not been analyzed or the role is unset.
	Role string `json:"role,omitempty"`
}

// cardImageURI returns the relative URL the frontend uses to fetch art
// for a card. It mirrors the existing /api/card-art/{name} contract,
// which 302s to Scryfall's named-fuzzy art_crop image. Frontend code
// can prepend the API base or rely on same-origin deployment.
func cardImageURI(name string) string {
	return "/api/card-art/" + url.PathEscape(name)
}

// handleCardSearch implements GET /api/cards/search?q=<query>&limit=<n>.
// Searches the loaded oracle card corpus by name with prefix/substring
// fuzzy match, returning the top results sorted by relevance.
//
// Defaults: limit=20, max=50. Empty `q` returns an empty list.
func (h *Handler) handleCardSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" || len(h.cardDB) == 0 {
		writeJSON(w, map[string]any{"query": q, "results": []CardSearchHit{}})
		return
	}
	needle := strings.ToLower(q)

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n := parseInt(v); n > 0 {
			limit = n
		}
	}
	if limit > 50 {
		limit = 50
	}

	hits := make([]CardSearchHit, 0, 64)
	for nameLower, info := range h.cardDB {
		score := matchScore(nameLower, needle)
		if score == 0 {
			continue
		}
		display := titleCaseCardName(nameLower)
		hits = append(hits, CardSearchHit{
			Name:     display,
			ManaCost: info.ManaCost,
			TypeLine: info.TypeLine,
			CMC:      info.CMC,
			Set:      info.Set,
			ImageURI: cardImageURI(display),
			Score:    score,
		})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Name < hits[j].Name
	})

	if len(hits) > limit {
		hits = hits[:limit]
	}

	writeJSON(w, map[string]any{
		"query":   q,
		"count":   len(hits),
		"results": hits,
	})
}

// handleCardByName implements GET /api/cards/{name}.
//
// Returns full card data plus a cross-reference of decks containing the
// card, with each deck's Freya role assignment (when strategy.json is
// available). 404 when the card isn't in the loaded corpus.
func (h *Handler) handleCardByName(w http.ResponseWriter, r *http.Request) {
	raw := r.PathValue("name")
	name, err := url.PathUnescape(raw)
	if err != nil || name == "" {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	key := strings.ToLower(strings.TrimSpace(name))
	info, ok := h.cardDB[key]
	if !ok {
		http.Error(w, "card not found", http.StatusNotFound)
		return
	}

	display := titleCaseCardName(key)
	detail := CardDetail{
		Name:       display,
		ManaCost:   info.ManaCost,
		TypeLine:   info.TypeLine,
		OracleText: info.OracleText,
		CMC:        info.CMC,
		Set:        info.Set,
		ImageURI:   cardImageURI(display),
		Printings:  []CardPrinting{},
		DecksUsing: h.findDecksUsingCard(key),
	}
	if info.Set != "" {
		detail.Printings = append(detail.Printings, CardPrinting{
			Set:      info.Set,
			ImageURI: detail.ImageURI,
		})
	}
	writeJSON(w, detail)
}

// findDecksUsingCard walks DecksDir for any deck whose card list contains
// `cardKey` (already lowercased). For each match it returns the Freya
// card-role assignment for that card if a strategy.json exists.
//
// Substring scan is intentionally cheap — read each deck file once,
// case-fold compare, and skip JSON-parse for non-matches. With a few
// hundred decks this stays well under 50ms even on cold cache.
func (h *Handler) findDecksUsingCard(cardKey string) []DeckRef {
	if h.DecksDir == "" || cardKey == "" {
		return []DeckRef{}
	}
	entries, err := os.ReadDir(h.DecksDir)
	if err != nil {
		return []DeckRef{}
	}

	out := make([]DeckRef, 0, 16)
	for _, ownerEntry := range entries {
		if !ownerEntry.IsDir() {
			continue
		}
		owner := ownerEntry.Name()
		// Skip server-internal directories the same way the rest of the
		// API does (see search.go).
		switch owner {
		case "freya", "benched", "test", "moxfield_300", ".versions":
			continue
		}

		deckDir := filepath.Join(h.DecksDir, owner)
		files, err := os.ReadDir(deckDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			fname := f.Name()
			ext := filepath.Ext(fname)
			if ext != ".txt" && ext != ".json" {
				continue
			}
			id := strings.TrimSuffix(fname, ext)

			data, err := os.ReadFile(filepath.Join(deckDir, fname))
			if err != nil {
				continue
			}
			// Quick reject: case-folded haystack must contain the card.
			lower := strings.ToLower(string(data))
			if !strings.Contains(lower, cardKey) {
				continue
			}
			// Confirm via parser so we don't hit on unrelated text in
			// the file (e.g. a similarly-spelled commander name in a
			// JSON metadata field).
			var cards []map[string]any
			if ext == ".json" {
				cards = parseDeckJSON(data)
			} else {
				cards = parseDeckList(string(data))
			}
			if !parsedDeckHasCard(cards, cardKey) {
				continue
			}

			displayName, _, _, cmdrCard := resolveDeckMetadata(h.DecksDir, owner, id, filepath.Join(deckDir, fname))
			role := lookupCardRole(h.DecksDir, owner, id, cardKey)
			out = append(out, DeckRef{
				Owner:     owner,
				ID:        id,
				Name:      displayName,
				Commander: cmdrCard,
				Role:      role,
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Owner != out[j].Owner {
			return out[i].Owner < out[j].Owner
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// parsedDeckHasCard returns true if the parsed deck contains a card whose
// name (with parenthetical set markers stripped) matches cardKey.
func parsedDeckHasCard(cards []map[string]any, cardKey string) bool {
	for _, c := range cards {
		raw, _ := c["name"].(string)
		clean := strings.ToLower(raw)
		if idx := strings.Index(clean, "("); idx > 0 {
			clean = strings.TrimSpace(clean[:idx])
		}
		if clean == cardKey {
			return true
		}
	}
	return false
}

// lookupCardRole reads <decksDir>/<owner>/freya/<id>.strategy.json and
// returns the card_roles entry for cardKey (case-folded match). Empty
// string when strategy.json is missing or the card has no role.
func lookupCardRole(decksDir, owner, id, cardKey string) string {
	p := filepath.Join(decksDir, owner, "freya", id+".strategy.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	var s struct {
		CardRoles map[string]string `json:"card_roles"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return ""
	}
	for k, v := range s.CardRoles {
		if strings.ToLower(k) == cardKey {
			return v
		}
	}
	return ""
}
