# Executando a demonstração

[English](../en-US/06-running-the-demo.md) | [Português do Brasil](../pt-BR/06-running-the-demo.md)

No Fedora, instale Go 1.24+, Node.js 24 LTS, npm e `make`; compilação Breez também exige toolchain C. No macOS, use instaladores oficiais ou Homebrew. No Windows, use PowerShell/WSL; o Breez nativo exige a DLL indicada na documentação. O mock é Go puro.

```bash
cp .env.example .env
make setup
make dev
```

Abra `http://localhost:5173`. Os dados são inseridos automaticamente. Use `mentor@example.com`, aprove, prepare, confirme e aguarde o sucesso. Um destino compatível com `fail` demonstra falha; `expired` demonstra expiração. `MOCK_PAYMENT_FAILURE=always` força falhas.

Use `make reset` com a API parada. Docker é opcional: `docker compose up --build`. Para modo real, leia [Integração Breez](07-breez-integration.md) e financie apenas um tesouro isolado de baixo valor.

<!-- nav-footer -->

---

<sub>📄 **Código:** [`cmd/api/main.go`](../../services/freedom-bounties-api/cmd/api/main.go)</sub>

**[🏠 README](../../README.pt-BR.md)**  ·  ◀ [Tesouraria e depósitos](18-treasury-and-deposits.md)  ·  [Integração com Breez SDK - Spark](07-breez-integration.md) ▶
