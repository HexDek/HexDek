package per_card

import (
	"testing"
)

// dev/etb-stub-handlers — tests for the new custom handlers added to
// fill gen_*.go ETB stubs. Mabel and Rendmaw kept their hand-edited
// era1 implementations (covered by era1_handlers_test.go); Lord Xander
// and Hidetsugu were already covered by era3_batch_test.go; this file
// covers the two genuinely-new handlers (Zegana adds an oracle-correct
// draw, and Karumonix replaces a parser_gap-only stub).

// ---------------------------------------------------------------------------
// Prime Speaker Zegana — additional coverage on top of era1 tests.
// Verifies the new oracle-correct draw count matches X (greatest power),
// not Zegana's post-counter power.
// ---------------------------------------------------------------------------

func TestPrimeSpeakerZeganaCustom_DrawsXNotPostCounterPower(t *testing.T) {
	gs := newGame(t, 2)
	zegana := addPerm(gs, 0, "Prime Speaker Zegana", "creature", "legendary")
	zegana.Card.BasePower = 1
	zegana.Card.BaseToughness = 1

	// Two other creatures: 3-power and 7-power. X = 7.
	a := addPerm(gs, 0, "Bear", "creature")
	a.Card.BasePower = 3
	a.Card.BaseToughness = 3
	b := addPerm(gs, 0, "Wurm", "creature")
	b.Card.BasePower = 7
	b.Card.BaseToughness = 7

	addLibrary(gs, 0, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8", "L9", "L10")

	primeSpeakerZeganaETB(gs, zegana)

	if zegana.Counters["+1/+1"] != 7 {
		t.Fatalf("expected 7 counters from greatest other power; got %d", zegana.Counters["+1/+1"])
	}
	if len(gs.Seats[0].Hand) != 7 {
		t.Fatalf("expected 7 cards drawn (X=7); hand=%d", len(gs.Seats[0].Hand))
	}
}

// ---------------------------------------------------------------------------
// Karumonix, the Rat King — ETB drops a poison counter on each opponent
// for each Rat the controller has on the battlefield (Karumonix herself
// counts since she's a Rat).
// ---------------------------------------------------------------------------

func TestKarumonix_ETBPoisonsEachOpponentPerRat(t *testing.T) {
	gs := newGame(t, 4)
	karu := addPerm(gs, 0, "Karumonix, the Rat King", "creature", "legendary", "rat")
	addPerm(gs, 0, "Pack Rat", "creature", "rat")
	addPerm(gs, 0, "Marrow Gnawer", "creature", "rat", "legendary")
	addPerm(gs, 0, "Plains", "land") // non-rat noise

	karumonixETBPoison(gs, karu)

	// 3 rats (Karumonix + Pack Rat + Marrow Gnawer) → 3 poison per opp.
	for i := 1; i < 4; i++ {
		if gs.Seats[i].PoisonCounters != 3 {
			t.Errorf("seat %d: expected 3 poison counters; got %d",
				i, gs.Seats[i].PoisonCounters)
		}
	}
	// Controller takes none.
	if gs.Seats[0].PoisonCounters != 0 {
		t.Errorf("controller should not poison themselves; got %d",
			gs.Seats[0].PoisonCounters)
	}
}

func TestKarumonix_ETBNoRatsNoPoison(t *testing.T) {
	gs := newGame(t, 2)
	// Place Karumonix without the rat type — defensive: should still
	// register zero rats and noop.
	karu := addPerm(gs, 0, "Karumonix, the Rat King", "creature", "legendary")

	karumonixETBPoison(gs, karu)

	if gs.Seats[1].PoisonCounters != 0 {
		t.Errorf("0 rats → 0 poison; opp got %d", gs.Seats[1].PoisonCounters)
	}
}

func TestKarumonix_ETBSkipsLostOpponents(t *testing.T) {
	gs := newGame(t, 4)
	karu := addPerm(gs, 0, "Karumonix, the Rat King", "creature", "legendary", "rat")
	gs.Seats[2].Lost = true

	karumonixETBPoison(gs, karu)

	if gs.Seats[1].PoisonCounters != 1 || gs.Seats[3].PoisonCounters != 1 {
		t.Errorf("living opps should each get 1 poison; got s1=%d s3=%d",
			gs.Seats[1].PoisonCounters, gs.Seats[3].PoisonCounters)
	}
	if gs.Seats[2].PoisonCounters != 0 {
		t.Errorf("eliminated seat should be skipped; got %d",
			gs.Seats[2].PoisonCounters)
	}
}
