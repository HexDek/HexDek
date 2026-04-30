# Trigger Dispatch

> Last updated: 2026-04-29
> Source: `internal/gameengine/triggers.go`, `trigger_stack_bridge.go`, `observer_triggers.go`, `event_aliases.go`
> CR refs: §603, §101.4

When abilities fire simultaneously, group by controller and push in [APNAP](APNAP.md) order; LIFO stack means last group resolves first.

## Trigger Lifecycle

```mermaid
flowchart TD
    Event[Game event<br/>cast / ETB / die /<br/>combat damage / etc.] --> Detect[Engine detects<br/>matching triggers]
    Detect --> AST[AST-declared triggers]
    Detect --> PerCardPer-card hooks]
    Detect --> Observer[Observer triggers<br/>passive on-board cards]
    AST --> Pool[Triggered pool]
    PerCard --> Pool
    Observer --> Pool
    Pool --> Order[OrderTriggersAPNAP]
    Order --> AP[Active player triggers<br/>pushed FIRST]
    Order --> NAP[Non-active players<br/>pushed AFTER]
    AP --> Stack[gs.Stack]
    NAP --> Stack
    Stack --> PriorityPriorityRound]
    Priority --> Resolve[ResolveStackTop<br/>NAP triggers resolve FIRST]
```

## APNAP Order Semantics

`OrderTriggersAPNAP` returns the PUSH order. Element [0] pushed first → resolves last. So the LAST player in APNAP order has triggers on top, resolving first. Counterintuitive but correct per §603.3b.

Within each controller's group, [Hat](Hat%20AI%20System.md) picks intra-group order via `OrderTriggers`.

## Trigger Guard

`per_card/registry.go:fireTrigger` enforces:
- **Per-chain depth: 8** — prevents recursive trigger storms
- **Total per game: 2000** — cumulative cap

Converged after the Freya bug audit that found infinite-loop trigger detection. Stable across 50K games.

## Event Aliases

`event_aliases.go` normalizes event names so per-card handlers don't have to know every spelling: `creature_died` → `dies`, `permanent_ltb` → `leaves_battlefield`, `card_discarded` → `discard`, etc. Added after the trigger audit found 8 dead per-card triggers because of name mismatches.

## Event Categories

- **Zone changes:** ETB, LTB, dies, exiled, discarded, milled, returned (see [Zone Changes](Zone%20Changes.md))
- **Cast:** spell_cast, noncreature_spell_cast, opponent_cast (Rhystic Study, Mystic Remora)
- **Combat:** attacks, blocks, deals_damage, deals_combat_damage_to_player
- **Phases:** upkeep, upkeep_controller (per-card, since 2026-04-26 fix), draw_step, end_step
- **Counters:** counter_added, counter_removed
- **Sacrifice:** typed by subject — `artifact_sacrificed`, `creature_sacrificed`, `food_sacrificed`

## Per-Card Hook Bridge

`per_card_hooks.go` exposes `TriggerHook` and `HasTriggerHook`. Per-card subpackage registers via `init()`. Engine calls `FireCardTrigger(name, ctx)` at every event site. See [Per-Card Handlers](Per-Card%20Handlers.md).

## Loop Shortcut Interplay

When a trigger pattern repeats (Kinnan-tap-untap, Ashling-counter), `loop_shortcut.go` (see [Stack and Priority](Stack%20and%20Priority.md)) fingerprints stack entries and projects deltas forward in one shot.

## Related

- [Stack and Priority](Stack%20and%20Priority.md)
- [Zone Changes](Zone%20Changes.md)
- [Per-Card Handlers](Per-Card%20Handlers.md)
- [Replacement Effects](Replacement%20Effects.md)
