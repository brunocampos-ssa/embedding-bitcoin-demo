# Infraestrutura de pagamento, não outra carteira

[English](../en-US/01-concept.md) | [Português do Brasil](../pt-BR/01-concept.md)

Aplicações comuns já sabem *por que* o dinheiro deve se mover: uma entrega foi aprovada ou uma oficina foi realizada. Elas devem declarar essa intenção e delegar protocolos a um serviço estreito.

FreedomBounties não cria nem recupera a carteira da destinatária. Ela escolhe uma carteira externa interoperável e fornece apenas endereço ou fatura. Assim, chaves e assinaturas ficam fora da aplicação, enquanto Lightning, Bitcoin e Spark entram em fluxos conhecidos.

O padrão favorece trabalho global e inclusão financeira porque ninguém fica preso a saldo ou conta proprietária. Marketplaces, folha de pagamento, plataformas de criadores e programas comunitários podem usar a mesma fronteira.

Novo nessas redes? A [introdução a Bitcoin, Lightning, Liquid e Spark](19-bitcoin-lightning-liquid-spark.md) explica cada camada — e aponta para o código — antes de você mergulhar nos meios de pagamento.

<!-- nav-footer -->

---

<sub>📄 **Código:** [`internal/payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  [Bitcoin, Lightning, Liquid e Spark](19-bitcoin-lightning-liquid-spark.md) ▶
