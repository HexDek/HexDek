package per_card

import (
	"reflect"
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// Regression tests for the dev/muninn-handlers-201-220 wave: dispatch-layer
// fallback so existing handlers fire on cascade/copy/token-renamed variants
// (cascade.go renames StackItem.Card.Name to "X (cascade)", Urza copies
// become "X (Urza copy)", Miirym tokens become "X (Miirym Token)", etc.).
// Before the fix, those variant names silently bypassed every fire*
// dispatcher except the partial " // " fallback in fireETB/fireTrigger.

func TestLookupCandidates_StripsTrailingParenthetical(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"Necromancy (cascade)", []string{"necromancy (cascade)", "necromancy"}},
		{"Phoenix Fleet Airship (Urza copy)", []string{"phoenix fleet airship (urza copy)", "phoenix fleet airship"}},
		{"Tiamat (Miirym Token)", []string{"tiamat (miirym token)", "tiamat"}},
		{"Claim Jumper (Restore-Relic token)", []string{"claim jumper (restorerelic token)", "claim jumper"}},
		// Nested suffix: strips only the outermost parenthetical (good enough
		// to reach the next-level variant, which itself can be stripped on
		// a subsequent lookup if needed).
		{"Crown of Gondor (Urza copy) (Urza copy)", []string{"crown of gondor (urza copy) (urza copy)", "crown of gondor (urza copy)"}},
		// Pure base name yields a single key.
		{"Lightning Bolt", []string{"lightning bolt"}},
		// DFC alone keeps the historical " // " front-face fallback.
		// normalizeName preserves "//" (only punctuation in []rune is stripped).
		{"Curious Homunculus // Voracious Reader", []string{"curious homunculus // voracious reader", "curious homunculus"}},
		// Cascade-fired DFC: strip paren, then split DFC.
		{"Eccentric Pestfinder // Turn Stones (cascade)", []string{"eccentric pestfinder // turn stones (cascade)", "eccentric pestfinder // turn stones", "eccentric pestfinder"}},
	}
	for _, tc := range cases {
		got := lookupCandidates(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("lookupCandidates(%q):\n  got  %q\n  want %q", tc.in, got, tc.want)
		}
	}
}

// Cascade-renamed ETB must reach the front-face handler.
func TestFireETB_DispatchesThroughCascadeRename(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	fired := 0
	Global().OnETB("Wave201CascadeETB", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		fired++
	})

	gs := newGame(t, 2)
	perm := addPerm(gs, 0, "Wave201CascadeETB (cascade)", "creature")

	gameengine.InvokeETBHook(gs, perm)

	if fired != 1 {
		t.Errorf("expected cascade-renamed ETB to dispatch via fallback, fired=%d", fired)
	}
}

// Urza copy must reach the front-face handler.
func TestFireETB_DispatchesThroughUrzaCopyRename(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	fired := 0
	Global().OnETB("Wave201UrzaCopy", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		fired++
	})

	gs := newGame(t, 2)
	perm := addPerm(gs, 0, "Wave201UrzaCopy (Urza copy)", "artifact")

	gameengine.InvokeETBHook(gs, perm)

	if fired != 1 {
		t.Errorf("expected Urza-copy ETB to dispatch via fallback, fired=%d", fired)
	}
}

// Cascade-renamed Trigger must reach the front-face handler.
func TestFireTrigger_DispatchesThroughCascadeRename(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	fired := 0
	Global().OnTrigger("Wave201CascadeTrigger", "end_step", func(gs *gameengine.GameState, perm *gameengine.Permanent, ctx map[string]interface{}) {
		fired++
	})

	gs := newGame(t, 2)
	addPerm(gs, 0, "Wave201CascadeTrigger (cascade)", "enchantment")

	gameengine.FireCardTrigger(gs, "end_step", map[string]interface{}{
		"controller_seat": 0,
	})

	if fired != 1 {
		t.Errorf("expected cascade-renamed trigger to dispatch via fallback, fired=%d", fired)
	}
}

// Pure base-name lookups still hit the direct key (no fallback regression).
func TestFireETB_DirectNameStillDispatches(t *testing.T) {
	Reset()
	t.Cleanup(Reset)

	fired := 0
	Global().OnETB("Wave201Direct", func(gs *gameengine.GameState, perm *gameengine.Permanent) {
		fired++
	})

	gs := newGame(t, 2)
	perm := addPerm(gs, 0, "Wave201Direct", "creature")

	gameengine.InvokeETBHook(gs, perm)

	if fired != 1 {
		t.Errorf("expected direct-name ETB to dispatch, fired=%d", fired)
	}
}
