# HexDek System Map

A package-level architecture diagram of the HexDek engine, AI, and platform.
All edges below were extracted from `import` blocks of the actual Go source
files under `internal/` and `cmd/hexdek-freya/`. Sub-modules listed inside
`internal/hat` (Amiibo, Tesla, Feynman, Ive, MicroNet, Octo, Poker, Neural)
are individual `.go` files in `package hat`, not separate packages — they are
shown as nodes for clarity.

## Mermaid graph

```mermaid
flowchart TB

    %% ================== ENGINE ==================
    subgraph Engine["⚔️ Engine — rules, cards, decks, tournaments"]
        direction TB
        gameast["<b>gameast</b><br/><i>parsed oracle text → AST</i><br/>CardAST · Ability · Effect · ManaCost · Trigger"]
        astload["<b>astload</b><br/><i>load AST corpus from JSONL</i><br/>Corpus · Load · LoadReader"]
        deckparser["<b>deckparser</b><br/><i>text decklist → engine cards</i><br/>TournamentDeck · ParseDeckFile · MetaDB · CloneLibrary"]
        gameengine["<b>gameengine</b><br/><i>MTG rules: zones, stack, combat, SBAs, triggers</i><br/>GameState · Card · Permanent · StackItem · Event<br/>NewGameState · CombatPhase · DrainStack · GainLife"]
        tournament["<b>tournament</b><br/><i>round-robin runner, ELO, mulligan, turn driver</i><br/>Run · TakeTurn · ELORatings · RunLondonMulligan"]
    end

    %% ================== INTELLIGENCE ==================
    subgraph Intelligence["🧠 Intelligence — the YggdrasilHat brain (package hat)"]
        direction TB

        subgraph IntelCore["Core — decision-making"]
            direction TB
            hat["<b>hat / Yggdrasil</b><br/><i>main AI player; MCTS + heuristics + Freya intel</i><br/>YggdrasilHat · NewYggdrasilHat · BudgetForELO · MjolnirStats"]
            strategy["<b>hat/strategy</b><br/><i>Freya strategy profile bridge</i><br/>StrategyProfile · WeaknessProfile · ComboPlan"]
            evaluator["<b>hat/evaluator</b><br/><i>8-dim board evaluation + weights</i>"]
            mcts["<b>hat/mcts + rollout</b><br/><i>tree search + simulated playouts</i>"]
        end

        subgraph IntelLearning["Learning — observation & evolution"]
            direction TB
            amiibo["<b>hat/amiibo</b><br/><i>evolvable per-deck DNA + dimension EMA</i><br/>AmiiboDNA · DimensionStats"]
            tesla["<b>hat/tesla</b><br/><i>causal pivot extraction (when did the game tilt?)</i><br/>CausalPivot"]
            feynman["<b>hat/feynman</b><br/><i>provably-correct invariant oracle</i><br/>OracleViolation · OracleResult"]
            ive["<b>hat/ive</b><br/><i>three-act spectator narrative builder</i><br/>GameNarrative · Act"]
            micronet["<b>hat/micronet</b><br/><i>per-deck tiny neural collector</i><br/>MicroNet · MicroSample · MicroCollector"]
            neural["<b>hat/neural</b><br/><i>state vector encoder for NN input</i><br/>StateVector · EncodeState"]
        end

        subgraph IntelPolicies["Policies & Analysis — alt brains and offline tools"]
            direction TB
            octo["<b>hat/octo</b><br/><i>OctoHat stress-test policy (chaos player)</i>"]
            poker["<b>hat/poker</b><br/><i>deprecated HOLD/CALL/RAISE policy</i>"]
            freya["<b>cmd/hexdek-freya</b><br/><i>deck analyzer: archetype, combos, roles, win lines</i><br/>FreyaReport · DeckProfile · BuildDeckProfile · ComputeEvalWeights"]
        end
    end

    %% ================== DISCOVERY ==================
    subgraph Discovery["🔭 Discovery — Heimdall sees, Huginn thinks, Muninn remembers"]
        direction TB
        heimdall["<b>heimdall</b><br/><i>game observer + replay system → Observations</i><br/>ReplayContext · NewReplayContext · ReadSeeds · ReplayWithObservation"]
        huginn["<b>huginn</b><br/><i>tier-graduates co-trigger patterns into combos</i><br/>FreyaInteraction · FreyaChain · Tier3Export"]
        muninn["<b>muninn</b><br/><i>persistent failure memory: gaps, crashes, dead triggers, invariants</i><br/>PersistParserGaps · PersistCrashLogs · PersistConcessions<br/>PersistInvariantViolations · AutoArchiveViolation"]
    end

    %% ================== PLATFORM ==================
    subgraph Platform["🌐 Platform — API, telemetry, post-game analytics"]
        direction TB
        analytics["<b>analytics</b><br/><i>post-game reports: combos, finishers, stalls, rivalries, threat graph</i><br/>GameAnalysis · AnalyzeGame · DetectMissedCombos · CardPerformance"]
        telemetry["<b>telemetry</b><br/><i>GA4 client (no internal deps)</i><br/>GA4Client · NewGA4Client"]
        hexapi["<b>hexapi</b><br/><i>HTTP/WS surface: handlers, showmatch, adapters</i><br/>Handler · DeckSummary · ImportRequest"]
    end

    %% ============ ENGINE INTERNAL FLOW ============
    gameast --> astload
    gameast --> gameengine
    astload --> gameengine
    gameast --> deckparser
    astload --> deckparser
    gameengine --> deckparser

    gameast --> tournament
    astload --> tournament
    deckparser --> tournament
    gameengine --> tournament

    %% ============ INTELLIGENCE INTERNAL ============
    strategy --> hat
    evaluator --> hat
    mcts --> hat
    amiibo --> hat
    tesla --> hat
    feynman --> hat
    ive --> hat
    micronet --> hat
    neural --> hat
    octo --> hat
    poker --> hat

    %% ============ ENGINE → INTELLIGENCE ============
    gameast --> hat
    gameengine --> hat
    gameengine --> analytics

    %% ============ DISCOVERY EDGES ============
    astload --> heimdall
    deckparser --> heimdall
    gameengine --> heimdall
    hat --> heimdall
    tournament --> heimdall

    analytics --> huginn
    analytics --> muninn

    %% ============ INTELLIGENCE FEEDBACK LOOPS ============
    huginn -. "tier3_for_freya.json" .-> freya
    freya -. "StrategyProfile JSON" .-> strategy
    heimdall -. "Observations" .-> huginn
    heimdall -. "Observations" .-> muninn
    tesla -. "pivot labels" .-> micronet
    feynman -. "invariant fails" .-> muninn

    %% ============ TOURNAMENT WIRING ============
    hat --> tournament
    huginn --> tournament
    muninn --> tournament
    analytics --> tournament

    %% ============ PLATFORM EDGES ============
    astload --> hexapi
    deckparser --> hexapi
    gameengine --> hexapi
    hat --> hexapi
    heimdall --> hexapi
    huginn --> hexapi
    muninn --> hexapi
    tournament --> hexapi
    analytics --> hexapi
    telemetry --> hexapi

    %% ============ STYLES ============
    classDef engine fill:#1e3a5f,stroke:#5b9bd5,color:#fff
    classDef intel fill:#3d2a5f,stroke:#a78bfa,color:#fff
    classDef discovery fill:#1f4d2a,stroke:#5cb85c,color:#fff
    classDef platform fill:#5f3a1e,stroke:#f0ad4e,color:#fff

    class gameast,astload,deckparser,gameengine,tournament engine
    class hat,strategy,evaluator,mcts,amiibo,tesla,feynman,ive,micronet,neural,octo,poker,freya intel
    class heimdall,huginn,muninn discovery
    class analytics,telemetry,hexapi platform
```

## Package summary

### Engine

| Package | One-liner | Imports (internal) |
|---|---|---|
| `gameast` | Parsed oracle text → structured AST (abilities, effects, mana). | — |
| `astload` | Loads JSONL AST corpus into memory. | `gameast` |
| `deckparser` | Decklist text → playable `gameengine.Card` slices, with metadata. | `astload`, `gameast`, `gameengine` |
| `gameengine` | MTG rules engine: zones, stack, combat, SBAs, triggers, mana. | `astload`, `gameast` |
| `tournament` | Tournament runner, ELO, mulligan, turn driver, round-robin/pool. | `analytics`, `astload`, `deckparser`, `gameast`, `gameengine`, `gameengine/per_card`, `hat`, `huginn`, `muninn`, `trueskill` |

### Intelligence (all sub-modules live in `package hat` except Freya)

| Module | One-liner |
|---|---|
| `hat / Yggdrasil` | Main AI player. Layered: heuristic → evaluator → MCTS by budget. |
| `hat/strategy` | `StrategyProfile` — Freya analysis injected into hat decisions. |
| `hat/evaluator` | 8-dim board scoring with archetype-aware weights. |
| `hat/mcts` | Monte Carlo tree search + rollout policies. |
| `hat/amiibo` | Per-deck evolvable DNA; EMA correlations between dim scores and wins. |
| `hat/tesla` | Causal pivot extraction — *when* did the game tilt? |
| `hat/feynman` | Provably-correct invariant oracle (engine-bug detector). |
| `hat/ive` | Three-act narrative builder for spectator UI. |
| `hat/micronet` | Per-deck tiny neural sample collector. |
| `hat/neural` | `StateVector` encoder for NN input. |
| `hat/octo` | OctoHat — chaos stress-test policy. |
| `hat/poker` | Deprecated HOLD/CALL/RAISE adaptive hat. |
| `cmd/hexdek-freya` | Deck analyzer: archetype, combos, roles, win lines, mana base, opening hands. Emits `StrategyProfile` JSON the hat consumes. |

`hat` itself imports only `gameast` and `gameengine`. Freya (`cmd/hexdek-freya`) imports `huginn` to fold tier-3 confirmed combos into `KnownCombos`.

### Discovery (Norse mythology: Heimdall watches, Huginn thinks, Muninn remembers)

| Package | One-liner | Imports (internal) |
|---|---|---|
| `heimdall` | Game observer + deterministic replay system; emits `Observation` records to sinks. | `astload`, `deckparser`, `gameengine`, `hat`, `tournament` |
| `huginn` | Tier-graduates co-trigger pairs (OBSERVED → RECURRING → CONFIRMED) and chains them into N-card lines. | `analytics` |
| `muninn` | Persistent failure memory: parser gaps, crashes, dead triggers, concessions, invariant violations, regressions. | `analytics` |

### Platform

| Package | One-liner | Imports (internal) |
|---|---|---|
| `analytics` | Post-game reports: missed combos/finishers, stall detection, rivalries, threat graph, card rankings. | `gameengine` |
| `telemetry` | GA4 client. No internal dependencies. | — |
| `hexapi` | HTTP/WebSocket surface; bridges every other package to clients. | `analytics`, `astload`, `db`, `deckparser`, `gameengine`, `hat`, `heimdall`, `huginn`, `matchmaking`, `muninn`, `telemetry`, `tournament`, `trueskill`, `versioning` |

## Data flow at a glance

```
Scryfall → thor → ast_dataset.jsonl
                     │
                     ▼
                 astload ──► gameast
                     │           │
                     ▼           ▼
                 deckparser ──► gameengine ──► hat (Yggdrasil)
                                    │              ▲
                                    ▼              │
                                tournament ────────┘
                                    │
                                    ▼
                                analytics
                                ╱   │   ╲
                              ▼     ▼    ▼
                         heimdall huginn muninn
                              │     │     │
                              │     ▼     │
                              │   freya   │
                              │     │     │
                              ▼     ▼     ▼
                                 hexapi  ──► clients (web, spectator)
```

The intelligence loop is deliberately bidirectional: tournament games feed
analytics → huginn graduates patterns → freya consumes tier-3 combos →
freya emits a `StrategyProfile` → hat reads that profile when playing the
next round of tournament games.
