package hexapi

import (
	"strings"

	"github.com/hexdek/hexdek/internal/heimdall"
	"github.com/hexdek/hexdek/internal/telemetry"
)

// telemetryAdapter implements heimdall.TelemetrySink by sending health
// pulses to GA4 via the telemetry package.
type telemetryAdapter struct {
	ga4 *telemetry.GA4Client
}

var _ heimdall.TelemetrySink = (*telemetryAdapter)(nil)

func (a *telemetryAdapter) Pulse(stats heimdall.HealthPulse) {
	if a.ga4 == nil {
		return
	}
	a.ga4.SendEvent("health_pulse", map[string]interface{}{
		"games_played":   stats.GamesPlayed,
		"parser_gaps":    stats.ParserGaps,
		"crashes":        stats.Crashes,
		"dead_triggers":  stats.DeadTriggers,
		"gap_cards":      strings.Join(stats.TopGapCards, ","),
		"engine_version": stats.EngineVersion,
	})
}
