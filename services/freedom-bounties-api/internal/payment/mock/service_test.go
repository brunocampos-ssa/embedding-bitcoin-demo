package mock

import (
	"context"
	"errors"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
	"testing"
	"time"
)

func TestParseDestinations(t *testing.T) {
	s := New(Config{})
	for input, want := range map[string]payment.DestinationType{"lnbc1demo": payment.DestinationBolt11, "person@example.com": payment.DestinationLightningAddress, "bc1qdemodestination": payment.DestinationBitcoin, "spark1demo": payment.DestinationSpark, "nonsense": payment.DestinationType("")} {
		got, err := s.ParseDestination(context.Background(), input)
		if want == "" {
			if !errors.Is(err, payment.ErrUnsupportedDestination) {
				t.Fatalf("%q: %v", input, err)
			}
			continue
		}
		if err != nil || got.Type != want {
			t.Fatalf("%q: got %+v, %v", input, got, err)
		}
	}
}
func TestSuccessFailureExpiryAndIdempotency(t *testing.T) {
	ctx := context.Background()
	s := New(Config{ProcessingDelay: time.Nanosecond})
	prep, err := s.PreparePayout(ctx, payment.PrepareRequest{SubmissionID: "s", Destination: "person@example.com", Asset: payment.AssetBTC, AmountBaseUnits: 100})
	if err != nil {
		t.Fatal(err)
	}
	first, err := s.SendPayout(ctx, prep.ProviderPreparationID, "same")
	if err != nil {
		t.Fatal(err)
	}
	again, _ := s.SendPayout(ctx, prep.ProviderPreparationID, "same")
	if first.ProviderPaymentID != again.ProviderPaymentID {
		t.Fatal("idempotency returned another payment")
	}
	time.Sleep(time.Millisecond)
	done, _ := s.GetPayment(ctx, first.ProviderPaymentID)
	if done.Status != payment.StatusSucceeded {
		t.Fatalf("got %s", done.Status)
	}
	f := New(Config{FailureMode: "always"})
	fp, _ := f.PreparePayout(ctx, payment.PrepareRequest{SubmissionID: "s", Destination: "person@example.com", Asset: payment.AssetBTC, AmountBaseUnits: 1})
	failed, _ := f.SendPayout(ctx, fp.ProviderPreparationID, "key")
	if failed.Status != payment.StatusFailed {
		t.Fatal("expected failure")
	}
	s.now = func() time.Time { return prep.ExpiresAt.Add(time.Second) }
	if _, err = s.SendPayout(ctx, prep.ProviderPreparationID, "late"); !errors.Is(err, payment.ErrExpired) {
		t.Fatalf("expected expiry, got %v", err)
	}
}
func TestInsufficientFunds(t *testing.T) {
	s := New(Config{TreasurySats: 10})
	_, err := s.PreparePayout(context.Background(), payment.PrepareRequest{SubmissionID: "s", Destination: "bc1qdemo", Asset: payment.AssetBTC, AmountBaseUnits: 11})
	if !errors.Is(err, payment.ErrInsufficientFunds) {
		t.Fatalf("got %v", err)
	}
}
func TestSelfPaymentGuardRejectsOwnWallet(t *testing.T) {
	ctx := context.Background()
	s := New(Config{})
	for _, dest := range []string{selfLightningAddress, selfSparkAddress, selfBitcoinAddress} {
		if _, err := s.ParseDestination(ctx, dest); !errors.Is(err, payment.ErrSelfPayment) {
			t.Fatalf("%q: expected self-payment, got %v", dest, err)
		}
		if _, err := s.PreparePayout(ctx, payment.PrepareRequest{SubmissionID: "s", Destination: dest, Asset: payment.AssetBTC, AmountBaseUnits: 100}); !errors.Is(err, payment.ErrSelfPayment) {
			t.Fatalf("%q prepare: expected self-payment, got %v", dest, err)
		}
	}
}
func TestDepositStartsEmptyAndCreditsAfterDelay(t *testing.T) {
	ctx := context.Background()
	base := time.Unix(1_700_000_000, 0).UTC()
	s := New(Config{StartEmpty: true, ProcessingDelay: time.Minute})
	s.now = func() time.Time { return base }

	info, err := s.TreasuryInfo(ctx)
	if err != nil || info.BalanceSats != 0 {
		t.Fatalf("empty start: %+v %v", info, err)
	}
	if info.Identity.LightningAddress != selfLightningAddress {
		t.Fatalf("identity not exposed: %+v", info.Identity)
	}
	// Cannot pay before funding.
	if _, err = s.PreparePayout(ctx, payment.PrepareRequest{SubmissionID: "s", Destination: "person@example.com", Asset: payment.AssetBTC, AmountBaseUnits: 100}); !errors.Is(err, payment.ErrInsufficientFunds) {
		t.Fatalf("expected insufficient before deposit, got %v", err)
	}
	q, err := s.Deposit(ctx, payment.DepositRequest{Rail: payment.RailLightning, AmountSats: 500})
	if err != nil || q.PaymentRequest == "" || q.ExpiresAt == nil {
		t.Fatalf("deposit quote: %+v %v", q, err)
	}
	// Not yet matured.
	if info, _ = s.TreasuryInfo(ctx); info.BalanceSats != 0 {
		t.Fatalf("credited too early: %d", info.BalanceSats)
	}
	// Advance past the processing delay: the deposit matures into balance.
	s.now = func() time.Time { return base.Add(2 * time.Minute) }
	if info, _ = s.TreasuryInfo(ctx); info.BalanceSats != 500 {
		t.Fatalf("expected 500 after delay, got %d", info.BalanceSats)
	}
	// A successful payout now debits the balance (amount + fee).
	prep, err := s.PreparePayout(ctx, payment.PrepareRequest{SubmissionID: "s", Destination: "person@example.com", Asset: payment.AssetBTC, AmountBaseUnits: 100})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.SendPayout(ctx, prep.ProviderPreparationID, "k"); err != nil {
		t.Fatal(err)
	}
	if info, _ = s.TreasuryInfo(ctx); info.BalanceSats != 399 {
		t.Fatalf("expected 399 after payout (500-100-1), got %d", info.BalanceSats)
	}
}
