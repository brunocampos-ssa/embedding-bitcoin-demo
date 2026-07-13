package payout

import (
	"errors"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
	"time"
)

type State string

const (
	Created          State = "CREATED"
	Validating       State = "VALIDATING"
	Prepared         State = "PREPARED"
	Processing       State = "PROCESSING"
	Succeeded        State = "SUCCEEDED"
	ValidationFailed State = "VALIDATION_FAILED"
	PaymentFailed    State = "PAYMENT_FAILED"
	Expired          State = "EXPIRED"
	Cancelled        State = "CANCELLED"
)

type Payout struct {
	ID                    string                  `json:"id"`
	SubmissionID          string                  `json:"submissionId"`
	State                 State                   `json:"state"`
	Asset                 payment.Asset           `json:"asset"`
	Rail                  payment.Rail            `json:"rail"`
	AmountBaseUnits       int64                   `json:"amountBaseUnits"`
	FeeBaseUnits          int64                   `json:"feeBaseUnits"`
	DestinationType       payment.DestinationType `json:"destinationType"`
	DestinationMasked     string                  `json:"destinationMasked"`
	ProviderPreparationID string                  `json:"-"`
	ProviderPaymentID     string                  `json:"providerPaymentId,omitempty"`
	IdempotencyKey        string                  `json:"-"`
	PreparedAt            time.Time               `json:"preparedAt"`
	ExpiresAt             time.Time               `json:"expiresAt"`
	UpdatedAt             time.Time               `json:"updatedAt"`
	FailureCode           string                  `json:"failureCode,omitempty"`
}

var (
	ErrNotApproved         = errors.New("submission is not approved")
	ErrAlreadyPaid         = errors.New("submission already has a successful payout")
	ErrIdempotencyRequired = errors.New("idempotency key is required")
	ErrPolicy              = errors.New("payout policy rejected request")
	ErrNotFound            = errors.New("payout not found")
)
