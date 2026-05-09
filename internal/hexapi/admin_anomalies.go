package hexapi

import (
	"net/http"

	"github.com/hexdek/hexdek/internal/anticheat"
)

// RegisterAdminAnomalies wires the legacy /api/admin/anomalies
// endpoint that the showmatch boot path expects. The handler is a
// minimal pass-through — Phase 2 introduces richer endpoints
// (/api/admin/verifications, /api/admin/sanctions) via
// AdminAnticheatHandler — but the showmatch boot code already calls
// this function, so we keep it available to avoid breaking the
// build. Auditor may be nil when persistence isn't wired.
func RegisterAdminAnomalies(mux *http.ServeMux, auditor *anticheat.StatisticalAuditor) {
	mux.HandleFunc("GET /api/admin/anomalies", func(w http.ResponseWriter, r *http.Request) {
		if auditor == nil {
			http.Error(w, "anomaly auditor not configured", http.StatusServiceUnavailable)
			return
		}
		// Phase 2 surfaces sanctions/verifications via the new
		// AdminAnticheatHandler. This endpoint stays as a stub so
		// existing scrapers don't 404; it returns an empty payload
		// rather than leaking detector internals in their pre-Phase-2
		// shape.
		writeAdminJSON(w, struct {
			Note string `json:"note"`
		}{
			Note: "see /api/admin/verifications and /api/admin/sanctions for Phase 2 anti-cheat data",
		})
	})
}
