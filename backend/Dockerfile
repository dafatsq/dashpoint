# Build stage (1.25: stdlib security fixes; see govulncheck in CI)
FROM golang:1.25-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage (3.22: 3.19 reached end of life)
FROM alpine:3.22@sha256:14358309a308569c32bdc37e2e0e9694be33a9d99e68afb0f5ff33cc1f695dce

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Preserve the expected upload path in container images.
RUN addgroup -S dashpoint \
    && adduser -S dashpoint -G dashpoint \
    && mkdir -p /app/uploads \
    && chown -R dashpoint:dashpoint /app/uploads

USER dashpoint

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]
