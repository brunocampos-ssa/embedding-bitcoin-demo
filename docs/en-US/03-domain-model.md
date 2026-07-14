# Domain model

[English](../en-US/03-domain-model.md) | [Português do Brasil](../pt-BR/03-domain-model.md)

Separate concepts are bounty, assignment, submission, approval, payout, and payment provider. Approval authorizes a payout but does not prove that money moved.

```mermaid
stateDiagram-v2
  DRAFT --> OPEN
  OPEN --> ASSIGNED
  ASSIGNED --> SUBMITTED
  SUBMITTED --> APPROVED
  APPROVED --> PAID: payout confirmed successful
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

Invariants: only approved submissions may be prepared; amount and destination freeze at preparation; a unique partial index permits only one successful payout per submission; confirmation requires an idempotency key; the key is persisted before the provider call; a different key cannot replace an existing payout. The bounty reaches `PAID` only while reconciling a provider `SUCCEEDED` result.

<!-- nav-footer -->

---

<sub>📄 **Code:** [`internal/payout/models.go`](../../services/freedom-bounties-api/internal/payout/models.go)</sub>

**[🏠 README](../../README.md)**  ·  ◀ [Architecture](02-architecture.md)  ·  [Payment lifecycle](05-payment-lifecycle.md) ▶
