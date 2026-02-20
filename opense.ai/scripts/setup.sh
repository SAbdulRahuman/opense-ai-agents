#!/usr/bin/env bash
# OpeNSE.ai Development Environment Setup
# Usage: bash scripts/setup.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "╔══════════════════════════════════════╗"
echo "║   OpeNSE.ai — Development Setup     ║"
echo "╚══════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; }

# ── Check Prerequisites ──

echo "Checking prerequisites..."
echo ""

# Go
if command -v go &>/dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    ok "Go: $GO_VERSION"
else
    fail "Go not found. Install from https://go.dev/dl/"
    exit 1
fi

# Node.js
if command -v node &>/dev/null; then
    NODE_VERSION=$(node -v)
    ok "Node.js: $NODE_VERSION"
else
    fail "Node.js not found. Install from https://nodejs.org/"
    exit 1
fi

# npm
if command -v npm &>/dev/null; then
    NPM_VERSION=$(npm -v)
    ok "npm: $NPM_VERSION"
else
    fail "npm not found."
    exit 1
fi

# Git
if command -v git &>/dev/null; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    ok "Git: $GIT_VERSION"
else
    warn "Git not found (optional but recommended)"
fi

echo ""

# ── Go Dependencies ──

echo "Installing Go dependencies..."
cd "$PROJECT_ROOT"
go mod download
go mod verify
ok "Go modules downloaded and verified"

# golangci-lint (optional)
if command -v golangci-lint &>/dev/null; then
    ok "golangci-lint already installed"
else
    warn "golangci-lint not found. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

echo ""

# ── Frontend Dependencies ──

echo "Installing frontend dependencies..."
cd "$PROJECT_ROOT/web"
npm ci
ok "Frontend dependencies installed"

echo ""

# ── Configuration ──

echo "Setting up configuration..."
cd "$PROJECT_ROOT"

if [ ! -f config/config.yaml ]; then
    if [ -f config/config.example.yaml ]; then
        cp config/config.example.yaml config/config.yaml
        ok "Created config/config.yaml from example"
        warn "Edit config/config.yaml to add your API keys"
    else
        warn "No config.example.yaml found — using defaults"
    fi
else
    ok "config/config.yaml already exists"
fi

echo ""

# ── Build Verification ──

echo "Verifying builds..."

# Go build
cd "$PROJECT_ROOT"
go build -o /dev/null ./cmd/openseai/
ok "Go backend builds successfully"

# Frontend build check (type-check only, skip full build for speed)
cd "$PROJECT_ROOT/web"
npx tsc --noEmit 2>/dev/null && ok "Frontend TypeScript checks pass" || warn "Frontend type errors found"

echo ""

# ── Run Tests ──

echo "Running tests..."

cd "$PROJECT_ROOT"
if go test -short ./... &>/dev/null; then
    ok "Go tests pass"
else
    warn "Some Go tests failed — run 'make test' for details"
fi

cd "$PROJECT_ROOT/web"
if npm test &>/dev/null; then
    ok "Frontend tests pass"
else
    warn "Some frontend tests failed — run 'make ui-test' for details"
fi

echo ""

# ── Summary ──

echo "╔══════════════════════════════════════╗"
echo "║   Setup Complete!                    ║"
echo "╚══════════════════════════════════════╝"
echo ""
echo "Quick start commands:"
echo "  make dev        — Start Go backend"
echo "  make ui-dev     — Start frontend dev server"
echo "  make test       — Run all Go tests"
echo "  make ui-test    — Run frontend tests"
echo "  make bench      — Run Go benchmarks"
echo "  make help       — Show all available targets"
echo ""

# Check for API keys
if [ -z "${OPENSEAI_LLM_OPENAI_KEY:-}" ] && [ -z "${OPENSEAI_LLM_GEMINI_KEY:-}" ]; then
    warn "No LLM API keys detected in environment."
    echo "  Set one of:"
    echo "    export OPENSEAI_LLM_OPENAI_KEY='sk-...'"
    echo "    export OPENSEAI_LLM_GEMINI_KEY='...'"
    echo "    export OPENSEAI_LLM_ANTHROPIC_KEY='sk-ant-...'"
    echo ""
fi
