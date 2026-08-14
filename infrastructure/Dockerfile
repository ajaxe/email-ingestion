# -----------------------------
# Stage 1: Build Environment
# -----------------------------
FROM golang:1.26.2-alpine AS builder

WORKDIR /src

# Leverage Docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build static Linux binary (adjust the package path if your entry is under ./cmd/...)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /out/email-service .

# -----------------------------
# Stage 2: Minimal Alpine Runtime
# -----------------------------
# Alpine OS allows console access/troubleshooting (/bin/sh) while remaining lightweight
FROM alpine:3.20

# Install CA certificates and timezone data for outbound HTTPS & DB connection support
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /out/email-service /app/email-service

# Entrypoint is fixed to the multi-service binary
ENTRYPOINT ["/app/email-service"]

# Fallback command if none passed
CMD ["--help"]