# Modelo de segurança

[English](../en-US/08-security-model.md) | [Português do Brasil](../pt-BR/08-security-model.md)

> As destinatárias mantêm a autocustódia. O FreedomBounties controla apenas seu tesouro operacional de pagamentos.

Segredos da destinatária nunca entram no sistema. Segredos do tesouro ficam no servidor; a interface recebe apenas dados normalizados e mascarados. A API limita corpos, gera request IDs, mapeia erros seguros, restringe origem e declara uma identidade apenas de desenvolvimento.

Controles: limites de valor/taxa/dia, aprovação, preparação congelada e expiração, idempotência obrigatória, unicidade no banco, auditoria e reconciliação antes de tentar novamente. O SQLite local contém destinos e é sensível.

O demo não resolve comprometimento do host, roubo de mnemonic, organização maliciosa, troca de endereço, DoS, autorização, rotação, backup e contabilidade. Produção requer assinador externo/HSM/MPC, cofre de segredos, armazenamento cifrado, papéis autenticados, revisão de destino, monitoramento, backup e auditoria independente.

<!-- nav-footer -->

---

<sub>📄 **Código:** [`internal/platform/httpserver/server.go`](../../services/freedom-bounties-api/internal/platform/httpserver/server.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Integração com Breez SDK - Spark](07-breez-integration.md)  ·  [Solução de problemas](10-troubleshooting.md) ▶
