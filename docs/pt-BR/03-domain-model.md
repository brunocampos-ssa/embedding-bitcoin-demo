# Modelo de domínio

[English](../en-US/03-domain-model.md) | [Português do Brasil](../pt-BR/03-domain-model.md)

Recompensa, atribuição, entrega, aprovação, payout e provedor são conceitos separados. Aprovação autoriza pagar; não prova que houve pagamento.

```mermaid
stateDiagram-v2
  DRAFT --> OPEN
  OPEN --> ASSIGNED
  ASSIGNED --> SUBMITTED
  SUBMITTED --> APPROVED
  APPROVED --> PAID: pagamento confirmado
  OPEN --> EXPIRED
  DRAFT --> CANCELLED
```

```mermaid
stateDiagram-v2
  CREATED --> VALIDATING
  VALIDATING --> PREPARED
  PREPARED --> PROCESSING
  PROCESSING --> SUCCEEDED
  VALIDATING --> VALIDATION_FAILED
  PROCESSING --> PAYMENT_FAILED
  PREPARED --> EXPIRED
```

Somente entregas aprovadas podem ser preparadas; valor e destino congelam; índice parcial permite um único sucesso por entrega; confirmação exige idempotência persistida antes do provedor. A recompensa vira `PAID` apenas após `SUCCEEDED` confirmado.

<!-- nav-footer -->

---

<sub>📄 **Código:** [`internal/payout/models.go`](../../services/freedom-bounties-api/internal/payout/models.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Arquitetura](02-architecture.md)  ·  [Ciclo do pagamento](05-payment-lifecycle.md) ▶
