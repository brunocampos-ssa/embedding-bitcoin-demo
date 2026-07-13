# 02 — Prepare a payment

[English](README.md) | [Português do Brasil](README.pt-BR.md)

With the mock API running and the seeded submission approved, `./run.sh` prepares an on-chain review without sending. Expect `PREPARED`, `100` sats, a deterministic fee, masked destination, and expiry. Read [payment lifecycle](../../docs/en-US/05-payment-lifecycle.md) and [`payout/service.go`](../../services/freedom-bounties-api/internal/payout/service.go). Production preparation lives in the [Breez adapter](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
