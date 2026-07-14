package httpserver

import (
	"context"
	"encoding/json"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment/mock"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payout"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/platform/database"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func handler(t *testing.T) http.Handler {
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	_ = database.Seed(context.Background(), db)
	m := mock.New(mock.Config{})
	p := payout.NewService(db, m, payout.Policy{MaxPayoutSats: 500, MaxFeeSats: 50, MaxDailyPayoutSats: 500})
	return New(db, p, m, "mock", "http://localhost:5173", slog.New(slog.NewTextHandler(io.Discard, nil)))
}
func TestValidationAndSafeErrors(t *testing.T) {
	h := handler(t)
	r := httptest.NewRequest(http.MethodPost, "/api/submissions/submission-finance/payouts/prepare", strings.NewReader(`{"destination":"x","extra":true}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var e apiError
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil || e.Code != "INVALID_REQUEST" || e.RequestID == "" {
		t.Fatalf("error=%+v err=%v", e, err)
	}
	if strings.Contains(w.Body.String(), "SQL") {
		t.Fatal("internal details leaked")
	}
}
func TestTreasuryDepositAndSelfPaymentGuard(t *testing.T) {
	h := handler(t)
	// Treasury balance is exposed; the identity pubkey is never serialized.
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/treasury", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"balanceSats"`) || !strings.Contains(w.Body.String(), `"lightningAddress"`) {
		t.Fatalf("treasury: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "pubkey") || strings.Contains(w.Body.String(), "Pubkey") {
		t.Fatalf("identity pubkey leaked: %s", w.Body.String())
	}
	// A Lightning deposit without an amount is rejected.
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/treasury/deposit", strings.NewReader(`{"rail":"lightning"}`)))
	if w.Code != 422 || !strings.Contains(w.Body.String(), "DEPOSIT_AMOUNT_REQUIRED") {
		t.Fatalf("deposit no-amount: %d %s", w.Code, w.Body.String())
	}
	// A valid Lightning deposit returns a payment request.
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/treasury/deposit", strings.NewReader(`{"rail":"lightning","amountSats":500}`)))
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"paymentRequest"`) {
		t.Fatalf("deposit: %d %s", w.Code, w.Body.String())
	}
	// Paying the treasury's own wallet is rejected after approval.
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/submissions/submission-finance/approve", strings.NewReader(`{}`)))
	if w.Code != 200 {
		t.Fatalf("approve: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/submissions/submission-finance/payouts/prepare", strings.NewReader(`{"destination":"spark1freedomtreasurydemo","asset":"BTC"}`)))
	if w.Code != 422 || !strings.Contains(w.Body.String(), "SELF_PAYMENT_REJECTED") {
		t.Fatalf("self-payment: %d %s", w.Code, w.Body.String())
	}
}
func TestHealthAndOrigin(t *testing.T) {
	h := handler(t)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"demoMode":true`) {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	r := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	r.Header.Set("Origin", "https://evil.example")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("status=%d", w.Code)
	}
}
