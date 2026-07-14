package main

import (
	"context"
	"errors"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/config"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment/breez"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payment/mock"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/payout"
	"github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/platform/database"
	httpplatform "github.com/freedom-bounties/embedding-bitcoin-demo/services/freedom-bounties-api/internal/platform/httpserver"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		log.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	ctx := context.Background()
	db, err := database.Open(ctx, cfg.DatabasePath)
	if err != nil {
		log.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err = database.Seed(ctx, db); err != nil {
		log.Error("seed failed", "error", err)
		os.Exit(1)
	}
	var payments payment.Service
	var closeProvider func() error = func() error { return nil }
	if cfg.PaymentProvider == "mock" {
		payments = mock.New(mock.Config{FailureMode: cfg.MockFailure, StartEmpty: true})
	} else {
		payments, closeProvider, err = breez.New(breez.Config{APIKey: cfg.BreezAPIKey, Network: cfg.BreezNetwork, StorageDir: cfg.BreezStorageDir, Mnemonic: cfg.BreezMnemonic}, log)
		if err != nil {
			log.Error("Breez startup failed", "error", err)
			os.Exit(1)
		}
	}
	defer closeProvider()
	ps := payout.NewService(db, payments, payout.Policy{MaxPayoutSats: cfg.MaxPayoutSats, MaxFeeSats: cfg.MaxFeeSats, MaxDailyPayoutSats: cfg.MaxDailyPayoutSats})
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: httpplatform.New(db, ps, payments, cfg.PaymentProvider, cfg.AllowedOrigin, log), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	go func() {
		log.Info("FreedomBounties API listening", "addr", cfg.HTTPAddr, "payment_provider", cfg.PaymentProvider)
		if e := srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Error("server failed", "error", e)
			os.Exit(1)
		}
	}()
	stop, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()
	<-stop.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err = srv.Shutdown(shutdown); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
