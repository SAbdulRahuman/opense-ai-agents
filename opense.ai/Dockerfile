# ── Stage 1: Build Next.js static export ─────────────────────────
FROM node:20-alpine AS web-builder

WORKDIR /app/web

# Cache npm dependencies
COPY web/package.json web/package-lock.json* ./
RUN npm ci --prefer-offline

# Copy source and build static export (produces web/out/)
COPY web/ .
RUN npm run build

# ── Stage 2: Build Go binary (with embedded web UI) ──────────────
FROM golang:1.23-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Copy static export into web/out/ so go:embed picks it up
COPY --from=web-builder /app/web/out /app/web/out

# Build single binary with embedded web UI
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev) -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
    -o /app/build/openseai ./cmd/openseai

# ── Stage 3: Final runtime image ─────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S openseai && adduser -S openseai -G openseai

WORKDIR /app

# Copy single binary (includes embedded web UI)
COPY --from=go-builder /app/build/openseai /app/openseai

# Copy config
COPY config/config.example.yaml /app/config/config.yaml
COPY config/agents.yaml /app/config/agents.yaml

# Set permissions
RUN chown -R openseai:openseai /app

USER openseai

# Single port: Go API + embedded web UI
EXPOSE 8080

# Start server (web UI at /, API at /api/v1)
CMD ["/app/openseai", "serve"]
