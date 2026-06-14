# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/vplc ./cmd/vplc

# Stage 2: run
FROM alpine:3.21

RUN addgroup -S vplc && adduser -S -G vplc vplc

WORKDIR /app
COPY --from=builder /app/vplc /app/vplc
COPY configs/docker.json /app/configs/docker.json

RUN mkdir -p /app/data /app/logs && chown -R vplc:vplc /app

USER vplc

EXPOSE 8080

VOLUME ["/app/data", "/app/logs"]

ENTRYPOINT ["/app/vplc", "--config", "/app/configs/docker.json"]
