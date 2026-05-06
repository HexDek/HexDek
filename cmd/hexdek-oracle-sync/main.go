// hexdek-oracle-sync — Pull fresh Scryfall oracle bulk data, diff against
// the local snapshot, re-parse changed cards, and run Thor on just that
// subset.
//
// The local source of truth is data/rules/oracle-cards.json. The fresh
// download lands at data/rules/oracle-cards-new.json and is only swapped
// in when --promote is passed. The diff report is always written to
// data/rules/oracle-sync-report.md.
//
// Usage:
//
//	go run ./cmd/hexdek-oracle-sync/                 # download + diff + parse + Thor
//	go run ./cmd/hexdek-oracle-sync/ --dry-run       # download + diff only
//	go run ./cmd/hexdek-oracle-sync/ --promote       # also replace the snapshot
//	go run ./cmd/hexdek-oracle-sync/ --verbose       # print changed cards
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	scryfallBulkAPI   = "https://api.scryfall.com/bulk-data"
	userAgent         = "HexDek-OracleSync/1.0"
	scryfallRateLimit = 100 * time.Millisecond
	httpTimeout       = 10 * time.Minute
)

type bulkInfo struct {
	Type        string `json:"type"`
	DownloadURI string `json:"download_uri"`
	UpdatedAt   string `json:"updated_at"`
	Size        int64  `json:"size"`
}

type bulkResp struct {
	Data []bulkInfo `json:"data"`
}

// card captures the fields we diff. Anything else passes through inside
// the raw payload we keep alongside.
type card struct {
	OracleID   string     `json:"oracle_id"`
	Name       string     `json:"name"`
	OracleText string     `json:"oracle_text"`
	TypeLine   string     `json:"type_line"`
	ManaCost   string     `json:"mana_cost"`
	Layout     string     `json:"layout"`
	CardFaces  []cardFace `json:"card_faces,omitempty"`
}

type cardFace struct {
	Name       string `json:"name"`
	OracleText string `json:"oracle_text"`
	TypeLine   string `json:"type_line"`
	ManaCost   string `json:"mana_cost"`
}

type fieldChange struct {
	Field string
	Old   string
	New   string
}

type cardChange struct {
	OracleID string
	Name     string
	Fields   []fieldChange
	OldRaw   json.RawMessage
	NewRaw   json.RawMessage
}

func main() {
	dryRun := flag.Bool("dry-run", false, "download and diff only — skip re-parse and Thor")
	promote := flag.Bool("promote", false, "after validation, replace the local oracle-cards.json with the fresh download")
	verbose := flag.Bool("verbose", false, "print changed cards to stdout")
	outputDir := flag.String("output", "data/rules", "directory holding oracle-cards.json")
	skipDownload := flag.Bool("skip-download", false, "reuse oracle-cards-new.json from a previous run instead of re-downloading")
	skipThor := flag.Bool("skip-thor", false, "diff and re-parse only — skip the Thor subset run")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx, *outputDir, *dryRun, *promote, *verbose, *skipDownload, *skipThor); err != nil {
		log.Fatalf("oracle-sync: %v", err)
	}
}

func run(ctx context.Context, outputDir string, dryRun, promote, verbose, skipDownload, skipThor bool) error {
	currentPath := filepath.Join(outputDir, "oracle-cards.json")
	newPath := filepath.Join(outputDir, "oracle-cards-new.json")
	reportPath := filepath.Join(outputDir, "oracle-sync-report.md")

	if _, err := os.Stat(currentPath); err != nil {
		return fmt.Errorf("read current snapshot %s: %w", currentPath, err)
	}

	// ─── Step 1: pull bulk data ────────────────────────────────────────
	if skipDownload {
		if _, err := os.Stat(newPath); err != nil {
			return fmt.Errorf("--skip-download set but %s is missing: %w", newPath, err)
		}
		log.Printf("skipping download — reusing %s", newPath)
	} else {
		info, err := fetchBulkInfo(ctx)
		if err != nil {
			return fmt.Errorf("scryfall bulk-data lookup: %w", err)
		}
		log.Printf("scryfall: oracle_cards bulk dated %s, ~%s", info.UpdatedAt, humanBytes(info.Size))
		log.Printf("downloading %s → %s", info.DownloadURI, newPath)
		// Honour Scryfall's 100ms gap before the second hit.
		time.Sleep(scryfallRateLimit)
		if err := download(ctx, info.DownloadURI, newPath); err != nil {
			return fmt.Errorf("download bulk: %w", err)
		}
	}

	// ─── Step 2: diff ──────────────────────────────────────────────────
	log.Printf("loading current snapshot %s", currentPath)
	oldCards, err := loadOracle(currentPath)
	if err != nil {
		return fmt.Errorf("load current: %w", err)
	}
	log.Printf("  %d cards (current)", len(oldCards))

	log.Printf("loading fresh download %s", newPath)
	newCards, err := loadOracle(newPath)
	if err != nil {
		return fmt.Errorf("load new: %w", err)
	}
	log.Printf("  %d cards (new)", len(newCards))

	changes := diffCards(oldCards, newCards)
	log.Printf("changes detected: %d cards", len(changes))
	if verbose {
		for _, ch := range changes {
			fmt.Printf("  %s\n", ch.Name)
			for _, f := range ch.Fields {
				fmt.Printf("    %s: %q → %q\n", f.Field, ch.truncate(f.Old), ch.truncate(f.New))
			}
		}
	}

	// ─── Step 3 & 4: parse + Thor (skipped on --dry-run) ───────────────
	var (
		parseImpacts []parseImpact
		thorReport   string
	)
	if !dryRun && len(changes) > 0 {
		log.Printf("re-parsing %d changed cards via scripts/parser.py", len(changes))
		parseImpacts, err = reparseChanged(ctx, changes)
		if err != nil {
			log.Printf("WARNING: re-parse failed: %v", err)
		} else {
			parseDelta := 0
			for _, p := range parseImpacts {
				if p.Changed {
					parseDelta++
				}
			}
			log.Printf("  parse-result changes: %d", parseDelta)
		}

		if skipThor {
			log.Printf("skipping Thor run (--skip-thor)")
			thorReport = "_Skipped (--skip-thor)._"
		} else {
			log.Printf("running Thor on the %d-card subset", len(changes))
			thorReport, err = runThor(ctx, changes, outputDir)
			if err != nil {
				log.Printf("WARNING: Thor run failed: %v", err)
				thorReport = fmt.Sprintf("Thor run failed: %v", err)
			}
		}
	}

	// ─── Step 7: write report ──────────────────────────────────────────
	if err := writeReport(reportPath, len(newCards), changes, parseImpacts, thorReport, dryRun); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	log.Printf("report: %s", reportPath)

	// ─── Step 5: promote ───────────────────────────────────────────────
	if promote {
		if err := os.Rename(newPath, currentPath); err != nil {
			return fmt.Errorf("promote: rename %s → %s: %w", newPath, currentPath, err)
		}
		log.Printf("promoted: %s now reflects the fresh download", currentPath)
	}

	return nil
}

// ────────────────────────────────────────────────────────────────────────
// Scryfall bulk download
// ────────────────────────────────────────────────────────────────────────

func fetchBulkInfo(ctx context.Context) (*bulkInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", scryfallBulkAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("scryfall returned %s", resp.Status)
	}

	var br bulkResp
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return nil, fmt.Errorf("decode bulk index: %w", err)
	}
	for _, e := range br.Data {
		if e.Type == "oracle_cards" {
			return &e, nil
		}
	}
	return nil, errors.New("oracle_cards entry not found in bulk index")
}

func download(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("http %s", resp.Status)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	bw := bufio.NewWriterSize(f, 1<<20)
	if _, err := io.Copy(bw, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := bw.Flush(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dest)
}

// ────────────────────────────────────────────────────────────────────────
// Loading + diffing
// ────────────────────────────────────────────────────────────────────────

type loadedCard struct {
	c   card
	raw json.RawMessage
}

func loadOracle(path string) (map[string]loadedCard, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(bufio.NewReaderSize(f, 1<<20))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("read opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, fmt.Errorf("expected JSON array, got %v", tok)
	}

	out := make(map[string]loadedCard, 32_000)
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, fmt.Errorf("decode card: %w", err)
		}
		var c card
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, fmt.Errorf("unmarshal card: %w", err)
		}
		if c.OracleID == "" {
			// tokens, art cards, etc — skip silently.
			continue
		}
		// Multiple printings share an oracle_id. First one wins; oracle_cards
		// bulk is already deduplicated but be defensive.
		if _, dup := out[c.OracleID]; dup {
			continue
		}
		out[c.OracleID] = loadedCard{c: c, raw: raw}
	}
	return out, nil
}

func diffCards(oldCards, newCards map[string]loadedCard) []cardChange {
	var changes []cardChange
	for oid, nc := range newCards {
		oc, ok := oldCards[oid]
		if !ok {
			changes = append(changes, cardChange{
				OracleID: oid,
				Name:     nc.c.Name,
				Fields:   []fieldChange{{Field: "added", Old: "", New: nc.c.Name}},
				NewRaw:   nc.raw,
			})
			continue
		}
		fields := compareFields(oc.c, nc.c)
		if len(fields) == 0 {
			continue
		}
		changes = append(changes, cardChange{
			OracleID: oid,
			Name:     nc.c.Name,
			Fields:   fields,
			OldRaw:   oc.raw,
			NewRaw:   nc.raw,
		})
	}
	for oid, oc := range oldCards {
		if _, ok := newCards[oid]; !ok {
			changes = append(changes, cardChange{
				OracleID: oid,
				Name:     oc.c.Name,
				Fields:   []fieldChange{{Field: "removed", Old: oc.c.Name, New: ""}},
				OldRaw:   oc.raw,
			})
		}
	}
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Name < changes[j].Name
	})
	return changes
}

func compareFields(oldC, newC card) []fieldChange {
	var fc []fieldChange
	if oldC.Name != newC.Name {
		fc = append(fc, fieldChange{"name", oldC.Name, newC.Name})
	}
	if oldC.OracleText != newC.OracleText {
		fc = append(fc, fieldChange{"oracle_text", oldC.OracleText, newC.OracleText})
	}
	if oldC.TypeLine != newC.TypeLine {
		fc = append(fc, fieldChange{"type_line", oldC.TypeLine, newC.TypeLine})
	}
	if oldC.ManaCost != newC.ManaCost {
		fc = append(fc, fieldChange{"mana_cost", oldC.ManaCost, newC.ManaCost})
	}
	// Dual-faced cards stash the real text under card_faces. Compare each
	// face by index — adding/removing a face counts as a structural change.
	if len(oldC.CardFaces) != len(newC.CardFaces) {
		fc = append(fc, fieldChange{
			"card_faces.count",
			fmt.Sprintf("%d faces", len(oldC.CardFaces)),
			fmt.Sprintf("%d faces", len(newC.CardFaces)),
		})
	} else {
		for i := range oldC.CardFaces {
			if oldC.CardFaces[i].OracleText != newC.CardFaces[i].OracleText {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].oracle_text", i),
					oldC.CardFaces[i].OracleText,
					newC.CardFaces[i].OracleText,
				})
			}
			if oldC.CardFaces[i].TypeLine != newC.CardFaces[i].TypeLine {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].type_line", i),
					oldC.CardFaces[i].TypeLine,
					newC.CardFaces[i].TypeLine,
				})
			}
			if oldC.CardFaces[i].ManaCost != newC.CardFaces[i].ManaCost {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].mana_cost", i),
					oldC.CardFaces[i].ManaCost,
					newC.CardFaces[i].ManaCost,
				})
			}
		}
	}
	return fc
}

func (cc cardChange) truncate(s string) string {
	s = strings.ReplaceAll(s, "\n", " ↵ ")
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}

// ────────────────────────────────────────────────────────────────────────
// Re-parse via scripts/parser.py
// ────────────────────────────────────────────────────────────────────────

type parseImpact struct {
	Name      string
	OldSig    string
	NewSig    string
	Changed   bool
	OldErrors int
	NewErrors int
	ParseErr  string
}

// pyParseHelper is fed the changed cards as JSONL on stdin and emits one
// JSON line per card per version with the structural signature.
// We use repr(ast.abilities) — not signature() — because signature is a
// structural fingerprint that elides numeric parameters. A "deals 3
// damage" → "deals 4 damage" errata would have the same signature, so we
// need the full repr to detect parameter-level changes.
const pyParseHelper = `
import json, sys
from pathlib import Path
sys.path.insert(0, str(Path('.').resolve() / 'scripts'))
import parser as P

P.load_extensions()

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    item = json.loads(line)
    version = item.pop('__version__')
    oid = item.pop('__oracle_id__')
    try:
        ast = P.parse_card(item)
        errs = len(ast.parse_errors or ())
        sys.stdout.write(json.dumps({
            'oracle_id': oid,
            'version': version,
            'name': item.get('name', ''),
            'signature': repr(ast.abilities),
            'errors': errs,
        }) + '\n')
    except Exception as e:
        sys.stdout.write(json.dumps({
            'oracle_id': oid,
            'version': version,
            'name': item.get('name', ''),
            'error': f'{type(e).__name__}: {e}',
        }) + '\n')
`

func reparseChanged(ctx context.Context, changes []cardChange) ([]parseImpact, error) {
	// Build one JSONL stream of (version, oracle_id, raw card dict). Python
	// loads the parser once and consumes every entry — far cheaper than
	// shelling out per card and reloading the 165MB oracle file each time.
	cmd := exec.CommandContext(ctx, "python3", "-c", pyParseHelper)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Feed both versions of every changed card.
	go func() {
		defer stdin.Close()
		bw := bufio.NewWriter(stdin)
		defer bw.Flush()
		for _, ch := range changes {
			if len(ch.OldRaw) > 0 {
				writeParseRequest(bw, ch.OracleID, "old", ch.OldRaw)
			}
			if len(ch.NewRaw) > 0 {
				writeParseRequest(bw, ch.OracleID, "new", ch.NewRaw)
			}
		}
	}()

	results := make(map[string]map[string]parseEntry) // oracle_id → version → entry
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1<<20), 1<<22)
	for scanner.Scan() {
		var pe parseEntry
		if err := json.Unmarshal(scanner.Bytes(), &pe); err != nil {
			continue
		}
		if results[pe.OracleID] == nil {
			results[pe.OracleID] = make(map[string]parseEntry, 2)
		}
		results[pe.OracleID][pe.Version] = pe
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	impacts := make([]parseImpact, 0, len(changes))
	for _, ch := range changes {
		old := results[ch.OracleID]["old"]
		newE := results[ch.OracleID]["new"]
		impact := parseImpact{
			Name:      ch.Name,
			OldSig:    old.Signature,
			NewSig:    newE.Signature,
			OldErrors: old.Errors,
			NewErrors: newE.Errors,
			Changed:   old.Signature != newE.Signature,
		}
		if old.Error != "" || newE.Error != "" {
			impact.ParseErr = strings.TrimSpace(old.Error + " | " + newE.Error)
		}
		impacts = append(impacts, impact)
	}
	return impacts, nil
}

type parseEntry struct {
	OracleID  string `json:"oracle_id"`
	Version   string `json:"version"`
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Errors    int    `json:"errors"`
	Error     string `json:"error,omitempty"`
}

func writeParseRequest(w io.Writer, oid, version string, raw json.RawMessage) {
	// Inject the bookkeeping fields into the raw card dict by patching the
	// trailing `}`. Cheaper than re-marshalling the whole card.
	trimmed := []byte(strings.TrimSpace(string(raw)))
	if len(trimmed) < 2 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return
	}
	body := trimmed[1 : len(trimmed)-1]
	prefix := []byte(`{"__version__":` + jsonString(version) +
		`,"__oracle_id__":` + jsonString(oid))
	if len(body) > 0 {
		prefix = append(prefix, ',')
	}
	out := append(prefix, body...)
	out = append(out, '}', '\n')
	_, _ = w.Write(out)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// ────────────────────────────────────────────────────────────────────────
// Thor on the changed subset
// ────────────────────────────────────────────────────────────────────────

func runThor(ctx context.Context, changes []cardChange, outputDir string) (string, error) {
	if len(changes) == 0 {
		return "No changed cards — Thor not run.", nil
	}
	listPath := filepath.Join(outputDir, "oracle-sync-thor-cards.txt")
	thorReportPath := filepath.Join(outputDir, "oracle-sync-thor-report.md")

	f, err := os.Create(listPath)
	if err != nil {
		return "", err
	}
	bw := bufio.NewWriter(f)
	for _, ch := range changes {
		// Cards added/removed don't have an existing AST in the corpus, so
		// Thor can't test them against the engine. Skip those.
		if len(ch.OldRaw) == 0 || len(ch.NewRaw) == 0 {
			continue
		}
		fmt.Fprintln(bw, ch.Name)
	}
	if err := bw.Flush(); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/hexdek-thor",
		"--card-list", listPath,
		"--report", thorReportPath,
	)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	tail := tailLines(string(out), 60)
	if err != nil {
		return tail, fmt.Errorf("thor exit: %w", err)
	}
	return fmt.Sprintf("Thor report: %s\n\nTail of stdout:\n```\n%s\n```", thorReportPath, tail), nil
}

func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// ────────────────────────────────────────────────────────────────────────
// Report
// ────────────────────────────────────────────────────────────────────────

func writeReport(path string, totalCards int, changes []cardChange, impacts []parseImpact, thorReport string, dryRun bool) error {
	parseChangedCount := 0
	parseErrorCount := 0
	for _, p := range impacts {
		if p.Changed {
			parseChangedCount++
		}
		if p.ParseErr != "" {
			parseErrorCount++
		}
	}

	var sb strings.Builder
	fmt.Fprintln(&sb, "# Oracle Sync Report")
	fmt.Fprintf(&sb, "**Generated:** %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintln(&sb, "**Source:** Scryfall bulk data (oracle_cards)")
	fmt.Fprintf(&sb, "**Cards checked:** %d\n", totalCards)
	fmt.Fprintf(&sb, "**Changes detected:** %d\n", len(changes))
	if dryRun {
		fmt.Fprintln(&sb, "**Mode:** dry-run (parser + Thor skipped)")
	} else {
		fmt.Fprintf(&sb, "**Parse changes:** %d\n", parseChangedCount)
		fmt.Fprintf(&sb, "**Parse errors:** %d\n", parseErrorCount)
		fmt.Fprintln(&sb, "**Thor failures on changed set:** see Thor section below")
	}
	fmt.Fprintln(&sb)

	// Changed Cards table
	fmt.Fprintln(&sb, "## Changed Cards")
	fmt.Fprintln(&sb)
	if len(changes) == 0 {
		fmt.Fprintln(&sb, "_No changes._")
	} else {
		fmt.Fprintln(&sb, "| Card | Field | Before | After |")
		fmt.Fprintln(&sb, "|---|---|---|---|")
		for _, ch := range changes {
			for _, fc := range ch.Fields {
				fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n",
					mdEscape(ch.Name),
					mdEscape(fc.Field),
					mdEscape(truncate(fc.Old, 120)),
					mdEscape(truncate(fc.New, 120)),
				)
			}
		}
	}
	fmt.Fprintln(&sb)

	// Parse Impact
	fmt.Fprintln(&sb, "## Parse Impact")
	fmt.Fprintln(&sb)
	if dryRun {
		fmt.Fprintln(&sb, "_Skipped (--dry-run)._")
	} else if len(impacts) == 0 {
		fmt.Fprintln(&sb, "_No cards re-parsed._")
	} else {
		fmt.Fprintln(&sb, "| Card | AST changed | Old errors | New errors | Notes |")
		fmt.Fprintln(&sb, "|---|---|---:|---:|---|")
		for _, p := range impacts {
			marker := "no"
			if p.Changed {
				marker = "**yes**"
			}
			notes := p.ParseErr
			if notes == "" && p.Changed {
				notes = "signature differs"
			}
			fmt.Fprintf(&sb, "| %s | %s | %d | %d | %s |\n",
				mdEscape(p.Name),
				marker,
				p.OldErrors,
				p.NewErrors,
				mdEscape(truncate(notes, 120)),
			)
		}
	}
	fmt.Fprintln(&sb)

	// Thor Results
	fmt.Fprintln(&sb, "## Thor Results on Changed Set")
	fmt.Fprintln(&sb)
	if dryRun {
		fmt.Fprintln(&sb, "_Skipped (--dry-run)._")
	} else if thorReport == "" {
		fmt.Fprintln(&sb, "_Thor was not run._")
	} else {
		fmt.Fprintln(&sb, thorReport)
	}
	fmt.Fprintln(&sb)

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

// ────────────────────────────────────────────────────────────────────────
// helpers
// ────────────────────────────────────────────────────────────────────────

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ↵ ")
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}

func mdEscape(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
