package payment

import (
	"context"
	"errors"
	"time"
)

type Asset string

const (
	AssetBTC  Asset = "BTC"
	AssetUSDT Asset = "USDT"
	AssetUSDC Asset = "USDC"
)

type Rail string

const (
	RailLightning  Rail = "lightning"
	RailBitcoin    Rail = "bitcoin"
	RailSpark      Rail = "spark"
	RailCrossChain Rail = "cross-chain"
)

type DestinationType string

const (
	DestinationBolt11           DestinationType = "bolt11"
	DestinationLightningAddress DestinationType = "lightning-address"
	DestinationLNURLPay         DestinationType = "lnurl-pay"
	DestinationBitcoin          DestinationType = "bitcoin-address"
	DestinationSpark            DestinationType = "spark-address"
	DestinationSparkInvoice     DestinationType = "spark-invoice"
	DestinationCrossChain       DestinationType = "cross-chain-address"
)

type Status string

const (
	StatusProcessing Status = "PROCESSING"
	StatusSucceeded  Status = "SUCCEEDED"
	StatusFailed     Status = "PAYMENT_FAILED"
)

type TokenMetadata struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Ticker     Asset  `json:"ticker"`
	Decimals   uint8  `json:"decimals"`
	Network    string `json:"network"`
}
type Capability struct {
	Asset   Asset  `json:"asset"`
	Rail    Rail   `json:"rail"`
	Enabled bool   `json:"enabled"`
	Note    string `json:"note,omitempty"`
}
type ParsedDestination struct {
	Type            DestinationType `json:"type"`
	Asset           Asset           `json:"asset"`
	Rail            Rail            `json:"rail"`
	Masked          string          `json:"masked"`
	AmountBaseUnits *int64          `json:"amountBaseUnits,omitempty"`
	Token           *TokenMetadata  `json:"token,omitempty"`
	ExpiresAt       *time.Time      `json:"expiresAt,omitempty"`
	Raw             string          `json:"-"`
}
type PrepareRequest struct {
	SubmissionID    string
	Destination     string
	Asset           Asset
	AmountBaseUnits int64
	TokenIdentifier string
}
type Prepared struct {
	ProviderPreparationID string            `json:"providerPreparationId"`
	Destination           ParsedDestination `json:"destination"`
	Asset                 Asset             `json:"asset"`
	Rail                  Rail              `json:"rail"`
	AmountBaseUnits       int64             `json:"amountBaseUnits"`
	FeeBaseUnits          int64             `json:"feeBaseUnits"`
	ExpiresAt             time.Time         `json:"expiresAt"`
}
type Result struct {
	ProviderPaymentID string    `json:"providerPaymentId"`
	Status            Status    `json:"status"`
	FailureCode       string    `json:"failureCode,omitempty"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// WalletIdentity describes the treasury's own receiving identifiers. The public
// key is used server-side for the self-payment guard and is never serialized to
// the browser.
type WalletIdentity struct {
	Pubkey           string `json:"-"`
	LightningAddress string `json:"lightningAddress,omitempty"`
	SparkAddress     string `json:"sparkAddress,omitempty"`
}

// TreasuryInfo is the operational balance and identity of the payout treasury.
type TreasuryInfo struct {
	BalanceSats   int64            `json:"balanceSats"`
	Identity      WalletIdentity   `json:"identity"`
	TokenBalances map[string]int64 `json:"tokenBalances,omitempty"`
}

// DepositRequest asks the provider to mint a receive request that funds the
// treasury over the given rail. AmountSats is required by the mock (so it can
// simulate the incoming credit) and is optional for address-style rails in real
// mode, where funds may arrive later and out of band.
type DepositRequest struct {
	Rail       Rail
	AmountSats int64
}

// DepositQuote is a receive request (invoice or address) the depositor pays to
// fund the treasury.
type DepositQuote struct {
	Rail           Rail       `json:"rail"`
	PaymentRequest string     `json:"paymentRequest"`
	FeeSats        int64      `json:"feeSats"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

type Service interface {
	ParseDestination(context.Context, string) (*ParsedDestination, error)
	PreparePayout(context.Context, PrepareRequest) (*Prepared, error)
	SendPayout(context.Context, string, string) (*Result, error)
	GetPayment(context.Context, string) (*Result, error)
	Capabilities(context.Context) ([]Capability, error)
	TreasuryInfo(context.Context) (*TreasuryInfo, error)
	Deposit(context.Context, DepositRequest) (*DepositQuote, error)
}

var (
	ErrUnsupportedDestination = errors.New("unsupported payment destination")
	ErrExpired                = errors.New("payment request expired")
	ErrInsufficientFunds      = errors.New("insufficient treasury funds")
	ErrPaymentFailed          = errors.New("payment failed")
	ErrNotFound               = errors.New("payment not found")
	ErrSelfPayment            = errors.New("destination is the treasury's own wallet")
)
