# Roteiro da oficina de duas horas

[English](../en-US/09-workshop-walkthrough.md) | [Português do Brasil](../pt-BR/09-workshop-walkthrough.md)

- 18:00–18:10 — Contexto, objetivo e “aplicação versus carteira”.
- 18:10–18:25 — Arquitetura, estados e aprovação versus liquidação.
- 18:25–18:40 — `make dev` e fluxo mock completo.
- 18:40–19:00 — `PaymentService`, mock, serviço de payout e restrições.
- 19:00–19:20 — Interpretar, preparar, congelar e revisar taxas.
- 19:20–19:40 — Confirmar Lightning; idempotência e reconciliação. Sem infraestrutura real, use processamento/falha mock e o código Breez compilado.
- 19:40–19:50 — On-chain e Spark; regtest quando disponível.
- 19:50–20:00 — USDT/USDC, segurança, assinador externo e perguntas.

Plano de contingência: permaneça no mock, use destinos determinísticos, links oficiais e `go test -tags breez`. Nunca exponha ou financie uma mnemonic pessoal por pressão de tempo.
