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
	"fmt"
	"regexp"
	"strconv"
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

	// Layer the intervening-if priming on top of the event-based action.
	// Most "if X this turn / if you have the city's blessing / if you
	// control another <subtype>" conditions live in the conditional_effect
	// args text, not in trigger.Condition (which the parser leaves null
	// for these cards). The two priming layers are orthogonal.
	primeInterveningIf(gs, info, srcPerm, tr)
	return true
}

// ---------------------------------------------------------------------------
// Intervening-if priming.
//
// For triggered abilities of shape "AT/WHENEVER trigger, IF condition,
// EFFECT" the parser emits the trigger as Trigger{Event=phase/etb/...} and
// the rest as a single Modification(kind="conditional_effect",
// args=["if X, Y"]) at the effect slot. trigger.Condition and
// intervening_if are both null. Without priming, evalCondition's "raw"
// fallback defaults to TRUE so the effect fires anyway, but the resulting
// trace is misleading and any per-card listener that reads concrete state
// (Lathiel's lathiel_life_gained_this_turn flag, Sorin's seat
// life_gained_this_turn counter, Twilight Prophet's citys_blessing) will
// silently fail.
//
// primeInterveningIf walks info.condition and the conditional_effect text
// and applies the appropriate setup so both the engine's evalCondition
// path AND any per-card flag-reading handler observe the condition as
// satisfied.
// ---------------------------------------------------------------------------

var (
	reLifeMoreStarting = regexp.MustCompile(`at least (\d+) life more than your starting life total`)
	reControlAnother   = regexp.MustCompile(`you control another ([a-z][a-z\- ]*?)(?:[,.]| and| or|$)`)
	reControlSubtype   = regexp.MustCompile(`you control a ([a-z][a-z\- ]*?)(?:[,.]| and| or|$)`)
)

// extractInterveningText returns the lower-cased "if ..." clause text from
// either info.condition (typed Condition AST) or from a conditional_effect
// ModificationEffect's args[0].
func extractInterveningText(info *effectInfo) string {
	if info == nil {
		return ""
	}
	if info.condition != nil {
		// A typed Condition. Stringify Kind+Args so downstream
		// substring matchers pick up "raw" args[0] verbatim AND
		// well-known kinds via their Kind tag.
		parts := []string{strings.ToLower(info.condition.Kind)}
		for _, a := range info.condition.Args {
			parts = append(parts, strings.ToLower(fmt.Sprintf("%v", a)))
		}
		return strings.Join(parts, " ")
	}
	if t := conditionalEffectText(info.effect); t != "" {
		return t
	}
	return conditionalEffectText(info.fullEffect)
}

// conditionalEffectText extracts args[0] from a *gameast.ModificationEffect
// whose ModKind is "conditional_effect", recursively descending Sequence
// and Optional wrappers.
func conditionalEffectText(eff gameast.Effect) string {
	if eff == nil {
		return ""
	}
	switch e := eff.(type) {
	case *gameast.ModificationEffect:
		if e.ModKind == "conditional_effect" && len(e.Args) > 0 {
			if s, ok := e.Args[0].(string); ok {
				return strings.ToLower(s)
			}
		}
	case *gameast.Sequence:
		for _, item := range e.Items {
			if t := conditionalEffectText(item); t != "" {
				return t
			}
		}
	case *gameast.Optional_:
		return conditionalEffectText(e.Body)
	}
	return ""
}

// primeInterveningIf applies state mutations that satisfy embedded "if X"
// clauses found in info. Returns true when at least one prime was applied.
func primeInterveningIf(gs *gameengine.GameState, info *effectInfo, srcPerm *gameengine.Permanent, tr *Tracer) bool {
	text := extractInterveningText(info)
	if text == "" {
		return false
	}

	primed := false

	// Life gain/loss this turn. Order matters: the combined "gained or
	// lost" form must be checked before the individual matchers so we
	// don't double-apply.
	switch {
	case strings.Contains(text, "gained or lost life this turn"):
		primeGainedLife(gs, 3)
		primeLostLife(gs, 1)
		tr.Record("CONDITION_SETUP",
			"%q → GainLife(seat0, 3) + LoseLife(seat0, 1)",
			"gained or lost life this turn")
		primed = true
	case strings.Contains(text, "gained life this turn"):
		primeGainedLife(gs, 3)
		tr.Record("CONDITION_SETUP",
			"%q → GainLife(seat0, 3)",
			"gained life this turn")
		primed = true
	case strings.Contains(text, "lost life this turn"):
		primeLostLife(gs, 2)
		tr.Record("CONDITION_SETUP",
			"%q → LoseLife(seat0, 2)",
			"lost life this turn")
		primed = true
	}

	// Counter placement this turn. Lasting Tarfire, Lord Jyscal Guado.
	if strings.Contains(text, "put a counter") && strings.Contains(text, "this turn") {
		primeCounterPlaced(gs)
		tr.Record("CONDITION_SETUP",
			"%q → +1/+1 counter on friendly creature",
			"put a counter on a creature this turn")
		primed = true
	}

	// Life-vs-starting threshold. Angel of Destiny: "at least 15 life
	// more than your starting life total".
	if m := reLifeMoreStarting.FindStringSubmatch(text); m != nil {
		n, _ := strconv.Atoi(m[1])
		primeLifeMoreThanStarting(gs, n)
		tr.Record("CONDITION_SETUP",
			"%q → set seat0 Life=%d",
			fmt.Sprintf("at least %d life more than starting", n),
			gs.Seats[0].Life)
		primed = true
	}

	// Ascend / city's blessing. Twilight Prophet.
	if strings.Contains(text, "city's blessing") || strings.Contains(text, "citys blessing") {
		primeAscend(gs)
		tr.Record("CONDITION_SETUP",
			"%q → 10 permanents on seat0, citys_blessing flag set",
			"city's blessing (ascend)")
		primed = true
	}

	// Subtype control. Acclaimed Contender ("if you control another
	// Knight"). The "another <subtype>" matcher takes precedence over
	// the bare "a <subtype>" matcher because a card can satisfy both
	// clauses ("control another Knight" implies "control a Knight").
	if m := reControlAnother.FindStringSubmatch(text); m != nil {
		subtype := strings.TrimSpace(m[1])
		if subtype != "" && !isGenericWord(subtype) {
			placeNamedFriendlyCreatureWithSubtype(gs,
				strings.Title(subtype)+" Buddy", subtype)
			tr.Record("CONDITION_SETUP",
				"%q → placed %s creature on seat0",
				"control another "+subtype, subtype)
			primed = true
		}
	} else if m := reControlSubtype.FindStringSubmatch(text); m != nil {
		subtype := strings.TrimSpace(m[1])
		if subtype != "" && !isGenericWord(subtype) {
			placeNamedFriendlyCreatureWithSubtype(gs,
				strings.Title(subtype)+" Buddy", subtype)
			tr.Record("CONDITION_SETUP",
				"%q → placed %s creature on seat0",
				"you control a "+subtype, subtype)
			primed = true
		} else if subtype == "creature" {
			// Birthing Ritual ("if you control a creature"). The
			// source itself usually counts (creatures bring their
			// own permanent), but enchantment sources need a
			// separate witness.
			if primeFriendlyCreatureIfMissing(gs) {
				tr.Record("CONDITION_SETUP",
					"%q → placed witness creature on seat0",
					"you control a creature")
				primed = true
			}
		}
	}

	// Exile linkage. Smirking Spelljacker ("if a card is exiled with
	// it"). Place a card in seat0's exile and tag the source with a
	// has-exiled-card flag so the per-card handler can find it.
	if (strings.Contains(text, "exiled with it") ||
		strings.Contains(text, "exiled with this") ||
		strings.Contains(text, "card is exiled with")) && srcPerm != nil {
		primeExiledWith(gs, srcPerm)
		tr.Record("CONDITION_SETUP",
			"%q → exile-linked card placed",
			"a card is exiled with it")
		primed = true
	}

	return primed
}

// ---------------------------------------------------------------------------
// Intervening-if priming helpers.
// ---------------------------------------------------------------------------

// primeGainedLife fires GainLife so per-card listeners (Lathiel, Heliod,
// Vito) increment their own counters, AND tops up seat.Flags so handlers
// (Sorin, Shanna) that read the seat-level flag directly see the gain.
// Both paths matter because the engine has not centralised turn-scoped
// life tracking.
func primeGainedLife(gs *gameengine.GameState, amount int) {
	if gs == nil || amount <= 0 || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return
	}
	gameengine.GainLife(gs, 0, amount, "thor_priming")
	if gs.Seats[0].Flags == nil {
		gs.Seats[0].Flags = map[string]int{}
	}
	gs.Seats[0].Flags["life_gained_this_turn"] += amount
}

// primeLostLife decrements seat 0 life and sets the seat-level flag. The
// engine has no public LoseLife helper, so we mutate Life and synthesise
// the trigger event for any per-card listeners.
func primeLostLife(gs *gameengine.GameState, amount int) {
	if gs == nil || amount <= 0 || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return
	}
	gs.Seats[0].Life -= amount
	if gs.Seats[0].Flags == nil {
		gs.Seats[0].Flags = map[string]int{}
	}
	gs.Seats[0].Flags["life_lost_this_turn"] += amount
	gs.Seats[0].Flags["lost_life_this_turn"] += amount
	gameengine.FireCardTrigger(gs, "life_lost", map[string]interface{}{
		"seat":   0,
		"amount": amount,
		"source": "thor_priming",
	})
}

// primeCounterPlaced ensures a friendly creature has a +1/+1 counter
// placed on it during the current turn and fires the engine's
// counter_placed trigger so listeners increment their this-turn flags.
func primeCounterPlaced(gs *gameengine.GameState) {
	if gs == nil || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return
	}
	var target *gameengine.Permanent
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.IsCreature() {
			target = p
			break
		}
	}
	if target == nil {
		target = placeNamedFriendlyCreature(gs, "Counter Recipient")
	}
	target.AddCounter("+1/+1", 1)
	if gs.Seats[0].Flags == nil {
		gs.Seats[0].Flags = map[string]int{}
	}
	gs.Seats[0].Flags["counter_placed_this_turn"] = 1
	if gs.Flags == nil {
		gs.Flags = map[string]int{}
	}
	gs.Flags["counter_placed_this_turn"] = 1
	gameengine.FireCardTrigger(gs, "counter_placed", map[string]interface{}{
		"target_perm":  target,
		"target_seat":  0,
		"counter_kind": "+1/+1",
		"amount":       1,
		"source_card":  "thor_priming",
		"source_seat":  0,
	})
}

// primeLifeMoreThanStarting raises seat 0 life so it sits exactly n above
// StartingLife. Falls back to a Commander default of 40 when StartingLife
// is unset (the goldilocks factory builds seats at Life=20 without
// populating StartingLife).
func primeLifeMoreThanStarting(gs *gameengine.GameState, n int) {
	if gs == nil || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return
	}
	starting := gs.Seats[0].StartingLife
	if starting <= 0 {
		starting = 40
	}
	target := starting + n
	if gs.Seats[0].Life < target {
		gs.Seats[0].Life = target
	}
}

// primeAscend places enough vanilla permanents on seat 0 to satisfy the
// 10-permanent threshold and eagerly sets the citys_blessing flag so
// readers that bypass CheckAscend (Twilight Prophet's intervening-if
// path) still observe the blessing.
func primeAscend(gs *gameengine.GameState) {
	if gs == nil || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return
	}
	seat := gs.Seats[0]
	for len(seat.Battlefield) < 10 {
		seat.Battlefield = append(seat.Battlefield, &gameengine.Permanent{
			Card: &gameengine.Card{
				Name:  fmt.Sprintf("Ascend Filler %d", len(seat.Battlefield)),
				Owner: 0,
				Types: []string{"land"},
			},
			Controller: 0,
			Owner:      0,
			Flags:      map[string]int{},
			Counters:   map[string]int{},
		})
	}
	if seat.Flags == nil {
		seat.Flags = map[string]int{}
	}
	seat.Flags["citys_blessing"] = 1
	gameengine.CheckAscend(gs, 0)
}

// primeFriendlyCreatureIfMissing places a witness creature on seat 0 only
// when none currently exists. Returns true if a creature was placed.
// Intended for "if you control a creature" preconditions on
// non-creature sources (Birthing Ritual is an enchantment).
func primeFriendlyCreatureIfMissing(gs *gameengine.GameState) bool {
	if gs == nil || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return false
	}
	for _, p := range gs.Seats[0].Battlefield {
		if p != nil && p.IsCreature() {
			return false
		}
	}
	placeNamedFriendlyCreature(gs, "Creature Witness")
	return true
}

// primeExiledWith places a card in seat 0's exile and flags srcPerm as
// having an exiled-with companion. Smirking Spelljacker and similar
// imprint-style cards check for this on attack.
func primeExiledWith(gs *gameengine.GameState, srcPerm *gameengine.Permanent) {
	if gs == nil || srcPerm == nil || len(gs.Seats) == 0 || gs.Seats[0] == nil {
		return
	}
	gs.Seats[0].Exile = append(gs.Seats[0].Exile, &gameengine.Card{
		Name:  "Exiled Companion",
		Owner: 0,
		Types: []string{"instant"},
	})
	if srcPerm.Flags == nil {
		srcPerm.Flags = map[string]int{}
	}
	srcPerm.Flags["card_exiled_with"] = 1
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

// ---------------------------------------------------------------------------
// Raw-condition scaffolding (Category B / D fix).
//
// The parser surfaces "if ..." and "as long as ..." clauses as Conditions
// with kind = intervening_if / as_long_as / conditional and the verbatim
// English in args[0]. The existing setupCondition switch only handles
// canonicalised kinds (fateful_hour, threshold, morbid, etc.), so cards
// like Land Tax ("if an opponent controls more lands than you...") and
// Oversold Cemetery ("if there are four or more creature cards in your
// graveyard...") never had their preconditions satisfied.
//
// detectConditionScaffold inspects the raw text and returns a shape; apply
// mutates gs accordingly; describe returns a one-line summary used for
// CONDITION_SETUP traces.
// ---------------------------------------------------------------------------

type conditionScaffoldKind int

const (
	condScaffoldNone conditionScaffoldKind = iota
	condScaffoldOpponentMoreLands
	condScaffoldYouControlSubtype
	condScaffoldCreatureDiedThisTurn
	condScaffoldCreatureCardsInGraveyard
	condScaffoldCardInGraveyard
	condScaffoldEnergyThreshold

	// Trigger-condition scaffold kinds — handle raw condition text that
	// describes a trigger event precondition. These bridge the gap when
	// the AST parser emits a raw/intervening_if condition instead of a
	// structured Trigger node, or when a static ability's condition text
	// references a trigger-like predicate.
	condScaffoldGainedLifeThisTurn
	condScaffoldCastSpellThisTurn
	condScaffoldCreatureETBThisTurn
	condScaffoldDrawnCardThisTurn
	condScaffoldAttackedThisTurn
	condScaffoldSacrificedThisTurn
	condScaffoldCombatDamageDealt
	condScaffoldLandfallThisTurn
	condScaffoldDiscardedThisTurn
	condScaffoldEnchantedCreature
	condScaffoldOpponentLostLife
	condScaffoldLifeAboveThreshold
	condScaffoldLifeBelowThreshold
	condScaffoldUpkeepPhase

	// Ability-word / status scaffold kinds — each mirrors a Check<Word>
	// helper in the engine. Detection accepts both the ability-word slug
	// ("hellbent", "delirium") and the verbatim English description that
	// the AST sometimes surfaces ("if you have no cards in hand", "four or
	// more card types in your graveyard").
	condScaffoldHellbent
	condScaffoldMonarch
	condScaffoldInitiative
	condScaffoldDelirium
	condScaffoldSpellMastery
	condScaffoldRevolt
	condScaffoldMetalcraft
	condScaffoldFerocious
	condScaffoldFormidable

	// New engine structure scaffold kinds — bridge the gap between
	// the TurnCounters migration and the scaffold priming system.
	condScaffoldPermanentLeftBF     // disappear / void — Turn.PermanentsLeft
	condScaffoldSecondSpellThisTurn // CastRecord nth-spell — Turn.Casts
	condScaffoldDescendedThisTurn   // descended — Turn.Descended
	condScaffoldLifeLostThisTurn    // opponent/self lost life — Turn.LifeLost
	condScaffoldTokensCreatedCount  // X = tokens created — Turn.TokensCreated
	condScaffoldCastFromExile       // cast from exile — Turn.CastFromExile
	condScaffoldExileLinkedReturn   // exile-until-leaves — ExileLinked/ReturnLinkedExile
)

type conditionScaffold struct {
	kind        conditionScaffoldKind
	description string // populated lazily by apply
	rawText     string

	// Shape-specific payload.
	subtype   string // for condScaffoldYouControlSubtype: e.g. "wizard"
	count     int    // for graveyard count / energy threshold
	threshold int    // for life threshold conditions
}

var (
	rawTribalRe   = regexp.MustCompile(`(?:another|a|an)\s+([a-z]+)`)
	energyAmtRe   = regexp.MustCompile(`(\d+)\s+or\s+more\s+energy`)
	graveCntRe    = regexp.MustCompile(`(?:(\d+)|one|two|three|four|five|six|seven)\s*(?:or\s+more\s+)?creature\s+cards?`)
	lifeAboveRe   = regexp.MustCompile(`(?:you have|your life total is)\s+(\d+)\s+or\s+more\s+life`)
	lifeBelowRe   = regexp.MustCompile(`(?:you have|your life total is)\s+(\d+)\s+or\s+(?:less|fewer)\s+life`)
	lifeBelowAltRe = regexp.MustCompile(`life total is\s+(\d+)\s+or\s+less`)
	// CastRecord nth-spell pattern — "second spell", "third creature", etc.
	secondSpellRe = regexp.MustCompile(`(?:second|third)\s+(?:spell|creature|noncreature|instant|sorcery|artifact|enchantment)`)
	// Ability-word fingerprints. Each accepts either the ability word
	// itself or the canonical English description so we catch both AST
	// forms.
	deliriumRe       = regexp.MustCompile(`four\s+or\s+more\s+card\s+types`)
	spellMasteryRe   = regexp.MustCompile(`(?:two|2)\s+or\s+more\s+(?:instant|sorcery)`)
	metalcraftRe     = regexp.MustCompile(`(?:three|3)\s+or\s+more\s+artifacts?`)
	ferociousRe      = regexp.MustCompile(`creature\s+with\s+power\s+(?:four|4)\s+or\s+(?:more|greater)`)
	formidableRe     = regexp.MustCompile(`total\s+power\s+(?:eight|8)\s+or\s+(?:more|greater)`)
)

func conditionRawText(cond *gameast.Condition) string {
	if cond == nil || len(cond.Args) == 0 {
		return ""
	}
	s, _ := cond.Args[0].(string)
	return strings.ToLower(strings.TrimSpace(s))
}

// detectConditionScaffold examines a raw-text condition and returns the
// shape that scaffolding should match. The shape is then handed to apply
// or describe — both share this detection so the trace text and the
// mutation always agree.
func detectConditionScaffold(cond *gameast.Condition) conditionScaffold {
	if cond == nil {
		return conditionScaffold{}
	}
	switch strings.ToLower(cond.Kind) {
	case "intervening_if", "as_long_as", "conditional", "raw":
		// proceed
	default:
		return conditionScaffold{}
	}
	txt := conditionRawText(cond)
	if txt == "" {
		return conditionScaffold{}
	}
	cs := conditionScaffold{rawText: txt}

	// "an opponent controls more lands than you" / "more lands than you do"
	if (strings.Contains(txt, "more land") || strings.Contains(txt, "controls more")) &&
		strings.Contains(txt, "than you") {
		cs.kind = condScaffoldOpponentMoreLands
		return cs
	}

	// "if/as long as a creature died this turn"
	if strings.Contains(txt, "died this turn") {
		cs.kind = condScaffoldCreatureDiedThisTurn
		return cs
	}

	// Delirium — 4+ distinct card types in your graveyard. Must come
	// before the generic graveyard matchers below so the more permissive
	// CardInGraveyard case doesn't swallow it.
	if strings.Contains(txt, "delirium") || deliriumRe.MatchString(txt) {
		cs.kind = condScaffoldDelirium
		return cs
	}

	// Spell Mastery — 2+ instant/sorcery in your graveyard. Same
	// ordering rationale as Delirium.
	if strings.Contains(txt, "spell mastery") ||
		(spellMasteryRe.MatchString(txt) && strings.Contains(txt, "graveyard")) {
		cs.kind = condScaffoldSpellMastery
		return cs
	}

	// Graveyard count: "four or more creature cards in your graveyard"
	if strings.Contains(txt, "graveyard") &&
		(strings.Contains(txt, "creature card") || strings.Contains(txt, "creatures in")) {
		cs.kind = condScaffoldCreatureCardsInGraveyard
		cs.count = parseGraveyardCount(txt)
		if cs.count < 4 {
			cs.count = 4
		}
		return cs
	}

	// Generic "card in your graveyard" — Necromancy, return-from-graveyard.
	if strings.Contains(txt, "graveyard") {
		cs.kind = condScaffoldCardInGraveyard
		return cs
	}

	// Energy counter threshold ("if you have N or more energy counters").
	if strings.Contains(txt, "energy") {
		n := 30
		if m := energyAmtRe.FindStringSubmatch(txt); len(m) > 1 {
			if v, err := strconv.Atoi(m[1]); err == nil && v > 0 {
				n = v
			}
		}
		cs.kind = condScaffoldEnergyThreshold
		cs.count = n
		return cs
	}

	// Life gained this turn. Sorin, Lathiel, Heliod.
	if (strings.Contains(txt, "gained life") || strings.Contains(txt, "gain life")) &&
		strings.Contains(txt, "this turn") {
		cs.kind = condScaffoldGainedLifeThisTurn
		return cs
	}

	// Cast a spell this turn. Monastery Mentor, Prowess-adjacent.
	if (strings.Contains(txt, "cast a spell") || strings.Contains(txt, "cast a noncreature") ||
		strings.Contains(txt, "cast an instant") || strings.Contains(txt, "you cast")) &&
		strings.Contains(txt, "this turn") {
		cs.kind = condScaffoldCastSpellThisTurn
		return cs
	}

	// Creature ETB this turn. "if a creature entered the battlefield"
	if (strings.Contains(txt, "creature entered") || strings.Contains(txt, "creature enters")) &&
		(strings.Contains(txt, "this turn") || strings.Contains(txt, "battlefield")) {
		cs.kind = condScaffoldCreatureETBThisTurn
		return cs
	}

	// Drew a card this turn. "if you've drawn a card this turn"
	if (strings.Contains(txt, "drew a card") || strings.Contains(txt, "drawn a card") ||
		strings.Contains(txt, "draw a card")) &&
		strings.Contains(txt, "this turn") {
		cs.kind = condScaffoldDrawnCardThisTurn
		return cs
	}

	// Attacked this turn. "if you attacked this turn" / "if you attacked with a creature"
	if (strings.Contains(txt, "attacked this turn") || strings.Contains(txt, "you attacked") ||
		strings.Contains(txt, "creature attacked")) {
		cs.kind = condScaffoldAttackedThisTurn
		return cs
	}

	// Sacrificed this turn. "if you sacrificed a creature this turn"
	if strings.Contains(txt, "sacrific") && strings.Contains(txt, "this turn") {
		cs.kind = condScaffoldSacrificedThisTurn
		return cs
	}

	// Combat damage dealt this turn. "if a creature dealt combat damage"
	if strings.Contains(txt, "combat damage") &&
		(strings.Contains(txt, "this turn") || strings.Contains(txt, "to a player") ||
			strings.Contains(txt, "dealt")) {
		cs.kind = condScaffoldCombatDamageDealt
		return cs
	}

	// Landfall. "if a land entered the battlefield" / "landfall" / "if you played a land"
	if strings.Contains(txt, "landfall") ||
		(strings.Contains(txt, "land") && strings.Contains(txt, "entered")) ||
		(strings.Contains(txt, "played a land") && strings.Contains(txt, "this turn")) {
		cs.kind = condScaffoldLandfallThisTurn
		return cs
	}

	// Discarded this turn. "if you discarded a card this turn"
	if strings.Contains(txt, "discard") && strings.Contains(txt, "this turn") {
		cs.kind = condScaffoldDiscardedThisTurn
		return cs
	}

	// Enchanted creature. "enchanted creature" condition for auras.
	if strings.Contains(txt, "enchanted creature") {
		cs.kind = condScaffoldEnchantedCreature
		return cs
	}

	// Opponent lost life this turn. "if an opponent lost life this turn"
	if strings.Contains(txt, "opponent") &&
		(strings.Contains(txt, "lost life") || strings.Contains(txt, "lose life") ||
			strings.Contains(txt, "dealt damage")) &&
		strings.Contains(txt, "this turn") {
		cs.kind = condScaffoldOpponentLostLife
		return cs
	}

	// Life above threshold. "if you have 25 or more life"
	if m := lifeAboveRe.FindStringSubmatch(txt); m != nil {
		n, _ := strconv.Atoi(m[1])
		cs.kind = condScaffoldLifeAboveThreshold
		cs.threshold = n
		return cs
	}

	// Life below threshold. "if you have 5 or less life" / "your life total is 5 or less"
	if m := lifeBelowRe.FindStringSubmatch(txt); m != nil {
		n, _ := strconv.Atoi(m[1])
		cs.kind = condScaffoldLifeBelowThreshold
		cs.threshold = n
		return cs
	}
	if m := lifeBelowAltRe.FindStringSubmatch(txt); m != nil {
		n, _ := strconv.Atoi(m[1])
		cs.kind = condScaffoldLifeBelowThreshold
		cs.threshold = n
		return cs
	}

	// Upkeep phase condition. "during your upkeep" / "it's your upkeep"
	if strings.Contains(txt, "upkeep") {
		cs.kind = condScaffoldUpkeepPhase
		return cs
	}

	// Hellbent — "if you have no cards in hand" / ability word.
	if strings.Contains(txt, "hellbent") ||
		(strings.Contains(txt, "no cards in") && strings.Contains(txt, "hand")) {
		cs.kind = condScaffoldHellbent
		return cs
	}

	// Monarch — "if you're the monarch" / "you are the monarch".
	if strings.Contains(txt, "the monarch") ||
		strings.Contains(txt, "you're the monarch") ||
		strings.Contains(txt, "you are the monarch") {
		cs.kind = condScaffoldMonarch
		return cs
	}

	// Initiative — "if you have the initiative".
	if strings.Contains(txt, "the initiative") ||
		strings.Contains(txt, "have initiative") {
		cs.kind = condScaffoldInitiative
		return cs
	}

	// Revolt — "a permanent you controlled left the battlefield this turn".
	if strings.Contains(txt, "revolt") ||
		(strings.Contains(txt, "permanent") &&
			strings.Contains(txt, "left the battlefield") &&
			strings.Contains(txt, "this turn")) {
		cs.kind = condScaffoldRevolt
		return cs
	}

	// Metalcraft — "if you control three or more artifacts".
	if strings.Contains(txt, "metalcraft") ||
		(metalcraftRe.MatchString(txt) && strings.Contains(txt, "you control")) {
		cs.kind = condScaffoldMetalcraft
		return cs
	}

	// Ferocious — "if you control a creature with power 4 or greater".
	if strings.Contains(txt, "ferocious") || ferociousRe.MatchString(txt) {
		cs.kind = condScaffoldFerocious
		return cs
	}

	// Formidable — "creatures you control have total power 8 or greater".
	if strings.Contains(txt, "formidable") || formidableRe.MatchString(txt) {
		cs.kind = condScaffoldFormidable
		return cs
	}

	// Permanent left the battlefield. Disappear / Void ability words and
	// revolt-like conditions that don't use the "revolt" keyword.
	if (strings.Contains(txt, "permanent left") || strings.Contains(txt, "permanent left the battlefield")) &&
		!strings.Contains(txt, "revolt") {
		cs.kind = condScaffoldPermanentLeftBF
		return cs
	}

	// Second/third spell this turn. CastRecord-dependent conditions.
	if secondSpellRe.MatchString(txt) {
		cs.kind = condScaffoldSecondSpellThisTurn
		return cs
	}

	// Descended this turn. Ixalan keyword.
	if strings.Contains(txt, "descended") || strings.Contains(txt, "descend") {
		cs.kind = condScaffoldDescendedThisTurn
		return cs
	}

	// Life lost this turn (self or opponent).
	if (strings.Contains(txt, "lost life") || strings.Contains(txt, "lose life")) &&
		strings.Contains(txt, "this turn") &&
		!strings.Contains(txt, "opponent") {
		cs.kind = condScaffoldLifeLostThisTurn
		return cs
	}

	// Tokens created count. Thalisse / Vazi / Ellyn Harbreeze.
	if strings.Contains(txt, "tokens") &&
		(strings.Contains(txt, "created this turn") || strings.Contains(txt, "you created this turn")) {
		cs.kind = condScaffoldTokensCreatedCount
		return cs
	}

	// Cast from exile. Impulse draw and flashback conditions.
	if strings.Contains(txt, "cast") && strings.Contains(txt, "from exile") &&
		!strings.Contains(txt, "you may cast") {
		cs.kind = condScaffoldCastFromExile
		return cs
	}

	// Exile-linked return. O-Ring / Fiend Hunter patterns.
	if strings.Contains(txt, "exiled with") &&
		(strings.Contains(txt, "return") || strings.Contains(txt, "leaves")) {
		cs.kind = condScaffoldExileLinkedReturn
		return cs
	}

	// Tribal: "if you control another <subtype>" / "if you control a <subtype>".
	// Run last because it's the most permissive matcher.
	if strings.Contains(txt, "you control") {
		// Strip the leading "you control" segment so we match the subtype.
		idx := strings.Index(txt, "you control")
		tail := txt[idx+len("you control"):]
		if m := rawTribalRe.FindStringSubmatch(strings.TrimSpace(tail)); len(m) > 1 {
			subtype := m[1]
			// Filter out generic words that happen to follow "another"/"a".
			if !isGenericWord(subtype) {
				cs.kind = condScaffoldYouControlSubtype
				cs.subtype = subtype
				return cs
			}
		}
	}

	return conditionScaffold{}
}

func isGenericWord(s string) bool {
	switch s {
	case "creature", "permanent", "card", "spell", "ability",
		"player", "opponent", "thing", "land", "artifact",
		"enchantment", "planeswalker":
		return true
	}
	return false
}

func parseGraveyardCount(txt string) int {
	if m := graveCntRe.FindStringSubmatch(txt); len(m) > 0 {
		if m[1] != "" {
			if v, err := strconv.Atoi(m[1]); err == nil {
				return v
			}
		}
		// Word-form numbers.
		for word, n := range map[string]int{
			"one": 1, "two": 2, "three": 3, "four": 4,
			"five": 5, "six": 6, "seven": 7,
		} {
			if strings.Contains(m[0], word) {
				return n
			}
		}
	}
	return 0
}

// applyConditionScaffolding mutates gs to satisfy the condition's shape.
// Returns the same scaffold descriptor with description populated, or a
// zero-value when nothing applied. Idempotent across types where it
// matters (it tops up rather than appending unconditionally).
func applyConditionScaffolding(gs *gameengine.GameState, cond *gameast.Condition, srcPerm *gameengine.Permanent) conditionScaffold {
	cs := detectConditionScaffold(cond)
	if cs.kind == condScaffoldNone || gs == nil {
		return conditionScaffold{}
	}

	switch cs.kind {
	case condScaffoldOpponentMoreLands:
		seedSeatLands(gs, 1, 6, "Plains", "plains")
		cs.description = "seeded 6 Plains on seat 1"

	case condScaffoldYouControlSubtype:
		placeNamedFriendlyCreatureWithSubtype(gs, "Tribal "+cs.subtype, cs.subtype)
		cs.description = fmt.Sprintf("placed %s creature on seat 0", cs.subtype)

	case condScaffoldCreatureDiedThisTurn:
		// Engine reads seat.Turn.CreaturesDied for morbid checks (since
		// TurnCounters migration). Also set legacy flag + place creature
		// card in graveyard so graveyard scans see a death event.
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["creature_died_this_turn"] = 1
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.CreaturesDied++
			gs.Seats[0].Turn.PermanentsLeft++
			gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, &gameengine.Card{
				Name:          "Died Setup",
				Owner:         0,
				Types:         []string{"creature"},
				BasePower:     1,
				BaseToughness: 1,
			})
		}
		cs.description = "set Turn.CreaturesDied + legacy flag + added creature to graveyard"

	case condScaffoldCreatureCardsInGraveyard:
		topUpGraveyardCreatures(gs, 0, cs.count)
		cs.description = fmt.Sprintf("populated seat 0 graveyard with %d creature cards", cs.count)

	case condScaffoldCardInGraveyard:
		topUpGraveyardCreatures(gs, 0, 1)
		cs.description = "placed creature card in seat 0 graveyard"

	case condScaffoldEnergyThreshold:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["energy_counters"] = cs.count
		}
		cs.description = fmt.Sprintf("set seat 0 energy counters to %d", cs.count)

	case condScaffoldGainedLifeThisTurn:
		primeGainedLife(gs, 3)
		cs.description = "gained 3 life for seat 0 (life_gained_this_turn flag set)"

	case condScaffoldCastSpellThisTurn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].SpellsCastThisTurn++
			gs.Seats[0].Turn.SpellsCast++
			gs.Seats[0].Turn.Casts = append(gs.Seats[0].Turn.Casts, gameengine.CastRecord{
				CardName:  "Scaffold Spell",
				Types:     []string{"instant"},
				ManaValue: 2,
			})
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["cast_spell_this_turn"] = 1
			gs.Seats[0].Flags["spells_cast_this_turn"] = gs.Seats[0].Turn.SpellsCast
		}
		gs.SpellsCastThisTurn++
		cs.description = "incremented Turn.SpellsCast + legacy cast counters for seat 0"

	case condScaffoldCreatureETBThisTurn:
		placeNamedFriendlyCreature(gs, "ETB Witness")
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.CreaturesEntered++
		}
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["creature_etb_this_turn"] = 1
		cs.description = "set Turn.CreaturesEntered + legacy flag + placed ETB Witness"

	case condScaffoldDrawnCardThisTurn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.CardsDrawn++
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["drawn_card_this_turn"] = 1
			if len(gs.Seats[0].Library) < 5 {
				fillLibrary(gs, 0, 5)
			}
		}
		cs.description = "set Turn.CardsDrawn + legacy flag + filled library"

	case condScaffoldAttackedThisTurn:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["attacked_this_turn"] = 1
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.Attacked = true
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["attacked_this_turn"] = 1
		}
		cs.description = "set Turn.Attacked + legacy flag on seat 0 and game"

	case condScaffoldSacrificedThisTurn:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["sacrificed_this_turn"] = 1
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.Sacrificed++
			gs.Seats[0].Turn.PermanentsLeft++
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["sacrificed_this_turn"] = 1
			gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, &gameengine.Card{
				Name:          "Sac Victim Setup",
				Owner:         0,
				Types:         []string{"creature"},
				BasePower:     1,
				BaseToughness: 1,
			})
		}
		cs.description = "set Turn.Sacrificed + legacy flag + placed creature in graveyard"

	case condScaffoldCombatDamageDealt:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["combat_damage_dealt_this_turn"] = 1
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["combat_damage_dealt_this_turn"] = 1
		}
		cs.description = "set combat_damage_dealt_this_turn flag"

	case condScaffoldLandfallThisTurn:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["landfall_this_turn"] = 1
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.LandsPlayed++
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["landfall_this_turn"] = 1
			// Place a land on the battlefield to represent the landfall trigger source.
			gs.Seats[0].Battlefield = append(gs.Seats[0].Battlefield, &gameengine.Permanent{
				Card: &gameengine.Card{
					Name:  "Landfall Setup Land",
					Owner: 0,
					Types: []string{"land", "forest"},
				},
				Controller: 0,
				Owner:      0,
				Flags:      map[string]int{},
				Counters:   map[string]int{},
			})
		}
		cs.description = "set landfall_this_turn flag + placed land on seat 0"

	case condScaffoldDiscardedThisTurn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["discarded_this_turn"] = 1
			gs.Seats[0].Turn.Discarded++
			gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, &gameengine.Card{
				Name:  "Discarded Setup",
				Owner: 0,
				Types: []string{"instant"},
			})
			if len(gs.Seats[0].Hand) < 3 {
				fillHand(gs, 0, 3-len(gs.Seats[0].Hand))
			}
		}
		cs.description = "set Turn.Discarded + legacy flag + placed card in graveyard"

	case condScaffoldEnchantedCreature:
		// For aura conditions that reference "enchanted creature", we need
		// a creature on the battlefield with the source attached to it.
		target := placeNamedFriendlyCreature(gs, "Enchanted Target")
		if srcPerm != nil {
			srcPerm.AttachedTo = target
		}
		cs.description = "placed Enchanted Target creature for aura attachment"

	case condScaffoldOpponentLostLife:
		if len(gs.Seats) > 1 && gs.Seats[1] != nil {
			gs.Seats[1].Life -= 3
			if gs.Seats[1].Flags == nil {
				gs.Seats[1].Flags = map[string]int{}
			}
			gs.Seats[1].Flags["life_lost_this_turn"] = 3
			gs.Seats[1].Flags["lost_life_this_turn"] = 3
		}
		cs.description = "opponent (seat 1) lost 3 life this turn"

	case condScaffoldLifeAboveThreshold:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			if gs.Seats[0].Life < cs.threshold {
				gs.Seats[0].Life = cs.threshold
			}
		}
		cs.description = fmt.Sprintf("set seat 0 life to %d (above threshold)", cs.threshold)

	case condScaffoldLifeBelowThreshold:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			if gs.Seats[0].Life > cs.threshold {
				gs.Seats[0].Life = cs.threshold
			}
		}
		cs.description = fmt.Sprintf("set seat 0 life to %d (below threshold)", cs.threshold)

	case condScaffoldUpkeepPhase:
		gs.Phase = "beginning"
		gs.Step = "upkeep"
		cs.description = "set game phase to upkeep"

	case condScaffoldHellbent:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Hand = nil
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["hellbent"] = 1
		}
		cs.description = "emptied seat 0 hand (hellbent active)"

	case condScaffoldMonarch:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["has_monarch"] = 1
		gs.Flags["monarch_seat"] = 0
		cs.description = "made seat 0 the monarch"

	case condScaffoldInitiative:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["initiative_holder"] = 0
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["has_initiative"] = 1
		}
		cs.description = "gave seat 0 the initiative"

	case condScaffoldDelirium:
		seedDeliriumGraveyard(gs, 0)
		cs.description = "seeded seat 0 graveyard with 4 distinct card types"

	case condScaffoldSpellMastery:
		seedSpellMasteryGraveyard(gs, 0)
		cs.description = "seeded seat 0 graveyard with 2 instant/sorcery cards"

	case condScaffoldRevolt:
		// Revolt reads gs.EventLog for a destroy/sacrifice/exile/bounce
		// event by seat 0 this turn. Append a synthesised sacrifice event
		// rather than mutating actual zones — the engine's CheckRevolt only
		// consults the log.
		gs.EventLog = append(gs.EventLog, gameengine.Event{
			Kind:   "sacrifice",
			Seat:   0,
			Target: -1,
			Source: "thor_priming",
		})
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["revolt_active"] = 1
		cs.description = "logged sacrifice event for seat 0 (revolt active)"

	case condScaffoldMetalcraft:
		seedSeatArtifacts(gs, 0, 3)
		cs.description = "placed 3 artifacts on seat 0 (metalcraft active)"

	case condScaffoldFerocious:
		placePoweredCreature(gs, 0, "Ferocious Setup", 4, 4)
		cs.description = "placed 4/4 creature on seat 0 (ferocious active)"

	case condScaffoldFormidable:
		placePoweredCreature(gs, 0, "Formidable Setup A", 4, 4)
		placePoweredCreature(gs, 0, "Formidable Setup B", 4, 4)
		cs.description = "placed creatures totaling 8 power on seat 0 (formidable active)"

	case condScaffoldPermanentLeftBF:
		if gs.Flags == nil {
			gs.Flags = map[string]int{}
		}
		gs.Flags["permanent_left_bf"] = 1
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.PermanentsLeft++
			gs.Flags["permanent_left_bf_0"] = 1
		}
		gs.EventLog = append(gs.EventLog, gameengine.Event{
			Kind:   "sacrifice",
			Seat:   0,
			Target: -1,
			Source: "thor_priming",
		})
		cs.description = "set Turn.PermanentsLeft + permanent_left_bf flag (disappear/void)"

	case condScaffoldSecondSpellThisTurn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.SpellsCast = 2
			gs.Seats[0].SpellsCastThisTurn = 2
			gs.Seats[0].Turn.Casts = append(gs.Seats[0].Turn.Casts,
				gameengine.CastRecord{CardName: "Scaffold Spell 1", Types: []string{"instant"}, ManaValue: 1},
				gameengine.CastRecord{CardName: "Scaffold Spell 2", Types: []string{"sorcery"}, ManaValue: 2},
			)
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["spells_cast_this_turn"] = 2
			gs.Seats[0].Flags["cast_spell_this_turn"] = 1
		}
		gs.SpellsCastThisTurn = 2
		cs.description = "set Turn.SpellsCast=2 + 2 CastRecords (second spell scaffold)"

	case condScaffoldDescendedThisTurn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.Descended = true
			gs.Seats[0].DescendedThisTurn = true
			gs.Seats[0].Graveyard = append(gs.Seats[0].Graveyard, &gameengine.Card{
				Name:  "Descended Setup",
				Owner: 0,
				Types: []string{"creature"},
			})
		}
		cs.description = "set Turn.Descended + legacy DescendedThisTurn + creature in GY"

	case condScaffoldLifeLostThisTurn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.LifeLost += 3
			if gs.Seats[0].Flags == nil {
				gs.Seats[0].Flags = map[string]int{}
			}
			gs.Seats[0].Flags["life_lost_this_turn"] = 3
			gs.Seats[0].Flags["lost_life_this_turn"] = 3
		}
		cs.description = "set Turn.LifeLost=3 + legacy flags (life lost this turn)"

	case condScaffoldTokensCreatedCount:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.TokensCreated = 3
			gs.Seats[0].Turn.TreasuresCreated = 2
		}
		cs.description = "set Turn.TokensCreated=3, TreasuresCreated=2"

	case condScaffoldCastFromExile:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil {
			gs.Seats[0].Turn.CastFromExile++
		}
		cs.description = "incremented Turn.CastFromExile"

	case condScaffoldExileLinkedReturn:
		if len(gs.Seats) > 0 && gs.Seats[0] != nil && srcPerm != nil {
			exiledCard := &gameengine.Card{
				Name:  "Exile-Linked Target",
				Owner: 1,
				Types: []string{"creature"},
			}
			srcPerm.LinkedExile = append(srcPerm.LinkedExile, exiledCard)
			gs.Seats[1].Exile = append(gs.Seats[1].Exile, exiledCard)
		}
		cs.description = "attached exile-linked card to source permanent (O-Ring scaffold)"
	}
	return cs
}

// traceConditionScaffolding emits a CONDITION_SETUP entry describing the
// scaffold that would be / was applied for cond. Pure observation — no
// mutation. Safe to call with a nil tracer (no-op).
func traceConditionScaffolding(cond *gameast.Condition, tr *Tracer) {
	if tr == nil {
		return
	}
	cs := detectConditionScaffold(cond)
	if cs.kind == condScaffoldNone {
		return
	}
	// Build the description without mutating; mirrors apply's switch.
	desc := ""
	switch cs.kind {
	case condScaffoldOpponentMoreLands:
		desc = "seeded 6 Plains on seat 1"
	case condScaffoldYouControlSubtype:
		desc = fmt.Sprintf("placed %s creature on seat 0", cs.subtype)
	case condScaffoldCreatureDiedThisTurn:
		desc = "set creature_died_this_turn flag + added creature to graveyard"
	case condScaffoldCreatureCardsInGraveyard:
		desc = fmt.Sprintf("populated seat 0 graveyard with %d creature cards", cs.count)
	case condScaffoldCardInGraveyard:
		desc = "placed creature card in seat 0 graveyard"
	case condScaffoldEnergyThreshold:
		desc = fmt.Sprintf("set seat 0 energy counters to %d", cs.count)
	case condScaffoldGainedLifeThisTurn:
		desc = "gained 3 life for seat 0 (life_gained_this_turn flag set)"
	case condScaffoldCastSpellThisTurn:
		desc = "incremented spell cast counters for seat 0"
	case condScaffoldCreatureETBThisTurn:
		desc = "placed ETB Witness creature and set creature_etb_this_turn flag"
	case condScaffoldDrawnCardThisTurn:
		desc = "set drawn_card_this_turn flag + filled library"
	case condScaffoldAttackedThisTurn:
		desc = "set attacked_this_turn flag on seat 0 and game"
	case condScaffoldSacrificedThisTurn:
		desc = "set sacrificed_this_turn flag + placed creature in graveyard"
	case condScaffoldCombatDamageDealt:
		desc = "set combat_damage_dealt_this_turn flag"
	case condScaffoldLandfallThisTurn:
		desc = "set landfall_this_turn flag + placed land on seat 0"
	case condScaffoldDiscardedThisTurn:
		desc = "set discarded_this_turn flag + placed card in graveyard"
	case condScaffoldEnchantedCreature:
		desc = "placed Enchanted Target creature for aura attachment"
	case condScaffoldOpponentLostLife:
		desc = "opponent (seat 1) lost 3 life this turn"
	case condScaffoldLifeAboveThreshold:
		desc = fmt.Sprintf("set seat 0 life to %d (above threshold)", cs.threshold)
	case condScaffoldLifeBelowThreshold:
		desc = fmt.Sprintf("set seat 0 life to %d (below threshold)", cs.threshold)
	case condScaffoldUpkeepPhase:
		desc = "set game phase to upkeep"
	case condScaffoldHellbent:
		desc = "emptied seat 0 hand (hellbent active)"
	case condScaffoldMonarch:
		desc = "made seat 0 the monarch"
	case condScaffoldInitiative:
		desc = "gave seat 0 the initiative"
	case condScaffoldDelirium:
		desc = "seeded seat 0 graveyard with 4 distinct card types"
	case condScaffoldSpellMastery:
		desc = "seeded seat 0 graveyard with 2 instant/sorcery cards"
	case condScaffoldRevolt:
		desc = "logged sacrifice event for seat 0 (revolt active)"
	case condScaffoldMetalcraft:
		desc = "placed 3 artifacts on seat 0 (metalcraft active)"
	case condScaffoldFerocious:
		desc = "placed 4/4 creature on seat 0 (ferocious active)"
	case condScaffoldFormidable:
		desc = "placed creatures totaling 8 power on seat 0 (formidable active)"
	}
	tr.Record("CONDITION_SETUP", "%q → %s", cs.rawText, desc)
}

// ---------------------------------------------------------------------------
// Mutation helpers used by applyConditionScaffolding.
// ---------------------------------------------------------------------------

// seedSeatLands tops up `seat` so it controls at least `count` basic lands
// of the given subtype. Existing matching lands are kept; only the deficit
// is appended.
func seedSeatLands(gs *gameengine.GameState, seat, count int, name, subtype string) {
	if seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return
	}
	have := 0
	for _, p := range gs.Seats[seat].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		isLand := false
		for _, t := range p.Card.Types {
			if t == "land" {
				isLand = true
				break
			}
		}
		if isLand {
			have++
		}
	}
	for i := have; i < count; i++ {
		perm := &gameengine.Permanent{
			Card: &gameengine.Card{
				Name:  fmt.Sprintf("%s %d", name, i),
				Owner: seat,
				Types: []string{"land", subtype},
			},
			Controller: seat,
			Owner:      seat,
			Flags:      map[string]int{},
			Counters:   map[string]int{},
		}
		gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, perm)
	}
}

// placeNamedFriendlyCreatureWithSubtype is like placeNamedFriendlyCreature
// but appends a creature subtype (Wizard, Knight, etc.) so "you control
// another <subtype>" checks resolve true.
func placeNamedFriendlyCreatureWithSubtype(gs *gameengine.GameState, name, subtype string) *gameengine.Permanent {
	perm := &gameengine.Permanent{
		Card: &gameengine.Card{
			Name:          name,
			Owner:         0,
			Types:         []string{"creature", subtype},
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

// topUpGraveyardCreatures ensures `seat`'s graveyard contains at least n
// creature cards, appending tokens-by-name when short.
func topUpGraveyardCreatures(gs *gameengine.GameState, seat, n int) {
	if seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return
	}
	have := 0
	for _, c := range gs.Seats[seat].Graveyard {
		if c == nil {
			continue
		}
		for _, t := range c.Types {
			if t == "creature" {
				have++
				break
			}
		}
	}
	for i := have; i < n; i++ {
		gs.Seats[seat].Graveyard = append(gs.Seats[seat].Graveyard, &gameengine.Card{
			Name:          fmt.Sprintf("GraveCreatureSetup %d-%d", seat, i),
			Owner:         seat,
			Types:         []string{"creature"},
			BasePower:     2,
			BaseToughness: 2,
		})
	}
}

// seedDeliriumGraveyard tops up `seat`'s graveyard so CheckDelirium passes:
// 4 cards covering 4 distinct delirium-counted types. Existing graveyard
// contents are kept; only missing types are appended.
func seedDeliriumGraveyard(gs *gameengine.GameState, seat int) {
	if seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return
	}
	have := map[string]bool{}
	for _, c := range gs.Seats[seat].Graveyard {
		if c == nil {
			continue
		}
		for _, t := range c.Types {
			have[strings.ToLower(t)] = true
		}
	}
	// Cover the four most universal types — they don't overlap in a way
	// that would let CheckDelirium double-count.
	for _, ty := range []string{"creature", "instant", "sorcery", "artifact"} {
		if have[ty] {
			continue
		}
		gs.Seats[seat].Graveyard = append(gs.Seats[seat].Graveyard, &gameengine.Card{
			Name:  fmt.Sprintf("Delirium Setup %s", ty),
			Owner: seat,
			Types: []string{ty},
		})
	}
}

// seedSpellMasteryGraveyard tops up `seat`'s graveyard with 2 instant or
// sorcery cards so CheckSpellMastery passes. Counts existing entries so
// repeated calls don't pile up duplicates.
func seedSpellMasteryGraveyard(gs *gameengine.GameState, seat int) {
	if seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return
	}
	have := 0
	for _, c := range gs.Seats[seat].Graveyard {
		if c == nil {
			continue
		}
		for _, t := range c.Types {
			lower := strings.ToLower(t)
			if lower == "instant" || lower == "sorcery" {
				have++
				break
			}
		}
	}
	for i := have; i < 2; i++ {
		ty := "instant"
		if i%2 == 1 {
			ty = "sorcery"
		}
		gs.Seats[seat].Graveyard = append(gs.Seats[seat].Graveyard, &gameengine.Card{
			Name:  fmt.Sprintf("SpellMastery Setup %d", i),
			Owner: seat,
			Types: []string{ty},
		})
	}
}

// seedSeatArtifacts tops up `seat`'s battlefield so it controls at least
// `count` artifacts. Existing artifacts are kept; only the deficit is
// appended. Used for metalcraft priming.
func seedSeatArtifacts(gs *gameengine.GameState, seat, count int) {
	if seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return
	}
	have := 0
	for _, p := range gs.Seats[seat].Battlefield {
		if p == nil || p.Card == nil {
			continue
		}
		for _, t := range p.Card.Types {
			if t == "artifact" {
				have++
				break
			}
		}
	}
	for i := have; i < count; i++ {
		gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, &gameengine.Permanent{
			Card: &gameengine.Card{
				Name:  fmt.Sprintf("Metalcraft Setup %d", i),
				Owner: seat,
				Types: []string{"artifact"},
			},
			Controller: seat,
			Owner:      seat,
			Flags:      map[string]int{},
			Counters:   map[string]int{},
		})
	}
}

// placePoweredCreature appends a creature with the given P/T to `seat`'s
// battlefield. Used by ferocious / formidable priming where the threshold
// check depends on a specific power level.
func placePoweredCreature(gs *gameengine.GameState, seat int, name string, power, toughness int) *gameengine.Permanent {
	if seat >= len(gs.Seats) || gs.Seats[seat] == nil {
		return nil
	}
	perm := &gameengine.Permanent{
		Card: &gameengine.Card{
			Name:          name,
			Owner:         seat,
			Types:         []string{"creature"},
			BasePower:     power,
			BaseToughness: toughness,
		},
		Controller: seat,
		Owner:      seat,
		Flags:      map[string]int{},
		Counters:   map[string]int{},
	}
	gs.Seats[seat].Battlefield = append(gs.Seats[seat].Battlefield, perm)
	return perm
}
