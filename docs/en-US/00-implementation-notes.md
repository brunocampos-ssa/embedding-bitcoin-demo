# Implementation notes

[English](../en-US/00-implementation-notes.md) | [Português do Brasil](../pt-BR/00-implementation-notes.md)

Verified on 2026-07-13 against the official [Breez SDK - Spark documentation](https://sdk-doc-spark.breez.technology/) and official [`breez-sdk-spark-go`](https://github.com/breez/breez-sdk-spark-go) source.

## Pinned integration

- Go module: `github.com/breez/breez-sdk-spark-go v0.15.1`; package: `github.com/breez/breez-sdk-spark-go/breez_sdk_spark`.
- The binding uses CGO and bundled native libraries for Linux amd64/arm64, macOS amd64/arm64, Windows amd64, Android, and iOS. Default mock builds avoid it with the `breez` build tag.
- Supported SDK networks in this release are `NetworkMainnet` and `NetworkRegtest`. Regtest supports Spark, on-chain, and token development; Lightning testing requires small-value mainnet payments.
- Initialization uses `DefaultConfig`, `SeedMnemonic`, `ConnectRequest`, and `Connect`; shutdown uses `Disconnect`.
- Parsing uses `Parse`. General sends use `PrepareSendPayment`, `SendPayment`, and `GetPayment`. Lightning addresses/LNURL-Pay use `PrepareLnurlPay` and `LnurlPay`. Events use `AddEventListener` with `EventListener.OnEvent`.
- The SDK instance is long-lived because this application has one treasury. Official server mode (`DefaultServerConfig`) targets multi-tenant servers that create an ephemeral wallet instance per request, so it is intentionally not used.

## Important deviations and limits

The current documentation describes cross-chain USDC/USDT delivery to EVM-family chains, Solana, and Tron from Spark BTC sats or USDB through a two-leg provider flow. However, released Go binding `v0.15.1` does not expose `CrossChainAddress`, route discovery, or cross-chain payment types. Real USDT/USDC is therefore disabled; mock mode and the normalized asset model remain available. This must be revisited when a released Go tag exposes the documented API.

Real BTC Lightning, on-chain, and Spark code compiles against the native SDK but was not end-to-end funded or credential-tested in this repository. LNURL resolution is network-dependent. SDK calls are synchronous at the Go boundary and do not accept `context.Context`; the adapter checks cancellation before calls, while HTTP timeouts bound the caller but cannot forcibly interrupt native work already in progress.

Provider preparation values are held in memory and application preparation metadata is persisted. After a process restart, a prepared-but-unsent payout must expire and be prepared again; a payout with a provider payment ID is reconciled with `GetPayment`. A potentially successful payment is never replaced with a new intent automatically.

Mnemonic-in-environment is demo-only. Production should use the documented external signer interfaces and a secrets manager. The SDK idempotency key must be a UUID for supported Spark-based transfers; the web client uses `crypto.randomUUID()`.

<!-- nav-footer -->

---

**[🏠 README](../../README.md)**  ·  ◀ [Troubleshooting](10-troubleshooting.md)  ·  [Next steps](11-next-steps.md) ▶
