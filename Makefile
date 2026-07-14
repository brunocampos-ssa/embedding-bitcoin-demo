SHELL := /bin/sh
API_DIR := services/freedom-bounties-api
WEB_DIR := apps/freedom-bounties-web

# Load .env (if present) into the environment. Shell-sourced, so any value with
# spaces — notably BREEZ_MNEMONIC — must be quoted in .env.
LOAD_ENV = set -a; [ -f .env ] && . ./.env; set +a
# Build the Breez binding only when the breez provider is selected; the default
# (mock) build stays CGO-free.
BREEZ_TAG = $$([ "$$PAYMENT_PROVIDER" = breez ] && printf %s '-tags breez')

.PHONY: setup dev seed check reset check-docs api web
setup:
	cd $(API_DIR) && go mod download
	cd $(WEB_DIR) && npm install

dev:
	@$(LOAD_ENV); \
	 echo "Starting API (provider: $${PAYMENT_PROVIDER:-mock}) on $${HTTP_ADDR:-:8080} and web app on :5173; Ctrl-C stops both."; \
	 trap 'kill 0' INT TERM EXIT; \
	 (cd $(API_DIR) && go run $(BREEZ_TAG) ./cmd/api) & \
	 (cd $(WEB_DIR) && npm run dev) & \
	 wait

api:
	@$(LOAD_ENV); cd $(API_DIR) && go run $(BREEZ_TAG) ./cmd/api

web:
	cd $(WEB_DIR) && npm run dev

seed:
	cd $(API_DIR) && PAYMENT_PROVIDER=mock go run ./cmd/api >/dev/null 2>&1 & pid=$$!; sleep 1; kill $$pid 2>/dev/null || true

check-docs:
	./scripts/check-doc-parity.sh

check:
	cd $(API_DIR) && gofmt -w . && go vet ./... && go test ./...
	cd $(WEB_DIR) && npm run typecheck && npm test && npm run build
	./scripts/check-doc-parity.sh

reset:
	./scripts/reset-demo.sh
