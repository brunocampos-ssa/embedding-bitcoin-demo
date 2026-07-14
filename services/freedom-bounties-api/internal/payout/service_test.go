package payout

import (
	"context"
	"errors"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment/mock"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/platform/database"
	"testing"
	"time"
)

func setup(t *testing.T) (*Service, *mock.Service) {
	t.Helper()
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = database.Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	m := mock.New(mock.Config{ProcessingDelay: time.Nanosecond})
	return NewService(db, m, Policy{MaxPayoutSats: 500, MaxFeeSats: 20, MaxDailyPayoutSats: 500}), m
}
func TestPrepareRequiresApprovalAndPolicy(t *testing.T) {
	s, _ := setup(t)
	ctx := context.Background()
	if _, err := s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC); !errors.Is(err, ErrNotApproved) {
		t.Fatalf("got %v", err)
	}
	if err := s.Approve(ctx, "submission-finance"); err != nil {
		t.Fatal(err)
	}
	s.policy.MaxPayoutSats = 50
	if _, err := s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC); !errors.Is(err, ErrPolicy) {
		t.Fatalf("got %v", err)
	}
	s.policy.MaxPayoutSats = 500
	s.policy.MaxFeeSats = 0
	if _, err := s.Prepare(ctx, "submission-finance", "bc1qdemo", payment.AssetBTC); !errors.Is(err, ErrPolicy) {
		t.Fatalf("fee policy: %v", err)
	}
}
func TestIdempotentConfirmationAndPaidOnlyAfterSuccess(t *testing.T) {
	s, _ := setup(t)
	ctx := context.Background()
	if err := s.Approve(ctx, "submission-finance"); err != nil {
		t.Fatal(err)
	}
	p, err := s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Confirm(ctx, p.ID, ""); !errors.Is(err, ErrIdempotencyRequired) {
		t.Fatal(err)
	}
	first, err := s.Confirm(ctx, p.ID, "key-1")
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.Confirm(ctx, p.ID, "key-1")
	if err != nil || again.ProviderPaymentID != first.ProviderPaymentID {
		t.Fatalf("idempotent result %+v %v", again, err)
	}
	if _, err = s.Confirm(ctx, p.ID, "different"); !errors.Is(err, ErrAlreadyPaid) {
		t.Fatal("different key must fail")
	}
	time.Sleep(time.Millisecond)
	done, err := s.Get(ctx, p.ID)
	if err != nil || done.State != Succeeded {
		t.Fatalf("got %+v %v", done, err)
	}
	var state string
	_ = s.db.QueryRow(`SELECT state FROM bounties WHERE id='bounty-finance'`).Scan(&state)
	if state != "PAID" {
		t.Fatalf("bounty state %s", state)
	}
	if _, err = s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC); !errors.Is(err, ErrAlreadyPaid) {
		t.Fatalf("duplicate payout: %v", err)
	}
}
func TestListReturnsEmptySliceNotNil(t *testing.T) {
	s, _ := setup(t)
	out, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("List returned nil — the API would then emit null and crash the history view")
	}
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d", len(out))
	}
}

// syncProvider forces SendPayout to report immediate success, as fast Lightning
// can, to exercise the synchronous-success path in Confirm.
type syncProvider struct{ *mock.Service }

func (p syncProvider) SendPayout(ctx context.Context, prepID, key string) (*payment.Result, error) {
	r, err := p.Service.SendPayout(ctx, prepID, key)
	if err != nil {
		return nil, err
	}
	r.Status = payment.StatusSucceeded
	return r, nil
}

func TestConfirmMarksBountyPaidOnSynchronousSuccess(t *testing.T) {
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = database.Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	s := NewService(db, syncProvider{mock.New(mock.Config{})}, Policy{MaxPayoutSats: 500, MaxFeeSats: 20, MaxDailyPayoutSats: 500})
	ctx := context.Background()
	if err = s.Approve(ctx, "submission-finance"); err != nil {
		t.Fatal(err)
	}
	p, err := s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC)
	if err != nil {
		t.Fatal(err)
	}
	done, err := s.Confirm(ctx, p.ID, "k")
	if err != nil || done.State != Succeeded {
		t.Fatalf("confirm: state=%v err=%v", done.State, err)
	}
	var state string
	_ = s.db.QueryRow(`SELECT state FROM bounties WHERE id='bounty-finance'`).Scan(&state)
	if state != "PAID" {
		t.Fatalf("bounty must be PAID after a synchronous success, got %s", state)
	}
}

func TestPrepareRequiresTreasuryBalance(t *testing.T) {
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err = database.Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	m := mock.New(mock.Config{StartEmpty: true, ProcessingDelay: time.Nanosecond})
	s := NewService(db, m, Policy{MaxPayoutSats: 500, MaxFeeSats: 20, MaxDailyPayoutSats: 500})
	ctx := context.Background()
	if err = s.Approve(ctx, "submission-finance"); err != nil {
		t.Fatal(err)
	}
	// Empty treasury: prepare must be blocked up front.
	if _, err = s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC); !errors.Is(err, payment.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds before deposit, got %v", err)
	}
	// Fund the treasury; the credit matures immediately (nanosecond delay).
	if _, err = m.Deposit(ctx, payment.DepositRequest{Rail: payment.RailLightning, AmountSats: 1000}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	if _, err = s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC); err != nil {
		t.Fatalf("expected prepare to succeed after deposit, got %v", err)
	}
}
func TestExpiredPreparation(t *testing.T) {
	s, _ := setup(t)
	ctx := context.Background()
	_ = s.Approve(ctx, "submission-finance")
	p, _ := s.Prepare(ctx, "submission-finance", "person@example.com", payment.AssetBTC)
	s.now = func() time.Time { return p.ExpiresAt.Add(time.Second) }
	if _, err := s.Confirm(ctx, p.ID, "late"); !errors.Is(err, payment.ErrExpired) {
		t.Fatalf("got %v", err)
	}
}
