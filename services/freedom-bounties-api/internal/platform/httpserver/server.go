package httpserver

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payout"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	db               *sql.DB
	payouts          *payout.Service
	payments         payment.Service
	provider, origin string
	log              *slog.Logger
	mux              *http.ServeMux
}

func New(db *sql.DB, p *payout.Service, pay payment.Service, provider, origin string, log *slog.Logger) http.Handler {
	s := &Server{db: db, payouts: p, payments: pay, provider: provider, origin: origin, log: log, mux: http.NewServeMux()}
	s.routes()
	return s.middleware(s.mux)
}
func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /api/bounties", s.listBounties)
	s.mux.HandleFunc("GET /api/bounties/{id}", s.getBounty)
	s.mux.HandleFunc("POST /api/bounties/{id}/assign", s.assign)
	s.mux.HandleFunc("POST /api/bounties/{id}/submissions", s.submit)
	s.mux.HandleFunc("POST /api/submissions/{id}/approve", s.approve)
	s.mux.HandleFunc("POST /api/submissions/{id}/payouts/prepare", s.prepare)
	s.mux.HandleFunc("POST /api/payouts/{id}/confirm", s.confirm)
	s.mux.HandleFunc("GET /api/payouts/{id}", s.getPayout)
	s.mux.HandleFunc("GET /api/payments", s.listPayments)
	s.mux.HandleFunc("GET /api/capabilities", s.capabilities)
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func reqID(r *http.Request) string { v, _ := r.Context().Value(requestIDKey{}).(string); return v }
func fail(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	write(w, status, apiError{code, msg, reqID(r)})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "The request body is invalid.")
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Only one JSON object is allowed.")
		return false
	}
	return true
}

type requestIDKey struct{}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", s.origin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key, X-Demo-Actor")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if o := r.Header.Get("Origin"); o != "" && o != s.origin {
			fail(w, r, http.StatusForbidden, "ORIGIN_NOT_ALLOWED", "Origin is not allowed.")
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", s.origin)
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		id := hex.EncodeToString(b)
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		start := time.Now()
		next.ServeHTTP(w, r.WithContext(ctx))
		s.log.Info("http request", "request_id", id, "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		fail(w, r, 503, "DATABASE_UNAVAILABLE", "Service is unavailable.")
		return
	}
	write(w, 200, map[string]any{"status": "ok", "paymentProvider": s.provider, "demoMode": s.provider == "mock", "authentication": "development-only demo actor"})
}
func (s *Server) listBounties(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,title,description,format,language,reward_sats,state,created_at FROM bounties ORDER BY created_at`)
	if err != nil {
		fail(w, r, 500, "INTERNAL_ERROR", "Could not load bounties.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, title, desc, format, lang, state, created string
		var reward int64
		if rows.Scan(&id, &title, &desc, &format, &lang, &reward, &state, &created) != nil {
			continue
		}
		out = append(out, map[string]any{"id": id, "title": title, "description": desc, "format": format, "language": lang, "rewardSats": reward, "state": state, "createdAt": created})
	}
	write(w, 200, out)
}
func (s *Server) getBounty(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var title, desc, format, lang, state, created string
	var reward int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT title,description,format,language,reward_sats,state,created_at FROM bounties WHERE id=?`, id).Scan(&title, &desc, &format, &lang, &reward, &state, &created); errors.Is(err, sql.ErrNoRows) {
		fail(w, r, 404, "BOUNTY_NOT_FOUND", "Bounty was not found.")
		return
	} else if err != nil {
		fail(w, r, 500, "INTERNAL_ERROR", "Could not load the bounty.")
		return
	}
	subs := []map[string]any{}
	rows, _ := s.db.QueryContext(r.Context(), `SELECT id,actor,evidence_url,notes,state,created_at,approved_at FROM submissions WHERE bounty_id=?`, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var sid, actor, url, notes, ss, sc string
			var approved sql.NullString
			_ = rows.Scan(&sid, &actor, &url, &notes, &ss, &sc, &approved)
			subs = append(subs, map[string]any{"id": sid, "actor": actor, "evidenceUrl": url, "notes": notes, "state": ss, "createdAt": sc, "approvedAt": approved.String})
		}
	}
	write(w, 200, map[string]any{"id": id, "title": title, "description": desc, "format": format, "language": lang, "rewardSats": reward, "state": state, "createdAt": created, "submissions": subs})
}
func actor(r *http.Request) string {
	if a := strings.TrimSpace(r.Header.Get("X-Demo-Actor")); a != "" {
		return a
	}
	return "demo-organizer"
}
func (s *Server) audit(ctx context.Context, r *http.Request, action, kind, id, outcome string) {
	_, _ = s.db.ExecContext(ctx, `INSERT INTO audit_records(request_id,actor,action,entity_type,entity_id,outcome,created_at)VALUES(?,?,?,?,?,?,?)`, reqID(r), actor(r), action, kind, id, outcome, time.Now().UTC().Format(time.RFC3339Nano))
}
func (s *Server) approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.payouts.Approve(r.Context(), id); err != nil {
		s.audit(r.Context(), r, "approve", "submission", id, "rejected")
		fail(w, r, 409, "INVALID_TRANSITION", "Only a submitted entry can be approved.")
		return
	}
	s.audit(r.Context(), r, "approve", "submission", id, "approved")
	write(w, 200, map[string]any{"id": id, "state": "APPROVED"})
}
func (s *Server) prepare(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Destination string        `json:"destination"`
		Asset       payment.Asset `json:"asset"`
	}
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Destination) == "" {
		fail(w, r, 422, "DESTINATION_REQUIRED", "Payment address or invoice is required.")
		return
	}
	p, err := s.payouts.Prepare(r.Context(), r.PathValue("id"), in.Destination, in.Asset)
	if err != nil {
		s.mapError(w, r, err)
		return
	}
	s.audit(r.Context(), r, "prepare", "payout", p.ID, "prepared")
	write(w, 201, p)
}
func (s *Server) confirm(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	p, err := s.payouts.Confirm(r.Context(), r.PathValue("id"), key)
	if err != nil {
		s.mapError(w, r, err)
		return
	}
	s.audit(r.Context(), r, "confirm", "payout", p.ID, string(p.State))
	write(w, 200, p)
}
func (s *Server) getPayout(w http.ResponseWriter, r *http.Request) {
	p, err := s.payouts.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		s.mapError(w, r, err)
		return
	}
	write(w, 200, p)
}
func (s *Server) listPayments(w http.ResponseWriter, r *http.Request) {
	p, err := s.payouts.List(r.Context())
	if err != nil {
		fail(w, r, 500, "INTERNAL_ERROR", "Could not load payment history.")
		return
	}
	write(w, 200, p)
}
func (s *Server) capabilities(w http.ResponseWriter, r *http.Request) {
	c, err := s.payments.Capabilities(r.Context())
	if err != nil {
		fail(w, r, 500, "INTERNAL_ERROR", "Could not load capabilities.")
		return
	}
	write(w, 200, c)
}
func (s *Server) assign(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Actor string `json:"actor"`
	}
	if !decode(w, r, &in) || strings.TrimSpace(in.Actor) == "" {
		return
	}
	id := r.PathValue("id")
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err == nil {
		aid := "assignment-" + id
		_, err = tx.ExecContext(r.Context(), `INSERT INTO assignments(id,bounty_id,actor,created_at)VALUES(?,?,?,?)`, aid, id, in.Actor, time.Now().UTC().Format(time.RFC3339Nano))
		if err == nil {
			_, err = tx.ExecContext(r.Context(), `UPDATE bounties SET state='ASSIGNED' WHERE id=? AND state='OPEN'`, id)
		}
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}
	if err != nil {
		fail(w, r, 409, "INVALID_TRANSITION", "Bounty cannot be assigned.")
		return
	}
	write(w, 201, map[string]string{"state": "ASSIGNED"})
}
func (s *Server) submit(w http.ResponseWriter, r *http.Request) {
	var in struct{ Actor, EvidenceURL, Notes string }
	if !decode(w, r, &in) {
		return
	}
	if in.Actor == "" || in.EvidenceURL == "" {
		fail(w, r, 422, "VALIDATION_FAILED", "Actor and evidenceURL are required.")
		return
	}
	id := r.PathValue("id")
	sid := fmt.Sprintf("submission-%d", time.Now().UnixNano())
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err == nil {
		_, err = tx.ExecContext(r.Context(), `INSERT INTO submissions(id,bounty_id,actor,evidence_url,notes,state,created_at)VALUES(?,?,?,?,?,'SUBMITTED',?)`, sid, id, in.Actor, in.EvidenceURL, in.Notes, time.Now().UTC().Format(time.RFC3339Nano))
		if err == nil {
			_, err = tx.ExecContext(r.Context(), `UPDATE bounties SET state='SUBMITTED' WHERE id=? AND state='ASSIGNED'`, id)
		}
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}
	if err != nil {
		fail(w, r, 409, "INVALID_TRANSITION", "Submission cannot be created.")
		return
	}
	write(w, 201, map[string]string{"id": sid, "state": "SUBMITTED"})
}
func (s *Server) mapError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, payout.ErrNotFound), errors.Is(err, payment.ErrNotFound):
		fail(w, r, 404, "NOT_FOUND", "The requested record was not found.")
	case errors.Is(err, payout.ErrNotApproved):
		fail(w, r, 409, "SUBMISSION_NOT_APPROVED", "Approve the submission before preparing a payout.")
	case errors.Is(err, payout.ErrAlreadyPaid):
		fail(w, r, 409, "PAYOUT_ALREADY_EXISTS", "This submission already has a payout or confirmation.")
	case errors.Is(err, payout.ErrIdempotencyRequired):
		fail(w, r, 400, "IDEMPOTENCY_KEY_REQUIRED", "An Idempotency-Key header is required.")
	case errors.Is(err, payout.ErrPolicy):
		fail(w, r, 422, "PAYOUT_POLICY_REJECTED", "The payout exceeds a configured safety limit.")
	case errors.Is(err, payment.ErrUnsupportedDestination):
		fail(w, r, 422, "UNSUPPORTED_DESTINATION", "Use a supported payment address or invoice.")
	case errors.Is(err, payment.ErrExpired):
		fail(w, r, 410, "PAYMENT_EXPIRED", "The payment request or preparation expired.")
	case errors.Is(err, payment.ErrInsufficientFunds):
		fail(w, r, 422, "INSUFFICIENT_TREASURY_FUNDS", "The demo treasury has insufficient funds.")
	default:
		s.log.Error("request failed", "request_id", reqID(r), "error", err)
		fail(w, r, 500, "INTERNAL_ERROR", "The request could not be completed.")
	}
}
