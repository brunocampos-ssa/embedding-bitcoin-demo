# Solução de problemas

[English](../en-US/10-troubleshooting.md) | [Português do Brasil](../pt-BR/10-troubleshooting.md)

- **Entrada incompatível:** use BOLT11, Lightning address, LNURL-Pay, endereço Bitcoin da rede ou Spark.
- **Expirado:** prepare novamente; não reutilize cotação vencida.
- **Chave/mnemonic ausente:** modo Breez falha fechado.
- **Rede:** apenas `mainnet` e `regtest` nesta versão.
- **Saldo/limite:** financie o tesouro isolado ou reduza o valor; confira políticas.
- **Inicialização:** confira plataforma nativa, permissões, chave, rede e conexão.
- **Banco:** pare a API e use `make reset`; nunca durante pagamento real pendente.
- **Frontend/API:** confira portas 5173/8080 e `VITE_API_BASE_URL`.
- **CORS:** `ALLOWED_ORIGIN` deve ser origem exata; wildcard é rejeitado fora de desenvolvimento.
- **Porta ocupada:** altere `HTTP_ADDR` ou a porta Vite.
- **Pendente:** consulte o mesmo payout/ID; não crie outra intenção.
- **Duplicado:** reutilize a mesma chave; outra é rejeitada.
- **Real desativado:** compile com `-tags breez`; stablecoin cross-chain segue indisponível no Go `v0.15.1`.

<!-- nav-footer -->

---

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Modelo de segurança](08-security-model.md)  ·  [Notas de implementação](00-implementation-notes.md) ▶
