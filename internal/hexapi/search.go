package hexapi

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SearchResult is one item in the universal-search dropdown.
//
// Kind discriminates the navigation behavior on the frontend:
//
//	"deck"      → navigate to /decks/{owner}/{id}
//	"commander" → navigate to /decks?tab=all&q={label}
//	"owner"     → navigate to /decks?tab=all&q={label}
//	"card"      → no navigation (informational); frontend may show a hint
type SearchResult struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Sub   string `json:"sub,omitempty"`
	Owner string `json:"owner,omitempty"`
	ID    string `json:"id,omitempty"`
	Score int    `json:"score,omitempty"`
}

// handleSearch implements GET /api/search?q=<query>&limit=<n>.
// Searches across deck names, commander names, deck owners (players), and
// — when the oracle card DB is loaded — card names. Returns up to `limit`
// results per kind (default 6, max 25).
func (h *Handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, map[string]any{"query": "", "results": []SearchResult{}})
		return
	}
	needle := strings.ToLower(q)

	limit := 6
	if v := r.URL.Query().Get("limit"); v != "" {
		var n int
		for _, c := range v {
			if c < '0' || c > '9' {
				n = 0
				break
			}
			n = n*10 + int(c-'0')
		}
		if n > 0 && n <= 25 {
			limit = n
		}
	}

	deckHits := []SearchResult{}
	commanderHits := []SearchResult{}
	ownerSet := map[string]int{}

	owners, err := os.ReadDir(h.DecksDir)
	if err == nil {
		for _, ownerEntry := range owners {
			if !ownerEntry.IsDir() {
				continue
			}
			owner := ownerEntry.Name()
			if owner == "freya" || owner == "benched" || owner == "test" ||
				owner == "moxfield_300" || owner == ".versions" {
				continue
			}
			ownerLower := strings.ToLower(owner)
			if score := matchScore(ownerLower, needle); score > 0 {
				if cur, ok := ownerSet[owner]; !ok || score > cur {
					ownerSet[owner] = score
				}
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
				if !strings.HasSuffix(fname, ".txt") && !strings.HasSuffix(fname, ".json") {
					continue
				}
				id := strings.TrimSuffix(fname, filepath.Ext(fname))
				name, _, _ := parseDeckFilename(id)
				cmdrCard := extractCommander(filepath.Join(deckDir, fname))

				deckScore := bestScore(needle, strings.ToLower(name), strings.ToLower(id))
				if deckScore > 0 {
					sub := owner
					if cmdrCard != "" {
						sub = owner + " · " + cmdrCard
					}
					deckHits = append(deckHits, SearchResult{
						Kind:  "deck",
						Label: name,
						Sub:   sub,
						Owner: owner,
						ID:    id,
						Score: deckScore,
					})
				}

				if cmdrCard != "" {
					if score := matchScore(strings.ToLower(cmdrCard), needle); score > 0 {
						commanderHits = append(commanderHits, SearchResult{
							Kind:  "commander",
							Label: cmdrCard,
							Sub:   owner + "/" + id,
							Owner: owner,
							ID:    id,
							Score: score,
						})
					}
				}
			}
		}
	}

	dedupCommanders(&commanderHits)

	ownerHits := make([]SearchResult, 0, len(ownerSet))
	for o, score := range ownerSet {
		ownerHits = append(ownerHits, SearchResult{
			Kind:  "owner",
			Label: o,
			Sub:   "PLAYER",
			Score: score,
		})
	}

	cardHits := []SearchResult{}
	if len(h.cardDB) > 0 {
		for cardLower, info := range h.cardDB {
			if score := matchScore(cardLower, needle); score > 0 {
				cardHits = append(cardHits, SearchResult{
					Kind:  "card",
					Label: titleCaseCardName(cardLower),
					Sub:   strings.TrimSpace(info.TypeLine),
					Score: score,
				})
				if len(cardHits) >= 200 {
					// Cap initial sweep — we re-sort and trim below.
					break
				}
			}
		}
	}

	sortByScoreThenLabel(deckHits)
	sortByScoreThenLabel(commanderHits)
	sortByScoreThenLabel(ownerHits)
	sortByScoreThenLabel(cardHits)

	if len(deckHits) > limit {
		deckHits = deckHits[:limit]
	}
	if len(commanderHits) > limit {
		commanderHits = commanderHits[:limit]
	}
	if len(ownerHits) > limit {
		ownerHits = ownerHits[:limit]
	}
	if len(cardHits) > limit {
		cardHits = cardHits[:limit]
	}

	combined := make([]SearchResult, 0, len(deckHits)+len(commanderHits)+len(ownerHits)+len(cardHits))
	combined = append(combined, ownerHits...)
	combined = append(combined, commanderHits...)
	combined = append(combined, deckHits...)
	combined = append(combined, cardHits...)

	writeJSON(w, map[string]any{
		"query": q,
		"results": map[string]any{
			"decks":      deckHits,
			"commanders": commanderHits,
			"owners":     ownerHits,
			"cards":      cardHits,
		},
		"top": combined,
	})
}

// matchScore returns a relevance score for `needle` inside `hay`.
// Higher is better. Returns 0 for no match.
//
// Heuristic ranking:
//
//	100  exact match
//	 80  prefix match
//	 60  word-boundary match (token after space/dash/underscore)
//	 30  substring match
func matchScore(hay, needle string) int {
	if hay == "" || needle == "" {
		return 0
	}
	if hay == needle {
		return 100
	}
	if strings.HasPrefix(hay, needle) {
		return 80
	}
	for _, sep := range []string{" ", "-", "_", ",", "/"} {
		if strings.Contains(hay, sep+needle) {
			return 60
		}
	}
	if strings.Contains(hay, needle) {
		return 30
	}
	return 0
}

func bestScore(needle string, hays ...string) int {
	best := 0
	for _, h := range hays {
		if s := matchScore(h, needle); s > best {
			best = s
		}
	}
	return best
}

func sortByScoreThenLabel(items []SearchResult) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		return items[i].Label < items[j].Label
	})
}

// dedupCommanders collapses commander results that share a label, keeping
// the highest-scoring entry. Commanders appear once per deck on disk; the
// dropdown should show each commander once.
func dedupCommanders(hits *[]SearchResult) {
	if len(*hits) == 0 {
		return
	}
	seen := map[string]int{}
	out := (*hits)[:0]
	for _, h := range *hits {
		key := strings.ToLower(h.Label)
		if idx, ok := seen[key]; ok {
			if h.Score > out[idx].Score {
				out[idx] = h
			}
			continue
		}
		seen[key] = len(out)
		out = append(out, h)
	}
	*hits = out
}

// titleCaseCardName converts a lowercased oracle card name back to a
// reasonable display form. The cardDB stores names lowercased for lookup;
// we don't carry the original casing alongside, so we approximate by
// upper-casing word starts. Commas/apostrophes/hyphens preserved.
func titleCaseCardName(s string) string {
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
