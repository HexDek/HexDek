// crossref.go — Muninn cross-reference: Thor vs live-game gap analysis.
//
// Compares Thor's deterministic per-card test failures against Muninn's
// live-game memory (parser gaps, dead triggers, crashes, invariant
// violations) to identify:
//   - True positives:  cards that fail in both systems (real bugs, fix first)
//   - False negatives: cards with Muninn issues but no Thor failure (Thor blind spots)
//   - False positives: cards with Thor failures but no Muninn issue (Thor over-reporting)
//
// Usage:
//
//	result, err := RunCrossRef(thorFailures, "data/muninn")
//	writeMarkdownReport("data/crossref-report.md", result)
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hexdek/hexdek/internal/muninn"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// CrossRefEntry describes a single card that appears in one or both systems.
type CrossRefEntry struct {
	CardName    string // display name (original casing from whichever source first supplied it)
	MuninnIssue string // gap type + snippet, empty if Thor-only
	ThorIssue   string // interaction + invariant, empty if Muninn-only
}

// CrossRefResult holds the complete cross-reference analysis.
type CrossRefResult struct {
	TruePositives     []CrossRefEntry // fail both Thor and Muninn
	FalseNegatives    []CrossRefEntry // fail Muninn only (Thor missed it)
	FalsePositives    []CrossRefEntry // fail Thor only (works in games)
	TrueNegativeCount int             // cards in neither set (estimated)
	ThorTotal         int             // unique card names in Thor failures
	MuninnTotal       int             // unique card names in Muninn data
}

// ---------------------------------------------------------------------------
// Card name normalization
// ---------------------------------------------------------------------------

// normalizeCrossRef lowercases, strips punctuation, and collapses whitespace.
// This mirrors per_card.NormalizeName so that card names from different
// sources match correctly.
func normalizeCrossRef(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	b.Grow(len(name))
	prevSpace := false
	for _, r := range name {
		switch r {
		case '\'', '’', ',', '.', '!', '?', ':', ';', '-', '—', '–':
			continue
		case ' ', '\t':
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Muninn data loading
// ---------------------------------------------------------------------------

// muninnCard holds the raw display name and a summary string for one card
// found in Muninn data.
type muninnCard struct {
	displayName string
	issue       string // human-readable summary
}

// loadMuninnCards reads all Muninn JSON files from dir and returns a map of
// normalized card name -> muninnCard. If the same card appears in multiple
// sources, the issues are concatenated.
func loadMuninnCards(dir string) (map[string]muninnCard, error) {
	cards := make(map[string]muninnCard)

	// merge adds a card to the map, appending the issue if the card already exists.
	merge := func(displayName, issue string) {
		key := normalizeCrossRef(displayName)
		if key == "" {
			return
		}
		if existing, ok := cards[key]; ok {
			existing.issue += "; " + issue
			cards[key] = existing
		} else {
			cards[key] = muninnCard{
				displayName: displayName,
				issue:       issue,
			}
		}
	}

	// 1. Parser gaps — snippet field is the card name.
	gaps, err := muninn.ReadParserGaps(dir)
	if err != nil {
		return nil, fmt.Errorf("crossref: read parser_gaps: %w", err)
	}
	for _, g := range gaps {
		snippet := strings.TrimSpace(g.Snippet)
		if snippet == "" {
			continue
		}
		merge(snippet, fmt.Sprintf("parser_gap (count=%d)", g.Count))
	}

	// 2. Dead triggers — card_name field.
	triggers, err := muninn.ReadDeadTriggers(dir)
	if err != nil {
		return nil, fmt.Errorf("crossref: read dead_triggers: %w", err)
	}
	for _, dt := range triggers {
		name := strings.TrimSpace(dt.CardName)
		if name == "" {
			continue
		}
		merge(name, fmt.Sprintf("dead_trigger:%s (count=%d, games=%d)", dt.TriggerName, dt.Count, dt.GamesSeen))
	}

	// 3. Crashes — extract commander names from deck keys.
	// Deck keys look like "moxfield/nicol_bolas_the_ravager_..._b2_user_ID"
	// or "josh/oloro_lifegain_b2_ageless_ascetic". The commander name is
	// embedded but not cleanly extractable. We store the full deck key
	// for reference but don't try to normalize to a card name — crashes
	// are game-level failures, not card-level.
	crashes, err := muninn.ReadCrashLogs(dir)
	if err != nil {
		return nil, fmt.Errorf("crossref: read crashes: %w", err)
	}
	for _, c := range crashes {
		// Try to extract card-like references from the stack trace.
		// Look for function names in per_card package which contain
		// card-derived identifiers (e.g., per_card.registerNicolBolas).
		for _, line := range strings.Split(c.StackTrace, "\n") {
			if strings.Contains(line, "/per_card") {
				// Extract function name after the last /per_card. segment.
				if idx := strings.Index(line, "/per_card."); idx >= 0 {
					rest := line[idx+len("/per_card."):]
					// Take until ( or end.
					if paren := strings.IndexByte(rest, '('); paren >= 0 {
						rest = rest[:paren]
					}
					// Skip generic functions.
					lower := strings.ToLower(rest)
					if lower == "register" || lower == "init" || strings.HasPrefix(lower, "register") {
						continue
					}
					merge(rest, "crash (stack_trace)")
				}
			}
		}
	}

	// 4. Invariant violations — extract card names from violation messages.
	// Messages like "[warning] zone_accounting (seat 0): ..." don't contain
	// card names directly. But violation_type can help categorize.
	// We skip these for card-level matching since they're game-level.
	// (Invariant violations are already captured by Thor through RunAllInvariants.)

	return cards, nil
}

// ---------------------------------------------------------------------------
// Thor data loading
// ---------------------------------------------------------------------------

// thorCard holds the display name and issue summary for one Thor failure.
type thorCard struct {
	displayName string
	issue       string
}

// buildThorCards builds a map of normalized card name -> thorCard from
// Thor failure results. Multiple failures for the same card are concatenated.
func buildThorCards(failures []failure) map[string]thorCard {
	cards := make(map[string]thorCard)
	for _, f := range failures {
		key := normalizeCrossRef(f.CardName)
		if key == "" {
			continue
		}

		var issue string
		if f.Panicked {
			issue = fmt.Sprintf("PANIC@%s: %s", f.Interaction, truncate(f.PanicMsg, 80))
		} else if f.Invariant != "" {
			issue = fmt.Sprintf("%s@%s: %s", f.Invariant, f.Interaction, truncate(f.Message, 80))
		} else {
			issue = fmt.Sprintf("%s: %s", f.Interaction, truncate(f.Message, 80))
		}

		if existing, ok := cards[key]; ok {
			existing.issue += "; " + issue
			cards[key] = existing
		} else {
			cards[key] = thorCard{
				displayName: f.CardName,
				issue:       issue,
			}
		}
	}
	return cards
}

// ---------------------------------------------------------------------------
// Cross-reference engine
// ---------------------------------------------------------------------------

// RunCrossRef performs the cross-reference analysis between Thor test
// failures and Muninn live-game data.
//
// thorFailures is the slice of failure structs from a Thor run.
// muninnDir is the path to the Muninn data directory (e.g., "data/muninn").
//
// Returns a CrossRefResult with the four quadrants of the confusion matrix.
func RunCrossRef(thorFailures []failure, muninnDir string) (*CrossRefResult, error) {
	muninnCards, err := loadMuninnCards(muninnDir)
	if err != nil {
		return nil, err
	}

	thorCards := buildThorCards(thorFailures)

	result := &CrossRefResult{
		ThorTotal:   len(thorCards),
		MuninnTotal: len(muninnCards),
	}

	// True Positives: cards in BOTH sets.
	// False Negatives: cards in Muninn ONLY.
	for key, mc := range muninnCards {
		if tc, ok := thorCards[key]; ok {
			result.TruePositives = append(result.TruePositives, CrossRefEntry{
				CardName:    mc.displayName,
				MuninnIssue: mc.issue,
				ThorIssue:   tc.issue,
			})
		} else {
			result.FalseNegatives = append(result.FalseNegatives, CrossRefEntry{
				CardName:    mc.displayName,
				MuninnIssue: mc.issue,
			})
		}
	}

	// False Positives: cards in Thor ONLY.
	for key, tc := range thorCards {
		if _, ok := muninnCards[key]; !ok {
			result.FalsePositives = append(result.FalsePositives, CrossRefEntry{
				CardName: tc.displayName,
				ThorIssue: tc.issue,
			})
		}
	}

	// Sort each bucket alphabetically for stable output.
	sortEntries := func(entries []CrossRefEntry) {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].CardName < entries[j].CardName
		})
	}
	sortEntries(result.TruePositives)
	sortEntries(result.FalseNegatives)
	sortEntries(result.FalsePositives)

	return result, nil
}

// ---------------------------------------------------------------------------
// Markdown report
// ---------------------------------------------------------------------------

// WriteCrossRefReport writes the cross-reference results to a markdown file.
func WriteCrossRefReport(path string, result *CrossRefResult) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("crossref: create report: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# Thor vs Muninn Cross-Reference Report\n\n")
	fmt.Fprintf(f, "**Date:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Summary.
	fmt.Fprintf(f, "## Summary\n\n")
	fmt.Fprintf(f, "| Metric | Count |\n")
	fmt.Fprintf(f, "|--------|-------|\n")
	fmt.Fprintf(f, "| Thor unique failing cards | %d |\n", result.ThorTotal)
	fmt.Fprintf(f, "| Muninn unique card issues | %d |\n", result.MuninnTotal)
	fmt.Fprintf(f, "| True Positives (both) | %d |\n", len(result.TruePositives))
	fmt.Fprintf(f, "| False Negatives (Muninn only) | %d |\n", len(result.FalseNegatives))
	fmt.Fprintf(f, "| False Positives (Thor only) | %d |\n", len(result.FalsePositives))
	fmt.Fprintln(f)

	if result.MuninnTotal > 0 {
		recall := float64(len(result.TruePositives)) / float64(result.MuninnTotal) * 100
		fmt.Fprintf(f, "**Thor recall:** %.1f%% (of Muninn-known issues, Thor catches this fraction)\n\n", recall)
	}
	if result.ThorTotal > 0 {
		precision := float64(len(result.TruePositives)) / float64(result.ThorTotal) * 100
		fmt.Fprintf(f, "**Thor precision:** %.1f%% (of Thor failures, this fraction are real game bugs)\n\n", precision)
	}

	// True Positives.
	fmt.Fprintf(f, "## True Positives (%d) — Real bugs, fix first\n\n", len(result.TruePositives))
	if len(result.TruePositives) == 0 {
		fmt.Fprintf(f, "_No overlap found._\n\n")
	} else {
		fmt.Fprintf(f, "| Card | Muninn Issue | Thor Issue |\n")
		fmt.Fprintf(f, "|------|-------------|------------|\n")
		for _, e := range result.TruePositives {
			fmt.Fprintf(f, "| %s | %s | %s |\n",
				escMd(e.CardName),
				escMd(truncate(e.MuninnIssue, 100)),
				escMd(truncate(e.ThorIssue, 100)))
		}
		fmt.Fprintln(f)
	}

	// False Negatives.
	fmt.Fprintf(f, "## False Negatives (%d) — Thor blind spots\n\n", len(result.FalseNegatives))
	fmt.Fprintf(f, "Cards with Muninn-detected issues that Thor does NOT catch.\n")
	fmt.Fprintf(f, "These indicate gaps in Thor's test coverage.\n\n")
	if len(result.FalseNegatives) == 0 {
		fmt.Fprintf(f, "_Thor catches everything Muninn sees._\n\n")
	} else {
		fmt.Fprintf(f, "| Card | Muninn Issue |\n")
		fmt.Fprintf(f, "|------|--------------|\n")
		for _, e := range result.FalseNegatives {
			fmt.Fprintf(f, "| %s | %s |\n",
				escMd(e.CardName),
				escMd(truncate(e.MuninnIssue, 120)))
		}
		fmt.Fprintln(f)
	}

	// False Positives.
	fmt.Fprintf(f, "## False Positives (%d) — Thor over-reporting\n\n", len(result.FalsePositives))
	fmt.Fprintf(f, "Cards that fail in Thor but work fine in live games.\n")
	fmt.Fprintf(f, "These may be test harness artifacts or edge cases that never occur in practice.\n\n")
	if len(result.FalsePositives) == 0 {
		fmt.Fprintf(f, "_All Thor failures are confirmed by Muninn._\n\n")
	} else {
		fmt.Fprintf(f, "| Card | Thor Issue |\n")
		fmt.Fprintf(f, "|------|------------|\n")
		limit := len(result.FalsePositives)
		if limit > 200 {
			limit = 200
		}
		for i := 0; i < limit; i++ {
			e := result.FalsePositives[i]
			fmt.Fprintf(f, "| %s | %s |\n",
				escMd(e.CardName),
				escMd(truncate(e.ThorIssue, 120)))
		}
		if len(result.FalsePositives) > 200 {
			fmt.Fprintf(f, "\n_... and %d more_\n", len(result.FalsePositives)-200)
		}
		fmt.Fprintln(f)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// truncate shortens a string to at most maxLen characters, appending "..."
// if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// escMd escapes pipe characters in markdown table cells.
func escMd(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
