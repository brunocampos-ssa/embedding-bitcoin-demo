# 02 — Preparar um pagamento

[English](README.md) | [Português do Brasil](README.pt-BR.md)

Com a API mock ativa e a entrega aprovada, `./run.sh` prepara uma revisão on-chain sem enviar. Espere `PREPARED`, `100` sats, taxa determinística, destino mascarado e validade. Leia [ciclo do pagamento](../../docs/pt-BR/05-payment-lifecycle.md), [`payout/service.go`](../../services/freedom-bounties-api/internal/payout/service.go) e o [adaptador Breez](../../services/freedom-bounties-api/internal/payment/breez/adapter.go).
