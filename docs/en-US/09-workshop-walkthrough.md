# Two-hour workshop walkthrough

[English](../en-US/09-workshop-walkthrough.md) | [Português do Brasil](../pt-BR/09-workshop-walkthrough.md)

A facilitator's script for a two-hour, hands-on session. By the end the audience can explain how to **embed Bitcoin payouts into an ordinary application** through a payment-service boundary — without building a wallet or ever holding a recipient's keys.

- **Audience:** developers comfortable reading Go and TypeScript. No Bitcoin or Lightning background assumed.
- **Goal:** trace one payout end to end — **fund → approve → validate → review → confirm → reconcile** — and see exactly where the Breez SDK plugs in.
- **Before you start:** `cp .env.example .env`, then `make setup`. Keep `PAYMENT_PROVIDER=mock` (safe, no real funds move). Have this repo and the [running guide](06-running-the-demo.md) open.

Times assume an 18:00 start; adjust freely. Each segment lists what to **do** on screen, what to **explain**, and the one **takeaway** to land.

## 1 · 18:00–18:10 — Framing: application vs. wallet

- **Do:** `make dev`, open <http://localhost:5173>, show the bounty list.
- **Explain:** we build the *payer* side of an everyday app. The recipient keeps their own wallet and keys; we operate only a small payout treasury. Name what we are **not** building — wallet creation, recovery, portfolio, swaps — versus what we **are**: a payment intent.
- **Takeaway:** embedding payments is not building a wallet.

## 2 · 18:10–18:25 — Architecture and domain states

- **Do:** walk the diagram in [architecture](02-architecture.md) and the states in [domain model](03-domain-model.md).
- **Explain:** the `PaymentService` port with two interchangeable adapters (mock and Breez). The bounty/submission lifecycle, and the difference between *approval* (a human decision) and *settlement* (money actually moving).
- **Takeaway:** the port is the seam that keeps Bitcoin details out of the domain.

## 3 · 18:25–18:45 — Fund the treasury, then run the seeded flow

- **Do:** in the Treasury panel, note the balance starts at **0**. Create a Lightning deposit (e.g. 1,000 sats) and watch the balance rise (~1s in mock). Then open the seeded bounty and run it: approve → paste `mentor@example.com` → validate → review amount and fee → confirm → watch it succeed.
- **Explain:** the lifecycle *begins with funding* — you cannot pay from an empty treasury. Demonstrate the two guards on purpose: skip funding to trigger `INSUFFICIENT_TREASURY_FUNDS`, and paste the treasury's own address to trigger `SELF_PAYMENT_REJECTED`.
- **Takeaway:** deposit → balance → payout is the whole money story.

## 4 · 18:45–19:05 — Read the code by concept

- **Do:** open, in order — [`payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go) (the port and normalized money model), [`mock/service.go`](../../services/freedom-bounties-api/internal/payment/mock/service.go), [`payout/service.go`](../../services/freedom-bounties-api/internal/payout/service.go) (policy, idempotency, reconciliation), and [`database.go`](../../services/freedom-bounties-api/internal/platform/database/database.go) (constraints).
- **Explain:** amounts are normalized as `AmountBaseUnits`; a partial unique index enforces one successful payout per submission; the balance precheck is provider-agnostic defense-in-depth.
- **Takeaway:** the safety lives in the payout service and the database constraints, not the UI.

## 5 · 19:05–19:25 — Parse, prepare, freeze, review fees

- **Do:** try each destination type (`lnbc…`, `bc1…`, `spark1…`, `name@example.com`). Prepare a payout and inspect the frozen amount, fee, and expiry in the review step.
- **Explain:** `ParseDestination` normalizes many inputs to one shape; preparation *freezes* the amount and destination so confirm cannot change them; fees differ by rail.
- **Takeaway:** review shows exactly what will be sent, and nothing changes after the freeze.

## 6 · 19:25–19:45 — Confirm, idempotency, reconciliation

- **Do:** confirm a payout. Add `fail` to a destination to force a failure. Refresh a processing payout to watch reconciliation.
- **Explain:** the `Idempotency-Key` header makes a retried confirm safe; PROCESSING → SUCCEEDED/FAILED is reconciled by provider ID; a bounty is marked PAID only on success. If Breez credentials or funding are unavailable, stay in mock and read the compiled Breez adapter instead.
- **Takeaway:** one confirmed intent yields at most one payment, and status is reconciled rather than assumed.

## 7 · 19:45–19:55 — Real Breez mode and SDK-driven wallet init

- **Do:** open [Breez integration](07-breez-integration.md). Explain that setting `PAYMENT_PROVIDER=breez` in `.env` makes `make dev` build with `-tags breez` automatically. With `BREEZ_MNEMONIC` left empty, the SDK **bootstraps a fresh treasury wallet**: a BIP39 mnemonic is generated, connected via `SeedMnemonic`, and persisted to `treasury.mnemonic` (mode `0600`) so the same wallet returns on the next run.
- **Explain:** compare on-chain versus Spark fee and settlement paths; prefer regtest; real mainnet Lightning moves real satoshis.
- **Takeaway:** the SDK can create and reconnect the wallet for you — you supply policy, not key management.

## 8 · 19:55–20:00 — Limits, security, and questions

- **Do:** cover the USDT/USDC release gap, treasury security, and the external-signer direction — [security model](08-security-model.md) and [production treasury security](17-production-treasury-security.md).
- **Takeaway:** name what changes on the road from demo to production.

**Fallback:** keep the API in mock mode, use the deterministic demo destinations, and show the official Breez guide links plus `go test -tags breez`. Never let workshop timing pressure justify exposing or funding a personal wallet mnemonic — in real mode, let the app generate a throwaway treasury instead.
