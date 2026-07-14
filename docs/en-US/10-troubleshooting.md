# Troubleshooting

[English](../en-US/10-troubleshooting.md) | [Português do Brasil](../pt-BR/10-troubleshooting.md)

- **Unsupported input:** use a BOLT11, Lightning address, LNURL-Pay, matching-network Bitcoin address, or Spark address/invoice. Mock fixtures are in the root README.
- **Expired invoice/preparation:** prepare again; never reuse an expired provider quote.
- **Missing API key/mnemonic:** Breez mode fails closed during configuration.
- **Invalid network:** only `mainnet` and `regtest` are accepted by this SDK release.
- **Insufficient balance / limit:** fund the isolated treasury or lower the demo amount; inspect configured policy limits.
- **SDK initialization:** check native platform support, storage permissions, API key, network, and connectivity.
- **Database:** stop the API and run `make reset`. Do not reset during a real pending payment.
- **Frontend/API:** verify ports 5173/8080 and `VITE_API_BASE_URL`; the Vite proxy handles local development.
- **CORS:** set `ALLOWED_ORIGIN` to the exact frontend origin. Wildcard is rejected outside development.
- **Port conflict:** change `HTTP_ADDR` or Vite's port.
- **Pending payment:** poll its existing payout/provider ID. Do not create another intent.
- **Duplicate confirmation:** reuse the same idempotency key; a different key is rejected.
- **Real payout disabled:** build with `-tags breez`; cross-chain stablecoins remain unavailable in Go `v0.15.1`.

<!-- nav-footer -->

---

**[🏠 README](../../README.md)**  ·  ◀ [Two-hour workshop walkthrough](09-workshop-walkthrough.md)  ·  [Implementation notes](00-implementation-notes.md) ▶
