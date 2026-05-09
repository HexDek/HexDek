// Distributed-compute (BOINC-style) protocol for hexdek.
//
// Community contributors run a lightweight client (cmd/hexdek-contrib)
// that connects via WebSocket, receives signed work chunks (deck
// matchups to simulate), runs them with the local game engine, and
// returns results. Validated results earn credits that surface on the
// operator profile.
//
// The protocol is two-way over a single WS connection:
//
//   client → server   "ready"      {worker_info}
//   server → client   "assignment" {chunk + signature}
//   client → server   "result"     {chunk_id, winners, hash, sig}
//   server → client   "ack"        {accepted, credits_awarded, reason?}
//
// Signing
// -------
// Server signs every assignment with HMAC-SHA256 over the canonical
// chunk bytes; client signs every result with the same key over a
// canonical result envelope. Either side rejects unsigned/forged
// messages outright. The shared key is rotated periodically and lives
// in HEXDEK_CONTRIB_HMAC_KEY (server-issued at session creation).
//
// Validation
// ----------
// The server spot-checks ~3% of returned chunks by re-running them
// locally with the same seed and comparing winner sequences. A
// mismatch flags the chunk; repeated mismatches from the same
// contributor trip the 3-sigma anomaly detector and freeze credits
// pending review.
package hexapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// ContribAssignment is a single unit of distributed work. The server
// marshals it canonically (sorted JSON keys) before signing; clients
// verify the signature before running anything.
type ContribAssignment struct {
	ChunkID     string   `json:"chunk_id"`
	IssuedAt    int64    `json:"issued_at"`
	NSeats      int      `json:"n_seats"`     // 2 or 4 typically
	GamesCount  int      `json:"games_count"` // games to simulate
	Seed        int64    `json:"seed"`        // master seed; per-game = seed + idx*1000+1
	Decks       []string `json:"decks"`       // raw decklist text, one per seat
	DeckKeys    []string `json:"deck_keys"`   // stable identifiers for each deck (for credit attribution)
	MaxTurns    int      `json:"max_turns"`   // per-game turn cap
	Difficulty  int      `json:"difficulty"`  // estimated work units (≈ games × seats)
	HMACKey     string   `json:"hmac_key,omitempty"` // only set on the welcome message; never on assignments themselves
	Signature   string   `json:"signature"`
}

// ContribResult is what the client returns after running an assignment.
// Winners is one entry per game (the deck index that won; -1 = draw).
// OutcomeHash is SHA-256 over the canonical winner+elimination bytes —
// the server compares hashes when spot-checking.
type ContribResult struct {
	ChunkID       string   `json:"chunk_id"`
	StartedAt     int64    `json:"started_at"`
	FinishedAt    int64    `json:"finished_at"`
	ElapsedMS     int64    `json:"elapsed_ms"`
	Winners       []int    `json:"winners"`         // len == GamesCount
	TurnCounts    []int    `json:"turn_counts"`     // len == GamesCount
	OutcomeHash   string   `json:"outcome_hash"`
	WorkerVersion string   `json:"worker_version"`
	Signature     string   `json:"signature"`
}

// ContribAck is the server's reply to a result submission.
type ContribAck struct {
	ChunkID         string  `json:"chunk_id"`
	Accepted        bool    `json:"accepted"`
	CreditsAwarded  int64   `json:"credits_awarded"`
	Reason          string  `json:"reason,omitempty"` // populated on reject or flag
	SpotChecked     bool    `json:"spot_checked"`
	AnomalyZScore   float64 `json:"anomaly_z_score,omitempty"`
}

// ContribReady is sent by the client when it's ready for work.
type ContribReady struct {
	WorkerVersion string `json:"worker_version"`
	NumCPU        int    `json:"num_cpu"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
}

// ContribWelcome is the server's initial reply on connect: confirms
// auth, names the contributor's owner slug, and ships the per-session
// HMAC key the client uses to sign result envelopes.
type ContribWelcome struct {
	Owner       string `json:"owner"`
	HMACKey     string `json:"hmac_key"`
	ServerTime  int64  `json:"server_time"`
	ProtocolVer int    `json:"protocol_version"`
}

// ContribEnvelope is the on-the-wire framing.
type ContribEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ProtocolVersion identifies wire-compat. Bump on breaking changes.
const ProtocolVersion = 1

// canonicalAssignmentBytes serializes the signable subset of an
// assignment (everything except Signature/HMACKey) with sorted keys so
// signing is deterministic across implementations.
func canonicalAssignmentBytes(a *ContribAssignment) ([]byte, error) {
	m := map[string]any{
		"chunk_id":    a.ChunkID,
		"issued_at":   a.IssuedAt,
		"n_seats":     a.NSeats,
		"games_count": a.GamesCount,
		"seed":        a.Seed,
		"decks":       a.Decks,
		"deck_keys":   a.DeckKeys,
		"max_turns":   a.MaxTurns,
		"difficulty":  a.Difficulty,
	}
	return canonicalJSON(m)
}

func canonicalResultBytes(r *ContribResult) ([]byte, error) {
	m := map[string]any{
		"chunk_id":       r.ChunkID,
		"started_at":     r.StartedAt,
		"finished_at":    r.FinishedAt,
		"elapsed_ms":     r.ElapsedMS,
		"winners":        r.Winners,
		"turn_counts":    r.TurnCounts,
		"outcome_hash":   r.OutcomeHash,
		"worker_version": r.WorkerVersion,
	}
	return canonicalJSON(m)
}

// canonicalJSON marshals a map with sorted top-level keys so two callers
// agree byte-for-byte. Nested values are encoded with the standard
// json.Marshal which is stable for slices and primitives but not for
// nested maps; we keep the assignment/result schemas flat to sidestep
// that. (sub-slices are encoded in their declared order — Decks is a
// slice not a map, so order is preserved.)
func canonicalJSON(m map[string]any) ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := []byte{'{'}
	for i, k := range keys {
		if i > 0 {
			out = append(out, ',')
		}
		kb, _ := json.Marshal(k)
		out = append(out, kb...)
		out = append(out, ':')
		vb, err := json.Marshal(m[k])
		if err != nil {
			return nil, err
		}
		out = append(out, vb...)
	}
	out = append(out, '}')
	return out, nil
}

// SignAssignment fills in a.Signature with HMAC-SHA256 over the
// canonical body. Returns the signature so callers can log it.
func SignAssignment(a *ContribAssignment, key []byte) (string, error) {
	body, err := canonicalAssignmentBytes(a)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(body)
	a.Signature = hex.EncodeToString(mac.Sum(nil))
	return a.Signature, nil
}

// VerifyAssignment returns nil iff a.Signature is a valid HMAC over the
// canonical body using `key`. Constant-time compared.
func VerifyAssignment(a *ContribAssignment, key []byte) error {
	if a == nil {
		return errors.New("nil assignment")
	}
	if a.Signature == "" {
		return errors.New("missing signature")
	}
	body, err := canonicalAssignmentBytes(a)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(body)
	want := mac.Sum(nil)
	got, err := hex.DecodeString(a.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !hmac.Equal(want, got) {
		return errors.New("signature mismatch")
	}
	return nil
}

// SignResult fills in r.Signature with HMAC-SHA256 over the canonical
// result body using `key`.
func SignResult(r *ContribResult, key []byte) (string, error) {
	body, err := canonicalResultBytes(r)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(body)
	r.Signature = hex.EncodeToString(mac.Sum(nil))
	return r.Signature, nil
}

// VerifyResult validates a result's signature.
func VerifyResult(r *ContribResult, key []byte) error {
	if r == nil {
		return errors.New("nil result")
	}
	if r.Signature == "" {
		return errors.New("missing signature")
	}
	body, err := canonicalResultBytes(r)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(body)
	want := mac.Sum(nil)
	got, err := hex.DecodeString(r.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !hmac.Equal(want, got) {
		return errors.New("signature mismatch")
	}
	return nil
}

// HashOutcomes computes the SHA-256 over the winner sequence + turn
// counts. Both sides must agree byte-for-byte for a chunk to validate.
func HashOutcomes(winners, turnCounts []int) string {
	h := sha256.New()
	// length-prefixed encoding so nil != [] != [0] etc.
	binAppend := func(n int) {
		var b [8]byte
		// little-endian, signed
		v := uint64(int64(n))
		for i := 0; i < 8; i++ {
			b[i] = byte(v >> (8 * i))
		}
		h.Write(b[:])
	}
	binAppend(len(winners))
	for _, w := range winners {
		binAppend(w)
	}
	binAppend(len(turnCounts))
	for _, t := range turnCounts {
		binAppend(t)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// CreditsForChunk computes the credit award for a validated chunk.
// Default formula: 1 credit per (game × seat). A 100-game 4-seat chunk
// = 400 credits. Anomaly-flagged chunks are credited at 0; spot-check
// failures at 0 with a strike recorded.
func CreditsForChunk(games, seats int) int64 {
	if games <= 0 || seats <= 0 {
		return 0
	}
	return int64(games) * int64(seats)
}
