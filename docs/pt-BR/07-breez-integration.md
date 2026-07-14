# Integração com Breez SDK - Spark

[English](../en-US/07-breez-integration.md) | [Português do Brasil](../pt-BR/07-breez-integration.md)

O adaptador está em [`internal/payment/breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go) e segue os guias oficiais de [inicialização](https://sdk-doc-spark.breez.technology/guide/initializing.html), [interpretação](https://sdk-doc-spark.breez.technology/guide/parse.html), [envio](https://sdk-doc-spark.breez.technology/guide/send_payment.html), [LNURL](https://sdk-doc-spark.breez.technology/guide/lnurl_pay.html), [consulta](https://sdk-doc-spark.breez.technology/guide/list_payments.html) e [eventos](https://sdk-doc-spark.breez.technology/guide/events.html).

`DefaultConfig` e `Connect` criam uma instância duradoura. `Parse` normaliza BOLT11, Lightning address, LNURL-Pay, Bitcoin e Spark. Lightning address usa `PrepareLnurlPay`/`LnurlPay`; demais rotas usam `PrepareSendPayment`/`SendPayment`. `GetInfo` e `ReceivePayment` sustentam o saldo e o depósito da tesouraria. `GetPayment` reconcilia e `Disconnect` encerra.

Inicialização da carteira da tesouraria: quando `BREEZ_MNEMONIC` está vazio, o adaptador gera uma nova mnemonic BIP39 ([`mnemonic.go`](../../services/freedom-bounties-api/internal/payment/breez/mnemonic.go)), conecta o SDK com ela via `SeedMnemonic` e a persiste em `BREEZ_STORAGE_DIR/treasury.mnemonic` (modo `0600`), reutilizando-a nas execuções seguintes. Isso demonstra a inicialização de carteira pelo SDK sem provisionar uma seed manualmente. A mnemonic nunca é registrada em log — apenas sua origem (gerada/carregada) e o caminho do arquivo.

```bash
cd services/freedom-bounties-api
go test -tags breez ./internal/payment/breez ./cmd/api
```

Validação manual requer chave Breez, plataforma nativa e fundos — a mnemonic da tesouraria é gerada automaticamente quando `BREEZ_MNEMONIC` está vazio (ou forneça a sua, de baixo valor). Regtest serve a Spark/on-chain; Lightning exige mainnet com poucos sats. O repositório não executou pagamento financiado. Stablecoins reais aguardam o binding Go descrito nas [notas](00-implementation-notes.md).

<!-- nav-footer -->

---

<sub>📄 **Código:** [`internal/payment/breez/adapter.go`](../../services/freedom-bounties-api/internal/payment/breez/adapter.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Executando a demonstração](06-running-the-demo.md)  ·  [Modelo de segurança](08-security-model.md) ▶
