# Payment infrastructure, not another wallet

[English](../en-US/01-concept.md) | [Português do Brasil](../pt-BR/01-concept.md)

Ordinary applications already know *why* money should move: an invoice was approved, a driver completed a trip, or a contributor delivered a workshop. They should express that payment intent and delegate protocol work to a narrow payment service.

FreedomBounties never creates or recovers the recipient's wallet. The recipient chooses an interoperable external wallet and supplies only a payment address or invoice. This keeps keys and signing material out of the application while allowing Lightning's global, low-value payments, Bitcoin settlement, and Spark routes to fit familiar product workflows.

The pattern matters for global work and financial inclusion because recipients are not forced into a platform balance or proprietary account. They can receive in compatible tools available to them. FreedomBounties is one example; marketplaces, payroll tools, creator platforms, and community programs can use the same boundary.

New to these networks? The [Bitcoin, Lightning, Liquid, and Spark primer](19-bitcoin-lightning-liquid-spark.md) explains each layer — and points at the code — before you dive into the rails.

<!-- nav-footer -->

---

<sub>📄 **Code:** [`internal/payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go)</sub>

**[🏠 README](../../README.md)**  ·  [Bitcoin, Lightning, Liquid, and Spark](19-bitcoin-lightning-liquid-spark.md) ▶
