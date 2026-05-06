package main

// conditional_setup.go — registry-driven priming for triggered-ability
// preconditions. The #1 source of Goldilocks gaps (~4.7K draw/life/damage
// failures) was Thor placing the card on the battlefield but never
// satisfying the trigger's CONDITION clause: a "whenever a creature you
// control dies, draw a card" card has nothing to die, so the trigger
// never fires and the test reports a dead effect.
//
// This file maps each canonical trigger event to a small priming action
// that builds the world so the trigger CAN fire when fireTriggerEvent
// runs after the snapshot. Priming runs BEFORE the snapshot so the
// before/after diff captures only the trigger's effect, not the priming
// entities.

import (
	"strings"

	"github.com/hexdek/hexdek/internal/gameast"
	"github.com/hexdek/hexdek/internal/gameengine"
)

// conditionAction primes the game state so a trigger's condition can be
// satisfied when fireTriggerEvent runs. Each action is idempotent: if
// the world already has what's needed (e.g. an opponent creature), the
// action is a no-op.
type conditionAction struct {
	// kind is the canonical condition slug used in trace entries.
	kind string
	// describe returns a one-line summary suitable for trace output.
	describe func(t *gameast.Trigger) string
	// apply mutates gs to prime the precondition. srcPerm is the card
	// under test; it may be nil if Goldilocks couldn't find the source
	// permanent (we still try a best-effort prime).
	apply func(gs *gameengine.GameState, srcPerm *gameengine.Permanent)
}

// classifyTrigger maps a Trigger AST node to a registry slug.
//
// We can't dispatch on Event alone — the AST also encodes the actor
// filter ("a creature you control" vs "a creature an opponent controls")
// and the phase ("upkeep" vs "end_step"), and the firing logic differs
// per slot.
func classifyTrigger(t *gameast.Trigger) string {
	if t == nil {
		return ""
	}
	event := strings.ToLower(strings.TrimSpace(t.Event))
	phase := strings.ToLower(strings.TrimSpace(t.Phase))

	actorBase := ""
	if t.Actor != nil {
		actorBase = strings.ToLower(t.Actor.Base)
	}
	opponentActor := strings.Contains(actorBase, "opponent") ||
		strings.Contains(actorBase, "an opponent")

	switch {
	case event == "dies" || strings.Contains(event, "dies") ||
		strings.Contains(event, "is put into a graveyard"):
		return "creature_dies"
	case event == "etb" || strings.Contains(event, "enters"):
		return "creature_etb"
	case event == "attacks" || strings.Contains(event, "attack"):
		return "attacks"
	case event == "deal_combat_damage" || event == "deals_combat_damage" ||
		strings.Contains(event, "combat damage"):
		return "combat_damage"
	case strings.Contains(event, "cast") || strings.Contains(event, "spell"):
		if opponentActor {
			return "opponent_cast"
		}
		return "cast_spell"
	case strings.Contains(event, "gain") && strings.Contains(event, "life"):
		return "gain_life"
	case strings.Contains(event, "lose") && strings.Contains(event, "life"):
		return "lose_life"
	case strings.Contains(event, "draw"):
		return "draw_card"
	case strings.Contains(event, "discard"):
		if opponentActor {
			return "opponent_discards"
		}
		return "discard"
	case strings.Contains(event, "leaves") || strings.Contains(event, "ltb"):
		return "ltb"
	case strings.Contains(event, "sacrific"):
		return "sacrifice"
	case event == "phase" || phase != "":
		switch phase {
		case "upkeep":
			return "upkeep"
		case "end_step", "end of turn", "endstep":
			return "end_step"
		case "combat_start", "begin_combat", "beginning of combat":
			return "begin_combat"
		case "untap":
			return "untap_step"
		case "draw_step":
			return "draw_step"
		}
		return "phase_" + phase
	}
	return ""
}

// triggerConditionActions is the dispatch table. The keys come from
// classifyTrigger.
var triggerConditionActions = map[string]conditionAction{
	"creature_dies": {
		kind: "creature_dies",
		describe: func(t *gameast.Trigger) string {
			return "place 'Setup Victim' 1/1 token on seat 0 (will die in fireTriggerEvent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			// Always add a fresh victim, distinct from srcPerm and any
			// existing prime, so 'whenever ANOTHER creature dies' triggers
			// also fire (a friendly creature already on the board may be
			// the source itself).
			placeNamedFriendlyCreature(gs, "Setup Victim")
		},
	},

	"creature_etb": {
		kind: "creature_etb",
		describe: func(t *gameast.Trigger) string {
			actor := ""
			if t.Actor != nil {
				actor = t.Actor.Base
			}
			if strings.Contains(strings.ToLower(actor), "another") {
				return "place 'ETB Buddy' 1/1 token (a second creature for 'another creature ETBs')"
			}
			return "no-op (source's own ETB will fire trigger)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			// Source's own ETB is fired by fireTriggerEvent. For
			// "another" variants we need a second creature already
			// present so the source's ETB can trigger off it (or vice
			// versa — fireTriggerEvent invokes the source's ETB hook).
			placeNamedFriendlyCreature(gs, "ETB Buddy")
		},
	},

	"attacks": {
		kind: "attacks",
		describe: func(t *gameast.Trigger) string {
			return "untap source so it can attack; ensure opponent has no blockers"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if srcPerm != nil {
				srcPerm.Tapped = false
				if srcPerm.Flags == nil {
					srcPerm.Flags = map[string]int{}
				}
				// Clear summoning sickness so the source is attack-eligible.
				delete(srcPerm.Flags, "summoning_sick")
			}
		},
	},

	"combat_damage": {
		kind: "combat_damage",
		describe: func(t *gameast.Trigger) string {
			return "ensure opponent creature exists as combat-damage target"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if !hasOpponentCreature(gs) {
				placeTargetCreatureOnOpponent(gs)
			}
		},
	},

	"cast_spell": {
		kind: "cast_spell",
		describe: func(t *gameast.Trigger) string {
			return "ensure 5 castable cards in seat 0 library"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if len(gs.Seats[0].Library) < 5 {
				fillLibrary(gs, 0, 5)
			}
		},
	},

	"opponent_cast": {
		kind: "opponent_cast",
		describe: func(t *gameast.Trigger) string {
			return "ensure 5 castable cards in seat 1 library (opponent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if len(gs.Seats) > 1 && len(gs.Seats[1].Library) < 5 {
				fillLibrary(gs, 1, 5)
			}
		},
	},

	"gain_life": {
		kind: "gain_life",
		describe: func(t *gameast.Trigger) string {
			return "no priming (life gain is fired in fireTriggerEvent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {},
	},

	"lose_life": {
		kind: "lose_life",
		describe: func(t *gameast.Trigger) string {
			return "raise seat 0 life to 30 so a fired loss is observable"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if gs.Seats[0].Life < 30 {
				gs.Seats[0].Life = 30
			}
		},
	},

	"draw_card": {
		kind: "draw_card",
		describe: func(t *gameast.Trigger) string {
			return "ensure 5 cards in seat 0 library"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if len(gs.Seats[0].Library) < 5 {
				fillLibrary(gs, 0, 5)
			}
		},
	},

	"discard": {
		kind: "discard",
		describe: func(t *gameast.Trigger) string {
			return "ensure 3 cards in seat 0 hand"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if len(gs.Seats[0].Hand) < 3 {
				fillHand(gs, 0, 3-len(gs.Seats[0].Hand))
			}
		},
	},

	"opponent_discards": {
		kind: "opponent_discards",
		describe: func(t *gameast.Trigger) string {
			return "ensure 3 cards in seat 1 hand (opponent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if len(gs.Seats) > 1 && len(gs.Seats[1].Hand) < 3 {
				fillHand(gs, 1, 3-len(gs.Seats[1].Hand))
			}
		},
	},

	"ltb": {
		kind: "ltb",
		describe: func(t *gameast.Trigger) string {
			return "no priming (source bounced in fireTriggerEvent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {},
	},

	"sacrifice": {
		kind: "sacrifice",
		describe: func(t *gameast.Trigger) string {
			return "place 'Sac Fodder' 1/1 token as a sacrifice candidate"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			placeNamedFriendlyCreature(gs, "Sac Fodder")
		},
	},

	"upkeep": {
		kind: "upkeep",
		describe: func(t *gameast.Trigger) string {
			return "no priming (phase advance happens in fireTriggerEvent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {},
	},

	"end_step": {
		kind: "end_step",
		describe: func(t *gameast.Trigger) string {
			return "no priming (phase advance happens in fireTriggerEvent)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {},
	},

	"begin_combat": {
		kind: "begin_combat",
		describe: func(t *gameast.Trigger) string {
			return "no priming (phase advance handled at fire time)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {},
	},

	"untap_step": {
		kind: "untap_step",
		describe: func(t *gameast.Trigger) string {
			return "no priming (phase advance handled at fire time)"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {},
	},

	"draw_step": {
		kind: "draw_step",
		describe: func(t *gameast.Trigger) string {
			return "ensure 5 cards in seat 0 library so the draw step succeeds"
		},
		apply: func(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
			if len(gs.Seats[0].Library) < 5 {
				fillLibrary(gs, 0, 5)
			}
		},
	},
}

// primeTriggerCondition runs the registry action that satisfies the
// card's trigger precondition, recording CONDITION_SETUP and TRIGGER_FIRE
// trace entries. Returns true if a registry entry matched, false on
// fallback (no action taken — Goldilocks falls back to existing
// behaviour, i.e. fireTriggerEvent does its own best-effort firing).
func primeTriggerCondition(gs *gameengine.GameState, info *effectInfo, srcPerm *gameengine.Permanent, tr *Tracer) bool {
	if info == nil || info.trigger == nil || gs == nil {
		return false
	}
	slug := classifyTrigger(info.trigger)
	if slug == "" {
		tr.Record("CONDITION_SETUP", "event=%q (unrecognised) → fallback to fireTriggerEvent default",
			info.trigger.Event)
		return false
	}
	action, ok := triggerConditionActions[slug]
	if !ok {
		// Phase events that didn't match a specific phase fall through to
		// here. We still classify so the trace is informative.
		tr.Record("CONDITION_SETUP", "slug=%s (no registry action) → fallback", slug)
		return false
	}

	tr.Record("CONDITION_SETUP", "%q → %s", info.trigger.Event, action.describe(info.trigger))
	action.apply(gs, srcPerm)

	cardName := "<unknown>"
	if srcPerm != nil && srcPerm.Card != nil {
		cardName = srcPerm.Card.Name
	}
	tr.Record("TRIGGER_FIRE", "%s primed on %q", slug, cardName)
	return true
}

// placeNamedFriendlyCreature is a tiny variant of placeFriendlyCreature
// that uses a caller-supplied name. The default helper hard-codes
// "Friendly Creature", but we want distinct trace-friendly names per
// priming role ("Setup Victim", "ETB Buddy", "Sac Fodder").
func placeNamedFriendlyCreature(gs *gameengine.GameState, name string) *gameengine.Permanent {
	perm := &gameengine.Permanent{
		Card: &gameengine.Card{
			Name:          name,
			Owner:         0,
			Types:         []string{"creature"},
			BasePower:     1,
			BaseToughness: 1,
		},
		Controller: 0,
		Owner:      0,
		Flags:      map[string]int{},
		Counters:   map[string]int{},
	}
	gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, perm)
	return perm
}
