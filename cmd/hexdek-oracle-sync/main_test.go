package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeCard builds a minimal Scryfall-format JSON card object.
func makeCard(oracleID, name, oracleText, typeLine, manaCost, power, toughness, layout string, keywords []string) json.RawMessage {
	c := map[string]interface{}{
		"oracle_id":   oracleID,
		"name":        name,
		"oracle_text": oracleText,
		"type_line":   typeLine,
		"mana_cost":   manaCost,
		"power":       power,
		"toughness":   toughness,
		"layout":      layout,
		"keywords":    keywords,
	}
	b, _ := json.Marshal(c)
	return b
}

func writeCardsFile(t *testing.T, dir, filename string, cards ...json.RawMessage) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	f.WriteString("[")
	for i, c := range cards {
		if i > 0 {
			f.WriteString(",")
		}
		f.Write(c)
	}
	f.WriteString("]")
	return path
}

// ────────────────────────────────────────────────────────────────────────
// normalizeWS
// ────────────────────────────────────────────────────────────────────────

func TestNormalizeWS(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello  world", "hello world"},
		{"  leading", "leading"},
		{"trailing  ", "trailing"},
		{"multiple   spaces   between", "multiple spaces between"},
		{"tabs\tand\nnewlines", "tabs and newlines"},
		{"", ""},
		{"single", "single"},
	}
	for _, tt := range tests {
		got := normalizeWS(tt.in)
		if got != tt.want {
			t.Errorf("normalizeWS(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────
// compareFields
// ────────────────────────────────────────────────────────────────────────

func TestCompareFields_NoChange(t *testing.T) {
	c := card{
		OracleText: "Draw a card.",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		Power:      "",
		Toughness:  "",
		Keywords:   nil,
		Layout:     "normal",
	}
	fc := compareFields(c, c)
	if len(fc) != 0 {
		t.Errorf("expected no changes, got %d: %+v", len(fc), fc)
	}
}

func TestCompareFields_WhitespaceOnly(t *testing.T) {
	old := card{OracleText: "Draw  a card.", TypeLine: "Instant", Layout: "normal"}
	new := card{OracleText: "Draw a card.", TypeLine: "Instant", Layout: "normal"}
	fc := compareFields(old, new)
	if len(fc) != 0 {
		t.Errorf("whitespace-only diff should not count as change, got %d: %+v", len(fc), fc)
	}
}

func TestCompareFields_OracleTextChange(t *testing.T) {
	old := card{
		OracleText: "Draw a card.",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		Layout:     "normal",
	}
	new := card{
		OracleText: "Draw two cards.",
		TypeLine:   "Instant",
		ManaCost:   "{U}",
		Layout:     "normal",
	}
	fc := compareFields(old, new)
	if len(fc) != 1 {
		t.Fatalf("expected 1 change, got %d: %+v", len(fc), fc)
	}
	if fc[0].Field != "oracle_text" {
		t.Errorf("expected field oracle_text, got %s", fc[0].Field)
	}
	if fc[0].Old != "Draw a card." {
		t.Errorf("expected old 'Draw a card.', got %q", fc[0].Old)
	}
	if fc[0].New != "Draw two cards." {
		t.Errorf("expected new 'Draw two cards.', got %q", fc[0].New)
	}
}

func TestCompareFields_PowerToughnessChange(t *testing.T) {
	old := card{Power: "2", Toughness: "3", Layout: "normal"}
	new := card{Power: "3", Toughness: "4", Layout: "normal"}
	fc := compareFields(old, new)
	if len(fc) != 2 {
		t.Fatalf("expected 2 changes (power + toughness), got %d: %+v", len(fc), fc)
	}
	fields := map[string]bool{}
	for _, f := range fc {
		fields[f.Field] = true
	}
	if !fields["power"] || !fields["toughness"] {
		t.Errorf("expected power and toughness changes, got fields: %+v", fc)
	}
}

func TestCompareFields_KeywordsChange(t *testing.T) {
	old := card{Keywords: []string{"Flying", "Haste"}, Layout: "normal"}
	new := card{Keywords: []string{"Flying", "Trample"}, Layout: "normal"}
	fc := compareFields(old, new)
	if len(fc) != 1 {
		t.Fatalf("expected 1 change (keywords), got %d: %+v", len(fc), fc)
	}
	if fc[0].Field != "keywords" {
		t.Errorf("expected keywords field, got %s", fc[0].Field)
	}
}

func TestCompareFields_KeywordsOrderIrrelevant(t *testing.T) {
	old := card{Keywords: []string{"Haste", "Flying"}, Layout: "normal"}
	new := card{Keywords: []string{"Flying", "Haste"}, Layout: "normal"}
	fc := compareFields(old, new)
	if len(fc) != 0 {
		t.Errorf("keyword order should not matter, got %d changes: %+v", len(fc), fc)
	}
}

func TestCompareFields_LayoutChange(t *testing.T) {
	old := card{Layout: "normal"}
	new := card{Layout: "transform"}
	fc := compareFields(old, new)
	if len(fc) != 1 || fc[0].Field != "layout" {
		t.Errorf("expected layout change, got %+v", fc)
	}
}

func TestCompareFields_MultipleChanges(t *testing.T) {
	old := card{
		OracleText: "Destroy target creature.",
		TypeLine:   "Instant",
		ManaCost:   "{B}{B}",
		Power:      "",
		Toughness:  "",
		Layout:     "normal",
	}
	new := card{
		OracleText: "Destroy target creature or planeswalker.",
		TypeLine:   "Sorcery",
		ManaCost:   "{1}{B}",
		Power:      "",
		Toughness:  "",
		Layout:     "normal",
	}
	fc := compareFields(old, new)
	if len(fc) != 3 {
		t.Fatalf("expected 3 changes, got %d: %+v", len(fc), fc)
	}
	fields := map[string]bool{}
	for _, f := range fc {
		fields[f.Field] = true
	}
	if !fields["oracle_text"] || !fields["type_line"] || !fields["mana_cost"] {
		t.Errorf("missing expected fields, got: %+v", fc)
	}
}

// ────────────────────────────────────────────────────────────────────────
// diffCards
// ────────────────────────────────────────────────────────────────────────

func TestDiffCards_Added(t *testing.T) {
	old := map[string]loadedCard{}
	new := map[string]loadedCard{
		"abc-123": {
			c:   card{OracleID: "abc-123", Name: "New Card"},
			raw: makeCard("abc-123", "New Card", "Draw a card.", "Instant", "{U}", "", "", "normal", nil),
		},
	}
	changes := diffCards(old, new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "added" {
		t.Errorf("expected kind=added, got %s", changes[0].Kind)
	}
	if changes[0].Name != "New Card" {
		t.Errorf("expected name 'New Card', got %s", changes[0].Name)
	}
}

func TestDiffCards_Removed(t *testing.T) {
	old := map[string]loadedCard{
		"abc-123": {
			c:   card{OracleID: "abc-123", Name: "Old Card"},
			raw: makeCard("abc-123", "Old Card", "Deal 3 damage.", "Instant", "{R}", "", "", "normal", nil),
		},
	}
	new := map[string]loadedCard{}
	changes := diffCards(old, new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "removed" {
		t.Errorf("expected kind=removed, got %s", changes[0].Kind)
	}
}

func TestDiffCards_Changed(t *testing.T) {
	old := map[string]loadedCard{
		"abc-123": {
			c:   card{OracleID: "abc-123", Name: "Lightning Bolt", OracleText: "Deal 3 damage.", Layout: "normal"},
			raw: makeCard("abc-123", "Lightning Bolt", "Deal 3 damage.", "Instant", "{R}", "", "", "normal", nil),
		},
	}
	new := map[string]loadedCard{
		"abc-123": {
			c:   card{OracleID: "abc-123", Name: "Lightning Bolt", OracleText: "Deal 4 damage.", Layout: "normal"},
			raw: makeCard("abc-123", "Lightning Bolt", "Deal 4 damage.", "Instant", "{R}", "", "", "normal", nil),
		},
	}
	changes := diffCards(old, new)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Kind != "changed" {
		t.Errorf("expected kind=changed, got %s", changes[0].Kind)
	}
}

func TestDiffCards_NoChange(t *testing.T) {
	c := loadedCard{
		c:   card{OracleID: "abc-123", Name: "Opt", OracleText: "Scry 1, then draw a card.", Layout: "normal"},
		raw: makeCard("abc-123", "Opt", "Scry 1, then draw a card.", "Instant", "{U}", "", "", "normal", nil),
	}
	old := map[string]loadedCard{"abc-123": c}
	new := map[string]loadedCard{"abc-123": c}
	changes := diffCards(old, new)
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestDiffCards_MixedOperations(t *testing.T) {
	old := map[string]loadedCard{
		"id-stay": {
			c:   card{OracleID: "id-stay", Name: "Stay Card", OracleText: "Old text.", Layout: "normal"},
			raw: makeCard("id-stay", "Stay Card", "Old text.", "Instant", "{W}", "", "", "normal", nil),
		},
		"id-remove": {
			c:   card{OracleID: "id-remove", Name: "Remove Card", OracleText: "Goodbye.", Layout: "normal"},
			raw: makeCard("id-remove", "Remove Card", "Goodbye.", "Creature", "{B}", "2", "2", "normal", nil),
		},
		"id-change": {
			c:   card{OracleID: "id-change", Name: "Change Card", OracleText: "Version 1.", Layout: "normal"},
			raw: makeCard("id-change", "Change Card", "Version 1.", "Sorcery", "{G}", "", "", "normal", nil),
		},
	}
	new := map[string]loadedCard{
		"id-stay": {
			c:   card{OracleID: "id-stay", Name: "Stay Card", OracleText: "Old text.", Layout: "normal"},
			raw: makeCard("id-stay", "Stay Card", "Old text.", "Instant", "{W}", "", "", "normal", nil),
		},
		"id-change": {
			c:   card{OracleID: "id-change", Name: "Change Card", OracleText: "Version 2.", Layout: "normal"},
			raw: makeCard("id-change", "Change Card", "Version 2.", "Sorcery", "{G}", "", "", "normal", nil),
		},
		"id-add": {
			c:   card{OracleID: "id-add", Name: "Add Card", OracleText: "Hello.", Layout: "normal"},
			raw: makeCard("id-add", "Add Card", "Hello.", "Enchantment", "{R}", "", "", "normal", nil),
		},
	}
	changes := diffCards(old, new)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes (add + remove + change), got %d: %+v", len(changes), changes)
	}
	kinds := map[string]int{}
	for _, ch := range changes {
		kinds[ch.Kind]++
	}
	if kinds["added"] != 1 || kinds["removed"] != 1 || kinds["changed"] != 1 {
		t.Errorf("expected 1 of each kind, got: %+v", kinds)
	}
}

// ────────────────────────────────────────────────────────────────────────
// CardDiff conversion
// ────────────────────────────────────────────────────────────────────────

func TestToCardDiff(t *testing.T) {
	ch := cardChange{
		OracleID: "abc",
		Name:     "Test Card",
		Kind:     "changed",
		Fields: []fieldChange{
			{Field: "oracle_text", Old: "Old oracle.", New: "New oracle."},
			{Field: "type_line", Old: "Instant", New: "Sorcery"},
		},
	}
	diff := ch.ToCardDiff()
	if diff.Name != "Test Card" {
		t.Errorf("name = %q, want 'Test Card'", diff.Name)
	}
	if diff.Kind != "changed" {
		t.Errorf("kind = %q, want 'changed'", diff.Kind)
	}
	if diff.OldOracle != "Old oracle." {
		t.Errorf("old oracle = %q, want 'Old oracle.'", diff.OldOracle)
	}
	if diff.NewOracle != "New oracle." {
		t.Errorf("new oracle = %q, want 'New oracle.'", diff.NewOracle)
	}
	if len(diff.ChangedFields) != 2 {
		t.Fatalf("changed fields = %d, want 2", len(diff.ChangedFields))
	}
	if diff.ChangedFields[0] != "oracle_text" || diff.ChangedFields[1] != "type_line" {
		t.Errorf("changed fields = %v, want [oracle_text, type_line]", diff.ChangedFields)
	}
}

func TestToCardDiff_Added(t *testing.T) {
	ch := cardChange{
		Name: "Brand New",
		Kind: "added",
		Fields: []fieldChange{
			{Field: "added", Old: "", New: "Brand New"},
		},
	}
	diff := ch.ToCardDiff()
	if diff.Kind != "added" {
		t.Errorf("kind = %q, want 'added'", diff.Kind)
	}
	// No oracle_text field change, so OldOracle/NewOracle should be empty.
	if diff.OldOracle != "" || diff.NewOracle != "" {
		t.Errorf("expected empty oracle for added card, got old=%q new=%q", diff.OldOracle, diff.NewOracle)
	}
}

// ────────────────────────────────────────────────────────────────────────
// loadOracle + round-trip
// ────────────────────────────────────────────────────────────────────────

func TestLoadOracle(t *testing.T) {
	dir := t.TempDir()
	cards := []json.RawMessage{
		makeCard("id-1", "Card A", "Flying.", "Creature", "{W}", "2", "2", "normal", []string{"Flying"}),
		makeCard("id-2", "Card B", "Haste.", "Creature", "{R}", "3", "1", "normal", []string{"Haste"}),
		makeCard("", "Token", "I'm a token", "Token", "", "1", "1", "normal", nil), // no oracle_id, should be skipped
	}
	path := writeCardsFile(t, dir, "oracle-cards.json", cards...)

	loaded, err := loadOracle(path)
	if err != nil {
		t.Fatalf("loadOracle: %v", err)
	}
	if len(loaded) != 2 {
		t.Errorf("expected 2 cards (token skipped), got %d", len(loaded))
	}
	if loaded["id-1"].c.Name != "Card A" {
		t.Errorf("card A name = %q", loaded["id-1"].c.Name)
	}
	if loaded["id-2"].c.Power != "3" {
		t.Errorf("card B power = %q, want '3'", loaded["id-2"].c.Power)
	}
}

func TestLoadOracle_DuplicateOracleID(t *testing.T) {
	dir := t.TempDir()
	cards := []json.RawMessage{
		makeCard("id-1", "First Printing", "Text A.", "Instant", "{U}", "", "", "normal", nil),
		makeCard("id-1", "Second Printing", "Text B.", "Instant", "{U}", "", "", "normal", nil),
	}
	path := writeCardsFile(t, dir, "oracle-cards.json", cards...)

	loaded, err := loadOracle(path)
	if err != nil {
		t.Fatalf("loadOracle: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("expected 1 card (dupe skipped), got %d", len(loaded))
	}
	// First one wins
	if loaded["id-1"].c.Name != "First Printing" {
		t.Errorf("expected first printing, got %q", loaded["id-1"].c.Name)
	}
}

// ────────────────────────────────────────────────────────────────────────
// Full diff via file I/O (integration-level, no Scryfall)
// ────────────────────────────────────────────────────────────────────────

func TestDiffViaFiles(t *testing.T) {
	dir := t.TempDir()

	oldCards := []json.RawMessage{
		makeCard("id-1", "Bolt", "Deal 3 damage to any target.", "Instant", "{R}", "", "", "normal", nil),
		makeCard("id-2", "Opt", "Scry 1, then draw a card.", "Instant", "{U}", "", "", "normal", nil),
		makeCard("id-3", "Grizzly Bears", "Vanilla bear.", "Creature", "{1}{G}", "2", "2", "normal", nil),
	}
	newCards := []json.RawMessage{
		makeCard("id-1", "Bolt", "Deal 3 damage to any target.", "Instant", "{R}", "", "", "normal", nil),   // unchanged
		makeCard("id-2", "Opt", "Scry 2, then draw a card.", "Instant", "{U}", "", "", "normal", nil),        // oracle text changed
		makeCard("id-4", "Counterspell", "Counter target spell.", "Instant", "{U}{U}", "", "", "normal", nil), // new card
		// id-3 removed
	}

	writeCardsFile(t, dir, "oracle-cards.json", oldCards...)
	writeCardsFile(t, dir, "oracle-cards-new.json", newCards...)

	old, err := loadOracle(filepath.Join(dir, "oracle-cards.json"))
	if err != nil {
		t.Fatal(err)
	}
	new, err := loadOracle(filepath.Join(dir, "oracle-cards-new.json"))
	if err != nil {
		t.Fatal(err)
	}

	changes := diffCards(old, new)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}

	byName := map[string]cardChange{}
	for _, ch := range changes {
		byName[ch.Name] = ch
	}

	if byName["Counterspell"].Kind != "added" {
		t.Errorf("Counterspell should be added, got %s", byName["Counterspell"].Kind)
	}
	if byName["Grizzly Bears"].Kind != "removed" {
		t.Errorf("Grizzly Bears should be removed, got %s", byName["Grizzly Bears"].Kind)
	}
	if byName["Opt"].Kind != "changed" {
		t.Errorf("Opt should be changed, got %s", byName["Opt"].Kind)
	}
	// Verify Opt has oracle_text in ChangedFields
	optDiff := byName["Opt"].ToCardDiff()
	found := false
	for _, f := range optDiff.ChangedFields {
		if f == "oracle_text" {
			found = true
		}
	}
	if !found {
		t.Errorf("Opt diff should include oracle_text, got %v", optDiff.ChangedFields)
	}
	if optDiff.OldOracle != "Scry 1, then draw a card." {
		t.Errorf("Opt old oracle = %q", optDiff.OldOracle)
	}
	if optDiff.NewOracle != "Scry 2, then draw a card." {
		t.Errorf("Opt new oracle = %q", optDiff.NewOracle)
	}
}

// ────────────────────────────────────────────────────────────────────────
// Report generation
// ────────────────────────────────────────────────────────────────────────

func TestWriteReport(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "oracle-sync-report.md")

	changes := []cardChange{
		{OracleID: "id-1", Name: "Added Card", Kind: "added",
			Fields: []fieldChange{{Field: "added", Old: "", New: "Added Card"}}},
		{OracleID: "id-2", Name: "Changed Card", Kind: "changed",
			Fields: []fieldChange{{Field: "oracle_text", Old: "Old text.", New: "New text."}}},
		{OracleID: "id-3", Name: "Removed Card", Kind: "removed",
			Fields: []fieldChange{{Field: "removed", Old: "Removed Card", New: ""}}},
	}

	err := writeReport(reportPath, 100, 101, changes, nil, "", true)
	if err != nil {
		t.Fatalf("writeReport: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	report := string(data)

	// Check summary
	if !strings.Contains(report, "Total changes | 3") {
		t.Error("report should contain total changes count")
	}
	if !strings.Contains(report, "Added | 1") {
		t.Error("report should contain added count")
	}
	if !strings.Contains(report, "Removed | 1") {
		t.Error("report should contain removed count")
	}
	if !strings.Contains(report, "Changed | 1") {
		t.Error("report should contain changed count")
	}

	// Check card entries in table
	if !strings.Contains(report, "Added Card") {
		t.Error("report should mention Added Card")
	}
	if !strings.Contains(report, "Changed Card") {
		t.Error("report should mention Changed Card")
	}
	if !strings.Contains(report, "Removed Card") {
		t.Error("report should mention Removed Card")
	}

	// Check re-parse section
	if !strings.Contains(report, "Cards Needing AST Re-Parse") {
		t.Error("report should have re-parse section")
	}
	if !strings.Contains(report, "2 cards need re-parsing") {
		t.Error("report should show 2 cards needing re-parse (added + changed)")
	}
}

// ────────────────────────────────────────────────────────────────────────
// sortedKeywords
// ────────────────────────────────────────────────────────────────────────

func TestSortedKeywords(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"Flying"}, "Flying"},
		{[]string{"Haste", "Flying", "Trample"}, "Flying, Haste, Trample"},
	}
	for _, tt := range tests {
		got := sortedKeywords(tt.in)
		if got != tt.want {
			t.Errorf("sortedKeywords(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{178257920, "170.0 MB"},
	}
	for _, tt := range tests {
		got := humanBytes(tt.in)
		if got != tt.want {
			t.Errorf("humanBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	short := "hello"
	if truncate(short, 10) != "hello" {
		t.Errorf("short string should not be truncated")
	}
	long := strings.Repeat("x", 200)
	result := truncate(long, 50)
	if len(result) != 50 {
		t.Errorf("truncated length = %d, want 50", len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Error("truncated string should end with ...")
	}
}

func TestMdEscape(t *testing.T) {
	if mdEscape("a|b|c") != `a\|b\|c` {
		t.Errorf("mdEscape should escape pipes")
	}
}

// ────────────────────────────────────────────────────────────────────────
// Card faces
// ────────────────────────────────────────────────────────────────────────

func TestCompareFields_CardFaces(t *testing.T) {
	old := card{
		Layout: "transform",
		CardFaces: []cardFace{
			{Name: "Front", OracleText: "Front text.", TypeLine: "Creature", ManaCost: "{1}{G}", Power: "2", Toughness: "2"},
			{Name: "Back", OracleText: "Back text.", TypeLine: "Creature", ManaCost: "", Power: "5", Toughness: "5"},
		},
	}
	new := card{
		Layout: "transform",
		CardFaces: []cardFace{
			{Name: "Front", OracleText: "Front text changed.", TypeLine: "Creature", ManaCost: "{1}{G}", Power: "2", Toughness: "2"},
			{Name: "Back", OracleText: "Back text.", TypeLine: "Creature", ManaCost: "", Power: "5", Toughness: "5"},
		},
	}
	fc := compareFields(old, new)
	if len(fc) != 1 {
		t.Fatalf("expected 1 change, got %d: %+v", len(fc), fc)
	}
	if fc[0].Field != "card_faces[0].oracle_text" {
		t.Errorf("expected card_faces[0].oracle_text, got %s", fc[0].Field)
	}
}

func TestCompareFields_CardFaceCount(t *testing.T) {
	old := card{
		Layout: "transform",
		CardFaces: []cardFace{
			{Name: "Front"},
		},
	}
	new := card{
		Layout: "transform",
		CardFaces: []cardFace{
			{Name: "Front"},
			{Name: "Back"},
		},
	}
	fc := compareFields(old, new)
	found := false
	for _, f := range fc {
		if f.Field == "card_faces.count" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected card_faces.count change, got: %+v", fc)
	}
}

func TestCompareFields_CardFacePowerToughness(t *testing.T) {
	old := card{
		Layout: "transform",
		CardFaces: []cardFace{
			{Name: "Front", Power: "1", Toughness: "1"},
		},
	}
	new := card{
		Layout: "transform",
		CardFaces: []cardFace{
			{Name: "Front", Power: "2", Toughness: "3"},
		},
	}
	fc := compareFields(old, new)
	fields := map[string]bool{}
	for _, f := range fc {
		fields[f.Field] = true
	}
	if !fields["card_faces[0].power"] || !fields["card_faces[0].toughness"] {
		t.Errorf("expected power and toughness face changes, got: %+v", fc)
	}
}
