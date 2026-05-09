package credits

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Register wires the credit endpoints onto mux.
//
// Routes:
//
//   GET  /api/credits           — current balance for the caller
//   GET  /api/credits/history   — recent transactions for the caller
//   GET  /api/credits/quota     — free-tier + balance posture for gauntlet UI
//   POST /api/credits/spend     — debit credits (used by clients that
//                                 want to purchase non-gauntlet
//                                 testing; gauntlet itself spends via
//                                 the Spend method directly)
//
// Every endpoint requires the X-HexDek-Owner header — the same
// auth shim the rest of the API uses. No header → 401.
func (s *Store) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/credits", s.handleBalance)
	mux.HandleFunc("GET /api/credits/history", s.handleHistory)
	mux.HandleFunc("GET /api/credits/quota", s.handleQuota)
	mux.HandleFunc("POST /api/credits/spend", s.handleSpend)
}

func (s *Store) handleBalance(w http.ResponseWriter, r *http.Request) {
	owner := callerOwner(r)
	if owner == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	bal, err := s.GetBalance(r.Context(), owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, Balance{Owner: owner, Credits: bal})
}

func (s *Store) handleHistory(w http.ResponseWriter, r *http.Request) {
	owner := callerOwner(r)
	if owner == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	txns, err := s.History(r.Context(), owner, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if txns == nil {
		txns = []Transaction{}
	}
	writeJSON(w, map[string]any{
		"owner":        owner,
		"transactions": txns,
	})
}

func (s *Store) handleQuota(w http.ResponseWriter, r *http.Request) {
	owner := callerOwner(r)
	if owner == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	q, err := s.QuotaState(r.Context(), owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, q)
}

// SpendRequest is the JSON body for POST /api/credits/spend. The
// reference field is optional and is stored verbatim in the ledger
// — clients use it to pin a spend to a specific deck or game id.
type SpendRequest struct {
	Amount    int64  `json:"amount"`
	Reason    string `json:"reason"`
	Reference string `json:"reference,omitempty"`
}

func (s *Store) handleSpend(w http.ResponseWriter, r *http.Request) {
	owner := callerOwner(r)
	if owner == "" {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	var req SpendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		http.Error(w, "reason is required", http.StatusBadRequest)
		return
	}
	bal, err := s.Spend(r.Context(), owner, req.Amount, req.Reason, req.Reference)
	if errors.Is(err, ErrInsufficientCredits) {
		// 402 Payment Required is the correct semantic; a few proxies
		// strip it but the body carries the message regardless.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   "insufficient_credits",
			"balance": bal,
			"needed":  req.Amount,
		})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, Balance{Owner: owner, Credits: bal})
}

func callerOwner(r *http.Request) string {
	return strings.ToLower(strings.TrimSpace(r.Header.Get("X-HexDek-Owner")))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
