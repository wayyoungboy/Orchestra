#!/bin/bash
# test_secretary_communication.sh - Tests core secretary communication flow
# Usage: ./test_secretary_communication.sh
set -e

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
WORKSPACE_ID="${WORKSPACE_ID:-01KNJ11EQNWH5V07J89BHV9QYK}"
CONV_ID="${CONV_ID:-conv_ee58e688}"
PASS=0
FAIL=0

pass() { echo "  ✅ PASS: $1"; PASS=$((PASS+1)); }
fail() { echo "  ❌ FAIL: $1"; FAIL=$((FAIL+1)); }
info() { echo "  ℹ️  INFO: $1"; }

echo "🧪 Secretary Communication Test Suite"
echo "======================================"
echo ""

# Step 1: Health check
echo "1. Health check"
HEALTH=$(curl -sf "$BASE_URL/health" 2>/dev/null)
if echo "$HEALTH" | grep -q '"status":"ok"'; then
    pass "Server is healthy"
else
    fail "Server health check failed"
    exit 1
fi
echo ""

# Step 2: Check secretary member has ACP enabled
echo "2. Secretary ACP configuration"
MEMBERS=$(curl -sf "$BASE_URL/api/workspaces/$WORKSPACE_ID/members" 2>/dev/null || curl -sf "http://127.0.0.1:8080/api/workspaces/$WORKSPACE_ID/members" 2>/dev/null)
SECRETARY_ID=$(echo "$MEMBERS" | python3 -c "
import sys, json
members = json.load(sys.stdin)
for m in members:
    if m['roleType'] == 'secretary':
        print(m['id'])
        break
" 2>/dev/null)

if [ -n "$SECRETARY_ID" ]; then
    info "Secretary ID: $SECRETARY_ID"
    ACP_ENABLED=$(echo "$MEMBERS" | python3 -c "
import sys, json
members = json.load(sys.stdin)
for m in members:
    if m['roleType'] == 'secretary':
        print('true' if m.get('acpEnabled') else 'false')
        break
" 2>/dev/null)
    if [ "$ACP_ENABLED" = "true" ]; then
        pass "Secretary ACP is enabled"
    else
        fail "Secretary ACP is NOT enabled"
        info "Fix: curl -X PUT $BASE_URL/api/workspaces/$WORKSPACE_ID/members/$SECRETARY_ID -d '{\"acpEnabled\":true,\"acpCommand\":\"claude\",\"acpArgs\":[\"--output-format\",\"stream-json\",\"--input-format\",\"stream-json\",\"--verbose\"]}'"
        exit 1
    fi
else
    fail "No secretary member found"
    exit 1
fi
echo ""

# Step 3: Send message to secretary
echo "3. Send message to secretary"
MSG_ID=$(curl -sf -X POST "$BASE_URL/api/workspaces/$WORKSPACE_ID/conversations/$CONV_ID/messages" \
    -H 'Content-Type: application/json' \
    -d "{\"text\": \"@1 Test $RANDOM - just reply OK.\", \"senderId\": \"owner\"}" \
    | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

if [ -n "$MSG_ID" ]; then
    pass "Message sent: $MSG_ID"
else
    fail "Failed to send message"
    exit 1
fi
echo ""

# Step 4: Wait for AI response
echo "4. Wait for AI response (timeout: 60s)"
AI_RESPONDED=false
for i in $(seq 1 12); do
    sleep 5
    AI_MSG=$(curl -sf "$BASE_URL/api/workspaces/$WORKSPACE_ID/conversations/$CONV_ID/messages" \
        | python3 -c "
import sys, json
msgs = json.load(sys.stdin)
ai = [m for m in msgs if m.get('isAi')]
if ai:
    latest = ai[-1]
    text = latest.get('content', {}).get('text', '')[:100]
    print(f'{latest[\"id\"]}|{text}')
" 2>/dev/null)

    if [ -n "$AI_MSG" ]; then
        AI_RESPONDED=true
        pass "AI responded: $AI_MSG"
        break
    fi
done

if [ "$AI_RESPONDED" = false ]; then
    fail "No AI response received after 60 seconds"
    info "Check: pkill -f 'claude.*stream-json' then restart backend"
fi
echo ""

# Summary
echo "======================================"
echo "Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
