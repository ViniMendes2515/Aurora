# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod files first — cached unless deps change
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/sensors-service/go.mod services/sensors-service/go.sum ./services/sensors-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/sensors-service && go mod download

# Copy source and build
COPY pkg/ ./pkg/
COPY services/sensors-service/ ./services/sensors-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/sensors-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /sensors-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /sensors-service .

EXPOSE 8081

CMD ["./sensors-service"]
