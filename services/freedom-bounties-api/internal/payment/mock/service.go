package mock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
)

// The mock treasury's own receiving identifiers. Pasting any of these as a
// payout destination trips the self-payment guard, reproducing the "I paid my
// own wallet" mistake as a teachable failure.
const (
	selfLightningAddress = "treasury@freedombounties.demo"
	selfSparkAddress     = "spark1freedomtreasurydemo"
	selfBitcoinAddress   = "bc1qfreedomtreasurydemo"
)

type Config struct {
	// TreasurySats seeds the starting balance. Zero means the default
	// (100_000) unless StartEmpty is set.
	TreasurySats int64
	// StartEmpty forces a zero starting balance so the deposit-first flow is
	// demonstrable. Used by the live demo; tests leave it false to stay funded.
	StartEmpty      bool
	FailureMode     string
	ProcessingDelay time.Duration
}

// pendingDeposit is a simulated incoming deposit that matures into balance once
// creditAt passes, mirroring how mock payments settle lazily on read.
type pendingDeposit struct {
	amount   int64
	creditAt time.Time
}

type Service struct {
	mu       sync.Mutex
	cfg      Config
	balance  int64
	identity payment.WalletIdentity
	deposits []pendingDeposit
	prepared map[string]payment.Prepared
	results  map[string]payment.Result
	byKey    map[string]string
	now      func() time.Time
}

func New(cfg Config) *Service {
	balance := cfg.TreasurySats
	if balance == 0 && !cfg.StartEmpty {
		balance = 100_000
	}
	if cfg.ProcessingDelay == 0 {
		cfg.ProcessingDelay = 750 * time.Millisecond
	}
	return &Service{
		cfg:      cfg,
		balance:  balance,
		identity: payment.WalletIdentity{Pubkey: "mock-treasury-pubkey", LightningAddress: selfLightningAddress, SparkAddress: selfSparkAddress},
		prepared: map[string]payment.Prepared{},
		results:  map[string]payment.Result{},
		byKey:    map[string]string{},
		now:      time.Now,
	}
}
func mask(s string) string {
	if len(s) < 12 {
		return "••••"
	}
	return s[:6] + "…" + s[len(s)-4:]
}
func id(prefix, s string) string {
	h := sha256.Sum256([]byte(s))
	return prefix + hex.EncodeToString(h[:8])
}

// isSelfDestination reports whether the trimmed, lowercased destination is one
// of the treasury's own receiving identifiers.
func isSelfDestination(low string) bool {
	return low == selfLightningAddress || low == selfSparkAddress || low == selfBitcoinAddress
}

// settleLocked matures any pending deposits whose credit time has passed. The
// caller must hold s.mu.
func (s *Service) settleLocked() {
	if len(s.deposits) == 0 {
		return
	}
	now := s.now()
	remaining := s.deposits[:0]
	for _, d := range s.deposits {
		if !now.Before(d.creditAt) {
			s.balance += d.amount
			continue
		}
		remaining = append(remaining, d)
	}
	s.deposits = remaining
}

func (s *Service) ParseDestination(ctx context.Context, raw string) (*payment.ParsedDestination, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	v := strings.TrimSpace(raw)
	low := strings.ToLower(v)
	if isSelfDestination(low) {
		return nil, payment.ErrSelfPayment
	}
	p := &payment.ParsedDestination{Asset: payment.AssetBTC, Masked: mask(v), Raw: v}
	switch {
	case strings.Contains(low, "expired"):
		return nil, payment.ErrExpired
	case strings.HasPrefix(low, "lnbc") || strings.HasPrefix(low, "lntb") || strings.HasPrefix(low, "lnbcrt"):
		p.Type = payment.DestinationBolt11
		p.Rail = payment.RailLightning
		t := s.now().Add(15 * time.Minute)
		p.ExpiresAt = &t
	case strings.HasPrefix(low, "lnurl"):
		p.Type = payment.DestinationLNURLPay
		p.Rail = payment.RailLightning
	case strings.Count(v, "@") == 1:
		p.Type = payment.DestinationLightningAddress
		p.Rail = payment.RailLightning
	case strings.HasPrefix(low, "bc1") || strings.HasPrefix(low, "tb1") || strings.HasPrefix(low, "bcrt1"):
		p.Type = payment.DestinationBitcoin
		p.Rail = payment.RailBitcoin
	case strings.HasPrefix(low, "spark1") || strings.HasPrefix(low, "sprt1"):
		p.Type = payment.DestinationSpark
		p.Rail = payment.RailSpark
	case strings.HasPrefix(low, "spark:"):
		p.Type = payment.DestinationSparkInvoice
		p.Rail = payment.RailSpark
	case strings.HasPrefix(low, "0x") || strings.HasPrefix(low, "ethereum:") || strings.HasPrefix(low, "solana:") || strings.HasPrefix(low, "tron:"):
		p.Type = payment.DestinationCrossChain
		p.Rail = payment.RailCrossChain
	default:
		return nil, payment.ErrUnsupportedDestination
	}
	return p, nil
}
func (s *Service) PreparePayout(ctx context.Context, r payment.PrepareRequest) (*payment.Prepared, error) {
	p, err := s.ParseDestination(ctx, r.Destination)
	if err != nil {
		return nil, err
	}
	if r.AmountBaseUnits <= 0 {
		return nil, payment.ErrPaymentFailed
	}
	p.Asset = r.Asset
	fee := int64(1)
	if p.Rail == payment.RailBitcoin {
		fee = 12
	}
	if p.Rail == payment.RailCrossChain {
		fee = 25
	}
	s.mu.Lock()
	s.settleLocked()
	if r.AmountBaseUnits+fee > s.balance {
		s.mu.Unlock()
		return nil, payment.ErrInsufficientFunds
	}
	exp := s.now().Add(10 * time.Minute)
	prep := payment.Prepared{ProviderPreparationID: id("mock-prep-", r.SubmissionID+r.Destination), Destination: *p, Asset: r.Asset, Rail: p.Rail, AmountBaseUnits: r.AmountBaseUnits, FeeBaseUnits: fee, ExpiresAt: exp}
	s.prepared[prep.ProviderPreparationID] = prep
	s.mu.Unlock()
	return &prep, nil
}
func (s *Service) SendPayout(ctx context.Context, prepID, key string) (*payment.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.byKey[key]; ok {
		r := s.results[old]
		return &r, nil
	}
	prep, ok := s.prepared[prepID]
	if !ok {
		return nil, payment.ErrNotFound
	}
	if !s.now().Before(prep.ExpiresAt) {
		return nil, payment.ErrExpired
	}
	pid := id("mock-pay-", prepID+key)
	status := payment.StatusProcessing
	failure := ""
	if s.cfg.FailureMode != "" || strings.Contains(strings.ToLower(prep.Destination.Raw), "fail") {
		status = payment.StatusFailed
		failure = "MOCK_PAYMENT_FAILED"
	} else {
		s.settleLocked()
		if prep.AmountBaseUnits+prep.FeeBaseUnits > s.balance {
			return nil, payment.ErrInsufficientFunds
		}
		s.balance -= prep.AmountBaseUnits + prep.FeeBaseUnits
	}
	r := payment.Result{ProviderPaymentID: pid, Status: status, FailureCode: failure, UpdatedAt: s.now()}
	s.results[pid] = r
	s.byKey[key] = pid
	return &r, nil
}
func (s *Service) GetPayment(ctx context.Context, id string) (*payment.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.results[id]
	if !ok {
		return nil, payment.ErrNotFound
	}
	if r.Status == payment.StatusProcessing && s.now().Sub(r.UpdatedAt) >= s.cfg.ProcessingDelay {
		r.Status = payment.StatusSucceeded
		r.UpdatedAt = s.now()
		s.results[id] = r
	}
	return &r, nil
}
func (s *Service) TreasuryInfo(ctx context.Context) (*payment.TreasuryInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settleLocked()
	return &payment.TreasuryInfo{BalanceSats: s.balance, Identity: s.identity}, nil
}
func (s *Service) Deposit(ctx context.Context, r payment.DepositRequest) (*payment.DepositQuote, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q := &payment.DepositQuote{Rail: r.Rail}
	switch r.Rail {
	case payment.RailLightning:
		q.PaymentRequest = id("lnbcdepositdemo", fmt.Sprintf("%d-%d", r.AmountSats, s.now().UnixNano()))
		exp := s.now().Add(15 * time.Minute)
		q.ExpiresAt = &exp
	case payment.RailBitcoin:
		q.PaymentRequest = selfBitcoinAddress
	case payment.RailSpark:
		q.PaymentRequest = selfSparkAddress
	default:
		return nil, payment.ErrUnsupportedDestination
	}
	// A positive amount schedules a simulated incoming credit; address-style
	// rails without an amount just return the address (funds arrive later).
	if r.AmountSats > 0 {
		s.mu.Lock()
		s.deposits = append(s.deposits, pendingDeposit{amount: r.AmountSats, creditAt: s.now().Add(s.cfg.ProcessingDelay)})
		s.mu.Unlock()
	}
	return q, nil
}
func (s *Service) Capabilities(context.Context) ([]payment.Capability, error) {
	return []payment.Capability{
		{Asset: payment.AssetBTC, Rail: payment.RailLightning, Enabled: true, Note: "mock"},
		{Asset: payment.AssetBTC, Rail: payment.RailBitcoin, Enabled: true, Note: "mock"},
		{Asset: payment.AssetBTC, Rail: payment.RailSpark, Enabled: true, Note: "mock"},
		{Asset: payment.AssetUSDT, Rail: payment.RailCrossChain, Enabled: true, Note: "mock extension"},
		{Asset: payment.AssetUSDC, Rail: payment.RailCrossChain, Enabled: true, Note: "mock extension"},
	}, nil
}
