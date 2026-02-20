#!/usr/bin/env bash
# OpeNSE.ai — NSE Data Source Smoke Tests
# Verifies that NSE data endpoints are accessible and returning valid data.
# Usage: bash scripts/test_nse.sh
set -euo pipefail

echo "╔══════════════════════════════════════╗"
echo "║   OpeNSE.ai — NSE Smoke Tests       ║"
echo "╚══════════════════════════════════════╝"
echo ""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0

pass() { echo -e "${GREEN}PASS${NC} $1"; ((PASS++)); }
fail() { echo -e "${RED}FAIL${NC} $1"; ((FAIL++)); }
skip() { echo -e "${YELLOW}SKIP${NC} $1"; ((SKIP++)); }

# Check if curl is available
if ! command -v curl &>/dev/null; then
    echo "curl is required but not found"
    exit 1
fi

# Check if jq is available (optional, for JSON parsing)
HAS_JQ=false
if command -v jq &>/dev/null; then
    HAS_JQ=true
fi

# ── NSE Website Accessibility ──

echo "1. NSE Website Accessibility"
echo "───────────────────────────"

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 \
    -H "User-Agent: Mozilla/5.0" \
    "https://www.nseindia.com/" 2>/dev/null || echo "000")

if [ "$HTTP_CODE" = "200" ]; then
    pass "NSE website accessible (HTTP $HTTP_CODE)"
else
    fail "NSE website returned HTTP $HTTP_CODE"
fi

echo ""

# ── NSE Quote API ──

echo "2. NSE Quote API"
echo "────────────────"

QUOTE_RESPONSE=$(curl -s --max-time 10 \
    -H "User-Agent: Mozilla/5.0" \
    -H "Accept: application/json" \
    "https://www.nseindia.com/api/quote-equity?symbol=TCS" 2>/dev/null || echo "error")

if echo "$QUOTE_RESPONSE" | grep -qi "TCS\|lastPrice\|priceInfo"; then
    pass "NSE quote API returns TCS data"
elif echo "$QUOTE_RESPONSE" | grep -qi "error\|forbidden\|blocked"; then
    skip "NSE quote API blocked (IP/rate limit) — normal in CI"
else
    skip "NSE quote API response unrecognized"
fi

echo ""

# ── NSE Option Chain API ──

echo "3. NSE Option Chain API"
echo "───────────────────────"

OC_RESPONSE=$(curl -s --max-time 10 \
    -H "User-Agent: Mozilla/5.0" \
    -H "Accept: application/json" \
    "https://www.nseindia.com/api/option-chain-indices?symbol=NIFTY" 2>/dev/null || echo "error")

if echo "$OC_RESPONSE" | grep -qi "strikePrice\|CE\|PE\|openInterest"; then
    pass "NSE option chain API returns NIFTY data"
elif echo "$OC_RESPONSE" | grep -qi "error\|forbidden\|blocked"; then
    skip "NSE option chain API blocked (IP/rate limit) — normal in CI"
else
    skip "NSE option chain API response unrecognized"
fi

echo ""

# ── Yahoo Finance API ──

echo "4. Yahoo Finance API"
echo "────────────────────"

YF_RESPONSE=$(curl -s --max-time 10 \
    "https://query1.finance.yahoo.com/v8/finance/chart/TCS.NS?interval=1d&range=5d" 2>/dev/null || echo "error")

if echo "$YF_RESPONSE" | grep -qi "chart\|timestamp\|close"; then
    pass "Yahoo Finance API returns TCS.NS data"
else
    fail "Yahoo Finance API not accessible"
fi

echo ""

# ── Local API Server (if running) ──

echo "5. Local API Server"
echo "────────────────────"

LOCAL_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 \
    "http://localhost:8080/api/v1/status" 2>/dev/null || echo "000")

if [ "$LOCAL_CODE" = "200" ]; then
    pass "Local API server running (HTTP $LOCAL_CODE)"
elif [ "$LOCAL_CODE" = "000" ]; then
    skip "Local API server not running (start with 'make serve')"
else
    skip "Local API server returned HTTP $LOCAL_CODE"
fi

echo ""

# ── Summary ──

echo "═══════════════════════════════════════"
TOTAL=$((PASS + FAIL + SKIP))
echo -e "Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}, ${YELLOW}${SKIP} skipped${NC} (${TOTAL} total)"

if [ "$FAIL" -gt 0 ]; then
    echo ""
    echo "Note: Some NSE endpoints may be blocked by IP/rate limits."
    echo "This is normal in CI environments or when behind corporate proxies."
    exit 1
fi

exit 0
