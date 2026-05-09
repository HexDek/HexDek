package hat

import (
	"testing"

	"github.com/hexdek/hexdek/internal/gameengine"
)

// dev/order-replacements — verifies YggdrasilHat.OrderReplacements
// promotes self-controlled and self-beneficial replacements ahead of
// opponent and neutral effects.

func TestYggdrasil_OrderReplacements_SelfBeforeOpponent(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)

	opp := &gameengine.ReplacementEffect{
		Timestamp:      1,
		ControllerSeat: 1,
		EventType:      "would_draw",
	}
	mine := &gameengine.ReplacementEffect{
		Timestamp:      5,
		ControllerSeat: 0,
		EventType:      "would_draw",
	}

	got := h.OrderReplacements(gs, 0, []*gameengine.ReplacementEffect{opp, mine})
	if len(got) != 2 || got[0] != mine || got[1] != opp {
		t.Fatalf("self-controlled replacement should sort first; got %v", got)
	}
}

func TestYggdrasil_OrderReplacements_LowLifePullsDamageReplacementToTop(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	gs.Seats[0].Life = 3 // critical

	gainSelf := &gameengine.ReplacementEffect{
		Timestamp:      1,
		ControllerSeat: 0,
		EventType:      "would_gain_life",
	}
	preventSelf := &gameengine.ReplacementEffect{
		Timestamp:      9,
		ControllerSeat: 0,
		EventType:      "would_be_dealt_damage",
	}

	got := h.OrderReplacements(gs, 0, []*gameengine.ReplacementEffect{gainSelf, preventSelf})
	if len(got) != 2 || got[0] != preventSelf {
		t.Fatalf("at low life, damage-prevention replacement must sort to the top; got %v", got)
	}
}

func TestYggdrasil_OrderReplacements_SingleCandidatePassthrough(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	r := &gameengine.ReplacementEffect{Timestamp: 1, ControllerSeat: 0, EventType: "would_draw"}
	got := h.OrderReplacements(gs, 0, []*gameengine.ReplacementEffect{r})
	if len(got) != 1 || got[0] != r {
		t.Fatalf("single-candidate input should pass through unchanged")
	}
}

func TestYggdrasil_OrderReplacements_TimestampBreaksTies(t *testing.T) {
	h := NewYggdrasilHatWithNoise(nil, 0, 0)
	gs := newTestGame(t, 2)
	older := &gameengine.ReplacementEffect{Timestamp: 1, ControllerSeat: 1, EventType: "would_draw"}
	newer := &gameengine.ReplacementEffect{Timestamp: 5, ControllerSeat: 1, EventType: "would_draw"}
	got := h.OrderReplacements(gs, 0, []*gameengine.ReplacementEffect{newer, older})
	if got[0] != older || got[1] != newer {
		t.Fatalf("timestamp ascending should break ties; got %v", got)
	}
}
