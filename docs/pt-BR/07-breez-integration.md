# Integração com Breez SDK - Spark

[English](../en-US/07-breez-integration.md) | [Português do Brasil](../pt-BR/07-breez-integration.md)

O adaptador está em [`internal/payment/breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go) e segue os guias oficiais de [inicialização](https://sdk-doc-spark.breez.technology/guide/initializing.html), [interpretação](https://sdk-doc-spark.breez.technology/guide/parse.html), [envio](https://sdk-doc-spark.breez.technology/guide/send_payment.html), [LNURL](https://sdk-doc-spark.breez.technology/guide/lnurl_pay.html), [consulta](https://sdk-doc-spark.breez.technology/guide/list_payments.html) e [eventos](https://sdk-doc-spark.breez.technology/guide/events.html).

`DefaultConfig` e `Connect` criam uma instância duradoura. `Parse` normaliza BOLT11, Lightning address, LNURL-Pay, Bitcoin e Spark. Lightning address usa `PrepareLnurlPay`/`LnurlPay`; demais rotas usam `PrepareSendPayment`/`SendPayment`. `GetPayment` reconcilia e `Disconnect` encerra.

```bash
cd services/freedom-bounties-api
go test -tags breez ./internal/payment/breez ./cmd/api
```

Validação manual requer chave Breez, nova mnemonic de tesouro mínimo, plataforma nativa e fundos. Regtest serve a Spark/on-chain; Lightning exige mainnet com poucos sats. O repositório não executou pagamento financiado. Stablecoins reais aguardam o binding Go descrito nas [notas](00-implementation-notes.md).
