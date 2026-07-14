# Running the demo

[English](../en-US/06-running-the-demo.md) | [Português do Brasil](../pt-BR/06-running-the-demo.md)

On Fedora install Go 1.24+, Node.js 24 LTS, npm, `make`, and a C toolchain only for Breez builds. On macOS use official Go/Node installers or Homebrew. On Windows use PowerShell/WSL; native Breez builds also require the packaged DLL placement described by Breez. Mock mode is pure Go and needs no CGO.

```bash
cp .env.example .env
make setup
make dev
```

Open `http://localhost:5173`. Seed data is automatic and idempotent. The first bounty already has a submitted workshop. Use `mentor@example.com`; then approve, prepare, confirm, and wait briefly for success. Use a supported value containing `fail` for a failure, `expired` for expiry, or set `MOCK_PAYMENT_FAILURE=always`.

Run `make reset` to remove only the configured demo `.db`; restart the API to seed again. Docker Compose is optional: `docker compose up --build`. For real mode follow [Breez integration](07-breez-integration.md), fund only a low-value treasury, and run the API with `-tags breez`.

<!-- nav-footer -->

---

<sub>📄 **Code:** [`cmd/api/main.go`](../../services/freedom-bounties-api/cmd/api/main.go)</sub>

**[🏠 README](../../README.md)**  ·  ◀ [Treasury and deposits](18-treasury-and-deposits.md)  ·  [Breez SDK - Spark integration](07-breez-integration.md) ▶
