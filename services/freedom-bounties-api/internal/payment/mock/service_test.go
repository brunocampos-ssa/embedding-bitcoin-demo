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
