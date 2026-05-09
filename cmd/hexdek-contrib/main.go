// hexdek-contrib — community distributed-compute client.
//
// Connects to a HexDek server via WebSocket, receives signed work
// chunks (deck matchups to simulate), runs them locally with the same
// engine the server uses, and returns signed results. Validated chunks
// earn credits visible on the operator profile.
//
// Usage:
//
//	hexdek-contrib \
//	    --server wss://hexdek.dev \
//	    --token  $HEXDEK_CONTRIB_TOKEN \
//	    --ast    data/rules/ast_dataset.jsonl \
//	    [--workers 0]                 # 0 = NumCPU; per-chunk parallelism
//	    [--max-chunks 0]              # 0 = run forever
//
// The client is intentionally lightweight: a single Go binary plus the
// ast_dataset.jsonl corpus the engine uses to parse oracle text. No
// disk writes, no telemetry, no network beyond the single WS to the
// server. Authentication is via a bearer token issued by the server
// (set HEXDEK_CONTRIB_TOKEN or pass --token).
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/hexdek/hexdek/internal/astload"
	"github.com/hexdek/hexdek/internal/deckparser"
	"github.com/hexdek/hexdek/internal/hexapi"
	"github.com/hexdek/hexdek/internal/tournament"
)

const workerVersion = "hexdek-contrib/1"

func main() {
	var (
		serverURL  = flag.String("server", "wss://hexdek.dev", "HexDek server URL (ws:// or wss://)")
		token      = flag.String("token", os.Getenv("HEXDEK_CONTRIB_TOKEN"), "bearer token; default $HEXDEK_CONTRIB_TOKEN")
		astPath    = flag.String("ast", "data/rules/ast_dataset.jsonl", "AST dataset path")
		oraclePath = flag.String("oracle", "", "optional oracle-cards.json for P/T supplement")
		workers    = flag.Int("workers", 0, "per-chunk parallelism (0 = NumCPU)")
		maxChunks  = flag.Int("max-chunks", 0, "stop after N chunks (0 = run forever)")
		quiet      = flag.Bool("quiet", false, "minimize log output")
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("--token (or $HEXDEK_CONTRIB_TOKEN) is required")
	}

	logger := log.New(os.Stderr, "contrib: ", log.LstdFlags)
	if *quiet {
		logger.SetFlags(0)
	}

	// Load engine corpus.
	logger.Printf("loading AST corpus %s ...", *astPath)
	corpus, err := astload.Load(*astPath)
	if err != nil {
		log.Fatalf("load AST corpus: %v", err)
	}
	meta, err := deckparser.LoadMetaFromJSONL(*astPath)
	if err != nil {
		log.Fatalf("load meta: %v", err)
	}
	if *oraclePath != "" {
		if err := meta.SupplementWithOracleJSON(*oraclePath); err != nil {
			logger.Printf("oracle supplement: %v (continuing)", err)
		}
	}
	logger.Printf("ready: %d cards in corpus, %d in meta, %d workers", corpus.Count(), meta.Count(), poolSize(*workers))

	// Build the WS URL.
	connectURL, err := buildConnectURL(*serverURL, *token)
	if err != nil {
		log.Fatalf("build URL: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Printf("dialing %s", redactToken(connectURL))
	wsConn, _, err := websocket.Dial(ctx, connectURL, nil)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer wsConn.Close(websocket.StatusNormalClosure, "client exit")
	wsConn.SetReadLimit(1 << 20)

	// Welcome.
	welcome, err := readEnvelope(ctx, wsConn)
	if err != nil {
		log.Fatalf("read welcome: %v", err)
	}
	if welcome.Type != "welcome" {
		log.Fatalf("expected welcome, got %q", welcome.Type)
	}
	var w hexapi.ContribWelcome
	if err := json.Unmarshal(welcome.Payload, &w); err != nil {
		log.Fatalf("decode welcome: %v", err)
	}
	if w.ProtocolVer != hexapi.ProtocolVersion {
		log.Fatalf("protocol mismatch: server=%d client=%d", w.ProtocolVer, hexapi.ProtocolVersion)
	}
	hmacKey, err := hex.DecodeString(w.HMACKey)
	if err != nil {
		log.Fatalf("decode hmac key: %v", err)
	}
	logger.Printf("connected as %s (server time %d, %d workers)", w.Owner, w.ServerTime, poolSize(*workers))

	// Main pull-run-submit loop.
	completed := 0
	for {
		if *maxChunks > 0 && completed >= *maxChunks {
			logger.Printf("reached --max-chunks=%d, exiting", *maxChunks)
			return
		}
		// Announce ready.
		ready := hexapi.ContribReady{
			WorkerVersion: workerVersion,
			NumCPU:        runtime.NumCPU(),
			OS:            runtime.GOOS,
			Arch:          runtime.GOARCH,
		}
		if err := writeEnvelope(ctx, wsConn, "ready", ready); err != nil {
			log.Fatalf("write ready: %v", err)
		}

		env, err := readEnvelope(ctx, wsConn)
		if err != nil {
			log.Fatalf("read: %v", err)
		}
		switch env.Type {
		case "wait":
			var msg struct {
				RetryAfterSeconds int `json:"retry_after_seconds"`
			}
			_ = json.Unmarshal(env.Payload, &msg)
			d := time.Duration(msg.RetryAfterSeconds) * time.Second
			if d <= 0 {
				d = 10 * time.Second
			}
			logger.Printf("queue empty, waiting %v", d)
			time.Sleep(d)
			continue

		case "assignment":
			var a hexapi.ContribAssignment
			if err := json.Unmarshal(env.Payload, &a); err != nil {
				logger.Printf("decode assignment: %v (skipping)", err)
				continue
			}
			if err := hexapi.VerifyAssignment(&a, hmacKey); err != nil {
				logger.Printf("REJECTED assignment %s: %v", a.ChunkID, err)
				continue
			}
			logger.Printf("accepted chunk %s: %d games × %d seats (seed=%d)", a.ChunkID, a.GamesCount, a.NSeats, a.Seed)

			result, err := runAssignment(&a, corpus, meta, *workers)
			if err != nil {
				logger.Printf("run failed for %s: %v", a.ChunkID, err)
				// We still need to clear the in-flight slot on the
				// server. Send a stub result with empty winners so
				// the server rejects it cleanly.
				stub := &hexapi.ContribResult{
					ChunkID:       a.ChunkID,
					StartedAt:     time.Now().Unix(),
					FinishedAt:    time.Now().Unix(),
					WorkerVersion: workerVersion,
				}
				_, _ = hexapi.SignResult(stub, hmacKey)
				_ = writeEnvelope(ctx, wsConn, "result", stub)
				_, _ = readEnvelope(ctx, wsConn) // drain ack
				continue
			}
			if _, err := hexapi.SignResult(result, hmacKey); err != nil {
				log.Fatalf("sign result: %v", err)
			}
			if err := writeEnvelope(ctx, wsConn, "result", result); err != nil {
				log.Fatalf("write result: %v", err)
			}
			ack, err := readEnvelope(ctx, wsConn)
			if err != nil {
				log.Fatalf("read ack: %v", err)
			}
			var a2 hexapi.ContribAck
			_ = json.Unmarshal(ack.Payload, &a2)
			if a2.Accepted {
				logger.Printf("ack chunk=%s accepted (+%d credits, z=%.2f, spot_check=%v)",
					a2.ChunkID, a2.CreditsAwarded, a2.AnomalyZScore, a2.SpotChecked)
			} else {
				logger.Printf("ack chunk=%s REJECTED: %s", a2.ChunkID, a2.Reason)
			}
			completed++

		case "error":
			logger.Printf("server error: %s", string(env.Payload))

		default:
			logger.Printf("unexpected envelope type %q (continuing)", env.Type)
		}
	}
}

// runAssignment parses the chunk's decks, runs the games via
// tournament.RunChunk, and packages the winners into a ContribResult.
func runAssignment(a *hexapi.ContribAssignment, corpus *astload.Corpus, meta *deckparser.MetaDB, workers int) (*hexapi.ContribResult, error) {
	if len(a.Decks) != a.NSeats {
		return nil, fmt.Errorf("assignment has %d decks but n_seats=%d", len(a.Decks), a.NSeats)
	}
	decks := make([]*deckparser.TournamentDeck, a.NSeats)
	for i, raw := range a.Decks {
		d, err := deckparser.ParseDeckReader(strings.NewReader(raw), corpus, meta)
		if err != nil {
			return nil, fmt.Errorf("deck %d parse: %w", i, err)
		}
		decks[i] = d
	}
	cfg := tournament.ChunkConfig{
		Decks:           decks,
		NSeats:          a.NSeats,
		NGames:          a.GamesCount,
		Seed:            a.Seed,
		MaxTurnsPerGame: a.MaxTurns,
		CommanderMode:   true,
		Workers:         workers,
	}
	startedAt := time.Now()
	outcomes, err := tournament.RunChunk(cfg)
	if err != nil {
		return nil, err
	}
	finishedAt := time.Now()
	winners := make([]int, len(outcomes))
	turns := make([]int, len(outcomes))
	for i, o := range outcomes {
		winners[i] = o.Winner
		turns[i] = o.Turns
	}
	return &hexapi.ContribResult{
		ChunkID:       a.ChunkID,
		StartedAt:     startedAt.Unix(),
		FinishedAt:    finishedAt.Unix(),
		ElapsedMS:     finishedAt.Sub(startedAt).Milliseconds(),
		Winners:       winners,
		TurnCounts:    turns,
		OutcomeHash:   hexapi.HashOutcomes(winners, turns),
		WorkerVersion: workerVersion,
	}, nil
}

// --- WS framing helpers ----------------------------------------------

func readEnvelope(ctx context.Context, c *websocket.Conn) (*hexapi.ContribEnvelope, error) {
	_, data, err := c.Read(ctx)
	if err != nil {
		return nil, err
	}
	var env hexapi.ContribEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}
	return &env, nil
}

func writeEnvelope(ctx context.Context, c *websocket.Conn, kind string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := hexapi.ContribEnvelope{Type: kind, Payload: json.RawMessage(body)}
	out, err := json.Marshal(env)
	if err != nil {
		return err
	}
	wctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return c.Write(wctx, websocket.MessageText, out)
}

// buildConnectURL turns a base server URL + token into the actual
// /api/contrib/connect WebSocket URL. Accepts both ws/wss and http/https
// (which we rewrite to ws/wss).
func buildConnectURL(server, token string) (string, error) {
	u, err := url.Parse(server)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		// fine
	default:
		return "", fmt.Errorf("unsupported scheme %q (want ws/wss/http/https)", u.Scheme)
	}
	u.Path = "/api/contrib/connect"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func redactToken(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return u
	}
	q := parsed.Query()
	if q.Get("token") != "" {
		q.Set("token", "REDACTED")
		parsed.RawQuery = q.Encode()
	}
	return parsed.String()
}

func poolSize(w int) int {
	if w <= 0 {
		return runtime.NumCPU()
	}
	return w
}
