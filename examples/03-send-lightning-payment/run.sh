#!/bin/sh
set -eu
prep=$(curl -sS -X POST http://localhost:8080/api/submissions/submission-finance/payouts/prepare -H 'Content-Type: application/json' -d '{"destination":"mentor@example.com","asset":"BTC"}')
id=$(printf '%s' "$prep" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
curl -sS -X POST "http://localhost:8080/api/payouts/$id/confirm" -H 'Content-Type: application/json' -H 'Idempotency-Key: 11111111-1111-4111-8111-111111111111' -d '{}'
echo
