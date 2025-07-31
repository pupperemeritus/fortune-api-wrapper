FROM golang:1.24-alpine AS builder

# Install fortune
RUN apk add --no-cache fortune

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fortune-api .

# Final stage
FROM alpine:3.22.1

# Install fortune and ca-certificates
RUN apk --no-cache add fortune ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/fortune-api .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./fortune-api"]

