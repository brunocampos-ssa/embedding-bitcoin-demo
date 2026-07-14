# Notas de implementação

[English](../en-US/00-implementation-notes.md) | [Português do Brasil](../pt-BR/00-implementation-notes.md)

Verificado em 13/07/2026 com a [documentação oficial do Breez SDK - Spark](https://sdk-doc-spark.breez.technology/) e o código oficial de [`breez-sdk-spark-go`](https://github.com/breez/breez-sdk-spark-go).

O módulo fixado é `github.com/breez/breez-sdk-spark-go v0.15.1`, pacote `breez_sdk_spark`. O binding usa CGO e bibliotecas nativas; o modo mock padrão não as carrega, graças à build tag `breez`. Esta versão expõe apenas `NetworkMainnet` e `NetworkRegtest`. Regtest atende Spark, on-chain e tokens; Lightning deve ser testado em mainnet com valores mínimos.

Inicialização: `DefaultConfig`, `SeedMnemonic`, `ConnectRequest`, `Connect` e `Disconnect`. Interpretação: `Parse`. Envio geral: `PrepareSendPayment`, `SendPayment`, `GetPayment`. Lightning address/LNURL-Pay: `PrepareLnurlPay` e `LnurlPay`. Eventos: `AddEventListener` e `EventListener.OnEvent`.

Há uma única instância duradoura para um único tesouro. O server mode oficial serve a ambientes multi-tenant que criam uma instância efêmera por carteira e requisição; seria complexidade inadequada aqui.

## Limites importantes

A documentação atual descreve entrega cross-chain de USDC/USDT em redes EVM, Solana e Tron, partindo de sats BTC ou USDB e usando duas etapas coordenadas por um provedor. Porém, o binding Go lançado em `v0.15.1` não contém tipos de endereço, descoberta de rotas nem pagamento cross-chain. USDT/USDC real fica desativado; o mock e o modelo normalizado permanecem disponíveis.

BTC por Lightning, on-chain e Spark compila contra o SDK nativo, mas não foi testado ponta a ponta com credenciais e fundos. Chamadas Go são síncronas e não recebem `context.Context`; o adaptador verifica cancelamento antes da chamada, mas não interrompe trabalho nativo já iniciado.

Preparações do SDK ficam em memória. Após reinício, uma preparação ainda não enviada deve expirar e ser refeita; pagamentos com ID do provedor são reconciliados por `GetPayment`. Nunca criamos automaticamente uma nova intenção diante de resultado ambíguo. Mnemonic em variável de ambiente serve apenas à demonstração; produção deve usar assinador externo. A chave idempotente do SDK precisa ser UUID; a interface usa `crypto.randomUUID()`.

<!-- nav-footer -->

---

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Solução de problemas](10-troubleshooting.md)  ·  [Próximos passos](11-next-steps.md) ▶
