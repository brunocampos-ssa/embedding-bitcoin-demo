# Roteiro da oficina de duas horas

[English](../en-US/09-workshop-walkthrough.md) | [Português do Brasil](../pt-BR/09-workshop-walkthrough.md)

Um roteiro para a pessoa facilitadora conduzir uma sessão prática de duas horas. Ao final, o público consegue explicar como **embutir pagamentos Bitcoin em uma aplicação comum** por meio de uma fronteira de serviço de pagamento — sem construir uma carteira e sem nunca guardar as chaves de quem recebe.

- **Público:** pessoas desenvolvedoras à vontade lendo Go e TypeScript. Não se pressupõe experiência com Bitcoin ou Lightning.
- **Objetivo:** percorrer um pagamento de ponta a ponta — **financiar → aprovar → validar → revisar → confirmar → reconciliar** — e ver exatamente onde o Breez SDK se encaixa.
- **Antes de começar:** `cp .env.example .env` e depois `make setup`. Mantenha `PAYMENT_PROVIDER=mock` (seguro, nenhum valor real se move). Deixe este repositório e o [guia de execução](06-running-the-demo.md) abertos.

Os horários assumem início às 18:00; ajuste à vontade. Cada trecho lista o que **fazer** na tela, o que **explicar** e a **ideia central** a fixar.

## 1 · 18:00–18:10 — Enquadramento: aplicação vs. carteira

- **Fazer:** `make dev`, abrir <http://localhost:5173>, mostrar a lista de recompensas.
- **Explicar:** construímos o lado *pagador* de uma aplicação comum. Quem recebe mantém a própria carteira e chaves; operamos apenas uma pequena tesouraria de pagamentos. Nomeie o que **não** construímos — criação de carteira, recuperação, portfólio, swaps — versus o que **construímos**: uma intenção de pagamento.
- **Ideia central:** embutir pagamentos não é construir uma carteira.

## 2 · 18:10–18:25 — Arquitetura e estados do domínio

- **Fazer:** percorrer o diagrama em [arquitetura](02-architecture.md) e os estados em [modelo de domínio](03-domain-model.md).
- **Explicar:** a porta `PaymentService` com dois adaptadores intercambiáveis (mock e Breez). O ciclo de recompensa/envio e a diferença entre *aprovação* (decisão humana) e *liquidação* (o dinheiro de fato se movendo).
- **Ideia central:** a porta é a costura que mantém os detalhes de Bitcoin fora do domínio.

## 3 · 18:25–18:45 — Financiar a tesouraria e rodar o fluxo semeado

- **Fazer:** no painel da Tesouraria, note que o saldo começa em **0**. Crie um depósito Lightning (ex.: 1.000 sats) e veja o saldo subir (~1s no mock). Depois abra a recompensa semeada e execute: aprovar → colar `mentor@example.com` → validar → revisar valor e taxa → confirmar → ver o sucesso.
- **Explicar:** o ciclo *começa pelo financiamento* — não se paga de uma tesouraria vazia. Demonstre as duas proteções de propósito: pule o financiamento para acionar `INSUFFICIENT_TREASURY_FUNDS`, e cole o próprio endereço da tesouraria para acionar `SELF_PAYMENT_REJECTED`.
- **Ideia central:** depósito → saldo → pagamento é toda a história do dinheiro.

## 4 · 18:45–19:05 — Ler o código por conceito

- **Fazer:** abrir, nesta ordem — [`payment/models.go`](../../services/freedom-bounties-api/internal/payment/models.go) (a porta e o modelo normalizado de valor), [`mock/service.go`](../../services/freedom-bounties-api/internal/payment/mock/service.go), [`payout/service.go`](../../services/freedom-bounties-api/internal/payout/service.go) (política, idempotência, reconciliação) e [`database.go`](../../services/freedom-bounties-api/internal/platform/database/database.go) (restrições).
- **Explicar:** valores são normalizados como `AmountBaseUnits`; um índice único parcial garante um pagamento bem-sucedido por envio; a verificação prévia de saldo é defesa em profundidade, independente do provedor.
- **Ideia central:** a segurança mora no serviço de pagamento e nas restrições do banco, não na interface.

## 5 · 19:05–19:25 — Interpretar, preparar, congelar, revisar taxas

- **Fazer:** experimente cada tipo de destino (`lnbc…`, `bc1…`, `spark1…`, `nome@example.com`). Prepare um pagamento e inspecione o valor congelado, a taxa e a validade na etapa de revisão.
- **Explicar:** `ParseDestination` normaliza muitas entradas em um único formato; a preparação *congela* valor e destino para que a confirmação não os altere; as taxas variam por trilho.
- **Ideia central:** a revisão mostra exatamente o que será enviado, e nada muda após o congelamento.

## 6 · 19:25–19:45 — Confirmar, idempotência, reconciliação

- **Fazer:** confirme um pagamento. Adicione `fail` a um destino para forçar uma falha. Atualize um pagamento em processamento para observar a reconciliação.
- **Explicar:** o cabeçalho `Idempotency-Key` torna segura uma confirmação repetida; PROCESSING → SUCCEEDED/FAILED é reconciliado pelo ID do provedor; uma recompensa só vira PAID após sucesso. Se faltarem credenciais ou fundos Breez, permaneça no mock e leia o adaptador Breez compilado.
- **Ideia central:** uma intenção confirmada gera no máximo um pagamento, e o estado é reconciliado, não presumido.

## 7 · 19:45–19:55 — Modo Breez real e inicialização da carteira pelo SDK

- **Fazer:** abrir [integração Breez](07-breez-integration.md). Explique que definir `PAYMENT_PROVIDER=breez` no `.env` faz o `make dev` compilar com `-tags breez` automaticamente. Com `BREEZ_MNEMONIC` vazio, o SDK **inicializa uma nova carteira de tesouraria**: uma mnemonic BIP39 é gerada, conectada via `SeedMnemonic` e persistida em `treasury.mnemonic` (modo `0600`), de modo que a mesma carteira retorna na próxima execução.
- **Explicar:** compare os caminhos de taxa e liquidação on-chain versus Spark; prefira regtest; Lightning em mainnet move satoshis reais.
- **Ideia central:** o SDK pode criar e reconectar a carteira por você — você fornece a política, não a gestão de chaves.

## 8 · 19:55–20:00 — Limites, segurança e perguntas

- **Fazer:** cobrir a lacuna de lançamento de USDT/USDC, a segurança da tesouraria e a direção de assinador externo — [modelo de segurança](08-security-model.md) e [segurança da tesouraria em produção](17-production-treasury-security.md).
- **Ideia central:** nomeie o que muda no caminho da demonstração até a produção.

**Plano B:** mantenha a API em modo mock, use os destinos determinísticos de demonstração e mostre os links dos guias oficiais do Breez mais `go test -tags breez`. Nunca deixe a pressão do tempo justificar expor ou financiar a mnemonic de uma carteira pessoal — no modo real, deixe a aplicação gerar uma tesouraria descartável.
