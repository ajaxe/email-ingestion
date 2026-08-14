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
# Stage 2: Minimal Runtime
# -----------------------------
# distroless includes root TLS certs and timezone data, perfect for outbound API calls/DB connections
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /out/email-service /bin/email-service

# Entrypoint is fixed to the multi-service binary
ENTRYPOINT ["/bin/email-service"]

# Fallback command if none passed
CMD ["--help"]