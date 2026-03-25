# ─── Stage 1: Build React web client ──────────────────────────────────────────
FROM node:20-alpine AS web-builder
WORKDIR /build

COPY web-client/package*.json ./web-client/
RUN cd web-client && npm ci --no-audit --no-fund

COPY web-client/ ./web-client/
RUN cd web-client && npm run build
# Output lands at /build/web-client/dist/public/

# ─── Stage 2: Build Go server binary ──────────────────────────────────────────
FROM golang:1.24-alpine AS go-builder
WORKDIR /build

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a fully-static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" -o mmb-server ./cmd/mmb-server

# ─── Stage 3: Minimal runtime image ───────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Server binary
COPY --from=go-builder /build/mmb-server ./mmb-server

# Static web assets – server resolves ./bin/web-client/ relative to CWD (/app)
COPY --from=web-builder /build/web-client/dist/public/ ./bin/web-client/

# Seed data files (mount a volume over /app/data to persist runtime changes)
COPY data/ ./data/

EXPOSE 3000

# ─── Runtime environment defaults ─────────────────────────────────────────────
# Override any of these in the Dokploy "Environment" panel.
ENV SERVER_PORT=3000
# Set ALLOWED_ORIGIN to your frontend domain if you host it separately,
# or leave * when the server itself serves the web client (default).
ENV ALLOWED_ORIGIN=*
# Set to "true" to run attacks without a proxy list.
ENV ALLOW_NO_PROXY=false
# Use structured JSON logs in production (easier to ingest by log collectors).
ENV LOG_FORMAT=json

ENTRYPOINT ["./mmb-server"]
