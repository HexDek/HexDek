package gameengine

// keywords_class.go — Class chapter mechanic (CR §716, Adventures
// in the Forgotten Realms 2021).
//
// CR §716.1: Class is a subtype of Enchantment. Each Class card has
//            one or more levels. A class enters the battlefield at
//            level 1 and gains levels via per-level activated
//            abilities. Each level grants the static / triggered
//            abilities printed in its level band.
// CR §716.3: Each level above 1 has an activated ability of the
//            form "{cost}: Level N: gain the level-N abilities."
//            These activations are sorcery-speed.
// CR §716.4: Levels must be gained in order — a class at level 1
//            can only activate the level-2 ability, not skip to
//            level 3.
// CR §716.5: Static abilities printed in a level band are active
//            only while the class's level is within that band.
//
// Engine model
// ------------
// Class state lives on perm.Flags["class_level"]. The §716.1
// "enters at level 1" semantic is enforced by ClassLevel reading
// max(1, stored_value) — a fresh ETB with no flag stamped reads
// level 1 without needing an ETB pre-pass.
//
// The level-band metadata is already extracted from each card's
// AST by parseLevelBracketsFromAST (keywords_levelup.go) — that
// reader walks Static nodes with ModKind=="class_level_band" and
// groups the following keywords / P/T into a LevelBracket. This
// file reuses that reader so the band data structure is shared
// between Level Up creatures and Class enchantments.
//
// Wiring:
//   - HasClass / ClassLevel / MaxClassLevel — type + level
//     readers.
//   - LevelUpClass — sorcery-speed cost-paying advancement step.
//     Caller passes the cost because the AST today doesn't carry
//     per-level cost data for Class cards; the cost is what the
//     printed activated ability says ("{R}: Level 2", etc.).
//     Returns true iff a level was gained.
//   - ClassLevelStaticActive(perm, bandLo) — predicate the layers
//     / per-card handlers consult to decide whether a band's
//     printed ability is currently functional.
//   - ActiveClassBrackets — returns the LevelBracket(s) currently
//     active on `perm` so AI/Hat policy + layers can iterate them
//     directly.
//
// Compatibility:
//   - The pre-existing AdvanceClassLevel + GetClassLevel helpers in
//     keywords_misc.go are kept for legacy callers; both read /
//     write the same perm.Flags["class_level"] backing field as
//     LevelUpClass / ClassLevel so the two surfaces stay in lock-
//     step. The new ClassLevel reader fixes the level-1 default
//     (GetClassLevel returned 0 for fresh-ETB classes); migration
//     to the new reader can land lazily as call sites are touched.

import (
	"strings"
)

// ---------------------------------------------------------------------------
// HasClass
// ---------------------------------------------------------------------------

// HasClass reports whether the card has the Class subtype.
// Lookup is case-insensitive and matches both "class" and
// title-cased "Class" so parser output variations are tolerated.
// Safe on nil card.
func HasClass(card *Card) bool {
	if card == nil {
		return false
	}
	for _, t := range card.Types {
		if strings.EqualFold(strings.TrimSpace(t), "class") {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// ClassLevel / MaxClassLevel
// ---------------------------------------------------------------------------

// ClassLevel returns the current level of `perm`. Per §716.1 a
// fresh Class enters the battlefield at level 1, so the reader
// returns max(1, perm.Flags["class_level"]) — a permanent whose
// flag has never been written still reads 1 without an explicit
// ETB initializer running.
//
// Returns 0 for nil perms or non-class permanents.
func ClassLevel(perm *Permanent) int {
	if perm == nil || !IsClass(perm) {
		return 0
	}
	if perm.Flags == nil {
		return 1
	}
	v := perm.Flags["class_level"]
	if v < 1 {
		return 1
	}
	return v
}

// MaxClassLevel returns the highest level the card can reach. If
// the card's AST carries level-band Modifications (the canonical
// shape produced by parseLevelBracketsFromAST), the max is the
// highest MaxLevel across all bands (or the highest MinLevel for
// open-ended top bands where MaxLevel == -1). Falls back to 3 for
// cards whose AST doesn't expose bands — every printed Class card
// in AFR / SNC / DFT had exactly three levels, so 3 is the safe
// default.
func MaxClassLevel(card *Card) int {
	if card == nil {
		return 0
	}
	if !HasClass(card) {
		return 0
	}
	brackets := parseLevelBracketsFromAST(card)
	max := 0
	for _, b := range brackets {
		hi := b.MaxLevel
		if hi < 0 {
			// Open-ended band ("3+") — the floor IS the cap because
			// no higher-level activation exists.
			hi = b.MinLevel
		}
		if hi > max {
			max = hi
		}
	}
	if max <= 0 {
		return 3 // default for printed Classes
	}
	return max
}

// IsClass reports whether `perm` is a Class enchantment.
// Convenience wrapper around HasClass(perm.Card).
func IsClass(perm *Permanent) bool {
	if perm == nil {
		return false
	}
	return HasClass(perm.Card)
}

// ---------------------------------------------------------------------------
// LevelUpClass
// ---------------------------------------------------------------------------

// LevelUpClass activates the next-level activated ability on a
// Class permanent. CR §716.3.
//
// Preconditions enforced here:
//   - perm is a Class enchantment
//   - sorcery timing (controller is active, in a main phase, stack
//     empty — via isSorceryTiming, which Level Up creatures also
//     use, keywords_levelup.go)
//   - current level < MaxClassLevel(card)
//   - controller can afford `levelUpCost`
//
// On success: perm.Flags["class_level"] is incremented by 1
// (level 1 → 2, 2 → 3, …), mana is paid, the per-card
// "class_level_up" trigger fires with ctx{source, new_level,
// previous_level}, and a class_level_up log event is emitted.
//
// Returns true iff a level was gained.
//
// The cost argument is explicit because the AST in this engine
// doesn't currently carry per-level cost data for Class cards;
// callers (UI / AI policy / per-card handlers) read the printed
// "{cost}: Level N: …" text and pass the parsed cost in. The
// printed-cost decoder can land in a follow-up PR without
// changing this entry point.
func LevelUpClass(gs *GameState, perm *Permanent, levelUpCost int) bool {
	if gs == nil || perm == nil {
		return false
	}
	if !IsClass(perm) {
		return false
	}
	seatIdx := perm.Controller
	if seatIdx < 0 || seatIdx >= len(gs.Seats) {
		return false
	}
	seat := gs.Seats[seatIdx]
	if seat == nil {
		return false
	}
	// CR §716.3 — Level-up activations are sorcery-speed.
	if !isSorceryTiming(gs, seatIdx) {
		gs.LogEvent(Event{
			Kind:   "class_level_up_rejected",
			Seat:   seatIdx,
			Source: sourceNameOf(perm),
			Details: map[string]interface{}{
				"reason": "sorcery_speed_only",
				"rule":   "716.3",
			},
		})
		return false
	}
	if levelUpCost < 0 {
		return false
	}
	current := ClassLevel(perm)
	max := MaxClassLevel(perm.Card)
	if current >= max {
		gs.LogEvent(Event{
			Kind:   "class_level_up_rejected",
			Seat:   seatIdx,
			Source: sourceNameOf(perm),
			Details: map[string]interface{}{
				"reason":  "already_at_max_level",
				"current": current,
				"max":     max,
				"rule":    "716.4",
			},
		})
		return false
	}
	if seat.ManaPool < levelUpCost {
		gs.LogEvent(Event{
			Kind:   "class_level_up_rejected",
			Seat:   seatIdx,
			Source: sourceNameOf(perm),
			Details: map[string]interface{}{
				"reason": "insufficient_mana",
				"cost":   levelUpCost,
			},
		})
		return false
	}
	seat.ManaPool -= levelUpCost
	SyncManaAfterSpend(seat)
	if levelUpCost > 0 {
		gs.LogEvent(Event{
			Kind:   "pay_mana",
			Seat:   seatIdx,
			Amount: levelUpCost,
			Source: sourceNameOf(perm),
			Details: map[string]interface{}{
				"reason":  "class_level_up",
				"keyword": "class",
				"rule":    "601.2f",
			},
		})
	}
	if perm.Flags == nil {
		perm.Flags = map[string]int{}
	}
	perm.Flags["class_level"] = current + 1
	gs.InvalidateCharacteristicsCache()

	gs.LogEvent(Event{
		Kind:   "class_level_up",
		Seat:   seatIdx,
		Source: sourceNameOf(perm),
		Amount: current + 1,
		Details: map[string]interface{}{
			"previous_level": current,
			"new_level":      current + 1,
			"cost":           levelUpCost,
			"rule":           "716.3",
		},
	})
	FireCardTrigger(gs, "class_level_up", map[string]interface{}{
		"source":          perm,
		"controller_seat": perm.Controller,
		"previous_level":  current,
		"new_level":       current + 1,
	})
	return true
}

// ---------------------------------------------------------------------------
// ClassLevelStaticActive / ActiveClassBrackets
// ---------------------------------------------------------------------------

// ClassLevelStaticActive reports whether a band starting at
// `bandLo` is currently functional on `perm` per §716.5. A band is
// active when the perm's current ClassLevel is at or above the
// band's MinLevel. Layers and per-card handlers that grant
// keywords / P/T from a specific band gate on this predicate.
//
// Returns false for nil perms, non-class perms, and bands above
// the perm's current level.
func ClassLevelStaticActive(perm *Permanent, bandLo int) bool {
	if perm == nil {
		return false
	}
	return ClassLevel(perm) >= bandLo
}

// ActiveClassBrackets returns the level brackets currently active
// on `perm` — every bracket whose MinLevel <= ClassLevel and
// (MaxLevel < 0 OR MaxLevel >= ClassLevel). Useful for layer
// computation + AI/Hat policy that wants to enumerate the
// currently-effective payoffs.
func ActiveClassBrackets(perm *Permanent) []LevelBracket {
	if perm == nil || perm.Card == nil || !IsClass(perm) {
		return nil
	}
	all := parseLevelBracketsFromAST(perm.Card)
	level := ClassLevel(perm)
	var active []LevelBracket
	for _, b := range all {
		if b.MinLevel > level {
			continue
		}
		if b.MaxLevel >= 0 && b.MaxLevel < level {
			continue
		}
		active = append(active, b)
	}
	return active
}

// sourceNameOf is a small log-helper that returns the perm's
// display name (or "<unknown>" when the card is unset).
func sourceNameOf(perm *Permanent) string {
	if perm == nil || perm.Card == nil {
		return "<unknown>"
	}
	return perm.Card.DisplayName()
}
