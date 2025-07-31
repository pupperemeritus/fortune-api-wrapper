# ---- Build Stage ----
# Use the official Go image based on Debian Bookworm
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app as a static binary
# -ldflags="-w -s" strips debugging information and symbols, reducing binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /fortune-api .


# ---- Final Stage ----
# Use a minimal Debian image which has the 'fortune-mod' package available
FROM debian:bookworm-slim

# Install runtime dependencies including fortune-mod
# - 'fortune-mod' is the package you need for extended features.
# - 'wget' is for the HEALTHCHECK.
# - 'ca-certificates' is for HTTPS support.
# We clean up the apt cache to keep the image small.
RUN apt-get update && apt-get install -y fortune-mod wget ca-certificates \
  && rm -rf /var/lib/apt/lists/*

# --- Security Best Practice: Run as non-root user ---
# Create a dedicated user and group for the application
RUN addgroup --system appgroup && adduser --system --ingroup appgroup appuser

# Switch to the non-root user
USER appuser

# Set the working directory in the user's home directory
WORKDIR /home/appuser/

# Copy the static binary from the builder stage
# And ensure the new user has ownership of the file
COPY --from=builder --chown=appuser:appgroup /fortune-api .

# Expose port 8080 to the outside world
EXPOSE 8080

# Health check to ensure the application is running correctly
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run the executable
CMD ["./fortune-api"]

