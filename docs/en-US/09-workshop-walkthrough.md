# Two-hour workshop walkthrough

[English](../en-US/09-workshop-walkthrough.md) | [Português do Brasil](../pt-BR/09-workshop-walkthrough.md)

- 18:00–18:10 — Context, objective, and “application versus wallet.”
- 18:10–18:25 — Architecture, domain states, approval versus settlement.
- 18:25–18:40 — `make dev`; complete the seeded mock flow.
- 18:40–19:00 — Inspect `PaymentService`, mock adapter, payout service, and constraints.
- 19:00–19:20 — Parse destinations, prepare, freeze amount/destination, review fees.
- 19:20–19:40 — Confirm the Lightning payout; discuss idempotency and reconciliation. If credentials/funding are unavailable, use mock processing/failure and inspect the compiled Breez code.
- 19:40–19:50 — Compare on-chain and Spark fee/settlement paths; use regtest where available.
- 19:50–20:00 — USDT/USDC release gap, treasury security, external signer, questions.

Fallback: keep the API in mock mode, use deterministic destinations, and show official API links plus `go test -tags breez`. Never let workshop timing pressure justify exposing or funding a personal wallet mnemonic.
