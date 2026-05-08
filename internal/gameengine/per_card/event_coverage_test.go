package per_card

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// TestAllRegisteredTriggersAreDispatched is a CI lint test that verifies every
// canonical event name used in OnTrigger registrations has at least one
// corresponding FireCardTrigger dispatch point in the engine source. If this
// test fails, either:
//   - a new per_card handler registered for an event that the engine never
//     fires (dead trigger), or
//   - a FireCardTrigger call was removed from the engine without removing
//     the handlers that depend on it.
func TestAllRegisteredTriggersAreDispatched(t *testing.T) {
	Reset()
	reg := Global()
	registered := reg.RegisteredTriggerEvents()

	dispatched := scanDispatchedEvents(t)

	// Build the set of canonical dispatched events (normalize each one).
	canonicalDispatched := map[string]bool{}
	for ev := range dispatched {
		canonical := gameengine.NormalizeEventSingle(ev)
		canonicalDispatched[canonical] = true
	}

	// Events dispatched through dedicated hooks (not FireCardTrigger).
	specialPaths := map[string]bool{
		"etb":       true, // fireETB hook
		"on_cast":   true, // fireOnCast hook
		"on_resolve": true, // fireOnResolve hook
		"activated":  true, // fireActivated hook
	}

	// Events that are catch-all parser fallbacks or meta-events, not
	// directly dispatched but matched via compound aliases.
	metaEvents := map[string]bool{
		"generic":       true,
		"generic_you":   true,
		"generic_self":  true,
		"generic_typed": true,
		"state_check":   true,
		"eot_trigger":   true,
		"cascading":     true,
		"damage_taken":  true, // subsumed by deals_damage / deals_combat_damage
	}

	var missing []string
	for ev := range registered {
		if specialPaths[ev] || metaEvents[ev] {
			continue
		}
		if !canonicalDispatched[ev] {
			missing = append(missing, ev)
		}
	}
	if len(missing) > 0 {
		t.Errorf("registered trigger events with no dispatch:\n  %s\n"+
			"Either add FireCardTrigger(gs, <event>, ...) in the engine, "+
			"add an alias in event_aliases.go, or add to the specialPaths/metaEvents allowlist.",
			strings.Join(missing, "\n  "))
	}
}

var fireCardTriggerRe = regexp.MustCompile(`FireCardTrigger\(gs,\s*"([^"]+)"`)

func scanDispatchedEvents(t *testing.T) map[string]bool {
	t.Helper()
	events := map[string]bool{}

	roots := []string{
		filepath.Join("..", "..", "..", "internal", "gameengine"),
		filepath.Join("..", "..", "..", "internal", "tournament"),
	}
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
				return nil
			}
			// Skip per_card handlers (they're consumers, not dispatchers).
			if strings.Contains(path, "per_card") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, match := range fireCardTriggerRe.FindAllSubmatch(data, -1) {
				events[string(match[1])] = true
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scanning %s: %v", root, err)
		}
	}
	return events
}
