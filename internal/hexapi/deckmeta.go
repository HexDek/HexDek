package hexapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// deck_meta stores per-deck overrides that don't belong in the on-disk
// deck files — currently just custom_name (a user-set display title that
// overrides the derived commander/filename name).
//
// Keyed by (owner, id) which mirrors the URL shape, so the frontend can
// PATCH a deck without us having to look up an internal UUID first.

const deckMetaSchema = `
CREATE TABLE IF NOT EXISTS deck_meta (
    owner        TEXT NOT NULL,
    id           TEXT NOT NULL,
    custom_name  TEXT NOT NULL DEFAULT '',
    cloned_from  TEXT NOT NULL DEFAULT '',
    updated_at   INTEGER NOT NULL,
    PRIMARY KEY (owner, id)
);

-- Per-user clone rate limiter. One row per clone attempt; the cap is
-- enforced by counting rows in the trailing hour for a given owner.
CREATE TABLE IF NOT EXISTS clone_log (
    owner       TEXT NOT NULL,
    src_key     TEXT NOT NULL,
    dst_key     TEXT NOT NULL,
    cloned_at   INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_clone_log_owner_time
    ON clone_log(owner, cloned_at DESC);
`

// EnsureDeckMetaSchema creates the deck_meta table idempotently. Call once
// at startup (right after handing the *sql.DB to the Handler). Also
// idempotently adds the cloned_from column on databases that predate the
// clone feature, so existing installs don't need a migration.
func EnsureDeckMetaSchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("hexapi: nil db for deck_meta schema")
	}
	if _, err := db.ExecContext(ctx, deckMetaSchema); err != nil {
		return err
	}
	// SQLite has no IF NOT EXISTS for ADD COLUMN; ignore the duplicate-
	// column error so this is safe to run repeatedly.
	if _, err := db.ExecContext(ctx,
		`ALTER TABLE deck_meta ADD COLUMN cloned_from TEXT NOT NULL DEFAULT ''`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	return nil
}

// SetDB attaches a database to the Handler. Must be called before requests
// arrive if you want PATCH /api/decks/{owner}/{id} to persist changes.
func (h *Handler) SetDB(db *sql.DB) { h.db = db }

func (h *Handler) loadCustomName(ctx context.Context, owner, id string) string {
	if h.db == nil {
		return ""
	}
	var name string
	err := h.db.QueryRowContext(ctx,
		`SELECT custom_name FROM deck_meta WHERE owner = ? AND id = ?`,
		owner, id).Scan(&name)
	if err != nil {
		return ""
	}
	return name
}

func (h *Handler) saveCustomName(ctx context.Context, owner, id, name string) error {
	if h.db == nil {
		return errors.New("hexapi: deck_meta storage not initialized")
	}
	_, err := h.db.ExecContext(ctx,
		`INSERT INTO deck_meta (owner, id, custom_name, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(owner, id) DO UPDATE SET
		   custom_name = excluded.custom_name,
		   updated_at  = excluded.updated_at`,
		owner, id, name, time.Now().Unix())
	return err
}

// loadClonedFrom returns the "owner/id" pointer recorded when this deck
// was created by /clone, or "" when the deck was imported directly.
func (h *Handler) loadClonedFrom(ctx context.Context, owner, id string) string {
	if h.db == nil {
		return ""
	}
	var src string
	err := h.db.QueryRowContext(ctx,
		`SELECT cloned_from FROM deck_meta WHERE owner = ? AND id = ?`,
		owner, id).Scan(&src)
	if err != nil {
		return ""
	}
	return src
}

// saveClonedFrom records the source deck a clone was based on. Upserts
// alongside any existing custom_name so the two are stored together.
func (h *Handler) saveClonedFrom(ctx context.Context, owner, id, src string) error {
	if h.db == nil {
		return errors.New("hexapi: deck_meta storage not initialized")
	}
	_, err := h.db.ExecContext(ctx,
		`INSERT INTO deck_meta (owner, id, cloned_from, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(owner, id) DO UPDATE SET
		   cloned_from = excluded.cloned_from,
		   updated_at  = excluded.updated_at`,
		owner, id, src, time.Now().Unix())
	return err
}

// CloneRateLimit is the cap on clone-creates per user per hour. Exported
// so tests can lower it without poking at unexported internals.
const CloneRateLimit = 10

// cloneCountSince returns how many clones the given owner has created
// since the supplied unix timestamp. Used by the clone handler to
// enforce CloneRateLimit; returns 0 (i.e. allow) when the DB is absent.
func (h *Handler) cloneCountSince(ctx context.Context, owner string, since int64) (int, error) {
	if h.db == nil {
		return 0, nil
	}
	var n int
	err := h.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM clone_log WHERE owner = ? AND cloned_at >= ?`,
		owner, since).Scan(&n)
	return n, err
}

// recordClone appends a row to clone_log so the rate limiter can see
// this attempt on the next call. Best-effort: a write error is logged
// elsewhere but does not roll back the clone the caller already wrote
// to disk.
func (h *Handler) recordClone(ctx context.Context, owner, srcKey, dstKey string) error {
	if h.db == nil {
		return nil
	}
	_, err := h.db.ExecContext(ctx,
		`INSERT INTO clone_log (owner, src_key, dst_key, cloned_at)
		 VALUES (?, ?, ?, ?)`,
		owner, srcKey, dstKey, time.Now().Unix())
	return err
}

// handlePatchDeck handles `PATCH /api/decks/{owner}/{id}` for lightweight
// metadata updates. Today: only `name` (the user-set display title). Other
// fields can be added later without changing the route.
//
// Body: {"name": "..."} — empty string clears the custom name and reverts
// the deck back to its derived display.
//
// Returns the updated DeckSummary-shaped JSON so the frontend can swap it
// into local state without a follow-up GET.
func (h *Handler) handlePatchDeck(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	id := r.PathValue("id")
	if !validatePathComponent(owner) || !validatePathComponent(id) {
		http.Error(w, "invalid owner or id", http.StatusBadRequest)
		return
	}
	if !checkOwnership(r, owner) {
		http.Error(w, "forbidden: not deck owner", http.StatusForbidden)
		return
	}

	if findDeckFile(h.DecksDir, owner, id) == "" {
		http.Error(w, "deck not found", http.StatusNotFound)
		return
	}

	var body struct {
		Name *string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.Name == nil {
		http.Error(w, "no patchable fields supplied", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(*body.Name)
	if len(name) > 120 {
		http.Error(w, "name too long (max 120 chars)", http.StatusBadRequest)
		return
	}
	// Disallow control chars and embedded newlines so the name renders
	// cleanly in the UI and as a hero title.
	for _, c := range name {
		if c < 0x20 || c == 0x7f {
			http.Error(w, "name contains control characters", http.StatusBadRequest)
			return
		}
	}

	if err := h.saveCustomName(r.Context(), owner, id, name); err != nil {
		http.Error(w, "save: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build a fresh summary so the client can swap state directly.
	commander, bracket, color := parseDeckFilename(id)
	displayName := commander
	if name != "" {
		displayName = name
	}
	resp := map[string]any{
		"owner":       owner,
		"id":          id,
		"name":        displayName,
		"commander":   commander,
		"custom_name": name,
		"bracket":     bracket,
		"color":       color,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
