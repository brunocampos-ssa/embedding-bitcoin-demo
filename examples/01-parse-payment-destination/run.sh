#!/bin/sh
set -eu
curl -sS -X POST http://localhost:8080/api/submissions/submission-finance/payouts/prepare -H 'Content-Type: application/json' -d '{"destination":"mentor@example.com","asset":"BTC"}'
echo
