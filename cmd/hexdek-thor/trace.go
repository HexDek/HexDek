package main

// trace.go — opt-in execution trace recorder for Thor tests.
//
// When --trace is passed, each test is given a Tracer that captures
// significant events (setup, trigger checks, handler entry, effect
// attempts, state changes, assertions). On failure, the trace is
// flushed to data/thor-traces/{slug}.{kind}.trace so we can see WHERE
// the chain broke instead of just the final assertion.
//
// When --trace is off, NewTracer returns nil and every method becomes
// a nil-receiver no-op, preserving zero overhead.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

// Package-level config. Set once in main() during flag parsing,
// read concurrently by workers without locks.
var (
	traceEnabled bool
	traceDir     string
	// traceSeq disambiguates trace files when the same card+kind fails
	// more than once across the run (different interactions, retries).
	traceSeq uint64
)

// Tracer accumulates step-numbered entries for a single test execution.
// A nil *Tracer is valid: every method becomes a no-op.
type Tracer struct {
	step    int
	entries []string
}

// NewTracer returns a fresh tracer if tracing is globally enabled,
// otherwise nil (callers must handle nil — methods on *Tracer do).
func NewTracer() *Tracer {
	if !traceEnabled {
		return nil
	}
	return &Tracer{entries: make([]string, 0, 32)}
}

// Record appends a step-numbered entry of the given kind.
// kind is the all-caps event class (SETUP, TRIGGER_CHECK, HANDLER_ENTER,
// EFFECT_ATTEMPT, STATE_CHANGE, ASSERT_PASS, ASSERT_FAIL, etc).
func (t *Tracer) Record(kind, format string, args ...interface{}) {
	if t == nil {
		return
	}
	t.step++
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	t.entries = append(t.entries, fmt.Sprintf("[%03d] %s: %s", t.step, kind, msg))
}

// Flush writes the trace to data/thor-traces/{slug}.{kind}.trace.
// Safe to call with nil receiver. Ignored when trace dir is unset
// (e.g. --trace was off).
func (t *Tracer) Flush(cardName, failureKind string) {
	if t == nil || len(t.entries) == 0 || traceDir == "" {
		return
	}
	seq := atomic.AddUint64(&traceSeq, 1)
	name := fmt.Sprintf("%s.%s.%06d.trace", slugify(cardName), slugify(failureKind), seq)
	path := filepath.Join(traceDir, name)
	body := strings.Join(t.entries, "\n") + "\n"
	// Best-effort: log a warning but don't fail the test if the
	// filesystem hiccups — traces are diagnostic, not load-bearing.
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "thor: trace write failed for %s: %v\n", path, err)
	}
}

// slugify lowercases a string and replaces any non-alphanumeric runs
// with a single hyphen, suitable for use as a filename component.
func slugify(s string) string {
	if s == "" {
		return "unknown"
	}
	var b strings.Builder
	b.Grow(len(s))
	prevHyphen := false
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen && b.Len() > 0 {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	out := strings.TrimRight(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}
