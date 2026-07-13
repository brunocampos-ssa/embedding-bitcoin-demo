# 01 — Interpretar um destino

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Requer API mock e a entrega inicial aprovada. Execute `make api`, aprove pela interface e rode `./run.sh`. O JSON esperado mostra `lightning-address`, `lightning`, destino mascarado, valor, taxa e validade. O endpoint combina interpretação, validação e cotação; a operação direta fica atrás de [`PaymentService`](../../services/freedom-bounties-api/internal/payment/models.go). Leia [ativos e meios](../../docs/pt-BR/04-payment-assets-and-rails.md) e o [adaptador Breez](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
