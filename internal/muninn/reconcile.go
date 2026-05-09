package muninn

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ArchiveResult summarizes what an ArchiveFixedCards run touched.
type ArchiveResult struct {
	DeadTriggersBefore  int      `json:"dead_triggers_before"`
	DeadTriggersAfter   int      `json:"dead_triggers_after"`
	DeadTriggersArchived int     `json:"dead_triggers_archived"`
	UnmatchedCards      []string `json:"unmatched_cards"`
	Timestamp           string   `json:"timestamp"`
}

// ArchivedDeadTrigger wraps a DeadTrigger with the reason it was archived
// so the archive file is self-describing.
type ArchivedDeadTrigger struct {
	DeadTrigger
	ArchivedAt   string `json:"archived_at"`
	ArchiveCause string `json:"archive_cause"`
}

const (
	deadTriggersArchiveFile = "dead_triggers_archive.json"
)

// ArchiveFixedCards moves entries out of dead_triggers.json whose CardName
// matches any name in fixedCards, recording them in dead_triggers_archive.json
// alongside a cause string. Names match case-insensitively after trimming
// whitespace; double-faced "A // B" names match either face.
//
// The intent: after a wave of handler upgrades, dead-trigger records keyed
// to the now-fixed cards no longer reflect real engine behavior. Archiving
// them keeps the live dead_triggers.json focused on genuinely-broken
// triggers without losing history.
//
// cause is stored on every archived row (e.g. "era5-unification PR #35").
// If cause is empty, "manual reconciliation" is recorded.
func ArchiveFixedCards(dir string, fixedCards []string, cause string) (ArchiveResult, error) {
	res := ArchiveResult{Timestamp: time.Now().UTC().Format(time.RFC3339)}
	if cause == "" {
		cause = "manual reconciliation"
	}
	if len(fixedCards) == 0 {
		return res, nil
	}

	fileMu.Lock()
	defer fileMu.Unlock()

	existing, err := ReadDeadTriggers(dir)
	if err != nil {
		return res, err
	}
	res.DeadTriggersBefore = len(existing)

	matchSet := make(map[string]bool, len(fixedCards)*2)
	for _, name := range fixedCards {
		for _, key := range nameKeys(name) {
			matchSet[key] = true
		}
	}

	matchedCards := make(map[string]bool)
	var kept []DeadTrigger
	var archivedNew []ArchivedDeadTrigger

	for _, dt := range existing {
		if isFixedCard(dt.CardName, matchSet) {
			archivedNew = append(archivedNew, ArchivedDeadTrigger{
				DeadTrigger:  dt,
				ArchivedAt:   res.Timestamp,
				ArchiveCause: cause,
			})
			for _, key := range nameKeys(dt.CardName) {
				matchedCards[key] = true
			}
		} else {
			kept = append(kept, dt)
		}
	}

	if len(archivedNew) == 0 {
		// Still report unmatched so the caller can see which fixed cards
		// had no live records.
		res.UnmatchedCards = unmatchedCards(fixedCards, matchedCards)
		return res, nil
	}

	res.DeadTriggersAfter = len(kept)
	res.DeadTriggersArchived = len(archivedNew)
	res.UnmatchedCards = unmatchedCards(fixedCards, matchedCards)

	prior, err := ReadDeadTriggersArchive(dir)
	if err != nil {
		return res, err
	}
	merged := append(prior, archivedNew...)

	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersFile), kept); err != nil {
		return res, err
	}
	if err := atomicWriteJSON(filepath.Join(dir, deadTriggersArchiveFile), merged); err != nil {
		return res, fmt.Errorf("muninn: write archive: %w", err)
	}
	return res, nil
}

// ReadDeadTriggersArchive returns the archived dead-trigger ledger.
func ReadDeadTriggersArchive(dir string) ([]ArchivedDeadTrigger, error) {
	var out []ArchivedDeadTrigger
	if err := readJSON(filepath.Join(dir, deadTriggersArchiveFile), &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []ArchivedDeadTrigger{}
	}
	return out, nil
}

// nameKeys returns lowercased, trimmed match keys for a card name. For a
// DFC "A // B" name, both faces and the joined form match.
func nameKeys(name string) []string {
	base := strings.ToLower(strings.TrimSpace(name))
	if base == "" {
		return nil
	}
	keys := []string{base}
	if strings.Contains(base, " // ") {
		for _, half := range strings.Split(base, " // ") {
			half = strings.TrimSpace(half)
			if half != "" && half != base {
				keys = append(keys, half)
			}
		}
	}
	return keys
}

func isFixedCard(name string, matchSet map[string]bool) bool {
	for _, k := range nameKeys(name) {
		if matchSet[k] {
			return true
		}
	}
	return false
}

func unmatchedCards(fixed []string, matched map[string]bool) []string {
	var out []string
	seen := make(map[string]bool)
	for _, name := range fixed {
		key := strings.ToLower(strings.TrimSpace(name))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		hit := false
		for _, k := range nameKeys(name) {
			if matched[k] {
				hit = true
				break
			}
		}
		if !hit {
			out = append(out, name)
		}
	}
	return out
}

// EraPassFixedCards is the canonical list of commander cards whose
// per_card handlers were upgraded by the 2026-05-09 era unification PRs
// (#30, #31, #32, #33, #34, #35) and the dead-stub cleanup (#36).
//
// Derived directly from the registrations in
// internal/gameengine/per_card/custom_*.go at the time of this commit.
// When a future era pass lands, append to (or replace) this list and
// re-run hexdek-muninn --reconcile-fixed.
var EraPassFixedCards = []string{
	"Amalia Benavides Aguirre",
	"Araumi of the Dead Tide",
	"Arcades, the Strategist",
	"Asmoranomardicadaistinaculdacar",
	"Chainer, Dementia Master",
	"Charix, the Raging Isle",
	"Choco, Seeker of Paradise",
	"Derevi, Empyrial Tactician",
	"Drivnod, Carnage Dominus",
	"Eddie Brock // Venom, Lethal Protector",
	"Eriette of the Charmed Apple",
	"Felothar the Steadfast",
	"Galazeth Prismari",
	"Ghyrson Starn, Kelermorph",
	"Giada, Font of Hope",
	"Inalla, Archmage Ritualist",
	"Isshin, Two Heavens as One",
	"Ixhel, Scion of Atraxa",
	"Jadzi, Oracle of Arcavios",
	"Kalamax, the Stormsire",
	"Karador, Ghost Chieftain",
	"Kardur, Doomscourge",
	"Lier, Disciple of the Drowned",
	"Mairsil, the Pretender",
	"Marchesa, the Black Rose",
	"Mayael the Anima",
	"Mondrak, Glory Dominus",
	"Quandrix, the Proof",
	"Ruric Thar, the Unbowed",
	"Saheeli, Radiant Creator",
	"Sakashima of a Thousand Faces",
	"Selenia, Dark Angel",
	"Shadow the Hedgehog",
	"Silverquill, the Disputant",
	"Sliver Gravemother",
	"Solphim, Mayhem Dominus",
	"Tiamat",
	"Tifa Lockhart",
	"Toxrill, the Corrosive",
	"Veyran, Voice of Duality",
	"Yasharn, Implacable Earth",
	"Yenna, Redtooth Regent",
	"Yurlok of Scorch Thrash",
	"Zopandrel, Hunger Dominus",
}
