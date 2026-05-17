package gameengine

// keywords_daynight.go — Daybound (CR §702.149) + Nightbound (§702.150)
// keyword surface, plus the ETB hook that drives CR §726.2.
//
// The day/night state machine itself lives in dfc.go:
//
//   - SetDayNight(gs, new, reason, rule)        — shared transition primitive
//   - MaybeBecomeDay(gs, reason)                — §726.2 "first daybound ETB"
//   - EvaluateDayNightAtTurnStart(gs)           — §726.3a turn-start transition
//   - ApplyDayboundNightboundTransforms(gs)     — §702.149/150 face sweep
//   - HasDayboundOrNightboundPermanent(gs)      — predicate for §726.2
//
// This file is the keyword-mechanic-facing surface: card-level keyword
// detection (HasDaybound / HasNightbound), gameplay-facing day-state
// queries (IsDay / IsNight / IsNeitherDayNorNight), and the ETB hook
// (OnDayboundOrNightboundETB) that production ETB code paths invoke so
// §726.2 actually fires at runtime.
//
// Comp-rules citations (rule numbers as of MidnightHunt-era printings):
//
//   §702.149 Daybound — A creature with daybound transforms when the
//                       game's day/night state changes to night.
//   §702.150 Nightbound — A creature with nightbound transforms when
//                         the game's day/night state changes to day.
//                         Cards with nightbound enter the battlefield
//                         with their back face up if the game has no
//                         day/night state when they enter.
//   §726.1   The game tracks a day/night designation that is "neither"
//            by default and toggles between day and night.
//   §726.2   If a permanent with daybound or nightbound enters the
//            battlefield while the game has no day/night designation,
//            the game becomes day.
//   §726.3a  At the beginning of each turn, if the game is day and the
//            previous active player cast no spells during their last
//            turn, it becomes night. If the game is night and the
//            previous active player cast two or more spells last turn,
//            it becomes day.

// ---------------------------------------------------------------------------
// HasDaybound / HasNightbound — keyword detection
// ---------------------------------------------------------------------------

// HasDaybound returns true if the card's currently-active face has the
// daybound keyword. Front-face daybound creatures (the human halves of
// Innistrad werewolves) carry the keyword on their printed AST.
func HasDaybound(card *Card) bool {
	return cardHasKeywordByName(card, "daybound")
}

// HasNightbound returns true if the card's currently-active face has
// the nightbound keyword. Back-face nightbound creatures carry the
// keyword on the back-face AST.
func HasNightbound(card *Card) bool {
	return cardHasKeywordByName(card, "nightbound")
}

// PermHasDaybound checks the permanent's currently-active face (the
// face perm.Card.AST points at). Use this on the battlefield rather
// than HasDaybound when you want to know whether the daybound static
// ability is currently in force.
func PermHasDaybound(perm *Permanent) bool {
	return permanentActiveFaceHasKeyword(perm, "daybound")
}

// PermHasNightbound checks the permanent's currently-active face. Use
// this on the battlefield rather than HasNightbound when you want to
// know whether the nightbound static ability is currently in force.
func PermHasNightbound(perm *Permanent) bool {
	return permanentActiveFaceHasKeyword(perm, "nightbound")
}

// ---------------------------------------------------------------------------
// Day/Night state queries
// ---------------------------------------------------------------------------

// IsDay reports whether the game's current day/night designation is
// "day". CR §726.1.
func IsDay(gs *GameState) bool {
	return gs != nil && gs.DayNight == DayNightDay
}

// IsNight reports whether the game's current day/night designation is
// "night". CR §726.1.
func IsNight(gs *GameState) bool {
	return gs != nil && gs.DayNight == DayNightNight
}

// IsNeitherDayNorNight reports whether the game has no day/night
// designation. CR §726.1 — this is the default state at game start and
// before any daybound/nightbound permanent has ETB'd. Accepts both the
// canonical DayNightNeither constant ("neither") and the empty-string
// zero value so callers that construct a GameState without going
// through NewGameState still get correct semantics.
func IsNeitherDayNorNight(gs *GameState) bool {
	if gs == nil {
		return true
	}
	return gs.DayNight == "" || gs.DayNight == DayNightNeither
}

// ---------------------------------------------------------------------------
// ETB hook — §726.2
// ---------------------------------------------------------------------------

// OnDayboundOrNightboundETB is the canonical ETB hook for the daybound/
// nightbound state-becomes-day check. Wired into FirePermanentETBTriggers
// so any permanent entering the battlefield is checked once. CR §726.2.
//
// Cheap fast-path: if the game already has a day/night designation, the
// underlying MaybeBecomeDay returns immediately. If the entering
// permanent doesn't carry the keyword we also early-out (so a board
// full of vanilla permanents pays no scan cost).
func OnDayboundOrNightboundETB(gs *GameState, perm *Permanent) {
	if gs == nil || perm == nil || perm.Card == nil {
		return
	}
	// Fast path: state is already set; no transition possible from ETB.
	if !IsNeitherDayNorNight(gs) {
		return
	}
	if !PermHasDaybound(perm) && !PermHasNightbound(perm) {
		// Either face may carry the keyword (some werewolves have
		// nightbound printed on the back face only). Check both.
		if !astHasKeyword(perm.FrontFaceAST, "daybound") &&
			!astHasKeyword(perm.FrontFaceAST, "nightbound") &&
			!astHasKeyword(perm.BackFaceAST, "daybound") &&
			!astHasKeyword(perm.BackFaceAST, "nightbound") {
			return
		}
	}
	MaybeBecomeDay(gs, "daybound_or_nightbound_etb")
}
