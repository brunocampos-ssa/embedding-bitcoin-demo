# Payment lifecycle

[English](../en-US/05-payment-lifecycle.md) | [Português do Brasil](../pt-BR/05-payment-lifecycle.md)

```mermaid
sequenceDiagram
  participant U as Organizer
  participant A as API
  participant P as PaymentService
  participant R as Provider
  U->>A: payout intent + destination
  A->>P: parse and prepare
  P->>R: validate route and quote
  R-->>P: amount, fee, expiry
  P-->>U: masked review
  U->>A: explicit confirmation + idempotency key
  A->>P: send frozen preparation
  P->>R: execute once
  R-->>P: payment ID and status
  A->>R: reconcile existing ID
  A->>A: mark bounty PAID only on success
```

The lifecycle opens earlier than this diagram: the treasury must be funded first. See [treasury and deposits](18-treasury-and-deposits.md) for the deposit step, the balance precheck, and the self-payment guard that gate `prepare`.

Errors are stable application codes, not SDK or SQL details. Pending and ambiguous sends are reconciled by provider ID. No retry creates a fresh payment intent automatically.
