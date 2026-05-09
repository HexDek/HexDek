package seedcontract

import (
	"encoding/json"
	"testing"
)

// fixtureInputs returns a representative game-start input set. Keep
// the fields varied (different deck strings, non-zero RNG, non-default
// NSeats) so a digest collision under naive concatenation surfaces in
// the determinism test.
func fixtureInputs() Inputs {
	return Inputs{
		RNGSeed:       0x4242_1337_DEAD_BEEF,
		DeckKeys:      [4]string{"alice/aggro", "bob/control", "carol/combo", "dave/midrange"},
		NSeats:        4,
		EngineVersion: "0.42.1+heimdall.b0d",
		SealedAtUnix:  1_700_000_000,
	}
}

func fixtureOutcome() Outcome {
	return Outcome{
		Winner:           2,
		Turns:            14,
		KillMethod:       "combo",
		EndReason:        "last_seat_standing",
		EliminationOrder: [4]int{2, 0, 1, 3},
		FinalLife:        [4]int{0, -3, 18, 0},
	}
}

func TestNew_InputDigestStable(t *testing.T) {
	a := New(fixtureInputs())
	b := New(fixtureInputs())
	if a.InputDigest == "" {
		t.Fatal("input digest empty")
	}
	if a.InputDigest != b.InputDigest {
		t.Fatalf("input digest unstable across constructions: %s vs %s", a.InputDigest, b.InputDigest)
	}
}

func TestSign_Verify_Roundtrip(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)
	if !c.Verify(key) {
		t.Fatal("Verify failed for freshly signed contract")
	}
	if err := c.CheckIntegrity(key); err != nil {
		t.Fatalf("CheckIntegrity failed: %v", err)
	}
}

func TestVerify_FailsWithWrongKey(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	good := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	bad := DeriveContractKey([]byte("master-secret"), "tournament:other")
	c.Sign(good)
	if c.Verify(bad) {
		t.Fatal("Verify accepted contract under wrong key")
	}
}

func TestTamper_Winner(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	// Forge a winner change without re-signing — the most basic attack.
	c.Outcome.Winner = 0
	if c.Verify(key) {
		// Sig is over the OLD outcome digest; Verify alone won't catch
		// the change because OutcomeDigest is still the old hex string.
		// CheckIntegrity must catch it via re-derivation.
	}
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered Outcome.Winner")
	}
}

func TestTamper_Turns(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	c.Outcome.Turns = 99
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered Outcome.Turns")
	}
}

func TestTamper_OutcomeDigestRecomputed(t *testing.T) {
	// Attacker also recomputes OutcomeDigest after tampering. Without
	// re-signing, the HMAC tag still binds to the old digest, so Verify
	// fails outright.
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	tampered := fixtureOutcome()
	tampered.Winner = 0
	c.Outcome = tampered
	c.OutcomeDigest = digestOutcome(tampered)
	if c.Verify(key) {
		t.Fatal("Verify passed with re-derived outcome digest but stale signature")
	}
}

func TestTamper_InputDigest(t *testing.T) {
	// Mutate a deck key after signing without recomputing InputDigest.
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	c.DeckKeys[0] = "mallory/forged"
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered DeckKeys")
	}
}

func TestTamper_RNGSeed(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	c.RNGSeed = c.RNGSeed + 1
	if err := c.CheckIntegrity(key); err == nil {
		t.Fatal("CheckIntegrity passed despite tampered RNGSeed")
	}
}

func TestTamper_SignatureFlip(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	if len(c.Sig) < 2 {
		t.Fatal("signature too short")
	}
	// Flip the first hex character.
	bad := []byte(c.Sig)
	if bad[0] == '0' {
		bad[0] = '1'
	} else {
		bad[0] = '0'
	}
	c.Sig = string(bad)
	if c.Verify(key) {
		t.Fatal("Verify accepted contract with flipped signature byte")
	}
}

func TestVerify_RejectsEmptySig(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	if c.Verify(key) {
		t.Fatal("Verify accepted unsigned contract")
	}
}

func TestVerify_RejectsMalformedSigHex(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)
	c.Sig = "ZZZZZ-not-hex"
	if c.Verify(key) {
		t.Fatal("Verify accepted contract with malformed sig hex")
	}
}

func TestJSON_Roundtrip_PreservesAllFields(t *testing.T) {
	c := New(fixtureInputs())
	c.Seal(fixtureOutcome())
	key := DeriveContractKey([]byte("master-secret"), "tournament:smoke")
	c.Sign(key)

	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var c2 SeedContract
	if err := json.Unmarshal(data, &c2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !c2.Verify(key) {
		t.Fatal("Verify failed after JSON round-trip")
	}
	if err := c2.CheckIntegrity(key); err != nil {
		t.Fatalf("CheckIntegrity failed after JSON round-trip: %v", err)
	}
}

func TestDeriveContractKey_StableUnderSameContext(t *testing.T) {
	master := []byte("super secret server master key")
	a := DeriveContractKey(master, "tournament:abc")
	b := DeriveContractKey(master, "tournament:abc")
	if string(a) != string(b) {
		t.Fatal("DeriveContractKey not stable across calls")
	}
	c := DeriveContractKey(master, "tournament:xyz")
	if string(a) == string(c) {
		t.Fatal("DeriveContractKey returned same key for different context")
	}
}

func TestCanonicalContextForTournament_OrderIndependent(t *testing.T) {
	a := CanonicalContextForTournament([]string{"build:abc", "tid:42", "fmt:cmdr"})
	b := CanonicalContextForTournament([]string{"fmt:cmdr", "tid:42", "build:abc"})
	if a != b {
		t.Fatalf("context not order-independent: %q vs %q", a, b)
	}
}

func TestDigestInputs_OrderSensitive(t *testing.T) {
	// Sanity-check: swapping deck keys MUST change the digest, otherwise
	// our canonicalization is collapsing fields.
	in := fixtureInputs()
	a := digestInputs(in)
	in.DeckKeys[0], in.DeckKeys[1] = in.DeckKeys[1], in.DeckKeys[0]
	b := digestInputs(in)
	if a == b {
		t.Fatal("input digest collapsed deck-key order")
	}
}

func TestSchemaConstant(t *testing.T) {
	c := New(fixtureInputs())
	if c.Schema != Schema {
		t.Fatalf("New() did not stamp Schema: got %d want %d", c.Schema, Schema)
	}
}
