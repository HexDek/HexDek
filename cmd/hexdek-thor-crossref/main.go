// hexdek-thor-crossref — Cross-reference Thor failures against Muninn live-game gaps.
//
// Thor finds failures in synthetic isolation. Muninn logs failures from live
// tournament games. These sources don't always agree:
//   - Some cards fail Thor but work fine in real games (false positives — Thor's
//     setup is wrong).
//   - Some cards work in Thor but fail in real games (false negatives — Thor
//     doesn't test the right interaction).
//
// This tool diffs the two sources and produces three lists:
//   - Both fail (confirmed bugs)
//   - Thor-only (likely test-harness issues)
//   - Muninn-only (Thor blind spots)
//
// Usage:
//
//	hexdek-thor-crossref \
//	  --thor-report data/muninn-gap-thor-report.md \
//	  --muninn-cards data/muninn-gap-cards.txt \
//	  --muninn-dir data/muninn \
//	  --output data/thor-muninn-crossref.md
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hexdek/hexdek/internal/muninn"
)

// thorFailure is a single row parsed from a Thor markdown report's
// "## Invariant Violations" table.
type thorFailure struct {
	Card        string
	Interaction string
	Invariant   string
	Message     string
}

// muninnContext aggregates the Muninn-side signals for a single card.
type muninnContext struct {
	ParserGaps          int // count from parser_gaps.json
	DeadTriggers        int // count from dead_triggers.json
	InvariantViolations int // count from invariant_violations.json (matched by message text)
}

func (c muninnContext) summary() string {
	parts := make([]string, 0, 3)
	if c.ParserGaps > 0 {
		parts = append(parts, fmt.Sprintf("parser_gaps=%d", c.ParserGaps))
	}
	if c.DeadTriggers > 0 {
		parts = append(parts, fmt.Sprintf("dead_triggers=%d", c.DeadTriggers))
	}
	if c.InvariantViolations > 0 {
		parts = append(parts, fmt.Sprintf("invariants=%d", c.InvariantViolations))
	}
	if len(parts) == 0 {
		return "(listed gap, no JSON detail)"
	}
	return strings.Join(parts, ", ")
}

func main() {
	var (
		thorReport   = flag.String("thor-report", "data/muninn-gap-thor-report.md", "path to Thor markdown report")
		muninnCards  = flag.String("muninn-cards", "data/muninn-gap-cards.txt", "path to Muninn gap card list (one per line)")
		muninnDir    = flag.String("muninn-dir", "data/muninn", "Muninn data directory (JSON files)")
		outputPath   = flag.String("output", "data/thor-muninn-crossref.md", "output markdown path")
	)
	flag.Parse()

	thorFailures, err := parseThorReport(*thorReport)
	if err != nil {
		fatalf("read thor report: %v", err)
	}
	if len(thorFailures) == 0 {
		fmt.Fprintf(os.Stderr, "warning: no failures parsed from %s\n", *thorReport)
	}

	muninnList, err := readCardList(*muninnCards)
	if err != nil {
		fatalf("read muninn card list: %v", err)
	}

	contexts, err := loadMuninnContexts(*muninnDir, muninnList)
	if err != nil {
		fatalf("load muninn JSON: %v", err)
	}

	report := buildReport(thorFailures, muninnList, contexts)

	if err := os.WriteFile(*outputPath, []byte(report), 0o644); err != nil {
		fatalf("write output: %v", err)
	}

	fmt.Printf("wrote %s\n", *outputPath)
}

// ----------------------------------------------------------------------------
// Parsing
// ----------------------------------------------------------------------------

// parseThorReport extracts failure rows from a Thor markdown report. It scans
// for tables under headings beginning with "## Invariant Violations" and
// collects rows of the form: | Card | Interaction | Invariant | Message |.
func parseThorReport(path string) ([]thorFailure, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var failures []thorFailure
	inViolations := false
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			inViolations = strings.Contains(strings.ToLower(trimmed), "invariant violations") ||
				strings.Contains(strings.ToLower(trimmed), "failures")
			continue
		}
		if !inViolations {
			continue
		}
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		// Skip header and separator rows.
		if strings.Contains(trimmed, "Card") && strings.Contains(trimmed, "Interaction") {
			continue
		}
		if strings.Contains(trimmed, "---") {
			continue
		}
		// Skip the "Failures by Interaction" summary table (no Card column).
		cells := splitMarkdownRow(trimmed)
		if len(cells) < 4 {
			continue
		}
		card := strings.TrimSpace(cells[0])
		if card == "" {
			continue
		}
		failures = append(failures, thorFailure{
			Card:        card,
			Interaction: strings.TrimSpace(cells[1]),
			Invariant:   strings.TrimSpace(cells[2]),
			Message:     strings.TrimSpace(cells[3]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return failures, nil
}

func splitMarkdownRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	parts := strings.Split(row, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func readCardList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name == "" || strings.HasPrefix(name, "#") {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out, scanner.Err()
}

// ----------------------------------------------------------------------------
// Muninn JSON aggregation
// ----------------------------------------------------------------------------

// loadMuninnContexts reads the Muninn JSON files and aggregates per-card
// signals. Card-name detection in invariant_violations.json is best-effort:
// the file doesn't carry a card column directly, so we substring-match the
// message against the card list.
func loadMuninnContexts(dir string, cardList []string) (map[string]*muninnContext, error) {
	contexts := make(map[string]*muninnContext)
	getCtx := func(name string) *muninnContext {
		c, ok := contexts[name]
		if !ok {
			c = &muninnContext{}
			contexts[name] = c
		}
		return c
	}

	gaps, err := muninn.ReadParserGaps(dir)
	if err != nil {
		return nil, fmt.Errorf("parser_gaps: %w", err)
	}
	for _, g := range gaps {
		// The snippet often is a card name verbatim, but can also be an
		// oracle text excerpt. Try direct match first, then substring.
		if name := matchCard(g.Snippet, cardList); name != "" {
			getCtx(name).ParserGaps += g.Count
		}
	}

	triggers, err := muninn.ReadDeadTriggers(dir)
	if err != nil {
		return nil, fmt.Errorf("dead_triggers: %w", err)
	}
	for _, t := range triggers {
		if t.CardName == "" {
			continue
		}
		getCtx(t.CardName).DeadTriggers += t.Count
	}

	violations, err := muninn.ReadInvariantViolations(dir)
	if err != nil {
		return nil, fmt.Errorf("invariant_violations: %w", err)
	}
	for _, v := range violations {
		if name := matchCard(v.Message, cardList); name != "" {
			getCtx(name).InvariantViolations++
		}
	}

	return contexts, nil
}

// matchCard returns the card name from cardList that appears in the haystack,
// preferring exact match, then case-insensitive substring. Returns "" if none.
func matchCard(haystack string, cardList []string) string {
	for _, name := range cardList {
		if name == haystack {
			return name
		}
	}
	hLower := strings.ToLower(haystack)
	for _, name := range cardList {
		if strings.Contains(hLower, strings.ToLower(name)) {
			return name
		}
	}
	return ""
}

// ----------------------------------------------------------------------------
// Cross-reference + report
// ----------------------------------------------------------------------------

func buildReport(thorFailures []thorFailure, muninnList []string, contexts map[string]*muninnContext) string {
	muninnSet := make(map[string]struct{}, len(muninnList))
	for _, name := range muninnList {
		muninnSet[name] = struct{}{}
	}

	// Group Thor failures by card.
	thorByCard := make(map[string][]thorFailure)
	for _, f := range thorFailures {
		thorByCard[f.Card] = append(thorByCard[f.Card], f)
	}

	// Categorize.
	var bothFail, thorOnly, muninnOnly []string

	for card := range thorByCard {
		if _, ok := muninnSet[card]; ok {
			bothFail = append(bothFail, card)
		} else {
			thorOnly = append(thorOnly, card)
		}
	}
	for _, card := range muninnList {
		if _, ok := thorByCard[card]; !ok {
			muninnOnly = append(muninnOnly, card)
		}
	}
	sort.Strings(bothFail)
	sort.Strings(thorOnly)
	// muninnOnly preserves the order of muninn-gap-cards.txt for stability,
	// then we sort to keep the report deterministic.
	sort.Strings(muninnOnly)

	var b strings.Builder
	fmt.Fprintf(&b, "# Thor ↔ Muninn Cross-Reference — %s\n\n", time.Now().Format("2006-01-02"))
	fmt.Fprintf(&b, "**Thor failures parsed:** %d (across %d unique cards)  \n", len(thorFailures), len(thorByCard))
	fmt.Fprintf(&b, "**Muninn gap cards:** %d  \n", len(muninnList))
	fmt.Fprintf(&b, "**Both fail (confirmed bugs):** %d  \n", len(bothFail))
	fmt.Fprintf(&b, "**Thor-only (likely harness issues):** %d  \n", len(thorOnly))
	fmt.Fprintf(&b, "**Muninn-only (Thor blind spots):** %d\n\n", len(muninnOnly))

	// --- Both fail ---
	fmt.Fprintf(&b, "## Both Fail (confirmed bugs) — %d\n\n", len(bothFail))
	if len(bothFail) == 0 {
		b.WriteString("_None._\n\n")
	} else {
		b.WriteString("| Card | Thor Failure | Muninn Context |\n")
		b.WriteString("|------|--------------|----------------|\n")
		for _, card := range bothFail {
			thor := summarizeThorFailures(thorByCard[card])
			ctx := summarizeMuninnContext(contexts[card])
			fmt.Fprintf(&b, "| %s | %s | %s |\n", escapePipe(card), escapePipe(thor), escapePipe(ctx))
		}
		b.WriteString("\n")
	}

	// --- Thor-only ---
	fmt.Fprintf(&b, "## Thor-Only (likely harness issues) — %d\n\n", len(thorOnly))
	if len(thorOnly) == 0 {
		b.WriteString("_None._\n\n")
	} else {
		b.WriteString("| Card | Thor Failure | Notes |\n")
		b.WriteString("|------|--------------|-------|\n")
		for _, card := range thorOnly {
			thor := summarizeThorFailures(thorByCard[card])
			fmt.Fprintf(&b, "| %s | %s | not in Muninn gap list — investigate harness setup |\n",
				escapePipe(card), escapePipe(thor))
		}
		b.WriteString("\n")
	}

	// --- Muninn-only ---
	fmt.Fprintf(&b, "## Muninn-Only (Thor blind spots) — %d\n\n", len(muninnOnly))
	if len(muninnOnly) == 0 {
		b.WriteString("_None._\n\n")
	} else {
		b.WriteString("| Card | Muninn Gap | Why Thor Misses It |\n")
		b.WriteString("|------|------------|--------------------|\n")
		for _, card := range muninnOnly {
			ctx := summarizeMuninnContext(contexts[card])
			fmt.Fprintf(&b, "| %s | %s | passes Thor — interaction not exercised by current battery |\n",
				escapePipe(card), escapePipe(ctx))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func summarizeThorFailures(fs []thorFailure) string {
	// Collapse to "interaction (xN); interaction (xN)" plus a representative message.
	counts := make(map[string]int)
	order := []string{}
	for _, f := range fs {
		if _, ok := counts[f.Interaction]; !ok {
			order = append(order, f.Interaction)
		}
		counts[f.Interaction]++
	}
	parts := make([]string, 0, len(order))
	for _, k := range order {
		if counts[k] == 1 {
			parts = append(parts, k)
		} else {
			parts = append(parts, fmt.Sprintf("%s (x%d)", k, counts[k]))
		}
	}
	// Append shortened first message for context.
	msg := ""
	if len(fs) > 0 {
		msg = trimMessage(fs[0].Message)
	}
	if msg != "" {
		return strings.Join(parts, "; ") + " — " + msg
	}
	return strings.Join(parts, "; ")
}

var wsRe = regexp.MustCompile(`\s+`)

func trimMessage(m string) string {
	m = wsRe.ReplaceAllString(m, " ")
	const max = 80
	if len(m) > max {
		return m[:max-1] + "…"
	}
	return m
}

func summarizeMuninnContext(c *muninnContext) string {
	if c == nil {
		return "(listed gap, no JSON detail)"
	}
	return c.summary()
}

func escapePipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "hexdek-thor-crossref: "+format+"\n", args...)
	os.Exit(1)
}
