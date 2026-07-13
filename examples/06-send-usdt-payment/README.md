# 06 — USDT payout capability

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Run `./run.sh` against mock mode to inspect capability reporting. USDT uses integer base units plus a canonical token ID and external network; ticker alone is insufficient. The released Go SDK `v0.15.1` lacks documented cross-chain route APIs, so this is a clearly marked mock/skeleton and must not send funds. Read [USDT/USDC](../../docs/en-US/13-usdt-usdc.md) and the [adapter capability](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
