# Ativos e meios de pagamento

[English](../en-US/04-payment-assets-and-rails.md) | [Português do Brasil](../pt-BR/04-payment-assets-and-rails.md)

```mermaid
flowchart LR
  A[Ativo<br/>BTC, USDT, USDC] --> R[Rota]
  P[Meio/protocolo<br/>Lightning, Bitcoin, Spark, cross-chain] --> R
  D[Destino interpretado] --> R
  V[Capacidade do provedor] --> R
```

**Ativo** é a unidade transferida; **meio** é como ela viaja; **destino** vem da destinatária; **rota** combina opções compatíveis; **provedor** executa. BTC usa satoshis inteiros. Tokens exigem unidades-base inteiras, decimais, identificador canônico e rede — nunca só ticker nem ponto flutuante.

O adaptador real lançado atende BTC por Lightning, Bitcoin on-chain e Spark. A entrega de stablecoin documentada pela Breez é cross-chain, em duas etapas a partir de sats ou USDB; não é um falso meio universal “stablecoin sobre Spark”.

<!-- nav-footer -->

---

<sub>📄 **Código:** [`internal/payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Bitcoin, Lightning, Liquid e Spark](19-bitcoin-lightning-liquid-spark.md)  ·  [Arquitetura](02-architecture.md) ▶
