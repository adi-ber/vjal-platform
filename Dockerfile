# syntax=docker/dockerfile:1

# ---------- Builder Stage ----------
FROM golang:1.21-bullseye AS builder

# Install garble and upx
RUN go install mvdan.cc/garble@latest \
 && apt-get update && apt-get install -y --no-install-recommends upx \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy source
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build with garble (obfuscation) and then compress with upx
RUN garble build -o vjal-app ./cmd/example \
 && upx --ultra-brute vjal-app

# ---------- Final Stage ----------
FROM scratch

# Copy only the compressed binary
COPY --from=builder /app/vjal-app /vjal-app

# Default entrypoint
ENTRYPOINT ["/vjal-app"]
