SHELL := /bin/sh
API_DIR := services/freedom-bounties-api
WEB_DIR := apps/freedom-bounties-web

.PHONY: setup dev seed check reset check-docs api web
setup:
	cd $(API_DIR) && go mod download
	cd $(WEB_DIR) && npm install

dev:
	@echo "Starting API on :8080 and web app on :5173; Ctrl-C stops both."
	@trap 'kill 0' INT TERM EXIT; (cd $(API_DIR) && go run ./cmd/api) & (cd $(WEB_DIR) && npm run dev) & wait

api:
	cd $(API_DIR) && go run ./cmd/api

web:
	cd $(WEB_DIR) && npm run dev

seed:
	cd $(API_DIR) && go run ./cmd/api >/dev/null 2>&1 & pid=$$!; sleep 1; kill $$pid 2>/dev/null || true

check-docs:
	./scripts/check-doc-parity.sh

check:
	cd $(API_DIR) && gofmt -w . && go vet ./... && go test ./...
	cd $(WEB_DIR) && npm run typecheck && npm test && npm run build
	./scripts/check-doc-parity.sh

reset:
	./scripts/reset-demo.sh
