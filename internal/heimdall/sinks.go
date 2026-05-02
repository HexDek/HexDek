package heimdall

// HuginnSink receives co-trigger observations for pattern discovery.
type HuginnSink interface {
	IngestCoTriggers(pairs []CoTriggerPair, deckNames []string)
}

// MuninnSink receives bug signals for persistent memory.
type MuninnSink interface {
	RecordParserGaps(gaps []string, gameID string)
	RecordDeadTriggers(triggers []DeadTrigger, gameID string)
	RecordCrash(panicMsg string, stackTrace string, deckKeys []string)
}

// TelemetrySink sends health pulses (GA4).
type TelemetrySink interface {
	Pulse(stats HealthPulse)
}
