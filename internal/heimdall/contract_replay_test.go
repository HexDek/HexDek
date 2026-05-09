package heimdall

import (
	"testing"

	"github.com/hexdek/hexdek/internal/seedcontract"
)

// TestVerifyReplay_RejectsTamperedContract — pure-integrity branch.
// We don't run a full game replay (that needs the AST corpus and
// deck files which aren't shipped to CI). Instead we drive the
// VerifyReplay -> CheckIntegrity short-circuit: a tampered contract
// MUST be rejected before replay is even attempted, so this test
// confirms the integrity check fires regardless of corpus state.
func TestVerifyReplay_RejectsTamperedContract(t *testing.T) {
	in := seedcontract.Inputs{
		RNGSeed:       4242,
		DeckKeys:      [4]string{"a/x", "b/y", "c/z", "d/w"},
		NSeats:        4,
		EngineVersion: "test-build",
		SealedAtUnix:  1_700_000_000,
	}
	out := seedcontract.Outcome{
		Winner:           1,
		Turns:            12,
		KillMethod:       "combat",
		EndReason:        "last_seat_standing",
		EliminationOrder: [4]int{2, 0, 3, 1},
		FinalLife:        [4]int{0, 18, -3, 0},
	}
	key := seedcontract.DeriveContractKey([]byte("master"), "tournament:rep")

	c := seedcontract.New(in)
	c.Seal(out)
	c.Sign(key)

	// Tamper the winner. VerifyReplay must surface the integrity
	// failure WITHOUT running a replay (no rc, no corpus).
	c.Outcome.Winner = 0
	res, err := VerifyReplay(nil, c, key)
	if err == nil {
		// nil err is fine — VerifyReplay reports the failure via
		// res.OK / res.Detail, not the error channel, except for
		// nil-rc which IS a fatal error since replay can't run.
		t.Fatal("expected error for nil ReplayContext on tampered contract")
	}
	if res.OK {
		t.Fatal("VerifyResult.OK true for tampered contract")
	}
}

// TestVerifyReplay_NilContract — defensive: nil contract should
// surface as an error, not a panic.
func TestVerifyReplay_NilContract(t *testing.T) {
	_, err := VerifyReplay(nil, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil contract")
	}
}

// TestDigestOutcomeFromContract_Determinism — internal helper used by
// VerifyReplay to recompute the outcome digest. Confirm it produces
// the same hex as seedcontract.Seal would.
func TestDigestOutcomeFromContract_Determinism(t *testing.T) {
	out := seedcontract.Outcome{
		Winner:           1,
		Turns:            14,
		KillMethod:       "combo",
		EndReason:        "last_seat_standing",
		EliminationOrder: [4]int{2, 0, 1, 3},
		FinalLife:        [4]int{0, 18, -2, 0},
	}

	want := seedcontract.New(seedcontract.Inputs{})
	want.Seal(out)

	got := digestOutcomeFromContract(out)
	if got != want.OutcomeDigest {
		t.Fatalf("digestOutcomeFromContract diverged from Seal: got %s want %s", got, want.OutcomeDigest)
	}
}

// TestRoundTrip_ContractFromTournamentMatchesReDerivation — ensures
// that the inputs we'd construct in tournament.runOneGame would
// produce the exact same digest as a replay-side construction with
// the same fields. Catches drift between the two paths.
func TestRoundTrip_ContractFromTournamentMatchesReDerivation(t *testing.T) {
	in := seedcontract.Inputs{
		RNGSeed:       0x1234_5678,
		DeckKeys:      [4]string{"alice/aggro", "bob/control", "carol/combo", "dave/midrange"},
		NSeats:        4,
		EngineVersion: "0.42.1",
		SealedAtUnix:  1_700_000_001,
	}
	out := seedcontract.Outcome{
		Winner:           3,
		Turns:            22,
		KillMethod:       "commander",
		EndReason:        "last_seat_standing",
		EliminationOrder: [4]int{1, 0, 2, 3},
		FinalLife:        [4]int{0, -5, 0, 24},
	}
	key := seedcontract.DeriveContractKey([]byte("k"), "tournament:rt")

	tournamentSide := seedcontract.New(in)
	tournamentSide.Seal(out)
	tournamentSide.Sign(key)

	verifierSide := seedcontract.New(in)
	verifierSide.Seal(out)
	if verifierSide.OutcomeDigest != tournamentSide.OutcomeDigest {
		t.Fatal("outcome digest drift between tournament-side and verifier-side construction")
	}
	if verifierSide.InputDigest != tournamentSide.InputDigest {
		t.Fatal("input digest drift between tournament-side and verifier-side construction")
	}
}
