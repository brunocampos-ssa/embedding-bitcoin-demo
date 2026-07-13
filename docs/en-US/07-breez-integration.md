# Breez SDK - Spark integration

[English](../en-US/07-breez-integration.md) | [Português do Brasil](../pt-BR/07-breez-integration.md)

The adapter is [`internal/payment/breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go). It uses verified Go APIs from the official [initialization](https://sdk-doc-spark.breez.technology/guide/initializing.html), [parsing](https://sdk-doc-spark.breez.technology/guide/parse.html), [sending](https://sdk-doc-spark.breez.technology/guide/send_payment.html), [LNURL](https://sdk-doc-spark.breez.technology/guide/lnurl_pay.html), [payment listing/retrieval](https://sdk-doc-spark.breez.technology/guide/list_payments.html), and [events](https://sdk-doc-spark.breez.technology/guide/events.html) guides.

`DefaultConfig` + `Connect` create one long-lived treasury instance. `Parse` normalizes BOLT11, Lightning address, LNURL-Pay, Bitcoin address, Spark address, and Spark invoice inputs. Lightning addresses use `PrepareLnurlPay`/`LnurlPay`; BOLT11, on-chain, and Spark use `PrepareSendPayment`/`SendPayment`. `GetPayment` reconciles status and `Disconnect` shuts down safely.

Compile verification:

```bash
cd services/freedom-bounties-api
go test -tags breez ./internal/payment/breez ./cmd/api
```

Manual validation needs a Breez API key, a new low-value treasury mnemonic, native platform support, and funding. Use regtest for Spark/on-chain; use very small mainnet amounts for Lightning. Verify parse, fee review, send, restart, and reconciliation. This repository did not execute a funded end-to-end payment. Cross-chain stablecoins are disabled for the release mismatch described in [implementation notes](00-implementation-notes.md).
