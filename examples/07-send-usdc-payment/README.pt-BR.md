# 07 — Capacidade de payout USDC

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Execute `./run.sh` em mock. Ele mostra a extensão sem alegar rota real. USDC exige metadados canônicos, validade, taxa, mínimo e slippage. Cross-chain não existe no Go `v0.15.1` lançado; o esqueleto nunca envia fundos. Leia [conversão](../../docs/pt-BR/14-asset-conversion.md) e o [adaptador](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
