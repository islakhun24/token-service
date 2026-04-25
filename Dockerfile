# ===== BUILD STAGE =====
FROM golang:1.25-bookworm AS buildstage

ENV TZ=Asia/Jakarta
ENV DEBIAN_FRONTEND=noninteractive
ENV GO111MODULE=on

WORKDIR /app

# Cache deps dulu (biar build lebih cepat)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -v -o token-service ./cmd/server/

# ===== DEPLOY STAGE =====
FROM debian:bookworm-slim AS deploystage

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
 && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Jakarta
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app

# Copy binary dari build stage
COPY --from=buildstage /app/token-service ./token-service

EXPOSE 8080

CMD ["./token-service"]