package heimdall

// HuginnSink receives co-trigger observations for pattern discovery.
type HuginnSink interface {
	IngestCoTriggers(pairs []CoTriggerPair, deckNames []string)
}

// HuginnZoneCastSink is an OPTIONAL extension to HuginnSink. Implementers
// receive zone-cast lifecycle events for "casts from elsewhere" pattern
// discovery. Observer.RecordObservation type-asserts and skips silently
// if not implemented.
type HuginnZoneCastSink interface {
	IngestZoneCastEvents(events []ZoneCastEvent, deckNames []string, gameID string)
}

// HuginnExileLinkSink is an OPTIONAL extension to HuginnSink. Implementers
// receive O-Ring style link/unlink events for blink-pattern discovery.
type HuginnExileLinkSink interface {
	IngestExileLinkEvents(events []ExileLinkEvent, deckNames []string, gameID string)
}

// MuninnSink receives bug signals for persistent memory.
type MuninnSink interface {
	RecordParserGaps(gaps []string, gameID string)
	RecordDeadTriggers(triggers []DeadTrigger, gameID string)
	RecordCrash(panicMsg string, stackTrace string, deckKeys []string)
}

// MuninnExileLinkSink is an OPTIONAL extension to MuninnSink. Implementers
// receive exile-link events to detect O-Ring parity divergences (linked
// cards that fail to return when the source leaves the battlefield).
type MuninnExileLinkSink interface {
	RecordExileLinkEvents(events []ExileLinkEvent, gameID string)
}

// TelemetrySink sends health pulses (GA4).
type TelemetrySink interface {
	Pulse(stats HealthPulse)
}
