# Breez SDK - Spark integration

[English](../en-US/07-breez-integration.md) | [Português do Brasil](../pt-BR/07-breez-integration.md)

The adapter is [`internal/payment/breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go). It uses verified Go APIs from the official [initialization](https://sdk-doc-spark.breez.technology/guide/initializing.html), [parsing](https://sdk-doc-spark.breez.technology/guide/parse.html), [sending](https://sdk-doc-spark.breez.technology/guide/send_payment.html), [LNURL](https://sdk-doc-spark.breez.technology/guide/lnurl_pay.html), [payment listing/retrieval](https://sdk-doc-spark.breez.technology/guide/list_payments.html), and [events](https://sdk-doc-spark.breez.technology/guide/events.html) guides.

`DefaultConfig` + `Connect` create one long-lived treasury instance. `Parse` normalizes BOLT11, Lightning address, LNURL-Pay, Bitcoin address, Spark address, and Spark invoice inputs. Lightning addresses use `PrepareLnurlPay`/`LnurlPay`; BOLT11, on-chain, and Spark use `PrepareSendPayment`/`SendPayment`. `GetInfo` and `ReceivePayment` back the treasury balance and deposit flow. `GetPayment` reconciles status and `Disconnect` shuts down safely.

Treasury wallet initialization: when `BREEZ_MNEMONIC` is unset, the adapter generates a fresh BIP39 mnemonic ([`mnemonic.go`](../../services/freedom-bounties-api/internal/payment/breez/mnemonic.go)), connects the SDK with it via `SeedMnemonic`, and persists it to `BREEZ_STORAGE_DIR/treasury.mnemonic` (mode `0600`), reusing it on later runs. This showcases SDK-driven wallet bootstrapping without pre-provisioning a seed. The mnemonic is never logged — only its source (generated/loaded) and file path are.

Compile verification:

```bash
cd services/freedom-bounties-api
go test -tags breez ./internal/payment/breez ./cmd/api
```

Manual validation needs a Breez API key, native platform support, and funding — the treasury mnemonic is generated for you when `BREEZ_MNEMONIC` is unset (or supply your own low-value one). Use regtest for Spark/on-chain; use very small mainnet amounts for Lightning. Verify parse, fee review, send, restart, and reconciliation. This repository did not execute a funded end-to-end payment. Cross-chain stablecoins are disabled for the release mismatch described in [implementation notes](00-implementation-notes.md).

<!-- nav-footer -->

---

<sub>📄 **Code:** [`internal/payment/breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go)</sub>

**[🏠 README](../../README.md)**  ·  ◀ [Running the demo](06-running-the-demo.md)  ·  [Security model](08-security-model.md) ▶
