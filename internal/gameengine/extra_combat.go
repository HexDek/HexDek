package gameengine

// PendingExtraCombat is one entry in the FIFO queue at
// GameState.PendingExtraCombats. Each entry represents a single extra
// combat phase waiting to be played out; the turn loop pops one per
// iteration, applies its metadata, and runs a combat phase.
//
// Restriction gates which creatures may be declared as attackers during
// this specific combat phase. Empty string = no restriction (vanilla).
// Recognized values:
//
//   ""                       — no restriction (Aggravated Assault, Najeela,
//                              Moraug, Seize the Day, Port Razer)
//   "land_creatures_only"    — only attackers with the Land type may
//                              attack (Bumi, Unleashed)
//
// New restriction tags should be added to passesCombatRestriction() in
// combat.go AND documented here. Keep tag values machine-readable
// (underscores, lowercase) so the gate logic is a simple switch.
//
// OnBegin runs once at the beginning_of_combat step for THIS extra combat
// only — not at every extra combat in the queue. Use it for "at the
// beginning of that combat" rider effects (Moraug's untap-all-creatures,
// Bumi Unleashed's untap-all-lands). Optional; may be nil for vanilla
// extra combats.
//
// SourceCard is the display name of the card that produced this extra
// combat. Used in log lines so spectators can attribute each extra
// combat to the right source. Optional but recommended.
type PendingExtraCombat struct {
	Restriction string
	OnBegin     func(gs *GameState)
	SourceCard  string
}

// AddExtraCombat appends an extra combat phase to the pending queue.
// Use this from per_card handlers and from resolveExtraCombat instead
// of mutating gs.PendingExtraCombats directly — the helper keeps the
// invariant that every entry carries explicit metadata (even if the
// metadata is the zero value, signaling a vanilla extra combat).
//
// Calling this from inside trigger resolution preserves APNAP order:
// the active player's chosen trigger order (resolved off the stack)
// becomes the order entries land in the queue, which is the order
// extra combats are played out.
func (gs *GameState) AddExtraCombat(c PendingExtraCombat) {
	if gs == nil {
		return
	}
	gs.PendingExtraCombats = append(gs.PendingExtraCombats, c)
}
