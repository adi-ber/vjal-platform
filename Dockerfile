# syntax=docker/dockerfile:1

#### 1. Module download & cache ####
FROM golang:1.24-bullseye AS modcache
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

#### 2. Builder with garble ####
FROM modcache AS builder
# Install a fixed garble version
RUN go install mvdan.cc/garble@v0.12.0
WORKDIR /app
COPY . .
RUN garble build -o vjal-app ./cmd/example

#### 3. Compressor ####
FROM builder AS compressor
RUN apt-get update && apt-get install -y --no-install-recommends upx \
    && upx --fast vjal-app \
    && rm -rf /var/lib/apt/lists/*

#### 4. Final image ####
FROM scratch
# Compressed binary
COPY --from=compressor /app/vjal-app /vjal-app
# Runtime assets
COPY --from=builder /app/config.json /config.json
COPY --from=builder /app/license.json /license.json
COPY --from=builder /app/forms /forms

ENTRYPOINT ["/vjal-app"]
