# 03 — Send a Lightning payment

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Warning: this script defaults to mock mode and sends no funds. Reset, start the API, approve the seeded submission, then run `./run.sh`; expect `PROCESSING`, followed by `SUCCEEDED` when queried. Real mode needs credentials, a funded isolated treasury, a `-tags breez` build, and a genuine invoice/address. Read [Breez integration](../../docs/en-US/07-breez-integration.md) and [adapter code](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
