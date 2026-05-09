package hexapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hexdek/hexdek/internal/db"
)

func sampleAssignment() *ContribAssignment {
	return &ContribAssignment{
		ChunkID:    "chunk-1",
		IssuedAt:   1700000000,
		NSeats:     4,
		GamesCount: 50,
		Seed:       42,
		Decks:      []string{"d0", "d1", "d2", "d3"},
		DeckKeys:   []string{"a/d0", "a/d1", "a/d2", "a/d3"},
		MaxTurns:   80,
		Difficulty: 200,
	}
}

func TestSignAssignment_RoundTrip(t *testing.T) {
	a := sampleAssignment()
	key := []byte("test-key")
	if _, err := SignAssignment(a, key); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if a.Signature == "" {
		t.Fatal("signature empty after sign")
	}
	if err := VerifyAssignment(a, key); err != nil {
		t.Errorf("verify with same key: %v", err)
	}
	if err := VerifyAssignment(a, []byte("wrong-key")); err == nil {
		t.Error("verify with wrong key should fail")
	}
}

func TestSignAssignment_Tamper(t *testing.T) {
	a := sampleAssignment()
	key := []byte("k")
	_, _ = SignAssignment(a, key)

	// Tamper a single field — signature must no longer verify.
	a.Seed = 99
	if err := VerifyAssignment(a, key); err == nil {
		t.Error("tampered assignment verified — signature is not catching field changes")
	}

	// Restore and re-sign; verify should succeed again.
	a.Seed = sampleAssignment().Seed
	_, _ = SignAssignment(a, key)
	if err := VerifyAssignment(a, key); err != nil {
		t.Errorf("re-signed assignment: %v", err)
	}
}

func TestSignAssignment_DeterministicCanonicalization(t *testing.T) {
	// Two clones should produce the same signature byte-for-byte
	// regardless of map iteration order.
	a := sampleAssignment()
	b := sampleAssignment()
	key := []byte("k")
	sa, _ := SignAssignment(a, key)
	sb, _ := SignAssignment(b, key)
	if sa != sb {
		t.Errorf("signatures differ across clones: %s vs %s", sa, sb)
	}
}

func TestSignResult_RoundTrip(t *testing.T) {
	r := &ContribResult{
		ChunkID:    "chunk-1",
		StartedAt:  1, FinishedAt: 5,
		ElapsedMS: 4000,
		Winners:   []int{0, 1, 0, -1},
		TurnCounts: []int{12, 8, 14, 80},
		OutcomeHash: HashOutcomes([]int{0, 1, 0, -1}, []int{12, 8, 14, 80}),
		WorkerVersion: "test/1",
	}
	key := []byte("k")
	if _, err := SignResult(r, key); err != nil {
		t.Fatalf("sign: %v", err)
	}
	if err := VerifyResult(r, key); err != nil {
		t.Errorf("verify: %v", err)
	}
	r.Winners[0] = 2 // tamper
	if err := VerifyResult(r, key); err == nil {
		t.Error("tampered result verified")
	}
}

func TestHashOutcomes_LengthSensitive(t *testing.T) {
	// Length is part of the prefix so {[0,0]} hashes differently from
	// {[0]}, and order matters: {[0,1]} != {[1,0]}. nil and []int{}
	// intentionally collide (both have length 0) — the wire format
	// can't distinguish them.
	cases := [][]int{
		{},
		{0},
		{0, 0},
		{0, 1},
		{1, 0},
	}
	seen := make(map[string]struct{})
	for _, c := range cases {
		h := HashOutcomes(c, c)
		if _, ok := seen[h]; ok {
			t.Errorf("hash collision on input %v", c)
		}
		seen[h] = struct{}{}
	}
}

func TestCreditsForChunk(t *testing.T) {
	if got := CreditsForChunk(100, 4); got != 400 {
		t.Errorf("100×4 = %d, want 400", got)
	}
	if got := CreditsForChunk(0, 4); got != 0 {
		t.Errorf("0×4 = %d, want 0", got)
	}
	if got := CreditsForChunk(10, -1); got != 0 {
		t.Errorf("10×-1 = %d, want 0", got)
	}
}

func TestParseSpotCheckPct(t *testing.T) {
	cases := []struct {
		in   string
		def  float64
		want float64
	}{
		{"", 0.05, 0.05},
		{"3", 0.05, 0.03},
		{"0", 0.05, 0.0},
		{"100", 0.05, 1.0},
		{"-1", 0.05, 0.05},   // out of range, fall back to default
		{"abc", 0.05, 0.05},  // unparseable
		{"101", 0.05, 0.05},  // out of range
	}
	for _, c := range cases {
		got := ParseSpotCheckPct(c.in, c.def)
		if got != c.want {
			t.Errorf("ParseSpotCheckPct(%q, %v) = %v, want %v", c.in, c.def, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------
// Dispatcher behavior — credit accrual, anomaly freeze, spot-check
// rejection. Uses an in-memory SQLite via internal/db.Open.
// ---------------------------------------------------------------------

func TestProcessResult_AcceptsValid(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	d := NewContribDispatcher(conn)
	d.SpotCheckPct = 0 // never spot-check in this test

	a := sampleAssignment()
	a.GamesCount = 10
	a.IssuedAt = time.Now().Unix()
	keyBytes := []byte("dispatcher-key-bytes-for-test-32") // 32 bytes
	if _, err := SignAssignment(a, keyBytes); err != nil {
		t.Fatalf("sign: %v", err)
	}

	winners := make([]int, 10)
	turns := make([]int, 10)
	for i := range winners {
		winners[i] = i % 4
		turns[i] = 10 + i
	}
	r := &ContribResult{
		ChunkID:       a.ChunkID,
		Winners:       winners,
		TurnCounts:    turns,
		OutcomeHash:   HashOutcomes(winners, turns),
		ElapsedMS:     5000,
		WorkerVersion: "test",
	}
	if _, err := SignResult(r, keyBytes); err != nil {
		t.Fatalf("sign result: %v", err)
	}

	// Pre-create the chunk row so FinalizeChunkRow has something to update.
	_ = db.RecordChunkAssignment(context.Background(), conn, a.ChunkID, "alice", a.GamesCount, a.NSeats, a.IssuedAt)

	ack := d.processResult(context.Background(), "alice", r, keyBytes, a)
	if !ack.Accepted {
		t.Fatalf("expected accept, got reject: %s", ack.Reason)
	}
	if ack.CreditsAwarded != int64(a.GamesCount*a.NSeats) {
		t.Errorf("credits: got %d want %d", ack.CreditsAwarded, a.GamesCount*a.NSeats)
	}
	if ack.SpotChecked {
		t.Error("SpotCheckPct=0 but spot check ran")
	}

	// Credits row should reflect the award.
	c, err := db.GetContributorCredits(context.Background(), conn, "alice")
	if err != nil {
		t.Fatalf("get credits: %v", err)
	}
	if c.CreditsTotal != int64(a.GamesCount*a.NSeats) {
		t.Errorf("credits_total: got %d want %d", c.CreditsTotal, a.GamesCount*a.NSeats)
	}
	if c.ChunksCompleted != 1 {
		t.Errorf("chunks_completed: %d, want 1", c.ChunksCompleted)
	}
	if c.Frozen {
		t.Error("contributor frozen on first chunk — anomaly detector firing too early")
	}
}

func TestProcessResult_RejectsBadSignature(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	d := NewContribDispatcher(conn)
	d.SpotCheckPct = 0

	a := sampleAssignment()
	keyBytes := []byte("real-key-bytes-32-bytes-padding!")
	_, _ = SignAssignment(a, keyBytes)

	winners := []int{0, 1, 2, 3}
	turns := []int{10, 11, 12, 13}
	r := &ContribResult{
		ChunkID:     a.ChunkID,
		Winners:     winners,
		TurnCounts:  turns,
		OutcomeHash: HashOutcomes(winners, turns),
		Signature:   "not-a-real-signature",
	}
	a.GamesCount = 4
	ack := d.processResult(context.Background(), "bob", r, keyBytes, a)
	if ack.Accepted {
		t.Error("forged-signature result was accepted")
	}
	if !strings.Contains(ack.Reason, "signature") {
		t.Errorf("reason should mention signature, got %q", ack.Reason)
	}
}

func TestProcessResult_RejectsHashMismatch(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	d := NewContribDispatcher(conn)
	d.SpotCheckPct = 0

	a := sampleAssignment()
	a.GamesCount = 4
	keyBytes := []byte("k")
	_, _ = SignAssignment(a, keyBytes)

	winners := []int{0, 1, 2, 3}
	turns := []int{10, 11, 12, 13}
	r := &ContribResult{
		ChunkID:     a.ChunkID,
		Winners:     winners,
		TurnCounts:  turns,
		OutcomeHash: "deadbeef", // wrong
	}
	_, _ = SignResult(r, keyBytes)
	ack := d.processResult(context.Background(), "carol", r, keyBytes, a)
	if ack.Accepted {
		t.Error("hash-mismatched result was accepted")
	}
	if !strings.Contains(ack.Reason, "hash") {
		t.Errorf("reason should mention hash, got %q", ack.Reason)
	}
}

func TestProcessResult_SpotCheckParityFailureRejects(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	d := NewContribDispatcher(conn)
	d.SpotCheckPct = 1.0 // always spot-check
	d.SpotCheck = func(*ContribAssignment) ([]int, []int, bool) {
		// Local re-run produces DIFFERENT winners than the client claimed.
		return []int{1, 0, 3, 2}, []int{20, 20, 20, 20}, true
	}

	a := sampleAssignment()
	a.GamesCount = 4
	keyBytes := []byte("k")
	_, _ = SignAssignment(a, keyBytes)

	winners := []int{0, 1, 2, 3}
	turns := []int{10, 11, 12, 13}
	r := &ContribResult{
		ChunkID:     a.ChunkID,
		Winners:     winners,
		TurnCounts:  turns,
		OutcomeHash: HashOutcomes(winners, turns),
	}
	_, _ = SignResult(r, keyBytes)
	ack := d.processResult(context.Background(), "dave", r, keyBytes, a)
	if ack.Accepted {
		t.Error("spot-check mismatch was accepted")
	}
	if !ack.SpotChecked {
		t.Error("ack should reflect spot_checked=true")
	}
	if !strings.Contains(ack.Reason, "spot-check") {
		t.Errorf("reason should mention spot-check, got %q", ack.Reason)
	}
}

func TestQueueFIFOAndRequeue(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	d := NewContribDispatcher(conn)

	a1 := &ContribAssignment{ChunkID: "1"}
	a2 := &ContribAssignment{ChunkID: "2"}
	d.Enqueue(a1)
	d.Enqueue(a2)

	if d.PendingCount() != 2 {
		t.Errorf("pending count: %d", d.PendingCount())
	}
	got := d.dequeue()
	if got.ChunkID != "1" {
		t.Errorf("FIFO violated: got %q first", got.ChunkID)
	}
	d.requeue(got)
	if d.PendingCount() != 2 {
		t.Errorf("after requeue pending: %d", d.PendingCount())
	}
	got = d.dequeue()
	if got.ChunkID != "1" {
		t.Errorf("requeue should restore at front, got %q", got.ChunkID)
	}
}

