//go:build breez

package breez

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"

	sdk "github.com/breez/breez-sdk-spark-go/breez_sdk_spark"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
)

type Config struct{ APIKey, Network, StorageDir, Mnemonic string }
type prepared struct {
	send  *sdk.PrepareSendPaymentResponse
	lnurl *sdk.PrepareLnurlPayResponse
}
type Service struct {
	sdk      *sdk.BreezSdk
	log      *slog.Logger
	mu       sync.Mutex
	prepared map[string]prepared
	identity payment.WalletIdentity
}
type eventListener struct{ log *slog.Logger }

func (e eventListener) OnEvent(event sdk.SdkEvent) {
	switch event.(type) {
	case sdk.SdkEventPaymentSucceeded:
		e.log.Info("Breez payment succeeded")
	case sdk.SdkEventPaymentPending:
		e.log.Info("Breez payment pending")
	case sdk.SdkEventPaymentFailed:
		e.log.Warn("Breez payment failed")
	case sdk.SdkEventSynced:
		e.log.Debug("Breez SDK synced")
	}
}

func New(c Config, log *slog.Logger) (payment.Service, func() error, error) {
	network := sdk.NetworkMainnet
	if strings.EqualFold(c.Network, "regtest") {
		network = sdk.NetworkRegtest
	} else if !strings.EqualFold(c.Network, "mainnet") {
		return nil, func() error { return nil }, fmt.Errorf("unsupported BREEZ_NETWORK %q", c.Network)
	}
	cfg := sdk.DefaultConfig(network)
	cfg.ApiKey = &c.APIKey
	seed := sdk.SeedMnemonic{Mnemonic: c.Mnemonic}
	client, err := sdk.Connect(sdk.ConnectRequest{Config: cfg, Seed: seed, StorageDir: c.StorageDir})
	if err != nil {
		return nil, func() error { return nil }, fmt.Errorf("connect Breez SDK: %w", err)
	}
	client.AddEventListener(eventListener{log})
	s := &Service{sdk: client, log: log, prepared: map[string]prepared{}}
	s.loadIdentity()
	return s, client.Disconnect, nil
}

// loadIdentity caches the treasury's own receiving identifiers so payouts to the
// treasury itself can be rejected. Failures are non-fatal: an identifier that
// cannot be loaded simply will not be matched by the self-payment guard.
func (s *Service) loadIdentity() {
	if info, err := s.sdk.GetInfo(sdk.GetInfoRequest{}); err == nil {
		s.identity.Pubkey = info.IdentityPubkey
	} else {
		s.log.Warn("Breez identity: GetInfo failed", "error", err)
	}
	if la, err := s.sdk.GetLightningAddress(); err == nil && la != nil {
		s.identity.LightningAddress = strings.ToLower(la.LightningAddress)
	}
	if r, err := s.sdk.ReceivePayment(sdk.ReceivePaymentRequest{PaymentMethod: sdk.ReceivePaymentMethodSparkAddress{}}); err == nil {
		s.identity.SparkAddress = strings.ToLower(r.PaymentRequest)
	}
}

// isSelfDestination reports whether the destination is one of the treasury's own
// cached receiving identifiers.
func (s *Service) isSelfDestination(low string) bool {
	return (s.identity.LightningAddress != "" && low == s.identity.LightningAddress) ||
		(s.identity.SparkAddress != "" && low == s.identity.SparkAddress)
}
func mask(v string) string {
	if len(v) < 12 {
		return "••••"
	}
	return v[:6] + "…" + v[len(v)-4:]
}
func prepID(v string) string {
	h := sha256.Sum256([]byte(v + time.Now().UTC().String()))
	return "breez-prep-" + hex.EncodeToString(h[:8])
}
func unix(t uint64) *time.Time { v := time.Unix(int64(t), 0).UTC(); return &v }
func (s *Service) ParseDestination(ctx context.Context, raw string) (*payment.ParsedDestination, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.isSelfDestination(strings.ToLower(strings.TrimSpace(raw))) {
		return nil, payment.ErrSelfPayment
	}
	in, err := s.sdk.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, mapError(err)
	}
	p := &payment.ParsedDestination{Asset: payment.AssetBTC, Masked: mask(raw), Raw: raw}
	switch v := in.(type) {
	case sdk.InputTypeBolt11Invoice:
		p.Type = payment.DestinationBolt11
		p.Rail = payment.RailLightning
		if v.Field0.AmountMsat != nil {
			a := int64(*v.Field0.AmountMsat / 1000)
			p.AmountBaseUnits = &a
		}
		p.ExpiresAt = unix(v.Field0.Timestamp + v.Field0.Expiry)
	case sdk.InputTypeLightningAddress:
		p.Type = payment.DestinationLightningAddress
		p.Rail = payment.RailLightning
	case sdk.InputTypeLnurlPay:
		p.Type = payment.DestinationLNURLPay
		p.Rail = payment.RailLightning
	case sdk.InputTypeBitcoinAddress:
		p.Type = payment.DestinationBitcoin
		p.Rail = payment.RailBitcoin
	case sdk.InputTypeSparkAddress:
		p.Type = payment.DestinationSpark
		p.Rail = payment.RailSpark
	case sdk.InputTypeSparkInvoice:
		p.Type = payment.DestinationSparkInvoice
		p.Rail = payment.RailSpark
		if v.Field0.ExpiryTime != nil {
			p.ExpiresAt = unix(*v.Field0.ExpiryTime)
		}
		if v.Field0.Amount != nil && (*v.Field0.Amount).IsInt64() {
			a := (*v.Field0.Amount).Int64()
			p.AmountBaseUnits = &a
		}
		if v.Field0.TokenIdentifier != nil {
			return nil, payment.ErrUnsupportedDestination
		}
	default:
		return nil, payment.ErrUnsupportedDestination
	}
	return p, nil
}
func (s *Service) PreparePayout(ctx context.Context, r payment.PrepareRequest) (*payment.Prepared, error) {
	parsed, err := s.ParseDestination(ctx, r.Destination)
	if err != nil {
		return nil, err
	}
	if r.Asset != payment.AssetBTC {
		return nil, payment.ErrUnsupportedDestination
	}
	id := prepID(r.SubmissionID)
	expires := time.Now().UTC().Add(10 * time.Minute)
	var fee int64
	var stored prepared
	input, err := s.sdk.Parse(strings.TrimSpace(r.Destination))
	if err != nil {
		return nil, mapError(err)
	}
	switch v := input.(type) {
	case sdk.InputTypeLightningAddress:
		res, e := s.sdk.PrepareLnurlPay(sdk.PrepareLnurlPayRequest{Amount: big.NewInt(r.AmountBaseUnits), PayRequest: v.Field0.PayRequest})
		if e != nil {
			return nil, mapError(e)
		}
		fee = int64(res.FeeSats)
		stored.lnurl = &res
	case sdk.InputTypeLnurlPay:
		res, e := s.sdk.PrepareLnurlPay(sdk.PrepareLnurlPayRequest{Amount: big.NewInt(r.AmountBaseUnits), PayRequest: v.Field0})
		if e != nil {
			return nil, mapError(e)
		}
		fee = int64(res.FeeSats)
		stored.lnurl = &res
	default:
		amount := big.NewInt(r.AmountBaseUnits)
		res, e := s.sdk.PrepareSendPayment(sdk.PrepareSendPaymentRequest{PaymentRequest: strings.TrimSpace(r.Destination), Amount: &amount})
		if e != nil {
			return nil, mapError(e)
		}
		fee = methodFee(res.PaymentMethod)
		stored.send = &res
	}
	if parsed.ExpiresAt != nil && parsed.ExpiresAt.Before(expires) {
		expires = *parsed.ExpiresAt
	}
	s.mu.Lock()
	s.prepared[id] = stored
	s.mu.Unlock()
	return &payment.Prepared{ProviderPreparationID: id, Destination: *parsed, Asset: r.Asset, Rail: parsed.Rail, AmountBaseUnits: r.AmountBaseUnits, FeeBaseUnits: fee, ExpiresAt: expires}, nil
}
func methodFee(m sdk.SendPaymentMethod) int64 {
	switch v := m.(type) {
	case sdk.SendPaymentMethodBolt11Invoice:
		return int64(v.LightningFeeSats)
	case sdk.SendPaymentMethodBitcoinAddress:
		return int64(v.FeeQuote.SpeedMedium.UserFeeSat)
	case sdk.SendPaymentMethodSparkAddress:
		if v.Fee.IsInt64() {
			return v.Fee.Int64()
		}
	case sdk.SendPaymentMethodSparkInvoice:
		if v.Fee.IsInt64() {
			return v.Fee.Int64()
		}
	}
	return 0
}
func (s *Service) SendPayout(ctx context.Context, id, key string) (*payment.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	p, ok := s.prepared[id]
	s.mu.Unlock()
	if !ok {
		return nil, payment.ErrNotFound
	}
	var paid sdk.Payment
	var err error
	if p.lnurl != nil {
		res, e := s.sdk.LnurlPay(sdk.LnurlPayRequest{PrepareResponse: *p.lnurl, IdempotencyKey: &key})
		paid = res.Payment
		err = e
	} else {
		res, e := s.sdk.SendPayment(sdk.SendPaymentRequest{PrepareResponse: *p.send, IdempotencyKey: &key})
		paid = res.Payment
		err = e
	}
	if err != nil {
		return nil, mapError(err)
	}
	return normalize(paid), nil
}
func (s *Service) GetPayment(ctx context.Context, id string) (*payment.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r, err := s.sdk.GetPayment(sdk.GetPaymentRequest{PaymentId: id})
	if err != nil {
		return nil, mapError(err)
	}
	return normalize(r.Payment), nil
}
func normalize(p sdk.Payment) *payment.Result {
	status := payment.StatusProcessing
	if p.Status == sdk.PaymentStatusCompleted {
		status = payment.StatusSucceeded
	} else if p.Status == sdk.PaymentStatusFailed {
		status = payment.StatusFailed
	}
	return &payment.Result{ProviderPaymentID: p.Id, Status: status, UpdatedAt: time.Unix(int64(p.Timestamp), 0).UTC()}
}
func (s *Service) TreasuryInfo(ctx context.Context) (*payment.TreasuryInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ensure := true
	info, err := s.sdk.GetInfo(sdk.GetInfoRequest{EnsureSynced: &ensure})
	if err != nil {
		return nil, mapError(err)
	}
	tokens := map[string]int64{}
	for id, tb := range info.TokenBalances {
		if tb.Balance != nil && tb.Balance.IsInt64() {
			tokens[id] = tb.Balance.Int64()
		}
	}
	if len(tokens) == 0 {
		tokens = nil
	}
	return &payment.TreasuryInfo{BalanceSats: int64(info.BalanceSats), Identity: s.identity, TokenBalances: tokens}, nil
}
func (s *Service) Deposit(ctx context.Context, r payment.DepositRequest) (*payment.DepositQuote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var method sdk.ReceivePaymentMethod
	switch r.Rail {
	case payment.RailLightning:
		var amount *uint64
		if r.AmountSats > 0 {
			a := uint64(r.AmountSats)
			amount = &a
		}
		method = sdk.ReceivePaymentMethodBolt11Invoice{Description: "FreedomBounties treasury deposit", AmountSats: amount}
	case payment.RailBitcoin:
		method = sdk.ReceivePaymentMethodBitcoinAddress{}
	case payment.RailSpark:
		method = sdk.ReceivePaymentMethodSparkAddress{}
	default:
		return nil, payment.ErrUnsupportedDestination
	}
	res, err := s.sdk.ReceivePayment(sdk.ReceivePaymentRequest{PaymentMethod: method})
	if err != nil {
		return nil, mapError(err)
	}
	var fee int64
	if res.Fee != nil && res.Fee.IsInt64() {
		fee = res.Fee.Int64()
	}
	return &payment.DepositQuote{Rail: r.Rail, PaymentRequest: res.PaymentRequest, FeeSats: fee}, nil
}
func (s *Service) Capabilities(context.Context) ([]payment.Capability, error) {
	return []payment.Capability{{Asset: payment.AssetBTC, Rail: payment.RailLightning, Enabled: true}, {Asset: payment.AssetBTC, Rail: payment.RailBitcoin, Enabled: true}, {Asset: payment.AssetBTC, Rail: payment.RailSpark, Enabled: true}, {Asset: payment.AssetUSDT, Rail: payment.RailCrossChain, Enabled: false, Note: "not exposed by released Go binding v0.15.1"}, {Asset: payment.AssetUSDC, Rail: payment.RailCrossChain, Enabled: false, Note: "not exposed by released Go binding v0.15.1"}}, nil
}
func mapError(err error) error {
	var sdkErr *sdk.SdkError
	if errors.As(err, &sdkErr) {
		text := strings.ToLower(err.Error())
		switch {
		case strings.Contains(text, "insufficient"):
			return payment.ErrInsufficientFunds
		case strings.Contains(text, "expired"):
			return payment.ErrExpired
		case strings.Contains(text, "not found"):
			return payment.ErrNotFound
		}
	}
	return fmt.Errorf("Breez SDK operation: %w", err)
}
