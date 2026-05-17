package hat

import (
	"testing"
	"time"
)

func TestConvictionRing_SnapshotOrderAndWrap(t *testing.T) {
	ResetConvictionTelemetry()
	t.Cleanup(ResetConvictionTelemetry)

	// Fill exactly capacity — no wrap yet.
	for i := 1; i <= convictionRingCapacity; i++ {
		pushConvictionEvent(ConvictionEvent{Turn: i})
	}
	events, total := SnapshotConvictionEvents(0, 0)
	if total != uint64(convictionRingCapacity) {
		t.Fatalf("total_seen = %d, want %d", total, convictionRingCapacity)
	}
	if len(events) != convictionRingCapacity {
		t.Fatalf("len(events) = %d, want %d", len(events), convictionRingCapacity)
	}
	if events[0].Turn != 1 || events[len(events)-1].Turn != convictionRingCapacity {
		t.Fatalf("order broken: first=%d last=%d", events[0].Turn, events[len(events)-1].Turn)
	}
	if events[0].Sequence != 1 || events[len(events)-1].Sequence != uint64(convictionRingCapacity) {
		t.Fatalf("sequence not monotonic from 1: first=%d last=%d",
			events[0].Sequence, events[len(events)-1].Sequence)
	}

	// Wrap by half a buffer — oldest entries should be evicted.
	for i := convictionRingCapacity + 1; i <= convictionRingCapacity+convictionRingCapacity/2; i++ {
		pushConvictionEvent(ConvictionEvent{Turn: i})
	}
	events, total = SnapshotConvictionEvents(0, 0)
	wantTotal := uint64(convictionRingCapacity + convictionRingCapacity/2)
	if total != wantTotal {
		t.Fatalf("total_seen after wrap = %d, want %d", total, wantTotal)
	}
	if len(events) != convictionRingCapacity {
		t.Fatalf("len(events) after wrap = %d, want %d (buffer cap)", len(events), convictionRingCapacity)
	}
	// After wrapping by N/2, the oldest retained turn should be N/2+1.
	if got := events[0].Turn; got != convictionRingCapacity/2+1 {
		t.Fatalf("oldest retained turn = %d, want %d", got, convictionRingCapacity/2+1)
	}
	if got := events[len(events)-1].Turn; got != int(wantTotal) {
		t.Fatalf("newest retained turn = %d, want %d", got, wantTotal)
	}
}

func TestConvictionRing_SinceAndLimit(t *testing.T) {
	ResetConvictionTelemetry()
	t.Cleanup(ResetConvictionTelemetry)

	for i := 1; i <= 10; i++ {
		pushConvictionEvent(ConvictionEvent{Turn: i})
	}
	events, _ := SnapshotConvictionEvents(7, 0)
	if len(events) != 3 {
		t.Fatalf("since=7 returned %d events, want 3", len(events))
	}
	if events[0].Sequence != 8 || events[2].Sequence != 10 {
		t.Fatalf("since=7 sequences = [%d..%d], want [8..10]", events[0].Sequence, events[2].Sequence)
	}

	events, _ = SnapshotConvictionEvents(0, 4)
	if len(events) != 4 {
		t.Fatalf("limit=4 returned %d events", len(events))
	}
	if events[0].Sequence != 7 || events[3].Sequence != 10 {
		t.Fatalf("limit=4 sequences = [%d..%d], want [7..10]", events[0].Sequence, events[3].Sequence)
	}
}

func TestConvictionRing_DisableSkipsCapture(t *testing.T) {
	ResetConvictionTelemetry()
	t.Cleanup(func() {
		SetConvictionTelemetryEnabled(true)
		ResetConvictionTelemetry()
	})

	SetConvictionTelemetryEnabled(false)
	pushConvictionEvent(ConvictionEvent{Turn: 1})
	events, total := SnapshotConvictionEvents(0, 0)
	if len(events) != 0 || total != 0 {
		t.Fatalf("disabled: events=%d total=%d, want 0/0", len(events), total)
	}

	SetConvictionTelemetryEnabled(true)
	pushConvictionEvent(ConvictionEvent{Turn: 2})
	events, total = SnapshotConvictionEvents(0, 0)
	if len(events) != 1 || total != 1 {
		t.Fatalf("re-enabled: events=%d total=%d, want 1/1", len(events), total)
	}
}

func TestConvictionRing_CapturedAtPopulated(t *testing.T) {
	ResetConvictionTelemetry()
	t.Cleanup(ResetConvictionTelemetry)

	before := time.Now().UTC().Add(-time.Second)
	pushConvictionEvent(ConvictionEvent{Turn: 1})
	after := time.Now().UTC().Add(time.Second)

	events, _ := SnapshotConvictionEvents(0, 0)
	if len(events) != 1 {
		t.Fatalf("len(events) = %d", len(events))
	}
	if events[0].CapturedAt.Before(before) || events[0].CapturedAt.After(after) {
		t.Fatalf("CapturedAt %v outside [%v, %v]", events[0].CapturedAt, before, after)
	}
}
