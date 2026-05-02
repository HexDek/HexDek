package hexapi

import (
	"github.com/hexdek/hexdek/internal/heimdall"
	"github.com/hexdek/hexdek/internal/muninn"
)

// muninnAdapter implements heimdall.MuninnSink by delegating to the real
// muninn persist functions. This is a thin bridge between Heimdall's
// per-observation interface and Muninn's append-only JSON files.
type muninnAdapter struct {
	dataDir string
}

var _ heimdall.MuninnSink = (*muninnAdapter)(nil)

// RecordParserGaps converts the per-game gap list into Muninn's
// map[string]int format (each gap counted once) and persists.
func (a *muninnAdapter) RecordParserGaps(gaps []string, gameID string) {
	if len(gaps) == 0 {
		return
	}
	counts := make(map[string]int, len(gaps))
	for _, g := range gaps {
		counts[g]++
	}
	_ = muninn.PersistParserGaps(a.dataDir, counts)
}

// RecordDeadTriggers converts Heimdall's lightweight DeadTrigger into
// Muninn's persistent format and merges into dead_triggers.json.
func (a *muninnAdapter) RecordDeadTriggers(triggers []heimdall.DeadTrigger, gameID string) {
	if len(triggers) == 0 {
		return
	}
	existing, err := muninn.ReadDeadTriggers(a.dataDir)
	if err != nil {
		return
	}

	// Index existing by trigger_name + card_name for O(1) merge.
	type key struct{ trigger, card string }
	idx := make(map[key]int, len(existing))
	for i, dt := range existing {
		idx[key{dt.TriggerName, dt.CardName}] = i
	}

	for _, t := range triggers {
		k := key{t.TriggerType, t.CardName}
		if i, ok := idx[k]; ok {
			existing[i].Count++
			existing[i].GamesSeen++
		} else {
			existing = append(existing, muninn.DeadTrigger{
				TriggerName: t.TriggerType,
				CardName:    t.CardName,
				Count:       1,
				GamesSeen:   1,
			})
			idx[k] = len(existing) - 1
		}
	}

	_ = muninn.PersistDeadTriggersRaw(a.dataDir, existing)
}

// RecordCrash delegates directly to Muninn's crash log persistence.
func (a *muninnAdapter) RecordCrash(panicMsg string, stackTrace string, deckKeys []string) {
	_ = muninn.PersistCrashLogs(a.dataDir, []string{stackTrace}, deckKeys, 0, 0)
}
