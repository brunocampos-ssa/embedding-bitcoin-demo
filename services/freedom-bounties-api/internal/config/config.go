package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv, HTTPAddr, DatabasePath, PaymentProvider, AllowedOrigin string
	MaxPayoutSats, MaxFeeSats, MaxDailyPayoutSats                  int64
	MockFailure                                                    string
	BreezAPIKey, BreezNetwork, BreezStorageDir, BreezMnemonic      string
	ShutdownTimeout                                                time.Duration
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func num(k string, d int64) int64 {
	v, err := strconv.ParseInt(env(k, strconv.FormatInt(d, 10)), 10, 64)
	if err != nil {
		return d
	}
	return v
}
func Load() (Config, error) {
	c := Config{AppEnv: env("APP_ENV", "development"), HTTPAddr: env("HTTP_ADDR", ":8080"), DatabasePath: env("DATABASE_PATH", "./data/freedom-bounties.db"), PaymentProvider: env("PAYMENT_PROVIDER", "mock"), AllowedOrigin: env("ALLOWED_ORIGIN", "http://localhost:5173"), MaxPayoutSats: num("MAX_PAYOUT_SATS", 5000), MaxFeeSats: num("MAX_FEE_SATS", 500), MaxDailyPayoutSats: num("MAX_DAILY_PAYOUT_SATS", 20000), MockFailure: env("MOCK_PAYMENT_FAILURE", ""), BreezAPIKey: os.Getenv("BREEZ_API_KEY"), BreezNetwork: env("BREEZ_NETWORK", "mainnet"), BreezStorageDir: env("BREEZ_STORAGE_DIR", "./data/breez"), BreezMnemonic: os.Getenv("BREEZ_MNEMONIC"), ShutdownTimeout: 10 * time.Second}
	if c.PaymentProvider != "mock" && c.PaymentProvider != "breez" {
		return c, fmt.Errorf("PAYMENT_PROVIDER must be mock or breez")
	}
	if c.PaymentProvider == "breez" && (strings.TrimSpace(c.BreezAPIKey) == "" || strings.TrimSpace(c.BreezMnemonic) == "") {
		return c, fmt.Errorf("Breez mode requires BREEZ_API_KEY and BREEZ_MNEMONIC")
	}
	if c.AppEnv != "development" && c.AllowedOrigin == "*" {
		return c, fmt.Errorf("wildcard CORS is forbidden outside development")
	}
	return c, nil
}
