package payout

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/bounty"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
	"time"
)

type Policy struct {
	MaxPayoutSats      int64
	MaxFeeSats         int64
	MaxDailyPayoutSats int64
	PreparationTTL     time.Duration
}
type Service struct {
	db       *sql.DB
	payments payment.Service
	policy   Policy
	now      func() time.Time
}

func NewService(db *sql.DB, p payment.Service, policy Policy) *Service {
	if policy.PreparationTTL == 0 {
		policy.PreparationTTL = 10 * time.Minute
	}
	return &Service{db: db, payments: p, policy: policy, now: time.Now}
}
func newID(prefix string) string { return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()) }
func (s *Service) Prepare(ctx context.Context, submissionID, destination string, asset payment.Asset) (*Payout, error) {
	var state string
	var amount int64
	err := s.db.QueryRowContext(ctx, `SELECT submissions.state,bounties.reward_sats FROM submissions JOIN bounties ON bounties.id=submissions.bounty_id WHERE submissions.id=?`, submissionID).Scan(&state, &amount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if state != "APPROVED" {
		return nil, ErrNotApproved
	}
	if asset == "" {
		asset = payment.AssetBTC
	}
	if asset == payment.AssetBTC && amount > s.policy.MaxPayoutSats {
		return nil, ErrPolicy
	}
	var paid int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payouts WHERE submission_id=? AND state='SUCCEEDED'`, submissionID).Scan(&paid)
	if paid > 0 {
		return nil, ErrAlreadyPaid
	}
	var daily int64
	day := s.now().UTC().Add(-24 * time.Hour).Format(time.RFC3339Nano)
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount_base_units),0) FROM payouts WHERE state='SUCCEEDED' AND asset='BTC' AND updated_at>=?`, day).Scan(&daily)
	if asset == payment.AssetBTC && daily+amount > s.policy.MaxDailyPayoutSats {
		return nil, ErrPolicy
	}
	prep, err := s.payments.PreparePayout(ctx, payment.PrepareRequest{SubmissionID: submissionID, Destination: destination, Asset: asset, AmountBaseUnits: amount})
	if err != nil {
		return nil, err
	}
	if asset == payment.AssetBTC && prep.FeeBaseUnits > s.policy.MaxFeeSats {
		return nil, ErrPolicy
	}
	// Provider-agnostic balance precheck. This is defense-in-depth: the mock
	// adapter already rejects underfunded preparations, but the Breez adapter's
	// PreparePayout does not verify balance, so this is the load-bearing guard
	// for the real provider. Do not remove it as "redundant".
	if asset == payment.AssetBTC {
		info, ierr := s.payments.TreasuryInfo(ctx)
		if ierr != nil {
			return nil, ierr
		}
		if info.BalanceSats < amount+prep.FeeBaseUnits {
			return nil, payment.ErrInsufficientFunds
		}
	}
	now := s.now().UTC()
	expires := prep.ExpiresAt
	if max := now.Add(s.policy.PreparationTTL); expires.After(max) {
		expires = max
	}
	p := &Payout{ID: newID("payout-"), SubmissionID: submissionID, State: Prepared, Asset: asset, Rail: prep.Rail, AmountBaseUnits: amount, FeeBaseUnits: prep.FeeBaseUnits, DestinationType: prep.Destination.Type, DestinationMasked: prep.Destination.Masked, ProviderPreparationID: prep.ProviderPreparationID, PreparedAt: now, ExpiresAt: expires, UpdatedAt: now}
	_, err = s.db.ExecContext(ctx, `INSERT INTO payouts(id,submission_id,asset,rail,amount_base_units,fee_base_units,destination_type,destination_masked,destination_raw,provider_preparation_id,state,prepared_at,expires_at,updated_at)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, p.ID, p.SubmissionID, p.Asset, p.Rail, p.AmountBaseUnits, p.FeeBaseUnits, p.DestinationType, p.DestinationMasked, destination, p.ProviderPreparationID, p.State, p.PreparedAt.Format(time.RFC3339Nano), p.ExpiresAt.Format(time.RFC3339Nano), p.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (s *Service) Confirm(ctx context.Context, id, key string) (*Payout, error) {
	if key == "" {
		return nil, ErrIdempotencyRequired
	}
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p.IdempotencyKey != "" {
		if p.IdempotencyKey != key {
			return nil, ErrAlreadyPaid
		}
		return p, nil
	}
	if p.State != Prepared {
		return nil, ErrAlreadyPaid
	}
	if !s.now().Before(p.ExpiresAt) {
		_, _ = s.db.ExecContext(ctx, `UPDATE payouts SET state='EXPIRED',updated_at=? WHERE id=?`, s.now().UTC().Format(time.RFC3339Nano), id)
		return nil, payment.ErrExpired
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE payouts SET idempotency_key=?,state='PROCESSING',updated_at=? WHERE id=? AND idempotency_key IS NULL AND state='PREPARED'`, key, s.now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		tx.Rollback()
		return s.Get(ctx, id)
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	r, err := s.payments.SendPayout(ctx, p.ProviderPreparationID, key)
	if err != nil {
		_, _ = s.db.ExecContext(ctx, `UPDATE payouts SET state='PAYMENT_FAILED',failure_code=?,updated_at=? WHERE id=?`, "PROVIDER_SEND_FAILED", s.now().UTC().Format(time.RFC3339Nano), id)
		return nil, err
	}
	state := Processing
	if r.Status == payment.StatusSucceeded {
		state = Succeeded
	}
	if r.Status == payment.StatusFailed {
		state = PaymentFailed
	}
	_, err = s.db.ExecContext(ctx, `UPDATE payouts SET provider_payment_id=?,state=?,failure_code=?,updated_at=? WHERE id=?`, r.ProviderPaymentID, state, r.FailureCode, s.now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return nil, err
	}
	// A provider may complete synchronously (e.g. fast Lightning), so mark the
	// bounty PAID here too — not only in Get's PROCESSING→SUCCEEDED reconciliation.
	if state == Succeeded {
		_, _ = s.db.ExecContext(ctx, `UPDATE bounties SET state='PAID' WHERE id=(SELECT bounty_id FROM submissions WHERE id=?) AND state='APPROVED'`, p.SubmissionID)
	}
	return s.Get(ctx, id)
}
func (s *Service) Get(ctx context.Context, id string) (*Payout, error) {
	p := &Payout{}
	var prepared, expires, updated string
	var providerID, idem, failure sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id,submission_id,state,asset,rail,amount_base_units,fee_base_units,destination_type,destination_masked,provider_preparation_id,provider_payment_id,idempotency_key,prepared_at,expires_at,updated_at,failure_code FROM payouts WHERE id=?`, id).Scan(&p.ID, &p.SubmissionID, &p.State, &p.Asset, &p.Rail, &p.AmountBaseUnits, &p.FeeBaseUnits, &p.DestinationType, &p.DestinationMasked, &p.ProviderPreparationID, &providerID, &idem, &prepared, &expires, &updated, &failure)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.ProviderPaymentID = providerID.String
	p.IdempotencyKey = idem.String
	p.FailureCode = failure.String
	p.PreparedAt, _ = time.Parse(time.RFC3339Nano, prepared)
	p.ExpiresAt, _ = time.Parse(time.RFC3339Nano, expires)
	p.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	if p.State == Processing && p.ProviderPaymentID != "" {
		r, e := s.payments.GetPayment(ctx, p.ProviderPaymentID)
		if e == nil && r.Status != payment.StatusProcessing {
			next := Succeeded
			if r.Status == payment.StatusFailed {
				next = PaymentFailed
			}
			_, _ = s.db.ExecContext(ctx, `UPDATE payouts SET state=?,failure_code=?,updated_at=? WHERE id=?`, next, r.FailureCode, s.now().UTC().Format(time.RFC3339Nano), id)
			p.State = next
			p.FailureCode = r.FailureCode
			if next == Succeeded {
				_, _ = s.db.ExecContext(ctx, `UPDATE bounties SET state='PAID' WHERE id=(SELECT bounty_id FROM submissions WHERE id=?) AND state='APPROVED'`, p.SubmissionID)
			}
		}
	}
	return p, nil
}
func (s *Service) List(ctx context.Context) ([]Payout, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM payouts ORDER BY prepared_at DESC`)
	if err != nil {
		return nil, err
	}
	// Collect ids first, then close the result set before calling Get per row:
	// Get issues its own queries, and the pool is capped at one connection.
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	// Always return a non-nil slice so the API emits [] rather than null.
	out := []Payout{}
	for _, id := range ids {
		p, err := s.Get(ctx, id)
		if err != nil {
			// Skip a row that can't be loaded rather than failing the whole list.
			continue
		}
		out = append(out, *p)
	}
	return out, nil
}
func (s *Service) Approve(ctx context.Context, submissionID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var bid, state string
	if err = tx.QueryRowContext(ctx, `SELECT bounty_id,state FROM submissions WHERE id=?`, submissionID).Scan(&bid, &state); err != nil {
		return err
	}
	if state != "SUBMITTED" {
		return fmt.Errorf("invalid transition")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE submissions SET state='APPROVED',approved_at=? WHERE id=?`, s.now().UTC().Format(time.RFC3339Nano), submissionID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE bounties SET state=? WHERE id=? AND state=?`, bounty.Approved, bid, bounty.Submitted)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return fmt.Errorf("invalid transition")
	}
	return tx.Commit()
}
