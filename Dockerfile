# ── Stage 1: Build Go binary ─────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

# Copy source and build
COPY backend/ ./backend/
RUN cd backend && CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/api

# ── Stage 2: Minimal runtime image ───────────────────────────────────────────
FROM alpine:3.21

# Copy ca-certificates from builder (golang:alpine already includes them)
# so we don't need a network call to an Alpine mirror.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

# Copy compiled binary
COPY --from=builder /app/server ./server

# Copy migrations (server reads these relative to its working directory)
COPY backend/migrations/ ./migrations/

# Copy frontend assets (served by the Go binary)
COPY frontend/ ./frontend/

# Persistent data directory (mount a volume here)
RUN mkdir -p ./data

EXPOSE 8080

ENV PORT=8080
ENV DATABASE_PATH=/app/data/app.db
ENV FRONTEND_DIR=/app/frontend

ENTRYPOINT ["./server"]

