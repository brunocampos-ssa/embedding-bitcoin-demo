# Security model

[English](../en-US/08-security-model.md) | [Português do Brasil](../pt-BR/08-security-model.md)

> Recipients remain self-custodial. FreedomBounties controls only its operational payout treasury.

Recipient secrets never cross this system. Treasury secrets remain server-side and the frontend sees only normalized, masked payment data. Destination values are masked in info logs. APIs bound request bodies, use request IDs, map safe errors, enforce an origin allowlist, and expose an explicit development-only actor rather than fake authentication.

Controls include amount/fee/daily limits, approval checks, frozen preparations, expiry, mandatory idempotency, database uniqueness, audit records, and reconciliation before retry. The local SQLite database contains payment destinations and should be treated as sensitive; its directory is created with owner-only permissions.

Threats not solved by this demo include compromised server/host, stolen mnemonic, malicious organizer, recipient address substitution, denial of service, authorization, secret rotation, backup, and treasury accounting. Environment mnemonics are demo-only. Production evolution should use an external signer/HSM/MPC, secret manager, encrypted storage, authenticated roles, destination-change review, monitoring, backups, and independent security review.

<!-- nav-footer -->

---

<sub>📄 **Code:** [`internal/platform/httpserver/server.go`](../../services/freedom-bounties-api/internal/platform/httpserver/server.go)</sub>

**[🏠 README](../../README.md)**  ·  ◀ [Breez SDK - Spark integration](07-breez-integration.md)  ·  [Two-hour workshop walkthrough](09-workshop-walkthrough.md) ▶
