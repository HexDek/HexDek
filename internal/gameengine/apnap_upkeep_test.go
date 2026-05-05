package gameengine

// Odin Golden-Game Oracle — APNAP simultaneous-trigger ordering.
//
// CR §603.3b: when multiple abilities have triggered since the last time
// a player received priority, each player puts theirs on the stack in
// APNAP order — Active Player first, then non-active players in turn
// order. Within each controller's set the controller chooses the order.
//
// Stack consequence (LIFO):
//
//   push order (bottom → top):  AP, NAP1, NAP2, NAP3
//   resolve order (top first):  NAP3, NAP2, NAP1, AP
//
// AP's triggers go on the stack first, so they resolve LAST.
//
// These golden tests stand up a 4-seat game where every seat has a
// permanent with an "at the beginning of upkeep" triggered ability and
// route the simultaneous batch through PushSimultaneousTriggers — the
// engine's CR §603.3b entry point. Assertions cover three things:
//
//   1. Stack push order is APNAP starting from gs.Active.
//   2. Resolution order via stack_resolve events is reverse-APNAP.
//   3. Multiple triggers controlled by the same seat stay grouped, and
//      the active seat rotates correctly when gs.Active is mid-table.

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameast"
)

// upkeepTriggerPerm puts a creature-shaped permanent on a seat's
// battlefield carrying a single "at the beginning of upkeep" triggered
// ability. The trigger has Controller="each" so it would fire on every
// player's upkeep — matters only if a future test routes through
// FirePhaseTriggers; here the trigger is just a tag the collector reads.
func upkeepTriggerPerm(gs *GameState, seat int, name string) *Permanent {
	ast := &gameast.CardAST{
		Name: name,
		Abilities: []gameast.Ability{
			&gameast.Triggered{
				Trigger: gameast.Trigger{
					Event:      "upkeep",
					Phase:      "upkeep",
					Controller: "each",
				},
				// Effect deliberately nil — ResolveStackTop logs the
				// stack_resolve event regardless and skips ResolveEffect
				// when item.Effect is nil. We only care about ordering.
				Effect: nil,
			},
		},
	}
	return addBattlefieldWithAST(gs, seat, name, 1, 1, ast, "creature")
}

// collectUpkeepTriggerItems walks every battlefield in seat-index order
// and builds a StackItem for each upkeep-triggered ability. Mirrors the
// collection half of FirePhaseTriggers but stops short of pushing —
// returning the batch lets the test feed it through the canonical CR
// §603.3b entry point (PushSimultaneousTriggers) without depending on
// FirePhaseTriggers' internal sort. Input order is therefore raw
// seat-index order, which is the worst case for APNAP correctness.
func collectUpkeepTriggerItems(gs *GameState) []*StackItem {
	var out []*StackItem
	for _, seat := range gs.Seats {
		if seat == nil || seat.Lost {
			continue
		}
		for _, perm := range seat.Battlefield {
			if perm == nil || perm.Card == nil || perm.Card.AST == nil {
				continue
			}
			for _, ab := range perm.Card.AST.Abilities {
				trig, ok := ab.(*gameast.Triggered)
				if !ok {
					continue
				}
				if trig.Trigger.Phase != "upkeep" && trig.Trigger.Event != "upkeep" {
					continue
				}
				out = append(out, &StackItem{
					Controller: perm.Controller,
					Source:     perm,
					Card:       perm.Card,
					Effect:     trig.Effect,
					Kind:       "triggered",
				})
			}
		}
	}
	return out
}

// resolveOrderFromEvents walks gs.EventLog and returns the source names
// of every stack_resolve event in the order they fired. This is the
// observable resolution sequence for APNAP assertions.
func resolveOrderFromEvents(gs *GameState) []string {
	var names []string
	for _, ev := range gs.EventLog {
		if ev.Kind == "stack_resolve" {
			names = append(names, ev.Source)
		}
	}
	return names
}

// ---------------------------------------------------------------------------
// Test 1 — Active player at seat 0. Push order = [0,1,2,3];
// resolve order = [3,2,1,0].
// ---------------------------------------------------------------------------

func TestOdin_APNAP_FourSeatUpkeep_ActiveSeat0(t *testing.T) {
	gs := NewGameState(4, nil, nil)
	gs.Active = 0
	gs.Phase = "beginning"
	gs.Step = "upkeep"
	for i := 0; i < 4; i++ {
		gs.Seats[i].Life = 40
	}

	upkeepTriggerPerm(gs, 0, "Trigger-Seat0")
	upkeepTriggerPerm(gs, 1, "Trigger-Seat1")
	upkeepTriggerPerm(gs, 2, "Trigger-Seat2")
	upkeepTriggerPerm(gs, 3, "Trigger-Seat3")

	triggers := collectUpkeepTriggerItems(gs)
	if len(triggers) != 4 {
		t.Fatalf("expected 4 collected upkeep triggers, got %d", len(triggers))
	}

	PushSimultaneousTriggers(gs, triggers)

	// CR §603.3b: APNAP push from active=0 → [0,1,2,3].
	if len(gs.Stack) != 4 {
		t.Fatalf("expected 4 stack items, got %d", len(gs.Stack))
	}
	wantPush := []int{0, 1, 2, 3}
	for i, want := range wantPush {
		if gs.Stack[i].Controller != want {
			t.Fatalf("stack[%d] controller = %d, want %d (APNAP push order)",
				i, gs.Stack[i].Controller, want)
		}
	}

	if !hasEventOfKind(gs, "triggers_ordered") {
		t.Fatal("expected triggers_ordered event from PushSimultaneousTriggers")
	}

	// Resolve top-down. Top of stack resolves first → seat 3 first, AP last.
	for len(gs.Stack) > 0 {
		ResolveStackTop(gs)
	}
	got := resolveOrderFromEvents(gs)
	want := []string{"Trigger-Seat3", "Trigger-Seat2", "Trigger-Seat1", "Trigger-Seat0"}
	if len(got) != len(want) {
		t.Fatalf("expected %d resolve events, got %d (%v)", len(want), len(got), got)
	}
	for i, name := range want {
		if got[i] != name {
			t.Fatalf("resolve[%d] = %q, want %q (full order: %v)", i, got[i], name, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 2 — Active player at seat 2 (mid-game rotation).
// APNAP push = [2,3,0,1]; resolve = [1,0,3,2].
// ---------------------------------------------------------------------------

func TestOdin_APNAP_FourSeatUpkeep_ActiveSeat2(t *testing.T) {
	gs := NewGameState(4, nil, nil)
	gs.Active = 2
	gs.Phase = "beginning"
	gs.Step = "upkeep"
	for i := 0; i < 4; i++ {
		gs.Seats[i].Life = 40
	}

	upkeepTriggerPerm(gs, 0, "Trigger-Seat0")
	upkeepTriggerPerm(gs, 1, "Trigger-Seat1")
	upkeepTriggerPerm(gs, 2, "Trigger-Seat2")
	upkeepTriggerPerm(gs, 3, "Trigger-Seat3")

	triggers := collectUpkeepTriggerItems(gs)
	if len(triggers) != 4 {
		t.Fatalf("expected 4 collected upkeep triggers, got %d", len(triggers))
	}

	PushSimultaneousTriggers(gs, triggers)

	wantPush := []int{2, 3, 0, 1}
	if len(gs.Stack) != 4 {
		t.Fatalf("expected 4 stack items, got %d", len(gs.Stack))
	}
	for i, want := range wantPush {
		if gs.Stack[i].Controller != want {
			t.Fatalf("stack[%d] controller = %d, want %d (APNAP from active=2)",
				i, gs.Stack[i].Controller, want)
		}
	}

	for len(gs.Stack) > 0 {
		ResolveStackTop(gs)
	}
	got := resolveOrderFromEvents(gs)
	want := []string{"Trigger-Seat1", "Trigger-Seat0", "Trigger-Seat3", "Trigger-Seat2"}
	if len(got) != len(want) {
		t.Fatalf("expected %d resolve events, got %d (%v)", len(want), len(got), got)
	}
	for i, name := range want {
		if got[i] != name {
			t.Fatalf("resolve[%d] = %q, want %q (full order: %v)", i, got[i], name, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 3 — Multiple triggers per seat stay grouped under their controller.
// AP=1 with two upkeep permanents on seat 1 plus one each on 0,2,3.
// Push order: [1a,1b, 2, 3, 0]; resolve: [0, 3, 2, 1b, 1a] (default Hat
// keeps insertion order within a group, so 1a stays below 1b on stack).
// ---------------------------------------------------------------------------

func TestOdin_APNAP_FourSeatUpkeep_MultiTriggerPerSeat(t *testing.T) {
	gs := NewGameState(4, nil, nil)
	gs.Active = 1
	gs.Phase = "beginning"
	gs.Step = "upkeep"
	for i := 0; i < 4; i++ {
		gs.Seats[i].Life = 40
	}

	upkeepTriggerPerm(gs, 0, "Trigger-Seat0")
	upkeepTriggerPerm(gs, 1, "Trigger-Seat1a")
	upkeepTriggerPerm(gs, 1, "Trigger-Seat1b")
	upkeepTriggerPerm(gs, 2, "Trigger-Seat2")
	upkeepTriggerPerm(gs, 3, "Trigger-Seat3")

	triggers := collectUpkeepTriggerItems(gs)
	if len(triggers) != 5 {
		t.Fatalf("expected 5 collected upkeep triggers, got %d", len(triggers))
	}

	PushSimultaneousTriggers(gs, triggers)

	// APNAP from active=1: [1, 1, 2, 3, 0]. Both seat-1 triggers come
	// before any non-active seat's trigger.
	if len(gs.Stack) != 5 {
		t.Fatalf("expected 5 stack items, got %d", len(gs.Stack))
	}
	wantCtl := []int{1, 1, 2, 3, 0}
	for i, want := range wantCtl {
		if gs.Stack[i].Controller != want {
			t.Fatalf("stack[%d] controller = %d, want %d", i, gs.Stack[i].Controller, want)
		}
	}

	// Within seat 1's group the default Hat (nil → identity) preserves
	// the input order: 1a before 1b. Push order [1a, 1b], so 1b ends up
	// higher on the stack and resolves before 1a.
	if gs.Stack[0].Card.Name != "Trigger-Seat1a" {
		t.Fatalf("stack[0] (bottom of seat-1 group) = %q, want Trigger-Seat1a",
			gs.Stack[0].Card.Name)
	}
	if gs.Stack[1].Card.Name != "Trigger-Seat1b" {
		t.Fatalf("stack[1] = %q, want Trigger-Seat1b", gs.Stack[1].Card.Name)
	}

	for len(gs.Stack) > 0 {
		ResolveStackTop(gs)
	}
	got := resolveOrderFromEvents(gs)
	want := []string{
		"Trigger-Seat0",
		"Trigger-Seat3",
		"Trigger-Seat2",
		"Trigger-Seat1b",
		"Trigger-Seat1a",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d resolve events, got %d (%v)", len(want), len(got), got)
	}
	for i, name := range want {
		if got[i] != name {
			t.Fatalf("resolve[%d] = %q, want %q (full order: %v)", i, got[i], name, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Test 4 — APNAP rotates correctly through every seat as the active seat
// advances. Each iteration rebuilds a fresh game; the assertion is that
// the bottom of the stack always belongs to the active seat (CR §603.3b
// says AP goes on the stack first).
// ---------------------------------------------------------------------------

func TestOdin_APNAP_FourSeatUpkeep_RotateAcrossAllSeats(t *testing.T) {
	for active := 0; active < 4; active++ {
		gs := NewGameState(4, nil, nil)
		gs.Active = active
		gs.Phase = "beginning"
		gs.Step = "upkeep"
		for i := 0; i < 4; i++ {
			gs.Seats[i].Life = 40
		}
		for i := 0; i < 4; i++ {
			upkeepTriggerPerm(gs, i, namePerSeat(i))
		}

		PushSimultaneousTriggers(gs, collectUpkeepTriggerItems(gs))

		if len(gs.Stack) != 4 {
			t.Fatalf("active=%d: expected 4 stack items, got %d", active, len(gs.Stack))
		}

		// Bottom of stack must be active player's trigger (AP pushed
		// first per APNAP). Top must be the seat immediately before AP
		// in turn order — i.e. (active + 3) mod 4 — since they're last
		// in APNAP and therefore on top of the stack.
		if gs.Stack[0].Controller != active {
			t.Fatalf("active=%d: stack bottom controller = %d, want %d",
				active, gs.Stack[0].Controller, active)
		}
		topWant := (active + 3) % 4
		if gs.Stack[3].Controller != topWant {
			t.Fatalf("active=%d: stack top controller = %d, want %d",
				active, gs.Stack[3].Controller, topWant)
		}

		// Validate the full APNAP sequence: stack[i].Controller == (active+i) mod 4.
		for i := 0; i < 4; i++ {
			want := (active + i) % 4
			if gs.Stack[i].Controller != want {
				t.Fatalf("active=%d: stack[%d] controller = %d, want %d",
					active, i, gs.Stack[i].Controller, want)
			}
		}
	}
}

func namePerSeat(seat int) string {
	switch seat {
	case 0:
		return "Trigger-Seat0"
	case 1:
		return "Trigger-Seat1"
	case 2:
		return "Trigger-Seat2"
	case 3:
		return "Trigger-Seat3"
	}
	return "Trigger-SeatX"
}
