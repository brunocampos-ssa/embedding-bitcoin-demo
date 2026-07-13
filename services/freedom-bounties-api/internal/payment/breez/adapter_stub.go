//go:build !breez

package breez

import (
	"errors"
	"log/slog"

	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
)

type Config struct {
	APIKey     string
	Network    string
	StorageDir string
	Mnemonic   string
}

func New(Config, *slog.Logger) (payment.Service, func() error, error) {
	return nil, func() error { return nil }, errors.New("Breez provider is not included in this binary; rebuild with -tags breez")
}
