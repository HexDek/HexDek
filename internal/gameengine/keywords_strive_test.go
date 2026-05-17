package gameengine

import (
	"math/rand"
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// ---------------------------------------------------------------------------
// Test helpers — Strive (CR §702.118)
// ---------------------------------------------------------------------------

func newStriveGame(t *testing.T) *GameState {
	t.Helper()
	rng := rand.New(rand.NewSource(23))
	return NewGameState(2, rng, nil)
}

// striveHandCard builds an instant with explicit base CMC and a strive
// keyword carrying `strivePerTarget` as its arg. The AST is tagged with
// the modern Keyword node so HasStrive's keyword path covers it; a
// separate test exercises the oracle-text fallback.
func striveHandCard(gs *GameState, seat int, name string, baseCMC, strivePerTarget int) *Card {
	c := &Card{
		Name:  name,
		Owner: seat,
		CMC:   baseCMC,
		Types: []string{"instant"},
		Colors: []string{"W"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "strive", Args: []any{float64(strivePerTarget)}},
			},
		},
	}
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, c)
	return c
}

// striveOracleOnlyCard builds a strive card whose AST does NOT carry the
// keyword tag — strive shows up only in the raw oracle text via a
// Static ability. Mirrors how older Theros-block cards parse.
func striveOracleOnlyCard(gs *GameState, seat int, name string, baseCMC int) *Card {
	c := &Card{
		Name:  name,
		Owner: seat,
		CMC:   baseCMC,
		Types: []string{"sorcery"},
		AST: &gameast.CardAST{
			Name: name,
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Strive — This spell costs {1}{W} more to cast for each target beyond the first. Up to N target creatures get +2/+2 until end of turn."},
			},
		},
	}
	gs.Seats[seat].Hand = append(gs.Seats[seat].Hand, c)
	return c
}

// ===========================================================================
// (a) StriveCastCost: 1, 2, 5 target arithmetic
// ===========================================================================

func TestStriveCastCost_OneTargetCostsBase(t *testing.T) {
	c := &Card{CMC: 3}
	if got := StriveCastCost(c, 2, 1); got != 3 {
		t.Errorf("1 target: want base 3, got %d", got)
	}
}

func TestStriveCastCost_TwoTargetsAddsOneStrive(t *testing.T) {
	c := &Card{CMC: 3}
	if got := StriveCastCost(c, 2, 2); got != 5 {
		t.Errorf("2 targets: want 3 + 1*2 = 5, got %d", got)
	}
}

func TestStriveCastCost_FiveTargetsAddsFourStrives(t *testing.T) {
	c := &Card{CMC: 3}
	if got := StriveCastCost(c, 2, 5); got != 11 {
		t.Errorf("5 targets: want 3 + 4*2 = 11, got %d", got)
	}
}

func TestStriveCastCost_Table(t *testing.T) {
	// Cross-check table form so any future regression in the formula
	// is obvious in failure output.
	type row struct {
		cmc, strive, n, want int
	}
	rows := []row{
		{cmc: 1, strive: 1, n: 1, want: 1},
		{cmc: 1, strive: 1, n: 2, want: 2},
		{cmc: 1, strive: 1, n: 4, want: 4},
		{cmc: 2, strive: 3, n: 3, want: 2 + 2*3},
		{cmc: 4, strive: 2, n: 5, want: 4 + 4*2},
		{cmc: 0, strive: 5, n: 6, want: 0 + 5*5},
	}
	for _, r := range rows {
		got := StriveCastCost(&Card{CMC: r.cmc}, r.strive, r.n)
		if got != r.want {
			t.Errorf("StriveCastCost(cmc=%d strive=%d n=%d) = %d, want %d",
				r.cmc, r.strive, r.n, got, r.want)
		}
	}
}

// ===========================================================================
// (b) Zero-target case
// ===========================================================================

func TestStriveCastCost_ZeroTargetsPaysOnlyBase(t *testing.T) {
	c := &Card{CMC: 4}
	if got := StriveCastCost(c, 3, 0); got != 4 {
		t.Errorf("0 targets: want base 4, got %d (the max(0, -1) guard must clamp)", got)
	}
}

func TestStriveCastCost_NilCardIsZero(t *testing.T) {
	if got := StriveCastCost(nil, 2, 5); got != 0 {
		t.Errorf("nil card: want 0, got %d", got)
	}
}

func TestStriveCastCost_NegativeStriveCostClamped(t *testing.T) {
	// Malformed strive arg should never refund mana.
	c := &Card{CMC: 3}
	if got := StriveCastCost(c, -2, 5); got != 3 {
		t.Errorf("negative strive cost should clamp to 0 extra; want base 3, got %d", got)
	}
}

// ===========================================================================
// (c) CostMeta["strive_targets"] stamped on the StackItem
// ===========================================================================

func TestCastWithStrive_StampsStriveTargetsOnCostMeta(t *testing.T) {
	gs := newStriveGame(t)
	gs.Active = 0
	// Cast Triplicate Spirits-style: base 4, strive {1}, 3 targets ⇒ 4 + 2*1 = 6
	gs.Seats[0].ManaPool = 6
	card := striveHandCard(gs, 0, "Triplicate Spirits", 4, 1)

	res, err := CastWithStrive(gs, 0, card, 1, 3)
	if err != nil {
		t.Fatalf("CastWithStrive: %v", err)
	}
	if res == nil {
		t.Fatal("CastWithStrive returned nil result on success")
	}

	if len(gs.Stack) != 1 {
		t.Fatalf("expected 1 stack item, got %d", len(gs.Stack))
	}
	meta := gs.Stack[0].CostMeta
	if meta == nil {
		t.Fatal("StackItem.CostMeta should not be nil")
	}
	if v, _ := meta["strive_targets"].(int); v != 3 {
		t.Errorf("CostMeta[strive_targets] = %v, want 3", meta["strive_targets"])
	}
	if v, _ := meta["strive_per_target"].(int); v != 1 {
		t.Errorf("CostMeta[strive_per_target] = %v, want 1", meta["strive_per_target"])
	}
	if v, _ := meta["strive_total_cost"].(int); v != 6 {
		t.Errorf("CostMeta[strive_total_cost] = %v, want 6", meta["strive_total_cost"])
	}
	if v, _ := meta["strive_extra_paid"].(int); v != 2 {
		t.Errorf("CostMeta[strive_extra_paid] = %v, want 2 (2 extra targets * {1})", meta["strive_extra_paid"])
	}
	if v, _ := meta["strive"].(bool); !v {
		t.Error("CostMeta[strive] should be true")
	}
	if v, _ := meta["alt_cost"].(string); v != "strive" {
		t.Errorf("CostMeta[alt_cost] = %v, want \"strive\"", meta["alt_cost"])
	}
	if gs.Stack[0].CastZone != ZoneHand {
		t.Errorf("CastZone should be ZoneHand, got %v", gs.Stack[0].CastZone)
	}

	// pay_mana and strive_cast events with the right amounts.
	sawPay, sawCast := false, false
	for _, ev := range gs.EventLog {
		if ev.Kind == "pay_mana" && ev.Amount == 6 {
			if r, _ := ev.Details["reason"].(string); r == "strive_cast" {
				sawPay = true
			}
		}
		if ev.Kind == "strive_cast" && ev.Amount == 6 {
			if rule, _ := ev.Details["rule"].(string); rule == "702.118a" {
				sawCast = true
			}
		}
	}
	if !sawPay {
		t.Error("expected a pay_mana event reason=strive_cast amount=6")
	}
	if !sawCast {
		t.Error("expected a strive_cast event with rule 702.118a and amount 6")
	}

	// Per-turn flag for "if a strive spell was cast this turn" predicates.
	if !SpellStriveThisTurn(gs, 0) {
		t.Error("SpellStriveThisTurn should be true after a strive cast")
	}
}

func TestCastWithStrive_OneTargetPaysBaseOnly(t *testing.T) {
	// Boundary: declaring 1 target must NOT add any strive surcharge.
	gs := newStriveGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 3
	card := striveHandCard(gs, 0, "Roused Striver", 3, 2)

	if _, err := CastWithStrive(gs, 0, card, 2, 1); err != nil {
		t.Fatalf("CastWithStrive: %v", err)
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("1 target should pay only base 3; mana left=%d", gs.Seats[0].ManaPool)
	}
	if v, _ := gs.Stack[0].CostMeta["strive_extra_paid"].(int); v != 0 {
		t.Errorf("strive_extra_paid should be 0 for 1 target, got %d", v)
	}
}

func TestCastWithStrive_ZeroTargetsAllowedAndPaysBase(t *testing.T) {
	// Some strive spells allow zero targets ("any number of target
	// creatures"). With 0 targets, total = base, CostMeta records 0.
	gs := newStriveGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 2
	card := striveHandCard(gs, 0, "Optional Strive", 2, 3)

	if _, err := CastWithStrive(gs, 0, card, 3, 0); err != nil {
		t.Fatalf("CastWithStrive with 0 targets: %v", err)
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("0 targets should pay only base 2; mana left=%d", gs.Seats[0].ManaPool)
	}
	if v, _ := gs.Stack[0].CostMeta["strive_targets"].(int); v != 0 {
		t.Errorf("strive_targets should be 0, got %d", v)
	}
	if v, _ := gs.Stack[0].CostMeta["strive_extra_paid"].(int); v != 0 {
		t.Errorf("strive_extra_paid should be 0 for 0 targets, got %d", v)
	}
}

// ===========================================================================
// (d) HasStrive detection — keyword tag + oracle text
// ===========================================================================

func TestHasStrive_DetectsKeyword(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "strive", Args: []any{float64(2)}},
			},
		},
	}
	if !HasStrive(c) {
		t.Fatal("HasStrive should detect modern Keyword tagging")
	}
}

func TestHasStrive_DetectsOracleText(t *testing.T) {
	gs := newStriveGame(t)
	c := striveOracleOnlyCard(gs, 0, "Old Strive Card", 2)
	if !HasStrive(c) {
		t.Fatal("HasStrive should detect via oracle-text \"strive —\" prefix when AST lacks the keyword tag")
	}
}

func TestHasStrive_DetectsAsciiHyphenOracleText(t *testing.T) {
	// Some older corpus dumps use ASCII '-' instead of em-dash '—'.
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Strive - This spell costs {2} more to cast for each target beyond the first."},
			},
		},
	}
	if !HasStrive(c) {
		t.Fatal("HasStrive should detect ASCII-hyphen \"strive - this spell costs\"")
	}
}

func TestHasStrive_NegativeNoKeywordNoText(t *testing.T) {
	c := &Card{AST: &gameast.CardAST{}}
	if HasStrive(c) {
		t.Fatal("HasStrive should be false for a card without keyword and without oracle text")
	}
	if HasStrive(nil) {
		t.Fatal("HasStrive(nil) should be false")
	}
	// A non-strive card whose text incidentally contains the word "strive"
	// but not as the keyword prefix should NOT match. We require
	// "strive —" or "strive -" specifically.
	c2 := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Static{Raw: "Players strive for victory in this format."},
			},
		},
	}
	if HasStrive(c2) {
		t.Fatal("HasStrive should NOT match incidental uses of the word 'strive' in oracle text")
	}
}

func TestStrivePerTargetCost_ParsesArg(t *testing.T) {
	c := &Card{
		AST: &gameast.CardAST{
			Abilities: []gameast.Ability{
				&gameast.Keyword{Name: "strive", Args: []any{float64(2)}},
			},
		},
	}
	if got := StrivePerTargetCost(c); got != 2 {
		t.Errorf("StrivePerTargetCost: want 2, got %d", got)
	}
}

// ===========================================================================
// CastWithStrive — rejection paths + side-effect guards
// ===========================================================================

func TestCastWithStrive_RejectsCardWithoutKeyword(t *testing.T) {
	gs := newStriveGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 10
	c := &Card{Name: "Plain", Owner: 0, CMC: 2, Types: []string{"instant"},
		AST: &gameast.CardAST{Name: "Plain"}}
	gs.Seats[0].Hand = append(gs.Seats[0].Hand, c)

	res, err := CastWithStrive(gs, 0, c, 1, 2)
	if err == nil || res != nil {
		t.Fatal("CastWithStrive should reject a card without the strive keyword")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "no_strive_keyword" {
		t.Errorf("expected CastError(no_strive_keyword), got %T %v", err, err)
	}
	if gs.Seats[0].ManaPool != 10 {
		t.Errorf("mana should be untouched, got %d", gs.Seats[0].ManaPool)
	}
}

func TestCastWithStrive_RejectsInsufficientMana(t *testing.T) {
	gs := newStriveGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 4 // need 5 (base 3 + 1*2)
	card := striveHandCard(gs, 0, "Triplicate Spirits", 3, 2)

	res, err := CastWithStrive(gs, 0, card, 2, 2)
	if err == nil || res != nil {
		t.Fatal("CastWithStrive should reject when mana < total strive cost")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "insufficient_mana" {
		t.Errorf("expected CastError(insufficient_mana), got %T %v", err, err)
	}
	// Card still in hand, mana untouched, stack empty.
	if len(gs.Seats[0].Hand) != 1 {
		t.Errorf("card should remain in hand, hand=%d", len(gs.Seats[0].Hand))
	}
	if gs.Seats[0].ManaPool != 4 {
		t.Errorf("mana pool should be untouched, got %d", gs.Seats[0].ManaPool)
	}
	if len(gs.Stack) != 0 {
		t.Errorf("stack should be empty, got %d", len(gs.Stack))
	}
}

func TestCastWithStrive_RejectsNegativeTargets(t *testing.T) {
	gs := newStriveGame(t)
	gs.Active = 0
	gs.Seats[0].ManaPool = 5
	card := striveHandCard(gs, 0, "Roused Striver", 3, 2)

	_, err := CastWithStrive(gs, 0, card, 2, -1)
	if err == nil {
		t.Fatal("CastWithStrive should reject numTargets < 0")
	}
	if ce, ok := err.(*CastError); !ok || ce.Reason != "negative_targets" {
		t.Errorf("expected CastError(negative_targets), got %T %v", err, err)
	}
}

func TestCastWithStrive_StrivePerTargetMinus1FallsBackToKeywordArg(t *testing.T) {
	gs := newStriveGame(t)
	gs.Active = 0
	// strive {2} keyword arg; 3 targets ⇒ 3 + 2*2 = 7
	gs.Seats[0].ManaPool = 7
	card := striveHandCard(gs, 0, "Roused Striver", 3, 2)

	if _, err := CastWithStrive(gs, 0, card, -1, 3); err != nil {
		t.Fatalf("CastWithStrive with -1 fallback: %v", err)
	}
	if gs.Seats[0].ManaPool != 0 {
		t.Errorf("expected 0 mana after paying 7; got %d", gs.Seats[0].ManaPool)
	}
	if v, _ := gs.Stack[0].CostMeta["strive_per_target"].(int); v != 2 {
		t.Errorf("fallback should resolve to keyword arg 2, got %d", v)
	}
}

func TestCastWithStrive_NilSafety(t *testing.T) {
	if _, err := CastWithStrive(nil, 0, nil, 0, 0); err == nil {
		t.Fatal("CastWithStrive(nil...) should error")
	}
	gs := newStriveGame(t)
	if _, err := CastWithStrive(gs, -1, nil, 0, 0); err == nil {
		t.Fatal("CastWithStrive(invalid seat) should error")
	}
	if _, err := CastWithStrive(gs, 0, nil, 0, 0); err == nil {
		t.Fatal("CastWithStrive(nil card) should error")
	}
}

func TestSpellStriveThisTurn_FalseBeforeCast(t *testing.T) {
	gs := newStriveGame(t)
	if SpellStriveThisTurn(gs, 0) {
		t.Fatal("SpellStriveThisTurn should be false before any cast")
	}
}
