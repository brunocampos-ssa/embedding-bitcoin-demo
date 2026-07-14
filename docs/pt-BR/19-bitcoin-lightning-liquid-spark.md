# Bitcoin, Lightning, Liquid e Spark — uma introdução

[English](../en-US/19-bitcoin-lightning-liquid-spark.md) | [Português do Brasil](../pt-BR/19-bitcoin-lightning-liquid-spark.md)

Esta aplicação move dinheiro por alguns "meios" (rails) diferentes do Bitcoin. Você não precisa ser especialista em Bitcoin para acompanhar a demonstração, mas um modelo mental de cada camada faz os capítulos de [ativos e meios de pagamento](04-payment-assets-and-rails.md) e [integração Breez](07-breez-integration.md) fazerem sentido. Cada seção termina com **onde isso aparece no código**.

## Bitcoin — a camada base de liquidação

Bitcoin é a rede subjacente. As transações são agrupadas em blocos a cada cerca de dez minutos e, uma vez confirmadas, são praticamente definitivas. Isso torna o Bitcoin on-chain excelente para liquidações maiores e menos sensíveis ao tempo, porém mais lento e com uma taxa de mineração variável que não é ideal para pagamentos pequenos do dia a dia.

Os valores são sempre **satoshis** inteiros (1 BTC = 100.000.000 sats) — nunca ponto flutuante. Um destino on-chain se parece com `bc1q…` (mainnet) ou `bcrt1…` (regtest).

- **Nesta demonstração:** um endereço `bc1…` é interpretado como o meio `bitcoin`. Veja `ParseDestination` em [`mock/service.go`](../../services/freedom-bounties-api/internal/payment/mock/service.go) e o [`breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go) real; os satoshis são a unidade base em [`payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go).

## Lightning — a Camada 2 rápida para pagamentos pequenos

Lightning é uma rede construída *sobre* o Bitcoin usando canais de pagamento. Os pagamentos liquidam em bem menos de um segundo, custam de uma fração de satoshi a poucos sats e são ideais para transferências globais de baixo valor — exatamente a recompensa de `100 sats` da oficina. Os fundos continuam sendo Bitcoin; a Lightning é o mecanismo de entrega.

Quem recebe pode compartilhar três formatos, todos entendidos por esta aplicação:
- uma **fatura BOLT11** — uma string `lnbc…` de uso único com valor e validade embutidos;
- um **Lightning address** — um amigável `nome@domínio` (resolvido via LNURL nos bastidores);
- um link **LNURL-pay**.

- **Nesta demonstração:** `lnbc…`, `mentor@example.com` e `lnurl…` são interpretados como o meio `lightning`. O adaptador real usa `PrepareLnurlPay`/`LnurlPay` para endereços e `PrepareSendPayment`/`SendPayment` para faturas — veja [`breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).

## Liquid — uma sidechain do Bitcoin (contexto)

Liquid é uma **sidechain** do Bitcoin: uma rede federada separada, atrelada ao Bitcoin, com blocos de ~1 minuto, valores confidenciais e a capacidade de emitir ativos (por exemplo L-BTC e stablecoins na Liquid). É a rede de liquidação por trás do **SDK Nodeless (Liquid)** da Breez, irmão do SDK Spark.

Este projeto **não** roda na Liquid — ela é incluída aqui porque escolher um SDK Breez significa escolher uma rede, e a Liquid é a alternativa mais comum ao Spark. Se o seu produto precisasse de transferências confidenciais ou stablecoins emitidas na Liquid, o SDK Nodeless (Liquid) seria a escolha natural, e a mesma porta `PaymentService` deste repositório hospedaria esse adaptador.

- **Nesta demonstração:** não usada. A porta que tornaria a troca de rede uma mudança de um único adaptador é [`payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go) (`Service`).

## Spark — a Camada 2 usada nesta demonstração

Spark é uma Camada 2 mais recente do Bitcoin para transferências rápidas, baratas e auto-custodiais de Bitcoin e tokens. É o alvo do **Breez SDK – Spark** e a rede que este projeto integra. Quem recebe compartilha um endereço `spark1…` ou uma fatura Spark.

- **Nesta demonstração:** destinos `spark1…` são interpretados como o meio `spark`; o provedor real é o binding `breez-sdk-spark-go` conectado em [`breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go) (com build tag `breez`). As restrições atuais de SDK/lançamento — incluindo por que stablecoins cross-chain estão desativadas — estão nas [notas de implementação](00-implementation-notes.md).

## Em resumo

| Meio | Camada | Velocidade | Taxas | Melhor para | Destino na demo |
|------|--------|------------|-------|-------------|-----------------|
| Bitcoin | Cadeia base | ~10 min | Variável (mineração) | Liquidação maior e definitiva | `bc1…` |
| Lightning | L2 (canais) | Menos de 1s | Ínfimas | Pequeno, global, frequente | `lnbc…`, `nome@domínio` |
| Liquid | Sidechain | ~1 min | Baixas, fixas | Confidencial / ativos emitidos | *(não usada aqui)* |
| Spark | L2 | Rápida | Baixas | BTC + tokens auto-custodiais | `spark1…` |

A ideia da aplicação é que o domínio nunca escolhe um meio na mão: ele entrega um destino ao serviço de pagamento, que o interpreta e escolhe uma rota compatível. Continue com [ativos e meios de pagamento](04-payment-assets-and-rails.md) para ver como essa escolha é modelada.

<!-- nav-footer -->

---

<sub>📄 **Código:** [`internal/payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Infraestrutura de pagamento, não outra carteira](01-concept.md)  ·  [Ativos e meios de pagamento](04-payment-assets-and-rails.md) ▶
