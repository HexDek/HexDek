package hexapi

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hexdek/hexdek/internal/heimdall"
	"github.com/hexdek/hexdek/internal/huginn"
)

// huginnAdapter implements heimdall.HuginnSink by converting Heimdall's
// lightweight CoTriggerPair observations into Huginn's RawObservation
// format and appending them to the raw_observations.json file for later
// ingestion.
type huginnAdapter struct {
	dataDir string
}

var _ heimdall.HuginnSink = (*huginnAdapter)(nil)

// IngestCoTriggers converts Heimdall CoTriggerPairs into Huginn's
// RawObservation format and persists them. The EffectPattern is
// synthesized from the card pair since Heimdall doesn't carry the
// full effect string; Huginn's NormalizePattern handles it gracefully.
func (a *huginnAdapter) IngestCoTriggers(pairs []heimdall.CoTriggerPair, deckNames []string) {
	if len(pairs) == 0 {
		return
	}

	dir := filepath.Join(a.dataDir, "huginn")

	existing, err := huginn.ReadRawObservations(dir)
	if err != nil {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for _, p := range pairs {
		// Synthesize an effect pattern from the pair. The format matches
		// what analytics.DetectCoTriggers produces so NormalizePattern
		// can extract the resource flow ("produces X → consumes X").
		// When no resource flow is available, use the card pair directly
		// so Huginn still groups observations by the same card combo.
		effectPattern := fmt.Sprintf("%s produces synergy, %s consumes synergy", p.CardA, p.CardB)

		existing = append(existing, huginn.RawObservation{
			CardA:         p.CardA,
			CardB:         p.CardB,
			ImpactScore:   p.ImpactScore,
			TurnWindow:    p.TurnWindow,
			EffectPattern: effectPattern,
			DeckNames:     append([]string(nil), deckNames...),
			Timestamp:     now,
		})
	}

	_ = huginn.PersistRawObservationsRaw(dir, existing)
}
