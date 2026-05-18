package gameengine

// CR §603.3b batched simultaneous-trigger placement.
//
// When multiple triggered abilities fire from a single game event, the rules
// require all of them to be placed on the stack together, in APNAP order
// across players and in controller-chosen order within each player's bucket.
// The R37 stack/priority audit (docs/stack-priority-audit-r37.md, BUG-37-A)
// flagged that production fire sites used the one-at-a-time PushTriggeredAbility
// path, which opens a priority round and resolves the new trigger inline
// before the next sibling trigger is even placed on the stack. That produces
// depth-first arrival-order resolution instead of §603.3b batched ordering.
//
// The fix here is a re-entrant batch frame:
//
//   - BeginTriggerBatch increments gs.triggerBatchDepth and returns whether
//     this caller opened the outermost frame.
//   - While depth > 0, PushTriggeredAbility appends to gs.pendingTriggers
//     instead of push+resolve.
//   - EndTriggerBatch on the outermost frame runs OrderTriggersAPNAP +
//     PushSimultaneousTriggers, then drains the stack via PriorityRound +
//     ResolveStackTop until the items it pushed are gone.
//
// Inner fire sites called during drain see depth == 0 again and open their
// own batches — so a creature that ETBs from another ETB's resolution gets
// its own §603.3b ordering pass, not absorbed into the parent batch.
//
// The standard idiom at fire sites:
//
//   defer EndTriggerBatch(gs, BeginTriggerBatch(gs))
//
// Single-trigger callers and tests that invoke PushTriggeredAbility directly
// continue to work — outside a batch frame, PushTriggeredAbility takes the
// legacy push+resolve path unchanged.

// BeginTriggerBatch opens a §603.3b batch frame. Returns true if this call
// opened the outermost frame (the caller is responsible for the matching
// EndTriggerBatch with opened=true); returns false if a batch was already
// active (the caller still must call EndTriggerBatch with the same value).
//
// Safe with nil gs (returns false).
func BeginTriggerBatch(gs *GameState) (opened bool) {
	if gs == nil {
		return false
	}
	gs.triggerBatchDepth++
	return gs.triggerBatchDepth == 1
}

// EndTriggerBatch closes a §603.3b batch frame. When opened==true (the
// outermost frame), it orders the collected triggers per APNAP +
// controller-choice, pushes them via PushSimultaneousTriggers, then drains
// the stack through priority/resolve cycles until those items are gone.
//
// Pair with BeginTriggerBatch via defer:
//
//   defer EndTriggerBatch(gs, BeginTriggerBatch(gs))
func EndTriggerBatch(gs *GameState, opened bool) {
	if gs == nil {
		return
	}
	if gs.triggerBatchDepth > 0 {
		gs.triggerBatchDepth--
	}
	if !opened {
		return
	}

	batch := gs.pendingTriggers
	gs.pendingTriggers = nil
	if len(batch) == 0 {
		return
	}
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	if gs.Flags["ended"] == 1 {
		return
	}

	stackBefore := len(gs.Stack)
	PushSimultaneousTriggers(gs, batch)
	if len(gs.Stack) <= stackBefore {
		return
	}

	// Drain: open a priority window per resolution and resolve the top
	// item until the batch we pushed is gone (or the game has ended).
	// Mirrors PushTriggeredAbility's single-trigger drain pattern but
	// loops across the whole batch.
	gs.Flags["resolve_depth"]++
	defer func() { gs.Flags["resolve_depth"]-- }()
	if gs.Flags["resolve_depth"] > maxResolveDepth {
		return
	}

	for len(gs.Stack) > stackBefore {
		if gs.Flags["ended"] == 1 {
			return
		}
		PriorityRound(gs)
		if len(gs.Stack) <= stackBefore {
			return
		}
		ResolveStackTop(gs)
	}
}

// collectTrigger is the engine-internal entry point used by
// PushTriggeredAbility when a batch is open. It appends to the pending
// slice without touching the stack or opening priority.
func collectTrigger(gs *GameState, item *StackItem) {
	if gs == nil || item == nil {
		return
	}
	if item.Kind == "" {
		item.Kind = "triggered"
	}
	gs.pendingTriggers = append(gs.pendingTriggers, item)
	name := ""
	if item.Card != nil {
		name = item.Card.DisplayName()
	}
	GlobalStackTrace.Log("trigger_collect", name, item.Controller, len(gs.pendingTriggers), "603.3b_pending")
}
