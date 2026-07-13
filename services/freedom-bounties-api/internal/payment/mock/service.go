package mock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
)

type Config struct {
	TreasurySats    int64
	FailureMode     string
	ProcessingDelay time.Duration
}
type Service struct {
	mu       sync.Mutex
	cfg      Config
	prepared map[string]payment.Prepared
	results  map[string]payment.Result
	byKey    map[string]string
	now      func() time.Time
}

func New(cfg Config) *Service {
	if cfg.TreasurySats == 0 {
		cfg.TreasurySats = 100_000
	}
	if cfg.ProcessingDelay == 0 {
		cfg.ProcessingDelay = 750 * time.Millisecond
	}
	return &Service{cfg: cfg, prepared: map[string]payment.Prepared{}, results: map[string]payment.Result{}, byKey: map[string]string{}, now: time.Now}
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

func (s *Service) ParseDestination(ctx context.Context, raw string) (*payment.ParsedDestination, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	v := strings.TrimSpace(raw)
	low := strings.ToLower(v)
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
	if r.AmountBaseUnits > s.cfg.TreasurySats {
		return nil, payment.ErrInsufficientFunds
	}
	p.Asset = r.Asset
	fee := int64(1)
	if p.Rail == payment.RailBitcoin {
		fee = 12
	}
	if p.Rail == payment.RailCrossChain {
		fee = 25
	}
	exp := s.now().Add(10 * time.Minute)
	prep := payment.Prepared{ProviderPreparationID: id("mock-prep-", r.SubmissionID+r.Destination), Destination: *p, Asset: r.Asset, Rail: p.Rail, AmountBaseUnits: r.AmountBaseUnits, FeeBaseUnits: fee, ExpiresAt: exp}
	s.mu.Lock()
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
func (s *Service) Capabilities(context.Context) ([]payment.Capability, error) {
	return []payment.Capability{
		{Asset: payment.AssetBTC, Rail: payment.RailLightning, Enabled: true, Note: "mock"},
		{Asset: payment.AssetBTC, Rail: payment.RailBitcoin, Enabled: true, Note: "mock"},
		{Asset: payment.AssetBTC, Rail: payment.RailSpark, Enabled: true, Note: "mock"},
		{Asset: payment.AssetUSDT, Rail: payment.RailCrossChain, Enabled: true, Note: "mock extension"},
		{Asset: payment.AssetUSDC, Rail: payment.RailCrossChain, Enabled: true, Note: "mock extension"},
	}, nil
}
