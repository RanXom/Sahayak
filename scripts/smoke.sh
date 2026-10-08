#!/usr/bin/env bash
# End-to-end API smoke test. Run from any shell:
#   bash scripts/smoke.sh
#   ./scripts/smoke.sh
set -euo pipefail

BASE="${BASE:-http://localhost:8080/api/v1}"
PHONE="${PHONE:-9000000009}"
OTP="${OTP:-123456}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required (it is provided by the Nix dev shell)." >&2
  exit 1
fi

echo "== config =="
curl -s "$BASE/config" | jq .

R=$(curl -s -X POST "$BASE/auth/otp/request" -H 'Content-Type: application/json' -d "{\"phone\":\"$PHONE\"}")
CH=$(echo "$R" | jq -r .challenge_id)
echo "== otp requested (challenge ${CH:0:8}…, demo code $(echo "$R" | jq -r '.demo_otp // "n/a"')) =="

T=$(curl -s -X POST "$BASE/auth/otp/verify" -H 'Content-Type: application/json' \
      -d "{\"challenge_id\":\"$CH\",\"otp\":\"$OTP\"}" | jq -r .access_token)
if [ -z "$T" ] || [ "$T" = "null" ]; then
  echo "login failed" >&2
  exit 1
fi
echo "== logged in =="

echo "== profile (widow farmer, Rajasthan) =="
curl -s -X PATCH "$BASE/profile/facts" -H "Authorization: Bearer $T" -H 'Content-Type: application/json' \
  -d '{"state_id":8,"facts":{"age":45,"gender":"female","marital_status":"widowed","income_annual":120000,"land_acres":2,"occupation":"farmer","caste_category":"sc","residence_type":"rural","ration_card_type":"bpl"}}' \
  | jq '{completed, facts}'

echo "== eligibility summary =="
curl -s -X POST "$BASE/eligibility/evaluate" -H "Authorization: Bearer $T" | jq '.summary'

echo "== top eligible schemes =="
curl -s -X POST "$BASE/eligibility/evaluate" -H "Authorization: Bearer $T" \
  | jq -r '.results[] | select(.status=="eligible") | "  - \(.scheme.name_en)"' | head -8
