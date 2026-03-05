# === Stage 1: Build the Go binary ===
FROM golang:1.25-bookworm AS builder

# Install templ CLI for template generation
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate templ files and build binary
# CGO_ENABLED=1 is required for mattn/go-sqlite3
RUN templ generate && \
    CGO_ENABLED=1 go build -o exile-tracker ./cmd

# === Stage 2: Runtime image ===
FROM debian:bookworm-slim

# Install runtime dependencies and clone PoB, then remove git (only needed at build time)
# - luajit: required by Path of Building headless mode
# - ca-certificates: required for HTTPS calls to PoE API and build upload sites
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        luajit \
        lua-zlib \
        git \
        ca-certificates && \
    git clone --depth 1 https://github.com/Chander-00/PathOfBuilding.git /app/PathOfBuilding && \
    apt-get purge -y git && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Create directories for volume mounts (DB and SSH host key)
RUN mkdir -p /app/data /app/.ssh

# Copy the compiled binary from builder stage
COPY --from=builder /build/exile-tracker .

# Copy migrations (goose runs them at startup via Go code)
COPY --from=builder /build/migrations ./migrations

# Ports: 3000 (HTTP API + Web), 2222 (SSH TUI)
EXPOSE 3000 2222

ENTRYPOINT ["./exile-tracker"]
