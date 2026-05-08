// hexdek-oracle-sync — Pull fresh Scryfall oracle bulk data, diff against
// the local snapshot, flag cards needing AST re-parse, and run Thor on the
// changed subset.
//
// The local source of truth is data/rules/oracle-cards.json.
//
// Usage:
//
//	hexdek-oracle-sync --live          # download, diff, update, re-parse, Thor
//	hexdek-oracle-sync --dry-run       # download + diff only, no writes
//	hexdek-oracle-sync --diff-only     # compare existing local files only
//	hexdek-oracle-sync --report        # show last sync report
//	hexdek-oracle-sync --live --verbose # print changed cards to stdout
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
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	scryfallBulkAPI   = "https://api.scryfall.com/bulk-data"
	userAgent         = "HexDek/1.0"
	scryfallRateLimit = 100 * time.Millisecond
	httpTimeout       = 10 * time.Minute
)

// ────────────────────────────────────────────────────────────────────────
// Domain types
// ────────────────────────────────────────────────────────────────────────

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
	Power      string     `json:"power"`
	Toughness  string     `json:"toughness"`
	Keywords   []string   `json:"keywords"`
	Layout     string     `json:"layout"`
	CardFaces  []cardFace `json:"card_faces,omitempty"`
}

type cardFace struct {
	Name       string `json:"name"`
	OracleText string `json:"oracle_text"`
	TypeLine   string `json:"type_line"`
	ManaCost   string `json:"mana_cost"`
	Power      string `json:"power"`
	Toughness  string `json:"toughness"`
}

// CardDiff represents a single card that changed between snapshots.
type CardDiff struct {
	Name          string
	Kind          string // "added", "removed", "changed"
	OldOracle     string
	NewOracle     string
	ChangedFields []string
}

type cardChange struct {
	OracleID string
	Name     string
	Kind     string // "added", "removed", "changed"
	Fields   []fieldChange
	OldRaw   json.RawMessage
	NewRaw   json.RawMessage
}

type fieldChange struct {
	Field string
	Old   string
	New   string
}

// ToCardDiff converts internal cardChange to the public CardDiff type.
func (cc cardChange) ToCardDiff() CardDiff {
	fields := make([]string, 0, len(cc.Fields))
	for _, f := range cc.Fields {
		fields = append(fields, f.Field)
	}
	var oldOracle, newOracle string
	for _, f := range cc.Fields {
		if f.Field == "oracle_text" {
			oldOracle = f.Old
			newOracle = f.New
			break
		}
	}
	return CardDiff{
		Name:          cc.Name,
		Kind:          cc.Kind,
		OldOracle:     oldOracle,
		NewOracle:     newOracle,
		ChangedFields: fields,
	}
}

// ────────────────────────────────────────────────────────────────────────
// Entry point
// ────────────────────────────────────────────────────────────────────────

func main() {
	live := flag.Bool("live", false, "download from Scryfall, diff, update local files")
	dryRun := flag.Bool("dry-run", false, "download + diff only, no writes")
	diffOnly := flag.Bool("diff-only", false, "compare existing local files only (no download)")
	report := flag.Bool("report", false, "show last sync report")
	verbose := flag.Bool("verbose", false, "print changed cards to stdout")
	outputDir := flag.String("output", "data/rules", "directory holding oracle-cards.json")
	reportDir := flag.String("report-dir", "data", "directory for oracle-sync-report.md")
	skipThor := flag.Bool("skip-thor", false, "diff and re-parse only — skip the Thor subset run")
	flag.Parse()

	// Mutual exclusion
	modeCount := 0
	if *live {
		modeCount++
	}
	if *dryRun {
		modeCount++
	}
	if *diffOnly {
		modeCount++
	}
	if *report {
		modeCount++
	}
	if modeCount == 0 {
		fmt.Fprintln(os.Stderr, "error: specify one of --live, --dry-run, --diff-only, or --report")
		flag.Usage()
		os.Exit(1)
	}
	if modeCount > 1 {
		fmt.Fprintln(os.Stderr, "error: --live, --dry-run, --diff-only, and --report are mutually exclusive")
		flag.Usage()
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *report {
		reportPath := filepath.Join(*reportDir, "oracle-sync-report.md")
		data, err := os.ReadFile(reportPath)
		if err != nil {
			log.Fatalf("oracle-sync: cannot read report: %v", err)
		}
		fmt.Print(string(data))
		return
	}

	if err := run(ctx, *outputDir, *reportDir, *live, *dryRun, *diffOnly, *verbose, *skipThor); err != nil {
		log.Fatalf("oracle-sync: %v", err)
	}
}

func run(ctx context.Context, outputDir, reportDir string, live, dryRun, diffOnly, verbose, skipThor bool) error {
	currentPath := filepath.Join(outputDir, "oracle-cards.json")
	newPath := filepath.Join(outputDir, "oracle-cards-new.json")
	reportPath := filepath.Join(reportDir, "oracle-sync-report.md")

	if _, err := os.Stat(currentPath); err != nil {
		return fmt.Errorf("read current snapshot %s: %w", currentPath, err)
	}

	// ─── Step 1: download (unless --diff-only) ────────────────────────
	if diffOnly {
		if _, err := os.Stat(newPath); err != nil {
			return fmt.Errorf("--diff-only set but %s is missing: %w", newPath, err)
		}
		log.Printf("diff-only mode — reusing %s", newPath)
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

	// Compute summary counts
	var addedCount, removedCount, changedCount int
	for _, ch := range changes {
		switch ch.Kind {
		case "added":
			addedCount++
		case "removed":
			removedCount++
		case "changed":
			changedCount++
		}
	}
	log.Printf("  added: %d, removed: %d, changed: %d", addedCount, removedCount, changedCount)

	if verbose {
		for _, ch := range changes {
			fmt.Printf("  [%s] %s\n", ch.Kind, ch.Name)
			for _, f := range ch.Fields {
				fmt.Printf("    %s: %q → %q\n", f.Field, truncateField(f.Old), truncateField(f.New))
			}
		}
	}

	// ─── Step 3: re-parse + Thor (--live mode only) ────────────────────
	var (
		parseImpacts []parseImpact
		thorReport   string
	)
	if live && len(changes) > 0 {
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

	// ─── Step 4: write report ──────────────────────────────────────────
	if err := writeReport(reportPath, len(oldCards), len(newCards), changes, parseImpacts, thorReport, dryRun || diffOnly); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	log.Printf("report: %s", reportPath)

	// ─── Step 5: update local file (--live only) ───────────────────────
	if live {
		if err := os.Rename(newPath, currentPath); err != nil {
			return fmt.Errorf("promote: rename %s → %s: %w", newPath, currentPath, err)
		}
		log.Printf("updated: %s now reflects the fresh download", currentPath)

		// Flag cards needing AST re-parse
		needReparse := 0
		for _, ch := range changes {
			if ch.Kind == "added" || ch.Kind == "changed" {
				needReparse++
			}
		}
		if needReparse > 0 {
			log.Printf("REPARSE NEEDED: %d cards need AST re-parse (run hexdek-thor to rebuild)", needReparse)
		}
	}

	return nil
}

// ────────────────────────────────────────────────────────────────────────
// Scryfall bulk download (streaming — no 170MB in memory)
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

// wsNorm collapses whitespace runs to single spaces and trims.
var wsRe = regexp.MustCompile(`\s+`)

func normalizeWS(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

func diffCards(oldCards, newCards map[string]loadedCard) []cardChange {
	var changes []cardChange
	for oid, nc := range newCards {
		oc, ok := oldCards[oid]
		if !ok {
			changes = append(changes, cardChange{
				OracleID: oid,
				Name:     nc.c.Name,
				Kind:     "added",
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
			Kind:     "changed",
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
				Kind:     "removed",
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
	if normalizeWS(oldC.OracleText) != normalizeWS(newC.OracleText) {
		fc = append(fc, fieldChange{"oracle_text", oldC.OracleText, newC.OracleText})
	}
	if normalizeWS(oldC.TypeLine) != normalizeWS(newC.TypeLine) {
		fc = append(fc, fieldChange{"type_line", oldC.TypeLine, newC.TypeLine})
	}
	if normalizeWS(oldC.ManaCost) != normalizeWS(newC.ManaCost) {
		fc = append(fc, fieldChange{"mana_cost", oldC.ManaCost, newC.ManaCost})
	}
	if normalizeWS(oldC.Power) != normalizeWS(newC.Power) {
		fc = append(fc, fieldChange{"power", oldC.Power, newC.Power})
	}
	if normalizeWS(oldC.Toughness) != normalizeWS(newC.Toughness) {
		fc = append(fc, fieldChange{"toughness", oldC.Toughness, newC.Toughness})
	}
	if normalizeWS(oldC.Layout) != normalizeWS(newC.Layout) {
		fc = append(fc, fieldChange{"layout", oldC.Layout, newC.Layout})
	}
	// Compare keywords (sorted for stability)
	oldKW := sortedKeywords(oldC.Keywords)
	newKW := sortedKeywords(newC.Keywords)
	if oldKW != newKW {
		fc = append(fc, fieldChange{"keywords", oldKW, newKW})
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
			if normalizeWS(oldC.CardFaces[i].OracleText) != normalizeWS(newC.CardFaces[i].OracleText) {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].oracle_text", i),
					oldC.CardFaces[i].OracleText,
					newC.CardFaces[i].OracleText,
				})
			}
			if normalizeWS(oldC.CardFaces[i].TypeLine) != normalizeWS(newC.CardFaces[i].TypeLine) {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].type_line", i),
					oldC.CardFaces[i].TypeLine,
					newC.CardFaces[i].TypeLine,
				})
			}
			if normalizeWS(oldC.CardFaces[i].ManaCost) != normalizeWS(newC.CardFaces[i].ManaCost) {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].mana_cost", i),
					oldC.CardFaces[i].ManaCost,
					newC.CardFaces[i].ManaCost,
				})
			}
			if normalizeWS(oldC.CardFaces[i].Power) != normalizeWS(newC.CardFaces[i].Power) {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].power", i),
					oldC.CardFaces[i].Power,
					newC.CardFaces[i].Power,
				})
			}
			if normalizeWS(oldC.CardFaces[i].Toughness) != normalizeWS(newC.CardFaces[i].Toughness) {
				fc = append(fc, fieldChange{
					fmt.Sprintf("card_faces[%d].toughness", i),
					oldC.CardFaces[i].Toughness,
					newC.CardFaces[i].Toughness,
				})
			}
		}
	}
	return fc
}

func sortedKeywords(kws []string) string {
	if len(kws) == 0 {
		return ""
	}
	cp := make([]string, len(kws))
	copy(cp, kws)
	sort.Strings(cp)
	return strings.Join(cp, ", ")
}

func truncateField(s string) string {
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

	results := make(map[string]map[string]parseEntry)
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
	trimmed := []byte(strings.TrimSpace(string(raw)))
	if len(trimmed) < 2 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return
	}
	body := trimmed[1 : len(trimmed)-1]
	prefix := []byte(`{"__version__":` + marshalString(version) +
		`,"__oracle_id__":` + marshalString(oid))
	if len(body) > 0 {
		prefix = append(prefix, ',')
	}
	out := append(prefix, body...)
	out = append(out, '}', '\n')
	_, _ = w.Write(out)
}

func marshalString(s string) string {
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

func writeReport(path string, oldTotal, newTotal int, changes []cardChange, impacts []parseImpact, thorReport string, readOnly bool) error {
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

	var addedCount, removedCount, changedCount int
	for _, ch := range changes {
		switch ch.Kind {
		case "added":
			addedCount++
		case "removed":
			removedCount++
		case "changed":
			changedCount++
		}
	}

	var sb strings.Builder
	fmt.Fprintln(&sb, "# Oracle Sync Report")
	fmt.Fprintf(&sb, "**Generated:** %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintln(&sb, "**Source:** Scryfall bulk data (oracle_cards)")
	fmt.Fprintf(&sb, "**Cards (old snapshot):** %d\n", oldTotal)
	fmt.Fprintf(&sb, "**Cards (new snapshot):** %d\n", newTotal)
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "## Summary")
	fmt.Fprintln(&sb)
	fmt.Fprintln(&sb, "| Metric | Count |")
	fmt.Fprintln(&sb, "|--------|------:|")
	fmt.Fprintf(&sb, "| Total changes | %d |\n", len(changes))
	fmt.Fprintf(&sb, "| Added | %d |\n", addedCount)
	fmt.Fprintf(&sb, "| Removed | %d |\n", removedCount)
	fmt.Fprintf(&sb, "| Changed | %d |\n", changedCount)
	if !readOnly {
		fmt.Fprintf(&sb, "| Parse-result changes | %d |\n", parseChangedCount)
		fmt.Fprintf(&sb, "| Parse errors | %d |\n", parseErrorCount)
	}
	fmt.Fprintln(&sb)

	if readOnly {
		fmt.Fprintf(&sb, "**Mode:** read-only (parser + Thor skipped)\n\n")
	}

	// Changed Cards table with before/after oracle text
	fmt.Fprintln(&sb, "## Changed Cards")
	fmt.Fprintln(&sb)
	if len(changes) == 0 {
		fmt.Fprintln(&sb, "_No changes._")
	} else {
		fmt.Fprintln(&sb, "| Card | Kind | Field | Before | After |")
		fmt.Fprintln(&sb, "|------|------|-------|--------|-------|")
		for _, ch := range changes {
			for _, fc := range ch.Fields {
				fmt.Fprintf(&sb, "| %s | %s | %s | %s | %s |\n",
					mdEscape(ch.Name),
					ch.Kind,
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
	if readOnly {
		fmt.Fprintln(&sb, "_Skipped (read-only mode)._")
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
	if readOnly {
		fmt.Fprintln(&sb, "_Skipped (read-only mode)._")
	} else if thorReport == "" {
		fmt.Fprintln(&sb, "_Thor was not run._")
	} else {
		fmt.Fprintln(&sb, thorReport)
	}
	fmt.Fprintln(&sb)

	// Cards needing re-parse
	fmt.Fprintln(&sb, "## Cards Needing AST Re-Parse")
	fmt.Fprintln(&sb)
	needReparse := 0
	for _, ch := range changes {
		if ch.Kind == "added" || ch.Kind == "changed" {
			needReparse++
		}
	}
	if needReparse == 0 {
		fmt.Fprintln(&sb, "_None._")
	} else {
		fmt.Fprintf(&sb, "%d cards need re-parsing:\n\n", needReparse)
		for _, ch := range changes {
			if ch.Kind == "added" || ch.Kind == "changed" {
				fmt.Fprintf(&sb, "- %s (%s)\n", ch.Name, ch.Kind)
			}
		}
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
