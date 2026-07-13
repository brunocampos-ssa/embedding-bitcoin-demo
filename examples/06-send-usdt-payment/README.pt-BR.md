# 06 — Capacidade de payout USDT

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Execute `./run.sh` em mock para inspecionar capacidades. USDT usa unidades-base, ID canônico e rede; ticker não basta. O Go `v0.15.1` não expõe as rotas cross-chain documentadas, portanto este é um esqueleto mock e não envia fundos. Leia [USDT/USDC](../../docs/pt-BR/13-usdt-usdc.md) e a [capacidade no adaptador](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
