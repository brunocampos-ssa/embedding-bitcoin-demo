# 01 — Parse a payment destination

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Requires the mock API and an approved seeded submission. Run `make api`, approve in the UI, then `./run.sh`. Expected JSON identifies `lightning-address`, `lightning`, masked destination, amount, fee, and expiry. The prepare endpoint intentionally combines parse/validation with a provider quote; direct parsing stays behind [`PaymentService`](../../services/freedom-bounties-api/internal/payment/models.go). Read [assets and rails](../../docs/en-US/04-payment-assets-and-rails.md) and the [Breez parser](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
