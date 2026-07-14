# Bitcoin, Lightning, Liquid, and Spark — a primer

[English](../en-US/19-bitcoin-lightning-liquid-spark.md) | [Português do Brasil](../pt-BR/19-bitcoin-lightning-liquid-spark.md)

This app moves money over a few different Bitcoin "rails." You do not need to be a Bitcoin expert to follow the demo, but a mental model of each layer makes the [assets and rails](04-payment-assets-and-rails.md) and [Breez integration](07-breez-integration.md) chapters click. Each section ends with **where it shows up in the code**.

## Bitcoin — the base settlement layer

Bitcoin is the underlying network. Transactions are grouped into blocks roughly every ten minutes and, once confirmed, are effectively final. That makes on-chain Bitcoin excellent for larger, less time-sensitive settlement, but slower and with a variable miner fee that is not ideal for tiny everyday payments.

Amounts are always integer **satoshis** (1 BTC = 100,000,000 sats) — never floating point. An on-chain destination looks like `bc1q…` (mainnet) or `bcrt1…` (regtest).

- **In this demo:** a `bc1…` address parses to the `bitcoin` rail. See `ParseDestination` in [`mock/service.go`](../../services/freedom-bounties-api/internal/payment/mock/service.go) and the real [`breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go); satoshis are the base unit in [`payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go).

## Lightning — the fast Layer 2 for small payments

Lightning is a network built *on top of* Bitcoin using payment channels. Payments settle in well under a second, cost a fraction of a satoshi to a few sats, and are ideal for global, low-value transfers — exactly the workshop's `100 sats` reward. Funds are still Bitcoin; Lightning is the delivery mechanism.

A recipient can share three shapes, all of which this app understands:
- a **BOLT11 invoice** — a one-time `lnbc…` string with amount and expiry baked in;
- a **Lightning address** — a friendly `name@domain` (resolved via LNURL behind the scenes);
- an **LNURL-pay** link.

- **In this demo:** `lnbc…`, `mentor@example.com`, and `lnurl…` all parse to the `lightning` rail. The real adapter uses `PrepareLnurlPay`/`LnurlPay` for addresses and `PrepareSendPayment`/`SendPayment` for invoices — see [`breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).

## Liquid — a Bitcoin sidechain (context)

Liquid is a Bitcoin **sidechain**: a separate, federated network pegged to Bitcoin with ~1-minute blocks, confidential amounts, and the ability to issue assets (for example L-BTC and stablecoins on Liquid). It is the settlement network behind Breez's **Nodeless (Liquid) SDK**, a sibling to the Spark SDK.

This project does **not** run on Liquid — it is included here because choosing a Breez SDK means choosing a network, and Liquid is the most common alternative to Spark. If your product needed confidential transfers or Liquid-issued stablecoins, the Nodeless (Liquid) SDK would be the natural pick, and the same `PaymentService` port in this repo would host that adapter instead.

- **In this demo:** not used. The port that would make swapping networks a one-adapter change is [`payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go) (`Service`).

## Spark — the Layer 2 this demo uses

Spark is a newer Bitcoin Layer 2 for fast, low-cost, self-custodial transfers of Bitcoin and tokens. It is what the **Breez SDK – Spark** targets, and it is the network this project integrates. Recipients share a `spark1…` address or a Spark invoice.

- **In this demo:** `spark1…` destinations parse to the `spark` rail; the real provider is the `breez-sdk-spark-go` binding wired in [`breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go) (build-tagged `breez`). Current SDK/release constraints — including why cross-chain stablecoins are disabled — are in [implementation notes](00-implementation-notes.md).

## At a glance

| Rail | Layer | Speed | Fees | Best for | Demo destination |
|------|-------|-------|------|----------|------------------|
| Bitcoin | Base chain | ~10 min | Variable (miner) | Larger, final settlement | `bc1…` |
| Lightning | L2 (channels) | Sub-second | Tiny | Small, global, frequent | `lnbc…`, `name@domain` |
| Liquid | Sidechain | ~1 min | Low, fixed | Confidential / issued assets | *(not used here)* |
| Spark | L2 | Fast | Low | Self-custodial BTC + tokens | `spark1…` |

The point of the app is that the domain never picks a rail by hand: it hands a destination to the payment service, which parses it and chooses a compatible route. Continue with [assets and rails](04-payment-assets-and-rails.md) to see how that choice is modeled.

<!-- nav-footer -->

---

<sub>📄 **Code:** [`internal/payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go)</sub>

**[🏠 README](../../README.md)**  ·  ◀ [Payment infrastructure, not another wallet](01-concept.md)  ·  [Payment assets and rails](04-payment-assets-and-rails.md) ▶
