# 07 — USDC payout capability

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Run `./run.sh` in mock mode. It reports the stablecoin extension without claiming a real route. USDC needs canonical token/network metadata, quote expiry, fee, minimum, and slippage checks. Cross-chain is unavailable in released Go `v0.15.1`, so this skeleton never sends funds. Read [conversion](../../docs/en-US/14-asset-conversion.md) and the [production adapter](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
