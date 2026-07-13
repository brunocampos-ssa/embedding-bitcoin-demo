# Ciclo do pagamento

[English](../en-US/05-payment-lifecycle.md) | [Português do Brasil](../pt-BR/05-payment-lifecycle.md)

```mermaid
sequenceDiagram
  participant U as Organização
  participant A as API
  participant P as PaymentService
  participant R as Provedor
  U->>A: intenção + destino
  A->>P: interpretar e preparar
  P->>R: validar rota e cotação
  R-->>U: valor, taxa, validade e destino mascarado
  U->>A: confirmação + chave idempotente
  A->>R: enviar preparação congelada uma vez
  R-->>A: ID e estado
  A->>R: reconciliar o mesmo ID
  A->>A: PAID somente após sucesso
```

Erros expõem códigos estáveis, não SQL ou detalhes do SDK. Resultado pendente ou ambíguo é reconciliado pelo ID existente; nunca se cria nova intenção automaticamente.
