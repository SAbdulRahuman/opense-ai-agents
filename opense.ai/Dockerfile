# ── Stage 1: Build Go backend ─────────────────────────────────────
FROM golang:1.23-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev) -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
    -o /app/build/openseai ./cmd/openseai

# ── Stage 2: Build Next.js frontend ──────────────────────────────
FROM node:20-alpine AS web-builder

WORKDIR /app/web

# Cache npm dependencies
COPY web/package.json web/package-lock.json* ./
RUN npm ci --prefer-offline

# Copy source and build
COPY web/ .
RUN npm run build

# ── Stage 3: Final runtime image ─────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S openseai && adduser -S openseai -G openseai

WORKDIR /app

# Copy Go binary
COPY --from=go-builder /app/build/openseai /app/openseai

# Copy Next.js standalone build
COPY --from=web-builder /app/web/.next/standalone /app/web/
COPY --from=web-builder /app/web/.next/static /app/web/.next/static
COPY --from=web-builder /app/web/public /app/web/public

# Copy config
COPY config/config.example.yaml /app/config/config.yaml
COPY config/agents.yaml /app/config/agents.yaml

# Set permissions
RUN chown -R openseai:openseai /app

USER openseai

# Expose ports: 8080 (Go API), 3000 (Next.js)
EXPOSE 8080 3000

# Default: start Go API server
CMD ["/app/openseai", "serve"]
